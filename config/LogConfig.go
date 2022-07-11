package config

type LogConfig struct {
    Path        string      `json:"path"`
    TimeFormat  string      `json:"time_format"`
    Prefix      string      `json:"prefix"`
}
