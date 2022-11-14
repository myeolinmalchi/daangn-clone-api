package repositories

import (
	"carrot-market-clone-api/models"

	"gorm.io/gorm"
)

type ChatRepository interface {
	GetChat(chatId int) (chat *models.Chat, err error)

	GetChatrooms(
		userId string,
		last *int,
		size *int,
	) (chatrooms []models.Chatroom, count int, err error)

	GetChatroom(chatroomId int) (chatroom *models.Chatroom, err error)

	GetChats(
		chatroomId int,
		last *int,
		size int,
	) (chats []models.Chat, count int, err error)

	GetChatUserId(chatroomId int, userId string) (chatUserId int)

	InsertChatroom(productId int, buyerId string) (chatroom *models.Chatroom, err error)

	InsertChat(chat *models.Chat) (err error)

	DeleteChatroom(chatroomId int) (err error)

	DeleteChat(chatId int) (err error)

	CheckChatroomExists(chatroomId int) (exists bool)

	CheckCorrectUser(userId string, chatroomId int) (isCorrect bool)
}

type ChatRepositoryImpl struct {
	db          *gorm.DB
	productRepo ProductRepository
}

func NewChatRepositoryImpl(db *gorm.DB, productRepo ProductRepository) ChatRepository {
	return &ChatRepositoryImpl{
		db:          db,
		productRepo: productRepo,
	}
}

func (r *ChatRepositoryImpl) GetChat(chatId int) (chat *models.Chat, err error) {
	chat = &models.Chat{}
	err = r.db.Table("v_chats").Where("id  = ?", chatId).First(chat).Error
	return
}

func (r *ChatRepositoryImpl) GetChats(
	chatroomId int,
	last *int,
	size int,
) (chats []models.Chat, count int, err error) {
	chats = []models.Chat{}

	query := r.db.Table("v_chats").Where("chatroom_id = ?", chatroomId)

	if last != nil {
		query = query.Where("id < ?", last)
	}

	query = query.Limit(size)
	r.db.Table("(?) as a", query).Select("count(*)").Find(&count)

	err = query.Find(&chats).Error
	return
}

func (r *ChatRepositoryImpl) GetChatrooms(
	userId string,
	last *int,
	size *int,
) (chatrooms []models.Chatroom, count int, err error) {
	chatrooms = []models.Chatroom{}
	query := r.db.Table("chatrooms").
		Select("chatrooms.*").
		Joins("inner join chat_users on chat_users.chatroom_id = chatrooms.id").
		Where("chat_users.user_id = ?", userId)

	if last != nil {
		query = query.Where("chatrooms.id < ?", last)
	}

	query = query.Preload("Product", func(db *gorm.DB) *gorm.DB {
		return db.Table("v_products").Select("content", "id", "price", "regdate", "title", "thumbnail")
	}).Preload("LastChat", func(db *gorm.DB) *gorm.DB {
		return db.Table("v_chats").Select("chatroom_id", "content", "send_date").Order("send_date desc")
	}).Preload("Seller", func(db *gorm.DB) *gorm.DB {
		return db.Select("chat_users.user_id", "chat_users.chatroom_id", "users.nickname", "users.profile_image").
			Joins("JOIN users ON users.id = chat_users.user_id").
			Where("chat_users.role = ?", models.SELLER)
	}).Preload("Buyer", func(db *gorm.DB) *gorm.DB {
		return db.Select("chat_users.user_id", "chat_users.chatroom_id", "users.nickname", "users.profile_image").
			Joins("JOIN users ON users.id = chat_users.user_id").
			Where("chat_users.role = ?", models.BUYER)
	})

	if size != nil {
		query = query.Limit(*size)
	}

	r.db.Table("(?) as a", query).Select("count(*)").Find(&count)

	err = query.Find(&chatrooms).Error

	return
}

