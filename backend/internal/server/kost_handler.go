package server

import (
	"errors"

	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/services"
	"github.com/frdavidh/nyarikos/internal/utils"
	"github.com/gin-gonic/gin"
)

type KostHandler struct {
	kostService   services.KostService
	uploadService *services.UploadService
}

func NewKostHandler(kostService services.KostService, uploadService *services.UploadService) *KostHandler {
	return &KostHandler{kostService: kostService, uploadService: uploadService}
}

func (h *KostHandler) Routes(api *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	kost := api.Group("/kost")
	kost.Use(middlewares...)
	kost.GET("/", h.GetAllKost)
	kost.POST("/", h.CreateKost)
	kost.GET("/:id", h.GetKost)
	kost.PUT("/:id", h.UpdateKost)
	kost.DELETE("/:id", h.DeleteKost)
	kost.POST("/:id/images", h.AddKostImage)
}

// @Tags			Kost
// @Summary		Create a new kost
// @Description	Create a new kost listing (requires pemilik role)
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			request	body		dto.CreateKostRequest					true	"Create Kost Request"
// @Success		201		{object}	utils.Response{data=dto.KostResponse}	"Kost created successfully"
// @Failure		400		{object}	utils.Response							"Invalid request"
// @Failure		401		{object}	utils.Response							"Unauthorized"
// @Failure		403		{object}	utils.Response							"Forbidden"
// @Failure		500		{object}	utils.Response							"Internal server error"
// @Router			/kost/ [post]
func (h *KostHandler) CreateKost(c *gin.Context) {
	var req dto.CreateKostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	ownerID := c.GetUint("user_id")
	kost, err := h.kostService.CreateKost(ownerID, &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUnauthorized):
			utils.ForbiddenResponse(c, "You are not allowed to create kost", nil)
		default:
			utils.InternalServerErrorResponse(c, "Failed to create kost", err)
		}
		return
	}

	utils.CreatedResponse(c, "Kost created successfully", kost)
}

// @Tags			Kost
// @Summary		Get all kosts
// @Description	Get a paginated list of all kosts
// @Produce		json
// @Security		BearerAuth
// @Param			page	query		int													false	"Page number"		default(1)
// @Param			limit	query		int													false	"Items per page"	default(10)
// @Success		200		{object}	utils.PaginatedResponse{data=[]dto.KostResponse}	"Kost fetched successfully"
// @Failure		401		{object}	utils.Response										"Unauthorized"
// @Failure		500		{object}	utils.Response										"Internal server error"
// @Router			/kost/ [get]
func (h *KostHandler) GetAllKost(c *gin.Context) {
	page, limit := parsePagination(c)

	kosts, total, err := h.kostService.GetAllKost(page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get all kost", err)
		return
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	utils.PaginatedSuccessResponse(c, "Kost fetched successfully", kosts, utils.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	})
}

// @Tags			Kost
// @Summary		Get kost by ID
// @Description	Get detailed information about a specific kost
// @Produce		json
// @Security		BearerAuth
// @Param			id	path		int										true	"Kost ID"
// @Success		200	{object}	utils.Response{data=dto.KostResponse}	"Kost fetched successfully"
// @Failure		401	{object}	utils.Response							"Unauthorized"
// @Failure		404	{object}	utils.Response							"Kost not found"
// @Failure		500	{object}	utils.Response							"Internal server error"
// @Router			/kost/{id} [get]
func (h *KostHandler) GetKost(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	kost, err := h.kostService.GetKost(id)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrKostNotFound):
			utils.NotFoundResponse(c, "kost not found", nil)
		default:
			utils.InternalServerErrorResponse(c, "Failed to get kost", err)
		}
		return
	}

	utils.SuccessResponse(c, "Kost fetched successfully", kost)
}

