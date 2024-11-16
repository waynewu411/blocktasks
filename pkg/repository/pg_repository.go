package repository

import (
	"fmt"
	"time"

	"github.com/waynewu411/blocktasks/pkg/config"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type pgRepository struct {
	lg  *zap.Logger
	cfg config.PgConfig
	db  *gorm.DB

	taskDao TaskDao
	logDao  LogDao
}

type customNamingStrategy struct {
	schema.NamingStrategy
	DbSchema string
}

func NewPgRepository(lg *zap.Logger, cfg config.PgConfig) Repository {
	gormCfg := &gorm.Config{
		NamingStrategy: customNamingStrategy{
			DbSchema: cfg.Schema,
		},
	}

	if cfg.DebugEnabled {
		gormCfg.Logger = gormLogger.Default.LogMode(gormLogger.Info)
	}

	db, err := gorm.Open(postgres.Open(cfg.Url), gormCfg)
	if err != nil {
		lg.Fatal("fail to connect to postgres database", zap.Error(err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		lg.Fatal("fail to get sql db", zap.Error(err))
	}

	if err := sqlDB.Ping(); err != nil {
		lg.Fatal("failed to ping postgres database", zap.Error(err))
	}
	lg.Info("successfully connected to postgres database")

	sqlDB.SetMaxIdleConns(int(cfg.MaxIdleConns))
	sqlDB.SetMaxOpenConns(int(cfg.MaxOpenConns))
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxConnLifeTime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.MaxConnIdleTime) * time.Second)

	pgRepository := &pgRepository{
		lg:      lg,
		cfg:     cfg,
		db:      db,
		taskDao: NewTaskDao(db),
		logDao:  NewLogDao(db),
	}

	return pgRepository
}

func (r *pgRepository) Transaction(fn func(Repository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		txRepo := &pgRepository{
			lg:      r.lg,
			cfg:     r.cfg,
			db:      tx,
			taskDao: NewTaskDao(tx),
			logDao:  NewLogDao(tx),
		}
		return fn(txRepo)
	})
}

func (r *pgRepository) TaskDao() TaskDao {
	return r.taskDao
}

func (r *pgRepository) LogDao() LogDao {
	return r.logDao
}

func (ns customNamingStrategy) TableName(table string) string {
	return fmt.Sprintf("%s.%s", ns.DbSchema, table)
}
