package services

import (
	"carrot-market-clone-api/models"
	"carrot-market-clone-api/repositories"
)

type ChatService interface {
	CreateChatroom(productId int, userId string) (chatroomId int, err error)
	InsertChat(chatroomId int, userId, content string) (err error)
	CheckCorrectUser(userId string, chatroomId int) (isCorrect bool)
}

type ChatServiceImpl struct {
	chatRepo repositories.ChatRepository
}

func NewChatServiceImpl(chatRepo repositories.ChatRepository) ChatService {
	return &ChatServiceImpl{
		chatRepo: chatRepo,
	}
}

func (s *ChatServiceImpl) CreateChatroom(productId int, userId string) (chatroomId int, err error) {
	chatroom, err := s.chatRepo.InsertChatroom(productId, userId)
	chatroomId = chatroom.ID
	return
}

func (s *ChatServiceImpl) InsertChat(chatroomId int, userId, content string) (err error) {
	err = s.chatRepo.InsertChat(&models.Chat{
		ChatroomID: chatroomId,
		ChatUserID: s.chatRepo.GetChatUserId(chatroomId, userId),
		Content:    content,
	})
	return
}

func (s *ChatServiceImpl) CheckCorrectUser(userId string, chatroomId int) (isCorrect bool) {
	isCorrect = s.chatRepo.CheckCorrectUser(userId, chatroomId)
	return
}
