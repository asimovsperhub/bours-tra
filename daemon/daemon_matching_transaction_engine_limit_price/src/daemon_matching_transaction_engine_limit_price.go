package main

import (
	"Bourse/app/models"
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

// ==================== 撮合限价后的下单消费队列脚本 ================

func init() {
	//初始化配置方式
	config.LoadConfig()
	inits.InitLog()

	file = "daemon_matching_transaction_engine_limit_price"

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
				//d.Nack(false, true)
				d.Ack(false)
				//logger.Warn(fmt.Sprintf("file:%s 设置消息为消费失败  Received a message:[%s] Time:[%s]", file, d.Body, time.Now().Format(time.DateTime)))
			} else {
				d.Ack(false)
			}
		}
	}()
	logger.Warn(fmt.Sprintf("【*】warting for messages, To exit press CCTRAL+C"))
	<-forever
}

func consumeSimple(body []byte) bool {
	// 写入持仓表---修改订单表状态---写入持仓缓存

	//redis初始化方式
	redisDb = inits.InitRedis()
	defer redisDb.Close()

	mysqlDb = inits.InitMysql()
	defer mysqlDb.Close()

	//解析消息
	msg := global.TradeOrderHoldingsMQMsg{}
	err := json.Unmarshal(body, &msg)
	if err != nil {
		return false
	}

	logger.Warn(fmt.Sprintf("file:%s  Failed:[init Translator failed] Time:[%s] Msg:[%v]", file, time.Now().Format(time.DateTime), string(body)))
	//校验是否已被消费
	//dataTradeOrderId, err := redisDb.HGet(context.Background(), msg.QueueHSetKey, fmt.Sprintf("%d", msg.TradeOrderId)).Result()
	//if err != nil || dataTradeOrderId == "" {
	//	logger.Warn(fmt.Sprintf("file:%s a111111 Time:[%s] dataTradeOrderId:[%s] err:[%s]", file, time.Now().Format(time.DateTime), dataTradeOrderId, err.Error()))
	//	return false
	//}

	Tx, err := mysqlDb.Begin()
	if err != nil {
		return false
	}
	//根据订单id查询订单信息
	sqlStr := "select id,userId,bourseType,systemBoursesId,symbol,stockSymbol,stockName,tradeType,isUpsDowns,orderType,entrustNum,entrustStatus,limitPrice,marketPrice,stopPrice,takeProfitPrice,orderAmount,orderAmountSum,earnestMoney,handlingFee,positionPrice,dealPrice,entrustOrderNumber,stockHoldingsId,entrustTime,updateTime from bo_user_trade_orders where id=?"
	rows, err := mysqlDb.Query(sqlStr, msg.TradeOrderId)
	if err != nil {
		logger.Warn(fmt.Sprintf("file:%s  a2222222 Time:[%s] Err:[%v]", file, time.Now().Format(time.DateTime), err))
		return false
	}
	defer rows.Close()
	var userTradeOrdersJson models.UserTradeOrdersJson
	for rows.Next() {
		var userTradeOrders models.UserTradeOrders
		rows.Scan(&userTradeOrders.Id, &userTradeOrders.UserId, &userTradeOrders.BourseType, &userTradeOrders.SystemBoursesId, &userTradeOrders.Symbol, &userTradeOrders.StockSymbol, &userTradeOrders.StockName, &userTradeOrders.TradeType, &userTradeOrders.IsUpsDowns, &userTradeOrders.OrderType, &userTradeOrders.EntrustNum, &userTradeOrders.EntrustStatus, &userTradeOrders.LimitPrice, &userTradeOrders.MarketPrice, &userTradeOrders.StopPrice, &userTradeOrders.TakeProfitPrice, &userTradeOrders.OrderAmount, &userTradeOrders.OrderAmountSum, &userTradeOrders.EarnestMoney, &userTradeOrders.HandlingFee, &userTradeOrders.PositionPrice, &userTradeOrders.DealPrice, &userTradeOrders.EntrustOrderNumber, &userTradeOrders.StockHoldingsId, &userTradeOrders.EntrustTime, &userTradeOrders.UpdateTime)
		userTradeOrdersJson = models.ConvUserTradeOrdersNull2Json(userTradeOrders, userTradeOrdersJson)
		break
	}

	//if userTradeOrdersJson.Id <= 0 || userTradeOrdersJson.EntrustStatus != 1 {
	//	//TODO 未处理的却变成已被处理 ...
	//	return false
	//}

	//if userTradeOrdersJson.TradeType == 2 { //卖出操作不在这里处理
	//	return false
	//}

	//写入持仓表
	addSql := fmt.Sprintf("insert into bo_user_stock_holdings (userId,bourseType,systemBoursesId,symbol,stockSymbol,stockName,isUpsDowns,orderType,stockHoldingNum,stockHoldingType,limitPrice,marketPrice,stopPrice,takeProfitPrice,orderAmount,orderAmountSum,earnestMoney,handlingFee,positionPrice,dealPrice,buyingPrice,entrustOrderNumber,openPositionTime) values (%d,%d,%d,'%s','%s','%s',%d,%d,%d,%d,%f,%f,%f,%f,%f,%f,%f,%f,%f,%f,%f,'%s','%s')", userTradeOrdersJson.UserId, userTradeOrdersJson.BourseType, userTradeOrdersJson.SystemBoursesId, userTradeOrdersJson.Symbol, userTradeOrdersJson.StockSymbol, userTradeOrdersJson.StockName, userTradeOrdersJson.IsUpsDowns, userTradeOrdersJson.OrderType, userTradeOrdersJson.EntrustNum, 1, userTradeOrdersJson.LimitPrice, userTradeOrdersJson.MarketPrice, userTradeOrdersJson.StopPrice, userTradeOrdersJson.TakeProfitPrice, userTradeOrdersJson.OrderAmount, userTradeOrdersJson.OrderAmountSum, userTradeOrdersJson.EarnestMoney, userTradeOrdersJson.HandlingFee, userTradeOrdersJson.PositionPrice, userTradeOrdersJson.DealPrice, msg.DealPrice, userTradeOrdersJson.EntrustOrderNumber, userTradeOrdersJson.EntrustTime)
	exec, err := Tx.Exec(addSql)
	if err != nil {
		logger.Warn(fmt.Sprintf("file:%s  a3333333 Time:[%s] Err:[%v]", file, time.Now().Format(time.DateTime), err))
		return false
	}
	lastInsertId, err := exec.LastInsertId()
	if err != nil || lastInsertId <= 0 {
		return false
	}

	//修改订单参数
	sqlStr = "update bo_user_trade_orders set entrustStatus=?,positionPrice=? where id=?"
	exec1, err := Tx.Exec(sqlStr, 2, msg.DealPrice, msg.TradeOrderId)
	if err != nil {
		logger.Warn(fmt.Sprintf("file:%s  a555555 Time:[%s] Err:[%v]", file, time.Now().Format(time.DateTime), err))
		Tx.Rollback()
		return false
	}
	rowsAffected2, err := exec1.RowsAffected()
	fmt.Println(rowsAffected2, msg.DealPrice, msg.TradeOrderId)
	if err != nil {
		Tx.Rollback()
		return false
	}

	err = Tx.Commit()
	if err != nil {
		return false
	}

	/*//判断是否有止损止盈设置(针对限价成交,市价成交在下单中直接设置)
	if userTradeOrdersJson.TradeType == 1 && userTradeOrdersJson.OrderType == 2 {
		//止损设置 (ZADD stock_order_stop_price_交易所id  价格 订单ID)
		tradeOrdersKey1 := fmt.Sprintf("stock_order_stop_price_%v", userTradeOrdersJson.SystemBoursesId)
		orderMoney1 := userTradeOrdersJson.StopPrice * global.PriceEnlargeMultiples
		redisDb.ZAdd(context.Background(), tradeOrdersKey1, redis.Z{Score: orderMoney1, Member: lastInsertId}).Err()

		//止赢设置 (ZADD stock_order_stop_price_交易所id  价格 订单ID)
		tradeOrdersKey2 := fmt.Sprintf("stock_order_take_profit_price_%v", userTradeOrdersJson.SystemBoursesId)
		orderMoney2 := userTradeOrdersJson.StopPrice * global.PriceEnlargeMultiples
		redisDb.ZAdd(context.Background(), tradeOrdersKey2, redis.Z{Score: orderMoney2, Member: lastInsertId}).Err()
	}*/

	//写入用户持仓 哈希缓存 redis的哈希key=user_stock_holdings_data_+ client.Uid + "_" + systemBoursesId ...
	sqlStr = "select id,userId,bourseType,systemBoursesId,symbol,stockSymbol,stockName,isUpsDowns,orderType,stockHoldingNum,stockHoldingType,limitPrice,marketPrice,stopPrice,takeProfitPrice,orderAmount,orderAmountSum,earnestMoney,handlingFee,positionPrice,dealPrice,buyingPrice,profitLossMoney,entrustOrderNumber,openPositionTime,updateTime from bo_user_stock_holdings where id=? and userId=? and stockHoldingType=?"
	rows, err = mysqlDb.Query(sqlStr, lastInsertId, userTradeOrdersJson.UserId, 1)
	if err != nil {
		logger.Warn(fmt.Sprintf("file:%s  a4444444 Time:[%s] Err:[%v]", file, time.Now().Format(time.DateTime), err))
		return false
	}
	defer rows.Close()
	var userStockJson models.UserStockHoldingsJson
	for rows.Next() {
		var userStock models.UserStockHoldings
		rows.Scan(&userStock.Id, &userStock.UserId, &userStock.BourseType, &userStock.SystemBoursesId, &userStock.Symbol, &userStock.StockSymbol, &userStock.StockName, &userStock.IsUpsDowns, &userStock.OrderType, &userStock.StockHoldingNum, &userStock.StockHoldingType, &userStock.LimitPrice, &userStock.MarketPrice, &userStock.StopPrice, &userStock.TakeProfitPrice, &userStock.OrderAmount, &userStock.OrderAmountSum, &userStock.EarnestMoney, &userStock.HandlingFee, &userStock.PositionPrice, &userStock.DealPrice, &userStock.BuyingPrice, &userStock.ProfitLossMoney, &userStock.EntrustOrderNumber, &userStock.OpenPositionTime, &userStock.UpdateTime)
		userStockJson = models.ConvUserStockHoldingsNull2Json(userStock, userStockJson)
		break
	}

	//获取用户持仓哈希缓存 redis的哈希key=user_stock_holdings_data_+ client.Uid + "_" +systemBoursesId ...
	cacheKey := fmt.Sprintf("user_stock_holdings_data_%d_%d", userTradeOrdersJson.UserId, userTradeOrdersJson.SystemBoursesId)
	marshal, _ := json.Marshal(userStockJson)
	err = redisDb.HSet(context.Background(), cacheKey, lastInsertId, marshal).Err()
	if err != nil {
		logger.Warn(fmt.Sprintf("file:%s 设置用户持仓哈希缓存失败 holdingsId:[%d]   a666666 Time:[%s] Err:[%v]", file, lastInsertId, time.Now().Format(time.DateTime), err))
		Tx.Rollback()
		return false
	}

	//删除队列持久化数据
	err = redisDb.HDel(context.Background(), msg.QueueHSetKey, fmt.Sprintf("%d", userTradeOrdersJson.Id)).Err()
	if err != nil {
		logger.Warn(fmt.Sprintf("file:%s 删除队列持久化缓存数据失败 userTradeOrdersJsonId:[%d]   a666666 Time:[%s] Err:[%v]", file, userTradeOrdersJson.Id, time.Now().Format(time.DateTime), err))
		Tx.Rollback()
		return false
	}

	return true
}

// 撮合限价下单处理队列 queueName=tradeOrderQueue_LimitPrice
func main() {
	tradeOrderConsumeSimpleMq(global.TradeorderqueueLimitprice)
}
