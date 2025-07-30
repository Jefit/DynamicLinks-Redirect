package config

import (
	"os"
)

type Config struct {
	Port                      string
	PreviewUrlStyle           string
	ExchangeShortLinkEndpoint string
	AppIconImageURL           string
	AppName                   string
	SSLEnabled                string
	SSLCertPath               string
	SSLKeyPath                string
}

func New() *Config {
	return &Config{
		Port:                      getEnv("PORT", "4040"),
		PreviewUrlStyle:           getEnv("PREVIEW_URL_STYLE", "hyphenated"), // hyphenated or subdomain
		ExchangeShortLinkEndpoint: getEnv("EXCHANGE_SHORT_LINK_ENDPOINT", "http://localhost:9010/v1/exchangeShortLink"),
		AppIconImageURL:           getEnv("APP_ICON_IMAGE_URL", "https://example.com/icon.png"),
		AppName:                   getEnv("APP_NAME", "My app name"),
		SSLEnabled:                getEnv("SSL_ENABLED", "false"),
		SSLCertPath:               getEnv("SSL_CERT_PATH", "./tls.crt"),
		SSLKeyPath:                getEnv("SSL_KEY_PATH", "./tls.key"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
