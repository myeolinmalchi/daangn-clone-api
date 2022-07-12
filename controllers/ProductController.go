package controllers

import (
	"carrot-market-clone-api/models"
	"carrot-market-clone-api/services"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"strconv"
    "strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductController interface {
    InsertProduct(c *gin.Context)

    UpdateProduct(c *gin.Context)

    DeleteProduct(c *gin.Context)

    GetProduct(c *gin.Context)
    GetProducts(c *gin.Context)
    GetUserProducts(c *gin.Context)
}

type ProductControllerImpl struct {
    client *s3.Client
    productService services.ProductSerivce
}

func NewProductControllerImpl(
    client *s3.Client,
    productService services.ProductSerivce,
) ProductController {
    return &ProductControllerImpl {
        client: client,
        productService: productService,
    }
}

type ProductForm struct {
    Files       []*multipart.FileHeader         `form:"files" binding:"required"`
    Json        string                          `form:"json" binding:"required"`
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

    result := p.productService.ValidateProduct(product)
    if result != nil {
        c.JSON(422, result)
        return
    }

    uploader := manager.NewUploader(p.client)
    for index, fileHeader := range form.Files {
        file, err := fileHeader.Open()
        if err != nil {
            log.Println(err)
            c.JSON(400, gin.H {
                "message": err.Error(),
            })
            return
        }
        filename := fmt.Sprintf("%s.png", uuid.NewString())
        _, err = uploader.Upload(context.TODO(), &s3.PutObjectInput {
            Bucket: aws.String(os.Getenv("AWS_S3_BUCKET")),
            Key: aws.String("images/"+filename),
            Body: file,
        })

        url := fmt.Sprintf("https://%s/images/%s", os.Getenv("AWS_S3_DOMAIN"), filename)
        product.Images = append(product.Images, models.ProductImage {
            URL: url,
            Sequence: index + 1,
        })
    }

    err := p.productService.InsertProduct(product)

    if err != nil {
        if err == gorm.ErrRecordNotFound { 
            c.Status(404)
        } else {
            c.JSON(400, gin.H{"message": err.Error()})
        }
        for _, image := range product.Images {
            filename := strings.Split("/", image.URL)[4]
            _, err = p.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
                Bucket: aws.String(os.Getenv("AWS_S3_BUCKET")),
                Key: aws.String("images/" + filename),
            })
            if err != nil {
                log.Println(err)
            } else {
                log.Println("이미지가 삭제되었습니다.")
            }
        }
        return
    } else {
        c.Status(201)
    }
}

