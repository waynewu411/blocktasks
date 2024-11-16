package do

type Task struct {
	Name                        string `json:"name" gorm:"column:name;primaryKey"`
	LastProcessedBlockNumber    int64  `json:"last_processed_block_number" gorm:"column:last_processed_block_number"`
	LastProcessedBlockTimestamp int64  `json:"last_processed_block_timestamp" gorm:"column:last_processed_block_timestamp"`
}

func (t *Task) TableName() string {
	return "Tasks"
}
