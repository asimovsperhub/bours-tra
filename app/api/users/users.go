package users

import (
	"Bourse/app/models"
	"Bourse/common/global"
	"Bourse/common/helper"
	"Bourse/common/inits"
	"Bourse/common/translators"
	"Bourse/config"
	"Bourse/config/constant"
	"database/sql"
	"encoding/json"
	"fmt"
	"gitee.com/phper95/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strconv"
	"time"
)

var (
	conn   *sql.DB
	client *redis.Client

	UserInfo *constant.UserInfoStruct
)

const (
	AccountTypeJson   = `[{"accountType":"1","accountName":"现货账户"},{"accountType":"2","accountName":"合约账户"},{"accountType":"3","accountName":"马股账户"},{"accountType":"4","accountName":"美股账户"},{"accountType":"5","accountName":"日股账户"},{"accountType":"6","accountName":"佣金账户"}]`
	WithdrawalChannel = `[{"channel":1,"channelName":"ERC-20 USDT"},{"channel":2,"channelName":"TRC-20 USDT"},{"channel":3,"channelName":"银行卡"}]`
)

type WithdrawMoneyInfo struct {
	AccountTypeJson   any `json:"accountTypeJson"`
	WithdrawalChannel any `json:"withdrawalChannel"`
	UserBankcardsInfo UserBankcardsReq
	UserUsdtsTrcInfo  *UserUsdtsReq
	UserUsdtsErcInfo  *UserUsdtsReq
}

type UserBankcardsReq struct {
	BankCardId     int64  `json:"bankCardId"`
	BankCardNumber string `json:"bankCardNumber"`
}

type UserUsdtsReq struct {
	AddreType   int32  `json:"addreType"`
	UsdtAddre   string `json:"usdtAddre"`
	UsdtTitle   string `json:"usdtTitle"`
	DefaultCard int32  `json:"defaultCard"`
}

type UserUccountBalanceList struct {
	TotalAmount                   float64                            `json:"totalAmount"` //总金额
	UserFundAccountsSpotsJson     *models.UserFundAccountsJson       `json:"userSpotsJson"`
	UserFundAccountsSwapsJson     *models.UserFundAccountsJson       `json:"userSwapsJson"`
	UserFundAccountsMalaysiaJson  *models.UserFundAccountsJson       `json:"userMalaysiaJson"`
	UserFundAccountsUSJson        *models.UserFundAccountsJson       `json:"userUSJson"`
	UserFundAccountsYenJson       *models.UserFundAccountsJson       `json:"userYenJson"`
	UserFundAccountsBrokerageJson *models.UserFundAccountsJson       `json:"userBrokerageJson"`
	SpotsJsonMap                  []*models.UserFundAccountSpotsJson `json:"spotsJsonMap"`
}

type UserTerminalEquipmentsList struct {
	List     []*models.UserTerminalEquipmentsJson `json:"list"`
	PageNum  int                                  `json:"pageNum"`
	PageSize int                                  `json:"pageSize"`
	Total    int32                                `json:"total"`
}

func init() {
	//初始化配置方式
	config.LoadConfig()
	inits.InitLog()

	//redis初始化方式
	client = inits.InitRedis()

	//多语言
	if err := translators.InitTranslators("zh"); err != nil {
		fmt.Sprintf("init Translator failed,err:%v\n", err)
	}
}

