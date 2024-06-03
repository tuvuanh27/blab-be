package dto

import (
	"blockchain-backend/service"
	"fmt"
)

type MineBlockData struct {
	MinerAddress string `json:"miner_address" binding:"required"`
	Data         string `json:"data"`
	BlockNumber  int64  `json:"block_number"`
}

type HashData struct {
	Data      string `json:"data" binding:"required"`
	Algorithm string `json:"algorithm" binding:"required"`
}

type SetDifficultyData struct {
	Difficulty int64 `json:"difficulty" binding:"required"`
}

type GenesisBlockData struct {
	Nonce      int64 `json:"nonce" binding:"required"`
	Difficulty int64 `json:"difficulty" binding:"required"`
}

type NewBlockData struct {
	BlockNumber  int64                 `json:"block_number" binding:"required"`
	Hash         string                `json:"hash"`
	Binary       string                `json:"binary"`
	ParentHash   string                `json:"parent_hash" binding:"required"`
	Nonce        int64                 `json:"nonce"`
	Difficulty   int64                 `json:"difficulty"`
	Miner        string                `json:"miner" binding:"required"`
	Transactions []service.Transaction `json:"transactions"`
	Data         string                `json:"data"`
}

func (h *HashData) Validate() error {
	// validate algorithm must be one of SHA256, SHA512, Keccak256
	if h.Algorithm != "SHA256" && h.Algorithm != "SHA512" && h.Algorithm != "Keccak256" {
		return fmt.Errorf("algorithm must be one of SHA256, SHA512, Keccak256")
	}
	return nil
}

func (s *SetDifficultyData) Validate() error {
	// validate difficulty must be greater than 0
	if s.Difficulty <= 0 {
		return fmt.Errorf("difficulty must be greater than 0")
	}
	return nil
}
