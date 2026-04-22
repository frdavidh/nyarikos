package services

import (
	"testing"

	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoomService_CreateRoom_Success(t *testing.T) {
	db := setupTestDB(t)
	service := NewRoomService(db)

	req := &dto.CreateRoomRequest{
		RoomType:      "Standard",
		PricePerMonth: decimal.NewFromInt(1000000),
		TotalRooms:    5,
	}

	resp, err := service.CreateRoom(1, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, req.RoomType, resp.RoomType)
	assert.Equal(t, req.TotalRooms, resp.TotalRooms)
	assert.Equal(t, uint(1), resp.KostID)
	assert.NotZero(t, resp.ID)
}

func TestRoomService_CreateRoom_WithFacilities(t *testing.T) {
	db := setupTestDB(t)
	service := NewRoomService(db)

	facility := models.Facility{Name: "AC"}
	err := db.Create(&facility).Error
	require.NoError(t, err)

	req := &dto.CreateRoomRequest{
		RoomType:      "Deluxe",
		PricePerMonth: decimal.NewFromInt(1500000),
		TotalRooms:    3,
		FacilityIDs:   []uint{facility.ID},
	}

	resp, err := service.CreateRoom(1, req)

	assert.NoError(t, err)
	assert.Len(t, resp.Facilities, 1)
	assert.Equal(t, "AC", resp.Facilities[0].Name)
}

func TestRoomService_GetRoomByID_Success(t *testing.T) {
	db := setupTestDB(t)
	service := NewRoomService(db)

	room := models.Room{
		KostID:        1,
		RoomType:      "Standard",
		PricePerMonth: decimal.NewFromInt(1000000),
		TotalRooms:    5,
	}
	err := db.Create(&room).Error
	require.NoError(t, err)

	resp, err := service.GetRoomByID(room.ID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, room.RoomType, resp.RoomType)
}

func TestRoomService_GetRoomByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewRoomService(db)

	resp, err := service.GetRoomByID(999)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrRoomNotFound)
}

func TestRoomService_GetRoomByKostID(t *testing.T) {
	db := setupTestDB(t)
	service := NewRoomService(db)

	for i := 0; i < 3; i++ {
		room := models.Room{
			KostID:        1,
			RoomType:      "Standard",
			PricePerMonth: decimal.NewFromInt(1000000),
			TotalRooms:    2,
		}
		err := db.Create(&room).Error
		require.NoError(t, err)
	}

	resp, err := service.GetRoomByKostID(1)

	assert.NoError(t, err)
	assert.Len(t, resp, 3)
}

func TestRoomService_UpdateRoom_Success(t *testing.T) {
	db := setupTestDB(t)
	service := NewRoomService(db)

	room := models.Room{
		KostID:        1,
		RoomType:      "Standard",
		PricePerMonth: decimal.NewFromInt(1000000),
		TotalRooms:    5,
	}
	err := db.Create(&room).Error
	require.NoError(t, err)

	newType := "Deluxe"
	req := &dto.UpdateRoomRequest{
		RoomType: newType,
	}

	resp, err := service.UpdateRoom(room.ID, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Deluxe", resp.RoomType)
	assert.Equal(t, 5, resp.TotalRooms)
}

func TestRoomService_UpdateRoom_NotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewRoomService(db)

	newType := "Deluxe"
	req := &dto.UpdateRoomRequest{
		RoomType: newType,
	}

	resp, err := service.UpdateRoom(999, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrRoomNotFound)
}

func TestRoomService_DeleteRoom_Success(t *testing.T) {
	db := setupTestDB(t)
	service := NewRoomService(db)

	room := models.Room{
		KostID:        1,
		RoomType:      "Standard",
		PricePerMonth: decimal.NewFromInt(1000000),
		TotalRooms:    5,
	}
	err := db.Create(&room).Error
	require.NoError(t, err)

	err = service.DeleteRoom(room.ID)

	assert.NoError(t, err)

	var count int64
	db.Model(&models.Room{}).Where("id = ?", room.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestRoomService_DeleteRoom_NotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewRoomService(db)

	err := service.DeleteRoom(999)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrRoomNotFound)
}

func TestRoomService_GetAllFacilities(t *testing.T) {
	db := setupTestDB(t)
	service := NewRoomService(db)

	facilities := []models.Facility{
		{Name: "AC"},
		{Name: "WiFi"},
		{Name: "TV"},
	}
	for _, f := range facilities {
		err := db.Create(&f).Error
		require.NoError(t, err)
	}

	resp, err := service.GetAllFacilities()

	assert.NoError(t, err)
	assert.Len(t, resp, 3)
}

func TestRoomService_CreateFacility_Success(t *testing.T) {
	db := setupTestDB(t)
	service := NewRoomService(db)

	req := &dto.CreateFacilityRequest{
		Name:    "AC",
		IconURL: strPtr("https://example.com/ac.png"),
	}

	resp, err := service.CreateFacility(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "AC", resp.Name)
	assert.Equal(t, "https://example.com/ac.png", *resp.IconURL)
}

