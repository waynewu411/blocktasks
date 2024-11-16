package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/waynewu411/blocktasks/pkg/config"
	"github.com/waynewu411/blocktasks/pkg/do"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func getTestConfig() *config.PgConfig {
	return &config.PgConfig{
		Schema:       "public",
		Url:          "postgres://root:root@localhost:5432/blocktasks?sslmode=disable",
		DebugEnabled: true,
	}
}

func TestPgRepository_NewPgRepository(t *testing.T) {
	pgRepo := NewPgRepository(zaptest.NewLogger(t, zaptest.Level(zapcore.DebugLevel)), getTestConfig())
	require.NotNil(t, pgRepo)
}

func TestPgRepository_Transaction(t *testing.T) {
	pgRepo := NewPgRepository(zaptest.NewLogger(t, zaptest.Level(zapcore.DebugLevel)), getTestConfig())
	require.NotNil(t, pgRepo)

	ctx := context.Background()

	err := pgRepo.Transaction(func(repo Repository) error {
		taskDao := repo.TaskDao()
		eventDao := repo.EventDao()

		task := do.Task{
			Name:                        "Task_1",
			LastProcessedBlockNumber:    1000,
			LastProcessedBlockTimestamp: 2000,
		}
		task, err := taskDao.InsertTask(ctx, task)
		if err != nil {
			return err
		}

		event := do.Event{
			ChainId:     1,
			BlockNumber: 1000,
			BlockHash:   "0x123",
			Data:        "data",
			Topics:      []string{"topic1", "topic2"},
			TxnHash:     "0x456",
			LogIndex:    1,
			Removed:     false,
			Timestamp:   2000,
		}
		event, err = eventDao.InsertEvent(ctx, event)
		if err != nil {
			return err
		}

		return nil
	})

	require.NoError(t, err)
}

func TestPgRepository_TransactionRevert(t *testing.T) {
	pgRepo := NewPgRepository(zaptest.NewLogger(t, zaptest.Level(zapcore.DebugLevel)), getTestConfig())
	require.NotNil(t, pgRepo)

	ctx := context.Background()

	err := pgRepo.Transaction(func(repo Repository) error {
		taskDao := repo.TaskDao()
		eventDao := repo.EventDao()

		task := do.Task{
			Name:                        "Task_1",
			LastProcessedBlockNumber:    1000,
			LastProcessedBlockTimestamp: 2000,
		}
		task, err := taskDao.InsertTask(ctx, task)
		if err != nil {
			return err
		}

		event := do.Event{
			ChainId:     1,
			BlockNumber: 1000,
			BlockHash:   "0x123",
			Data:        "data",
			Topics:      []string{"topic1", "topic2"},
			TxnHash:     "0x456",
			LogIndex:    1,
			Removed:     false,
			Timestamp:   2000,
		}
		event, err = eventDao.InsertEvent(ctx, event)
		if err != nil {
			return err
		}

		return errors.New("Revert")
	})

	require.Error(t, err)
}
