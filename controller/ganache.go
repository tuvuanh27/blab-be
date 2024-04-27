package controller

import (
	"blockchain-backend/infras/redis"

	"github.com/gin-gonic/gin"
)

type IGanacheController interface {
	SetupRoutes(group *gin.RouterGroup)
	GetHistory() func(c *gin.Context)
}

type ganacheController struct {
}

func NewGanacheController() IGanacheController {
	return &ganacheController{}
}

func (gc *ganacheController) SetupRoutes(group *gin.RouterGroup) {
	group.GET("/history", gc.GetHistory())
}

func (gc *ganacheController) GetHistory() func(c *gin.Context) {

	return func(c *gin.Context) {

		address := c.Query("address")
		if address == "" {
			c.JSON(400, gin.H{
				"error": "Address is required",
			})
			return
		}

		transactions := redis.RedisService.GetSet(redis.HistoryTransactionsKey + address)

		c.JSON(200, gin.H{
			"message": "Get history",
			"data":    transactions,
		})
	}
}
