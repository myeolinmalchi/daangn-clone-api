package models

import (
	"time"
)

type Product struct {
	ID           int            `json:"id,omitempty" gorm:"primaryKey"`
	Title        string         `json:"title"`
	Content      string         `json:"content,omitempty"`
	Price        *int           `json:"price,omitempty" gorm:"column:price"`
	CategoryID   int            `json:"categoryId,omitempty"`
	UserID       string         `json:"userId,omitempty"`
	Nickname     string         `json:"nickname,omitempty" gorm:"->"`
	ProfileImage string         `json:"profileImage,omitempty" gorm:"->"`
	Regdate      time.Time      `json:"regdate,omitempty" gorm:"->"`
	Views        int            `json:"views" gorm:"->"`
	Wishes       int            `json:"wishes" gorm:"->"`
	Chatrooms    int            `json:"chatrooms" gorm:"->"`
	Thumbnail    string         `json:"thumbnail,omitempty" gorm:"->"`
	Images       []ProductImage `json:"images,omitempty" gorm:"foreignKey:ProductID"`
}

type ProductW struct {
	Product
	Wished     bool `json:"wished,omitempty"`
	ChatroomID int  `json:"chatroomId,omitempty"`
}

type ProductImage struct {
	ProductID int    `json:"productId,omitempty"`
	ID        int    `json:"id,omitempty"`
	URL       string `json:"url"`
	Sequence  int    `json:"sequence"`
}

type Wish struct {
	ProductID int    `json:"productId"`
	ID        int    `json:"id"`
	UserID    string `json:"userId"`
}

type View struct {
	ProductID int       `json:"productId"`
	ID        int       `json:"id"`
	IP        string    `json:"ip"`
	ViewDate  time.Time `json:"viewDate,omitempty" gorm:"->"`
}

type ProductValidationResult struct {
	Title      *string `json:"title,omitempty"`
	Content    *string `json:"content,omitempty"`
	Price      *string `json:"price,omitempty"`
	CategoryID *string `json:"categoryId,omitempty"`
}

func (r *ProductValidationResult) GetOrNil() *ProductValidationResult {
	if r.Title == nil && r.Content == nil && r.Price == nil && r.CategoryID == nil {
		return nil
	}
	return r
}
