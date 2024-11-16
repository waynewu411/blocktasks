package repository

import (
	"context"

	"github.com/waynewu411/blocktasks/pkg/do"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TaskDao interface {
	InsertTask(ctx context.Context, task do.Task) (do.Task, error)
	UpdateTask(ctx context.Context, task do.Task) (do.Task, error)
	GetTask(ctx context.Context, name string) (do.Task, error)
	GetTaskForUpdate(ctx context.Context, name string) (do.Task, error)
}

type taskDao struct {
	db *gorm.DB
}

func NewTaskDao(db *gorm.DB) TaskDao {
	return &taskDao{db: db}
}

func (t *taskDao) InsertTask(ctx context.Context, task do.Task) (do.Task, error) {
	if err := t.db.WithContext(ctx).Create(&task).Error; err != nil {
		return do.Task{}, transformGormError(err)
	}
	return task, nil
}

func (t *taskDao) UpdateTask(ctx context.Context, task do.Task) (do.Task, error) {
	if err := t.db.WithContext(ctx).Save(&task).Error; err != nil {
		return do.Task{}, transformGormError(err)
	}
	return task, nil
}

func (t *taskDao) GetTask(ctx context.Context, name string) (do.Task, error) {
	var task do.Task
	err := t.db.WithContext(ctx).Where("name = ?", name).First(&task).Error
	if err != nil {
		return do.Task{}, transformGormError(err)
	}
	return task, nil
}

func (t *taskDao) GetTaskForUpdate(ctx context.Context, name string) (do.Task, error) {
	var task do.Task
	err := t.db.WithContext(ctx).Where("name = ?", name).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&task).Error
	if err != nil {
		return do.Task{}, transformGormError(err)
	}
	return task, nil
}
