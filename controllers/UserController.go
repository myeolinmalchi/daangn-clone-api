package controllers

import (
	"carrot-market-clone-api/models"
	"carrot-market-clone-api/models/chat"
	"carrot-market-clone-api/services"
	"mime/multipart"
	"strconv"

	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type UserController interface {
	Login(c *gin.Context)
	Register(c *gin.Context)
	UpdateUser(c *gin.Context)
	DeleteUser(c *gin.Context)
	GetUserData(c *gin.Context)
	NewChatConnection(c *gin.Context)
	CreateChatroom(c *gin.Context)
}

type UserControllerImpl struct {
	userService services.UserService
	authService services.AuthService
	awsService  services.AWSService
	chatService services.ChatService
	client      *s3.Client
	chatHub     chat.ChatHub
}

func NewUserControllerImpl(
	userService services.UserService,
	authService services.AuthService,
	awsService services.AWSService,
	chatService services.ChatService,
	client *s3.Client,
	chatHub chat.ChatHub,
) UserController {
	return &UserControllerImpl{
		userService: userService,
		authService: authService,
		awsService:  awsService,
		chatService: chatService,
		client:      client,
		chatHub:     chatHub,
	}
}

// POST /api/v1/users/auth/login
func (u *UserControllerImpl) Login(c *gin.Context) {
	user := &models.User{}
	err := c.ShouldBind(user)
	if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	if ok, err := u.userService.Login(user); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Status(404)
			return
		} else {
			c.JSON(400, gin.H{"message": err})
			return
		}
	} else if !ok {
		c.Status(401)
		return
	} else {
		userDetail, err := u.userService.GetUserByEmail(user.Email)
		if err != nil {
			c.JSON(400, gin.H{"message": err})
			return
		}
		at, err := u.authService.CreateAccessToken(userDetail.ID)
		if err != nil {
			c.JSON(400, gin.H{"message": err})
			return
		}
		c.Header("Authorization", at)
		c.Status(200)
	}
}

type UserForm struct {
	File *multipart.FileHeader `form:"file" binding:"omitempty"`
	Json string                `form:"json" binding:"required"`
}

// POST /api/v1/user
func (u *UserControllerImpl) Register(c *gin.Context) {

	form := UserForm{}

	if err := c.ShouldBind(&form); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}

	user := &models.User{}
	err := json.Unmarshal([]byte(form.Json), user)

	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}

	validationResult := &models.UserValidationResult{}

	if form.File == nil {
		validationResult, err = u.userService.Register(nil, user)
	} else {
		file, err := form.File.Open()
		if err != nil {
			c.JSON(400, gin.H{"message": err.Error()})
			return
		}
		validationResult, err = u.userService.Register(file, user)
	}

	if validationResult != nil {
		c.IndentedJSON(422, validationResult)
		return
	}

	if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	c.Status(201)

}

// PUT /api/v1/users/{userId}
func (u *UserControllerImpl) UpdateUser(c *gin.Context) {

	form := UserForm{}

	if err := c.ShouldBind(&form); err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	user := &models.User{}
	err := json.Unmarshal([]byte(form.Json), user)

	if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	userId := c.Param("userId")
	if user.ID != userId {
		c.Status(403)
		return
	}

	file, err := form.File.Open()
	if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	validationResult, err := u.userService.Update(file, user)

	if err == gorm.ErrRecordNotFound {
		c.Status(404)
		return
	} else if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	if validationResult != nil {
		c.IndentedJSON(422, validationResult)
		return
	}

	c.Status(200)
}

// DELETE /api/v1/users/{userId}
func (u *UserControllerImpl) DeleteUser(c *gin.Context) {
	userId := c.Param("userId")

	err := u.userService.Delete(userId)

	if err == gorm.ErrRecordNotFound {
		c.Status(404)
		return
	} else if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	c.Status(200)
}

// GET /api/v1/users/{userId}
func (u *UserControllerImpl) GetUserData(c *gin.Context) {
	userId := c.Param("userId")

	user, err := u.userService.GetUserByID(userId)

	if err == gorm.ErrRecordNotFound {
		c.Status(404)
		return
	} else if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	c.IndentedJSON(200, user)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// WS /api/v1/users/{userId}/chats/{chatroomId}
func (u *UserControllerImpl) NewChatConnection(c *gin.Context) {

	userId := c.Param("userId")
	chatroomId, err := strconv.Atoi(c.Param("chatroomId"))

	if err != nil {
		c.JSON(400, gin.H{"message": "chatroomdId는 정수값이어야 합니다."})
		return
	}

	w := c.Writer
	r := c.Request
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		c.JSON(400, gin.H{"message": "연결을 생성하지 못했습니다."})
		return
	}
	chatroom := &chat.Chatroom{
		ID: chatroomId,
	}

	u.chatHub.Register <- &chat.Client{
		Chatroom: chatroom,
		UserID:   userId,
		Send:     make(chan []byte),
		Conn:     conn,
	}
	c.Status(200)
}

// POST /api/v1/users/{userId}/products/{productId}/chats
func (u *UserControllerImpl) CreateChatroom(c *gin.Context) {
	userId := c.Param("userId")
	productId, err := strconv.Atoi(c.Param("productId"))

	if err != nil {
		c.JSON(400, gin.H{"message": "productId는 정수값이어야 합니다."})
		return
	}

	chatroomId, err := u.chatService.CreateChatroom(productId, userId)

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
