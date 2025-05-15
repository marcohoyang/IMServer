package router

import (
	"os"
	"path/filepath"

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
	// 获取可执行文件所在目录
	exe, _ := os.Executable()
	dir := filepath.Dir(exe)

	// 构建模板路径（比如 ../view）
	templatePath := filepath.Join(dir, "..", "view", "*")
	r.LoadHTMLGlob(templatePath)

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
