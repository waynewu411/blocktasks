package main

import (
	"context"

	"github.com/waynewu411/blocktasks/pkg/chain"
	"github.com/waynewu411/blocktasks/pkg/config"
	"github.com/waynewu411/blocktasks/pkg/httpclient"
	"github.com/waynewu411/blocktasks/pkg/logger"
	"github.com/waynewu411/blocktasks/pkg/repository"
	"github.com/waynewu411/blocktasks/pkg/request"
	"github.com/waynewu411/blocktasks/pkg/tasks"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var (
	Version string = "latest"
	Build   string = ""
)

func logVersionAndBuild(lg *zap.Logger) {
	lg.Info("blocktasks", zap.String("version", Version), zap.String("build", Build))
}

func main() {
	lg, closer := logger.NewLogger()
	defer closer()

	logVersionAndBuild(lg)

	cfg := config.LoadConfig(lg)

	pgRepo := repository.NewPgRepository(lg, cfg.PgConfig)

	eg, ctx := errgroup.WithContext(context.Background())

	if cfg.BaseEventMonitorConfig.Enabled {
		request := request.NewRequest(
			lg,
			request.WithHttpClient(
				httpclient.NewHttpClient(lg, cfg.BaseEventMonitorConfig.ChainConfig.HttpClientConfig),
			),
		)
		baseChain := chain.NewBaseChain(lg, cfg.BaseEventMonitorConfig.ChainConfig, request)
		eg.Go(func() error {
			return tasks.NewLogMonitor(lg, tasks.TaskBaseLogMonitor, cfg.BaseEventMonitorConfig, pgRepo, baseChain).Start(ctx)
		})
	}

	if err := eg.Wait(); err != nil {
		lg.Fatal("blocktasks stopped", zap.Error(err))
	}
}
