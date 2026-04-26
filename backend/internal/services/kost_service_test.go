package services

import (
	"testing"

	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKostService_CreateKost_Success(t *testing.T) {
	db := setupTestDB(t)
	service := NewKostService(db)

	req := &dto.CreateKostRequest{
		Name:        "Kost Merdeka",
		Description: "Kost nyaman di pusat kota",
		Address:     "Jl. Merdeka No. 1",
		City:        "Jakarta",
		KostType:    "putra",
	}

	resp, err := service.CreateKost(1, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, req.Name, resp.Name)
	assert.Equal(t, req.Address, resp.Address)
	assert.Equal(t, req.City, resp.City)
	assert.Equal(t, req.KostType, resp.KostType)
	assert.Equal(t, uint(1), resp.OwnerID)
	assert.False(t, resp.IsPremium)
	assert.NotZero(t, resp.ID)
}

func TestKostService_CreateKost_WithPremium(t *testing.T) {
	db := setupTestDB(t)
	service := NewKostService(db)

	isPremium := true
	req := &dto.CreateKostRequest{
		Name:        "Kost Premium",
		Description: "Kost mewah",
		Address:     "Jl. Sudirman No. 1",
		City:        "Jakarta",
		IsPremium:   &isPremium,
		KostType:    "putri",
	}

	resp, err := service.CreateKost(1, req)

	assert.NoError(t, err)
	assert.True(t, resp.IsPremium)
}

func TestKostService_GetKost_Success(t *testing.T) {
	db := setupTestDB(t)
	service := NewKostService(db)

	kost := models.Kost{
		OwnerID:  1,
		Name:     "Kost Merdeka",
		Address:  "Jl. Merdeka No. 1",
		City:     "Jakarta",
		KostType: "putra",
	}
	err := db.Create(&kost).Error
	require.NoError(t, err)

	resp, err := service.GetKost(kost.ID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, kost.Name, resp.Name)
	assert.Equal(t, kost.City, resp.City)
}

func TestKostService_GetKost_NotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewKostService(db)

	resp, err := service.GetKost(999)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrKostNotFound)
}

func TestKostService_GetAllKost(t *testing.T) {
	db := setupTestDB(t)
	service := NewKostService(db)

	for i := 0; i < 5; i++ {
		kost := models.Kost{
			OwnerID:  uint(i + 1),
			Name:     "Kost Test",
			Address:  "Jl. Test",
			City:     "Jakarta",
			KostType: "putra",
		}
		err := db.Create(&kost).Error
		require.NoError(t, err)
	}

	req := &dto.SearchKostRequest{Page: 1, Limit: 2}
	resp, total, err := service.GetAllKost(req)

	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, resp, 2)
}

func TestKostService_GetAllKost_FilterByName(t *testing.T) {
	db := setupTestDB(t)
	service := NewKostService(db)

	kost1 := models.Kost{OwnerID: 1, Name: "Kost Merdeka", Address: "Jl. A", City: "Jakarta", KostType: "putra"}
	kost2 := models.Kost{OwnerID: 1, Name: "Kost Sudirman", Address: "Jl. B", City: "Jakarta", KostType: "putra"}
	require.NoError(t, db.Create(&kost1).Error)
	require.NoError(t, db.Create(&kost2).Error)

	req := &dto.SearchKostRequest{Q: "Merdeka", Page: 1, Limit: 10}
	resp, total, err := service.GetAllKost(req)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, resp, 1)
	assert.Equal(t, "Kost Merdeka", resp[0].Name)
}

func TestKostService_GetAllKost_FilterByCity(t *testing.T) {
	db := setupTestDB(t)
	service := NewKostService(db)

	kost1 := models.Kost{OwnerID: 1, Name: "Kost A", Address: "Jl. A", City: "Jakarta", KostType: "putra"}
	kost2 := models.Kost{OwnerID: 1, Name: "Kost B", Address: "Jl. B", City: "Bandung", KostType: "putra"}
	require.NoError(t, db.Create(&kost1).Error)
	require.NoError(t, db.Create(&kost2).Error)

	req := &dto.SearchKostRequest{City: "Bandung", Page: 1, Limit: 10}
	resp, total, err := service.GetAllKost(req)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Kost B", resp[0].Name)
}

