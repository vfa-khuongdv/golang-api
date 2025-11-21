package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vfa-khuongdv/golang-cms/internal/constants"
)

func CalculateTotalPages(totalRows int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	totalPages := int(totalRows) / limit
	if int(totalRows)%limit != 0 {
		totalPages++
	}
	return totalPages
}

func ParsePageAndLimit(c *gin.Context) (int, int) {
	var pageInt, limitInt int64
	var page, limit string

	page = c.Query("page")
	limit = c.Query("limit")

	pageInt, err := strconv.ParseInt(page, 10, 64)
	if err != nil {
		pageInt = 1
	}
	if pageInt <= 0 {
		pageInt = 1
	}

	limitInt, err2 := strconv.ParseInt(limit, 10, 64)
	if err2 != nil {
		limitInt = int64(constants.LIMIT)
	}
	if limitInt <= 0 {
		limitInt = int64(constants.LIMIT)
	}

	return int(pageInt), int(limitInt)
}

type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
	Data       any `json:"data"`
}
