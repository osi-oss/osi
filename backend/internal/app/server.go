package server

import (
	"github.com/gin-gonic/gin"
	"github.com/osi-oss/osi/internal/config"
	"github.com/osi-oss/osi/internal/db"
)

func Start(cfg *config.Config) {

	DB := db.Connect(cfg)
	db.SyncDb(DB)

	r := gin.Default()
	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(
			200, gin.H{
				"message": "pong",
			})
	})

	r.Run()

}
