package user_fund_accounts

import (
	"Bourse/app/models"
	"Bourse/common/global"
	"Bourse/common/inits"
	"Bourse/common/translators"
	"Bourse/config"
	"Bourse/config/constant"
	"database/sql"
	"fmt"
	"gitee.com/phper95/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
	"time"
)

var (
	conn   *sql.DB
	client *redis.Client

	UserInfo *constant.UserInfoStruct
)

func init() {
	//初始化配置方式
	config.LoadConfig()
	inits.InitLog()

	//多语言
	if err := translators.InitTranslators("zh"); err != nil {
		fmt.Sprintf("init Translator failed,err:%v\n", err)
	}
}

// GetListUserFundAccounts 获取用户资产账户列表
func GetListUserFundAccounts(ctx *gin.Context) {
	//mysql初始化方式
	conn = inits.InitMysql()
	defer conn.Close() //关闭数据库连接
	var msg any

	//redis初始化方式
	client = inits.InitRedis()
	UserInfo = global.GetUserInfo(client)

	sqlStr := GetListUserFundAccountsSql()
	Rows, err := conn.Query(sqlStr, UserInfo.Id)
	if err != nil {
		logger.Warn(fmt.Sprintf("Err:[Query Sql failed.] SQl:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err.Error()))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	defer Rows.Close() //关闭查询

	var userFundAccountsJsonMap []*models.UserFundAccountsJson
	for Rows.Next() {
		var userFundAccounts models.UserFundAccounts
		var userFundAccountsJson models.UserFundAccountsJson
		Rows.Scan(&userFundAccounts.Id, &userFundAccounts.UserId, &userFundAccounts.FundAccount, &userFundAccounts.FundAccountType, &userFundAccounts.Money)
		userFundAccountsJson = models.ConvUserFundAccountsNull2Json(userFundAccounts, userFundAccountsJson)
		userFundAccountsJsonMap = append(userFundAccountsJsonMap, &userFundAccountsJson)
	}

	config.Success(ctx, http.StatusOK, msg, &userFundAccountsJsonMap)
	return
}
