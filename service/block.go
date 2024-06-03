package service

import (
	redisPkg "blockchain-backend/infras/redis"
	"blockchain-backend/util"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Block struct {
	BlockNumber  int64         `json:"block_number"`
	Hash         string        `json:"hash"`
	Binary       string        `json:"binary"`
	ParentHash   string        `json:"parent_hash"`
	Nonce        int64         `json:"nonce"`
	Difficulty   int64         `json:"difficulty"`
	Timestamp    int64         `json:"timestamp"`
	Miner        string        `json:"miner"`
	Transactions []Transaction `json:"transactions"`
	Data         string        `json:"data"`
}

type IBlockService interface {
	Genesis(nonce, difficulty int64) *Block
	adjustDifficulty(originalBlock Block, timestamp int64) int64
	NewBlock(lastBlock Block, transactions []Transaction, data, miner string, position int64) (*Block, error)
	SetDifficulty(difficulty int64)
	GetDifficulty() int64
	HashBlock(block *Block, lastHash string) string
}

type blockService struct {
	difficulty         int64
	mineRate           int64
	transactionPoolSvc ITransactionPoolService
}

func NewBlockService(transactionPoolSvc ITransactionPoolService) IBlockService {
	return &blockService{
		difficulty:         10,
		mineRate:           10000,
		transactionPoolSvc: transactionPoolSvc,
	}
}

func (bs *blockService) SetDifficulty(difficulty int64) {
	bs.difficulty = difficulty
	redisPkg.RedisService.Set(redisPkg.DifficultyKey, strconv.FormatInt(difficulty, 10))
}

func (bs *blockService) GetDifficulty() int64 {
	return bs.difficulty
}

func (bs *blockService) HashBlock(block *Block, lastHash string) string {
	transaction, _ := json.Marshal(block.Transactions)

	return util.CryptoHash([]byte(strconv.FormatInt(block.BlockNumber, 10) + lastHash + string(rune(block.Nonce)) + strconv.FormatInt(block.Difficulty, 10) + strconv.FormatInt(block.Timestamp, 10) + block.Miner + string(transaction) + block.Data)).Hex()
}

func (bs *blockService) Genesis(nonce, difficulty int64) *Block {
	return &Block{
		BlockNumber:  1,
		Hash:         "0x",
		ParentHash:   "0x",
		Nonce:        nonce,
		Difficulty:   difficulty,
		Timestamp:    0,
		Miner:        "0x",
		Transactions: []Transaction{},
	}
}

func (bs *blockService) adjustDifficulty(originalBlock Block, timestamp int64) int64 {
	difficulty := originalBlock.Difficulty
	if difficulty < 1 {
		return 1
	}

	if timestamp-originalBlock.Timestamp > bs.mineRate {
		return difficulty - 1
	}

	return difficulty + 1
}

// NewBlock creates a new block, if position is -1, the block will be mined with all transactions
func (bs *blockService) NewBlock(lastBlock Block, transactions []Transaction, data, miner string, position int64) (*Block, error) {
	if len(transactions) == 0 {
		return nil, fmt.Errorf("no transactions to mine")
	}

	if position == -1 {
		position = lastBlock.BlockNumber + 1
	}

	lastHash := lastBlock.Hash
	blockNumber := position
	difficulty := bs.difficulty
	nonce := 0
	var timestamp int64 = 0
	var blockHash string = "0x"
	var binary string

	newBlock := &Block{
		BlockNumber:  blockNumber,
		Hash:         blockHash,
		Binary:       binary,
		ParentHash:   lastHash,
		Nonce:        int64(nonce),
		Difficulty:   difficulty,
		Timestamp:    timestamp,
		Miner:        miner,
		Transactions: transactions,
		Data:         data,
	}

	for {
		nonce += 1
		timestamp = time.Now().Unix() // time in seconds
		//difficulty = bs.adjustDifficulty(lastBlock, timestamp)

		blockHash = bs.HashBlock(newBlock, lastHash)

		binary, _ = util.HexToBin(blockHash)

		if binary[:difficulty] == strings.Repeat("0", int(difficulty)) {
			break
		}
		nonce++

	}
	newBlock.Hash = blockHash
	newBlock.Nonce = int64(nonce)
	newBlock.Timestamp = timestamp
	newBlock.Binary = binary

	bs.transactionPoolSvc.Clear()

	// sync to redis
	//chain := Chain{}
	//blockChain := redis.RedisService.Get(redis.ChainKey)
	//if blockChain != "" {
	//	err := json.Unmarshal([]byte(blockChain), &chain)
	//	if err != nil {
	//		return Block{}, err
	//	}
	//}
	//
	//chain.Blocks = append(chain.Blocks, *newBlock)
	//blockChainBytes, _ := json.Marshal(chain)
	//redis.RedisService.Set(redis.ChainKey, string(blockChainBytes))

	//redis.RedisService.Publish(redis.ChannelSyncNodeKey, string(blockChainBytes))

	return newBlock, nil
}
