package main

import (
	"Bourse/config"
	"Bourse/routers"
	"fmt"
	"gitee.com/phper95/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	router := routers.InitRouter() //启动路由

	err := router.Run(fmt.Sprintf(":%d", config.Cfg.App.HttpPort)) //启动服务
	if err != nil {
		logger.Error("http server start error", zap.Error(err))
	}
	return
}
