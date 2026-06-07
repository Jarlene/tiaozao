package handler

import (
	"flea-market/pkg/errors"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Code:    errors.Success,
		Message: errors.MessageMap[errors.Success],
		Data:    data,
	})
}

func Error(c *gin.Context, httpStatus int, code int) {
	msg, ok := errors.MessageMap[code]
	if !ok {
		msg = "未知错误"
	}
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: msg,
	})
}

func ErrorWithMsg(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
	})
}
