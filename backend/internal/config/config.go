package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	AWS      AWSConfig
	Upload   UploadConfig
	OAuth2   OAuth2Config
	Midtrans MidtransConfig
}
type ServerConfig struct {
	Port    string
	GinMode string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret           string
	ExpiresIn        time.Duration
	RefreshExpiresIn time.Duration
}

type AWSConfig struct {
	Region          string
	AccessKey       string
	SecretAccessKey string
	S3BucketName    string
	S3Endpoint      string
}

type UploadConfig struct {
	Path        string
	MaxFileSize int64
	Provider    string
}

type OAuth2Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scope        []string
	Endpoint     oauth2.Endpoint
}

type MidtransConfig struct {
	ServerKey    string
	ClientKey    string
	MerchantID   string
	IsProduction bool
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	jwtExpiresIn, err := time.ParseDuration(getEnv("JWT_EXPIRES_IN", "24h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRES_IN: %w", err)
	}
	refreshExpiresIn, err := time.ParseDuration(getEnv("JWT_REFRESH_EXPIRES_IN", "72h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_EXPIRES_IN: %w", err)
	}
	maxUploadSize, err := strconv.ParseInt(getEnv("MAX_UPLOAD_SIZE", "10485760"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid MAX_UPLOAD_SIZE: %w", err)
	}

	return &Config{
		Server: ServerConfig{
			Port:    getEnv("PORT", "8080"),
			GinMode: getEnv("GIN_MODE", "release"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "postgres"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:           getEnv("JWT_SECRET", "secret"),
			ExpiresIn:        jwtExpiresIn,
			RefreshExpiresIn: refreshExpiresIn,
		},
		AWS: AWSConfig{
			Region:          getEnv("AWS_REGION", "us-east-1"),
			AccessKey:       getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
			S3BucketName:    getEnv("AWS_S3_BUCKET_NAME", ""),
			S3Endpoint:      getEnv("AWS_S3_ENDPOINT", ""),
		},
		Upload: UploadConfig{
			Path:        getEnv("UPLOAD_PATH", "uploads"),
			MaxFileSize: maxUploadSize,
			Provider:    getEnv("UPLOAD_PROVIDER", "local"),
		},
		OAuth2: OAuth2Config{
			ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", ""),
			Scope:        []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
		Midtrans: MidtransConfig{
			ServerKey:    getEnv("MIDTRANS_SERVER_KEY", ""),
			ClientKey:    getEnv("MIDTRANS_CLIENT_KEY", ""),
			MerchantID:   getEnv("MIDTRANS_MERCHANT_ID", ""),
			IsProduction: getEnv("MIDTRANS_ENV", "sandbox") == "production",
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
