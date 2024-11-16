package chain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"github.com/waynewu411/blocktasks/pkg/config"
	"github.com/waynewu411/blocktasks/pkg/request"
	"github.com/waynewu411/blocktasks/pkg/request/jsonrpc"
	"go.uber.org/zap"
)

type BaseBlockWithFullTxns struct {
	BlockNumber string    `json:"number"`
	BlockHash   string    `json:"hash"`
	Txns        []BaseTxn `json:"transactions"`
	GasLimit    string    `json:"gasLimit"`
	GasUsed     string    `json:"gasUsed"`
	Timestamp   string    `json:"timestamp"`
}

type BaseBlockWithoutFullTxns struct {
	BlockNumber string   `json:"number"`
	BlockHash   string   `json:"hash"`
	Txns        []string `json:"transactions"`
	GasLimit    string   `json:"gasLimit"`
	GasUsed     string   `json:"gasUsed"`
	Timestamp   string   `json:"timestamp"`
}

type BaseTxn struct {
	BlockHash   string `json:"blockHash"`
	BlockNumber string `json:"blockNumber"`
	TxnHash     string `json:"hash"`
	TxnIndex    string `json:"transactionIndex"`
	Type        string `json:"type"`
	Nonce       string `json:"nonce"`
	Input       string `json:"input"`
	R           string `json:"r"`
	S           string `json:"s"`
	V           string `json:"v"`
	Gas         string `json:"gas"`
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
	GasPrice    string `json:"gasPrice"`
}

type BaseLog struct {
	Address     string   `json:"address"`
	BlockHash   string   `json:"blockHash"`
	BlockNumber string   `json:"blockNumber"`
	Data        string   `json:"data"`
	LogIndex    string   `json:"logIndex"`
	Removed     bool     `json:"removed"`
	Topics      []string `json:"topics"`
	TxnHash     string   `json:"transactionHash"`
	TxnIndex    string   `json:"transactionIndex"`
}

type getLogsParam struct {
	Addresses       []string `json:"address"`
	FromBlockNumber string   `json:"fromBlock"`
	ToBlockNumber   string   `json:"toBlock"`
	Topics          []string `json:"topics"`
}

type BaseChain struct {
	lg      *zap.Logger
	cfg     config.ChainConfig
	request request.Request
}

func NewBaseChain(lg *zap.Logger, cfg config.ChainConfig, request request.Request) Chain {
	return &BaseChain{lg: lg, cfg: cfg, request: request}
}

func (b *BaseChain) getApiUrl() string {
	return fmt.Sprintf("%s/%s", b.cfg.ApiEndpoint, b.cfg.ApiKey)
}

func (b *BaseChain) GetChainId() int64 {
	return ChainIdBaseMainnet
}

func (b *BaseChain) GetBlockByNumber(ctx context.Context, blockNumber int64, fullTxns bool) (Block, error) {
	var blockNumberStr string
	switch {
	case blockNumber == BlockNumberLatest:
		blockNumberStr = "latest"
	case blockNumber == BlockNumberFinalized:
		blockNumberStr = "finalized"
	case blockNumber == BlockNumberSafe:
		blockNumberStr = "safe"
	case blockNumber >= 0:
		blockNumberStr = "0x" + strconv.FormatInt(blockNumber, 16)
	default:
		return Block{}, errors.New("invalid block number")
	}

	req := &jsonrpc.Request{
		Method: "eth_getBlockByNumber",
		Params: []any{
			blockNumberStr,
			fullTxns,
		},
		Id:      1,
		JsonRpc: "2.0",
	}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return Block{}, err
	}

	response, err := b.request.MakeRequest(
		http.MethodPost,
		b.getApiUrl(),
		map[string]string{},
		string(reqBody),
	)
	if err != nil {
		return Block{}, err
	}

	if fullTxns {
		var respBody jsonrpc.Response[BaseBlockWithFullTxns]
		if err := json.Unmarshal(response, &respBody); err != nil {
			return Block{}, err
		}

		blockNumber, err = strconv.ParseInt(strings.TrimPrefix(respBody.Result.BlockNumber, "0x"), 16, 64)
		if err != nil {
			return Block{}, err
		}

		timestamp, err := strconv.ParseInt(strings.TrimPrefix(respBody.Result.Timestamp, "0x"), 16, 64)
		if err != nil {
			return Block{}, err
		}
		timestamp *= 1000

		block := Block{
			BlockNumber: blockNumber,
			BlockHash:   respBody.Result.BlockHash,
			Timestamp:   timestamp,
		}

		block.Txns = lo.Map(respBody.Result.Txns, func(txn BaseTxn, _ int) Txn {
			return Txn{
				BlockHash:   txn.BlockHash,
				BlockNumber: txn.BlockNumber,
				TxnHash:     txn.TxnHash,
				Type:        txn.Type,
				From:        txn.From,
				To:          txn.To,
				Value:       txn.Value,
			}
		})

		return block, nil
	}

	var respBody jsonrpc.Response[BaseBlockWithoutFullTxns]
	if err := json.Unmarshal(response, &respBody); err != nil {
		return Block{}, err
	}

	blockNumber, err = strconv.ParseInt(strings.TrimPrefix(respBody.Result.BlockNumber, "0x"), 16, 64)
	if err != nil {
		return Block{}, err
	}

	timestamp, err := strconv.ParseInt(strings.TrimPrefix(respBody.Result.Timestamp, "0x"), 16, 64)
	if err != nil {
		return Block{}, err
	}
	timestamp *= 1000

	block := Block{
		ChainId:     b.GetChainId(),
		BlockNumber: blockNumber,
		BlockHash:   respBody.Result.BlockHash,
		Timestamp:   timestamp,
	}

	block.TxnHashes = respBody.Result.Txns

	return block, nil
}

