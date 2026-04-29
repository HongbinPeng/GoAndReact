package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket 配置
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 消息类型
const (
	MessageTypeChat   = "chat"   // 普通聊天消息
	MessageTypeJoin   = "join"   // 用户加入
	MessageTypeLeave  = "leave"  // 用户离开
	MessageTypeOnline = "online" // 在线列表
)

// 消息结构体
type Message struct {
	Type      string    `json:"type"`      // 消息类型
	Username  string    `json:"username"`  // 用户名
	Content   string    `json:"content"`   // 消息内容
	Timestamp time.Time `json:"timestamp"` // 时间戳
	Online    []string  `json:"online"`    // 在线用户列表
}

// 客户端管理
type ClientManager struct {
	clients   map[*websocket.Conn]string // 连接 -> 用户名
	broadcast chan Message               // 广播通道
	mu        sync.RWMutex               // 互斥锁
}

var manager = ClientManager{
	clients:   make(map[*websocket.Conn]string),
	broadcast: make(chan Message),
}

// 广播消息到所有客户端
func (cm *ClientManager) broadcastMessage(msg Message) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for conn := range cm.clients {
		err := conn.WriteJSON(msg)
		if err != nil {
			log.Println("发送失败:", err)
			conn.Close()
			delete(cm.clients, conn)
		}
	}
}

// 获取在线用户列表
func (cm *ClientManager) getOnlineUsers() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var users []string
	for _, username := range cm.clients {
		users = append(users, username)
	}
	return users
}

// WebSocket 处理函数
func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("升级失败:", err)
		return
	}
	defer conn.Close()

	// 第一步：获取用户名（从 URL 参数或消息中获取）
	username := r.URL.Query().Get("username")
	if username == "" {
		// 如果 URL 没有用户名，等待客户端发送
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("获取用户名失败:", err)
			return
		}
		username = msg.Username
	}

	// 注册客户端
	manager.mu.Lock()
	manager.clients[conn] = username
	manager.mu.Unlock()

	log.Printf("用户 [%s] 加入，当前在线: %d", username, len(manager.clients))

	// 发送用户加入通知
	joinMsg := Message{
		Type:      MessageTypeJoin,
		Username:  username,
		Content:   "加入了聊天室",
		Timestamp: time.Now(),
		Online:    manager.getOnlineUsers(),
	}
	manager.broadcastMessage(joinMsg)

	// 发送当前在线列表给新用户
	onlineMsg := Message{
		Type:      MessageTypeOnline,
		Online:    manager.getOnlineUsers(),
		Timestamp: time.Now(),
	}
	conn.WriteJSON(onlineMsg)

	// 循环读取消息
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("用户 [%s] 离开", username)

			// 发送用户离开通知
			manager.mu.Lock()
			delete(manager.clients, conn)
			manager.mu.Unlock()

			leaveMsg := Message{
				Type:      MessageTypeLeave,
				Username:  username,
				Content:   "离开了聊天室",
				Timestamp: time.Now(),
				Online:    manager.getOnlineUsers(),
			}
			manager.broadcastMessage(leaveMsg)
			break
		}

		// 设置消息属性
		msg.Type = MessageTypeChat
		msg.Timestamp = time.Now()

		log.Printf("[%s] %s: %s", msg.Timestamp.Format("15:04:05"), msg.Username, msg.Content)
		manager.broadcastMessage(msg)
	}
}

// 提供静态页面服务
func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func main() {
	// 注册路由
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", wsHandler)

	log.Println("群聊服务器启动在 :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
