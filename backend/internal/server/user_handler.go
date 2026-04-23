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

func (h *UserHandler) Routes(api *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	user := api.Group("/user")
	user.Use(middlewares...)
	user.GET("/profile", h.GetProfile)
	user.PUT("/profile", h.UpdateProfile)
}

// @Tags			User
// @Summary		Get user profile
// @Description	Get the profile of the currently authenticated user
// @Produce		json
// @Security		BearerAuth
// @Success		200	{object}	utils.Response{data=dto.UserResponse}	"Success"
// @Failure		401	{object}	utils.Response							"Unauthorized"
// @Failure		404	{object}	utils.Response							"User not found"
// @Router			/user/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	profile, err := h.userService.GetProfile(userID)
	if err != nil {
		utils.NotFoundResponse(c, "User not found", err)
		return
	}

	utils.SuccessResponse(c, "Success", profile)
}

// @Tags			User
// @Summary		Update user profile
// @Description	Update the profile of the currently authenticated user
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			request	body		dto.UpdateProfileRequest				true	"Update Profile Request"
// @Success		200		{object}	utils.Response{data=dto.UserResponse}	"Profile updated successfully"
// @Failure		400		{object}	utils.Response							"Invalid request"
// @Failure		401		{object}	utils.Response							"Unauthorized"
// @Failure		500		{object}	utils.Response							"Internal server error"
// @Router			/user/profile [put]
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
