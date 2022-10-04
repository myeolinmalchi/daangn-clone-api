package services

import (
	"carrot-market-clone-api/models"
	"carrot-market-clone-api/repositories"
)

type ChatService interface {
	CreateChatroom(productId int, userId string) (chatroomId int, err error)
	InsertChat(chatroomId int, userId, content string) (err error)
	CheckCorrectUser(userId string, chatroomId int) (isCorrect bool)
	GetChatroom(chatroomId int) (chatroom *models.Chatroom, err error)
	GetChatrooms(
		userId string,
		last *int,
		size int,
	) (chatrooms []models.Chatroom, count int, err error)
	GetChats(
		chatroomId int,
		last *int,
		size int,
	) (chats []models.Chat, count int, err error)
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

func (s *ChatServiceImpl) GetChatroom(chatroomId int) (chatroom *models.Chatroom, err error) {
	return s.chatRepo.GetChatroom(chatroomId)
}

func (s *ChatServiceImpl) GetChats(
	chatroomId int,
	last *int,
	size int,
) (chats []models.Chat, count int, err error) {
	return s.chatRepo.GetChats(chatroomId, last, size)
}

func (s *ChatServiceImpl) GetChatrooms(
	userId string,
	last *int,
	size int,
) (chatrooms []models.Chatroom, count int, err error) {
	return s.chatRepo.GetChatrooms(userId, last, size)
}
