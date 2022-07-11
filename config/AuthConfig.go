package config

type AuthConfig struct {
    AccessSecret    string      `json:"access_secret"`
    RefreshSecret   string      `json:"refreshSecret"`
}
