package server

import (
	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/services"
	"github.com/frdavidh/nyarikos/internal/utils"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Routes(api *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	user := api.Group("/user")
	user.Use(authMiddleware)
	user.GET("/profile", h.GetProfile)
	user.PUT("/profile", h.UpdateProfile)
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	profile, err := h.userService.GetProfile(userID)
	if err != nil {
		utils.NotFoundResponse(c, "User not found", err)
		return
	}

	utils.SuccessResponse(c, "Success", profile)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	profile, err := h.userService.UpdateProfile(userID, &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update profile", err)
		return
	}

	utils.SuccessResponse(c, "Profile updated successfully", profile)
}
