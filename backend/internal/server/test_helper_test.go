package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func makeRequest(t *testing.T, router *gin.Engine, method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, path, reqBody)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func makeMultipartRequest(t *testing.T, router *gin.Engine, method, path string, body *bytes.Buffer, contentType string, headers map[string]string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", contentType)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func parseResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	return resp
}

type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) Register(req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *mockAuthService) Login(req *dto.LoginRequest) (*dto.AuthResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *mockAuthService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *mockAuthService) Logout(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *mockAuthService) GoogleLogin(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *mockAuthService) GoogleCallback(ctx context.Context, code, state string) (*dto.AuthResponse, error) {
	args := m.Called(ctx, code, state)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

var _ services.AuthService = (*mockAuthService)(nil)

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) GetProfile(userID uint) (*dto.UserResponse, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}

func (m *mockUserService) UpdateProfile(userID uint, req *dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}

var _ services.UserService = (*mockUserService)(nil)

type mockKostService struct {
	mock.Mock
}

func (m *mockKostService) CreateKost(ownerID uint, req *dto.CreateKostRequest) (*dto.KostResponse, error) {
	args := m.Called(ownerID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.KostResponse), args.Error(1)
}

func (m *mockKostService) UpdateKost(kostID, userID uint, req *dto.UpdateKostRequest) (*dto.KostResponse, error) {
	args := m.Called(kostID, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.KostResponse), args.Error(1)
}

func (m *mockKostService) DeleteKost(kostID, userID uint) (*dto.KostResponse, error) {
	args := m.Called(kostID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.KostResponse), args.Error(1)
}

func (m *mockKostService) GetAllKost(req *dto.SearchKostRequest) ([]dto.KostResponse, int64, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]dto.KostResponse), args.Get(1).(int64), args.Error(2)
}

func (m *mockKostService) GetKost(id uint) (*dto.KostResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.KostResponse), args.Error(1)
}

func (m *mockKostService) AddKostImage(kostID, userID uint, url, altText string) error {
	args := m.Called(kostID, userID, url, altText)
	return args.Error(0)
}

var _ services.KostService = (*mockKostService)(nil)

type mockRoomService struct {
	mock.Mock
}

func (m *mockRoomService) CreateRoom(kostID uint, req *dto.CreateRoomRequest) (*dto.RoomResponse, error) {
	args := m.Called(kostID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RoomResponse), args.Error(1)
}

func (m *mockRoomService) UpdateRoom(roomID uint, req *dto.UpdateRoomRequest) (*dto.RoomResponse, error) {
	args := m.Called(roomID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RoomResponse), args.Error(1)
}

func (m *mockRoomService) DeleteRoom(roomID uint) error {
	args := m.Called(roomID)
	return args.Error(0)
}

func (m *mockRoomService) GetRoomByID(roomID uint) (*dto.RoomResponse, error) {
	args := m.Called(roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RoomResponse), args.Error(1)
}

func (m *mockRoomService) GetRoomByKostID(kostID uint) ([]dto.RoomResponse, error) {
	args := m.Called(kostID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.RoomResponse), args.Error(1)
}

func (m *mockRoomService) GetAllFacilities() ([]dto.FacilityResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.FacilityResponse), args.Error(1)
}

func (m *mockRoomService) CreateFacility(req *dto.CreateFacilityRequest) (*dto.FacilityResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.FacilityResponse), args.Error(1)
}

func (m *mockRoomService) UpdateFacility(facilityID uint, req *dto.UpdateFacilityRequest) (*dto.FacilityResponse, error) {
	args := m.Called(facilityID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.FacilityResponse), args.Error(1)
}

func (m *mockRoomService) DeleteFacility(facilityID uint) error {
	args := m.Called(facilityID)
	return args.Error(0)
}

func (m *mockRoomService) CreateRoomFacility(roomID uint, req *dto.CreateRoomFacilityRequest) (*dto.RoomFacilityResponse, error) {
	args := m.Called(roomID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RoomFacilityResponse), args.Error(1)
}

func (m *mockRoomService) DeleteRoomFacility(roomID uint, facilityID uint) error {
	args := m.Called(roomID, facilityID)
	return args.Error(0)
}

var _ services.RoomService = (*mockRoomService)(nil)

type mockBookingService struct {
	mock.Mock
}

func (m *mockBookingService) CreateBooking(ctx context.Context, userID uint, req *dto.CreateBookingRequest) (*dto.BookingResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.BookingResponse), args.Error(1)
}

var _ services.BookingService = (*mockBookingService)(nil)

type mockPaymentService struct {
	mock.Mock
}

func (m *mockPaymentService) CreatePayment(userID uint, req *dto.CreatePaymentRequest) (*dto.PaymentResponse, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaymentResponse), args.Error(1)
}

func (m *mockPaymentService) HandleWebhook(req *dto.MidtransWebhookRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

var _ services.PaymentService = (*mockPaymentService)(nil)

func setAuthContext(c *gin.Context, userID uint, role string) {
	c.Set("user_id", userID)
	c.Set("email", "test@example.com")
	c.Set("role", role)
}

func authMiddlewareForTest(userID uint, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		setAuthContext(c, userID, role)
		c.Next()
	}
}

func strPtr(s string) *string {
	return &s
}
