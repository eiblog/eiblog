// Package swag provides ...
package swag

import (
	_ "github.com/eiblog/eiblog/pkg/core/backup/docs" // docs

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// RegisterRoutes register routes
func RegisterRoutes(group gin.IRoutes) {
	group.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
