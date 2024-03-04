package controller

import (
	"blockchain-backend/service"
	"github.com/gin-gonic/gin"
)

type IWalletController interface {
	SetupRoutes(group *gin.RouterGroup)
	generateKeyPair() func(c *gin.Context)
	getBalance() func(c *gin.Context)
}

type walletController struct {
	walletSvc service.IWalletService
}

func NewWalletController(walletService service.IWalletService) IWalletController {
	return &walletController{
		walletSvc: walletService,
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
