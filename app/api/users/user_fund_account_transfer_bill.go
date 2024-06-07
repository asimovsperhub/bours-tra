package users

import (
	"Bourse/app/models"
	"Bourse/common/global"
	"Bourse/common/inits"
	"Bourse/common/ordernum"
	"Bourse/common/translators"
	"Bourse/config"
	"fmt"
	"gitee.com/phper95/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
	"strconv"
	"time"
)

type UserFundAccountTransferBillForm struct {
	UserId         int64   `json:"userId" form:"userId"`                                    //用户ID
	TransferId     int64   `json:"transferId" form:"transferId" binding:"required"`         //被划账户ID
	TransferAmount float64 `json:"transferAmount" form:"transferAmount" binding:"required"` //划转金额
	CollectId      int64   `json:"collectId" form:"collectId" binding:"required"`           //收款账户ID
	CollectAmount  float64 `json:"collectAmount" form:"collectAmount" binding:"required"`   //划转收款金额
	ExchangeRate   float64 `json:"exchangeRate" form:"exchangeRate" binding:"required"`     //汇率
}

type SystemExchangeRates struct {
	SourceCode     string  `json:"sourceCode" form:"sourceCode" binding:"required"` //原货币简码
	TargetCode     string  `json:"targetCode" form:"targetCode" binding:"required"` //目标货币简码
	TransferAmount float64 `json:"transferAmount" form:"transferAmount"`            //划转金额
	CollectAmount  float64 `json:"collectAmount" form:"collectAmount"`              //划转收款金额
	ExchangeRate   float64 `json:"exchangeRate" form:"exchangeRate"`                //汇率
}

type FundAccountTransferList struct {
	List     []*models.UserFundAccountTransferBillJson `json:"list"`
	PageNum  int                                       `json:"pageNum"`
	PageSize int                                       `json:"pageSize"`
	Total    int32                                     `json:"total"`
}

