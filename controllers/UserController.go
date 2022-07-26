package controllers

import (
	"carrot-market-clone-api/models"
	"carrot-market-clone-api/services"
	"mime/multipart"

	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserController interface {
    Login(c *gin.Context)
    Register(c *gin.Context)
    UpdateUser(c *gin.Context)
    DeleteUser(c *gin.Context)
}

type UserControllerImpl struct {
    userService services.UserService
    authService services.AuthService
    awsService services.AWSService
    client *s3.Client
}

func NewUserControllerImpl(
    userService services.UserService,
    authService services.AuthService,
    awsService services.AWSService,
    client *s3.Client,
) UserController {
    return &UserControllerImpl {
        userService: userService,
        authService: authService,
        awsService: awsService,
        client: client,
    }
}
// POST /api/v1/user/auth/login
func (u *UserControllerImpl) Login(c *gin.Context) {
    user := &models.User{}
    err := c.ShouldBind(user)
    if err != nil { c.JSON(400, gin.H{"message": err}); return }

    if ok, err := u.userService.Login(user); err != nil {
        if err == gorm.ErrRecordNotFound {
            c.Status(404)
            return
        } else {
            c.JSON(400, gin.H {"message": err})
            return
        }
    } else if !ok {
        c.Status(401)
        return
    } else {
        userDetail, err := u.userService.GetUserByEmail(user.Email)
        if err != nil {
            c.JSON(400, gin.H {"message": err})
            return
        }
        at, err := u.authService.CreateAccessToken(userDetail.ID)
        if err != nil {
            c.JSON(400, gin.H {"message": err})
            return
        }
        c.Header("Authorization", at)
        c.Status(200)
    }
}

type UserForm struct {
    File        *multipart.FileHeader       `form:"file" binding:"omitempty"`
    Json        string                      `form:"json" binding:"required"`
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

// PUT /api/v1/user/{userId}
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
    if err != nil { c.JSON(400, gin.H{"message": err}); return }

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
}
