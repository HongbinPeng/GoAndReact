package util

import "github.com/gin-gonic/gin"

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PaginationMeta struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

func Success(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, APIResponse{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, APIResponse{
		Code:    status,
		Message: message,
	})
}
