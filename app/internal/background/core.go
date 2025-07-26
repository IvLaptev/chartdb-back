package background

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/IvLaptev/chartdb-back/internal/storage"
	"github.com/IvLaptev/chartdb-back/pkg/ctxlog"
	"github.com/IvLaptev/chartdb-back/pkg/s3client"
)

type Worker interface {
	Start(context.Context) error
	Stop() error
}

type worker struct {
	logger *slog.Logger

	stopCh chan struct{}
	jobs   []Job
}

func (w *worker) run() {
out:
	for {
		select {
		case <-w.stopCh:
			break out
		case <-time.After(500 * time.Millisecond):
			now := time.Now().Unix()
			for _, job := range w.jobs {
				go func() {
					defer func() {
						if err := recover(); err != nil {
							ctxlog.Error(context.Background(), w.logger, "panic", slog.Any("error", err), slog.String("details", string(debug.Stack())))
						}
					}()

					job.Run(now)
				}()
			}
		}
	}
}

func (w *worker) Start(ctx context.Context) error {
	var jobNames []string
	for _, job := range w.jobs {
		jobNames = append(jobNames, job.Name())
	}
	ctxlog.Info(ctx, w.logger, "starting background worker", slog.Any("jobs", jobNames))

	go w.run()

	return nil
}

func (w *worker) Stop() error {
	w.stopCh <- struct{}{}
	ctxlog.Info(context.Background(), w.logger, "background worker stopped")
	return nil
}

func NewWorker(
	logger *slog.Logger,
	s3client s3client.Client,
	storage storage.Storage,
) Worker {
	return &worker{
		logger: logger,
		stopCh: make(chan struct{}),
		jobs: []Job{
			&CleanObjectStorageJob{
				logger:   logger,
				s3client: s3client,
				storage:  storage,
				period:   1 * time.Hour,
			},
		},
	}
}

type Job interface {
	Name() string
	Run(int64)
}
