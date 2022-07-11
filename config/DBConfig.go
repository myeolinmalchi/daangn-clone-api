package config

import (
    "fmt"
)

type DBConfig struct {
    User        string      `json:"user"`
    Password    string      `json:"password"`
    Host        string      `json:"host"`
    Port        int         `json:"port"`
    Database    string      `json:"database"`
}

func (c *DBConfig) ToDNS() string {
    return fmt.Sprintf(
        "%s:%s@tcp(%s:%d)/%s?parseTime=true",
        c.User,
        c.Password,
        c.Host,
        c.Port,
        c.Database,
    )
}
