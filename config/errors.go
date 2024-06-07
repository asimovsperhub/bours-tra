package config

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Errors(ctx *gin.Context, code int, msg any) {
	ctx.JSON(http.StatusUnauthorized, gin.H{
		"code": code,
		"msg":  msg,
	})
	return
}

func Success(ctx *gin.Context, code int, msg any, data any) {
	if msg == nil {
		msg = "请求成功"
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  msg,
		"data": data,
	})
	return
}
