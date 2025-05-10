package router

import (
	"github.com/gin-gonic/gin"
	docs "github.com/hoyang/imserver/src/docs"
	"github.com/hoyang/imserver/src/service"
	"github.com/hoyang/imserver/src/utils"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Router(service *service.UserService) *gin.Engine {
	r := gin.Default()
	docs.SwaggerInfo.BasePath = ""
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	r.GET("/index", service.GetIndex)


	r.POST("/register", service.Register)
	r.POST("/login", service.Login)
	user := r.Group("/user")
	user.Use(utils.JWTAuthMiddleware())
	{
		user.POST("/updateUser", service.UpdateUser)
		user.GET("/getUser", service.GetUser)
		user.GET("/ws", service.SwapToWebSocket)
	}

	return r
}
