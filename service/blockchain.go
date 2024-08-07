package service

import (
	redisPkg "blockchain-backend/infras/redis"
	"blockchain-backend/util"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"log"
	"strconv"
	"strings"
)

type State string

const (
	Valid   State = "VALID"
	Invalid State = "INVALID"
)

type Chain struct {
	Blocks []Block `json:"blocks"`
	//State            State   `json:"state"`
	//BlockNumberValid int64   `json:"block_number_valid"`
}

type IBlockchainService interface {
	Reset()
	GetBlocks() Chain
	GetBlock(blockNumber int64) (Block, error)
	//AddBlock(block Block)
	IsValidChain(chain Chain) (bool, int64)
	IsValidTransactionData(chain Chain) bool
	ReplaceChain(chain Chain) error
	BlockLength() int
	GetTransactionHistory(address string) []Transaction
	GetTransaction(transactionHash string) (Transaction, error)
	//SyncNode(pubsub *redis.PubSub)
	ReplaceBlock(block *Block)
	NewBlock(data, miner string, position int64) (*Block, error)
}

type blockchainService struct {
	chain                  Chain
	blockService           IBlockService
	transactionPoolService ITransactionPoolService
}

func NewBlockchainService(blockService IBlockService, transactionPoolService ITransactionPoolService, chain Chain) IBlockchainService {
	if len(chain.Blocks) == 0 {
		chain = Chain{
			Blocks: []Block{},
			//State:            Valid,
			//BlockNumberValid: 0,
		}
		blockChainBytes, _ := json.Marshal(chain)
		redisPkg.RedisService.Set(redisPkg.ChainKey, string(blockChainBytes))
	}

	return &blockchainService{
		chain:                  chain,
		blockService:           blockService,
		transactionPoolService: transactionPoolService,
	}
}

func (bls *blockchainService) NewBlock(data, miner string, position int64) (*Block, error) {
	var lastBlockNumber int64 = position - 2
	if position == -1 {
		lastBlockNumber = int64(len(bls.chain.Blocks) - 1)
	}
	if position == 1 {
		lastBlockNumber = 0
	}
	lastBlock := bls.chain.Blocks[lastBlockNumber]
	block, err := bls.blockService.NewBlock(lastBlock, bls.transactionPoolService.GetTransactions(), data, miner, position)
	if err != nil {
		return nil, err
	}

	bls.ReplaceBlock(block)
	return block, nil
}

func (bls *blockchainService) ReplaceBlock(block *Block) {
	blockNumber := block.BlockNumber
	log.Println("blockNumber: ", block)
	if blockNumber == 1 && len(bls.chain.Blocks) == 0 {
		bls.chain.Blocks = []Block{*block}
	}
	if blockNumber > int64(len(bls.chain.Blocks)) {
		bls.chain.Blocks = append(bls.chain.Blocks, *block)
	} else {
		for i, b := range bls.chain.Blocks {
			if b.BlockNumber == blockNumber {
				bls.chain.Blocks[i] = *block
				break
			}
		}

	}
}

func (bls *blockchainService) Reset() {
	bls.chain = Chain{
		Blocks: []Block{*bls.blockService.Genesis(0, bls.blockService.GetDifficulty())},
	}

	blockChainBytes, _ := json.Marshal(bls.chain)
	redisPkg.RedisService.Set(redisPkg.ChainKey, string(blockChainBytes))

	redisPkg.RedisService.Publish(redisPkg.ChannelSyncNodeKey, string(blockChainBytes))

	// clear transaction pool
	bls.transactionPoolService.Clear()
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

//func (bls *blockchainService) AddBlock(block Block) {
//	bls.chain.Blocks = append(bls.chain.Blocks, block)
//}

func (bls *blockchainService) IsValidChain(chain Chain) (bool, int64) {
	//if !reflect.DeepEqual(chain.Blocks[0], bls.blockService.Genesis()) {
	//	log.Println("Genesis block is invalid")
	//	return false
	//}

	for i := 1; i < len(chain.Blocks); i++ {
		lastHash := chain.Blocks[i-1].Hash
		//lastDifficulty := chain.Blocks[i-1].Difficulty
		if lastHash != chain.Blocks[i].ParentHash {
			return false, chain.Blocks[i].BlockNumber
		}

		block := chain.Blocks[i]

		transaction, _ := json.Marshal(block.Transactions)

		validHash := util.CryptoHash([]byte(strconv.FormatInt(block.BlockNumber, 10) + lastHash + string(rune(block.Nonce)) + strconv.FormatInt(block.Difficulty, 10) + strconv.FormatInt(block.Timestamp, 10) + block.Miner + string(transaction) + block.Data)).Hex()
		if strings.Compare(validHash, chain.Blocks[i].Hash) != 0 {
			log.Println("Invalid hash at block ", i)
			return false, chain.Blocks[i].BlockNumber
		}

		//if math.Abs(float64(lastDifficulty-difficulty)) > 1 {
		//	log.Println("Invalid difficulty at block ", i)
		//	return false, chain.Blocks[i].BlockNumber
		//}
	}

	return true, int64(len(chain.Blocks))
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

func (bls *blockchainService) ReplaceChain(chain Chain) error {
	if len(chain.Blocks) <= len(bls.chain.Blocks) {
		return fmt.Errorf("received chain is not longer than the current chain")
	}

	if isValid, _ := bls.IsValidChain(chain); !isValid {
		return fmt.Errorf("received chain is invalid")
	}
	bls.chain = chain

	blockChainBytes, _ := json.Marshal(chain)
	redisPkg.RedisService.Set(redisPkg.ChainKey, string(blockChainBytes))

	//redisPkg.RedisService.Publish(redisPkg.ChannelSyncNodeKey, string(blockChainBytes))
	return nil
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

func (bls *blockchainService) GetTransaction(transactionHash string) (Transaction, error) {
	for _, block := range bls.chain.Blocks {
		for _, transaction := range block.Transactions {
			if strings.Compare(transaction.Hash, transactionHash) == 0 {
				return transaction, nil
			}
		}
	}
	return Transaction{}, fmt.Errorf("transaction not found")
}

//func (bls *blockchainService) SyncNode(pubsub *redis.PubSub) {
//	defer func(pubsub *redis.PubSub) {
//		err := pubsub.Close()
//		if err != nil {
//			log.Fatalln(err)
//		}
//	}(pubsub)
//
//	for {
//		msg, err := pubsub.ReceiveMessage(redisPkg.Ctx)
//		if err != nil {
//			log.Fatalln(err)
//		}
//		log.Println("Received message: ", msg.Payload, " from channel: ", msg.Channel)
//
//		if strings.Compare(msg.Channel, redisPkg.ChannelSyncNodeKey) == 0 {
//
//			var chain Chain
//			err = json.Unmarshal([]byte(msg.Payload), &chain)
//			if err != nil {
//				log.Fatalln(err)
//			}
//
//			err = bls.ReplaceChain(chain)
//			if err != nil {
//				log.Fatalln(err)
//			}
//		}
//
//		if strings.Compare(msg.Channel, redisPkg.ChannelSyncTransactionKey) == 0 {
//
//			var transaction Transaction
//			err = json.Unmarshal([]byte(msg.Payload), &transaction)
//			if err != nil {
//				log.Fatalln(err)
//			}
//
//			bls.transactionPoolService.SetTransaction(transaction)
//		}
//	}
//
//}
