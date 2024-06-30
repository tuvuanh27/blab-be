package service

import (
	"blockchain-backend/util"
	"encoding/hex"
)

type IWalletService interface {
	GenerateKeyPair(seedPhrase string) (util.KeyPair, error)
	SignTransaction(tx Transaction, privateKey string) (string, error)
	CalculateBalance(address string) int64
	CalculateAllBalances() map[string]int64
}

type walletService struct {
	blockChainSvc IBlockchainService
}

func NewWalletService(blockChainSvc IBlockchainService) IWalletService {
	return &walletService{
		blockChainSvc: blockChainSvc,
	}
}

func (ws *walletService) CalculateAllBalances() map[string]int64 {
	balances := make(map[string]int64)

	chain := ws.blockChainSvc.GetBlocks()

	for _, block := range chain.Blocks {
		for _, transaction := range block.Transactions {
			if _, ok := balances[transaction.From]; !ok {
				balances[transaction.From] = 0
			}

			if _, ok := balances[transaction.To]; !ok {
				balances[transaction.To] = 0
			}

			balances[transaction.From] -= transaction.Value
			balances[transaction.To] += transaction.Value
		}
	}

	// remove address 0x0000000000000000000000000000000000000000
	delete(balances, "0x0000000000000000000000000000000000000000")

	return balances
}

func (ws *walletService) GenerateKeyPair(seedPhrase string) (util.KeyPair, error) {
	return util.GenerateKeyPair(seedPhrase)
}

func (ws *walletService) SignTransaction(tx Transaction, privateKey string) (string, error) {
	data, err := hex.DecodeString(tx.Hash)
	if err != nil {
		return "", err
	}

	return util.Sign(data, privateKey)
}

func (ws *walletService) CalculateBalance(address string) int64 {
	var balance int64 = 0
	chain := ws.blockChainSvc.GetBlocks()

	for _, block := range chain.Blocks {
		for _, transaction := range block.Transactions {
			if transaction.From == address {
				balance -= transaction.Value
			}

			if transaction.To == address {
				balance += transaction.Value
			}
		}

	}

	return balance
}
