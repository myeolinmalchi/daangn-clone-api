package repositories_test

import (
	"carrot-market-clone-api/config"
	"carrot-market-clone-api/models"
	"carrot-market-clone-api/repositories"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChatRepository(t *testing.T) {

	conf, err := config.LoadTestConfig()
	if err != nil {
		assert.Error(t, err)
	}

	db, err := conf.InitDBConnection()
	if err != nil {
		assert.Error(t, err)
	}

	productRepo := repositories.NewProductRepositoryImpl(db)
	r := repositories.NewChatRepositoryImpl(db, productRepo)

	// insert test product
	product := &models.Product{
		Title:      "test title",
		Content:    "test content",
		Price:      30000,
		CategoryID: 1,
		UserID:     "517ff837-98ef-4851-b87a-c8199a8d465c",
		Images: []models.ProductImage{
			{URL: "test url", Sequence: 1},
		},
	}

	productRepo.InsertProduct(product)

	// insert chatroom
	buyerId := "7e2cfeea-1e1f-4fd0-9542-0f802e1dd954"
	chatroom, err := r.InsertChatroom(product.ID, buyerId)
	if err != nil {
		assert.Error(t, err)
	}

	// insert chats
	chats := make([]models.Chat, 10)
	for i := 0; i < len(chats); i++ {
		chats[i] = models.Chat{
			ChatUserID: chatroom.Seller.ID,
			Content:    fmt.Sprintf("test content %d", i+1),
		}
	}

	for i := 0; i < len(chats); i++ {
		if err := r.InsertChat(&chats[i]); err != nil {
			assert.Error(t, err)
		}
	}

	// get chat
	testChat, err := r.GetChat(chats[0].ID)
	if err != nil {
		assert.Error(t, err)
	}

	assert.Equal(t, "test content 1", testChat.Content)
	assert.Equal(t, chatroom.Seller.ID, testChat.ChatUserID)
	assert.Equal(t, models.SELLER, testChat.Role)

	// get chats
	testChats, count, err := r.GetChats(chatroom.ID, nil, 3)
	assert.Equal(t, 3, count)
	assert.Equal(t, 3, len(testChats))

	// get chatrooms
	testChatrooms, count, err := r.GetChatrooms(chatroom.Seller.UserID, nil, 5)
	assert.Equal(t, 1, count)
	assert.Equal(t, 1, len(testChatrooms))
	assert.Equal(t, "test content 10", testChatrooms[0].LastChat.Content)
	assert.Equal(t, buyerId, testChatrooms[0].Buyer.UserID)

	// delete chatroom
	if err := r.DeleteChatroom(chatroom.ID); err != nil {
		assert.Error(t, err)
	}

	// delete test product
	if err := productRepo.DeleteProduct(product.ID); err != nil {
		assert.Error(t, err)
	}
}
