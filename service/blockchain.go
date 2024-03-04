package service

import (
	"blockchain-backend/util"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/redis/go-redis/v9"
	"log"
	"math"
	"reflect"
	"strconv"
	"strings"
)

type Chain struct {
	Blocks []Block `json:"blocks"`
}

type IBlockchainService interface {
	GetBlocks() Chain
	GetBlock(blockNumber int64) (Block, error)
	AddBlock(block Block)
	IsValidChain(chain Chain) bool
	IsValidTransactionData(chain Chain) bool
	ReplaceChain(chain Chain, validTransactions []Transaction, onSuccess bool)
	BlockLength() int
	GetTransactionHistory(address string) []Transaction
	SyncNode(pubsub *redis.PubSub)
}

type blockchainService struct {
	chain        Chain
	blockService IBlockService
}

func NewBlockchainService(blockService IBlockService, chain Chain) IBlockchainService {
	if len(chain.Blocks) == 0 {
		chain.Blocks = append(chain.Blocks, blockService.Genesis())
	}

	return &blockchainService{
		chain:        chain,
		blockService: blockService,
	}
}

func (bls *blockchainService) GetBlocks() Chain {
	return bls.chain
}

func (bls *blockchainService) GetBlock(blockNumber int64) (Block, error) {
	for _, block := range bls.chain.Blocks {
		if block.BlockNumber == blockNumber {
			return block, nil
		}
	}

	return Block{}, fmt.Errorf("block not found")
}

func (bls *blockchainService) AddBlock(block Block) {
	bls.chain.Blocks = append(bls.chain.Blocks, block)
}

func (bls *blockchainService) IsValidChain(chain Chain) bool {
	if reflect.DeepEqual(chain.Blocks[0], bls.blockService.Genesis()) {
		return false
	}

	for i := 1; i < len(chain.Blocks); i++ {
		lastHash := chain.Blocks[i-1].Hash
		lastDifficulty := chain.Blocks[i-1].Difficulty
		if lastHash != chain.Blocks[i].ParentHash {
			return false
		}

		var blockNumber = chain.Blocks[i].BlockNumber
		var difficulty = chain.Blocks[i].Difficulty
		var nonce = chain.Blocks[i].Nonce
		var timestamp = chain.Blocks[i].Timestamp
		var miner = chain.Blocks[i].Miner
		var data, _ = json.Marshal(chain.Blocks[i].Transactions)

		validHash := util.CryptoHash([]byte(strconv.FormatInt(blockNumber, 10) + lastHash + string(rune(nonce)) + strconv.FormatInt(difficulty, 10) + strconv.FormatInt(timestamp, 10) + miner + string(data)))
		if strings.Compare(validHash.Hex(), chain.Blocks[i].Hash) != 0 {
			return false
		}

		if math.Abs(float64(lastDifficulty-difficulty)) > 1 {
			return false
		}
	}

	return true
}

func (bls *blockchainService) IsValidTransactionData(chain Chain) bool {
	for i := 1; i < len(chain.Blocks); i++ {
		rewardTransactionCount := 0
		for _, transaction := range chain.Blocks[i].Transactions {
			if strings.Compare(transaction.From, common.Address{}.Hex()) == 0 {
				rewardTransactionCount += 1
				if rewardTransactionCount > 1 {
					log.Fatalln("Miner rewards exceed limit")
					return false
				}

				if transaction.Value != util.MinersReward {
					log.Fatalln("Miner reward amount is invalid")
					return false
				}
			} else {

			}

		}
	}

	return true
}

// ReplaceChain TODO: Implement this function, call when subcribe to new block
func (bls *blockchainService) ReplaceChain(chain Chain, validTransactions []Transaction, onSuccess bool) {
	if len(chain.Blocks) <= len(bls.chain.Blocks) {
		log.Fatalln("Received chain is not longer than the current chain")
		return
	}

	if !bls.IsValidChain(chain) {
		log.Fatalln("Received chain is invalid")
		return
	}

}

func (bls *blockchainService) BlockLength() int {
	return len(bls.chain.Blocks)
}

func (bls *blockchainService) GetTransactionHistory(address string) []Transaction {
	var transactions []Transaction
	for _, block := range bls.chain.Blocks {
		for _, transaction := range block.Transactions {
			if strings.Compare(transaction.From, address) == 0 {
				transactions = append(transactions, transaction)
			}
		}
	}

	return transactions
}

func (bls *blockchainService) SyncNode(pubsub *redis.PubSub) {

}
