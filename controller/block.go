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
}

type blockController struct {
	blockSvc           service.IBlockService
	blockChainSvc      service.IBlockchainService
	transactionPoolSvc service.ITransactionPoolService
}

func NewBlockController(blockSvc service.IBlockService, blockChainSvc service.IBlockchainService, transactionPoolSvc service.ITransactionPoolService) IBlockController {
	return &blockController{
		blockSvc:           blockSvc,
		blockChainSvc:      blockChainSvc,
		transactionPoolSvc: transactionPoolSvc,
	}
}

func (bc *blockController) SetupRoutes(group *gin.RouterGroup) {
	group.GET("/", bc.getBlocks())
	group.GET("/:blockNumber", bc.getBlock())
	group.POST("/mine", bc.mine())
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
