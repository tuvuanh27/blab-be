package dto

type MineBlockQuery struct {
	MinerAddress string `json:"miner_address" binding:"required"`
}
