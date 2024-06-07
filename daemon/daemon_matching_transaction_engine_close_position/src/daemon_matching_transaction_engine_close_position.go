package main

import (
	"Bourse/app/models"
	"Bourse/app/servers"
	"Bourse/common/global"
	"Bourse/common/inits"
	"Bourse/common/translators"
	"Bourse/config"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"gitee.com/phper95/pkg/logger"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	mysqlDb *sql.DB
	redisDb *redis.Client

	file string
)

// ==================== 平仓、强平、止损止盈的消费队列脚本 ================

func init() {
	//初始化配置方式
	config.LoadConfig()
	inits.InitLog()

	file = "daemon_matching_transaction_engine_close_position"

	//多语言
	if err := translators.InitTranslators("zh"); err != nil {
		fmt.Sprintf("init Translator failed,err:%v\n", err)
		logger.Warn(fmt.Sprintf("file:%s  Failed:[init Translator failed] Time:[%s] Err:[%v]", file, time.Now().Format(time.DateTime), err))
	}
}

func tradeOrderConsumeSimpleMq(queueKey string) {
	//接收者
	r := inits.NewRabbitMQSimple(queueKey)

	//rabbitmq.ConsumeSimple()
	//1、申请队列，如果队列存在就跳过，不存在创建
	//优点：保证队列存在，消息能发送到队列中
	_, err := r.Channel.QueueDeclare(
		//队列名称
		r.QueueName,
		//是否持久化
		true,
		//是否为自动删除 当最后一个消费者断开连接之后，是否把消息从队列中删除
		false,
		//是否具有排他性
		false,
		//是否阻塞
		false,
		//额外数学系
		nil,
	)
	if err != nil {
		logger.Warn(fmt.Sprintf("file:%s Time:[%s] Err:[%v]", file, time.Now().Format(time.DateTime), err))
		fmt.Println(err)
	}
	//接收消息
	msgs, err := r.Channel.Consume(
		r.QueueName,
		//用来区分多个消费者
		"",
		//是否自动应答
		false,
		//是否具有排他性
		false,
		//如果设置为true,表示不能同一个connection中发送的消息传递给这个connection中的消费者
		false,
		//队列是否阻塞
		false,
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}
	forever := make(chan bool)

	//启用协程处理
	go func() {
		for d := range msgs {
			//处理的逻辑函数
			if ok := consumeSimple(d.Body); !ok {
				//设置消息为消费失败
				d.Nack(false, true)
				logger.Warn(fmt.Sprintf("file:%s 设置消息为消费失败  Received a message:[%s] Time:[%s]", file, d.Body, time.Now().Format(time.DateTime)))
			} else {
				d.Ack(false)
			}
		}
	}()
	logger.Warn(fmt.Sprintf("【*】warting for messages, To exit press CCTRAL+C"))
	<-forever
}

