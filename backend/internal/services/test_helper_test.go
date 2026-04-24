package services

import (
	"database/sql"
	"testing"
	"time"

	"github.com/frdavidh/nyarikos/internal/config"
	"github.com/frdavidh/nyarikos/internal/models"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *gorm.DB {
	sqlDB, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.User{},
		&models.Kost{},
		&models.KostImage{},
		&models.Room{},
		&models.Booking{},
		&models.Facility{},
		&models.RoomFacility{},
		&models.RefreshToken{},
		&models.SocialAccount{},
		&models.Payment{},
	)
	require.NoError(t, err)

	return db
}

func testConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret:           "test-secret-key",
			ExpiresIn:        time.Hour,
			RefreshExpiresIn: time.Hour * 24,
		},
		OAuth2: config.OAuth2Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			RedirectURL:  "http://localhost:8080/auth/google/callback",
			Scope:        []string{"openid", "email", "profile"},
		},
		Midtrans: config.MidtransConfig{
			ServerKey:  "SB-Mid-server-test-key",
			ClientKey:  "SB-Mid-client-test-key",
			MerchantID: "G000000000",
		},
	}
}
