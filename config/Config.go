package config

import (
	"encoding/json"
	"os"
	"time"
    "context"
    "strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/aws/aws-sdk-go-v2/credentials"
    "github.com/aws/aws-sdk-go-v2/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
    DBConfig        DBConfig        `json:"db"`
    AWSConfig       AWSConfig       `json:"aws"`
    LogConfig       LogConfig       `json:"log"`
    AuthConfig      AuthConfig      `json:"auth"`
}

func LoadConfig() (*Config, error){
    file, err := os.Open("config.json")
    defer file.Close()
    config := &Config{}
    jsonParser := json.NewDecoder(file)
    jsonParser.Decode(config)
    return config, err
}

func LoadTestConfig() (*Config, error){
    file, err := os.Open("../config.json")
    defer file.Close()
    config := &Config{}
    jsonParser := json.NewDecoder(file)
    jsonParser.Decode(config)
    return config, err
}

func (c *Config) InitDBConnection() (db *gorm.DB, err error) {
    dsn := c.DBConfig.ToDNS()
    db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
        NowFunc: func() time.Time {
            ti, _ := time.LoadLocation("Asia/Seoul")
            return time.Now().In(ti)
        },
    })
    return
}

func (c *Config) InitS3Client() (*s3.Client, error) {
    creds := credentials.NewStaticCredentialsProvider(c.AWSConfig.AccessKeyID, c.AWSConfig.SecretKey, "")
    conf, err := config.LoadDefaultConfig(
        context.TODO(),
        config.WithCredentialsProvider(creds),
        config.WithRegion(c.AWSConfig.Region),
    )
    if err != nil { return nil, err }

    return s3.NewFromConfig(conf), nil
}

func (c *Config) InitLogger() (*os.File, error) {
    startTime := time.Now().Format(c.LogConfig.TimeFormat)
    fileName := c.LogConfig.Path + "/" + c.LogConfig.Prefix + "-" + startTime
    return os.Create(strings.TrimSpace(fileName))
}

func (c *Config) InitAuth() {
    os.Setenv("ACCESS_SECRET", c.AuthConfig.AccessSecret)
    os.Setenv("REFRESH_SECRET", c.AuthConfig.RefreshSecret)
}
