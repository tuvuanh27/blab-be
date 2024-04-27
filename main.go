package main

import (
	"blockchain-backend/config"
	"blockchain-backend/controller"
	"blockchain-backend/infras/redis"
	"blockchain-backend/service"
	"encoding/json"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"log"
	"time"
)

var timestampInSeconds validator.Func = func(fl validator.FieldLevel) bool {
	// validate that the timestamp is in seconds
	timestamp := fl.Field().Int()
	return timestamp > 0 && timestamp < time.Now().Unix()
}

func main() {

	port := config.ConfigEnv.Port
	engine := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("timestampInSeconds", timestampInSeconds)
		if err != nil {
			return
		}
	}

	engine.ForwardedByClientIP = true
	err := engine.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		return
	}

	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	engine.GET("/", func(c *gin.Context) {

		c.JSON(200, gin.H{
			"message": "Welcome to the blab blockchain",
		})
	})

	// sync block
	var chain = service.Chain{}
	blockChain := redis.RedisService.Get(redis.ChainKey)
	if blockChain != "" {
		err = json.Unmarshal([]byte(blockChain), &chain)
		if err != nil {
			log.Fatal(err)
		}
	}

	transactionSvc := service.NewTransactionService()
	transactionPoolSvc := service.NewTransactionPoolService(transactionSvc)
	blockSvc := service.NewBlockService(transactionPoolSvc)
	blockChainSvc := service.NewBlockchainService(blockSvc, transactionPoolSvc, chain)
	walletSvc := service.NewWalletService(blockChainSvc)
	ganacheSvc := service.NewGanacheService()

	// sync node
	go func() {
		blockChainSvc.SyncNode(redis.RedisService.Subscribe(redis.ChannelSyncNodeKey))
	}()

	// cron crawl block
	go func() {
		err := ganacheSvc.CrawlBlock()
		if err != nil {
			return
		}
	}()

	walletController := controller.NewWalletController(walletSvc, transactionPoolSvc)
	transactionController := controller.NewTransactionController(transactionSvc, transactionPoolSvc, blockChainSvc, walletSvc)
	blockController := controller.NewBlockController(blockSvc, blockChainSvc, transactionPoolSvc, transactionSvc)
	ganacheController := controller.NewGanacheController()

	walletGroup := engine.Group("/wallet")
	transactionGroup := engine.Group("/transaction")
	blockGroup := engine.Group("/block")
	ganacheGroup := engine.Group("/ganache")

	walletController.SetupRoutes(walletGroup)
	transactionController.SetupRoutes(transactionGroup)
	blockController.SetupRoutes(blockGroup)
	ganacheController.SetupRoutes(ganacheGroup)

	if err := engine.Run(
		":" + port,
	); err != nil {
		log.Fatalln(err)
	}

	log.Println("Server is running on port " + port + "...")

}
