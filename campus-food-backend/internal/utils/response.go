package utils

import "github.com/gin-gonic/gin"

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Success: true,
		Message: "操作成功",
		Data:    data,
	})
}

func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(200, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, message string) {
	c.JSON(400, Response{
		Success: false,
		Message: message,
		Error:   message,
	})
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(401, Response{
		Success: false,
		Message: message,
		Error:   message,
	})
}

func InternalError(c *gin.Context, message string) {
	c.JSON(500, Response{
		Success: false,
		Message: message,
		Error:   message,
	})
}
