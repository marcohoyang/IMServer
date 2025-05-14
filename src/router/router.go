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

	r.Static("/asset", "asset/")
	r.LoadHTMLGlob("../view/*")

	r.GET("/", service.GetIndex)
	r.GET("/index", service.GetIndex)

	r.POST("/api/register", service.Register)
	r.POST("/api/login", service.Login)

	user := r.Group("/api/user")
	user.GET("/ws", utils.JWTAuthMiddlewareForWS(), service.UpgradeWebSocket)
	user.GET("/friends", utils.JWTAuthMiddlewareForWS(), service.GetFriends)
	user.POST("addfriend", utils.JWTAuthMiddlewareForWS(), service.AddFriend)
	user.Use(utils.JWTAuthMiddleware())
	{
		user.POST("/updateUser", service.UpdateUser)
		user.GET("/getUser", service.GetUser)
	}

	return r
}