func TestKostService_GetAllKost_FilterByKostType(t *testing.T) {
	db := setupTestDB(t)
	service := NewKostService(db)

	kost1 := models.Kost{OwnerID: 1, Name: "Kost A", Address: "Jl. A", City: "Jakarta", KostType: "putra"}
	kost2 := models.Kost{OwnerID: 1, Name: "Kost B", Address: "Jl. B", City: "Jakarta", KostType: "putri"}
	require.NoError(t, db.Create(&kost1).Error)
	require.NoError(t, db.Create(&kost2).Error)

	req := &dto.SearchKostRequest{KostType: "putri", Page: 1, Limit: 10}
	resp, total, err := service.GetAllKost(req)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Kost B", resp[0].Name)
}

func TestKostService_GetAllKost_FilterByPriceRange(t *testing.T) {
	db := setupTestDB(t)
	service := NewKostService(db)

	kost1 := models.Kost{OwnerID: 1, Name: "Kost A", Address: "Jl. A", City: "Jakarta", KostType: "putra"}
	kost2 := models.Kost{OwnerID: 1, Name: "Kost B", Address: "Jl. B", City: "Jakarta", KostType: "putra"}
	require.NoError(t, db.Create(&kost1).Error)
	require.NoError(t, db.Create(&kost2).Error)

	room1 := models.Room{KostID: kost1.ID, RoomType: "single", PricePerMonth: decimal.NewFromFloat(1000000), TotalRooms: 1}
	room2 := models.Room{KostID: kost2.ID, RoomType: "single", PricePerMonth: decimal.NewFromFloat(2000000), TotalRooms: 1}
	require.NoError(t, db.Create(&room1).Error)
	require.NoError(t, db.Create(&room2).Error)

	req := &dto.SearchKostRequest{MinPrice: 1500000, MaxPrice: 2500000, Page: 1, Limit: 10}
	resp, total, err := service.GetAllKost(req)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Kost B", resp[0].Name)
}

func TestKostService_GetAllKost_FilterByRoomType(t *testing.T) {
	db := setupTestDB(t)
	service := NewKostService(db)

	kost1 := models.Kost{OwnerID: 1, Name: "Kost A", Address: "Jl. A", City: "Jakarta", KostType: "putra"}
	kost2 := models.Kost{OwnerID: 1, Name: "Kost B", Address: "Jl. B", City: "Jakarta", KostType: "putra"}
	require.NoError(t, db.Create(&kost1).Error)
	require.NoError(t, db.Create(&kost2).Error)

	room1 := models.Room{KostID: kost1.ID, RoomType: "single", PricePerMonth: decimal.NewFromFloat(1000000), TotalRooms: 1}
	room2 := models.Room{KostID: kost2.ID, RoomType: "double", PricePerMonth: decimal.NewFromFloat(2000000), TotalRooms: 1}
	require.NoError(t, db.Create(&room1).Error)
	require.NoError(t, db.Create(&room2).Error)

	req := &dto.SearchKostRequest{RoomType: "double", Page: 1, Limit: 10}
	resp, total, err := service.GetAllKost(req)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Kost B", resp[0].Name)
}

func TestKostService_GetAllKost_FilterByFacilities(t *testing.T) {
	db := setupTestDB(t)
	service := NewKostService(db)

	kost1 := models.Kost{OwnerID: 1, Name: "Kost A", Address: "Jl. A", City: "Jakarta", KostType: "putra"}
	kost2 := models.Kost{OwnerID: 1, Name: "Kost B", Address: "Jl. B", City: "Jakarta", KostType: "putra"}
	require.NoError(t, db.Create(&kost1).Error)
	require.NoError(t, db.Create(&kost2).Error)

	fac1 := models.Facility{Name: "WiFi"}
	fac2 := models.Facility{Name: "AC"}
	require.NoError(t, db.Create(&fac1).Error)
	require.NoError(t, db.Create(&fac2).Error)

	room1 := models.Room{KostID: kost1.ID, RoomType: "single", PricePerMonth: decimal.NewFromFloat(1000000), TotalRooms: 1}
	room2 := models.Room{KostID: kost2.ID, RoomType: "single", PricePerMonth: decimal.NewFromFloat(1500000), TotalRooms: 1}
	require.NoError(t, db.Create(&room1).Error)
	require.NoError(t, db.Create(&room2).Error)

	require.NoError(t, db.Create(&models.RoomFacility{RoomID: room1.ID, FacilityID: fac1.ID}).Error)
	require.NoError(t, db.Create(&models.RoomFacility{RoomID: room2.ID, FacilityID: fac1.ID}).Error)
	require.NoError(t, db.Create(&models.RoomFacility{RoomID: room2.ID, FacilityID: fac2.ID}).Error)

	req := &dto.SearchKostRequest{FacilityIDs: []uint{fac1.ID, fac2.ID}, Page: 1, Limit: 10}
	resp, total, err := service.GetAllKost(req)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Kost B", resp[0].Name)
}

