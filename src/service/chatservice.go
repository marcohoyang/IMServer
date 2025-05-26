package service

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/hoyang/imserver/src/models"
	rpcClient "github.com/hoyang/imserver/src/rpc"
	"github.com/hoyang/imserver/src/utils"
	"github.com/redis/go-redis/v9"
)

type Node struct {
	Conn      *websocket.Conn
	DataQueue chan models.Message
	wg        sync.WaitGroup
}

func CreateNode(c *websocket.Conn) *Node {
	var node Node
	queueSize := 10
	node.DataQueue = make(chan models.Message, queueSize)
	node.Conn = c
	return &node
}

type ChatService struct {
	clientMap map[uint64]*Node
	rwLocker  sync.RWMutex
	redisDB   *redis.Client
	pool      *rpcClient.ClientPool
}

func NewChatService(redisDB *redis.Client, pool *rpcClient.ClientPool) *ChatService {
	s := &ChatService{redisDB: redisDB, pool: pool}
	s.clientMap = make(map[uint64]*Node, 10)
	return s
}

func (s *ChatService) Subscription() {
	ctx := context.Background()
	go func() {
		for {
			msg, err := utils.Subscription(s.redisDB, ctx, "msgChannel")
			if err != nil {
				log.Printf("recive msg err %v", err)
				continue
			}
			// 根据targetId转发消息到对应的user node，可能会导致消息顺序错误
			go func() {
				log.Println("Subscription revice:", msg)
				message, _ := models.MessageFromString(msg)
				log.Println("targetId,", message.ToID)
				s.rwLocker.RLock()
				node := s.clientMap[message.ToID]
				s.rwLocker.RUnlock()
				if node != nil {
					node.DataQueue <- message
				}
			}()
		}
	}()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *ChatService) Chat(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("升级websocket失败")
		c.JSON(400, gin.H{
			"mseeage": "升级ws失败",
		})
		return
	}
	defer conn.Close()
	node := CreateNode(conn)
	node.Conn = conn
	userId, exist := c.Get("user_id")
	if !exist {
		log.Println("升级websocket失败, userid不存在")
		c.JSON(400, gin.H{
			"mseeage": "升级ws失败",
		})
		return
	}
	s.rwLocker.Lock()
	s.clientMap[userId.(uint64)] = node
	s.rwLocker.Unlock()
	defer delete(s.clientMap, userId.(uint64))

	log.Println("升级websocke成功")
	response := map[string]interface{}{
		"action":  "switchToChat",
		"message": "WebSocket 连接成功",
	}

	// 发送消息给客户端
	err = node.Conn.WriteJSON(response)
	if err != nil {
		log.Println("发送消息失败:", err)
		return
	}
	s.handlerWebsocket(node, c)

	node.wg.Wait()
	log.Println("handlerWebsocket msg eixt, userID:", userId)
	// TODO: 更新user status to offline
}

// 设置心跳参数
const (
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

func (s *ChatService) handlerWebsocket(node *Node, c *gin.Context) {
	closeNotify := make(chan any)
	var closeOnce sync.Once
	closeFunc := func() {
		closeOnce.Do(func() {
			close(closeNotify)
		})
	}
	//订阅redis消息
	node.wg.Add(1)
	go func() {
		defer node.wg.Done()
		for {
			select {
			case <-closeNotify:
				return
			case msg := <-node.DataQueue:
				content, err := json.Marshal(&msg)
				if err != nil {
					log.Println("解析失败", err)
				}
				err = node.Conn.WriteMessage(websocket.TextMessage, content)
				if err != nil {
					log.Printf("WriteMessage err %v", err)
					closeFunc()
					return
				}
			}
		}
	}()
	node.wg.Add(1)
	//将客户端消息publish到redis
	go func() {
		defer node.wg.Done()

		for {
			select {
			case <-closeNotify:
				return
			default:
				_, message, err := node.Conn.ReadMessage()
				if err != nil {
					log.Printf("ReadMessage err %v", err)
					closeFunc()
					return
				}
				log.Println("receive message:", string(message))

				var msg models.Message
				err = json.Unmarshal(message, &msg)
				if err != nil {
					log.Println("解析错误:", err)
					return
				}

				utils.Publish(s.redisDB, c, "msgChannel", msg.String())
				conn := s.pool.Get()
				defer s.pool.Put(conn)
				rpcClient.NewMessageProxy(conn).StoreMessage(msg.FromID, msg.ToID, msg.Type, msg.Content)
			}
		}
	}()

	// 启动心跳机制
	node.wg.Add(1)
	go func() {
		defer node.wg.Done()
		pingTicker := time.NewTicker(pingPeriod)
		defer pingTicker.Stop()

		for {
			select {
			case <-pingTicker.C:
				if err := node.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second)); err != nil {
					log.Printf("WebSocket ping 发送失败: %v", err)
					closeFunc() // 通知其他 goroutine 关闭
					return
				}
			case <-closeNotify:
				return
			}
		}
	}()
}