// PostFundAccountTransfer 用户账户资金划转
func PostFundAccountTransfer(ctx *gin.Context) {
	var billForm UserFundAccountTransferBillForm
	var msg any

	UserInfo = global.GetUserInfo(client)
	//1、获取参数，并验证
	if errs := ctx.ShouldBindJSON(&billForm); errs != nil {
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
	conn = inits.InitMysql()
	defer conn.Close()

	//开启事务
	Tx, err := conn.Begin()
	if err != nil {
		logger.Warn(fmt.Sprintf("Err:[开启事务失败 failed.] Time:[%s] Err:[%v]", time.Now().Format(time.DateTime), err.Error()))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	//查询被划转账户资金
	sqlStr := "select id,userId,fundAccount,fundAccountType,money,unit,addTime,updateTime from bo_user_fund_accounts where id in (?,?) and userId=?"
	rows, err := Tx.Query(sqlStr, billForm.TransferId, billForm.CollectId, UserInfo.Id)
	if err != nil {
		Tx.Rollback()
		logger.Warn(fmt.Sprintf("Err:[Query Sql failed.] SQl:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err.Error()))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}

	var userFundAccountsJson1 models.UserFundAccountsJson
	var userFundAccountsJson2 models.UserFundAccountsJson
	for rows.Next() {
		var userFundAccounts models.UserFundAccounts
		rows.Scan(&userFundAccounts.Id, &userFundAccounts.UserId, &userFundAccounts.FundAccount, &userFundAccounts.FundAccountType, &userFundAccounts.Money, &userFundAccounts.Unit, &userFundAccounts.AddTime, &userFundAccounts.UpdateTime)
		if billForm.TransferId == userFundAccounts.Id.Int64 { //被划账户
			userFundAccountsJson1 = models.ConvUserFundAccountsNull2Json(userFundAccounts, userFundAccountsJson1)
		}
		if billForm.CollectId == userFundAccounts.Id.Int64 { //收款账户
			userFundAccountsJson2 = models.ConvUserFundAccountsNull2Json(userFundAccounts, userFundAccountsJson2)
		}
	}

	//处理错误
	if err = rows.Err(); err != nil {
		rows.Close()
		Tx.Rollback()
		logger.Warn(fmt.Sprintf("Err:[查询账户数据失败 failed.] SQl:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err.Error()))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	if userFundAccountsJson1.Id == 0 || userFundAccountsJson2.Id == 0 {
		Tx.Rollback()
		logger.Warn(fmt.Sprintf("Err:[被划转账户不存在 failed.] SQl:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err.Error()))
		code := 500
		msg = "查询资金账户失败"
		config.Errors(ctx, code, msg)
		return
	}

	//判断资金余额是否足够
	if userFundAccountsJson1.Money < billForm.TransferAmount {
		logger.Warn(fmt.Sprintf("Err:[被划转账户余额不足 failed.] SQl:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err.Error()))
		code := 500
		msg = "被划转账户余额不足"
		config.Errors(ctx, code, msg)
		return
	}

	newTime := time.Now().Format(time.DateTime)
	money1 := userFundAccountsJson1.Money - billForm.TransferAmount
	money2 := userFundAccountsJson2.Money + billForm.CollectAmount

	//修改资金
	sqlStr1 := "update bo_user_fund_accounts set money=?,updateTime=? where id=? and userId=?"
	exec, err := Tx.Exec(sqlStr1, money1, newTime, userFundAccountsJson1.Id, userFundAccountsJson1.UserId)
	if err != nil {
		Tx.Rollback()
		code := 500
		msg = "划转失败"
		config.Errors(ctx, code, msg)
		return
	}
	rowsAffected1, err := exec.RowsAffected()
	if err != nil || rowsAffected1 != 1 {
		Tx.Rollback()
		code := 500
		msg = "划转失败"
		config.Errors(ctx, code, msg)
		return
	}

	result, err := Tx.Exec(sqlStr1, money2, newTime, userFundAccountsJson2.Id, userFundAccountsJson2.UserId)
	if err != nil {
		Tx.Rollback()
		code := 500
		msg = "划转失败"
		config.Errors(ctx, code, msg)
		return
	}
	rowsAffected2, err := result.RowsAffected()
	if err != nil || rowsAffected2 != 1 {
		Tx.Rollback()
		code := 500
		msg = "划转失败"
		config.Errors(ctx, code, msg)
		return
	}
	//fmt.Println(money2, newTime, userFundAccountsJson2.Id, userFundAccountsJson2.UserId)

	//写入划转流水表
	sqlStr = "insert into bo_user_fund_account_transfer_bill (userId,transferId,transferAmount,transferFundAccountType,transferUnit,collectId,collectAmount,collectFundAccountType,collectUnit,orderNo,exchangeRate,status,addTime) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)"
	exec1, err := Tx.Exec(sqlStr, UserInfo.Id, billForm.TransferId, billForm.TransferAmount, userFundAccountsJson1.FundAccountType, userFundAccountsJson1.Unit, billForm.CollectId, billForm.CollectAmount, userFundAccountsJson2.FundAccountType, userFundAccountsJson2.Unit, ordernum.GetOrderNo(), billForm.ExchangeRate, 1, newTime)
	if err != nil {
		Tx.Rollback()
		code := 500
		msg = "写入划转流水失败"
		config.Errors(ctx, code, msg)
		return
	}
	lastId, err := exec1.LastInsertId()
	if err != nil || lastId == 0 {
		Tx.Rollback()
		code := 500
		msg = "写入划转流水失败"
		config.Errors(ctx, code, msg)
		return
	}

	err = Tx.Commit()
	if err != nil {
		Tx.Rollback()
		code := 500
		msg = "划转失败"
		config.Errors(ctx, code, msg)
		return
	}
	config.Success(ctx, http.StatusOK, "划转成功", nil)
	return
}

// GetFundAccountTransferList 获取用户资金账户划转记录列表
func GetFundAccountTransferList(ctx *gin.Context) {
	var msg any
	UserInfo = global.GetUserInfo(client)

	conn = inits.InitMysql()
	defer conn.Close()

	pageNum, _ := strconv.Atoi(ctx.DefaultQuery("pageNum", "1"))   //获取分页数量 pageNum
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "1")) //获取当前页 pageSize total

	//计算总数
	sqlStr := "select count(1) as total from bo_user_fund_account_transfer_bill where userId=?"
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

	sqlStr = "select id,userId,transferId,transferAmount,transferFundAccountType,transferUnit,collectId,collectAmount,collectFundAccountType,collectUnit,orderNo,exchangeRate,status,addTime from bo_user_fund_account_transfer_bill where userId=? limit ? offset ?"
	rows, err = conn.Query(sqlStr, UserInfo.Id, pageSize, (pageNum-1)*pageSize)
	if err != nil {
		code := 500
		msg = "系统内部错误"
		config.Errors(ctx, code, msg)
		return
	}
	defer rows.Close()

	var transferBillJsonMap []*models.UserFundAccountTransferBillJson
	for rows.Next() {
		var transferBill models.UserFundAccountTransferBill
		var transferBillJson models.UserFundAccountTransferBillJson
		rows.Scan(&transferBill.Id, &transferBill.UserId, &transferBill.TransferId, &transferBill.TransferAmount, &transferBill.TransferFundAccountType, &transferBill.TransferUnit, &transferBill.CollectId, &transferBill.CollectAmount, &transferBill.CollectFundAccountType, &transferBill.CollectUnit, &transferBill.OrderNo, &transferBill.ExchangeRate, &transferBill.Status, &transferBill.AddTime)
		transferBillJson = models.ConvUserFundAccountTransferBillNull2Json(transferBill, transferBillJson)
		transferBillJsonMap = append(transferBillJsonMap, &transferBillJson)
	}
	var fundAccountTransferList FundAccountTransferList
	fundAccountTransferList.List = transferBillJsonMap
	fundAccountTransferList.Total = Total
	fundAccountTransferList.PageNum = pageNum
	fundAccountTransferList.PageSize = pageSize
	config.Success(ctx, http.StatusOK, "请求成功", &fundAccountTransferList)
	return
}

// CountExchangeRates 计算汇率
func CountExchangeRates(ctx *gin.Context) {
	var ratesForm SystemExchangeRates
	var msg any

	UserInfo = global.GetUserInfo(client)
	//1、获取参数，并验证
	if errs := ctx.ShouldBind(&ratesForm); errs != nil {
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

	if len(ratesForm.TargetCode) == 0 || len(ratesForm.SourceCode) == 0 {
		code := 500
		msg = "提交参数错误"
		config.Errors(ctx, code, msg)
		return
	}

	conn = inits.InitMysql()
	defer conn.Close()

	//查询汇率
	sqlStr := "select exchangeRate from bo_system_exchange_rates where sourceCode=? and targetCode=? limit 1"
	err := conn.QueryRow(sqlStr, ratesForm.SourceCode, ratesForm.TargetCode).Scan(&ratesForm.ExchangeRate)
	if err != nil {
		logger.Warn(fmt.Sprintf("Err:[查询失败 failed.] Time:[%s] Err:[%v]", time.Now().Format(time.DateTime), err.Error()))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}

	ratesForm.CollectAmount = ratesForm.TransferAmount * ratesForm.ExchangeRate
	config.Success(ctx, http.StatusOK, "请求成功", &ratesForm)
	return
}