// PUT api/v1/user/{user_id}/product/{product_id}
func (p *ProductControllerImpl) UpdateProduct(c *gin.Context) {
    userId := c.Param("userId")
    product := &models.Product{}
    if err := c.ShouldBind(product); err != nil {
        c.JSON(400, gin.H{"message": err.Error()})
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

// DELETE api/v1/user/{user_id}/products{product_id}
func (p *ProductControllerImpl) DeleteProduct(c *gin.Context) {
    productId, err := strconv.Atoi(c.Param("productId"))
    if err != nil { c.JSON(400, gin.H{"message": err }); return }

    userId := c.Param("userId")

    err = p.productService.DeleteProduct(userId, productId)
    if err == gorm.ErrRecordNotFound { c.Status(404); return }
    if err != nil { c.JSON(400, gin.H{"message": err}); return}

    c.Status(200)
}

// GET api/v1/products/{product_id}
func (p *ProductControllerImpl) GetProduct(c *gin.Context) {
    productId, err := strconv.Atoi(c.Param("productId"))
    if err != nil { c.JSON(400, gin.H{"message": err }); return }

    product, err := p.productService.GetProduct(productId)
    if err == gorm.ErrRecordNotFound { c.Status(404); return }
    if err != nil { c.JSON(400, gin.H{"message": err}); return}

    c.IndentedJSON(200, product)
}

// GET api/v1/products
// Query String: 
//      keyword(optional) 
//      size(default: 10) 
//      category(optional)
//      last(optional) 
func (p *ProductControllerImpl) GetProducts(c *gin.Context) {
    
    var err error
    var (
        keyword *string
        size int
        category *int
        last *int
    )

    size, err = strconv.Atoi(c.DefaultQuery("size", "10")) 
    if err != nil { c.JSON(400, gin.H{"message": err}); return }

    if keywordStr, keywordExists := c.GetQuery("keyword"); keywordExists {
        keyword = &keywordStr
    } else {
        keyword = nil
    }

    if categoryStr, categoryExists := c.GetQuery("category"); categoryExists {
        temp, err := strconv.Atoi(categoryStr)
        if err != nil { c.JSON(400, gin.H{"message": err}); return }
        category = &temp
    } else {
        category = nil
    }

    if lastStr, lastExists := c.GetQuery("last"); lastExists {
        temp, err := strconv.Atoi(lastStr)
        if err != nil { c.JSON(400, gin.H{"message": err}); return }
        last = &temp
    } else {
        last = nil
    }

    getProductsFuncMap := map[string]services.GetProductsFunc {
        "price": p.productService.GetProductsOrderByPrice(true),
        "pricedesc": p.productService.GetProductsOrderByPrice(false),
        "regdate": p.productService.GetProductsOrderByRegdate(true),
        "regdatedesc": p.productService.GetProductsOrderByRegdate(false),
    }

    var products []models.Product
    if sortStr, sortExists := c.GetQuery("sort"); sortExists {
        if getProductsFunc := getProductsFuncMap[sortStr]; getProductsFunc != nil {
            products, _, err = getProductsFunc(keyword, category, last, size)
        } else {
            products, _, err = getProductsFuncMap["regdatedesc"](keyword, category, last, size)
        }
    } else {
        products, _, err = getProductsFuncMap["regdatedesc"](keyword, category, last, size)
    }
    if err != nil { c.JSON(400, gin.H{"message": err}); return }

    c.IndentedJSON(200, gin.H {
        "size": size,
        "products": products,
    })

}

// GET api/v1/user/{user_id}/products
// Query String: 
//      size(default: 10) 
//      last(optional) 
func (p *ProductControllerImpl) GetUserProducts(c *gin.Context) {
    var err error
    var (
        userId string
        last *int
        size int
    )

    userId = c.Param("userId")

    if lastStr, lastExists := c.GetQuery("last"); lastExists {
        temp, err := strconv.Atoi(lastStr)
        if err != nil { c.JSON(400, gin.H{"message": err}); return }
        last = &temp
    }

    size, err = strconv.Atoi(c.DefaultQuery("size", "10")) 
    if err != nil { c.JSON(400, gin.H{"message": err}); return }

    getUserProductsFuncMap := map[string]services.GetUserProductsFunc {
        "price": p.productService.GetUserProductsOrderByPrice(true),
        "pricedesc": p.productService.GetUserProductsOrderByPrice(false),
        "regdate": p.productService.GetUserProductsOrderByRegdate(true),
        "regdatedesc": p.productService.GetUserProductsOrderByRegdate(false),
    }

    var products []models.Product
    if sortStr, sortExists := c.GetQuery("sort"); sortExists {
        if getProductsFunc := getUserProductsFuncMap[sortStr]; getProductsFunc != nil {
            products, _, err = getProductsFunc(userId, last, size)
        } else {
            products, _, err = getUserProductsFuncMap["regdatedesc"](userId, last, size)
        }
    } else {
        products, _, err = getUserProductsFuncMap["regdatedesc"](userId, last, size)
    }
    if err == gorm.ErrRecordNotFound {
        c.Status(404)
        return
    } else if err != nil {
        c.JSON(400, gin.H{"message": err})
        return 
    }

    c.IndentedJSON(200, gin.H {
        "size": size,
        "userId": userId,
        "products": products,
    })
}

