package repository

import (
	"context"

	"github.com/waynewu411/blocktasks/pkg/do"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LogDao interface {
	InsertLogs(ctx context.Context, logs []do.Log) error
}

type logDao struct {
	db *gorm.DB
}

func NewLogDao(db *gorm.DB) LogDao {
	return &logDao{db: db}
}

func (e *logDao) InsertLogs(ctx context.Context, logs []do.Log) error {
	err := e.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "chain_id"}, {Name: "txn_hash"}, {Name: "log_index"}},
			DoNothing: true,
		}).
		Create(&logs).Error
	if err != nil {
		return transformGormError(err)
	}
	return nil
}
