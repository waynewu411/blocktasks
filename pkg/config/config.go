package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type PgConfig struct {
	Schema          string `mapstructure:"SCHEMA"`
	Url             string `mapstructure:"URL"`
	DebugEnabled    bool   `mapstructure:"DEBUG_ENABLED"`
	MaxIdleConns    int64  `mapstructure:"MAX_IDLE_CONNS"`
	MaxOpenConns    int64  `mapstructure:"MAX_OPEN_CONNS"`
	MaxConnLifeTime int64  `mapstructure:"MAX_CONN_LIFETIME"`  // in seconds
	MaxConnIdleTime int64  `mapstructure:"MAX_CONN_IDLE_TIME"` // in seconds
}

type HttpClientConfig struct {
	DebugEnabled bool  `mapstructure:"DEBUG_ENABLED"`
	RateLimit    int64 `mapstructure:"RATE_LIMIT"` // request per second
}

type ChainConfig struct {
	HttpClientConfig HttpClientConfig `mapstructure:"HTTP_CLIENT_CONFIG"`
	ApiEndpoint      string           `mapstructure:"API_ENDPOINT"`
	ApiKey           string           `mapstructure:"API_KEY"`
}

type EventMonitorConfig struct {
	Enabled                    bool        `mapstructure:"ENABLED"`
	ChainConfig                ChainConfig `mapstructure:"CHAIN_CONFIG"`
	PollInterval               int64       `mapstructure:"POLL_INTERVAL"`     // in seconds
	QueryMaxBlocks             int64       `mapstructure:"QUERY_MAX_BLOCKS"`  // maximum blocks in each query
	MaxBlockRetries            int64       `mapstructure:"MAX_BLOCK_RETRIES"` // maximum retries on failure for each block1
	BlockDistance              int64       `mapstructure:"BLOCK_DISTANCE"`    // the distance to the latest block
	MonitoredContractAddresses []string    `mapstructure:"MONITORED_CONTRACT_ADDRESSES"`
}

type Config struct {
	PgConfig               PgConfig           `mapstructure:"PG_CONFIG"`
	BaseEventMonitorConfig EventMonitorConfig `mapstructure:"BASE_EVENT_MONITOR_CONFIG"`
}

var (
	cfg Config
)

func initDefaultValues() {
	viper.SetDefault("PG_CONFIG", PgConfig{
		Schema:          "public",
		Url:             "postgres://root:root@localhost:5432/postgres?sslmode=disable",
		DebugEnabled:    true,
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		MaxConnLifeTime: 30 * 60, // 30 minutes
		MaxConnIdleTime: 10 * 60, // 10 minutes
	})
	viper.SetDefault("BASE_EVENT_MONITOR_CONFIG",
		EventMonitorConfig{
			Enabled: true,
			ChainConfig: ChainConfig{
				HttpClientConfig: HttpClientConfig{
					DebugEnabled: true,
					RateLimit:    100,
				},
				ApiEndpoint: "https://base-mainnet.g.alchemy.com/v2",
				ApiKey:      "",
			},
			PollInterval:               3,
			QueryMaxBlocks:             50,
			MaxBlockRetries:            3,
			BlockDistance:              0,
			MonitoredContractAddresses: []string{},
		},
	)
}

func init() {
	initDefaultValues()
}

func LoadConfig(logger *zap.Logger) *Config {
	viper.AutomaticEnv()

	configFiles := os.Getenv("CONFIG_FILES")
	if configFiles == "" {
		loadSingleConfig(logger, "./config.json")
		if err := viper.ReadInConfig(); err != nil {
			logger.Fatal("failed to read default config", zap.Error(err))
		}
	} else {
		// Load multiple config files
		files := strings.Split(configFiles, ",")
		for i, file := range files {
			loadSingleConfig(logger, strings.TrimSpace(file))
			if i == 0 {
				err := viper.ReadInConfig()
				if err != nil {
					logger.Fatal("failed to read config", zap.String("file", file), zap.Error(err))
				}
			} else {
				err := viper.MergeInConfig()
				if err != nil {
					logger.Fatal("failed to merge config", zap.String("file", file), zap.Error(err))
				}
			}
		}
	}

	err := viper.Unmarshal(&cfg)
	if err != nil {
		logger.Fatal("failed to unmarshal config from config file(s)", zap.Error(err))
	}

	logger.Info("config", zap.Any("config", cfg))

	return &cfg
}

func loadSingleConfig(logger *zap.Logger, configFile string) {
	dir := filepath.Dir(configFile)
	base := filepath.Base(configFile)
	ext := filepath.Ext(base)
	configName := strings.TrimSuffix(base, ext)
	configType := strings.TrimPrefix(ext, ".")
	logger.Info("config file details",
		zap.String("configFile", configFile),
		zap.String("path", dir),
		zap.String("name", configName),
		zap.String("type", configType))
	viper.AddConfigPath(dir)
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)
}