func consumeSimple(body []byte) bool {

	//redis初始化方式
	redisDb = inits.InitRedis()

	mysqlDb = inits.InitMysql()
	defer mysqlDb.Close()

	//解析消息
	msg := global.TradeOrderHoldingsMQMsg{}
	err := json.Unmarshal(body, &msg)
	if err != nil {
		return false
	}
	fmt.Printf(" ========1 %+v ==========", msg)
	//校验是否已被消费 HSet
	dataStockHoldingsId, err := redisDb.HGet(context.Background(), msg.QueueHSetKey, fmt.Sprintf("%d", msg.StockHoldingsId)).Result()
	if err != nil {
		logger.Warn(fmt.Sprintf("file:%s 7777777 Time:[%s] Err:[%v]", file, time.Now().Format(time.DateTime), err))
		return false
	}
	fmt.Printf(" ========2 %+v ==========", dataStockHoldingsId)
	Tx, err := mysqlDb.Begin()
	if err != nil {
		logger.Warn(fmt.Sprintf("file:%s 111111 Time:[%s] Err:[%v]", file, time.Now().Format(time.DateTime), err))
		return false
	}
	//根据持仓id查询持仓信息
	sqlStr := "select id,userId,bourseType,systemBoursesId,symbol,stockSymbol,stockName,isUpsDowns,orderType,stockHoldingNum,stockHoldingType,limitPrice,marketPrice,stopPrice,takeProfitPrice,orderAmount,orderAmountSum,earnestMoney,handlingFee,positionPrice,dealPrice,buyingPrice,profitLossMoney,entrustOrderNumber,openPositionTime,updateTime from bo_user_stock_holdings where id=?"
	rows, err := mysqlDb.Query(sqlStr, msg.StockHoldingsId)
	if err != nil {
		logger.Warn(fmt.Sprintf("file:%s 2222222 Time:[%s] Err:[%v]", file, time.Now().Format(time.DateTime), err))
		return false
	}
	defer rows.Close()
	var stockHoldingsJson models.UserStockHoldingsJson
	for rows.Next() {
		var userStockHoldings models.UserStockHoldings
		rows.Scan(&userStockHoldings.Id, &userStockHoldings.UserId, &userStockHoldings.BourseType, &userStockHoldings.SystemBoursesId, &userStockHoldings.Symbol, &userStockHoldings.StockSymbol, &userStockHoldings.StockName, &userStockHoldings.IsUpsDowns, &userStockHoldings.OrderType, &userStockHoldings.StockHoldingNum, &userStockHoldings.StockHoldingType, &userStockHoldings.LimitPrice, &userStockHoldings.MarketPrice, &userStockHoldings.StopPrice, &userStockHoldings.TakeProfitPrice, &userStockHoldings.OrderAmount, &userStockHoldings.OrderAmountSum, &userStockHoldings.EarnestMoney, &userStockHoldings.HandlingFee, &userStockHoldings.PositionPrice, &userStockHoldings.DealPrice, &userStockHoldings.BuyingPrice, &userStockHoldings.ProfitLossMoney, &userStockHoldings.EntrustOrderNumber, &userStockHoldings.OpenPositionTime, &userStockHoldings.UpdateTime)
		stockHoldingsJson = models.ConvUserStockHoldingsNull2Json(userStockHoldings, stockHoldingsJson)
		break
	}
	fmt.Printf(" ========3 %+v ==========", stockHoldingsJson)
	if stockHoldingsJson.Id <= 0 || stockHoldingsJson.StockHoldingType != 1 {
		//TODO 未处理的却变成已被处理 ...
		return false
	}

	//计算盈亏 (持仓价-成交价)*持仓数量  ===有保险设置，所有持仓价>成交价
	profitLossMoney := (stockHoldingsJson.PositionPrice - msg.DealPrice) * float64(stockHoldingsJson.StockHoldingNum)

	//结算调取资金账户server服务
	var userFundAccountsJson models.UserFundAccountsJson
	sqlStr = "select id,userId,fundAccount,fundAccountType,money,unit,addTime,updateTime from bo_user_fund_accounts where userId=? and fundAccountType=? limit 1"
	rows, err = mysqlDb.Query(sqlStr, stockHoldingsJson.UserId, stockHoldingsJson.SystemBoursesId)
	if err != nil {
		logger.Warn(fmt.Sprintf("file:%s 3333333 Time:[%s] Err:[%v]", file, time.Now().Format(time.DateTime), err))
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var userFundAccounts models.UserFundAccounts
		rows.Scan(&userFundAccounts.Id, &userFundAccounts.UserId, &userFundAccounts.FundAccount, &userFundAccounts.FundAccountType, &userFundAccounts.Money, &userFundAccounts.Unit, &userFundAccounts.AddTime, &userFundAccounts.UpdateTime)
		userFundAccountsJson = models.ConvUserFundAccountsNull2Json(userFundAccounts, userFundAccountsJson)
		break
	}

	//修改资金账户余额 & 修改Redis中用户账户资金余额(买入操作)
	/*userFundAccountServe := servers.UserFundAccountServe{
		Id:              userFundAccountsJson.Id,
		UserId:          userFundAccountsJson.UserId,
		FundAccount:     userFundAccountsJson.FundAccount,
		FundAccountType: userFundAccountsJson.FundAccountType,
		Operate:         1,
		OperateMoney:    profitLossMoney,
	}*/
	userFundAccountServe := servers.UserFundAccountServe{
		Id:              userFundAccountsJson.Id,
		UserId:          userFundAccountsJson.UserId,
		FundAccount:     userFundAccountsJson.FundAccount,
		FundAccountType: userFundAccountsJson.FundAccountType,
		Operate:         1,
		OperateMoney:    stockHoldingsJson.OrderAmountSum + profitLossMoney,
		OrderAmount:     stockHoldingsJson.OrderAmount,
		OrderAmountSum:  stockHoldingsJson.OrderAmountSum,
		EarnestMoney:    stockHoldingsJson.EarnestMoney,
		HandlingFee:     stockHoldingsJson.HandlingFee,
		OperationType:   1,
		OrderNo:         stockHoldingsJson.EntrustOrderNumber,
		Remark:          "",
	}
	service, err := servers.UserFundAccountService(&userFundAccountServe, mysqlDb)
	if err != nil || service == false {
		logger.Warn(fmt.Sprintf("file:%s 4444444 Time:[%s]", file, time.Now().Format(time.DateTime)))
		return false
	}

	//修改支持表状态
	sqlStr = "update bo_user_stock_holdings set stockHoldingType=?,dealPrice=?,profitLossMoney=?,updateTime=? where id=? and stockHoldingType=?"
	exec, err := Tx.Exec(sqlStr, 2, msg.DealPrice, profitLossMoney, time.Now().Format(time.DateTime), stockHoldingsJson.Id, 1)
	if err != nil {
		return false
	}
	affected, err := exec.RowsAffected()
	if err != nil || affected != 1 {
		logger.Warn(fmt.Sprintf("file:%s 555555 Time:[%s]", file, time.Now().Format(time.DateTime)))
		Tx.Rollback()
		return false
	}
	// 移除止损撮合集合
	/*tradeOrdersKey1 := fmt.Sprintf("stock_order_stop_price_%v", stockHoldingsJson.SystemBoursesId)
	redisDb.ZRem(context.Background(), tradeOrdersKey1, stockHoldingsJson.Id).Err()
	// 移除止盈撮合集合
	tradeOrdersKey2 := fmt.Sprintf("stock_order_take_profit_price_%v", stockHoldingsJson.SystemBoursesId)
	redisDb.ZRem(context.Background(), tradeOrdersKey2, stockHoldingsJson.Id).Err()*/

	//移除用户持仓列表
	cacheKey := fmt.Sprintf("user_stock_holdings_data_%d_%d", stockHoldingsJson.UserId, stockHoldingsJson.SystemBoursesId)
	redisDb.HDel(context.Background(), cacheKey, fmt.Sprintf("%d", stockHoldingsJson.Id))

	//删除队列持久化数据
	redisDb.HDel(context.Background(), msg.QueueHSetKey, fmt.Sprintf("%d", stockHoldingsJson.Id))

	err = Tx.Commit()
	if err != nil {
		logger.Warn(fmt.Sprintf("file:%s 666666 Time:[%s] Err:[%v]", file, time.Now().Format(time.DateTime), err))
		return false
	}
	return true
}

// 平仓、强平、止损止盈处理队列 queueName=tradeOrderQueue_CloseOut
func main() {
	tradeOrderConsumeSimpleMq(global.TradeorderqueueCloseout)
}
