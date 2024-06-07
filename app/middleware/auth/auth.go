package auth

import (
	"Bourse/common/inits"
	"Bourse/config"
	"Bourse/config/constant"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"gitee.com/phper95/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"net/http"
	"time"
)

var (
	conn          *sql.DB
	client        *redis.Client
	TokenUserInfo *UserInfoJson
)

type UserInfo struct {
	Id  sql.NullInt64  `json:"id"`
	Uid sql.NullString `json:"uid"`
}
type UserInfoJson struct {
	Id  int64  `json:"id"`
	Uid string `json:"uid"`
}

func init() {
	//初始化配置方式
	config.LoadConfig()
	inits.InitLog()

	conn = inits.InitMysql()
	client = inits.InitRedis()
}

func Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//获取请求头部信息
		token := ctx.GetHeader("Authorization")
		if token == "" {
			token = ctx.Query("token")
		}
		fmt.Println(token)

		if len(token) == 0 {
			logger.Warn("Login Token verification expired. Time:" + time.Now().Format("2006.01.02 15:04:05"))
			config.Errors(ctx, http.StatusUnauthorized, "Login Token verification expired")
			ctx.Abort()
			return
		}

		//验证token + token处理
		if ok := isLogin(token[7:]); !ok {
			logger.Warn("Login Token verification expired. Time:" + time.Now().Format("2006.01.02 15:04:05"))
			config.Errors(ctx, http.StatusUnauthorized, "Login Token verification expired.")
			ctx.Abort()
			return
		}

		//解析参数
		err := ctx.Request.ParseForm()
		if err != nil {
			logger.Warn("token解析出差. Time:" + time.Now().Format("2006.01.02 15:04:05"))
			config.Errors(ctx, http.StatusUnauthorized, "Login Token verification failed.")
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

// 登陆验证
func isLogin(token string) bool {
	sqlstr := "select id,uid from bo_users where accessToken=? and status=1 and deleteTime is null limit 1"
	rows, err := conn.Query(sqlstr, token)
	if err != nil {
		logger.Warn("Login Token verification failed.", zap.Error(err))
		return false
	}
	defer rows.Close()
	var userInfo UserInfo
	var userInfoJson UserInfoJson
	for rows.Next() {
		rows.Scan(&userInfo.Id, &userInfo.Uid)
		userInfoJson.Id = userInfo.Id.Int64
		userInfoJson.Uid = userInfo.Uid.String
		break
	}
	if userInfoJson.Id <= 0 {
		return false
	}
	TokenUserInfo = &userInfoJson
	//写入缓存
	marshal, _ := json.Marshal(userInfoJson)
	userInfoKey := fmt.Sprintf("%s%v", constant.UserInfoUidKey, userInfoJson.Id)
	err = client.Set(context.Background(), userInfoKey, marshal, 0).Err()
	if err != nil {
		logger.Warn("Login Token verification failed. Time:" + time.Now().Format("2006.01.02 15:04:05"))
		fmt.Println(err.Error())
		return false
	}
	return true
}
