package services

import (
	"testing"

	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestUserService_GetProfile_Success(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	user := models.User{
		Email:       "test@example.com",
		Name:        "Test User",
		PhoneNumber: strPtr("08123456789"),
		Role:        models.RolePencari,
		IsActive:    true,
	}
	err := db.Create(&user).Error
	require.NoError(t, err)

	resp, err := service.GetProfile(user.ID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, user.ID, resp.ID)
	assert.Equal(t, user.Email, resp.Email)
	assert.Equal(t, user.Name, resp.Name)
	assert.Equal(t, "pencari", resp.Role)
	assert.True(t, resp.IsActive)
}

func TestUserService_GetProfile_NotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	resp, err := service.GetProfile(999)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestUserService_UpdateProfile_Success(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	user := models.User{
		Email:       "test@example.com",
		Name:        "Old Name",
		PhoneNumber: strPtr("08123456789"),
		Role:        models.RolePencari,
	}
	err := db.Create(&user).Error
	require.NoError(t, err)

	newName := "New Name"
	newPhone := "08987654321"
	req := &dto.UpdateProfileRequest{
		Name:        &newName,
		PhoneNumber: &newPhone,
	}

	resp, err := service.UpdateProfile(user.ID, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "New Name", resp.Name)
	assert.Equal(t, "08987654321", *resp.PhoneNumber)
}

func TestUserService_UpdateProfile_NotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	newName := "New Name"
	req := &dto.UpdateProfileRequest{
		Name: &newName,
	}

	resp, err := service.UpdateProfile(999, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestUserService_UpdateProfile_PartialUpdate(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	user := models.User{
		Email:       "test@example.com",
		Name:        "Original Name",
		PhoneNumber: strPtr("08123456789"),
		Role:        models.RolePencari,
	}
	err := db.Create(&user).Error
	require.NoError(t, err)

	newName := "Updated Name"
	req := &dto.UpdateProfileRequest{
		Name: &newName,
	}

	resp, err := service.UpdateProfile(user.ID, req)

	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", resp.Name)
	assert.Equal(t, "08123456789", *resp.PhoneNumber)
}
