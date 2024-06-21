package controller

import (
	"blockchain-backend/controller/dto"
	"blockchain-backend/service"
	"blockchain-backend/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
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
	group.POST("/", wc.generateKeyPair())
	group.GET("/balance/:address", wc.getBalance())
}

func (wc *walletController) generateKeyPair() func(c *gin.Context) {
	return func(c *gin.Context) {
		var body *dto.GenerateWalletData
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		// get initBalance from query
		initBalance := c.Query("initBalance")
		var balanceValue int64
		/// check if initBalance is not empty
		if initBalance != "" {
			balanceValue, _ = strconv.ParseInt(initBalance, 10, 64)
		} else {
			balanceValue = 1000
		}

		keyPair, err := wc.walletSvc.GenerateKeyPair(body.SeedPhrase)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		data := util.CryptoHash([]byte(common.Address{}.Hex() + keyPair.Address + strconv.FormatInt(balanceValue, 10) + "" + strconv.FormatInt(time.Now().Unix(), 10))).Bytes()
		signature, _ := util.Sign(data, keyPair.PrivateKey)

		// add transaction send 1000 to keyPair.Address
		transaction := &service.Transaction{
			From:      common.Address{}.Hex(),
			To:        keyPair.Address,
			Value:     balanceValue,
			Data:      "",
			Timestamp: time.Now().Unix(),
			Signature: signature,
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
