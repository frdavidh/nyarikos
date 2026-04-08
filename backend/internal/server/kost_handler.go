package server

import (
	"errors"
	"strconv"

	"github.com/frdavidh/nyarikos/internal/dto"
	"github.com/frdavidh/nyarikos/internal/services"
	"github.com/frdavidh/nyarikos/internal/utils"
	"github.com/gin-gonic/gin"
)

type KostHandler struct {
	kostService services.KostService
}

func NewKostHandler(kostService services.KostService) *KostHandler {
	return &KostHandler{kostService: kostService}
}

func (h *KostHandler) Routes(api *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	kost := api.Group("/kost")
	kost.Use(middlewares...)
	kost.GET("/", h.GetAllKost)
	kost.POST("/", h.CreateKost)
	kost.GET("/:id", h.GetKost)
	kost.PUT("/:id", h.UpdateKost)
	kost.DELETE("/:id", h.DeleteKost)
}

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

func (h *KostHandler) GetAllKost(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 {
		limit = 10
	}

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

func (h *KostHandler) GetKost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.BadRequestResponse(c, "invalid kost ID", err)
		return
	}
	kost, err := h.kostService.GetKost(uint(id))
	// if err != nil {
	// 	utils.InternalServerErrorResponse(c, "Failed to get kost", err)
	// 	return
	// }
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

func (h *KostHandler) UpdateKost(c *gin.Context) {
	idStr := c.Param("id")
	kostID, err := strconv.Atoi(idStr)
	if err != nil {
		utils.BadRequestResponse(c, "invalid kost ID", err)
		return
	}

	userID := c.GetUint("user_id")

	var req dto.UpdateKostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "invalid request body", err)
		return
	}

	kost, err := h.kostService.UpdateKost(uint(kostID), userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUnauthorized):
			utils.ForbiddenResponse(c, "You are not allowed to update kost", nil)
		default:
			utils.InternalServerErrorResponse(c, "Failed to update kost", err)
		}
		return
	}

	utils.SuccessResponse(c, "Kost updated successfully", kost)
}

func (h *KostHandler) DeleteKost(c *gin.Context) {
	idStr := c.Param("id")
	kostID, err := strconv.Atoi(idStr)
	if err != nil {
		utils.BadRequestResponse(c, "invalid kost ID", err)
		return
	}

	kost, err := h.kostService.DeleteKost(uint(kostID))
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