// GetWithdrawMoneyInfo 获取提款账户类型、渠道-默认账户信息
func GetWithdrawMoneyInfo(ctx *gin.Context) {
	//mysql初始化方式
	conn = inits.InitMysql()
	defer conn.Close() //关闭数据库连接
	var msg any

	UserInfo = global.GetUserInfo(client)

	var withdrawMoneyInfo WithdrawMoneyInfo
	_ = json.Unmarshal([]byte(AccountTypeJson), &withdrawMoneyInfo.AccountTypeJson)
	_ = json.Unmarshal([]byte(WithdrawalChannel), &withdrawMoneyInfo.WithdrawalChannel)

	//1 获取默认银行卡
	sqlStr := GetUserBankcardsSql()
	Rows, err := conn.Query(sqlStr, UserInfo.Id, 1)
	if err != nil {
		logger.Warn(fmt.Sprintf("Err:[Query Sql failed.] SQl:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err.Error()))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	defer Rows.Close()
	for Rows.Next() {
		var userBankcards models.UserBankcards
		Rows.Scan(&userBankcards.BankCardId, &userBankcards.BankCardNumber)
		withdrawMoneyInfo.UserBankcardsInfo.BankCardId = userBankcards.BankCardId.Int64
		withdrawMoneyInfo.UserBankcardsInfo.BankCardNumber = helper.ParseBank(userBankcards.BankCardNumber.String)
		break
	}

	//2 获取默认USDT地址
	sqlStr = GetUserUsdtsDefaultSql()
	Rows, err = conn.Query(sqlStr, UserInfo.Id, 1)
	if err != nil {
		logger.Warn(fmt.Sprintf("Err:[Query Sql failed.] SQl:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err.Error()))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	defer Rows.Close()
	var userUsdtsMap []*UserUsdtsReq
	for Rows.Next() {
		var userUsdts models.UserUsdts
		var userUsdtsReq UserUsdtsReq
		Rows.Scan(&userUsdts.AddreType, &userUsdts.UsdtAddre, &userUsdts.UsdtTitle, &userUsdts.DefaultCard)
		userUsdtsReq.AddreType = userUsdts.AddreType.Int32
		userUsdtsReq.UsdtAddre = userUsdts.UsdtAddre.String
		userUsdtsReq.UsdtTitle = userUsdts.UsdtTitle.String
		userUsdtsReq.DefaultCard = userUsdts.DefaultCard.Int32
		userUsdtsMap = append(userUsdtsMap, &userUsdtsReq)
	}

	for _, usdts := range userUsdtsMap {
		switch usdts.AddreType {
		case 1:
			withdrawMoneyInfo.UserUsdtsErcInfo = usdts
		case 2:
			withdrawMoneyInfo.UserUsdtsTrcInfo = usdts
		}
	}

	config.Success(ctx, 200, "ok", &withdrawMoneyInfo)
	return
}

func GetUserAccountBalance(ctx *gin.Context) {
	//mysql初始化方式 === 获取对应资金账户余额
	conn = inits.InitMysql()
	defer conn.Close() //关闭数据库连接

	//获取请求参数
	fundAccountTypeStr := ctx.Query("fundAccountType")
	fundAccountName := ctx.Query("fundAccountName")
	fundAccountType, _ := strconv.Atoi(fundAccountTypeStr)

	//账户总余额  获取现货 获取合约 获取股票
	if fundAccountType == 1 {
		//获取现货子账号
		sqlStr := "select id,fundAccountsId,userId,fundAccountSpot,fundAccountName,spotMoney,unit from bo_user_fund_account_spots where userId=? and fundAccountName=? limit 1"
		rows, err := conn.Query(sqlStr, UserInfo.Id, 1, fundAccountName)
		if err != nil {
			logger.Warn(fmt.Sprintf("Err:[Query Sql failed.] SQl:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err.Error()))
			code := 500
			msg := "网络繁忙"
			config.Errors(ctx, code, msg)
			return
		}
		defer rows.Close()
		var spotsJson models.UserFundAccountSpotsJson
		for rows.Next() {
			var spots models.UserFundAccountSpots
			rows.Scan(&spots.Id, &spots.FundAccountsId, &spots.UserId, &spots.FundAccountSpot, &spots.FundAccountName, &spots.SpotMoney, &spots.Unit)
			spotsJson = models.ConvUserFundAccountSpotsNull2Json(spots, spotsJson)
			break
		}

		config.Success(ctx, http.StatusOK, "请求成功", spotsJson)
		return
	} else {
		//马股资产/美股资产/日股资产 === 获取fundAccountType资产账号
		sqlStr := "select id,userId,fundAccount,fundAccountType,money,unit,addTime,updateTime from bo_user_fund_accounts where userId=? and fundAccountType=? limit 1"
		rows, err := conn.Query(sqlStr, UserInfo.Id, fundAccountType)
		if err != nil {
			logger.Warn(fmt.Sprintf("Err:[Query Sql failed.] SQl:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err.Error()))
			code := 500
			msg := "网络繁忙"
			config.Errors(ctx, code, msg)
			return
		}
		defer rows.Close()
		var swapsJson models.UserFundAccountsJson
		for rows.Next() {
			var userFundAccounts models.UserFundAccounts
			rows.Scan(&userFundAccounts.Id, &userFundAccounts.UserId, &userFundAccounts.FundAccount, &userFundAccounts.FundAccountType, &userFundAccounts.Money, &userFundAccounts.Unit)
			swapsJson = models.ConvUserFundAccountsNull2Json(userFundAccounts, swapsJson)
			break
		}

		config.Success(ctx, http.StatusOK, "请求成功", swapsJson)
		return
	}
	return
}

func GetUserAccountsBalanceList(ctx *gin.Context) {
	//mysql初始化方式 === 获取对应资金账户余额
	conn = inits.InitMysql()
	defer conn.Close() //关闭数据库连接

	UserInfo = global.GetUserInfo(client)

	//获取现货子账号列表
	/*sqlStr := "select id,fundAccountsId,userId,fundAccountSpot,fundAccountName,spotMoney,unit from bo_user_fund_account_spots where userId=?"
	rows, err := conn.Query(sqlStr, UserInfo.Id)
	if err != nil {
		logger.Warn(fmt.Sprintf("Err:[Query Sql failed.] SQl:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err.Error()))
		code := 500
		msg := "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	defer rows.Close()
	var spotsJsonMap []*models.UserFundAccountSpotsJson
	for rows.Next() {
		var spotsJson models.UserFundAccountSpotsJson
		var spots models.UserFundAccountSpots
		rows.Scan(&spots.Id, &spots.FundAccountsId, &spots.UserId, &spots.FundAccountSpot, &spots.FundAccountName, &spots.SpotMoney, &spots.Unit)
		spotsJson = models.ConvUserFundAccountSpotsNull2Json(spots, spotsJson)
		spotsJsonMap = append(spotsJsonMap, &spotsJson)
	}*/

	sqlStr := "select id,userId,fundAccountType,fundAccount,money,unit,addTime,updateTime from bo_user_fund_accounts where userId=? and fundAccountType in(3,4,5)"
	rows, err := conn.Query(sqlStr, UserInfo.Id)
	if err != nil {
		logger.Warn(fmt.Sprintf("Err:[Query Sql failed.] SQl:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err.Error()))
		code := 500
		msg := "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	defer rows.Close()
	var userUccountBalanceList UserUccountBalanceList
	for rows.Next() {
		var userFundAccounts models.UserFundAccounts
		var userFund models.UserFundAccountsJson
		rows.Scan(&userFundAccounts.Id, &userFundAccounts.UserId, &userFundAccounts.FundAccountType, &userFundAccounts.FundAccount, &userFundAccounts.Money, &userFundAccounts.Unit, &userFundAccounts.AddTime, &userFundAccounts.UpdateTime)
		userFund = models.ConvUserFundAccountsNull2Json(userFundAccounts, userFund)
		if userFund.FundAccountType == 3 { //马来户
			userUccountBalanceList.UserFundAccountsMalaysiaJson = &userFund
		}
		if userFund.FundAccountType == 4 { //美户
			userUccountBalanceList.UserFundAccountsUSJson = &userFund
		}
		if userFund.FundAccountType == 5 { //日户
			userUccountBalanceList.UserFundAccountsYenJson = &userFund
		}
	}

	sqlStr = "select (select exchangeRate from bo_system_exchange_rates where sourceCode='USD' and targetCode='MYR' limit 1) as myr, (select exchangeRate from bo_system_exchange_rates where sourceCode='USD' and targetCode='JPY' limit 1) as jpy from bo_system_exchange_rates limit 1"
	rows, err = conn.Query(sqlStr)
	if err != nil {
		logger.Warn(fmt.Sprintf("Err:[Query Sql failed.] SQl:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err.Error()))
		code := 500
		msg := "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	defer rows.Close()
	var myr, jpy float64
	for rows.Next() {
		rows.Scan(&myr, &jpy)
		break
	}

	userUccountBalanceList.TotalAmount = userUccountBalanceList.UserFundAccountsMalaysiaJson.Money*myr + userUccountBalanceList.UserFundAccountsUSJson.Money + userUccountBalanceList.UserFundAccountsYenJson.Money*jpy

	//var userUccountBalanceList UserUccountBalanceList
	//userUccountBalanceList.UserFundAccountsSpotsJson = GetFundAccountsFind(ctx, conn, 1)
	//userUccountBalanceList.UserFundAccountsSwapsJson = GetFundAccountsFind(ctx, conn, 2)
	/*userUccountBalanceList.UserFundAccountsMalaysiaJson = GetFundAccountsFind(ctx, conn, 3)
	userUccountBalanceList.UserFundAccountsUSJson = GetFundAccountsFind(ctx, conn, 4)
	userUccountBalanceList.UserFundAccountsYenJson = GetFundAccountsFind(ctx, conn, 5)*/
	//userUccountBalanceList.UserFundAccountsBrokerageJson = GetFundAccountsFind(ctx, conn, 6)
	//userUccountBalanceList.SpotsJsonMap = spotsJsonMap

	//userUccountBalanceList.TotalAmount = userUccountBalanceList.UserFundAccountsSpotsJson.Money + userUccountBalanceList.UserFundAccountsSwapsJson.Money + userUccountBalanceList.UserFundAccountsMalaysiaJson.Money + userUccountBalanceList.UserFundAccountsUSJson.Money + userUccountBalanceList.UserFundAccountsYenJson.Money + userUccountBalanceList.UserFundAccountsBrokerageJson.Money

	config.Success(ctx, http.StatusOK, "请求成功", userUccountBalanceList)
	return
}

// GetUserTerminalEquipmentsList 获取用户设备信息列表
func GetUserTerminalEquipmentsList(ctx *gin.Context) {
	var msg any
	UserInfo = global.GetUserInfo(client)

	conn = inits.InitMysql()
	defer conn.Close()

	pageNum, _ := strconv.Atoi(ctx.DefaultQuery("pageNum", "1"))   //获取分页数量 pageNum
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "1")) //获取当前页 pageSize total

	//计算总数
	sqlStr := "select count(1) as total from bo_user_terminal_equipments where userId=?"
	rows, err := conn.Query(sqlStr, UserInfo.Id)
	if err != nil {
		code := 500
		msg = "系统内部错误"
		config.Errors(ctx, code, msg)
		return
	}
	defer rows.Close()
	var Total int32
	for rows.Next() {
		rows.Scan(&Total)
	}

	sqlStr = "select terminalEquipmentId,userId,terminalEquipmentName,loginCountry,loginCity,loginIp,addTime from bo_user_terminal_equipments where userId=? limit ? offset ?"
	rows, err = conn.Query(sqlStr, UserInfo.Id, pageSize, (pageNum-1)*pageSize)
	if err != nil {
		code := 500
		msg = "系统内部错误"
		config.Errors(ctx, code, msg)
		return
	}
	defer rows.Close()

	var userTerminalEquipmentsJsonMap []*models.UserTerminalEquipmentsJson
	for rows.Next() {
		var terminalEquipments models.UserTerminalEquipments
		var terminalEquipmentsJson models.UserTerminalEquipmentsJson
		rows.Scan(&terminalEquipments.TerminalEquipmentId, &terminalEquipments.UserId, &terminalEquipments.TerminalEquipmentName, &terminalEquipments.LoginCountry, &terminalEquipments.LoginCity, &terminalEquipments.LoginIp, &terminalEquipments.AddTime)
		terminalEquipmentsJson = models.ConvUserTerminalEquipmentsNull2Json(terminalEquipments, terminalEquipmentsJson)
		userTerminalEquipmentsJsonMap = append(userTerminalEquipmentsJsonMap, &terminalEquipmentsJson)
	}
	var userTerminalEquipmentsListList UserTerminalEquipmentsList
	userTerminalEquipmentsListList.List = userTerminalEquipmentsJsonMap
	userTerminalEquipmentsListList.Total = Total
	userTerminalEquipmentsListList.PageNum = pageNum
	userTerminalEquipmentsListList.PageSize = pageSize

	config.Success(ctx, http.StatusOK, "请求成功", userTerminalEquipmentsListList)
	return
}

func GetFundAccountsFind(ctx *gin.Context, conn *sql.DB, fundAccountType int) *models.UserFundAccountsJson {
	sqlStr := "select id,userId,fundAccountType,fundAccount,money,unit,addTime,updateTime from bo_user_fund_accounts where userId=? and fundAccountType=? limit 1"
	rows, err := conn.Query(sqlStr, UserInfo.Id, fundAccountType)
	if err != nil {
		logger.Warn(fmt.Sprintf("Err:[Query Sql failed.] SQl:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err.Error()))
		code := 500
		msg := "网络繁忙"
		config.Errors(ctx, code, msg)
		return nil
	}
	defer rows.Close()
	var userFundAccountsMalaysiaJson models.UserFundAccountsJson
	for rows.Next() {
		var userFundAccounts models.UserFundAccounts
		rows.Scan(&userFundAccounts.Id, &userFundAccounts.UserId, &userFundAccounts.FundAccountType, &userFundAccounts.FundAccount, &userFundAccounts.Money, &userFundAccounts.Unit, &userFundAccounts.AddTime, &userFundAccounts.UpdateTime)
		userFundAccountsMalaysiaJson = models.ConvUserFundAccountsNull2Json(userFundAccounts, userFundAccountsMalaysiaJson)
		break
	}
	return &userFundAccountsMalaysiaJson
}
