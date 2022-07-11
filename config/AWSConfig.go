package config

type AWSConfig struct {
    AccessKeyID string          `json:"access_key_id"`
    SecretKey   string          `json:"secret_key"`
    Region      string          `json:"region"`
    Bucket      string          `json:"bucket"`
    Domain      string          `json:"domain"`
}
