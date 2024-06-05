package controller

import (
	"blockchain-backend/controller/dto"
	"blockchain-backend/service"
	"blockchain-backend/util"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gin-gonic/gin"
	"strconv"
)

type ITransactionController interface {
	SetupRoutes(group *gin.RouterGroup)
	signTransaction() func(c *gin.Context)
	createTransaction() func(c *gin.Context)
	getTransactionPool() func(c *gin.Context)
	getTransactionHistory() func(c *gin.Context)
	getTransaction() func(c *gin.Context)
	verifySignature() func(c *gin.Context)
	configTransactionPool() func(c *gin.Context)
	getConfigTransactionPool() func(c *gin.Context)
}

type transactionController struct {
	transactionSvc     service.ITransactionService
	transactionPoolSvc service.ITransactionPoolService
	blockchainService  service.IBlockchainService
	walletSvc          service.IWalletService
}

func NewTransactionController(transactionSvc service.ITransactionService, transactionPoolSvc service.ITransactionPoolService, blockchainService service.IBlockchainService, walletSvc service.IWalletService) ITransactionController {
	return &transactionController{
		transactionSvc:     transactionSvc,
		transactionPoolSvc: transactionPoolSvc,
		blockchainService:  blockchainService,
		walletSvc:          walletSvc,
	}
}

func (tc *transactionController) SetupRoutes(group *gin.RouterGroup) {
	group.POST("/", tc.createTransaction())
	group.POST("/sign", tc.signTransaction())
	group.GET("/pool", tc.getTransactionPool())
	group.GET("/history/:address", tc.getTransactionHistory())
	group.GET("/:hash", tc.getTransaction())
	group.POST("/verify", tc.verifySignature())
	group.POST("/config-tx-pool", tc.configTransactionPool())
	group.GET("/get-config-tx-pool", tc.getConfigTransactionPool())
}

func (tc *transactionController) getConfigTransactionPool() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"data": tc.transactionPoolSvc.GetConfigTransactionPool(),
		})
	}

}

func (tc *transactionController) configTransactionPool() func(c *gin.Context) {
	return func(c *gin.Context) {
		var body *dto.ConfigTransactionPoolData
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

		tc.transactionPoolSvc.ConfigTransactionPool(service.TxPoolConfigSource(body.Type))

		c.JSON(200, gin.H{
			"message": "Transaction pool configured",
		})
	}
}

func (tc *transactionController) verifySignature() func(c *gin.Context) {
	return func(c *gin.Context) {
		var body dto.VerifySignatureData

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		hash, err := hexutil.Decode(body.TxHash)
		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		ok := util.VerifySignature(body.PublicKey, hash, body.Signature)

		c.JSON(200, gin.H{
			"data": ok,
		})
	}
}

func (tc *transactionController) signTransaction() func(c *gin.Context) {
	return func(c *gin.Context) {
		var body dto.SignTransactionRequest

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		if body.Value > tc.walletSvc.CalculateBalance(body.From) {
			c.JSON(400, gin.H{
				"error": "Insufficient balance",
			})
			return
		}

		data := util.CryptoHash([]byte(body.From + body.To + strconv.FormatInt(body.Value, 10) + body.Data + strconv.FormatInt(body.Timestamp, 10))).Bytes()

		signature, err := util.Sign(data, body.PrivateKey)
		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"data": signature,
		})
	}
}

func (tc *transactionController) createTransaction() func(c *gin.Context) {
	return func(c *gin.Context) {
		var body dto.CreateTransactionRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		transaction, err := tc.transactionSvc.CreateTransaction(body.From, body.To, body.Value, body.Data, body.Timestamp, body.Signature, body.PublicKey)
		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		tc.transactionPoolSvc.SetTransaction(transaction)

		c.JSON(200, gin.H{
			"data": transaction,
		})
	}
}

func (tc *transactionController) getTransactionPool() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"data": tc.transactionPoolSvc.GetTransactionPool(),
		})
	}
}

func (tc *transactionController) getTransactionHistory() func(c *gin.Context) {
	return func(c *gin.Context) {
		address := c.Param("address")
		if address == "" {
			c.JSON(400, gin.H{
				"error": "address is required",
			})
			return
		}
		c.JSON(200, gin.H{
			"data": tc.blockchainService.GetTransactionHistory(address),
		})
	}
}

func (tc *transactionController) getTransaction() func(c *gin.Context) {
	return func(c *gin.Context) {
		txHash := c.Param("hash")
		if txHash == "" {
			c.JSON(400, gin.H{
				"error": "transaction hash is required",
			})
			return
		}

		transaction, err := tc.blockchainService.GetTransaction(txHash)
		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"data": transaction,
		})
	}
}
