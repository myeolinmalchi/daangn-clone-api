package controllers

import (
	"carrot-market-clone-api/models/chat"
	"carrot-market-clone-api/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type ChatController interface {
	NewChatConnection(c *gin.Context)
	CreateChatroom(c *gin.Context)
	GetChatroom(c *gin.Context)
	GetChatrooms(c *gin.Context)
	GetChats(c *gin.Context)
}

type ChatControllerImpl struct {
	chatService services.ChatService
	chatHub     chat.ChatHub
}

func NewChatControllerImpl(
	chatService services.ChatService,
	chatHub chat.ChatHub,
) ChatController {
	go chatHub.Run()
	return &ChatControllerImpl{
		chatService: chatService,
		chatHub:     chatHub,
	}
}

// GET ws:// ~ /api/v1/users/{userId}/chatrooms/{chatroomId}/ws
func (t *ChatControllerImpl) NewChatConnection(c *gin.Context) {
	userId := c.Param("userId")
	chatroomId, err := strconv.Atoi(c.Param("chatroomId"))

	if err != nil {
		c.JSON(400, gin.H{"message": "chatroomdId는 정수값이어야 합니다."})
		return
	}

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	w := c.Writer
	r := c.Request
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		c.JSON(400, gin.H{"message": "연결을 생성하지 못했습니다."})
		return
	}

	chatroom := &chat.Chatroom{
		ID:      chatroomId,
		Hub:     &t.chatHub,
		Clients: make(map[*chat.Client]bool),
		Send:    make(chan chat.Chat),
	}

	t.chatHub.Register <- &chat.Client{
		Chatroom: chatroom,
		UserID:   userId,
		Send:     make(chan chat.Chat),
		Conn:     conn,
	}
	c.Status(200)
}

// POST /api/v1/users/{userId}/products/{productId}/chatrooms
func (t *ChatControllerImpl) CreateChatroom(c *gin.Context) {
	userId := c.Param("userId")
	productId, err := strconv.Atoi(c.Param("productId"))

	if err != nil {
		c.JSON(400, gin.H{"message": "productId는 정수값이어야 합니다."})
		return
	}

	chatroomId, err := t.chatService.CreateChatroom(productId, userId)

	if err == gorm.ErrInvalidValue {
		c.JSON(403, gin.H{"message": "본인과 채팅할 수 없습니다."})
		return
	}

	if err == gorm.ErrRecordNotFound {
		c.JSON(404, gin.H{"message": err})
		return
	}

	if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	c.JSON(201, gin.H{"chatroomId": chatroomId})

}

// GET /api/v1/users/{userId}/chatrooms/{chatroomId}
func (t *ChatControllerImpl) GetChatroom(c *gin.Context) {
	userId := c.Param("userId")
	chatroomId, err := strconv.Atoi(c.Param("chatroomId"))

	if err != nil {
		c.JSON(400, gin.H{"message": "chatroomdId는 정수값이어야 합니다."})
		return
	}

	if ok := t.chatService.CheckCorrectUser(userId, chatroomId); !ok {
		c.JSON(403, gin.H{"message": "접근 권한이 없습니다"})
		return
	}

	chatroom, err := t.chatService.GetChatroom(chatroomId)

	if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	c.JSON(200, chatroom)

}

// GET /api/v1/users/{userId}/chatrooms
func (t *ChatControllerImpl) GetChatrooms(c *gin.Context) {
	userId := c.Param("userId")
	var (
		err  error
		last *int
		size int
	)

	size, err = strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	if lastStr, lastExists := c.GetQuery("last"); lastExists {
		temp, err := strconv.Atoi(lastStr)
		if err != nil {
			c.JSON(400, gin.H{"message": err})
			return
		}
		last = &temp
	} else {
		last = nil
	}

	chatrooms, count, err := t.chatService.GetChatrooms(userId, last, size)
	if err == gorm.ErrRecordNotFound {
		c.JSON(404, gin.H{"message": err})
		return
	}

	if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	c.JSON(200, gin.H{
		"chatrooms": chatrooms,
		"size":      count,
		"userId":    userId,
	})
}

// GET /api/v1/users/{userId}/chatrooms/{chatroomId}/chats
func (t *ChatControllerImpl) GetChats(c *gin.Context) {
	userId := c.Param("userId")
	chatroomId, err := strconv.Atoi(c.Param("chatroomId"))

	if err != nil {
		c.JSON(400, gin.H{"message": "chatroomdId는 정수값이어야 합니다."})
		return
	}

	if ok := t.chatService.CheckCorrectUser(userId, chatroomId); !ok {
		c.JSON(403, gin.H{"message": "접근 권한이 없습니다"})
		return
	}

	var (
		last *int
		size int
	)

	size, err = strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	if lastStr, lastExists := c.GetQuery("last"); lastExists {
		temp, err := strconv.Atoi(lastStr)
		if err != nil {
			c.JSON(400, gin.H{"message": err})
			return
		}
		last = &temp
	} else {
		last = nil
	}

	chats, count, err := t.chatService.GetChats(chatroomId, last, size)

	if err == gorm.ErrRecordNotFound {
		c.JSON(404, gin.H{"message": err})
		return
	}

	if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	c.JSON(200, gin.H{
		"chatroomId": chatroomId,
		"size":       count,
		"chats":      chats,
	})
}
