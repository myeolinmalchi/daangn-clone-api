package repositories_test

import (
	"fmt"
	"testing"

	"carrot-market-clone-api/config"
	"carrot-market-clone-api/models"
	"carrot-market-clone-api/repositories"

	"github.com/stretchr/testify/assert"
)

func TestProductRepository(t *testing.T) {

	conf, err := config.LoadTestConfig()
	if err != nil {
		assert.Error(t, err)
	}

	db, err := conf.InitDBConnection()
	if err != nil {
		assert.Error(t, err)
	}

	r := repositories.NewProductRepositoryImpl(db)

	products := make([]models.Product, 5)
	for i := 0; i < len(products); i++ {
		products[i] = models.Product{
			Title:      fmt.Sprintf("test title %d", i+1),
			Content:    fmt.Sprintf("test content %d", i+1),
			Price:      (i + 1) * 10000,
			CategoryID: (i + 1),
			UserID:     "user",
			Images: []models.ProductImage{
				{URL: "test url 1", Sequence: 1},
				{URL: "test url 2", Sequence: 2},
				{URL: "test url 3", Sequence: 3},
			},
		}
	}

	// insert
	for i := 0; i < len(products); i++ {
		if err := r.InsertProduct(&products[i]); err != nil {
			assert.Error(t, err)
		}
	}

	// update
	err = r.UpdateProduct(&models.Product{
		ID:      products[0].ID,
		Title:   "update test title",
		Content: "update test content",
		Images: []models.ProductImage{
			{URL: "update test url", Sequence: 1},
		},
	})
	if err != nil {
		assert.Error(t, err)
	}

	// select
	product1, err := r.GetProduct(products[0].ID)
	if err != nil {
		assert.Error(t, err)
	}
	assert.Equal(t, "update test title", product1.Title)
	assert.Equal(t, 1, len(product1.Images))

	selectedProducts, count, err := r.GetProductsByUserID("user", nil, 2, "id ASC")

	assert.Equal(t, 2, count)
	assert.Equal(t, 2, len(selectedProducts))
	assert.Equal(t, "update test title", selectedProducts[0].Title)
	assert.Equal(t, "update test url", selectedProducts[0].Images[0].URL)

	keyword := "test"
	categoryId := 1
	last := products[3].ID

	selectedProducts, count, err = r.GetProducts(&keyword, &categoryId, nil, len(products), "id ASC")
	if err != nil {
		assert.Error(t, err)
	}

	assert.Equal(t, 1, count)
	assert.Equal(t, 1, len(selectedProducts))

	selectedProducts, count, err = r.GetProducts(nil, nil, &last, len(products), "id ASC")
	if err != nil {
		assert.Error(t, err)
	}

	assert.Equal(t, 3, count)
	assert.Equal(t, 3, len(selectedProducts))

	// delete
	for i := 0; i < len(products); i++ {
		if err := r.DeleteProduct(products[i].ID); err != nil {
			assert.Error(t, err)
		}
	}

}
