package main

import (
	"carrot-market-clone-api/config"
	"carrot-market-clone-api/module"
	"io"
	"log"
	"os"
    "time"

	"github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
)

func main() {
    conf, err := config.LoadConfig()
    if err != nil {
        log.Println("설정 파일을 불러오지 못했습니다. 서버를 종료합니다.")
        log.Println(err)
        return
    }

    if file, err := conf.InitLogger(); err != nil {
        log.Println("로그 파일을 생성하지 못했습니다. 서버를 종료합니다.")
        log.Println(err)
        return
    } else {
        gin.DefaultWriter = io.MultiWriter(file)
        log.SetOutput(file)
    }

    db, err := conf.InitDBConnection()
    if err != nil {
        log.Println("DB 연결에 실패했습니다. 서버를 종료합니다.")
        log.Println(err)
        return
    }

    s3, err := conf.InitS3Client()
    if err != nil {
        log.Println("AWS S3 연결에 실패했습니다. 서버를 종료합니다.")
        log.Println(err)
        return
    }

    os.Setenv("ACCESS_SECRET", conf.AuthConfig.AccessSecret)
    os.Setenv("REFRESH_SECRET", conf.AuthConfig.RefreshSecret)
    os.Setenv("AWS_S3_BUCKET", conf.AWSConfig.Bucket)
    os.Setenv("AWS_S3_DOMAIN", conf.AWSConfig.Domain)

    route := gin.New()
    route.Use(cors.New(cors.Config {
        AllowAllOrigins:    true,
        AllowMethods:       []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders:       []string{"Content-Type", "Authorization"},
        ExposeHeaders:      []string{"Authorization"},
        AllowCredentials:   true,
        MaxAge: 12 * time.Hour,
    }))
    route.Use(gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: []string{"/"}}))
    route.Use(gin.Recovery())

    productController := module.InitProductController(db, s3)
    userController := module.InitUserController(db, s3)
    authMiddleware := module.InitAuthMiddleware(db)

    route.GET("/", func(c *gin.Context) {
        c.Status(200)
    })

    v1 := route.Group("/api/v1")
    {
        v1.GET("/products/:productId", productController.GetProduct)
        v1.GET("/products", productController.GetProducts)
        
        v1.GET("/user/:userId/products", productController.GetUserProducts)
        v1.POST("/user/:userId/products", authMiddleware.UserAuth, productController.InsertProduct)
        v1.PUT("/user/:userId/products/:productId", authMiddleware.UserAuth, productController.UpdateProduct)
        v1.DELETE("/user/:userId/products/:productId", authMiddleware.UserAuth, productController.DeleteProduct)

        v1.POST("/user/auth/login", userController.Login)
        v1.POST("/user", userController.Register)
        v1.PUT("/user/:userId", authMiddleware.UserAuth, userController.UpdateUser)
        v1.DELETE("/user/:userId", authMiddleware.UserAuth, userController.DeleteUser)
    }
    route.Run(":3000")
    
    
}
