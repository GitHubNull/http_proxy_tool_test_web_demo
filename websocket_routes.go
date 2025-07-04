package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocket升级器
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
}

// WebSocket消息类型
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
	ID        string      `json:"id"`
}

// WebSocket连接管理器
type WebSocketManager struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.RWMutex
}

var wsManager = &WebSocketManager{
	clients:    make(map[*websocket.Conn]bool),
	broadcast:  make(chan []byte),
	register:   make(chan *websocket.Conn),
	unregister: make(chan *websocket.Conn),
}

// 初始化WebSocket路由
func initWebSocketRoutes(r *gin.Engine) {
	// 启动WebSocket管理器
	go wsManager.run()

	ws := r.Group("/ws")
	{
		// 基础WebSocket连接
		ws.GET("/connect", handleWebSocketConnect)

		// 回声WebSocket
		ws.GET("/echo", handleWebSocketEcho)

		// 广播WebSocket
		ws.GET("/broadcast", handleWebSocketBroadcast)

		// 实时数据推送
		ws.GET("/realtime", handleWebSocketRealtime)

		// 心跳检测
		ws.GET("/heartbeat", handleWebSocketHeartbeat)

		// 二进制数据传输
		ws.GET("/binary", handleWebSocketBinary)

		// 聊天室
		ws.GET("/chat", handleWebSocketChat)

		// 性能测试
		ws.GET("/performance", handleWebSocketPerformance)
	}

	// 广播API
	r.POST("/api/broadcast", handleBroadcastAPI)
}

// WebSocket管理器运行
func (manager *WebSocketManager) run() {
	for {
		select {
		case client := <-manager.register:
			manager.mutex.Lock()
			manager.clients[client] = true
			manager.mutex.Unlock()
			log.Printf("WebSocket客户端连接: %v", client.RemoteAddr())

			// 发送欢迎消息
			welcomeMsg := WebSocketMessage{
				Type:      "welcome",
				Data:      "欢迎连接WebSocket服务器",
				Timestamp: time.Now().Unix(),
				ID:        generateRequestID(),
			}
			manager.sendToClient(client, welcomeMsg)

		case client := <-manager.unregister:
			manager.mutex.Lock()
			if _, ok := manager.clients[client]; ok {
				delete(manager.clients, client)
				client.Close()
				log.Printf("WebSocket客户端断开: %v", client.RemoteAddr())
			}
			manager.mutex.Unlock()

		case message := <-manager.broadcast:
			manager.mutex.RLock()
			for client := range manager.clients {
				if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
					delete(manager.clients, client)
					client.Close()
				}
			}
			manager.mutex.RUnlock()
		}
	}
}

// 发送消息给特定客户端
func (manager *WebSocketManager) sendToClient(client *websocket.Conn, message WebSocketMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("消息序列化失败: %v", err)
		return
	}

	if err := client.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("发送消息失败: %v", err)
		manager.unregister <- client
	}
}

// 广播消息给所有客户端
func (manager *WebSocketManager) broadcastToAll(message WebSocketMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("消息序列化失败: %v", err)
		return
	}

	manager.mutex.RLock()
	for client := range manager.clients {
		if err := client.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("广播消息失败: %v", err)
			manager.unregister <- client
		}
	}
	manager.mutex.RUnlock()
}

// 基础WebSocket连接处理
func handleWebSocketConnect(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	wsManager.register <- conn

	defer func() {
		wsManager.unregister <- conn
	}()

	for {
		var message WebSocketMessage
		if err := conn.ReadJSON(&message); err != nil {
			log.Printf("读取消息失败: %v", err)
			break
		}

		// 处理收到的消息
		response := WebSocketMessage{
			Type:      "response",
			Data:      fmt.Sprintf("服务器收到消息: %v", message),
			Timestamp: time.Now().Unix(),
			ID:        generateRequestID(),
		}

		wsManager.sendToClient(conn, response)
	}
}

// 回声WebSocket处理
func handleWebSocketEcho(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	defer conn.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("读取消息失败: %v", err)
			break
		}

		// 回声消息
		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Printf("发送回声消息失败: %v", err)
			break
		}
	}
}

// 广播WebSocket处理
func handleWebSocketBroadcast(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	wsManager.register <- conn

	defer func() {
		wsManager.unregister <- conn
	}()

	for {
		var message WebSocketMessage
		if err := conn.ReadJSON(&message); err != nil {
			log.Printf("读取消息失败: %v", err)
			break
		}

		// 广播消息
		broadcastMsg := WebSocketMessage{
			Type:      "broadcast",
			Data:      message.Data,
			Timestamp: time.Now().Unix(),
			ID:        generateRequestID(),
		}

		wsManager.broadcastToAll(broadcastMsg)
	}
}

// 实时数据推送WebSocket处理
func handleWebSocketRealtime(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	defer conn.Close()

	// 定期推送实时数据
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	counter := 0
	for {
		<-ticker.C
		counter++
		message := WebSocketMessage{
			Type: "realtime",
			Data: map[string]interface{}{
				"counter":   counter,
				"timestamp": time.Now().Unix(),
				"random":    time.Now().Nanosecond(),
			},
			Timestamp: time.Now().Unix(),
			ID:        generateRequestID(),
		}

		if err := conn.WriteJSON(message); err != nil {
			log.Printf("发送实时数据失败: %v", err)
			return
		}

		// 检查连接是否关闭
		if err := conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
			log.Printf("设置读取超时失败: %v", err)
			return
		}
		if _, _, err := conn.ReadMessage(); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket连接异常关闭: %v", err)
			}
			return
		}
	}
}

