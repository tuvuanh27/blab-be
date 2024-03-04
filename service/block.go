package service

import (
	"blockchain-backend/infras/redis"
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
}

type IBlockService interface {
	Genesis() Block
	adjustDifficulty(originalBlock Block, timestamp int64) int64
	MineBlock(lastBlock Block, transactions []Transaction, miner string) (Block, error)
}

type blockService struct {
	initDifficulty     int64
	mineRate           int64
	transactionPoolSvc ITransactionPoolService
}

func NewBlockService(transactionPoolSvc ITransactionPoolService) IBlockService {
	return &blockService{
		initDifficulty:     10,
		mineRate:           10000,
		transactionPoolSvc: transactionPoolSvc,
	}
}

func (bs *blockService) Genesis() Block {
	return Block{
		BlockNumber:  1,
		Hash:         "0x",
		ParentHash:   "0x",
		Nonce:        0,
		Difficulty:   bs.initDifficulty,
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

func (bs *blockService) MineBlock(lastBlock Block, transactions []Transaction, miner string) (Block, error) {
	if len(transactions) == 0 {
		return Block{}, fmt.Errorf("no transactions to mine")
	}

	lastHash := lastBlock.Hash
	blockNumber := lastBlock.BlockNumber + 1
	difficulty := lastBlock.Difficulty
	nonce := 0
	var timestamp int64 = 0
	var blockHash string = "0x"
	var binary string
	data, _ := json.Marshal(transactions)

	for {
		nonce += 1
		timestamp = time.Now().Unix() // time in seconds
		difficulty = bs.adjustDifficulty(lastBlock, timestamp)

		blockHash = (util.CryptoHash([]byte(strconv.FormatInt(blockNumber, 10) + lastHash + string(rune(nonce)) + strconv.FormatInt(difficulty, 10) + strconv.FormatInt(timestamp, 10) + miner + string(data)))).Hex()

		binary, _ = util.HexToBin(blockHash)

		if binary[:difficulty] == strings.Repeat("0", int(difficulty)) {
			break
		}
		nonce++

	}

	bs.transactionPoolSvc.Clear()

	// sync to redis
	chain := Chain{}
	blockChain := redis.RedisService.Get(redis.ChainKey)
	if blockChain != "" {
		err := json.Unmarshal([]byte(blockChain), &chain)
		if err != nil {
			return Block{}, err
		}
	}

	newBlock := Block{
		BlockNumber:  blockNumber,
		Hash:         blockHash,
		Binary:       binary,
		ParentHash:   lastHash,
		Nonce:        int64(nonce),
		Difficulty:   difficulty,
		Timestamp:    timestamp,
		Miner:        miner,
		Transactions: transactions,
	}

	chain.Blocks = append(chain.Blocks, newBlock)
	blockChainBytes, _ := json.Marshal(chain)
	redis.RedisService.Set(redis.ChainKey, string(blockChainBytes))

	redis.RedisService.Publish(redis.ChannelSyncNodeKey, string(blockChainBytes))

	return newBlock, nil
}
