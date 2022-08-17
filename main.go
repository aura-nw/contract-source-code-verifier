package main

import (
	"smart-contract-verify/cloud"
	docs "smart-contract-verify/docs" // docs is generated by Swag CLI, you have to import it.
	"smart-contract-verify/repository"
	"smart-contract-verify/util"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Verify Smart Contract API
// @version 1.0
// @description This is a smart contract verify application
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email soberkoder@gmail.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host verify.serenity.aurascan.io
// @BasePath /
func main() {
	util.DownloadAllRustOptimizerImages()
	util.DownloadAllWorkspaceOptimizerImages()

	smartContractRepo := repository.New()
	router := gin.Default()

	session := cloud.ConnectS3()

	router.Use(func(c *gin.Context) {
		c.Set("session", session)
		c.Next()
	})

	docs.SwaggerInfo.BasePath = "/api/v1"
	v1 := router.Group("/api/v1")
	{
		eg := v1.Group("/smart-contract")
		{
			eg.POST("/verify", smartContractRepo.CallVerifyContractCode)
			eg.GET("/get-hash/:contractId", smartContractRepo.CallGetContractHash)
		}
	}

	url := ginSwagger.URL("https://verify.serenity.aurascan.io/swagger/doc.json") // The url pointing to API definition
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler, url))
	router.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
