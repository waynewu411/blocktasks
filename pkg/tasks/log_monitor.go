package tasks

import (
	"context"
	"time"

	"github.com/samber/lo"
	"github.com/waynewu411/blocktasks/pkg/chain"
	"github.com/waynewu411/blocktasks/pkg/config"
	"github.com/waynewu411/blocktasks/pkg/do"
	"github.com/waynewu411/blocktasks/pkg/repository"
	"go.uber.org/zap"
)

type LogMonitor struct {
	baseTask
	cfg                      config.EventMonitorConfig
	repo                     repository.Repository
	chain                    chain.Chain
	lastProcessedBlockNumber int64
	lastProcessedTimestamp   int64
}

func NewLogMonitor(lg *zap.Logger, name string, cfg config.EventMonitorConfig, repo repository.Repository, chain chain.Chain) Task {
	return &LogMonitor{
		baseTask: baseTask{
			lg:   lg,
			name: name,
		},
		cfg:   cfg,
		repo:  repo,
		chain: chain,
	}
}

func (m *LogMonitor) Start(ctx context.Context) error {
	err := m.init(ctx)
	if err != nil {
		m.lg.Error("fail to initialize", zap.String("name", m.name), zap.Error(err))
		return err
	}

	err = m.run(ctx)
	if err != nil {
		m.lg.Error("fail to run", zap.String("name", m.name), zap.Error(err))
		return err
	}

	return nil
}

func (m *LogMonitor) init(ctx context.Context) error {
	m.lg.Debug("initializing...", zap.String("name", m.name))

	task, err := m.repo.TaskDao().GetTask(ctx, m.name)
	if err != nil && err != repository.ErrRecordNotFound {
		return err
	}

	if err == nil {
		m.lg.Debug("task found", zap.String("name", m.name), zap.Any("task", task))
		m.lastProcessedBlockNumber = task.LastProcessedBlockNumber
		m.lastProcessedTimestamp = task.LastProcessedBlockTimestamp
		return nil
	}

	latestConfirmedBlock, err := m.chain.GetBlockByNumber(ctx, chain.BlockNumberSafe, false)
	if err != nil {
		m.lg.Error("fail to get latest confirmed block", zap.Error(err))
		return err
	}

	lastProcessedBlockNumber := latestConfirmedBlock.BlockNumber
	if lastProcessedBlockNumber > 0 {
		lastProcessedBlockNumber = lastProcessedBlockNumber - 1
	}
	lastProcessedBlockTimestamp := latestConfirmedBlock.Timestamp
	if lastProcessedBlockTimestamp > 0 {
		lastProcessedBlockTimestamp = lastProcessedBlockTimestamp - 1
	}

	task = do.Task{
		Name:                        m.name,
		LastProcessedBlockNumber:    lastProcessedBlockNumber,
		LastProcessedBlockTimestamp: lastProcessedBlockTimestamp,
	}

	task, err = m.repo.TaskDao().InsertTask(ctx, task)
	if err != nil {
		m.lg.Error("fail to insert task", zap.Any("task", task), zap.Error(err))
		return err
	}

	m.lastProcessedBlockNumber = lastProcessedBlockNumber
	m.lastProcessedTimestamp = lastProcessedBlockTimestamp

	m.lg.Debug("initialized", zap.String("name", m.name), zap.Any("task", task))

	return nil
}

func (m *LogMonitor) run(ctx context.Context) error {
	ticker := time.NewTicker(time.Duration(m.cfg.PollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.lg.Error("stopped", zap.String("name", m.name))
			return ctx.Err()
		case <-ticker.C:
			func() {
				defer func() {
					if r := recover(); r != nil {
						m.lg.Error("panic", zap.String("name", m.name), zap.Any("error", r), zap.Stack("stack"))
					}
				}()

				lastProcessedBlockNumber := m.lastProcessedBlockNumber

				latestConfirmedBlock, err := m.chain.GetBlockByNumber(ctx, chain.BlockNumberSafe, false)
				if err != nil {
					m.lg.Error("fail to get latest confirmed block", zap.String("name", m.name), zap.Error(err))
					return
				}
				m.lg.Debug("latest confirmed block", zap.String("name", m.name), zap.Int64("blockNumber", latestConfirmedBlock.BlockNumber))
				latestBlockNumber := latestConfirmedBlock.BlockNumber

				if latestBlockNumber <= (lastProcessedBlockNumber + m.cfg.BlockDistance) {
					return
				}

				startBlockNumber := lastProcessedBlockNumber + 1
				endBlockNumber := latestBlockNumber - m.cfg.BlockDistance
				_ = m.processBlocks(ctx, startBlockNumber, endBlockNumber)
			}()
		}
	}
}

