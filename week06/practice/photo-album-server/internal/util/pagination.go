package util

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultPage     = 1
	defaultPageSize = 10
	maxPageSize     = 50
)

func ParsePagination(c *gin.Context) (int, int) {
	page := parsePositiveInt(c.DefaultQuery("page", strconv.Itoa(defaultPage)), defaultPage)
	pageSize := parsePositiveInt(c.DefaultQuery("page_size", strconv.Itoa(defaultPageSize)), defaultPageSize)
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	return page, pageSize
}

func BuildListData(list interface{}, page, pageSize int, total int64) gin.H {
	return gin.H{
		"list": list,
		"pagination": PaginationMeta{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}
}

func Offset(page, pageSize int) int {
	return (page - 1) * pageSize
}

func parsePositiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}
