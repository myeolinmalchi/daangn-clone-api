package models

import "time"

type UserRole string

const (
	BUYER  UserRole = "BUYER"
	SELLER UserRole = "SELLER"
)

type Chat struct {
	ID         int       `json:"id,omitempty"`
	ChatroomID int       `json:"chatroomId,omitempty" gorm:"->"`
	ChatUserID int       `json:"chatUserId,omitempty"`
	Role       UserRole  `json:"role,omitempty" gorm:"->"`
	Content    string    `json:"content"`
	SendDate   time.Time `json:"sendDate,omitempty" gorm:"->"`
}

type Chatroom struct {
	ID        int      `json:"id,omitempty"`
	ProductID int      `json:"productId,omitempty"`
	Seller    ChatUser `json:"seller,omitempty" gorm:"foreignKey:ChatroomID"`
	Buyer     ChatUser `json:"buyer,omitempty" gorm:"foreignKey:ChatroomID"`
	Product   Product  `json:"product,omitempty" gorm:"foreignKey:ID;references:ProductID"`
	LastChat  Chat     `json:"lastChat,omitempty" gorm:"->"`
}

type ChatUser struct {
	ID           int      `json:"id,omitempty"`
	UserID       string   `json:"userId,omitempty"`
	ChatroomID   int      `json:"chatroomId,omitempty"`
	Role         UserRole `json:"role,omitempty"`
	Nickname     string   `json:"nickname,omitempty" gorm:"->"`
	ProfileImage string   `json:"profileImage,omitempty" gorm:"->"`
}
