package inits

import (
	"Bourse/config"
	"gitee.com/phper95/pkg/logger"
	"gitee.com/phper95/pkg/timeutil"
)

func InitLog() {
	// 初始化 logger
	logger.InitLogger(
		logger.WithDisableConsole(),
		logger.WithTimeLayout(timeutil.CSTLayout),
		logger.WithFileRotationP(config.Cfg.App.AppLogPath),
	)
	defer func() {
		logger.Sync()
	}()
	return
}
