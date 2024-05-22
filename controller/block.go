package controller

import (
	"blockchain-backend/controller/dto"
	"blockchain-backend/service"
	"github.com/gin-gonic/gin"
	"strconv"
)

type IBlockController interface {
	SetupRoutes(group *gin.RouterGroup)
	getBlocks() func(c *gin.Context)
	getBlock() func(c *gin.Context)
	mine() func(c *gin.Context)
	replaceChain() func(c *gin.Context)
}

type blockController struct {
	blockSvc           service.IBlockService
	blockChainSvc      service.IBlockchainService
	transactionPoolSvc service.ITransactionPoolService
	transactionSvc     service.ITransactionService
}

func NewBlockController(blockSvc service.IBlockService, blockChainSvc service.IBlockchainService, transactionPoolSvc service.ITransactionPoolService, transactionSvc service.ITransactionService) IBlockController {
	return &blockController{
		blockSvc:           blockSvc,
		blockChainSvc:      blockChainSvc,
		transactionPoolSvc: transactionPoolSvc,
		transactionSvc:     transactionSvc,
	}
}

func (bc *blockController) SetupRoutes(group *gin.RouterGroup) {
	group.GET("/", bc.getBlocks())
	group.GET("/:blockNumber", bc.getBlock())
	group.POST("/mine", bc.mine())
	group.POST("/replace-chain", bc.replaceChain())

}

func (bc *blockController) getBlocks() func(c *gin.Context) {
	return func(c *gin.Context) {

		c.JSON(200, gin.H{
			"data": bc.blockChainSvc.GetBlocks(),
		})
	}
}

func (bc *blockController) getBlock() func(c *gin.Context) {
	return func(c *gin.Context) {

		blockNumber, err := strconv.ParseInt(c.Param("blockNumber"), 10, 64)
		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		block, err := bc.blockChainSvc.GetBlock(blockNumber)
		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"data": block,
		})
	}
}

func (bc *blockController) mine() func(c *gin.Context) {
	return func(c *gin.Context) {
		body := dto.MineBlockQuery{}

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		blockLength := bc.blockChainSvc.BlockLength()
		lastBlock, err := bc.blockChainSvc.GetBlock(int64(blockLength))
		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}
		transactions := bc.transactionPoolSvc.GetTransactions()
		if len(transactions) == 0 {
			c.JSON(400, gin.H{
				"error": "no transactions to mine",
			})
			return
		}
		rewardTransaction := bc.transactionSvc.RewardTransaction(body.MinerAddress)
		transactions = append(transactions, rewardTransaction)
		newBlock, err := bc.blockSvc.MineBlock(lastBlock, transactions, body.MinerAddress)
		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}
		bc.blockChainSvc.AddBlock(newBlock)

		c.JSON(200, gin.H{
			"message": "mined block successfully",
			"data":    newBlock,
		})
	}
}

func (bc *blockController) replaceChain() func(c *gin.Context) {
	return func(c *gin.Context) {
		body := service.Chain{}

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		if bc.blockChainSvc.IsValidChain(body) {
			bc.blockChainSvc.ReplaceChain(body)
			bc.transactionPoolSvc.Clear()
			c.JSON(200, gin.H{
				"message": "chain replaced successfully",
			})
			return
		}

		c.JSON(400, gin.H{
			"error": "invalid chain",
		})
	}
}
