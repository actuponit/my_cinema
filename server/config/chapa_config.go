package config

import "os"

type ChapaConfig struct {
	SecretKey string
	BaseURL   string
}

func NewChapaConfig() *ChapaConfig {
	return &ChapaConfig{
		SecretKey: os.Getenv("CHAPA_SECRET_KEY"),
		BaseURL:   "https://api.chapa.co/v1",
	}
}
