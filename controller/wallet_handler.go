package controller

import (
	"blockchain-backend/service"
	"blockchain-backend/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"strconv"
)

type IWalletController interface {
	SetupRoutes(group *gin.RouterGroup)
	generateKeyPair() func(c *gin.Context)
	getBalance() func(c *gin.Context)
}

type walletController struct {
	walletSvc          service.IWalletService
	transactionPoolSvc service.ITransactionPoolService
}

func NewWalletController(walletService service.IWalletService, transactionPoolSvc service.ITransactionPoolService) IWalletController {
	return &walletController{
		walletSvc:          walletService,
		transactionPoolSvc: transactionPoolSvc,
	}
}

func (wc *walletController) SetupRoutes(group *gin.RouterGroup) {
	group.GET("/", wc.generateKeyPair())
	group.GET("/balance/:address", wc.getBalance())
}

func (wc *walletController) generateKeyPair() func(c *gin.Context) {
	return func(c *gin.Context) {
		keyPair, err := wc.walletSvc.GenerateKeyPair()
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		// add transaction send 1000 to keyPair.Address
		transaction := service.Transaction{
			From:      common.Address{}.Hex(),
			To:        keyPair.Address,
			Value:     1000,
			Data:      "",
			Timestamp: 0,
		}

		transaction.Hash = util.CryptoHash([]byte(transaction.From + transaction.To + strconv.FormatInt(transaction.Value, 10) + transaction.Data + strconv.FormatInt(transaction.Timestamp, 10))).Hex()
		wc.transactionPoolSvc.SetTransaction(transaction)
		c.JSON(200, gin.H{
			"data": keyPair,
		})
	}
}

func (wc *walletController) getBalance() func(c *gin.Context) {
	return func(c *gin.Context) {
		address := c.Param("address")
		if address == "" {
			c.JSON(400, gin.H{
				"error": "address is required",
			})
			return
		}

		balance := wc.walletSvc.CalculateBalance(address)
		c.JSON(200, gin.H{
			"data": balance,
		})
	}
}
