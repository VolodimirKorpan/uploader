package main

import (
	"github.com/VolodimirKorpan/uploader/controllers"
	"github.com/VolodimirKorpan/uploader/store"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20 // 8Mb

	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"admin": "admin",
	}))

	authorized.POST("/upload", controllers.Upload)
	authorized.GET("/download/:id", controllers.DownloadFile)

	store.Connect()

	r.Run(":8080")
}