func TestKostService_GetAllKost_CombinedFilters(t *testing.T) {
	db := setupTestDB(t)
	service := NewKostService(db)

	kost1 := models.Kost{OwnerID: 1, Name: "Kost Merdeka", Address: "Jl. A", City: "Jakarta", KostType: "putra"}
	kost2 := models.Kost{OwnerID: 1, Name: "Kost Sudirman", Address: "Jl. B", City: "Jakarta", KostType: "putra"}
	require.NoError(t, db.Create(&kost1).Error)
	require.NoError(t, db.Create(&kost2).Error)

	room1 := models.Room{KostID: kost1.ID, RoomType: "single", PricePerMonth: decimal.NewFromFloat(1000000), TotalRooms: 1}
	room2 := models.Room{KostID: kost2.ID, RoomType: "double", PricePerMonth: decimal.NewFromFloat(2000000), TotalRooms: 1}
	require.NoError(t, db.Create(&room1).Error)
	require.NoError(t, db.Create(&room2).Error)

	req := &dto.SearchKostRequest{Q: "Sudirman", City: "Jakarta", MinPrice: 1500000, RoomType: "double", Page: 1, Limit: 10}
	resp, total, err := service.GetAllKost(req)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Kost Sudirman", resp[0].Name)
}

func TestKostService_GetAllKost_NoResults(t *testing.T) {
	db := setupTestDB(t)
	service := NewKostService(db)

	kost := models.Kost{OwnerID: 1, Name: "Kost A", Address: "Jl. A", City: "Jakarta", KostType: "putra"}
	require.NoError(t, db.Create(&kost).Error)

	req := &dto.SearchKostRequest{Q: "NonExistent", Page: 1, Limit: 10}
	resp, total, err := service.GetAllKost(req)

	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, resp)
}

func TestKostService_UpdateKost_Success(t *testing.T) {
	db := setupTestDB(t)
	service := NewKostService(db)

	kost := models.Kost{
		OwnerID:  1,
		Name:     "Kost Lama",
		Address:  "Jl. Lama",
		City:     "Jakarta",
		KostType: "putra",
	}
	err := db.Create(&kost).Error
	require.NoError(t, err)

	newName := "Kost Baru"
	newCity := "Bandung"
	req := &dto.UpdateKostRequest{
		Name: &newName,
		City: &newCity,
	}

	resp, err := service.UpdateKost(kost.ID, 1, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Kost Baru", resp.Name)
	assert.Equal(t, "Bandung", resp.City)
	assert.Equal(t, "Jl. Lama", resp.Address)
}

func TestKostService_UpdateKost_NotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewKostService(db)

	newName := "Kost Baru"
	req := &dto.UpdateKostRequest{
		Name: &newName,
	}

	resp, err := service.UpdateKost(999, 1, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrKostNotFound)
}

