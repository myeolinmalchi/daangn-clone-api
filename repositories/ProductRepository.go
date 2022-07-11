package repositories

import (
	"carrot-market-clone-api/models"

	"gorm.io/gorm"
)

type ProductRepository interface {
    GetProduct(postId int)                  (product *models.Product, err error)

    GetProductsByUserID(
        userId string,
        last *int,
        size int,
        orderBy ... string,
    )                                       (products []models.Product, count int, err error)

    GetProducts(
        keyword *string,
        categoryId *int,
        last *int,
        size int,
        orderBy ... string,
    )                                       (products []models.Product, count int, err error)

    InsertProduct(product *models.Product)  (err error)

    UpdateProduct(product *models.Product)  (err error)

    DeleteProduct(productId int)            (err error)

    CheckProductExists(productId int)       (exists bool)

    GetOwnerId(productId int)               (userId string)
}

type ProductRepositoryImpl struct {
    db *gorm.DB
}

func NewProductRepositoryImpl(
    db *gorm.DB,
) ProductRepository {
    return &ProductRepositoryImpl{ db: db }
}

func (r *ProductRepositoryImpl) GetProduct(productId int) (product *models.Product, err error) {
    product = &models.Product{}
    err = r.db.Table("v_products").Preload("Images", func(db *gorm.DB) *gorm.DB {
        return db.Order("product_images.sequence ASC")
    }).Where("id = ?", productId).First(product).Error
    return
}

func (r *ProductRepositoryImpl) GetProductsByUserID(
    userId string,
    last *int,
    size int,
    orderBy ...string,
) (products []models.Product, count int, err error) {

    products = []models.Product{}

    query := r.db.Table("v_products").Omit("Content", "CategoryID", "Views").Where("user_id = ?", userId)

    if(last != nil) {
        query = query.Where("id < ?", last)
    }

    r.db.Table("(?) as a", query).Select("count(*)").Find(&count)
    for _, order := range orderBy {
        query = query.Order(order)
    }
    err = query.Preload("Images", func(db *gorm.DB) *gorm.DB {
        return db.Order("product_images.sequence ASC").Limit(1)
    }).Limit(size).Find(&products).Error

    return
}

func (r *ProductRepositoryImpl) GetProducts(
    keyword *string,
    categoryId *int,
    last *int,
    size int,
    orderBy ... string,
) (products []models.Product, count int, err error) {

    products = []models.Product{}

    query := r.db.Table("v_products").Omit("Content", "CategoryID", "Views")

    if(keyword != nil) {
        query = query.Where("title LIKE ? OR content LIKE ?", "%"+*keyword+"%", "%"+*keyword+"%")
    }

    if(categoryId != nil) {
        query = query.Where("category_id = ?", categoryId)
    }

    if(last != nil) {
        query = query.Where("id < ?", last)
    }

    r.db.Table("(?) as a", query).Select("count(*)").Find(&count)

    for _, order := range orderBy {
        query = query.Order(order)
    }

    err = query.Preload("Images", func(db *gorm.DB) *gorm.DB {
        return db.Order("product_images.sequence ASC").Limit(1)
    }).Limit(size).Find(&products).Error

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

