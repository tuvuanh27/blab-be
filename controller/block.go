package controller

import (
	"blockchain-backend/controller/dto"
	"blockchain-backend/service"
	"blockchain-backend/util"
	"github.com/gin-gonic/gin"
	"strconv"
)

type IBlockController interface {
	SetupRoutes(group *gin.RouterGroup)
	getBlocks() func(c *gin.Context)
	getBlock() func(c *gin.Context)
	mine() func(c *gin.Context)
	replaceChain() func(c *gin.Context)
	hash() func(c *gin.Context)
	reset() func(c *gin.Context)
	setDifficulty() func(c *gin.Context)
	getDifficulty() func(c *gin.Context)
	newGenesisBlock() func(c *gin.Context)
	checkValidChain() func(c *gin.Context)
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
	group.POST("/hash", bc.hash()) // use-case 1
	group.POST("/reset", bc.reset())
	group.POST("/set-difficulty", bc.setDifficulty())
	group.GET("/get-difficulty", bc.getDifficulty())
	group.POST("/new-genesis-block", bc.newGenesisBlock())
	group.GET("/check-valid-chain", bc.checkValidChain())
}

func (bc *blockController) getDifficulty() func(c *gin.Context) {
	return func(c *gin.Context) {
		difficulty := bc.blockSvc.GetDifficulty()

		c.JSON(200, gin.H{
			"difficulty": difficulty,
		})
	}
}

func (bc *blockController) checkValidChain() func(c *gin.Context) {
	return func(c *gin.Context) {
		chain := bc.blockChainSvc.GetBlocks()
		ok, blockNumber := bc.blockChainSvc.IsValidChain(chain)

		c.JSON(200, gin.H{
			"is_valid":     ok,
			"block_number": blockNumber,
		})
	}
}

func (bc *blockController) newGenesisBlock() func(c *gin.Context) {
	return func(c *gin.Context) {
		var body *dto.GenesisBlockData

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		genesisBlock := bc.blockSvc.Genesis(body.Nonce, body.Difficulty)
		bc.blockChainSvc.ReplaceBlock(genesisBlock)

		c.JSON(200, gin.H{
			"message": "genesis block created successfully",
		})
	}
}

// @BasePath /block

// @Summary Set difficulty
// @Description Set difficulty
// @Tags block
// @Accept json
// @Produce json
// @Param difficulty body dto.SetDifficultyData true "Difficulty"
// @Success 201
// @Router /block/set-difficulty [post]
func (bc *blockController) setDifficulty() func(c *gin.Context) {
	return func(c *gin.Context) {
		var body *dto.SetDifficultyData

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		if err := body.Validate(); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		bc.blockSvc.SetDifficulty(body.Difficulty)

		c.JSON(200, gin.H{
			"message": "difficulty set successfully",
			"data":    bc.blockSvc.GetDifficulty(),
		})
	}

}

func (bc *blockController) reset() func(c *gin.Context) {
	return func(c *gin.Context) {
		bc.blockChainSvc.Reset()
		c.JSON(200, gin.H{
			"message": "blockchain reset successfully",
		})
	}
}

// @BasePath /block

// @Summary Get blocks
// @Description Get blocks
// @Tags block
// @Accept json
// @Produce json
// @Success 200
// @Router /block [get]
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
		body := dto.MineBlockData{}

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		position := body.BlockNumber
		if position == 0 {
			position = -1
		} else {
			position = body.BlockNumber
		}

		rewardTransaction := bc.transactionSvc.RewardTransaction(body.MinerAddress)
		bc.transactionPoolSvc.SetTransaction(rewardTransaction)

		newBlock, err := bc.blockChainSvc.NewBlock(body.Data, body.MinerAddress, position)
		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

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

		if ok, _ := bc.blockChainSvc.IsValidChain(body); ok {
			err := bc.blockChainSvc.ReplaceChain(body)
			if err != nil {
				c.JSON(400, gin.H{
					"error": err.Error(),
				})
				return
			}
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

func (bc *blockController) hash() func(c *gin.Context) {
	return func(c *gin.Context) {

		var body *dto.HashData
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		if err := body.Validate(); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		hash := util.Hash([]byte(body.Data), util.HashAlgorithm(body.Algorithm))

		c.JSON(200, gin.H{
			"data": hash,
		})
	}
}
