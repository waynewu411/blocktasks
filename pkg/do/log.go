package do

import "github.com/lib/pq"

type Log struct {
	ChainId     int64          `json:"chain_id" gorm:"column:chain_id;primaryKey"`
	BlockNumber int64          `json:"block_number" gorm:"column:block_number"`
	BlockHash   string         `json:"block_hash" gorm:"column:block_hash"`
	Data        string         `json:"data" gorm:"column:data"`
	Topics      pq.StringArray `json:"topics" gorm:"column:topics;type:text[]"`
	TxnHash     string         `json:"txn_hash" gorm:"column:txn_hash;primaryKey"`
	LogIndex    int64          `json:"log_index" gorm:"column:log_index;primaryKey"`
	Removed     bool           `json:"removed" gorm:"column:removed"`
	Timestamp   int64          `json:"timestamp" gorm:"column:timestamp"`
}

func (l *Log) TableName() string {
	return "Logs"
}
