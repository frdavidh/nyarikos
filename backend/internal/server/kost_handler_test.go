package server

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/interfaces"
	"github.com/frdavidh/nyarikos/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestKostHandler_CreateKost_Success(t *testing.T) {
	router := setupTestRouter()
	mockKost := new(mockKostService)
	mockUpload := new(mockUploadProvider)
	uploadService := services.NewUploadService(mockUpload)
	handler := NewKostHandler(mockKost, uploadService)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	reqBody := dto.CreateKostRequest{
		Name:     "Kost Merdeka",
		Address:  "Jl. Merdeka",
		City:     "Jakarta",
		KostType: "putra",
	}

	mockKost.On("CreateKost", uint(1), &reqBody).Return(&dto.KostResponse{
		ID:       1,
		OwnerID:  1,
		Name:     "Kost Merdeka",
		Address:  "Jl. Merdeka",
		City:     "Jakarta",
		KostType: "putra",
	}, nil)

	w := makeRequest(t, router, "POST", "/api/v1/kost/", reqBody, nil)

	assert.Equal(t, http.StatusCreated, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	mockKost.AssertExpectations(t)
}

func TestKostHandler_CreateKost_InvalidBody(t *testing.T) {
	router := setupTestRouter()
	mockKost := new(mockKostService)
	mockUpload := new(mockUploadProvider)
	uploadService := services.NewUploadService(mockUpload)
	handler := NewKostHandler(mockKost, uploadService)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	w := makeRequest(t, router, "POST", "/api/v1/kost/", "not-json", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestKostHandler_GetAllKost_Success(t *testing.T) {
	router := setupTestRouter()
	mockKost := new(mockKostService)
	mockUpload := new(mockUploadProvider)
	uploadService := services.NewUploadService(mockUpload)
	handler := NewKostHandler(mockKost, uploadService)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	mockKost.On("GetAllKost", 1, 10).Return([]dto.KostResponse{
		{ID: 1, Name: "Kost A"},
		{ID: 2, Name: "Kost B"},
	}, int64(2), nil)

	w := makeRequest(t, router, "GET", "/api/v1/kost/", nil, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	mockKost.AssertExpectations(t)
}

func TestKostHandler_GetKost_Success(t *testing.T) {
	router := setupTestRouter()
	mockKost := new(mockKostService)
	mockUpload := new(mockUploadProvider)
	uploadService := services.NewUploadService(mockUpload)
	handler := NewKostHandler(mockKost, uploadService)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	mockKost.On("GetKost", uint(1)).Return(&dto.KostResponse{
		ID:   1,
		Name: "Kost Merdeka",
	}, nil)

	w := makeRequest(t, router, "GET", "/api/v1/kost/1", nil, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	mockKost.AssertExpectations(t)
}

func TestKostHandler_GetKost_NotFound(t *testing.T) {
	router := setupTestRouter()
	mockKost := new(mockKostService)
	mockUpload := new(mockUploadProvider)
	uploadService := services.NewUploadService(mockUpload)
	handler := NewKostHandler(mockKost, uploadService)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	mockKost.On("GetKost", uint(999)).Return(nil, services.ErrKostNotFound)

	w := makeRequest(t, router, "GET", "/api/v1/kost/999", nil, nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "kost not found", resp["message"])
	mockKost.AssertExpectations(t)
}

func TestKostHandler_GetKost_InvalidID(t *testing.T) {
	router := setupTestRouter()
	mockKost := new(mockKostService)
	mockUpload := new(mockUploadProvider)
	uploadService := services.NewUploadService(mockUpload)
	handler := NewKostHandler(mockKost, uploadService)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	w := makeRequest(t, router, "GET", "/api/v1/kost/invalid", nil, nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.Contains(t, resp["message"], "invalid id ID")
}

func TestKostHandler_UpdateKost_Success(t *testing.T) {
	router := setupTestRouter()
	mockKost := new(mockKostService)
	mockUpload := new(mockUploadProvider)
	uploadService := services.NewUploadService(mockUpload)
	handler := NewKostHandler(mockKost, uploadService)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	newName := "Kost Baru"
	reqBody := dto.UpdateKostRequest{Name: &newName}

	mockKost.On("UpdateKost", uint(1), uint(1), &reqBody).Return(&dto.KostResponse{
		ID:   1,
		Name: "Kost Baru",
	}, nil)

	w := makeRequest(t, router, "PUT", "/api/v1/kost/1", reqBody, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	mockKost.AssertExpectations(t)
}

func TestKostHandler_UpdateKost_Unauthorized(t *testing.T) {
	router := setupTestRouter()
	mockKost := new(mockKostService)
	mockUpload := new(mockUploadProvider)
	uploadService := services.NewUploadService(mockUpload)
	handler := NewKostHandler(mockKost, uploadService)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	newName := "Kost Baru"
	reqBody := dto.UpdateKostRequest{Name: &newName}

	mockKost.On("UpdateKost", uint(1), uint(1), &reqBody).Return(nil, services.ErrUnauthorized)

	w := makeRequest(t, router, "PUT", "/api/v1/kost/1", reqBody, nil)

	assert.Equal(t, http.StatusForbidden, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "you are not allowed to update this kost", resp["message"])
	mockKost.AssertExpectations(t)
}

func TestKostHandler_DeleteKost_Success(t *testing.T) {
	router := setupTestRouter()
	mockKost := new(mockKostService)
	mockUpload := new(mockUploadProvider)
	uploadService := services.NewUploadService(mockUpload)
	handler := NewKostHandler(mockKost, uploadService)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	mockKost.On("DeleteKost", uint(1), uint(1)).Return(&dto.KostResponse{
		ID:   1,
		Name: "Kost Merdeka",
	}, nil)

	w := makeRequest(t, router, "DELETE", "/api/v1/kost/1", nil, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	mockKost.AssertExpectations(t)
}

func TestKostHandler_DeleteKost_NotFound(t *testing.T) {
	router := setupTestRouter()
	mockKost := new(mockKostService)
	mockUpload := new(mockUploadProvider)
	uploadService := services.NewUploadService(mockUpload)
	handler := NewKostHandler(mockKost, uploadService)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	mockKost.On("DeleteKost", uint(999), uint(1)).Return(nil, services.ErrKostNotFound)

	w := makeRequest(t, router, "DELETE", "/api/v1/kost/999", nil, nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "kost not found", resp["message"])
	mockKost.AssertExpectations(t)
}

func TestKostHandler_AddKostImage_InvalidFileType(t *testing.T) {
	router := setupTestRouter()
	mockKost := new(mockKostService)
	mockUpload := new(mockUploadProvider)
	uploadService := services.NewUploadService(mockUpload)
	handler := NewKostHandler(mockKost, uploadService)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	part, _ := writer.CreateFormFile("image", "test.pdf")
	part.Write([]byte("fake content"))
	writer.WriteField("alt_text", "Test Image")
	writer.Close()

	w := makeMultipartRequest(t, router, "POST", "/api/v1/kost/1/images", &b, writer.FormDataContentType(), nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "invalid file type", resp["message"])
}

type mockUploadProvider struct {
	mock.Mock
}

func (m *mockUploadProvider) UploadFile(file *multipart.FileHeader, path string) (string, error) {
	args := m.Called(file, path)
	return args.String(0), args.Error(1)
}

func (m *mockUploadProvider) DeleteFile(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

var _ interfaces.UploadProvider = (*mockUploadProvider)(nil)
