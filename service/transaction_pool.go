package service

import (
	redisPkg "blockchain-backend/infras/redis"
	"encoding/json"
	"log"
)

type TxPoolConfigSource string

const (
	Mempool TxPoolConfigSource = "Mempool"
	Redis   TxPoolConfigSource = "Redis"
)

type ITransactionPoolService interface {
	Clear()
	SetTransaction(transaction *Transaction)
	GetTransactionPool() map[string]Transaction
	GetTransactions() []Transaction
	ConfigTransactionPool(sourceType TxPoolConfigSource)
	GetConfigTransactionPool() TxPoolConfigSource
}

type transactionPoolService struct {
	sourceType         TxPoolConfigSource
	transactionMap     map[string]Transaction
	transactionService ITransactionService
}

func NewTransactionPoolService(transactionService ITransactionService) ITransactionPoolService {
	return &transactionPoolService{
		transactionMap:     make(map[string]Transaction),
		transactionService: transactionService,
	}
}

func (tps *transactionPoolService) GetConfigTransactionPool() TxPoolConfigSource {
	return tps.sourceType
}

func (tps *transactionPoolService) ConfigTransactionPool(sourceType TxPoolConfigSource) {
	if sourceType == Mempool {
		tps.sourceType = Mempool
	}

	if sourceType == Redis {
		tps.sourceType = Redis
	}
}

func (tps *transactionPoolService) Clear() {
	tps.transactionMap = make(map[string]Transaction)
}

func (tps *transactionPoolService) SetTransaction(transaction *Transaction) {
	isExist := false
	if _, ok := tps.transactionMap[transaction.Hash]; ok {
		isExist = true
	}

	if !isExist {
		tps.transactionMap[transaction.Hash] = *transaction
	}
	transactions := make([]Transaction, 0)
	for _, tx := range tps.transactionMap {
		transactions = append(transactions, tx)
	}
	tsxBytes, _ := json.Marshal(transactions)

	redisPkg.RedisService.Set(redisPkg.TransactionPoolKey, string(tsxBytes))
}

func (tps *transactionPoolService) GetTransactionPool() map[string]Transaction {
	if tps.sourceType == Redis {
		txPool := redisPkg.RedisService.Get(redisPkg.TransactionPoolKey)
		var txs []Transaction
		if txPool != "" {
			err := json.Unmarshal([]byte(txPool), &txs)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return tps.transactionMap
}

func (tps *transactionPoolService) GetTransactions() []Transaction {
	transactions := make([]Transaction, 0)
	for _, transaction := range tps.GetTransactionPool() {
		transactions = append(transactions, transaction)
	}
	return transactions
}
