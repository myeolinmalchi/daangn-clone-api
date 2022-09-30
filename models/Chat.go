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
	SendDate   time.Time `json:"sendDate"`
}

type Chatroom struct {
	ID        int      `json:"id,omitempty"`
	ProductID int      `json:"product_id,omitempty"`
	Seller    ChatUser `json:"seller,omitempty" gorm:"foreignKey:ChatroomID"`
	Buyer     ChatUser `json:"buyer,omitempty" gorm:"foreignKey:ChatroomID"`
	LastChat  Chat     `json:"lastChat,omitempty" gorm:"->"`
}

type ChatUser struct {
	ID         int      `json:"id,omitempty"`
	UserID     string   `json:"userId,omitempty"`
	ChatroomID int      `json:"chatroomId,omitempty"`
	Role       UserRole `json:"role,omitempty"`
}