func TestRoomService_UpdateFacility_Success(t *testing.T) {
	db := setupTestDB(t)
	service := NewRoomService(db)

	facility := models.Facility{Name: "AC"}
	err := db.Create(&facility).Error
	require.NoError(t, err)

	newName := "Air Conditioner"
	req := &dto.UpdateFacilityRequest{
		Name: newName,
	}

	resp, err := service.UpdateFacility(facility.ID, req)

	assert.NoError(t, err)
	assert.Equal(t, "Air Conditioner", resp.Name)
}

func TestRoomService_UpdateFacility_NotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewRoomService(db)

	req := &dto.UpdateFacilityRequest{Name: "AC"}
	resp, err := service.UpdateFacility(999, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrFacilityNotFound)
}

func TestRoomService_DeleteFacility_Success(t *testing.T) {
	db := setupTestDB(t)
	service := NewRoomService(db)

	facility := models.Facility{Name: "AC"}
	err := db.Create(&facility).Error
	require.NoError(t, err)

	err = service.DeleteFacility(facility.ID)

	assert.NoError(t, err)

	var count int64
	db.Model(&models.Facility{}).Where("id = ?", facility.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestRoomService_DeleteFacility_NotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewRoomService(db)

	err := service.DeleteFacility(999)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrFacilityNotFound)
}

func TestRoomService_CreateRoomFacility_Success(t *testing.T) {
	db := setupTestDB(t)
	service := NewRoomService(db)

	room := models.Room{
		KostID:        1,
		RoomType:      "Standard",
		PricePerMonth: decimal.NewFromInt(1000000),
		TotalRooms:    5,
	}
	err := db.Create(&room).Error
	require.NoError(t, err)

	facility := models.Facility{Name: "AC"}
	err = db.Create(&facility).Error
	require.NoError(t, err)

	req := &dto.CreateRoomFacilityRequest{
		FacilityID: facility.ID,
	}

	resp, err := service.CreateRoomFacility(room.ID, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, room.ID, resp.RoomID)
	assert.Equal(t, facility.ID, resp.FacilityID)
}

func TestRoomService_CreateRoomFacility_RoomNotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewRoomService(db)

	facility := models.Facility{Name: "AC"}
	err := db.Create(&facility).Error
	require.NoError(t, err)

	req := &dto.CreateRoomFacilityRequest{
		FacilityID: facility.ID,
	}

	resp, err := service.CreateRoomFacility(999, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrRoomNotFound)
}

func TestRoomService_CreateRoomFacility_FacilityNotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewRoomService(db)

	room := models.Room{
		KostID:        1,
		RoomType:      "Standard",
		PricePerMonth: decimal.NewFromInt(1000000),
		TotalRooms:    5,
	}
	err := db.Create(&room).Error
	require.NoError(t, err)

	req := &dto.CreateRoomFacilityRequest{
		FacilityID: 999,
	}

	resp, err := service.CreateRoomFacility(room.ID, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrFacilityNotFound)
}

func TestRoomService_DeleteRoomFacility_Success(t *testing.T) {
	db := setupTestDB(t)
	service := NewRoomService(db)

	room := models.Room{
		KostID:        1,
		RoomType:      "Standard",
		PricePerMonth: decimal.NewFromInt(1000000),
		TotalRooms:    5,
	}
	err := db.Create(&room).Error
	require.NoError(t, err)

	facility := models.Facility{Name: "AC"}
	err = db.Create(&facility).Error
	require.NoError(t, err)

	rf := models.RoomFacility{RoomID: room.ID, FacilityID: facility.ID}
	err = db.Create(&rf).Error
	require.NoError(t, err)

	err = service.DeleteRoomFacility(room.ID, facility.ID)

	assert.NoError(t, err)

	var count int64
	db.Model(&models.RoomFacility{}).Where("room_id = ? AND facility_id = ?", room.ID, facility.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestRoomService_DeleteRoomFacility_NotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewRoomService(db)

	err := service.DeleteRoomFacility(1, 1)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrRoomNotFound)
}

func TestToRoomResponse(t *testing.T) {
	room := &models.Room{
		ID:            1,
		KostID:        2,
		RoomType:      "Deluxe",
		PricePerMonth: decimal.NewFromInt(1500000),
		TotalRooms:    3,
		Facilities: []models.Facility{
			{ID: 1, Name: "AC"},
			{ID: 2, Name: "WiFi"},
		},
	}

	resp := toRoomResponse(room)

	assert.Equal(t, uint(1), resp.ID)
	assert.Equal(t, uint(2), resp.KostID)
	assert.Equal(t, "Deluxe", resp.RoomType)
	assert.Equal(t, 3, resp.TotalRooms)
	assert.Len(t, resp.Facilities, 2)
}

func TestToFacilityResponse(t *testing.T) {
	icon := "https://example.com/icon.png"
	facility := &models.Facility{
		ID:      1,
		Name:    "AC",
		IconURL: &icon,
	}

	resp := toFacilityResponse(facility)

	assert.Equal(t, uint(1), resp.ID)
	assert.Equal(t, "AC", resp.Name)
	assert.Equal(t, "https://example.com/icon.png", *resp.IconURL)
}

func TestToRoomFacilityResponse(t *testing.T) {
	rf := &models.RoomFacility{
		RoomID:     1,
		FacilityID: 2,
	}

	resp := toRoomFacilityResponse(rf)

	assert.Equal(t, uint(1), resp.RoomID)
	assert.Equal(t, uint(2), resp.FacilityID)
}
