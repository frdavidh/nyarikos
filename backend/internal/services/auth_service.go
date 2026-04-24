package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/frdavidh/nyarikos/internal/config"
	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/models"
	"github.com/frdavidh/nyarikos/internal/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrEmailAlreadyExists  = errors.New("email already registered")
	ErrInvalidPassword     = errors.New("invalid password")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRefreshTokenRevoked = errors.New("refresh token is revoked")
	ErrRefreshTokenExpired = errors.New("refresh token is expired")
	ErrUserInactive        = errors.New("user is inactive")
)

type AuthService interface {
	Register(req *dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(req *dto.LoginRequest) (*dto.AuthResponse, error)
	RefreshToken(req *dto.RefreshTokenRequest) (*dto.AuthResponse, error)
	Logout(refreshToken string) error
	GoogleLogin() string
	GoogleCallback(code string) (*dto.AuthResponse, error)
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type oauthExchanger interface {
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)
}

type oauth2Wrapper struct {
	config *oauth2.Config
}

func (w *oauth2Wrapper) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return w.config.Exchange(ctx, code)
}

type authService struct {
	db            *gorm.DB
	config        *config.Config
	httpClient    httpClient
	oauthExchange oauthExchanger
}

func NewAuthService(db *gorm.DB, config *config.Config) AuthService {
	svc := &authService{
		db:         db,
		config:     config,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
	svc.oauthExchange = &oauth2Wrapper{config: svc.oauthConfig()}
	return svc
}

func (s *authService) generateAuthResponse(user *models.User) (*dto.AuthResponse, error) {
	accessToken, refreshToken, err := utils.GenerateTokenPair(
		&s.config.JWT,
		user.ID,
		user.Email,
		string(user.Role),
	)
	if err != nil {
		return nil, err
	}

	refreshTokenRecord := models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(s.config.JWT.RefreshExpiresIn),
	}
	if err := s.db.Create(&refreshTokenRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.UserResponse{
			ID:          user.ID,
			Name:        user.Name,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
			Role:        string(user.Role),
			IsActive:    user.IsActive,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		},
	}, nil
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique constraint failed") || strings.Contains(msg, "duplicate key value violates unique constraint")
}

func (s *authService) Register(req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	hashedPassword, err := utils.HashPassword(req.Password, utils.DefaultParams)
	if err != nil {
		return nil, err
	}

	role := models.RolePencari
	if req.Role == string(models.RolePemilik) {
		role = models.RolePemilik
	}

	user := models.User{
		Email:       req.Email,
		Password:    &hashedPassword,
		Name:        req.Name,
		PhoneNumber: &req.PhoneNumber,
		Role:        role,
	}

	if err := s.db.Create(&user).Error; err != nil {
		if isDuplicateKeyError(err) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return s.generateAuthResponse(&user)
}

func (s *authService) Login(req *dto.LoginRequest) (*dto.AuthResponse, error) {
	var user models.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil, ErrUserNotFound
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	match, err := utils.VerifyPassword(req.Password, *user.Password)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, ErrInvalidPassword
	}

	return s.generateAuthResponse(&user)
}

func (s *authService) RefreshToken(req *dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	claims, err := utils.ValidateToken(req.RefreshToken, s.config.JWT.Secret)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	var refreshToken models.RefreshToken
	if err := s.db.Where("token = ?", req.RefreshToken).First(&refreshToken).Error; err != nil {
		return nil, ErrInvalidRefreshToken
	}

	if refreshToken.IsRevoked != nil && *refreshToken.IsRevoked {
		return nil, ErrRefreshTokenRevoked
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		return nil, ErrRefreshTokenExpired
	}

	var user models.User
	if err := s.db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		return nil, ErrUserNotFound
	}

	if err := s.db.Model(&refreshToken).Update("is_revoked", true).Error; err != nil {
		return nil, err
	}

	return s.generateAuthResponse(&user)
}

func (s *authService) Logout(refreshToken string) error {
	return s.db.Model(&models.RefreshToken{}).Where("token = ?", refreshToken).Update("is_revoked", true).Error
}

func (s *authService) oauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     s.config.OAuth2.ClientID,
		ClientSecret: s.config.OAuth2.ClientSecret,
		RedirectURL:  s.config.OAuth2.RedirectURL,
		Scopes:       s.config.OAuth2.Scope,
		Endpoint:     google.Endpoint,
	}
}

func (s *authService) GoogleLogin() string {
	return s.oauthConfig().AuthCodeURL("state", oauth2.AccessTypeOffline)
}

type googleUserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (s *authService) GoogleCallback(code string) (*dto.AuthResponse, error) {
	token, err := s.oauthExchange.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange oauth code: %w", err)
	}

	userInfo, err := s.fetchGoogleUserInfo(token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch google user info: %w", err)
	}

	if userInfo.ID == "" || userInfo.Email == "" {
		return nil, errors.New("incomplete user info from google")
	}

	var social models.SocialAccount
	err = s.db.Where("provider_name = ? AND provider_id = ?", models.ProviderGoogle, userInfo.ID).First(&social).Error
	if err == nil {
		var user models.User
		if err := s.db.First(&user, social.UserID).Error; err != nil {
			return nil, ErrUserNotFound
		}
		return s.generateAuthResponse(&user)
	}

	var user models.User
	if err := s.db.Where("email = ?", userInfo.Email).First(&user).Error; err != nil {
		user = models.User{
			Email:    userInfo.Email,
			Name:     userInfo.Name,
			Role:     models.RolePencari,
			IsActive: true,
		}
		if err := s.db.Create(&user).Error; err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	social = models.SocialAccount{
		UserID:       user.ID,
		ProviderName: models.ProviderGoogle,
		ProviderID:   userInfo.ID,
	}
	if err := s.db.Create(&social).Error; err != nil {
		return nil, fmt.Errorf("failed to link google account: %w", err)
	}

	return s.generateAuthResponse(&user)
}

func (s *authService) fetchGoogleUserInfo(accessToken string) (*googleUserInfo, error) {
	req, err := http.NewRequest(http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google userinfo returned status %d", resp.StatusCode)
	}

	var userInfo googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}
	return &userInfo, nil
}
