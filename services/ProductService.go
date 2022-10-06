package services

import (
	"carrot-market-clone-api/models"
	"carrot-market-clone-api/repositories"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"strings"

	"gorm.io/gorm"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type GetProductsFunc func(
	keyword *string,
	categoryId *int,
	last *int,
	size int,
) (products []models.Product, count int, err error)

type GetUserProductsFunc func(
	userId string,
	last *int,
	size int,
) (products []models.Product, count int, err error)

type ProductSerivce interface {
	GetProduct(productId int) (product *models.Product, err error)

	ViewProduct(productId int, ip string) (product *models.Product, err error)

	ViewProductW(productId int, userId, ip string) (product *models.ProductW, err error)

	GetProductsOrderByPrice(asc bool) GetProductsFunc

	GetProductsOrderByID(asc bool) GetProductsFunc

	GetUserProductsOrderByPrice(asc bool) GetUserProductsFunc

	GetUserProductsOrderByID(asc bool) GetUserProductsFunc

	GetWishProducts(
		userId string,
		last *int,
		size int,
	) (products []models.Product, count int, err error)

	ValidateProduct(prod *models.Product) (result *models.ProductValidationResult)

	InsertProduct(
		files []multipart.File,
		product *models.Product,
	) (result *models.ProductValidationResult, err error)

	UpdateProduct(product *models.Product) (err error)

	DeleteProduct(
		userId string,
		productId int,
	) (err error)

	WishProduct(wish *models.Wish) (exists bool, err error)

	DeleteWish(wish *models.Wish) (err error)
}

type ProductServiceImpl struct {
	productRepo repositories.ProductRepository
	userRepo    repositories.UserRepository
	awsService  AWSService
	client      *s3.Client
}

func NewProductServiceImpl(
	productRepo repositories.ProductRepository,
	userRepo repositories.UserRepository,
	awsService AWSService,
	client *s3.Client,
) ProductSerivce {
	return &ProductServiceImpl{
		productRepo: productRepo,
		userRepo:    userRepo,
		awsService:  awsService,
		client:      client,
	}
}

func (s *ProductServiceImpl) GetProduct(productId int) (product *models.Product, err error) {
	product, err = s.productRepo.GetProduct(productId)
	return
}

func (s *ProductServiceImpl) ViewProduct(productId int, ip string) (product *models.Product, err error) {
	product, err = s.productRepo.ViewProduct(productId, ip)
	return
}

func (s *ProductServiceImpl) ViewProductW(productId int, userId, ip string) (product *models.ProductW, err error) {
	product, err = s.productRepo.ViewProductW(productId, userId, ip)
	return
}

func (s *ProductServiceImpl) GetProductsOrderByPrice(asc bool) GetProductsFunc {
	orderBy := "price"
	if !asc {
		orderBy = orderBy + " DESC"
	}
	return func(keyword *string, categoryId *int, last *int, size int) (products []models.Product, count int, err error) {
		return s.productRepo.GetProducts(keyword, categoryId, last, size, orderBy)
	}
}

func (s *ProductServiceImpl) GetProductsOrderByID(asc bool) GetProductsFunc {
	orderBy := "id"
	if !asc {
		orderBy = orderBy + " DESC"
	}
	return func(keyword *string, categoryId *int, last *int, size int) (products []models.Product, count int, err error) {
		return s.productRepo.GetProducts(keyword, categoryId, last, size, orderBy)
	}
}

func (s *ProductServiceImpl) GetUserProductsOrderByPrice(asc bool) GetUserProductsFunc {
	orderBy := "price"
	if !asc {
		orderBy = orderBy + " DESC"
	}
	return func(userId string, last *int, size int) (products []models.Product, count int, err error) {
		if !s.userRepo.CheckUserExists("id", userId) {
			return nil, 0, gorm.ErrRecordNotFound
		}
		return s.productRepo.GetProductsByUserID(userId, last, size, orderBy)
	}
}

func (s *ProductServiceImpl) GetUserProductsOrderByID(asc bool) GetUserProductsFunc {
	orderBy := "id"
	if !asc {
		orderBy = orderBy + " DESC"
	}
	return func(userId string, last *int, size int) (products []models.Product, count int, err error) {
		if !s.userRepo.CheckUserExists("id", userId) {
			return nil, 0, gorm.ErrRecordNotFound
		}
		return s.productRepo.GetProductsByUserID(userId, last, size, orderBy)
	}
}

func (s *ProductServiceImpl) GetWishProducts(
	userId string,
	last *int,
	size int,
) (products []models.Product, count int, err error) {
	return s.productRepo.GetWishProducts(userId, last, size)
}

func (s *ProductServiceImpl) ValidateProduct(product *models.Product) (result *models.ProductValidationResult) {
	checkTitle := func(title string) *string {
		var msg string
		if len(title) > 200 {
			msg = "제목은 200자 이하까지 입력 가능합니다."
			return &msg
		}
		if len(title) < 2 {
			msg = "제목은 2자 이상 입력해야 합니다."
			return &msg
		}
		return nil
	}

	checkContent := func(content string) *string {
		var msg string
		if len(content) > 2000 {
			msg = "내용은 2000자 이하까지 입력 가능합니다."
			return &msg
		}
		if content == "" {
			msg = "내용은 필수 항목입니다."
			return &msg
		}
		return nil
	}

	checkPrice := func(price *int) *string {
		if price == nil {
			return nil
		}
		var msg string
		if *price < 0 {
			msg = "가격은 0원 이상이어야 합니다."
			return &msg
		}
		return nil
	}

	checkCategory := func(categoryId int) *string {
		var msg string
		if categoryId == 0 {
			msg = "카테고리는 필수 항목입니다."
			return &msg
		}
		if !s.productRepo.CheckCorrectCategory(categoryId) {
			msg = "존재하지 않는 카테고리입니다."
			return &msg
		}
		return nil
	}

	result = &models.ProductValidationResult{
		Title:      checkTitle(product.Title),
		Content:    checkContent(product.Content),
		Price:      checkPrice(product.Price),
		CategoryID: checkCategory(product.CategoryID),
	}
	return result.GetOrNil()
}

func (s *ProductServiceImpl) InsertProduct(
	files []multipart.File,
	product *models.Product,
) (result *models.ProductValidationResult, err error) {

	result = s.ValidateProduct(product)

	if result != nil {
		return
	}

	for index, file := range files {
		filename, err := s.awsService.UploadFile(file)
		if err != nil {
			return nil, err
		}

		url := fmt.Sprintf("https://%s/images/%s", os.Getenv("AWS_S3_DOMAIN"), filename)
		product.Images = append(product.Images, models.ProductImage{
			URL:      url,
			Sequence: index + 1,
		})
	}

	err = s.productRepo.InsertProduct(product)

	if err != nil {
		for _, image := range product.Images {
			filename := strings.Split("/", image.URL)[4]
			err := s.awsService.DeleteFile(filename)
			if err != nil {
				log.Println(err)
			}
		}
	}

	return
}

// TODO: 구현
func (s *ProductServiceImpl) UpdateProduct(product *models.Product) (err error) {
	err = s.productRepo.UpdateProduct(product)
	return
}

func (s *ProductServiceImpl) DeleteProduct(userId string, productId int) (err error) {
	product, err := s.productRepo.GetProduct(productId)
	if err != nil {
		return
	}

	err = s.productRepo.DeleteProduct(productId)
	if err != nil {
		return
	}

	for _, image := range product.Images {
		//"https://~~~/images/~~~.png"
		filename := strings.Split(image.URL, "/")[4]
		err := s.awsService.DeleteFile(filename)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("이미지가 삭제되었습니다: " + filename)
		}
	}
	return
}

func (s *ProductServiceImpl) WishProduct(wish *models.Wish) (exists bool, err error) {
	exists = s.productRepo.CheckWishExists(wish)
	if exists {
		return
	}

	if userExists := s.userRepo.CheckUserExists("id", wish.UserID); !userExists {
		err = gorm.ErrRecordNotFound
		return
	}

	if productExists := s.productRepo.CheckProductExists(wish.ProductID); !productExists {
		err = gorm.ErrRecordNotFound
		return
	}

	err = s.productRepo.InsertWish(wish)
	return
}

func (s *ProductServiceImpl) DeleteWish(wish *models.Wish) (err error) {
	err = s.productRepo.DeleteWish(wish)
	return
}
