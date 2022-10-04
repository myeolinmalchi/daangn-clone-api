package repositories

import (
	"carrot-market-clone-api/models"

	"gorm.io/gorm"
)

type ProductRepository interface {
	GetProduct(productId int) (product *models.Product, err error)

	ViewProduct(productId int, ip string) (products *models.Product, err error)

	GetProductsByUserID(
		userId string,
		last *int,
		size int,
		orderBy ...string,
	) (products []models.Product, count int, err error)

	GetProducts(
		keyword *string,
		categoryId *int,
		last *int,
		size int,
		orderBy ...string,
	) (products []models.Product, count int, err error)

	GetWishProducts(
		userId string,
		last *int,
		size int,
	) (products []models.Product, count int, err error)

	InsertProduct(product *models.Product) (err error)

	UpdateProduct(product *models.Product) (err error)

	DeleteProduct(productId int) (err error)

	CheckProductExists(productId int) (exists bool)

	CheckCorrectCategory(categoryId int) (correct bool)

	GetOwnerId(productId int) (userId string)

	CheckWishExists(wish *models.Wish) (exists bool)

	InsertWish(wish *models.Wish) (err error)

	DeleteWish(wish *models.Wish) (err error)

	InsertView(view *models.View) (err error)
}

type ProductRepositoryImpl struct {
	db *gorm.DB
}

func NewProductRepositoryImpl(
	db *gorm.DB,
) ProductRepository {
	return &ProductRepositoryImpl{db: db}
}

func (r *ProductRepositoryImpl) CheckCorrectCategory(categoryId int) (correct bool) {
	r.db.Table("categories").Select("count(*) > 0").Where("id = ?", categoryId).Find(&correct)
	return
}

func (r *ProductRepositoryImpl) GetProduct(productId int) (product *models.Product, err error) {
	product = &models.Product{}
	err = r.db.Table("v_products").Omit("Thumbnail").Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_images.sequence ASC")
	}).Where("id = ?", productId).First(product).Error
	return
}

func (r *ProductRepositoryImpl) ViewProduct(productId int, ip string) (product *models.Product, err error) {
	product = &models.Product{}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&models.View{
			ProductID: productId,
			IP:        ip,
		}).Error
		if err != nil {
			return err
		}
		err = tx.Table("v_products").Omit("Thumbnail").Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("product_images.sequence ASC")
		}).Where("id = ?", productId).First(product).Error
		if err != nil {
			return err
		}
		return nil
	})
	return
}

func (r *ProductRepositoryImpl) GetProductsByUserID(
	userId string,
	last *int,
	size int,
	orderBy ...string,
) (products []models.Product, count int, err error) {

	products = []models.Product{}

	query := r.db.Table("v_products").
		Omit("Content", "CategoryID", "Views", "UserID", "Nickname", "ProfileImage").
		Where("user_id = ?", userId)

	if last != nil {
		query = query.Where("id < ?", last)
	}

	for _, order := range orderBy {
		query = query.Order(order)
	}
	query = query.Limit(size)

	r.db.Table("(?) as a", query).Select("count(*)").Find(&count)

	err = query.Find(&products).Error

	return
}

func (r *ProductRepositoryImpl) GetProducts(
	keyword *string,
	categoryId *int,
	last *int,
	size int,
	orderBy ...string,
) (products []models.Product, count int, err error) {

	products = []models.Product{}

	query := r.db.Table("v_products").Omit("Content", "CategoryID", "Views", "UserID", "Nickname", "ProfileImage")

	if keyword != nil {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+*keyword+"%", "%"+*keyword+"%")
	}

	if categoryId != nil {
		query = query.Where("category_id = ?", categoryId)
	}

	if last != nil {
		query = query.Where("id < ?", last)
	}

	for _, order := range orderBy {
		query = query.Order(order)
	}

	query = query.Limit(size)

	r.db.Table("(?) as a", query).Select("count(*)").Find(&count)

	err = query.Find(&products).Error

	return
}

func (r *ProductRepositoryImpl) GetWishProducts(
	userId string,
	last *int,
	size int,
) (products []models.Product, count int, err error) {

	products = []models.Product{}

	query := r.db.Table("v_products").Omit("Content", "CategoryID", "Views", "UserID", "Nickname", "ProfileImage").
		Joins("JOIN wishes ON v_products.id = wishes.product_id").
		Order("v_products.id desc")

	if last != nil {
		query = query.Where("wishes.product_id < ?", last)
	}

	query = query.Where("wishes.user_id = ?", userId).Limit(size)

	r.db.Table("(?) as a", query).Select("count(*)").Find(&count)

	err = query.Find(&products).Error

	return
}

func (r *ProductRepositoryImpl) InsertProduct(product *models.Product) (err error) {
	err = r.db.Create(product).Error
	return
}

func (r *ProductRepositoryImpl) UpdateProduct(product *models.Product) (err error) {
	err = r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.UpdateColumns(product).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.ProductImage{}, "product_id = ?", product.ID).Error; err != nil {
			return err
		}
		if len(product.Images) < 1 {
			return nil
		}
		if err := tx.Create(product.Images).Error; err != nil {
			return err
		}
		return nil
	})
	return
}

func (r *ProductRepositoryImpl) DeleteProduct(productId int) (err error) {
	err = r.db.Delete(&models.Product{}, "id = ?", productId).Error
	return
}

func (r *ProductRepositoryImpl) CheckProductExists(productId int) (exists bool) {
	r.db.Model(&models.Product{}).Select("count(*) > 0").Where("id = ?", productId).Find(&exists)
	return
}

func (r *ProductRepositoryImpl) GetOwnerId(productId int) (userId string) {
	r.db.Model(&models.Product{}).Select("user_id").Where("id = ?", productId).Find(&userId)
	return
}

func (r *ProductRepositoryImpl) InsertWish(wish *models.Wish) (err error) {
	err = r.db.Create(wish).Error
	return
}

func (r *ProductRepositoryImpl) DeleteWish(wish *models.Wish) (err error) {
	err = r.db.Delete(models.Wish{}, "user_id = ? AND product_id = ?", wish.UserID, wish.ProductID).Error
	return
}

func (r *ProductRepositoryImpl) InsertView(view *models.View) (err error) {
	err = r.db.Create(view).Error
	return
}

func (r *ProductRepositoryImpl) CheckWishExists(wish *models.Wish) (exists bool) {
	r.db.Table("wishes").
		Select("count(*) > 0").
		Where("user_id = ? AND product_id = ?", wish.UserID, wish.ProductID).
		Find(&exists)
	return
}