func (b *BaseChain) GetBlocks(ctx context.Context, fromBlockNumber int64, toBlockNumber int64, fullTxns bool, includeLogs bool, addresses []string, topics []string) ([]Block, error) {
	// use alchemy batch request to get the event in blocks
	// and the information of all the blocks in one request
	// https://docs.alchemy.com/reference/batch-requests
	// The purpose of getting block information is to populate the timestamp
	req := make([]jsonrpc.Request, 0)
	id := int64(0)
	if includeLogs {
		id++
		req = append(req, jsonrpc.Request{
			Method: "eth_getLogs",
			Params: []any{
				getLogsParam{
					FromBlockNumber: "0x" + strconv.FormatInt(fromBlockNumber, 16),
					ToBlockNumber:   "0x" + strconv.FormatInt(toBlockNumber, 16),
					Addresses:       addresses,
					Topics:          topics,
				},
			},
			Id:      id,
			JsonRpc: "2.0",
		})
	}
	for i := fromBlockNumber; i <= toBlockNumber; i++ {
		id++
		req = append(req, jsonrpc.Request{
			Method: "eth_getBlockByNumber",
			Params: []any{
				"0x" + strconv.FormatInt(i, 16),
				fullTxns,
			},
			Id:      id,
			JsonRpc: "2.0",
		})
	}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	response, err := b.request.MakeRequest(
		http.MethodPost,
		b.getApiUrl(),
		map[string]string{},
		string(reqBody),
	)
	if err != nil {
		return nil, err
	}

	var respBody []jsonrpc.Response[json.RawMessage]
	if err := json.Unmarshal(response, &respBody); err != nil {
		return nil, err
	}

	blocks := make([]Block, 0)
	var logs []Log
	for _, resp := range respBody {
		// Unmarshal as block first
		// because there will be multiple entries for blocks
		// and only one entry for logs
		if fullTxns {
			var baseBlock BaseBlockWithFullTxns
			err = json.Unmarshal(resp.Result, &baseBlock)
			if err == nil {
				blockNumber, err := strconv.ParseInt(strings.TrimPrefix(baseBlock.BlockNumber, "0x"), 16, 64)
				if err != nil {
					continue
				}
				blockTimestamp, err := strconv.ParseInt(strings.TrimPrefix(baseBlock.Timestamp, "0x"), 16, 64)
				if err != nil {
					continue
				}
				blockTimestamp *= 1000
				txns := lo.Map(baseBlock.Txns, func(txn BaseTxn, _ int) Txn {
					return Txn{
						BlockHash:   txn.BlockHash,
						BlockNumber: txn.BlockNumber,
						TxnHash:     txn.TxnHash,
						Type:        txn.Type,
						From:        txn.From,
						To:          txn.To,
						Value:       txn.Value,
					}
				})
				blocks = append(blocks, Block{
					ChainId:     b.GetChainId(),
					BlockNumber: blockNumber,
					BlockHash:   baseBlock.BlockHash,
					Timestamp:   blockTimestamp,
					Txns:        txns,
				})
				continue
			}
		} else {
			var baseBlock BaseBlockWithoutFullTxns
			err = json.Unmarshal(resp.Result, &baseBlock)
			if err == nil {
				blockNumber, err := strconv.ParseInt(strings.TrimPrefix(baseBlock.BlockNumber, "0x"), 16, 64)
				if err != nil {
					continue
				}
				blockTimestamp, err := strconv.ParseInt(strings.TrimPrefix(baseBlock.Timestamp, "0x"), 16, 64)
				if err != nil {
					continue
				}
				blockTimestamp *= 1000
				blocks = append(blocks, Block{
					ChainId:     b.GetChainId(),
					BlockNumber: blockNumber,
					BlockHash:   baseBlock.BlockHash,
					Timestamp:   blockTimestamp,
					TxnHashes:   baseBlock.Txns,
				})
				continue
			}
		}

		var baseLogs []BaseLog
		err = json.Unmarshal(resp.Result, &baseLogs)
		if err != nil {
			continue
		}
		logs = lo.Map(baseLogs, func(log BaseLog, _ int) Log {
			blockNumber, err := strconv.ParseInt(strings.TrimPrefix(log.BlockNumber, "0x"), 16, 64)
			if err != nil {
				return Log{}
			}
			logIndex, err := strconv.ParseInt(strings.TrimPrefix(log.LogIndex, "0x"), 16, 64)
			if err != nil {
				return Log{}
			}
			return Log{
				Address:     log.Address,
				BlockNumber: blockNumber,
				BlockHash:   log.BlockHash,
				Data:        log.Data,
				Topics:      log.Topics,
				TxnHash:     log.TxnHash,
				LogIndex:    logIndex,
				Removed:     log.Removed,
			}
		})
	}

	if includeLogs {
		logsByBlock := lo.GroupBy(logs, func(log Log) int64 {
			return log.BlockNumber
		})

		blocks = lo.Map(blocks, func(block Block, _ int) Block {
			if blockLogs, ok := logsByBlock[block.BlockNumber]; ok {
				sort.Slice(blockLogs, func(i, j int) bool {
					return blockLogs[i].LogIndex < blockLogs[j].LogIndex
				})
				block.Logs = blockLogs
			}
			return block
		})
	}

	return blocks, nil
}
