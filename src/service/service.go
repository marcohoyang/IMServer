package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hoyang/imserver/src/conveter"
	"github.com/hoyang/imserver/src/models"
	im "github.com/hoyang/imserver/src/proto"
	rpcClient "github.com/hoyang/imserver/src/rpc"

	"github.com/hoyang/imserver/src/utils"
	"github.com/redis/go-redis/v9"
)

type UserService struct {
	pool        *rpcClient.ClientPool
	redisDB     *redis.Client
	chatService *ChatService
}

// NewUserService 构造函数
func NewUserService(pool *rpcClient.ClientPool, redisDB *redis.Client) *UserService {
	chatService := NewChatService(redisDB)
	chatService.Subscription()
	return &UserService{pool: pool, redisDB: redisDB, chatService: chatService}
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
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "登录页面",
	})
}

func (s *UserService) GetChatHtml(c *gin.Context) {
	c.HTML(http.StatusOK, "chat1.html", gin.H{
		"title": "聊天页面",
	})
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login
// @Summary 登录
// @Tags 用户模块
// @param username formData string false "用户名"
// @param password formData string false "密码"
// @Success 200 {string} userId
// @Router /login [post]
func (s *UserService) Login(c *gin.Context) {
	var loginRequest LoginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}
	user := models.IMUser{}
	user.Name = loginRequest.Username
	password := loginRequest.Password
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

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: false, // 防止XSS攻击
		Secure:   false, // 开发环境设为false，生产环境设为true
	})
	c.JSON(200, gin.H{
		"message": "ok",
		"userID":  dbUser.ID,
	})
}

func (s *UserService) GetFriends(c *gin.Context) {
	userId, exist := c.Get("user_id")
	if !exist {
		log.Println("GetFriends失败, userid不存在")
		c.JSON(400, gin.H{
			"mseeage": "GetFriends失败",
		})
		return
	}
	friends, err := s.getFriends(userId.(uint))
	if err != nil {
		log.Println("GetFriends失败")
		c.JSON(400, gin.H{
			"mseeage": "GetFriends失败",
		})
		return
	}
	prin, _ := json.Marshal(friends)
	log.Println(string(prin))
	c.JSON(200, friends)
}

type AddFriendReq struct {
	Username       string `json:"username"`
	UserID         uint   `json:"userID"`
	FriendUsername string `json:"friendUsername"`
}

func (s *UserService) AddFriend(c *gin.Context) {
	var addFriendReq AddFriendReq
	if err := c.ShouldBindJSON(&addFriendReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	user := models.IMUser{}
	user.Name = addFriendReq.FriendUsername
	friend, err := s.getUser(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "用户不存在",
		})
		return
	}
	var userShips models.Contact
	userShips.OwnerId = addFriendReq.UserID
	userShips.TargetId = friend.ID
	err = s.addFriend(&userShips)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "添加好友失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "添加好友成功",
	})
}

func (s *UserService) addFriend(userShips *models.Contact) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn := s.pool.Get()
	defer s.pool.Put(conn)
	client := im.NewUserServiceClient(conn)
	contact := im.Contact{}
	contact.OwnerId = uint64(userShips.OwnerId)
	contact.TargetId = uint64(userShips.TargetId)
	_, err := client.AddFriend(ctx, &contact)
	if err != nil {
		log.Printf("addFriend failed %v\n", err)
		return err
	}

	return nil
}

// Register
// @Summary 注册用户
// @Tags 用户模块
// @param username formData string false "用户名"
// @param password formData string false "密码"
// @param repassword formData string false "确认密码"
// @param phone formData string false "手机号"
// @param email formData string false "邮箱"
// @Success 200 {string} userId
// @Router /register [post]
func (s *UserService) Register(c *gin.Context) {
	var loginRequest LoginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}
	user := models.IMUser{}
	user.Name = loginRequest.Username
	if user.Name == "" {
		c.JSON(400, gin.H{
			"message": "用户名不能为空",
		})
		return
	}
	password := loginRequest.Password
	salt := fmt.Sprintf("%06d", rand.Int31())
	user.Password = utils.MakePassword(password, salt)
	user.Salt = salt
	Phone := c.PostForm("phone")
	user.Phone = &Phone
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
// @param username formData string false "用户名"
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
	user.Name = c.PostForm("username")
	password := c.PostForm("password")
	Phone := c.PostForm("phone")
	user.Phone = &Phone
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
		log.Printf("CreateUser failed %v", err)
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
	s.chatService.Chat(c)
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
		log.Printf("GetUser failed %v\n", err)
		return nil, err
	}

	return conveter.ToDBIMUser(result), nil
}

func (s *UserService) createUser(user *models.IMUser) error {
	log.Println("service call CreateUser")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn := s.pool.Get()
	defer s.pool.Put(conn)
	client := im.NewUserServiceClient(conn)
	result, err := client.CreateUser(ctx, conveter.ToPBIMUser(user))
	if err != nil {
		log.Printf("CreateUser failed %v", err)
		return err
	}
	*user = *conveter.ToDBIMUser(result)

	return nil
}

func (s *UserService) getFriends(id uint) ([]models.FriendView, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn := s.pool.Get()
	defer s.pool.Put(conn)
	client := im.NewUserServiceClient(conn)
	req := im.UserRequest{Id: uint64(id)}
	result, err := client.GetFriends(ctx, &req)
	if err != nil {
		log.Printf("GetFriends failed %v", err)
		return []models.FriendView{}, err
	}
	friendViews := conveter.ProtosToFriendViews(result)
	return friendViews, nil
}
