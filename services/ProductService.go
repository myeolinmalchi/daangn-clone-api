package services

import (
	"carrot-market-clone-api/models"
	"carrot-market-clone-api/repositories"
	"gorm.io/gorm"
	"context"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
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

    GetProduct(productId int)               (product *models.Product, err error)

    GetProductsOrderByPrice(asc bool)       GetProductsFunc

    GetProductsOrderByRegdate(asc bool)     GetProductsFunc

    GetUserProductsOrderByPrice(asc bool)   GetUserProductsFunc

    GetUserProductsOrderByRegdate(asc bool) GetUserProductsFunc

    ValidateProduct(prod *models.Product)   (result *models.ProductValidationResult)

    InsertProduct(product *models.Product)  (err error)

    UpdateProduct(product *models.Product)  (err error)

    DeleteProduct(
        userId string, 
        productId int,
    )                                       (err error)

}

type ProductServiceImpl struct {
    productRepo repositories.ProductRepository
    userRepo repositories.UserRepository
    client *s3.Client
}

func NewProductServiceImpl(
    productRepo repositories.ProductRepository,
    userRepo repositories.UserRepository, 
    client *s3.Client,
) ProductSerivce {
    return &ProductServiceImpl{ 
        productRepo: productRepo,
        userRepo: userRepo,
        client: client,
    }
}

func (s *ProductServiceImpl) GetProduct(productId int) (product *models.Product, err error) {
    return s.productRepo.GetProduct(productId)
}

func (s *ProductServiceImpl) GetProductsOrderByPrice(asc bool) GetProductsFunc {
    orderBy := "price"
    if !asc { orderBy = orderBy + " DESC" }
    return func(keyword *string, categoryId *int, last *int, size int) (products []models.Product, count int, err error) {
        return s.productRepo.GetProducts(keyword, categoryId, last, size, orderBy)
    }
}

func (s *ProductServiceImpl) GetProductsOrderByRegdate(asc bool) GetProductsFunc {
    orderBy := "regdate"
    if !asc { orderBy = orderBy + " DESC" }
    return func(keyword *string, categoryId *int, last *int, size int) (products []models.Product, count int, err error) {
        return s.productRepo.GetProducts(keyword, categoryId, last, size, orderBy)
    }
}

func (s *ProductServiceImpl)  GetUserProductsOrderByPrice(asc bool) GetUserProductsFunc {
    orderBy := "price"
    if !asc { orderBy = orderBy + " DESC" }
    return func(userId string, last *int, size int) (products []models.Product, count int, err error) {
        if !s.userRepo.CheckUserExists("id", userId) {
            return nil, 0, gorm.ErrRecordNotFound
        }
        return s.productRepo.GetProductsByUserID(userId, last, size, orderBy)
    }
}

func (s *ProductServiceImpl)  GetUserProductsOrderByRegdate(asc bool) GetUserProductsFunc {
    orderBy := "regdate"
    if !asc { orderBy = orderBy + " DESC" }
    return func(userId string, last *int, size int) (products []models.Product, count int, err error) {
        if !s.userRepo.CheckUserExists("id", userId) {
            return nil, 0, gorm.ErrRecordNotFound
        }
        return s.productRepo.GetProductsByUserID(userId, last, size, orderBy)
    }
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

    checkPrice := func(price int) *string {
        var msg string 
        if price == 0 {
            return nil
        }
        if price < 1000 { 
            msg = "가격은 1000원 이상이어야 합니다."
            return &msg
        }
        return nil
    }

    result = &models.ProductValidationResult {
        Title: checkTitle(product.Title),
        Content: checkContent(product.Content),
        Price: checkPrice(product.Price),
    }    
    return result.GetOrNil()
}

func (s *ProductServiceImpl) InsertProduct(product *models.Product) (err error) {
    if !s.userRepo.CheckUserExists("id", product.UserID) {
        return gorm.ErrRecordNotFound
    }
    err = s.productRepo.InsertProduct(product)
    return
}
func (s *ProductServiceImpl) UpdateProduct(product *models.Product) (err error) {
    err = s.productRepo.UpdateProduct(product)
    return
}
func (s *ProductServiceImpl) DeleteProduct(userId string, productId int) (err error) {
    product, err :=  s.productRepo.GetProduct(productId)
    if err != nil { return }

    err = s.productRepo.DeleteProduct(productId)
    if err != nil { return }

    for _, image := range product.Images {
        //"https://~~~/images/~~~.png"
        filename := strings.Split(image.URL, "/")[4]
        _, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput {
            Bucket: aws.String(os.Getenv("AWS_S3_BUCKET")),
            Key:    aws.String("images/" + filename),
        })
        if err != nil {
            log.Println(err)
        } else {
            log.Println("이미지가 삭제되었습니다: "+filename)
        }
    }
    return
}


