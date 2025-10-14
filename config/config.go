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
	EnableFallback            string
	FallbackHost              string
}

func New() *Config {
	return &Config{
		Port:                      getEnv("PORT", "4040"),
		PreviewUrlStyle:           getEnv("PREVIEW_URL_STYLE", "hyphenated"), // hyphenated or subdomain
		ExchangeShortLinkEndpoint: getEnv("EXCHANGE_SHORT_LINK_ENDPOINT", "http://localhost:9010/v1/exchangeShortLink"),
		AppIconImageURL:           getEnv("APP_ICON_IMAGE_URL", "/static/appIcon.svg"),
		AppName:                   getEnv("APP_NAME", "My app name"),
		SSLEnabled:                getEnv("SSL_ENABLED", "false"),
		SSLCertPath:               getEnv("SSL_CERT_PATH", "./tls.crt"),
		SSLKeyPath:                getEnv("SSL_KEY_PATH", "./tls.key"),
		EnableFallback:            getEnv("ENABLE_FALLBACK", "false"),
		FallbackHost:              getEnv("FALLBACK_HOST", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
