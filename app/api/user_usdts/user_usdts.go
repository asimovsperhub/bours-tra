package user_usdts

import (
	"Bourse/app/models"
	"Bourse/common/global"
	"Bourse/common/inits"
	"Bourse/common/translators"
	"Bourse/config"
	"Bourse/config/constant"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"gitee.com/phper95/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"net/http"
	"time"
)

var (
	conn   *sql.DB
	client *redis.Client

	UserInfo *constant.UserInfoStruct
)

type UserUsdtsForm struct {
	UsdtId int64 `json:"usdtId" form:"usdtId"`
	//FundAccountsId int64  `json:"fundAccountsId" form:"fundAccountsId" binding:"required"`
	AddreType   int32  `json:"addreType" form:"addreType" binding:"required"`
	UsdtAddre   string `json:"usdtAddre" form:"usdtAddre" binding:"required"`
	UsdtTitle   string `json:"usdtTitle" form:"usdtTitle" binding:"required"`
	DefaultCard int32  `json:"defaultCard" form:"defaultCard"`
}

type UserUsdtsDeleteForm struct {
	UsdtId int64 `json:"usdtId" form:"usdtId" binding:"required"`
}

type UserUsdtsReq struct {
	UsdtId         int64  `json:"usdtId"`
	FundAccountsId int64  `json:"fundAccountsId"`
	UserId         int64  `json:"userId"`
	AddreType      int32  `json:"addreType"`
	UsdtAddre      string `json:"usdtAddre"`
	UsdtTitle      string `json:"usdtTitle"`
	DefaultCard    int32  `json:"defaultCard"`
}

func init() {
	//初始化配置方式
	config.LoadConfig()
	inits.InitLog()

	//多语言
	if err := translators.InitTranslators("zh"); err != nil {
		fmt.Sprintf("init Translator failed,err:%v\n", err)
	}
}

