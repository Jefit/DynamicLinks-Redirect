package config

import (
	"os"

	_ "github.com/lib/pq"
)

type Config struct {
	Port                      string
	PreviewUrlStyle           string
	ExchangeShortLinkEndpoint string
	AppIconImageURL           string
	AppName                   string
}

func New() *Config {
	return &Config{
		Port:                      getEnv("PORT", "4040"),
		PreviewUrlStyle:           getEnv("PREVIEW_URL_STYLE", "hyphenated"), // hyphenated or subdomain
		ExchangeShortLinkEndpoint: getEnv("EXCHANGE_SHORT_LINK_ENDPOINT", "http://localhost:9010/v1/exchangeShortLink"),
		AppIconImageURL:           getEnv("APP_ICON_IMAGE_URL", "https://example.com/icon.png"),
		AppName:                   getEnv("APP_NAME", "My app name"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
