package auth

import (
	"Bourse/config"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Cors 跨越
func Cors() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		method := ctx.Request.Method
		origin := ctx.Request.Header.Get("Origin")
		if origin != "" {
			ctx.Header("Access-Control-Allow-Origin", "*")
			ctx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			ctx.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token,Authorization,language")
			ctx.Header("Access-Control-Allow-Credentials", "true")
			ctx.Set("content-type", "application/json")
		}
		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			ctx.AbortWithStatus(http.StatusNoContent)
		}
		ctx.Next()
	}
}

// Recovery 铺货异常
func Recovery(ctx *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("系统内部错误")
			config.Errors(ctx, http.StatusInternalServerError, "系统内部错误")
		}
		ctx.Next()
	}()
}