// @Tags			Kost
// @Summary		Update kost
// @Description	Update a kost listing (requires ownership)
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id		path		int										true	"Kost ID"
// @Param			request	body		dto.UpdateKostRequest					true	"Update Kost Request"
// @Success		200		{object}	utils.Response{data=dto.KostResponse}	"Kost updated successfully"
// @Failure		400		{object}	utils.Response							"Invalid request"
// @Failure		401		{object}	utils.Response							"Unauthorized"
// @Failure		403		{object}	utils.Response							"Forbidden"
// @Failure		404		{object}	utils.Response							"Kost not found"
// @Failure		500		{object}	utils.Response							"Internal server error"
// @Router			/kost/{id} [put]
func (h *KostHandler) UpdateKost(c *gin.Context) {
	kostID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	userID := c.GetUint("user_id")
	var req dto.UpdateKostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	kost, err := h.kostService.UpdateKost(kostID, userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrKostNotFound):
			utils.NotFoundResponse(c, "kost not found", nil)
		case errors.Is(err, services.ErrUnauthorized):
			utils.ForbiddenResponse(c, "you are not allowed to update this kost", nil)
		default:
			utils.InternalServerErrorResponse(c, "failed to update kost", err)
		}
		return
	}

	utils.SuccessResponse(c, "Kost updated successfully", kost)
}

// @Tags			Kost
// @Summary		Delete kost
// @Description	Delete a kost listing (requires ownership)
// @Produce		json
// @Security		BearerAuth
// @Param			id	path		int										true	"Kost ID"
// @Success		200	{object}	utils.Response{data=dto.KostResponse}	"Kost deleted successfully"
// @Failure		401	{object}	utils.Response							"Unauthorized"
// @Failure		403	{object}	utils.Response							"Forbidden"
// @Failure		404	{object}	utils.Response							"Kost not found"
// @Failure		500	{object}	utils.Response							"Internal server error"
// @Router			/kost/{id} [delete]
func (h *KostHandler) DeleteKost(c *gin.Context) {
	kostID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	userID := c.GetUint("user_id")
	kost, err := h.kostService.DeleteKost(kostID, userID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrKostNotFound):
			utils.NotFoundResponse(c, "kost not found", nil)
		case errors.Is(err, services.ErrUnauthorized):
			utils.ForbiddenResponse(c, "you are not allowed to delete this kost", nil)
		default:
			utils.InternalServerErrorResponse(c, "failed to delete kost", err)
		}
		return
	}

	utils.SuccessResponse(c, "Kost deleted successfully", kost)
}

// @Tags			Kost
// @Summary		Add image to kost
// @Description	Upload and add an image to a kost listing (requires ownership)
// @Accept			multipart/form-data
// @Produce		json
// @Security		BearerAuth
// @Param			id			path		int				true	"Kost ID"
// @Param			image		formData	file			true	"Image file"
// @Param			alt_text	formData	string			true	"Alt text for the image"
// @Success		201			{object}	utils.Response	"Image added successfully"
// @Failure		400			{object}	utils.Response	"Invalid request"
// @Failure		401			{object}	utils.Response	"Unauthorized"
// @Failure		403			{object}	utils.Response	"Forbidden"
// @Failure		404			{object}	utils.Response	"Kost not found"
// @Failure		500			{object}	utils.Response	"Internal server error"
// @Router			/kost/{id}/images [post]
func (h *KostHandler) AddKostImage(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		utils.BadRequestResponse(c, "no file uploaded", err)
		return
	}

	altText := c.PostForm("alt_text")
	if altText == "" {
		utils.BadRequestResponse(c, "alt text is required", nil)
		return
	}

	url, err := h.uploadService.UploadKostImage(id, file)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidFileType):
			utils.BadRequestResponse(c, "invalid file type", err)
		default:
			utils.InternalServerErrorResponse(c, "failed to upload image", err)
		}
		return
	}

	userID := c.GetUint("user_id")
	err = h.kostService.AddKostImage(id, userID, url, altText)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrKostNotFound):
			utils.NotFoundResponse(c, "kost not found", nil)
		case errors.Is(err, services.ErrUnauthorized):
			utils.ForbiddenResponse(c, "you are not allowed to add image to this kost", nil)
		default:
			utils.InternalServerErrorResponse(c, "failed to add image", err)
		}
		return
	}

	utils.CreatedResponse(c, "Image added successfully", nil)
}
