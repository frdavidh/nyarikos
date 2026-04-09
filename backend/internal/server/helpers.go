package server

import (
	"strconv"

	"github.com/frdavidh/nyarikos/internal/utils"
	"github.com/gin-gonic/gin"
)

func parseUintParam(c *gin.Context, param string) (uint, bool) {
	idStr := c.Param(param)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.BadRequestResponse(c, "invalid "+param+" ID", err)
		return 0, false
	}
	return uint(id), true
}

func parsePagination(c *gin.Context) (page, limit int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ = strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 {
		limit = 10
	}
	return page, limit
}
