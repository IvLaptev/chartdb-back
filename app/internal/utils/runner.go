package utils

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

type RunnerConfig struct {
	GraceTimeout time.Duration `yaml:"grace_timeout"`
	ForceTimeout time.Duration `yaml:"force_timeout"`
}

var ErrSignalExit = errors.New("signal")

type runner struct {
	logger *slog.Logger
	cfg    RunnerConfig

	errG     *errgroup.Group
	aliveCTX context.Context
}

func NewRunner(ctx context.Context, logger *slog.Logger, cfg RunnerConfig) (runner, context.Context) {
	errG, ctx := errgroup.WithContext(ctx)

	return runner{
		logger: logger,
		cfg:    cfg,

		aliveCTX: ctx,
		errG:     errG,
	}, ctx
}

func (r runner) Run(run, grace func() error) {
	r.errG.Go(run)
	r.register(grace, nil)
}

func (r runner) RunContext(run func(ctx context.Context) error, grace func() error) {
	r.errG.Go(func() error {
		return run(r.aliveCTX)
	})
	r.register(grace, nil)
}

func (r runner) RunExternal(grace func() error) {
	r.register(grace, nil)
}

func (r runner) RunGraceForce(run, grace, force func() error) {
	r.errG.Go(run)
	r.register(grace, force)
}

func (r runner) RunGraceContext(run func() error, grace func(ctx context.Context) error) {
	r.errG.Go(run)
	r.register(func() error {
		// at this point aliveCTX is canceled. We need to un cancel it for grace period to give grace a chance
		ctx := context.WithoutCancel(r.aliveCTX)
		ctx, cancel := context.WithTimeout(ctx, r.cfg.GraceTimeout)
		defer cancel()

		return grace(ctx)
	}, nil)
}

func (r runner) Wait() error {
	r.errG.Go(func() error {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(ch)

		select {
		case sig := <-ch:
			return fmt.Errorf("got signal %v: %w", sig, ErrSignalExit)
		case <-r.aliveCTX.Done():
			return nil
		}
	})

	return r.errG.Wait()
}

func (r runner) register(grace, force func() error) {
	r.errG.Go(func() error {
		<-r.aliveCTX.Done()
		var err error
		if grace != nil {
			graceDone := make(chan struct{})
			go func() {
				defer close(graceDone)
				err = grace()
			}()
			graceTimer := time.NewTimer(r.cfg.GraceTimeout)
			defer graceTimer.Stop()
			select {
			case <-graceDone:
				if err == nil {
					return nil
				}
			case <-graceTimer.C:
				r.logger.Error("grace timeout")
			}
		}

		if force != nil {
			forceDone := make(chan struct{})
			go func() {
				defer close(forceDone)
				err = force()
			}()
			forceTimer := time.NewTimer(r.cfg.ForceTimeout)
			defer forceTimer.Stop()
			select {
			case <-forceDone:
			case <-forceTimer.C:
				r.logger.Error("force timeout")
			}
		}
		return err
	})
}
