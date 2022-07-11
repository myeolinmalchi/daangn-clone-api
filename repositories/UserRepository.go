package repositories

import (
	"carrot-market-clone-api/models"

	"gorm.io/gorm"
)

type UserRepository interface {
    CheckUserExists(column, value string)   (exists bool)

    GetUser(column, value string)           (user *models.User, err error)

    InsertUser(user *models.User)           (err error)

    UpdateUser(user *models.User)           (err error)

    DeleteUser(userId string)               (err error)
}

type UserRepositoryImpl struct {
    db *gorm.DB
}

func NewUserRepositoryImpl(db *gorm.DB) UserRepository {
    return &UserRepositoryImpl{ db: db }
}

func (r *UserRepositoryImpl) CheckUserExists(column, value string) (exists bool) {
    r.db.Model(&models.User{}).Select("count(*) > 0").Where(column + " = ?", value).Find(&exists)
    return
}

func (r *UserRepositoryImpl) GetUser(column, value string) (user *models.User, err error) {
    user = &models.User{}
    err = r.db.Model(&models.User{}).First(user, column + " = ?", value).Error
    return
}

func (r *UserRepositoryImpl) InsertUser(user *models.User) (err error) {
    err = r.db.Create(user).Error
    return
}

func (r *UserRepositoryImpl) UpdateUser(user *models.User) (err error) {
    err = r.db.UpdateColumns(user).Error
    return
}

func (r *UserRepositoryImpl) DeleteUser(userId string) (err error) {
    err = r.db.Delete(&models.User{}, "id = ?", userId).Error
    return
}
