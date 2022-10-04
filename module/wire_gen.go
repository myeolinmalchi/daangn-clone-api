// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package module

import (
	"carrot-market-clone-api/controllers"
	"carrot-market-clone-api/middlewares"
	"carrot-market-clone-api/models/chat"
	"carrot-market-clone-api/repositories"
	"carrot-market-clone-api/services"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gorm.io/gorm"
)

// Injectors from wire.go:

func InitProductController(db *gorm.DB, s3_2 *s3.Client) controllers.ProductController {
	productRepository := repositories.NewProductRepositoryImpl(db)
	userRepository := repositories.NewUserRepositoryImpl(db)
	awsService := services.NewAWSServiceImpl(s3_2)
	productSerivce := services.NewProductServiceImpl(productRepository, userRepository, awsService, s3_2)
	productController := controllers.NewProductControllerImpl(s3_2, productSerivce)
	return productController
}

func InitAuthMiddleware(db *gorm.DB) middlewares.AuthMiddleware {
	userRepository := repositories.NewUserRepositoryImpl(db)
	authService := services.NewAuthServiceImpl(userRepository)
	authMiddleware := middlewares.NewAuthMiddlewareImpl(authService)
	return authMiddleware
}

func InitUserController(db *gorm.DB, s3_2 *s3.Client) controllers.UserController {
	userRepository := repositories.NewUserRepositoryImpl(db)
	awsService := services.NewAWSServiceImpl(s3_2)
	userService := services.NewUserServiceImpl(userRepository, awsService, s3_2)
	authService := services.NewAuthServiceImpl(userRepository)
	userController := controllers.NewUserControllerImpl(userService, authService, awsService, s3_2)
	return userController
}

func InitChatController(db *gorm.DB) controllers.ChatController {
	productRepository := repositories.NewProductRepositoryImpl(db)
	chatRepository := repositories.NewChatRepositoryImpl(db, productRepository)
	chatService := services.NewChatServiceImpl(chatRepository)
	chatHub := chat.NewChatHub(chatService)
	chatController := controllers.NewChatControllerImpl(chatService, chatHub)
	return chatController
}
