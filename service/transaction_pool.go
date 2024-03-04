package service

type ITransactionPoolService interface {
	Clear()
	SetTransaction(transaction Transaction)
	GetTransactionPool() map[string]Transaction
	GetTransactions() []Transaction
}

type transactionPoolService struct {
	transactionMap     map[string]Transaction
	transactionService ITransactionService
}

func NewTransactionPoolService(transactionService ITransactionService) ITransactionPoolService {
	return &transactionPoolService{
		transactionMap:     make(map[string]Transaction),
		transactionService: transactionService,
	}
}

func (tps *transactionPoolService) Clear() {
	tps.transactionMap = make(map[string]Transaction)
}

func (tps *transactionPoolService) SetTransaction(transaction Transaction) {
	isExist := false
	if _, ok := tps.transactionMap[transaction.Hash]; ok {
		isExist = true
	}

	if !isExist {
		tps.transactionMap[transaction.Hash] = transaction
	}
}

func (tps *transactionPoolService) GetTransactionPool() map[string]Transaction {
	return tps.transactionMap
}

func (tps *transactionPoolService) GetTransactions() []Transaction {
	transactions := make([]Transaction, 0)
	for _, transaction := range tps.transactionMap {
		transactions = append(transactions, transaction)
	}
	return transactions
}