func (m *LogMonitor) processBlocks(ctx context.Context, fromBlockNumber int64, toBlockNumber int64) error {
	i := fromBlockNumber
	for i <= toBlockNumber {
		j := i + m.cfg.QueryMaxBlocks
		if j > toBlockNumber {
			j = toBlockNumber
		}
		err := m.queryAndFilterLogsInBlocksWithRetry(ctx, i, j)
		if err != nil {
			return err
		}
		i = j + 1
	}

	return nil
}

func (m *LogMonitor) queryAndFilterLogsInBlocksWithRetry(ctx context.Context, fromBlockNumber int64, toBlockNumber int64) error {
	var err error
	for i := int64(0); i < m.cfg.MaxBlockRetries; i++ {
		err = m.queryAndFilterLogsInBlocks(ctx, fromBlockNumber, toBlockNumber)
		if err == nil {
			m.lg.Debug(
				"query and filter logs in blocks succeeded",
				zap.String("name", m.name),
				zap.Int64("fromBlockNumber", fromBlockNumber),
				zap.Int64("toBlockNumber", toBlockNumber),
			)
			break
		}
		m.lg.Error(
			"query and filter logs in blocks failed",
			zap.String("name", m.name),
			zap.Int64("fromBlockNumber", fromBlockNumber),
			zap.Int64("toBlockNumber", toBlockNumber),
			zap.Error(err),
		)
	}

	return err
}

func (m *LogMonitor) queryAndFilterLogsInBlocks(ctx context.Context, fromBlockNumber int64, toBlockNumber int64) error {
	blocks, err := m.chain.GetBlocks(
		ctx,
		fromBlockNumber,
		toBlockNumber,
		false,
		true,
		m.cfg.MonitoredContractAddresses,
		[]string{},
	)
	if err != nil {
		m.lg.Error(
			"fail to get blocks",
			zap.String("name", m.name),
			zap.Int64("fromBlockNumber", fromBlockNumber),
			zap.Int64("toBlockNumber", toBlockNumber),
			zap.Error(err),
		)
		return err
	}

	for blockNumber := fromBlockNumber; blockNumber <= toBlockNumber; blockNumber++ {
		block := blocks[blockNumber-fromBlockNumber]
		txErr := m.repo.Transaction(func(repo repository.Repository) error {
			logDOs := lo.Map(block.Logs, func(log chain.Log, _ int) do.Log {
				return do.Log{
					ChainId:     m.chain.GetChainId(),
					BlockNumber: log.BlockNumber,
					BlockHash:   log.BlockHash,
					Data:        log.Data,
					Topics:      log.Topics,
					TxnHash:     log.TxnHash,
					LogIndex:    log.LogIndex,
					Removed:     log.Removed,
					Timestamp:   block.Timestamp,
				}
			})
			err := repo.LogDao().InsertLogs(ctx, logDOs)
			if err != nil {
				m.lg.Error("fail to insert logs", zap.String("name", m.name), zap.Int64("blockNumber", blockNumber), zap.Error(err))
				return err
			}
			m.lg.Debug("logs inserted", zap.String("name", m.name), zap.Int64("blockNumber", blockNumber), zap.Int("logs", len(logDOs)))

			task := do.Task{
				Name:                        m.name,
				LastProcessedBlockNumber:    block.BlockNumber,
				LastProcessedBlockTimestamp: block.Timestamp,
			}
			task, err = repo.TaskDao().UpdateTask(ctx, task)
			if err != nil {
				m.lg.Error("fail to update task", zap.String("name", m.name), zap.Error(err))
				return err
			}
			m.lg.Debug("task updated", zap.String("name", m.name), zap.Any("task", task))

			return nil
		})
		if txErr != nil {
			return txErr
		}
	}

	return nil
}
