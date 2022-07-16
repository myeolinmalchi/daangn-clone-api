//go:build wireinject
// +build wireinject

package module

import (
	"carrot-market-clone-api/controllers"
	"carrot-market-clone-api/repositories"
	"carrot-market-clone-api/middlewares"
	"carrot-market-clone-api/services"
	"gorm.io/gorm"

	"github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/google/wire"
)

func InitProductController(db *gorm.DB, s3 *s3.Client) (c controllers.ProductController) {
    wire.Build(
        repositories.NewProductRepositoryImpl,
        repositories.NewUserRepositoryImpl,
        services.NewAWSServiceImpl,
        services.NewProductServiceImpl,
        controllers.NewProductControllerImpl,
    )
    return
}

func InitAuthMiddleware(db *gorm.DB) (m middlewares.AuthMiddleware) {
    wire.Build (
        repositories.NewUserRepositoryImpl,
        services.NewAuthServiceImpl,
        middlewares.NewAuthMiddlewareImpl,
    )
    return
}

func InitUserController(db *gorm.DB, s3 *s3.Client) (c controllers.UserController) {
    wire.Build (
        repositories.NewUserRepositoryImpl,
        services.NewAWSServiceImpl,
        services.NewAuthServiceImpl,
        services.NewUserServiceImpl,
        controllers.NewUserControllerImpl,
    )
    return
}
