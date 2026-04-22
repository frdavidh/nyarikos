package server

import (
	"net/http"
	"testing"

	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/services"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRoomHandler_CreateRoom_Success(t *testing.T) {
	router := setupTestRouter()
	mockRoom := new(mockRoomService)
	handler := NewRoomHandler(mockRoom)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	reqBody := dto.CreateRoomRequest{
		RoomType:      "Standard",
		PricePerMonth: decimal.NewFromInt(1000000),
		TotalRooms:    5,
	}

	mockRoom.On("CreateRoom", uint(1), &reqBody).Return(&dto.RoomResponse{
		ID:            1,
		KostID:        1,
		RoomType:      "Standard",
		PricePerMonth: decimal.NewFromInt(1000000),
		TotalRooms:    5,
	}, nil)

	w := makeRequest(t, router, "POST", "/api/v1/kost/1/room/", reqBody, nil)

	assert.Equal(t, http.StatusCreated, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	mockRoom.AssertExpectations(t)
}

func TestRoomHandler_CreateRoom_InvalidBody(t *testing.T) {
	router := setupTestRouter()
	mockRoom := new(mockRoomService)
	handler := NewRoomHandler(mockRoom)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	w := makeRequest(t, router, "POST", "/api/v1/kost/1/room/", "not-json", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRoomHandler_GetRoomByKostID_Success(t *testing.T) {
	router := setupTestRouter()
	mockRoom := new(mockRoomService)
	handler := NewRoomHandler(mockRoom)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	mockRoom.On("GetRoomByKostID", uint(1)).Return([]dto.RoomResponse{
		{ID: 1, RoomType: "Standard"},
		{ID: 2, RoomType: "Deluxe"},
	}, nil)

	w := makeRequest(t, router, "GET", "/api/v1/kost/1/room/", nil, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	mockRoom.AssertExpectations(t)
}

func TestRoomHandler_GetRoomByID_Success(t *testing.T) {
	router := setupTestRouter()
	mockRoom := new(mockRoomService)
	handler := NewRoomHandler(mockRoom)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	mockRoom.On("GetRoomByID", uint(1)).Return(&dto.RoomResponse{
		ID:       1,
		RoomType: "Standard",
	}, nil)

	w := makeRequest(t, router, "GET", "/api/v1/kost/1/room/1", nil, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	mockRoom.AssertExpectations(t)
}

func TestRoomHandler_GetRoomByID_NotFound(t *testing.T) {
	router := setupTestRouter()
	mockRoom := new(mockRoomService)
	handler := NewRoomHandler(mockRoom)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	mockRoom.On("GetRoomByID", uint(999)).Return(nil, services.ErrRoomNotFound)

	w := makeRequest(t, router, "GET", "/api/v1/kost/1/room/999", nil, nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "room not found", resp["message"])
	mockRoom.AssertExpectations(t)
}

func TestRoomHandler_GetRoomByID_InvalidID(t *testing.T) {
	router := setupTestRouter()
	mockRoom := new(mockRoomService)
	handler := NewRoomHandler(mockRoom)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	w := makeRequest(t, router, "GET", "/api/v1/kost/1/room/invalid", nil, nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRoomHandler_UpdateRoom_Success(t *testing.T) {
	router := setupTestRouter()
	mockRoom := new(mockRoomService)
	handler := NewRoomHandler(mockRoom)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	reqBody := dto.UpdateRoomRequest{
		RoomType: "Deluxe",
	}

	mockRoom.On("UpdateRoom", uint(1), mock.AnythingOfType("*dto.UpdateRoomRequest")).Return(&dto.RoomResponse{
		ID:       1,
		RoomType: "Deluxe",
	}, nil)

	w := makeRequest(t, router, "PUT", "/api/v1/kost/1/room/1", reqBody, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	mockRoom.AssertExpectations(t)
}

func TestRoomHandler_UpdateRoom_NotFound(t *testing.T) {
	router := setupTestRouter()
	mockRoom := new(mockRoomService)
	handler := NewRoomHandler(mockRoom)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	reqBody := dto.UpdateRoomRequest{
		RoomType: "Deluxe",
	}

	mockRoom.On("UpdateRoom", uint(999), mock.AnythingOfType("*dto.UpdateRoomRequest")).Return(nil, services.ErrRoomNotFound)

	w := makeRequest(t, router, "PUT", "/api/v1/kost/1/room/999", reqBody, nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "room not found", resp["message"])
	mockRoom.AssertExpectations(t)
}

func TestRoomHandler_DeleteRoom_Success(t *testing.T) {
	router := setupTestRouter()
	mockRoom := new(mockRoomService)
	handler := NewRoomHandler(mockRoom)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	mockRoom.On("DeleteRoom", uint(1)).Return(nil)

	w := makeRequest(t, router, "DELETE", "/api/v1/kost/1/room/1", nil, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	mockRoom.AssertExpectations(t)
}

func TestRoomHandler_DeleteRoom_NotFound(t *testing.T) {
	router := setupTestRouter()
	mockRoom := new(mockRoomService)
	handler := NewRoomHandler(mockRoom)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	mockRoom.On("DeleteRoom", uint(999)).Return(services.ErrRoomNotFound)

	w := makeRequest(t, router, "DELETE", "/api/v1/kost/1/room/999", nil, nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "room not found", resp["message"])
	mockRoom.AssertExpectations(t)
}

func TestRoomHandler_GetAllFacilities_Success(t *testing.T) {
	router := setupTestRouter()
	mockRoom := new(mockRoomService)
	handler := NewRoomHandler(mockRoom)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	mockRoom.On("GetAllFacilities").Return([]dto.FacilityResponse{
		{ID: 1, Name: "AC"},
		{ID: 2, Name: "WiFi"},
	}, nil)

	w := makeRequest(t, router, "GET", "/api/v1/facilities/", nil, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	mockRoom.AssertExpectations(t)
}

func TestRoomHandler_CreateFacility_Success(t *testing.T) {
	router := setupTestRouter()
	mockRoom := new(mockRoomService)
	handler := NewRoomHandler(mockRoom)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	reqBody := dto.CreateFacilityRequest{Name: "AC"}

	mockRoom.On("CreateFacility", &reqBody).Return(&dto.FacilityResponse{
		ID:   1,
		Name: "AC",
	}, nil)

	w := makeRequest(t, router, "POST", "/api/v1/facilities/", reqBody, nil)

	assert.Equal(t, http.StatusCreated, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	mockRoom.AssertExpectations(t)
}

func TestRoomHandler_UpdateFacility_Success(t *testing.T) {
	router := setupTestRouter()
	mockRoom := new(mockRoomService)
	handler := NewRoomHandler(mockRoom)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	reqBody := dto.UpdateFacilityRequest{Name: "Air Conditioner"}

	mockRoom.On("UpdateFacility", uint(1), &reqBody).Return(&dto.FacilityResponse{
		ID:   1,
		Name: "Air Conditioner",
	}, nil)

	w := makeRequest(t, router, "PUT", "/api/v1/facilities/1", reqBody, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	mockRoom.AssertExpectations(t)
}

func TestRoomHandler_UpdateFacility_NotFound(t *testing.T) {
	router := setupTestRouter()
	mockRoom := new(mockRoomService)
	handler := NewRoomHandler(mockRoom)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	reqBody := dto.UpdateFacilityRequest{Name: "AC"}

	mockRoom.On("UpdateFacility", uint(999), &reqBody).Return(nil, services.ErrFacilityNotFound)

	w := makeRequest(t, router, "PUT", "/api/v1/facilities/999", reqBody, nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "facility not found", resp["message"])
	mockRoom.AssertExpectations(t)
}

func TestRoomHandler_DeleteFacility_Success(t *testing.T) {
	router := setupTestRouter()
	mockRoom := new(mockRoomService)
	handler := NewRoomHandler(mockRoom)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	mockRoom.On("DeleteFacility", uint(1)).Return(nil)

	w := makeRequest(t, router, "DELETE", "/api/v1/facilities/1", nil, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "Facility deleted successfully", resp["message"])
	mockRoom.AssertExpectations(t)
}

func TestRoomHandler_DeleteFacility_NotFound(t *testing.T) {
	router := setupTestRouter()
	mockRoom := new(mockRoomService)
	handler := NewRoomHandler(mockRoom)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	mockRoom.On("DeleteFacility", uint(999)).Return(services.ErrFacilityNotFound)

	w := makeRequest(t, router, "DELETE", "/api/v1/facilities/999", nil, nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "facility not found", resp["message"])
	mockRoom.AssertExpectations(t)
}

func TestRoomHandler_CreateRoomFacility_Success(t *testing.T) {
	router := setupTestRouter()
	mockRoom := new(mockRoomService)
	handler := NewRoomHandler(mockRoom)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	reqBody := dto.CreateRoomFacilityRequest{
		RoomID:     1,
		FacilityID: 2,
	}

	mockRoom.On("CreateRoomFacility", uint(1), &reqBody).Return(&dto.RoomFacilityResponse{
		RoomID:     1,
		FacilityID: 2,
	}, nil)

	w := makeRequest(t, router, "POST", "/api/v1/room-facilities/", reqBody, nil)

	assert.Equal(t, http.StatusCreated, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp["success"].(bool))
	mockRoom.AssertExpectations(t)
}

func TestRoomHandler_CreateRoomFacility_RoomNotFound(t *testing.T) {
	router := setupTestRouter()
	mockRoom := new(mockRoomService)
	handler := NewRoomHandler(mockRoom)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	reqBody := dto.CreateRoomFacilityRequest{
		RoomID:     999,
		FacilityID: 2,
	}

	mockRoom.On("CreateRoomFacility", uint(999), &reqBody).Return(nil, services.ErrRoomNotFound)

	w := makeRequest(t, router, "POST", "/api/v1/room-facilities/", reqBody, nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "room not found", resp["message"])
	mockRoom.AssertExpectations(t)
}

func TestRoomHandler_DeleteRoomFacility_Success(t *testing.T) {
	router := setupTestRouter()
	mockRoom := new(mockRoomService)
	handler := NewRoomHandler(mockRoom)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	reqBody := dto.DeleteRoomFacilityRequest{
		RoomID:     1,
		FacilityID: 2,
	}

	mockRoom.On("DeleteRoomFacility", uint(1), uint(2)).Return(nil)

	w := makeRequest(t, router, "DELETE", "/api/v1/room-facilities/", reqBody, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "Room facility deleted successfully", resp["message"])
	mockRoom.AssertExpectations(t)
}

func TestRoomHandler_DeleteRoomFacility_NotFound(t *testing.T) {
	router := setupTestRouter()
	mockRoom := new(mockRoomService)
	handler := NewRoomHandler(mockRoom)

	api := router.Group("/api/v1")
	handler.Routes(api, authMiddlewareForTest(1, "pemilik"))

	reqBody := dto.DeleteRoomFacilityRequest{
		RoomID:     999,
		FacilityID: 2,
	}

	mockRoom.On("DeleteRoomFacility", uint(999), uint(2)).Return(services.ErrRoomNotFound)

	w := makeRequest(t, router, "DELETE", "/api/v1/room-facilities/", reqBody, nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseResponse(t, w)
	assert.Equal(t, "room not found", resp["message"])
	mockRoom.AssertExpectations(t)
}
