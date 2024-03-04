package dto

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
