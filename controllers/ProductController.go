package controllers

import (
	"carrot-market-clone-api/models"
	"carrot-market-clone-api/services"
	"encoding/json"
	"log"
	"mime/multipart"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductController interface {
	InsertProduct(c *gin.Context)

	UpdateProduct(c *gin.Context)

	DeleteProduct(c *gin.Context)

	GetProduct(c *gin.Context)

	GetProductW(c *gin.Context)

	GetProducts(c *gin.Context)

	GetUserProducts(c *gin.Context)

	GetWishProducts(c *gin.Context)

	WishProduct(c *gin.Context)

	DeleteWish(c *gin.Context)
}

type ProductControllerImpl struct {
	client         *s3.Client
	productService services.ProductSerivce
	chatService    services.ChatService
}

func NewProductControllerImpl(
	client *s3.Client,
	productService services.ProductSerivce,
) ProductController {
	return &ProductControllerImpl{
		client:         client,
		productService: productService,
	}
}

type ProductForm struct {
	Files []*multipart.FileHeader `form:"files" binding:"omitempty"`
	Json  string                  `form:"json" binding:"required"`
}

// POST api/v1/user/{user_id}/products
func (p *ProductControllerImpl) InsertProduct(c *gin.Context) {
	form := ProductForm{}
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}

	userId := c.Param("userId")

	product := &models.Product{}
	json.Unmarshal([]byte(form.Json), product)

	if product.UserID != userId {
		c.Status(403)
		return
	}

	files := []multipart.File{}
	for _, fileHeader := range form.Files {
		file, err := fileHeader.Open()
		if err != nil {
			log.Println(err)
			c.JSON(400, gin.H{"message": err})
			return
		}
		files = append(files, file)
	}

	validationResult, err := p.productService.InsertProduct(files, product)

	if validationResult != nil {
		c.IndentedJSON(422, validationResult)
		return
	}

	if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	c.Status(201)
}

// PUT api/v1/user/{user_id}/product/{product_id}
// TODO: 구현
func (p *ProductControllerImpl) UpdateProduct(c *gin.Context) {
	userId := c.Param("userId")
	product := &models.Product{}
	if err := c.ShouldBind(product); err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	if product.UserID != userId {
		c.Status(403)
		return
	}

	_, err := p.productService.GetProduct(product.ID)

	if err == gorm.ErrRecordNotFound {
		c.Status(404)
		return
	} else if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

}

// DELETE api/v1/user/{user_id}/products/{product_id}
func (p *ProductControllerImpl) DeleteProduct(c *gin.Context) {
	productId, err := strconv.Atoi(c.Param("productId"))
	if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	userId := c.Param("userId")

	err = p.productService.DeleteProduct(userId, productId)
	if err == gorm.ErrRecordNotFound {
		c.Status(404)
		return
	}
	if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	c.Status(200)
}

// GET api/v1/products/{product_id}
func (p *ProductControllerImpl) GetProduct(c *gin.Context) {
	productId, err := strconv.Atoi(c.Param("productId"))
	ip := c.ClientIP()
	if err != nil {
		c.JSON(400, gin.H{"message": "productId는 정수값이어야 합니다."})
		return
	}

	product, err := p.productService.ViewProduct(productId, ip)
	if err == gorm.ErrRecordNotFound {
		c.Status(404)
		return
	}
	if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	c.IndentedJSON(200, product)
}

// GET /api/v1/users/{userId}/products/{productId}
func (p *ProductControllerImpl) GetProductW(c *gin.Context) {
	userId := c.Param("userId")
	productId, err := strconv.Atoi(c.Param("productId"))
	ip := c.ClientIP()
	if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	product, err := p.productService.ViewProductW(productId, userId, ip)

	if err == gorm.ErrRecordNotFound {
		c.Status(404)
		return
	}
	if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	c.IndentedJSON(200, product)
}

