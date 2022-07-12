package controllers

import (
	"carrot-market-clone-api/models"
	"carrot-market-clone-api/services"
	"fmt"
	"mime/multipart"
	"os"
	"strings"

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
    File        *multipart.FileHeader       `form:"file" binding:"required"`
    Json        string                      `form:"json" binding:"required"`
}

// POST /api/v1/user
// TODO 로직 service 레이어로 옮기기
func (u *UserControllerImpl) Register(c *gin.Context) {
    
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

    validationResult := u.userService.UserRegistValidation(user)

    if validationResult != nil {
        c.IndentedJSON(422, validationResult)
        return
    }

    file, err := form.File.Open()
    if err != nil {
        c.JSON(400, gin.H{"message": err})
        return
    }

    filename, err := u.awsService.UploadFile(file)
    if err != nil {
        c.JSON(400, gin.H{"message": err})
        return
    }

    url := fmt.Sprintf("https://%s/images/%s", os.Getenv("AWS_S3_DOMAIN"), filename)
    user.ProfileImage = url

    err = u.userService.Register(user)
    if err != nil {
        c.JSON(400, gin.H{"message": err})

        //TODO
        // Delete image
        return
    }

    c.JSON(201, gin.H{"id": user.ID})

}

// TODO: 로직 service layer로 옮기기
func (u *UserControllerImpl) UpdateUser(c *gin.Context) {

    // ------------- 요청 form 바인딩 ----------
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

    beforeUser, err := u.userService.GetUserByID(userId)

    if err == gorm.ErrRecordNotFound {
        c.Status(404)
        return
    } else if err != nil {
        c.JSON(400, gin.H{"message": err})
        return
    }
    
    // -------------- 유효성 검사 ---------------
    validationResult := u.userService.UserUpdateValidation(user)

    if validationResult != nil {
        c.IndentedJSON(422, validationResult)
        return
    }

    // ---------- 변경할 이미지 업로드 ----------
    afterFilename, err := u.awsService.UploadFile(file)
    if err != nil { c.JSON(400, gin.H{"message": err}); return }

    // ------------ db 업데이트 ---------------
    user.ProfileImage = fmt.Sprintf("https://%s/images/%s", os.Getenv("AWS_S3_DOMAIN"), afterFilename)
    err = u.userService.Update(user)

    // db 업데이트 실패시 업로드한 이미지 삭제
    if err != nil {
        if err := u.awsService.DeleteFile(afterFilename); err != nil {
             c.JSON(400, gin.H{"message": err})
             return
        }
        c.JSON(400, gin.H{"message": err})
        return
    }

    // 모든 작업 성공시 기존 이미지 삭제
    beforeFilename := strings.Split(beforeUser.ProfileImage, "/")[4]

    // 기존 이미지 삭제에 실패해도 오류 코드를 보내지 않는다.
    err = u.awsService.DeleteFile(beforeFilename)

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
