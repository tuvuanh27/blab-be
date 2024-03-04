package service

import (
	"blockchain-backend/util"
	"encoding/hex"
)

type IWalletService interface {
	GenerateKeyPair() (util.KeyPair, error)
	SignTransaction(tx Transaction, privateKey string) (string, error)
	CalculateBalance(address string) int64
}

type walletService struct {
	blockChainSvc IBlockchainService
}

func NewWalletService(blockChainSvc IBlockchainService) IWalletService {
	return &walletService{
		blockChainSvc: blockChainSvc,
	}
}

func (ws *walletService) GenerateKeyPair() (util.KeyPair, error) {
	return util.GenerateKeyPair()
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