func CreateUserUsdts(ctx *gin.Context) {
	//获取表单数据,并验证参数
	var userUsdtsForm UserUsdtsForm
	var msg any

	//redis初始化方式
	client = inits.InitRedis()
	UserInfo = global.GetUserInfo(client)

	//mysql初始化方式
	conn = inits.InitMysql()
	defer conn.Close() //关闭数据库连接

	if errs := ctx.ShouldBindJSON(&userUsdtsForm); errs != nil {
		code := 500
		err, ok := errs.(validator.ValidationErrors)
		msg = err.Translate(translators.Trans)
		if !ok {
			msg = errs.Error()
		}
		config.Errors(ctx, code, msg)
		return
	}

	connTx, err := conn.Begin()
	if err != nil {
		connTx.Rollback()
		logger.Warn(fmt.Sprintf("Err:[事务开启失败] Time:[%s] Err:[%v]", time.Now().Format(time.DateTime), err.Error()))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}

	//查询该用户是否已设置默认卡
	sqlStr := GetUserUsdtsDefaultCardSql()
	Rows, err := conn.Query(sqlStr, UserInfo.Id, userUsdtsForm.AddreType)
	if err != nil {
		connTx.Rollback()
		logger.Warn(fmt.Sprintf("Err:[GetUserUsdtsDefaultCard Query Sql failed.] SQL:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err.Error()))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}

	var userUsdtsGet UserUsdtsReq
	for Rows.Next() {
		var userUsdts models.UserUsdts
		Rows.Scan(&userUsdtsGet.UsdtId, &userUsdtsGet.UserId, &userUsdtsGet.AddreType, &userUsdtsGet.UsdtAddre, &userUsdtsGet.UsdtTitle, &userUsdtsGet.DefaultCard)
		userUsdtsGet = ConvUserUsdtsNull2Json(userUsdts, userUsdtsGet)
		break
	}

	//如果存在默认卡先将默认卡修改为普通卡，否则跳过
	if userUsdtsGet.UsdtId > 0 {
		sqlStr = GetCreateInUpdateUserUsdtsSql()
		Ret, err := connTx.Exec(sqlStr, 0, userUsdtsGet.UsdtId, UserInfo.Id)
		if err != nil {
			connTx.Rollback()
			logger.Warn(fmt.Sprintf("Exec SQL Failed:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err))
			code := 500
			msg = "网络繁忙"
			config.Errors(ctx, code, msg)
			return
		}
		RowsAff2, err := Ret.RowsAffected()
		if err != nil || RowsAff2 != 1 {
			connTx.Rollback()
			logger.Warn(fmt.Sprintf("Get LastInsertId Failed:[%s] RowsAff2:[%v] Time:[%s] Err:[%v]", sqlStr, RowsAff2, time.Now().Format(time.DateTime), err.Error()))
			code := 500
			msg = "网络繁忙"
			config.Errors(ctx, code, msg)
			return
		}
	} else {
		userUsdtsForm.DefaultCard = 1
	}

	//创建USDT地址
	sqlStr = GetCreateUserUsdtsSql()
	Ret, err := connTx.Exec(sqlStr, UserInfo.Id, userUsdtsForm.AddreType, userUsdtsForm.UsdtAddre, userUsdtsForm.UsdtTitle, userUsdtsForm.DefaultCard, time.Now().Format(time.DateTime))
	if err != nil {
		connTx.Rollback()
		logger.Warn(fmt.Sprintf("Exec SQL Failed:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	LastId, err := Ret.LastInsertId()
	if err != nil {
		connTx.Rollback()
		logger.Warn(fmt.Sprintf("Get LastInsertId Failed:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}

	if LastId > 0 {
		connTx.Commit()
	} else {
		connTx.Rollback()
		logger.Warn(fmt.Sprintf("CreateUserUsdts Failed, LastId:[%v] Time:[%s]", LastId, time.Now().Format(time.DateTime)))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}

	//设置缓存
	if userUsdtsForm.DefaultCard == 1 {
		//获取用户USDT默认地址
		sqlStr = "select usdtId,userId,addreType,usdtAddre,usdtTitle,defaultCard,addTime,updateTime from " + TableUserUsdts + " where userId=? and addreType=? and defaultCard=? and deleteTime is null"
		Rows, err = conn.Query(sqlStr, UserInfo.Id, userUsdtsForm.AddreType, 1)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer Rows.Close()
		var userUsdtsJson models.UserUsdtsJson
		for Rows.Next() {
			var userUsdts models.UserUsdts
			Rows.Scan(&userUsdts.UsdtId, &userUsdts.UserId, &userUsdts.AddreType, &userUsdts.UsdtAddre, &userUsdts.UsdtTitle, &userUsdts.DefaultCard, &userUsdts.AddTime)
			userUsdtsJson.UsdtId = userUsdts.UsdtId.Int64
			userUsdtsJson.UserId = userUsdts.UserId.Int64
			userUsdtsJson.AddreType = userUsdts.AddreType.Int32
			userUsdtsJson.UsdtAddre = userUsdts.UsdtAddre.String
			userUsdtsJson.UsdtTitle = userUsdts.UsdtTitle.String
			userUsdtsJson.DefaultCard = userUsdts.DefaultCard.Int32
			userUsdtsJson.AddTime = userUsdts.AddTime.String
			break
		}
		var key = ""
		if userUsdtsJson.AddreType == 1 {
			key = fmt.Sprintf("%s%s", constant.UsdtsDefaultErcKey, UserInfo.PhoneNumber)
		} else if userUsdtsJson.AddreType == 2 {
			key = fmt.Sprintf("%s%s", constant.UsdtsDefaultTrcKey, UserInfo.PhoneNumber)
		}
		userUsdtsMarshal, err := json.Marshal(&userUsdtsJson)
		if err != nil {
			logger.Warn(fmt.Sprintf("userId:[%v] usdtId:[%v]  默认USDT卡，缓存设设置失败--json解析失败 Time:[%s] Err:[%v]", userUsdtsJson.UserId, userUsdtsJson.UsdtId, time.Now().Format(time.DateTime), err.Error()))
		}
		err = client.Set(context.Background(), key, userUsdtsMarshal, 0).Err()
		if err != nil {
			logger.Warn(fmt.Sprintf("userId:[%v] usdtId:[%v]  默认USDT卡，缓存设设置失败 Time:[%s] Err:[%v]", userUsdtsJson.UserId, userUsdtsJson.UsdtId, time.Now().Format(time.DateTime), err.Error()))
		}
	}

	config.Success(ctx, http.StatusOK, msg, LastId)
	return
}

func DeleteUserUsdts(ctx *gin.Context) {
	//mysql初始化方式
	conn = inits.InitMysql()
	defer conn.Close() //关闭数据库连接
	var deleteForm UserUsdtsDeleteForm
	var msg any

	//redis初始化方式
	client = inits.InitRedis()
	UserInfo = global.GetUserInfo(client)

	//1、获取参数，并验证
	if errs := ctx.ShouldBindJSON(&deleteForm); errs != nil {
		code := 500
		err, ok := errs.(validator.ValidationErrors)
		msg = err.Translate(translators.Trans)
		if !ok {
			msg = errs.Error()
		}
		logger.Warn(fmt.Sprintf("Time:[%s] Err:[%v]", time.Now().Format(time.DateTime), err))
		config.Errors(ctx, code, msg)
		return
	}

	//查询是否为默认卡
	sqlStr := "select usdtId,userId,addreType,usdtAddre,usdtTitle,defaultCard from bo_user_usdts where usdtId=? and userId=? and defaultCard=1 and deleteTime is null limit 1"
	Rows, err := conn.Query(sqlStr, deleteForm.UsdtId, UserInfo.Id)
	if err != nil {
		logger.Warn(fmt.Sprintf("Err:[GetUserUsdtsDefaultCard Query Sql failed.] SQL:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err.Error()))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	defer Rows.Close()

	var userUsdtsGet UserUsdtsReq
	for Rows.Next() {
		var userUsdts models.UserUsdts
		Rows.Scan(&userUsdts.UsdtId, &userUsdts.UserId, &userUsdts.AddreType, &userUsdts.UsdtAddre, &userUsdts.UsdtTitle, &userUsdts.DefaultCard)
		userUsdtsGet = ConvUserUsdtsNull2Json(userUsdts, userUsdtsGet)
		break
	}

	var userUsdtsUp UserUsdtsReq
	if userUsdtsGet.UsdtId > 0 {
		//更改默认卡-查询一条新数据作为默认卡
		sqlStr = "select usdtId,userId,addreType,usdtAddre,usdtTitle,defaultCard from bo_user_usdts where usdtId not in(?) and userId=? and addreType=? and defaultCard=0 and deleteTime is null limit 1"
		Rows, err = conn.Query(sqlStr, userUsdtsGet.UsdtId, UserInfo.Id, userUsdtsGet.AddreType)
		if err != nil {
			logger.Warn(fmt.Sprintf("Err:[GetUserUsdtsDefaultCard Query Sql failed.] SQL:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err.Error()))
			code := 500
			msg = "网络繁忙"
			config.Errors(ctx, code, msg)
			return
		}
		defer Rows.Close()
		for Rows.Next() {
			var userUsdts models.UserUsdts
			Rows.Scan(&userUsdts.UsdtId, &userUsdts.UserId, &userUsdts.AddreType, &userUsdts.UsdtAddre, &userUsdts.UsdtTitle, &userUsdts.DefaultCard)
			userUsdtsUp = ConvUserUsdtsNull2Json(userUsdts, userUsdtsUp)
			break
		}
	}

	connTx, err := conn.Begin()
	if err != nil {
		connTx.Rollback()
		logger.Warn(fmt.Sprintf("Err:[事务开启失败] Time:[%s] Err:[%v]", time.Now().Format(time.DateTime), err.Error()))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}

	if userUsdtsUp.UsdtId > 0 {
		//更改默认卡-更改默认卡
		sqlStr = "update bo_user_usdts set defaultCard=1 where usdtId=? and userId=? and addreType=? and deleteTime is null"
		Ret, err := connTx.Exec(sqlStr, userUsdtsUp.UsdtId, UserInfo.Id, userUsdtsUp.AddreType)
		if err != nil {
			connTx.Rollback()
			logger.Warn(fmt.Sprintf("Exec SQL Failed:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err))
			code := 500
			msg = "网络繁忙"
			config.Errors(ctx, code, msg)
			return
		}
		Row, err := Ret.RowsAffected()
		if err != nil || Row <= 0 {
			connTx.Rollback()
			logger.Warn(fmt.Sprintf("Get RowsAffected Failed:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err))
			code := 500
			msg = "网络繁忙"
			config.Errors(ctx, code, msg)
			return
		}
		//修改缓存数据
		var key = ""
		if userUsdtsUp.AddreType == 1 {
			key = fmt.Sprintf("%s%s", constant.UsdtsDefaultErcKey, UserInfo.Id)
		} else if userUsdtsUp.AddreType == 2 {
			key = fmt.Sprintf("%s%s", constant.UsdtsDefaultTrcKey, UserInfo.Id)
		}
		userUsdtsUp.DefaultCard = 1
		userUsdtsUpMarshal, err := json.Marshal(&userUsdtsUp)
		if err != nil {
			connTx.Rollback()
			logger.Warn(fmt.Sprintf("userId:[%v] usdtId:[%v]  默认USDT卡，缓存设设置失败--json解析失败 Time:[%s] Err:[%v]", userUsdtsUp.UserId, userUsdtsUp.UsdtId, time.Now().Format(time.DateTime), err.Error()))
			code := 500
			msg = "网络繁忙"
			config.Errors(ctx, code, msg)
			return
		}
		err = client.Set(context.Background(), key, userUsdtsUpMarshal, 0).Err()
		if err != nil {
			connTx.Rollback()
			logger.Warn(fmt.Sprintf("userId:[%v] usdtId:[%v]  默认USDT卡，缓存设设置失败 Time:[%s] Err:[%v]", userUsdtsUp.UserId, userUsdtsUp.UsdtId, time.Now().Format(time.DateTime), err.Error()))
			code := 500
			msg = "网络繁忙"
			config.Errors(ctx, code, msg)
			return
		}
	}

	sqlStr = GetDeleteUserUsdtsSql()
	Ret, err := conn.Exec(sqlStr, time.Now().Format(time.DateTime), deleteForm.UsdtId, UserInfo.Id)
	if err != nil {
		connTx.Rollback()
		logger.Warn(fmt.Sprintf("Exec SQL Failed:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	Row, err := Ret.RowsAffected()
	if err != nil {
		connTx.Rollback()
		logger.Warn(fmt.Sprintf("Get RowsAffected Failed:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}

	if Row > 0 {
		connTx.Commit()
	} else {
		connTx.Rollback()
		logger.Warn(fmt.Sprintf("DeleteUserUsdts Failed, RowsAffected:[%v] SQL:[%s] 参数:[usdtId:%v,UsdtId:%v] Time:[%s]", Row, sqlStr, deleteForm.UsdtId, UserInfo.Id, time.Now().Format(time.DateTime)))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}

	config.Success(ctx, http.StatusOK, msg, Row)
	return
}

// GetUserUsdts 获取默认用户USDT地址
func GetUserUsdts(ctx *gin.Context) {
	//mysql初始化方式
	conn = inits.InitMysql()
	defer conn.Close() //关闭数据库连接
	var msg any

	//redis初始化方式
	client = inits.InitRedis()
	UserInfo = global.GetUserInfo(client)

	usdtId := ctx.Query("usdtId")
	if usdtId == "" {
		logger.Warn(fmt.Sprintf("Err:[QueryForm usdtId parameter error.] Time:[%s] ", time.Now().Format(time.DateTime)))
		code := 500
		msg = "参数错误"
		config.Errors(ctx, code, msg)
		return
	}
	var userUsdtsReq UserUsdtsReq
	sqlStr := GetUserUsdtsSql()
	Rows, err := conn.Query(sqlStr, usdtId, UserInfo.Id, 1)
	if err != nil {
		logger.Warn(fmt.Sprintf("Err:[Query Sql failed.] SQl:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err.Error()))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	defer Rows.Close()

	for Rows.Next() {
		var userUsdts models.UserUsdts
		Rows.Scan(&userUsdts.UsdtId, &userUsdts.UserId, &userUsdts.AddreType, &userUsdts.UsdtAddre, &userUsdts.UsdtTitle, &userUsdts.DefaultCard)
		userUsdtsReq = ConvUserUsdtsNull2Json(userUsdts, userUsdtsReq)
		break
	}
	config.Success(ctx, http.StatusOK, msg, &userUsdtsReq)
	return
}

// GetListUserUsdts 获取用户USDT地址列表
func GetListUserUsdts(ctx *gin.Context) {
	//mysql初始化方式
	conn = inits.InitMysql()
	defer conn.Close() //关闭数据库连接
	var msg any

	//redis初始化方式
	client = inits.InitRedis()
	UserInfo = global.GetUserInfo(client)

	sqlStr := GetUserAllUsdtsSql()
	fmt.Println(sqlStr)
	Rows, err := conn.Query(sqlStr, UserInfo.Id)
	if err != nil {
		logger.Warn(fmt.Sprintf("Err:[Query Sql failed.] SQl:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err.Error()))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	defer Rows.Close() //关闭查询

	var userUsdtsMap []*UserUsdtsReq
	for Rows.Next() {
		var userUsdtsReq UserUsdtsReq
		var userUsdts models.UserUsdts
		Rows.Scan(&userUsdts.UsdtId, &userUsdts.UserId, &userUsdts.AddreType, &userUsdts.UsdtAddre, &userUsdts.UsdtTitle, &userUsdts.DefaultCard)
		userUsdtsReq = ConvUserUsdtsNull2Json(userUsdts, userUsdtsReq)
		userUsdtsMap = append(userUsdtsMap, &userUsdtsReq)
	}

	config.Success(ctx, http.StatusOK, msg, &userUsdtsMap)
	return
}

func ConvUserUsdtsNull2Json(userUsdts models.UserUsdts, userUsdtsGet UserUsdtsReq) UserUsdtsReq {
	userUsdtsGet.UsdtId = userUsdts.UsdtId.Int64
	userUsdtsGet.UserId = userUsdts.UserId.Int64
	userUsdtsGet.AddreType = userUsdts.AddreType.Int32
	userUsdtsGet.UsdtAddre = userUsdts.UsdtAddre.String
	userUsdtsGet.UsdtTitle = userUsdts.UsdtTitle.String
	userUsdtsGet.DefaultCard = userUsdts.DefaultCard.Int32
	return userUsdtsGet
}
