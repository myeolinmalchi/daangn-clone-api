package services

import (
	"mime/multipart"
    "os"
    "context"

	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
)

type AWSService interface {
    UploadFile(file multipart.File)                     (filename string, err error)
    DeleteFile(filename string)                         (err error)
}

type AWSServiceImpl struct {
    client *s3.Client
}

func NewAWSServiceImpl(client *s3.Client) AWSService {
    return &AWSServiceImpl{ client: client }
}

func (s *AWSServiceImpl) UploadFile(file multipart.File) (filename string, err error) {
    filename = uuid.NewString() + ".png"
    uploader := manager.NewUploader(s.client)
    _, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
        Bucket: aws.String(os.Getenv("AWS_S3_BUCKET")),
        Key: aws.String("images/"+filename),
        Body: file,
    })
    return
}

func (s *AWSServiceImpl) DeleteFile(filename string) (err error) {
    _, err = s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput {
        Bucket: aws.String(os.Getenv("AWS_S3_BUCKET")),
        Key: aws.String("images/"+filename),
    })
    return
}
