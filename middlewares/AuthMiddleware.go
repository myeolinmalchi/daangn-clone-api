package middlewares

import (
	"carrot-market-clone-api/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)


type AuthMiddleware interface {
    UserAuth(c *gin.Context)
}

type AuthMiddlewareImpl struct {
    authService services.AuthService
}

func NewAuthMiddlewareImpl(authService services.AuthService) AuthMiddleware {
    return &AuthMiddlewareImpl{ authService: authService }
}

func (a *AuthMiddlewareImpl) UserAuth(c *gin.Context) {

    userId := c.Param("userId")

    token := c.Request.Header.Get("Authorization")

    if token == "" {
        c.JSON(401, gin.H{"message": "access token is empty."})
        c.Abort()
    } else if claims, err := a.authService.VerifyAccessToken(token); err != nil {
        if v, _ := err.(*jwt.ValidationError); v.Errors == jwt.ValidationErrorExpired {
            c.JSON(401, gin.H{"message": "access token is expired"})
            c.Abort()
        } else  {
            c.JSON(401, gin.H{"message": "invalid access token"})
            c.Abort()
        }
    } else {
        tokenUserId := claims["user_id"].(string)
        tokenRole := claims["role"].(string)
        if tokenRole != "user" || tokenUserId != userId {
            c.AbortWithStatus(403)
        }
    }
    
}
