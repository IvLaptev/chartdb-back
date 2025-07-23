package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/IvLaptev/chartdb-back/internal/storage"
	"github.com/IvLaptev/chartdb-back/pkg/ctxlog"
	xstorage "github.com/IvLaptev/chartdb-back/pkg/storage"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
)

const (
	instrumentationName = "storage/postgres"
)

type Storage struct {
	db               *sqlx.DB
	logger           *slog.Logger
	defaultDBContext *dbContext
}

func (s *Storage) DB(ctx context.Context) sqlx.ExtContext {
	if currentDBCtx, ok := getDBContext(ctx); ok {
		return currentDBCtx.db
	}

	return s.db
}

func (s *Storage) DoInTransaction(ctx context.Context, f func(ctx context.Context) error) error {
	if currentDBCtx, ok := getDBContext(ctx); ok {
		if currentDBCtx.isTx {
			return f(ctx)
		}
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	var txID string
	err = tx.GetContext(ctx, &txID, "SELECT pg_current_xact_id()::text")
	if err != nil {
		return fmt.Errorf("get transaction id: %w", err)
	}

	dbCtx := &dbContext{
		db:   tx,
		now:  s.defaultDBContext.now,
		isTx: true,
	}

	ctx = setDBContext(ctx, dbCtx)

	defer func() {
		if p := recover(); p != nil {
			if err := tx.Rollback(); err != nil {
				ctxlog.Warn(ctx, s.logger, "can't rollback transaction", slog.Any("error", err))
			}
			panic(p)
		}
	}()

	err = f(ctx)
	txerr := handleTx(tx, err)
	if txerr != nil {
		txerr = fmt.Errorf("tx error: %w", txerr)
	}

	return errors.Join(err, txerr)
}

func (s *Storage) Now(ctx context.Context) xstorage.Timestamp {
	dbContext, ok := getDBContext(ctx)
	if ok {
		return dbContext.now()
	}

	return s.defaultDBContext.now()
}

func (s *Storage) Shutdown() error {
	return s.db.Close()
}

func (s *Storage) Diagram() storage.DiagramRepository {
	return s
}

func (s *Storage) User() storage.UserRepository {
	return s
}

func (s *Storage) UserConfirmation() storage.UserConfirmationRepository {
	return s
}

func handleTx(tx *sqlx.Tx, err error) error {
	if err != nil {
		return tx.Rollback()
	}
	return tx.Commit()
}

type dbContext struct {
	db   sqlx.ExtContext
	now  func() xstorage.Timestamp
	isTx bool
}

type dbContextKey struct{}

func setDBContext(ctx context.Context, dbCtx *dbContext) context.Context {
	return context.WithValue(ctx, dbContextKey{}, dbCtx)
}

func getDBContext(ctx context.Context) (*dbContext, bool) {
	value, ok := ctx.Value(dbContextKey{}).(*dbContext)
	return value, ok
}

type Config struct {
	Host        string `yaml:"host"`
	Port        uint64 `yaml:"port"`
	Database    string `yaml:"database"`
	User        string `yaml:"user"`
	Password    string `yaml:"password"`
	SSLMode     string `yaml:"ssl_mode"`
	SSLRootCert string `yaml:"ssl_root_cert"`

	MaxIdleTime time.Duration `yaml:"max_idle_time"`
	MaxLifeTime time.Duration `yaml:"max_life_time"`
	MaxIdleConn int           `yaml:"max_idle_conn"`
	MaxOpenConn int           `yaml:"max_open_conn"`

	LogLevel    string        `yaml:"log_level"`
	InitTimeout time.Duration `yaml:"init_timeout"`
}

func NewStorage(cfg Config, logger *slog.Logger) (*Storage, error) {
	logger = logger.With(slog.String("name", instrumentationName))
	dsn := connectionString(cfg)

	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	config.PreferSimpleProtocol = true
	config.Logger = newLogger(logger)
	config.RuntimeParams["timezone"] = "UTC" // get UTC values from db
	logLevel, err := pgx.LogLevelFromString(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}
	config.LogLevel = logLevel

	db := sqlx.NewDb(stdlib.OpenDB(*config), "pgx")
	db.SetConnMaxIdleTime(cfg.MaxIdleTime)
	db.SetConnMaxLifetime(cfg.MaxLifeTime)
	db.SetMaxIdleConns(cfg.MaxIdleConn)
	db.SetMaxOpenConns(cfg.MaxOpenConn)

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("start postgres connection: %w", err)
	}

	defaultDBContext := &dbContext{
		db: db,
		now: func() xstorage.Timestamp {
			return xstorage.NewTimestamp(time.Now())
		},
		isTx: false,
	}

	return &Storage{
		db:               db,
		logger:           logger,
		defaultDBContext: defaultDBContext,
	}, err
}

func connectionString(cfg Config) string {
	values := map[string]string{
		"host":     cfg.Host,
		"port":     strconv.FormatUint(cfg.Port, 10),
		"database": cfg.Database,
		"user":     cfg.User,
	}
	if cfg.SSLMode != "" {
		values["sslmode"] = cfg.SSLMode
	}
	if cfg.SSLRootCert != "" {
		values["sslrootcert"] = cfg.SSLRootCert
	}
	password := cfg.Password
	if password != "" {
		values["password"] = password
	}

	keyValues := make([]string, 0, len(values))
	for key, value := range values {
		keyValues = append(keyValues, key+"="+value)
	}
	return strings.Join(keyValues, " ")
}
