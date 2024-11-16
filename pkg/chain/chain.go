package chain

import (
	"context"
)

const (
	ChainIdBaseMainnet int64 = 8453 // 0x2105
)

const (
	BlockNumberLatest    int64 = -1
	BlockNumberFinalized int64 = -2
	BlockNumberSafe      int64 = -3
)

type Block struct {
	ChainId     int64    `json:"chainId"`
	BlockNumber int64    `json:"number"`
	BlockHash   string   `json:"hash"`
	Timestamp   int64    `json:"timestamp"` // in milli seconds
	Txns        []Txn    `json:"transactions"`
	TxnHashes   []string `json:"transactionHashes"`
	Logs        []Log    `json:"logs"`
}

type Txn struct {
	BlockHash   string `json:"blockHash"`
	BlockNumber string `json:"blockNumber"`
	TxnHash     string `json:"hash"`
	Type        string `json:"type"`
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
}

type Log struct {
	Address     string   `json:"address"`
	BlockNumber int64    `json:"blockNumber"`
	BlockHash   string   `json:"blockHash"`
	Data        string   `json:"data"`
	Topics      []string `json:"topics"`
	TxnHash     string   `json:"transactionHash"`
	LogIndex    int64    `json:"logIndex"`
	Removed     bool     `json:"removed"` // in milli seconds
}

type Chain interface {
	GetChainId() int64
	GetBlockByNumber(ctx context.Context, blockNumber int64, fullTxns bool) (Block, error)
	GetBlocks(ctx context.Context, fromBlockNumber int64, toBlockNumber int64, fullTxns bool, includeLogs bool, addresses []string, topics []string) ([]Block, error)
}
