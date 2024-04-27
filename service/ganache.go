package service

import (
	"blockchain-backend/config"
	"blockchain-backend/infras/redis"
	"context"
	"log"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/robfig/cron/v3"
)

type IGanacheService interface {
	CrawlBlock() error
	currentBlock() uint64
}

type GanacheService struct {
	rpc string
	r   redis.IRedis
}

func NewGanacheService() IGanacheService {
	return &GanacheService{
		rpc: config.ConfigEnv.Rpc,
		r:   redis.RedisService,
	}
}

func (gs *GanacheService) currentBlock() uint64 {
	currentBlock := gs.r.Get(redis.CurrentBlockCrawledKey)
	if currentBlock != "" {
		block, err := strconv.Atoi(currentBlock)
		if err != nil {
			log.Println(err)
			return 0
		}
		return uint64(block)
	} else {
		return 0
	}
}

func (gs *GanacheService) CrawlBlock() error {
	c := cron.New(
		cron.WithParser(
			cron.NewParser(
				cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)))

	client, _ := ethclient.Dial(gs.rpc)

	_, err := c.AddFunc("*/1 * * * *", func() {
		crawledBlock := gs.currentBlock()

		// get block from ganache
		latestBlock, err := client.BlockNumber(context.Background())
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Crawled block", crawledBlock, "to", latestBlock)
		if crawledBlock >= latestBlock {
			return
		}

		for i := crawledBlock + 1; i <= latestBlock; i++ {
			// get transactions from block
			blockData, err := client.BlockByNumber(context.Background(), big.NewInt(int64(i)))
			if err != nil {
				log.Println(err)
				return
			}

			// transactions from block
			for _, tx := range blockData.Transactions() {
				from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
				if err != nil {
					log.Println(err)
					return
				}
				log.Println("Block", i, "Tx", tx.Hash().Hex(), "From", from.Hex(), "To", tx.To().Hex(), "Value", tx.Value().String())
				// key = HistoryTransactionsKey + From
				key := redis.HistoryTransactionsKey + from.Hex()

				gs.r.SetSet(key, tx.Hash().Hex())
			}

			// save block to redis
			gs.r.Set(redis.CurrentBlockCrawledKey, strconv.Itoa(int(i)))

		}
	})
	if err != nil {
		return err
	}

	c.Start()

	return nil
}