func TestKostService_UpdateKost_Unauthorized(t *testing.T) {
	db := setupTestDB(t)
	service := NewKostService(db)

	kost := models.Kost{
		OwnerID:  1,
		Name:     "Kost Merdeka",
		Address:  "Jl. Merdeka",
		City:     "Jakarta",
		KostType: "putra",
	}
	err := db.Create(&kost).Error
	require.NoError(t, err)

	newName := "Kost Baru"
	req := &dto.UpdateKostRequest{
		Name: &newName,
	}

	resp, err := service.UpdateKost(kost.ID, 2, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestKostService_DeleteKost_Success(t *testing.T) {
	db := setupTestDB(t)
	service := NewKostService(db)

	kost := models.Kost{
		OwnerID:  1,
		Name:     "Kost Merdeka",
		Address:  "Jl. Merdeka",
		City:     "Jakarta",
		KostType: "putra",
	}
	err := db.Create(&kost).Error
	require.NoError(t, err)

	resp, err := service.DeleteKost(kost.ID, 1)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, kost.Name, resp.Name)

	var count int64
	db.Model(&models.Kost{}).Where("id = ?", kost.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestKostService_DeleteKost_NotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewKostService(db)

	resp, err := service.DeleteKost(999, 1)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrKostNotFound)
}

func TestKostService_DeleteKost_Unauthorized(t *testing.T) {
	db := setupTestDB(t)
	service := NewKostService(db)

	kost := models.Kost{
		OwnerID:  1,
		Name:     "Kost Merdeka",
		Address:  "Jl. Merdeka",
		City:     "Jakarta",
		KostType: "putra",
	}
	err := db.Create(&kost).Error
	require.NoError(t, err)

	resp, err := service.DeleteKost(kost.ID, 2)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestKostService_AddKostImage_Success(t *testing.T) {
	db := setupTestDB(t)
	service := NewKostService(db)

	kost := models.Kost{
		OwnerID:  1,
		Name:     "Kost Merdeka",
		Address:  "Jl. Merdeka",
		City:     "Jakarta",
		KostType: "putra",
	}
	err := db.Create(&kost).Error
	require.NoError(t, err)

	err = service.AddKostImage(kost.ID, 1, "https://example.com/image.jpg", "Kost Image")

	assert.NoError(t, err)

	var image models.KostImage
	err = db.First(&image).Error
	require.NoError(t, err)
	assert.Equal(t, kost.ID, image.KostID)
	assert.Equal(t, "https://example.com/image.jpg", image.ImageURL)
	assert.True(t, image.IsMain)
}

func TestKostService_AddKostImage_NotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewKostService(db)

	err := service.AddKostImage(999, 1, "https://example.com/image.jpg", "Kost Image")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrKostNotFound)
}

func TestKostService_AddKostImage_Unauthorized(t *testing.T) {
	db := setupTestDB(t)
	service := NewKostService(db)

	kost := models.Kost{
		OwnerID:  1,
		Name:     "Kost Merdeka",
		Address:  "Jl. Merdeka",
		City:     "Jakarta",
		KostType: "putra",
	}
	err := db.Create(&kost).Error
	require.NoError(t, err)

	err = service.AddKostImage(kost.ID, 2, "https://example.com/image.jpg", "Kost Image")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestKostService_AddKostImage_SecondaryImage(t *testing.T) {
	db := setupTestDB(t)
	service := NewKostService(db)

	kost := models.Kost{
		OwnerID:  1,
		Name:     "Kost Merdeka",
		Address:  "Jl. Merdeka",
		City:     "Jakarta",
		KostType: "putra",
	}
	err := db.Create(&kost).Error
	require.NoError(t, err)

	err = service.AddKostImage(kost.ID, 1, "https://example.com/main.jpg", "Main")
	require.NoError(t, err)

	err = service.AddKostImage(kost.ID, 1, "https://example.com/second.jpg", "Second")
	require.NoError(t, err)

	var images []models.KostImage
	err = db.Where("kost_id = ?", kost.ID).Find(&images).Error
	require.NoError(t, err)
	assert.Len(t, images, 2)

	var mainCount, secondaryCount int
	for _, img := range images {
		if img.IsMain {
			mainCount++
		} else {
			secondaryCount++
		}
	}
	assert.Equal(t, 1, mainCount)
	assert.Equal(t, 1, secondaryCount)
}

func TestToKostResponse(t *testing.T) {
	kost := &models.Kost{
		ID:       1,
		OwnerID:  2,
		Name:     "Kost Test",
		Address:  "Jl. Test",
		City:     "Jakarta",
		KostType: "putra",
		Images: []models.KostImage{
			{ID: 1, ImageURL: "https://example.com/1.jpg"},
			{ID: 2, ImageURL: "https://example.com/2.jpg"},
		},
	}

	resp := toKostResponse(kost)

	assert.Equal(t, uint(1), resp.ID)
	assert.Equal(t, uint(2), resp.OwnerID)
	assert.Equal(t, "Kost Test", resp.Name)
	assert.Len(t, resp.Images, 2)
}
