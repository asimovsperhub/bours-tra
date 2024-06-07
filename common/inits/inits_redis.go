package inits

import (
	"Bourse/config"
	"context"
	"fmt"
	"gitee.com/phper95/pkg/logger"
	"github.com/redis/go-redis/v9"
	"time"
)

func InitRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Cfg.Redis.Host,
		Password: config.Cfg.Redis.Password,
		DB:       config.Cfg.Redis.DB,
	})
	//检测是否需要建立连接
	timeoutCtx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelFunc()
	//检查
	_, err := client.Ping(timeoutCtx).Result()
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed:[%s] Time:[%s] Err:[%v]", "Redis连接失败!", time.Now().Format(time.DateTime), err))
		return nil
	}
	return client
}
