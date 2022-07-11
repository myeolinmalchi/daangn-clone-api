package services

import (
	"carrot-market-clone-api/repositories"
	"gorm.io/gorm"
	"os"
	"time"
    "errors"

	"github.com/golang-jwt/jwt"
)

type AuthService interface {
    CreateAccessToken(userId string)        (at string, err error)
    VerifyAccessToken(at string)            (claims jwt.MapClaims, err error)      
}

type AuthServiceImpl struct {
    userRepo repositories.UserRepository
}

func NewAuthServiceImpl(userRepo repositories.UserRepository) AuthService {
    return &AuthServiceImpl{ userRepo: userRepo }
}

func (s *AuthServiceImpl) CreateAccessToken(userId string) (at string, err error) {
    atClaims := jwt.MapClaims{}
    if !s.userRepo.CheckUserExists("id", userId) {
        return "", gorm.ErrRecordNotFound
    }
    atClaims["authorized"] = true
    atClaims["user_id"] = userId
    atClaims["role"] = "user"
    atClaims["exp"] = time.Now().Add(time.Hour * 24 * 7).Unix()
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
    at, err = token.SignedString([]byte(os.Getenv("ACCESS_SECRET")))

    return
}

func (s *AuthServiceImpl) VerifyAccessToken(at string) (claims jwt.MapClaims, err error) {
    claims = jwt.MapClaims{}
    verifying := func(token *jwt.Token) (interface{}, error) {
        if token.Method != jwt.SigningMethodHS256 {
            return nil, errors.New("Unexpected Signing Method")
        }
        return []byte(os.Getenv("ACCESS_SECRET")), nil
    }
    _, err = jwt.ParseWithClaims(at, &claims, verifying)
    return
}


