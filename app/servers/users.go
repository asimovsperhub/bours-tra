package servers

import (
	"Bourse/app/models"
	"Bourse/common/inits"
	"Bourse/common/translators"
	"Bourse/config"
	"Bourse/config/constant"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	conn   *sql.DB
	client *redis.Client

	UserInfo *constant.UserInfoStruct
)

const TableUserFundAccounts = "bo_user_fund_accounts"

// UserFundAccountServe UserFundAccount 资金账户
type UserFundAccountServe struct {
	Id              int64   `json:"id"`              //资金账户id
	UserId          int64   `json:"userId"`          //用户id
	FundAccount     string  `json:"fundAccount"`     //资金账号
	FundAccountType int32   `json:"fundAccountType"` //资金账号类型：1 现货资产 2 合约资产 3 马股资产 4 美股资产 5 日股资产 6 佣金账户
	Operate         int32   `json:"operate"`         //操作：1增 2减
	OperateMoney    float64 `json:"operateMoney"`    //操作金额
	OrderAmount     float64 `json:"orderAmount"`     //订单金额
	OrderAmountSum  float64 `json:"orderAmountSum"`  //订单总金额(包含保证金或手续费)
	EarnestMoney    float64 `json:"earnestMoney"`    //保证金
	HandlingFee     float64 `json:"handlingFee"`     //手续费
	OperationType   int32   `json:"operationType"`   //操作类型: 1 购股 2 购股保证金 3 购股手续费 4 持仓结算 5 充值 6 提款 7 团队分佣 8 持仓结算手续费 9 平台操作
	OrderNo         string  `json:"orderNo"`         //订单号
	Remark          string  `json:"remark"`          //备注
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

// UserFundAccountService 用户资金账户金额变动服务 TODO 账户流水写入 ...
func UserFundAccountService(userFundAccountServe *UserFundAccountServe, conn *sql.DB) (bool, error) {
	//UserInfo = global.GetUserInfo(client)

	if userFundAccountServe.OperateMoney < 0 {
		return false, errors.New("操作失败，金额不对")
	}

	coonTx, err := conn.Begin()
	if err != nil {
		return false, errors.New(fmt.Sprintf("开启事务失败,Err:[%s]", err.Error()))
	}
	sqlStr := "select id,userId,fundAccountType,fundAccount,money,unit,addTime,updateTime from " + TableUserFundAccounts + " where id=? and userId=?"
	//TODO conn.Query 使用coonTx事务操作时会报错。。。
	Rows, err := conn.Query(sqlStr, userFundAccountServe.Id, userFundAccountServe.UserId)
	if err != nil {
		coonTx.Rollback()
		return false, errors.New(fmt.Sprintf("查询失败,Err:[%s]", err.Error()))
	}
	//defer Rows.Close()
	var userFundAccountJson models.UserFundAccountsJson
	for Rows.Next() {
		var userFundAccounts models.UserFundAccounts
		Rows.Scan(&userFundAccounts.Id, &userFundAccounts.UserId, &userFundAccounts.FundAccountType, &userFundAccounts.FundAccount, &userFundAccounts.Money, &userFundAccounts.Unit, &userFundAccounts.AddTime, &userFundAccounts.UpdateTime)
		userFundAccountJson = models.ConvUserFundAccountsNull2Json(userFundAccounts, userFundAccountJson)
		break
	}
	if err = Rows.Err(); err != nil || userFundAccountJson.Id == 0 {
		Rows.Close()
		coonTx.Rollback()
		return false, errors.New(fmt.Sprintf("查询资金账户失败 userFundAccountJson:[%d],Err:[%s]", userFundAccountJson.Id, err.Error()))
	}

	fundAccountRedisKey := fmt.Sprintf("%s%v_%v", constant.UserFundAccountsKey, userFundAccountJson.FundAccountType, userFundAccountJson.UserId)

	var money float64
	if userFundAccountServe.OperationType == 1 {
		userFundAccountServe.OperateMoney = userFundAccountServe.OrderAmount + userFundAccountServe.EarnestMoney + userFundAccountServe.HandlingFee
	}
	if userFundAccountServe.Operate == 1 {
		money = userFundAccountJson.Money + userFundAccountServe.OperateMoney
	} else if userFundAccountServe.Operate == 2 {
		if userFundAccountJson.Money < userFundAccountServe.OperateMoney {
			fmt.Println("账户余额不足,操作失败")
			return false, errors.New("账户余额不足,操作失败")
		}
		money = userFundAccountJson.Money - userFundAccountServe.OperateMoney
	}

	sqlStr = fmt.Sprintf("update bo_user_fund_accounts set money=%f where id=%v and userId=%v", money, userFundAccountJson.Id, userFundAccountJson.UserId)
	exec, err := coonTx.Exec(sqlStr)
	if err != nil {
		coonTx.Rollback()
		return false, errors.New(fmt.Sprintf("修改操作失败,Err:[%s] SQL:[%s]", err.Error(), sqlStr))
	}
	affected, err := exec.RowsAffected()
	if err != nil || affected != 1 {
		coonTx.Rollback()
		return false, errors.New(fmt.Sprintf("RowsAffected查询失败,Err:[%s]", err.Error()))
	}

	//写入资金账户账单流水
	//写入划转流水表
	/*sqlStr = ""
	if userFundAccountServe.OrderAmount > 0 {
		sqlStr = fmt.Sprintf("insert into bo_user_fund_account_flows (userId,fundAccountType,operationType,changeType,changeAmount,beforeAmount,afterAmount,unit,orderNo,remark,addTime) VALUES (%d,%d,%d,%d,%f,%f,%f,'%s','%s','%s','%s')", UserInfo.Id, userFundAccountJson.FundAccountType, userFundAccountServe.OperationType, userFundAccountServe.Operate, userFundAccountServe.OperateMoney, userFundAccountJson.Money, money, userFundAccountJson.Unit, userFundAccountServe.OrderNo, "", time.Now().Format(time.DateTime))*/
	/*exec1, err := coonTx.Exec(sqlStr, UserInfo.Id, userFundAccountJson.FundAccountType, userFundAccountServe.OperationType, userFundAccountServe.Operate, userFundAccountServe.OperateMoney, userFundAccountJson.Money, money, userFundAccountJson.Unit, userFundAccountServe.OrderNo, "", time.Now().Format(time.DateTime))*/

	/*} else if userFundAccountServe.EarnestMoney > 0 {
		sqlStr = sqlStr + fmt.Sprintf("(%d,%d,%d,%d,%f,%f,%f,'%s','%s','%s','%s')", UserInfo.Id, userFundAccountJson.FundAccountType, userFundAccountServe.OperationType, userFundAccountServe.Operate, userFundAccountServe.OperateMoney, userFundAccountJson.Money, money, userFundAccountJson.Unit, userFundAccountServe.OrderNo, "", time.Now().Format(time.DateTime))
	} else if userFundAccountServe.HandlingFee > 0 {

	}*/

	sqlStr = fmt.Sprintf("insert into bo_user_fund_account_flows (userId,fundAccountType,operationType,changeType,changeAmount,beforeAmount,afterAmount,unit,orderNo,remark,addTime) VALUES (%d,%d,%d,%d,%f,%f,%f,'%s','%s','%s','%s')", userFundAccountJson.UserId, userFundAccountJson.FundAccountType, userFundAccountServe.OperationType, userFundAccountServe.Operate, userFundAccountServe.OperateMoney, userFundAccountJson.Money, money, userFundAccountJson.Unit, userFundAccountServe.OrderNo, "", time.Now().Format(time.DateTime))
	exec1, err := coonTx.Exec(sqlStr)
	if err != nil {
		coonTx.Rollback()
		return false, errors.New(fmt.Sprintf("写入资金账户账单流水失败,Err:[%s]", err.Error()))
	}
	lastId, err := exec1.LastInsertId()
	if err != nil || lastId == 0 {
		coonTx.Rollback()
		return false, errors.New(fmt.Sprintf("写入资金账户账单流水失败,Err:[%s]", err.Error()))
	}

	//操作Redis
	clientPipe := client.TxPipeline()
	clientPipe.Set(context.Background(), fundAccountRedisKey, money, 0)
	_, err = clientPipe.Exec(context.Background())
	if err != nil {
		return false, errors.New(fmt.Sprintf("redis操作失败,Err:[%s]", err.Error()))
	}

	if lastId == 0 {
		coonTx.Rollback()
		return false, errors.New(fmt.Sprintf("服务操作失败,Err:[%s]", err.Error()))

	}
	coonTx.Commit()

	return true, nil
}
