package swag

import (
	_ "github.com/eiblog/eiblog/cmd/eiblog/docs" // docs

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// RegisterRoutes register routes
func RegisterRoutes(group gin.IRoutes) {
	group.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