// 心跳检测WebSocket处理
func handleWebSocketHeartbeat(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	defer conn.Close()

	// 设置心跳
	if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
		log.Printf("设置读取超时失败: %v", err)
		return
	}
	conn.SetPongHandler(func(string) error {
		if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			log.Printf("设置Pong读取超时失败: %v", err)
		}
		return nil
	})

	// 发送心跳
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			log.Printf("发送心跳失败: %v", err)
			return
		}

		// 读取消息
		var message WebSocketMessage
		if err := conn.ReadJSON(&message); err != nil {
			log.Printf("读取消息失败: %v", err)
			break
		}

		// 响应消息
		response := WebSocketMessage{
			Type:      "heartbeat_response",
			Data:      "心跳正常",
			Timestamp: time.Now().Unix(),
			ID:        generateRequestID(),
		}

		if err := conn.WriteJSON(response); err != nil {
			log.Printf("发送心跳响应失败: %v", err)
			break
		}
	}
}

// 二进制数据传输WebSocket处理
func handleWebSocketBinary(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	defer conn.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("读取消息失败: %v", err)
			break
		}

		if messageType == websocket.BinaryMessage {
			// 处理二进制数据
			log.Printf("收到二进制数据，长度: %d", len(message))

			// 发送二进制响应
			response := make([]byte, len(message))
			copy(response, message)

			if err := conn.WriteMessage(websocket.BinaryMessage, response); err != nil {
				log.Printf("发送二进制响应失败: %v", err)
				break
			}
		} else {
			// 处理文本消息
			var textMessage WebSocketMessage
			if err := json.Unmarshal(message, &textMessage); err != nil {
				log.Printf("解析文本消息失败: %v", err)
				continue
			}

			// 生成二进制数据
			binaryData := make([]byte, 1024)
			for i := range binaryData {
				binaryData[i] = byte(i % 256)
			}

			if err := conn.WriteMessage(websocket.BinaryMessage, binaryData); err != nil {
				log.Printf("发送二进制数据失败: %v", err)
				break
			}
		}
	}
}

// 聊天室WebSocket处理
func handleWebSocketChat(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	wsManager.register <- conn

	defer func() {
		wsManager.unregister <- conn
	}()

	for {
		var message WebSocketMessage
		if err := conn.ReadJSON(&message); err != nil {
			log.Printf("读取消息失败: %v", err)
			break
		}

		// 处理聊天消息
		chatMessage := WebSocketMessage{
			Type: "chat",
			Data: map[string]interface{}{
				"user":    message.Data,
				"content": message.Data,
				"room":    "general",
			},
			Timestamp: time.Now().Unix(),
			ID:        generateRequestID(),
		}

		wsManager.broadcastToAll(chatMessage)
	}
}

// 性能测试WebSocket处理
func handleWebSocketPerformance(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	defer conn.Close()

	// 获取测试参数
	countStr := c.Query("count")
	count, err := strconv.Atoi(countStr)
	if err != nil || count < 1 || count > 10000 {
		count = 100
	}

	intervalStr := c.Query("interval")
	interval, err := strconv.Atoi(intervalStr)
	if err != nil || interval < 1 || interval > 5000 {
		interval = 10
	}

	// 发送性能测试消息
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
	defer ticker.Stop()

	sent := 0
	for range ticker.C {
		if sent >= count {
			// 发送完成消息
			completeMsg := WebSocketMessage{
				Type:      "performance_complete",
				Data:      fmt.Sprintf("性能测试完成，共发送 %d 条消息", count),
				Timestamp: time.Now().Unix(),
				ID:        generateRequestID(),
			}

			if err := conn.WriteJSON(completeMsg); err != nil {
				log.Printf("发送完成消息失败: %v", err)
			}
			return
		}

		sent++
		message := WebSocketMessage{
			Type: "performance",
			Data: map[string]interface{}{
				"sequence": sent,
				"total":    count,
				"content":  fmt.Sprintf("性能测试消息 #%d", sent),
			},
			Timestamp: time.Now().Unix(),
			ID:        generateRequestID(),
		}

		if err := conn.WriteJSON(message); err != nil {
			log.Printf("发送性能测试消息失败: %v", err)
			return
		}
	}
}

// 广播API处理
func handleBroadcastAPI(c *gin.Context) {
	var requestData struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, ApiResponse{
			Code:      400,
			Message:   "请求数据格式错误",
			Data:      nil,
			Timestamp: time.Now().Unix(),
			RequestID: generateRequestID(),
		})
		return
	}

	// 广播消息
	broadcastMsg := WebSocketMessage{
		Type:      requestData.Type,
		Data:      requestData.Message,
		Timestamp: time.Now().Unix(),
		ID:        generateRequestID(),
	}

	wsManager.broadcastToAll(broadcastMsg)

	response := ApiResponse{
		Code:      200,
		Message:   "广播消息发送成功",
		Data:      map[string]interface{}{"message": requestData.Message},
		Timestamp: time.Now().Unix(),
		RequestID: generateRequestID(),
	}

	c.JSON(http.StatusOK, response)
}
