package services

import (
	"carrot-market-clone-api/models"
	"carrot-market-clone-api/repositories"
	"carrot-market-clone-api/utils/encryption"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"regexp"
	"unicode"
)

type UserService interface {

    GetUserByID(userId string)      (user *models.User, err error)
    GetUserByEmail(email string)    (user *models.User, err error)
    GetUserByPhone(phone string)    (user *models.User, err error)

    Login(*models.User)         (ok bool, err error)

    Register(*models.User)      (ok bool, result *models.UserValidationResult)

    Update(*models.User)        (ok bool, result *models.UserValidationResult)

    Delete(userid string)       (err error)

}

type UserServiceImpl struct {
    userRepo repositories.UserRepository
    client *s3.Client
}

func NewUserServiceImpl(
    userRepo repositories.UserRepository,
    client *s3.Client,
) UserService {
    return &UserServiceImpl {
        userRepo: userRepo,
        client: client,
    }
}

func (s *UserServiceImpl) GetUserByID(userId string) (*models.User, error) {
    return s.userRepo.GetUser("id", userId)
}

func (s *UserServiceImpl) GetUserByEmail(email string) (*models.User, error) {
    return s.userRepo.GetUser("email", email)
}

func (s *UserServiceImpl) GetUserByPhone(phone string) (*models.User, error) {
    return s.userRepo.GetUser("phone", phone)
}

func (s *UserServiceImpl) Login(user *models.User) (ok bool, err error) {
    insertedPassword := user.PW
    userDetail, err := s.userRepo.GetUser("email", user.Email)
    if err != nil { return false, err }
    return encryption.EncryptSHA256(insertedPassword) == userDetail.PW, nil
}

// Validate Admin PW. If valid, it returns nil.
func (s *UserServiceImpl) checkPW(pw string) *string {
    var msg string
    var (
        hasMinLen   = false
        hasUpper    = false
        hasLower    = false
        hasNumber   = false
        hasSpecial  = false
    )
    if len(pw) >= 8 {
        hasMinLen = true
    }
    for _, char := range pw {
        switch {
        case unicode.IsUpper(char):
            hasUpper = true
        case unicode.IsLower(char):
            hasLower = true
        case unicode.IsNumber(char):
            hasNumber = true
        case unicode.IsPunct(char) || unicode.IsSymbol(char):
            hasSpecial = true
        }
    }

    if hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial {
        return nil
    } else {
        msg = "비밀번호는 최소 8자이며, 대문자, 소문자, 숫자 및 특수문자를 하나 이상 포함해야 합니다."
    }
    return &msg
}

// Validate Admin Name. If valid, it returns nil.
func (s *UserServiceImpl) checkName(name string) *string {
    var msg string
    if match, _ := regexp.MatchString("^[가-힣]+$", name); !match {
        msg = "이름은 한글만 입력할 수 있습니다."
    } else {
        return nil
    }
    return &msg
}

func (s *UserServiceImpl) checkNickname(nickname string) *string {
    var msg string
    if len(nickname) > 30 {
        msg = "별칭은 30자 이하까지 입력할 수 있습니다."
    } else if nickname == ""{
        msg = "별칭은 필수 항목입니다."
    } else {
        return nil
    }
    return &msg
}

// Validate Admin Email. If valid, it returns nil.
func (s *UserServiceImpl) checkEmail(email string) *string {
    var msg string
    if match, _ := regexp.
        MatchString("^[0-9a-zA-Z]([-_.]?[0-9a-zA-Z])*@[0-9a-zA-Z]([-_.]?[0-9a-zA-Z])*.[a-zA-Z]{2,3}$", email); !match {
        msg = "이메일 형식이 아닙니다."
    } else if s.userRepo.CheckUserExists("email", email) {
        msg = "이미 사용중인 이메일입니다."
    } else {
        return nil
    }
    return &msg
}

// Validate Admin Phone. If valid, it returns nil.
func (s *UserServiceImpl) checkPhone(phone string) *string {
    var msg string
    if match, _ := regexp.MatchString("^\\d{3}-\\d{3,4}-\\d{4}$", phone); !match {
        msg = "전화번호 형식이 아닙니다. ('-' 포함)"
    } else if s.userRepo.CheckUserExists("phone", phone) {
        msg = "이미 사용중인 전화번호입니다."
    } else {
        return nil
    }
    return &msg
}

func (s *UserServiceImpl) userRegistValidation(user *models.User) *models.UserValidationResult {
    result := &models.UserValidationResult{
        PW: s.checkPW(user.PW),
        Email: s.checkEmail(user.Email),
        Phone: s.checkPhone(user.Phone),
        Name: s.checkName(user.Name),
        Nickname: s.checkNickname(user.Nickname),
    }
    return result.GetOrNil()
}

func (s *UserServiceImpl) userUpdateValidation(user *models.User) *models.UserValidationResult {
    result := &models.UserValidationResult{
        PW: s.checkPW(user.PW),
        Email: s.checkEmail(user.Email),
        Phone: s.checkPhone(user.Phone),
        Name: s.checkName(user.Name),
        Nickname: s.checkNickname(user.Nickname),
    }
    return result.GetOrNil()
}

func (s *UserServiceImpl) Register(user *models.User) (ok bool, result *models.UserValidationResult) {
    result = s.userRegistValidation(user)
    if result == nil {
        user.PW = encryption.EncryptSHA256(user.PW)
        if err := s.userRepo.InsertUser(user); err != nil {
            return false, nil
        }
        return true, nil
    }
    return false, result
}

func (s *UserServiceImpl) Update(user *models.User) (ok bool, result *models.UserValidationResult) {
    result = s.userUpdateValidation(user)
    if result == nil {
        user.PW = encryption.EncryptSHA256(user.PW)
        if err := s.userRepo.InsertUser(user); err != nil {
            return false, nil
        }
        return true, nil
    }
    return false, result
}

func (s *UserServiceImpl) Delete(userId string) (err error) {
    return s.userRepo.DeleteUser(userId)
}
