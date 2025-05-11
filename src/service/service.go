package service

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/hoyang/imserver/src/conveter"
	"github.com/hoyang/imserver/src/dbproxy/models"
	im "github.com/hoyang/imserver/src/proto"
	rpcClient "github.com/hoyang/imserver/src/rpc"
	"github.com/hoyang/imserver/src/utils"
	"github.com/redis/go-redis/v9"
)

type UserService struct {
	pool    *rpcClient.ClientPool
	redisDB *redis.Client
}

// NewUserService 构造函数
func NewUserService(pool *rpcClient.ClientPool, redisDB *redis.Client) *UserService {
	return &UserService{pool: pool, redisDB: redisDB}
}

// GetIndex
// @Summary ping example
// @Schemes
// @Description do ping
// @Tags example
// @Accept json
// @Produce json
// @Success 200 {string} Helloworld
// @Router /example/helloworld [get]
func (s *UserService) GetIndex(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

// Login
// @Summary 登录
// @Tags 用户模块
// @param name formData string false "用户名"
// @param password formData string false "密码"
// @Success 200 {string} userId
// @Router /login [post]
func (s *UserService) Login(c *gin.Context) {
	user := models.IMUser{}
	user.Name = c.PostForm("name")
	password := c.PostForm("password")
	dbUser, err := s.getUser(&user)
	if err != nil {
		c.JSON(400, gin.H{
			"message": "登录失败",
		})
		return
	}
	if !utils.VaildPassword(password, dbUser.Salt, dbUser.Password) {
		c.JSON(400, gin.H{
			"message": "登录失败",
		})
		return
	}
	token, err := utils.GenerateToken(dbUser.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": "生成Token失败"})
		return
	}
	c.JSON(200, gin.H{
		"message": "ok",
		"token":   token,
	})
}

// Register
// @Summary 注册用户
// @Tags 用户模块
// @param name formData string false "用户名"
// @param password formData string false "密码"
// @param repassword formData string false "确认密码"
// @param phone formData string false "手机号"
// @param email formData string false "邮箱"
// @Success 200 {string} userId
// @Router /register [post]
func (s *UserService) Register(c *gin.Context) {
	user := models.IMUser{}
	user.Name = c.PostForm("name")
	if user.Name == "" {
		c.JSON(400, gin.H{
			"message": "用户名不能为空",
		})
		return
	}
	password := c.PostForm("password")
	rePassword := c.PostForm("repassword")
	if password != rePassword {
		c.JSON(400, gin.H{
			"message": "两次输入的密码不同",
		})
		return
	}
	salt := fmt.Sprintf("%06d", rand.Int31())
	user.Password = utils.MakePassword(password, salt)
	user.Salt = salt
	user.Phone = c.PostForm("phone")
	email := c.PostForm("email")
	user.Email = &email

	err := s.createUser(&user)
	if err != nil {
		c.JSON(400, gin.H{
			"message": "注册失败",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "注册成功",
		"userid":  user.ID,
	})
}

// GetUser
// @Summary 查询用户
// @Tags 用户模块
// @param name query string false "用户名"
// @param Authorization header string true "Bearer token"
// @Security bearerAuth
// @Accept json
// @Produce json
// @Success 200 {string} ok
// @Router /user/getUser [get]
func (s *UserService) GetUser(c *gin.Context) {
	user := models.IMUser{}
	user.Name = c.Query("name")
	dbUser, err := s.getUser(&user)
	if err != nil {
		c.JSON(400, gin.H{
			"message": "查询失败",
		})
		return
	}
	c.JSON(200, gin.H{
		"userDate": dbUser,
	})
}

// UpdateUser
// @Summary 更新用户
// @Tags 用户模块
// @param id formData string false "Id"
// @param name formData string false "用户名"
// @param password formData string false "密码"
// @param phone formData string false "手机号"
// @param email formData string false "邮箱"
// @param Authorization header  string true "Bearer token"
// @Security bearerAuth
// @Success 200 {string} ok
// @Router /user/updateUser [post]
func (s *UserService) UpdateUser(c *gin.Context) {
	user := models.IMUser{}
	id, _ := strconv.Atoi(c.PostForm("id"))
	user.ID = uint(id)
	user.Name = c.PostForm("name")
	password := c.PostForm("password")
	user.Phone = c.PostForm("phone")
	email := c.PostForm("email")
	user.Email = &email

	if password != "" {
		salt := fmt.Sprintf("%06d", rand.Int31())
		user.Password = utils.MakePassword(password, salt)
		user.Salt = salt
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn := s.pool.Get()
	defer s.pool.Put(conn)
	client := im.NewUserServiceClient(conn)
	result, err := client.UpdateUser(ctx, conveter.ToPBIMUser(&user))
	if err != nil {
		fmt.Printf("CreateUser failed %v", err)
		c.JSON(400, gin.H{
			"message": "更新失败",
		})
		return
	}
	user = *conveter.ToDBIMUser(result)

	c.JSON(200, gin.H{
		"message": "ok",
		"userId":  user.ID,
	})
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// UpgradeWebSocket
// @Summary 升级websocket
// @Description Handle WebSocket upgrade request
// @Tags 用户模块
// @Accept json
// @Produce json
// @param Authorization header  string true "Bearer token"
// @Security bearerAuth
// @Router /user/upgradeWebSocket [get]
func (s *UserService) UpgradeWebSocket(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("升级websocket失败")
		c.JSON(400, gin.H{
			"mseeage": "升级ws失败",
		})
		return
	}

	fmt.Println("升级websocke成功")
	go s.handlerWebsocket(ws)
}

func (s *UserService) handlerWebsocket(ws *websocket.Conn) {
	defer ws.Close()
	for {
		messageType, message, err := ws.ReadMessage()
		if err != nil {
			return
		}

		fmt.Println(string(message))
		err = ws.WriteMessage(messageType, []byte("收到消息"))
		if err != nil {
			return
		}
	}
}

func (s *UserService) getUser(user *models.IMUser) (*models.IMUser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn := s.pool.Get()
	defer s.pool.Put(conn)
	client := im.NewUserServiceClient(conn)
	req := im.UserRequest{Name: user.Name}
	result, err := client.GetUser(ctx, &req)
	if err != nil {
		fmt.Printf("GetUser failed %v\n", err)
		return nil, err
	}

	return conveter.ToDBIMUser(result), nil
}

func (s *UserService) createUser(user *models.IMUser) error {
	fmt.Println("service call CreateUser")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn := s.pool.Get()
	defer s.pool.Put(conn)
	client := im.NewUserServiceClient(conn)
	result, err := client.CreateUser(ctx, conveter.ToPBIMUser(user))
	if err != nil {
		fmt.Printf("CreateUser failed %v", err)
		return err
	}
	*user = *conveter.ToDBIMUser(result)

	return nil
}