// GET api/v1/products
// Query String:
//
//	keyword(optional)
//	size(default: 10)
//	category(optional)
//	last(optional)
func (p *ProductControllerImpl) GetProducts(c *gin.Context) {

	var err error
	var (
		keyword  *string
		size     int
		category *int
		last     *int
	)

	size, err = strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	if keywordStr, keywordExists := c.GetQuery("keyword"); keywordExists {
		keyword = &keywordStr
	} else {
		keyword = nil
	}

	if categoryStr, categoryExists := c.GetQuery("category"); categoryExists {
		temp, err := strconv.Atoi(categoryStr)
		if err != nil {
			c.JSON(400, gin.H{"message": err})
			return
		}
		category = &temp
	} else {
		category = nil
	}

	if lastStr, lastExists := c.GetQuery("last"); lastExists {
		temp, err := strconv.Atoi(lastStr)
		if err != nil {
			c.JSON(400, gin.H{"message": err})
			return
		}
		last = &temp
	} else {
		last = nil
	}

	getProductsFuncMap := map[string]services.GetProductsFunc{
		"price":     p.productService.GetProductsOrderByPrice(true),
		"pricedesc": p.productService.GetProductsOrderByPrice(false),
		"id":        p.productService.GetProductsOrderByID(true),
		"iddesc":    p.productService.GetProductsOrderByID(false),
	}

	var products []models.Product
	var count int
	if sortStr, sortExists := c.GetQuery("sort"); sortExists {
		if getProductsFunc := getProductsFuncMap[sortStr]; getProductsFunc != nil {
			products, count, err = getProductsFunc(keyword, category, last, size)
		} else {
			products, count, err = getProductsFuncMap["iddesc"](keyword, category, last, size)
		}
	} else {
		products, count, err = getProductsFuncMap["iddesc"](keyword, category, last, size)
	}
	if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	c.IndentedJSON(200, gin.H{
		"size":     count,
		"products": products,
	})

}

// GET api/v1/user/{user_id}/products
// Query String:
//
//	size(default: 10)
//	last(optional)
func (p *ProductControllerImpl) GetUserProducts(c *gin.Context) {
	var err error
	var (
		userId string
		last   *int
		size   int
	)

	userId = c.Param("userId")

	if lastStr, lastExists := c.GetQuery("last"); lastExists {
		temp, err := strconv.Atoi(lastStr)
		if err != nil {
			c.JSON(400, gin.H{"message": err})
			return
		}
		last = &temp
	}

	size, err = strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	getUserProductsFuncMap := map[string]services.GetUserProductsFunc{
		"price":     p.productService.GetUserProductsOrderByPrice(true),
		"pricedesc": p.productService.GetUserProductsOrderByPrice(false),
		"id":        p.productService.GetUserProductsOrderByID(true),
		"iddesc":    p.productService.GetUserProductsOrderByID(false),
	}

	var products []models.Product
	var count int
	if sortStr, sortExists := c.GetQuery("sort"); sortExists {
		if getProductsFunc := getUserProductsFuncMap[sortStr]; getProductsFunc != nil {
			products, count, err = getProductsFunc(userId, last, size)
		} else {
			products, count, err = getUserProductsFuncMap["iddesc"](userId, last, size)
		}
	} else {
		products, count, err = getUserProductsFuncMap["iddesc"](userId, last, size)
	}
	if err == gorm.ErrRecordNotFound {
		c.Status(404)
		return
	} else if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	c.IndentedJSON(200, gin.H{
		"size":     count,
		"userId":   userId,
		"products": products,
	})
}

// GET /api/v1/users/{userId}/products_wish
func (p *ProductControllerImpl) GetWishProducts(c *gin.Context) {
	var err error
	var (
		userId string
		last   *int
		size   int
	)

	userId = c.Param("userId")

	if lastStr, lastExists := c.GetQuery("last"); lastExists {
		temp, err := strconv.Atoi(lastStr)
		if err != nil {
			c.JSON(400, gin.H{"message": err})
			return
		}
		last = &temp
	}

	size, err = strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	products, count, err := p.productService.GetWishProducts(userId, last, size)

	if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	c.IndentedJSON(200, gin.H{
		"size":     count,
		"userId":   userId,
		"products": products,
	})
}

// POST /api/v1/users/{userId}/products/{productId}/wish
func (p *ProductControllerImpl) WishProduct(c *gin.Context) {

	userId := c.Param("userId")
	productId, err := strconv.Atoi(c.Param("productId"))

	if err != nil {
		c.JSON(400, gin.H{"message": "productId는 정수값이어야 합니다."})
		return
	}

	exists, err := p.productService.WishProduct(&models.Wish{
		UserID:    userId,
		ProductID: productId,
	})

	// Wish already exists
	if exists {
		c.Status(403)
		return
	}

	if err == gorm.ErrRecordNotFound {
		c.Status(404)
		return
	}

	if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	c.Status(200)
}

// DELETE /api/v1/users/{userId}/products/{productId}/wish
func (p *ProductControllerImpl) DeleteWish(c *gin.Context) {

	userId := c.Param("userId")
	productId, err := strconv.Atoi(c.Param("productId"))

	if err != nil {
		c.JSON(400, gin.H{"message": "productId는 정수값이어야 합니다."})
		return
	}
	err = p.productService.DeleteWish(&models.Wish{
		UserID:    userId,
		ProductID: productId,
	})

	if err == gorm.ErrRecordNotFound {
		c.Status(404)
		return
	}

	if err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}

	c.Status(200)
}
