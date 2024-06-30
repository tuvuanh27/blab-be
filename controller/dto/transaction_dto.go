package dto

import "fmt"

type SignTransactionRequest struct {
	PrivateKey string `json:"private_key" binding:"required"`
	From       string `json:"from" binding:"required"`
	To         string `json:"to" binding:"required"`
	Value      int64  `json:"value" binding:"required"`
	Data       string `json:"data" binding:"required"`
	Timestamp  int64  `json:"timestamp" binding:"required,timestampInSeconds"`
}

type CreateTransactionRequest struct {
	From      string `json:"from" binding:"required"`
	To        string `json:"to" binding:"required"`
	Value     int64  `json:"value" binding:"required"`
	Data      string `json:"data" binding:"required"`
	Timestamp int64  `json:"timestamp" binding:"required,timestampInSeconds"`
	Signature string `json:"signature" binding:"required"`
	PublicKey string `json:"public_key" binding:"required"`
}

type VerifySignatureData struct {
	Signature string `json:"signature" binding:"required"`
	TxHash    string `json:"tx_hash" binding:"required"`
	PublicKey string `json:"public_key" binding:"required"`
}

type ConfigTransactionPoolData struct {
	Type string `json:"type" binding:"required"`
}

func (c *CreateTransactionRequest) Validate() error {
	if c.Value <= 0 {
		return fmt.Errorf("value must be greater than 0")
	}

	if c.From == c.To {
		return fmt.Errorf("from and to must be different")
	}

	return nil
}

func (c *ConfigTransactionPoolData) Validate() error {
	// validate type must be Mempool or Redis
	if c.Type != "Mempool" && c.Type != "Redis" {
		return fmt.Errorf("type must be Mempool or Redis")
	}

	return nil
}