func (r *ChatRepositoryImpl) GetChatroom(chatroomId int) (chatroom *models.Chatroom, err error) {
	chatroom = &models.Chatroom{}
	query := r.db.Table("chatrooms").
		Select("chatrooms.*").
		Joins("inner join chat_users on chat_users.chatroom_id = chatrooms.id").
		Where("chatrooms.id = ?", chatroomId)

	err = query.Preload("Product", func(db *gorm.DB) *gorm.DB {
		return db.Table("v_products").Select("content", "id", "price", "regdate", "title", "thumbnail")
	}).Preload("LastChat", func(db *gorm.DB) *gorm.DB {
		return db.Table("v_chats").Select("chatroom_id", "content", "send_date").Order("send_date desc")
	}).Preload("Seller", func(db *gorm.DB) *gorm.DB {
		return db.Select("chat_users.user_id", "chat_users.chatroom_id", "users.nickname", "users.profile_image").
			Joins("JOIN users ON users.id = chat_users.user_id").
			Where("chat_users.role = ?", models.SELLER)
	}).Preload("Buyer", func(db *gorm.DB) *gorm.DB {
		return db.Select("chat_users.user_id", "chat_users.chatroom_id", "users.nickname", "users.profile_image").
			Joins("JOIN users ON users.id = chat_users.user_id").
			Where("chat_users.role = ?", models.BUYER)
	}).First(chatroom).Error

	return
}

func (r *ChatRepositoryImpl) InsertChatroom(productId int, buyerId string) (chatroom *models.Chatroom, err error) {

	err = r.db.Transaction(func(tx *gorm.DB) error {
		err := r.db.
			Table("chatrooms").
			Select("chatrooms.*").
			Joins("INNER JOIN chat_users ON chat_users.chatroom_id = chatrooms.id").
			Where("chatrooms.product_id = ? AND chat_users.user_id = ?", productId, buyerId).
			First(&chatroom).
			Error

		if err == gorm.ErrRecordNotFound {
			var sellerId string
			err := tx.Model(&models.Product{}).
				Select("user_id").
				Where("id = ?", productId).
				Find(&sellerId).
				Error

			if err != nil {
				return err
			}

			if sellerId == buyerId {
				return gorm.ErrInvalidValue
			}

			chatroom = &models.Chatroom{
				ProductID: productId,
				Seller: models.ChatUser{
					UserID: sellerId,
					Role:   models.SELLER,
				},
				Buyer: models.ChatUser{
					UserID: buyerId,
					Role:   models.BUYER,
				},
			}

			err = tx.Create(chatroom).Error
			return err
		} else if err != nil {
			return err
		} else {
			return nil
		}
	})

	return
}

func (r *ChatRepositoryImpl) GetChatUserId(chatroomId int, userId string) (chatUserId int) {
	r.db.Table("chat_users").
		Select("id").
		Where("chatroom_id = ? AND user_id = ?", chatroomId, userId).
		Find(&chatUserId)
	return
}

func (r *ChatRepositoryImpl) InsertChat(chat *models.Chat) (err error) {
	err = r.db.Create(chat).Error
	return
}

func (r *ChatRepositoryImpl) DeleteChatroom(chatroomId int) (err error) {
	err = r.db.Delete(&models.Chatroom{}, "id = ?", chatroomId).Error
	return
}

func (r *ChatRepositoryImpl) DeleteChat(chatId int) (err error) {
	err = r.db.Delete(&models.Chat{}, "id = ?", chatId).Error
	return
}

func (r *ChatRepositoryImpl) CheckChatroomExists(chatroomId int) (exists bool) {
	r.db.Model(&models.Chatroom{}).Select("count(*) > 0").Where("id = ?", chatroomId).Find(&exists)
	return
}

func (r *ChatRepositoryImpl) CheckCorrectUser(
	userId string,
	chatroomId int,
) (isCorrect bool) {
	r.db.Model(&models.ChatUser{}).Select("count(*) > 0").
		Where("user_id = ? AND chatroom_id = ?", userId, chatroomId).
		Find(&isCorrect)
	return
}
