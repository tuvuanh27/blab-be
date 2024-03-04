package service

import (
	"blockchain-backend/util"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"strconv"
	"strings"
)

type Transaction struct {
	Hash      string `json:"hash"`
	Signature string `json:"signature"`
	From      string `json:"from"`
	To        string `json:"to"`
	Value     int64  `json:"value"`
	Data      string `json:"data"`
	Timestamp int64  `json:"timestamp"`
}

type ITransactionService interface {
	ValidTransaction(transaction Transaction, pubKey string) bool
	TxHash(transaction Transaction) string
	RewardTransaction(miner string) Transaction
	CreateTransaction(from string, to string, value int64, data string, timestamp int64, signature string, pubKey string) (Transaction, error)
}

type transactionService struct {
}

func NewTransactionService() ITransactionService {
	return &transactionService{}
}

func (ts *transactionService) TxHash(transaction Transaction) string {
	return util.CryptoHash([]byte(transaction.From + transaction.To + strconv.FormatInt(transaction.Value, 10) + transaction.Data + strconv.FormatInt(transaction.Timestamp, 10))).Hex()
}

func (ts *transactionService) ValidTransaction(transaction Transaction, pubKey string) bool {
	// check reward transaction
	if strings.Compare(transaction.From, common.Address{}.Hex()) == 0 {
		return true
	}

	if transaction.From == "" {
		return false
	}

	if transaction.To == "" {
		return false
	}

	if transaction.Value <= 0 {
		return false
	}

	if transaction.Data == "" {
		return false
	}

	if transaction.Timestamp <= 0 {
		return false
	}

	hashBytes := util.CryptoHash([]byte(transaction.From + transaction.To + strconv.FormatInt(transaction.Value, 10) + transaction.Data + strconv.FormatInt(transaction.Timestamp, 10))).Bytes()

	if !util.VerifySignature(
		pubKey,
		hashBytes,
		transaction.Signature,
	) {
		return false
	}

	return true
}

func (ts *transactionService) RewardTransaction(miner string) Transaction {
	transaction := Transaction{
		From:      common.Address{}.Hex(),
		To:        miner,
		Value:     util.MinersReward,
		Data:      "",
		Timestamp: 0,
	}

	transaction.Hash = ts.TxHash(transaction)

	return transaction
}

func (ts *transactionService) CreateTransaction(from string, to string, value int64, data string, timestamp int64, signature string, pubKey string) (Transaction, error) {
	transaction := Transaction{
		From:      from,
		To:        to,
		Value:     value,
		Data:      data,
		Timestamp: timestamp,
		Signature: signature,
	}

	if !ts.ValidTransaction(transaction, pubKey) {
		return Transaction{}, fmt.Errorf("invalid transaction")
	}

	transaction.Hash = ts.TxHash(transaction)

	return transaction, nil
}
