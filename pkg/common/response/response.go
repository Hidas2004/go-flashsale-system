package response

import "github.com/gin-gonic/gin"

type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// successResponse
func SuccessResponse(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, Response{
		Status:  code,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse
func ErrorResponse(c *gin.Context, code int, message string, err string) {
	c.JSON(code, Response{
		Status:  code,
		Message: message,
		Error:   err,
	})
}
