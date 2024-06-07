package socket

import (
	"Bourse/app/models"
	"Bourse/app/servers"
	"Bourse/common/inits"
	"Bourse/common/ordernum"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"gitee.com/phper95/pkg/logger"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Client 信息
type Client struct {
	Uid    string          //用户uid
	UserId int64           //用户表id
	Conn   *websocket.Conn //连接
	Send   chan *Message   //发送
	mux    sync.Mutex      //互斥锁
	rux    sync.RWMutex    //读写互斥锁
}

// Success ws响应成功结构体
type Success struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

// Errors ws响应错误结构体
type Errors struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}

type UserUccountBalanceList struct {
	UserFundAccountsSpotsJson     *models.UserFundAccountsJson       `json:"userSpotsJson"`
	UserFundAccountsSwapsJson     *models.UserFundAccountsJson       `json:"userSwapsJson"`
	UserFundAccountsMalaysiaJson  *models.UserFundAccountsJson       `json:"userMalaysiaJson"`
	UserFundAccountsUSJson        *models.UserFundAccountsJson       `json:"userUSJson"`
	UserFundAccountsYenJson       *models.UserFundAccountsJson       `json:"userYenJson"`
	UserFundAccountsBrokerageJson *models.UserFundAccountsJson       `json:"userBrokerageJson"`
	SpotsJsonMap                  []*models.UserFundAccountSpotsJson `json:"spotsJsonMap"`
}

// UserTradeOrdersForm 交易订单表-add入参
type UserTradeOrdersForm struct {
	BourseType         int32   `json:"bourseType" form:"bourseType" binding:"required"`      //所属类型：1 现货 2 合约 3 马股 4 美股 5 日股
	SystemBoursesId    int64   `json:"systemBoursesId" form:"systemBoursesId"`               //交易所股种id
	Symbol             string  `json:"symbol" form:"symbol"`                                 //交易代码
	StockSymbol        string  `json:"stockSymbol" form:"stockSymbol"`                       //股票代码
	StockName          string  `json:"stockName" form:"stockName"`                           //股票名字
	TradeType          int32   `json:"tradeType" form:"tradeType" binding:"required"`        //操作类型：1买入 2 卖出
	IsUpsDowns         int32   `json:"isUpsDowns" form:"isUpsDowns" binding:"required"`      //交易方向：1 买涨 2 买跌
	OrderType          int32   `json:"orderType" form:"orderType"`                           //交易方式：1 限价(需走撮合)  2 市价(下单直接成交(买涨))
	EntrustNum         int32   `json:"entrustNum" form:"entrustNum"`                         //委托数量
	LimitPrice         float64 `json:"limitPrice" form:"limitPrice"`                         //限价
	MarketPrice        float64 `json:"marketPrice" form:"marketPrice"`                       //市价
	StopPrice          float64 `json:"stopPrice" form:"stopPrice"`                           //止损价
	TakeProfitPrice    float64 `json:"takeProfitPrice" form:"takeProfitPrice"`               //止盈价
	OrderAmount        float64 `json:"orderAmount" form:"orderAmount" binding:"required"`    //订单金额
	OrderAmountSum     float64 `json:"orderAmountSum" form:"orderAmount" binding:"required"` //订单总金额
	EarnestMoney       float64 `json:"earnestMoney" form:"earnestMoney"`                     //保证金
	HandlingFee        float64 `json:"handlingFee" form:"handlingFee"`                       //手续费
	PositionPrice      float64 `json:"positionPrice" form:"positionPrice"`                   //持仓价
	ClosingPrice       float64 `json:"closingPrice" form:"closingPrice"`                     //平仓价
	BuyingPrice        float64 `json:"buyingPrice" form:"buyingPrice"`                       //买入价
	StockHoldingsId    int64   `json:"stockHoldingsId" form:"stockHoldingsId"`               //持仓id
	EntrustOrderNumber string  `json:"entrustOrderNumber" form:"entrustOrderNumber"`         //委托订单号
}

// UserTradeOrdersPendingWithdrawalClearing 交易订单表-list出参
type UserTradeOrdersPendingWithdrawalClearing struct {
	Pending    []*models.UserTradeOrdersJson `json:"pending"`
	Withdrawal []*models.UserTradeOrdersJson `json:"withdrawal"`
	Clearing   []*models.UserTradeOrdersJson `json:"clearing"`
}

// 写
func (c *Client) Write() {
	defer func() {
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				return
			}
			messageJson, _ := json.Marshal(message)
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(messageJson)
			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

//func (c *Client) Write2(msg *Message) {
//	time.Sleep(2 * time.Second) //延时2秒
//	c.mux.Lock()
//	messageJson, _ := json.Marshal(msg)
//	c.Conn.WriteMessage(websocket.TextMessage, messageJson)
//	c.mux.Unlock()
//	return
//}

// 读
func (c *Client) Read() {
	defer func() {
		MsServer.Unregister <- c
		c.Conn.Close()
	}()

	for {
		c.Conn.PongHandler()
		_, message, err := c.Conn.ReadMessage() //读取消息
		if err != nil {                         //读取消息失败，注销用户client
			MsServer.Unregister <- c
			c.Conn.Close()
			break
		}
		//解析消息
		msg := &SymbolMessage{}
		json.Unmarshal(message, msg)

		//处理消息类型
		switch msg.Type {
		case "ping": //接收ping回复pong
			c.Send <- &Message{
				Content: "pong",
			}
			break
		case "get-earnest-money":
			//根据股市ID获取保证金
			c.GetEarnestMoney(msg)
			break
		case "post-place-order":
			//下单
			c.PostPlaceOrder(msg)
			break
		case "post-withdraw-order":
			//撤单
			c.PostWithdrawOrder(msg)
			break
		case "post-close-position-order":
			//平仓
			c.PostClosePositionOrder(msg)
			break
		case "post-stop-profit-stop-loss-order": //止盈止损
			break
		case "get-user-account-balance":
			//获取对应资金账户余额
			c.GetUserUccountBalance(msg)
			break
		case "get-user-stock-holdings-num":
			//根据id获取持股数量
			c.GetUserStockHoldingsNum(msg)
			break
		case "get-list-user-trade-position":
			//获取用户交易持仓列表
			c.GetListUserTradePosition(msg)
			break
		case "get-list-user-trade-pending-withdrawal-clearing":
			//获取用户交易撤单-挂单-结算列表
			c.GetListUserTradePendingWithdrawalClearing(msg)
			break
		default:
			break
		}
	}
}

func (c *Client) GetEarnestMoney(msg *SymbolMessage) {
	//mysql初始化方式
	mysqlDb := inits.InitMysql()
	defer mysqlDb.Close() //关闭数据库连接
	orderAmount, _ := strconv.ParseFloat(msg.Content["orderAmount"], 64)

	//查询所有股市保证金
	sqlStr := "select id,bourseName,bourseType,earnestMoney,handlingFee,sort,status from bo_system_bourses where id=? and status=1 limit 1"
	rows, err := mysqlDb.Query(sqlStr, msg.Content["systemBoursesId"])
	if err != nil {
		c.Send <- &Message{
			ServersId: "get-earnest-money",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  "网络繁忙",
			},
			Uid: c.Uid,
		}
		return
	}
	defer rows.Close()

	var systemBoursesJson models.SystemBoursesJson
	for rows.Next() {
		var systemBourses models.SystemBourses
		rows.Scan(&systemBourses.Id, &systemBourses.BourseName, &systemBourses.BourseType, &systemBourses.EarnestMoney, &systemBourses.Sort, &systemBourses.Status)
		systemBoursesJson = models.ConvSystemBoursesNull2Json(systemBourses, systemBoursesJson)
		systemBoursesJson.HandlingFee = systemBoursesJson.HandlingFee * orderAmount
		break
	}
	earnestMoney := make(map[string]float64, 1)
	earnestMoney["earnestMoney"] = systemBoursesJson.EarnestMoney
	earnestMoney["handlingFee"] = systemBoursesJson.HandlingFee

	//查询计算保证金
	c.Send <- &Message{
		ServersId: "get-earnest-money",
		Content: &Success{
			Code: http.StatusOK,
			Msg:  "请求成功",
			Data: earnestMoney,
		},
		Uid: c.Uid,
	}
	return
}

func (c *Client) PostPlaceOrder(msg *SymbolMessage) {
	//-----下单--数字币 & 股票交易下单(买入卖出)===获取表单数据,并验证参数
	//redis初始化方式
	//redisDb := inits.InitRedis()

	//mysql初始化方式
	mysqlDb := inits.InitMysql()
	defer mysqlDb.Close() //关闭数据库连接

	// =========== RabbitMQ Simple模式 发送者 ===========
	//queueKey := "tradeOrderBuy_" //现货 spots 合约 contract
	//rabbitmq = inits.NewRabbitMQSimple("tradeOrderBuy_")

	//1、获取参数，并验证
	var userTradeOrdersForm UserTradeOrdersForm
	userTradeOrdersForm = ConvUserTradeOrdersForm2Msg(userTradeOrdersForm, msg)
	msg.Content["entrustOrderNumber"] = ordernum.GetOrderNo()

	if userTradeOrdersForm.EntrustNum%100 != 0 {
		c.Send <- &Message{
			ServersId: "post-place-order",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  "买入数量参数错误",
			},
			Uid: c.Uid,
		}
		return
	}

	//2、查询检验股票代码相关信息

	//3、获取根据交易类型用户资金账户信息-Redis查查不到-MySQL查
	/*userFundAccountsKey := fmt.Sprintf("%s%v_%v", constant.UserFundAccountsKey, userTradeOrdersForm.BourseType, c.UserId)
	result, err := redisDb.Get(context.Background(), userFundAccountsKey).Result()*/
	var userFundAccountsJson models.UserFundAccountsJson
	sqlStr := "select id,userId,fundAccount,fundAccountType,money,unit,addTime,updateTime from bo_user_fund_accounts where userId=? and fundAccountType=? limit 1"
	rows, err := mysqlDb.Query(sqlStr, c.UserId, userTradeOrdersForm.SystemBoursesId)
	if err != nil {
		c.Send <- &Message{
			ServersId: "post-place-order",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  fmt.Sprintf("网络繁忙 %s", err.Error()),
			},
			Uid: c.Uid,
		}
		return
	}
	defer rows.Close()
	for rows.Next() {
		var userFundAccounts models.UserFundAccounts
		rows.Scan(&userFundAccounts.Id, &userFundAccounts.UserId, &userFundAccounts.FundAccount, &userFundAccounts.FundAccountType, &userFundAccounts.Money, &userFundAccounts.Unit, &userFundAccounts.AddTime, &userFundAccounts.UpdateTime)
		userFundAccountsJson = models.ConvUserFundAccountsNull2Json(userFundAccounts, userFundAccountsJson)
		break
	}

	//4、判断用户账户余额是否充足
	if userFundAccountsJson.Money < userTradeOrdersForm.OrderAmountSum {
		logger.Warn(fmt.Sprintf("Time:[%s] Err:[用户账户资金不足] UserId:[%d] Money:[%f] OrderAmount:[%f]", time.Now().Format(time.DateTime), c.UserId, userFundAccountsJson.Money, userTradeOrdersForm.OrderAmount))
		c.Send <- &Message{
			ServersId: "post-place-order",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  "用户账户资金不足",
			},
			Uid: c.Uid,
		}
		return
	}

	//5、开启事务--MySQL-修改用户资金账户余额--修改Redis中用户账户资金余额--(写入MQ队列)写入资金账户交易订单表--写入交易订单表
	//5.1 修改资金账户余额&修改Redis中用户账户资金余额(买入操作)
	if userTradeOrdersForm.TradeType == 1 {
		userFundAccountServe := servers.UserFundAccountServe{
			Id:              userFundAccountsJson.Id,
			UserId:          userFundAccountsJson.UserId,
			FundAccount:     userFundAccountsJson.FundAccount,
			FundAccountType: userFundAccountsJson.FundAccountType,
			Operate:         2,
			OperateMoney:    userTradeOrdersForm.OrderAmountSum,
			OrderAmount:     userTradeOrdersForm.OrderAmount,
			OrderAmountSum:  userTradeOrdersForm.OrderAmountSum,
			EarnestMoney:    userTradeOrdersForm.EarnestMoney,
			HandlingFee:     userTradeOrdersForm.HandlingFee,
			OperationType:   2,
			OrderNo:         msg.Content["entrustOrderNumber"],
			Remark:          "",
		}
		service, err := servers.UserFundAccountService(&userFundAccountServe, mysqlDb)
		if err != nil || service == false {
			logger.Warn(fmt.Sprintf("Time:[%s] Err:[下单失败. %v]", time.Now().Format(time.DateTime), err.Error()))
			c.Send <- &Message{
				ServersId: "post-place-order",
				Content: &Errors{
					Code: http.StatusInternalServerError,
					Msg:  "下单失败",
				},
				Uid: c.Uid,
			}
			return
		}
	}

	//5.2 (写入MQ队列)写入资金账户交易订单表--写入交易订单表
	Tx, err := mysqlDb.Begin()
	if err != nil {
		logger.Warn(fmt.Sprintf("开启事务失败 Err:[%s] Time:[%s]", err.Error(), time.Now().Format(time.DateTime)))
		c.Send <- &Message{
			ServersId: "post-place-order",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  "网络繁忙1",
			},
			Uid: c.Uid,
		}
		return
	}
	entrustStatus := 1
	var lastHoldingsId int64
	if userTradeOrdersForm.TradeType == 1 { // 买入--买入市价
		if userTradeOrdersForm.OrderType == 2 { //市价
			entrustStatus = 2
			// 写入持仓表
			sqlStr = GetInsertUserStockHoldingsSql(userTradeOrdersForm, c.UserId, msg.Content["entrustOrderNumber"])
			exec, err := Tx.Exec(sqlStr)
			if err != nil {
				Tx.Rollback()
				logger.Warn(fmt.Sprintf("插入失败 SQL:[%s] Err:[%s] Time:[%s]", sqlStr, err.Error(), time.Now().Format(time.DateTime)))
				c.Send <- &Message{
					ServersId: "post-place-order",
					Content: &Errors{
						Code: http.StatusInternalServerError,
						Msg:  "网络繁忙2",
					},
					Uid: c.Uid,
				}
				return
			}
			lastHoldingsId, err = exec.LastInsertId()
			if err != nil || lastHoldingsId <= 0 {
				Tx.Rollback()
				c.Send <- &Message{
					ServersId: "post-place-order",
					Content: &Errors{
						Code: http.StatusInternalServerError,
						Msg:  "网络繁忙3",
					},
					Uid: c.Uid,
				}
				return
			}
			msg.Content["stockHoldingsId"] = fmt.Sprintf("%d", lastHoldingsId)
		}

		msg.Content["marketPrice"] = fmt.Sprintf("%f", userTradeOrdersForm.MarketPrice)

		//写入用户交易订单表
		sqlStr = models.GetWssTradeOrdersSql(msg.Content, c.UserId, entrustStatus)
		exec, err := Tx.Exec(sqlStr)
		if err != nil {
			Tx.Rollback()
			logger.Warn(fmt.Sprintf("插入失败 SQL:[%s] Err:[%s] Time:[%s]", sqlStr, err.Error(), time.Now().Format(time.DateTime)))
			c.Send <- &Message{
				ServersId: "post-place-order",
				Content: &Errors{
					Code: http.StatusInternalServerError,
					Msg:  "网络繁忙6",
				},
				Uid: c.Uid,
			}
			return
		}
		lastInsertId, err := exec.LastInsertId()
		if err != nil || lastInsertId <= 0 {
			Tx.Rollback()
			c.Send <- &Message{
				ServersId: "post-place-order",
				Content: &Errors{
					Code: http.StatusInternalServerError,
					Msg:  "网络繁忙7",
				},
				Uid: c.Uid,
			}
			return
		}

		//判断是否限价成交--写入限价成交集合
		/*if userTradeOrdersForm.TradeType == 1 && userTradeOrdersForm.OrderType == 1 {
			//  (ZADD stock_order_limit_price_交易所id  价格 订单ID)
			tradeOrdersKey := fmt.Sprintf("stock_order_limit_price_%v", userTradeOrdersForm.SystemBoursesId)
			orderMoney := userTradeOrdersForm.LimitPrice * global.PriceEnlargeMultiples
			err = redisDb.ZAdd(context.Background(), tradeOrdersKey, redis.Z{Score: orderMoney, Member: lastInsertId}).Err()
			if err != nil {
				fmt.Printf("ZAdd-3 失败 %s", err.Error())
				return
			}
		}*/

		//判断是否有止损止盈设置(针对市价成交,限价成交在撮合脚本中设置)
		/*if userTradeOrdersForm.TradeType == 1 && userTradeOrdersForm.OrderType == 2 {
			//  止损设置 (ZADD stock_order_stop_price_交易所id  价格 订单ID)
			tradeOrdersKey1 := fmt.Sprintf("stock_order_stop_price_%v", userTradeOrdersForm.SystemBoursesId)
			orderMoney1 := userTradeOrdersForm.StopPrice * global.PriceEnlargeMultiples
			err := redisDb.ZAdd(context.Background(), tradeOrdersKey1, redis.Z{Score: orderMoney1, Member: lastInsertId}).Err()
			if err != nil {
				fmt.Printf("ZAdd-1 失败 %s", err.Error())
				return
			}

			//  止赢设置 (ZADD stock_order_stop_price_交易所id  价格 订单ID)
			tradeOrdersKey2 := fmt.Sprintf("stock_order_take_profit_price_%v", userTradeOrdersForm.SystemBoursesId)
			orderMoney2 := userTradeOrdersForm.TakeProfitPrice * global.PriceEnlargeMultiples
			err = redisDb.ZAdd(context.Background(), tradeOrdersKey2, redis.Z{Score: orderMoney2, Member: lastInsertId}).Err()
			if err != nil {
				fmt.Printf("ZAdd-2 失败 %s", err.Error())
				return
			}
		}*/

		Tx.Commit()

		//限价和限价止损止盈写入缓存 TODO 后期改脚本写入 ...
		sqlStr = "select id,userId,bourseType,systemBoursesId,symbol,stockSymbol,stockName,tradeType,isUpsDowns,orderType,entrustNum,entrustStatus,limitPrice,marketPrice,stopPrice,takeProfitPrice,orderAmount,orderAmountSum,earnestMoney,handlingFee,positionPrice,dealPrice,entrustOrderNumber,stockHoldingsId,entrustTime,updateTime from bo_user_trade_orders where id=? and userId=? and entrustStatus=?"
		rows, err = mysqlDb.Query(sqlStr, lastInsertId, c.UserId, 1)
		if err != nil {
			logger.Warn(fmt.Sprintf("写入缓存，查写入限价数据失败 Err:[%s] Time:[%s]", err.Error(), time.Now().Format(time.DateTime)))
			c.Send <- &Message{
				ServersId: "post-place-order",
				Content: &Errors{
					Code: http.StatusInternalServerError,
					Msg:  "网络繁忙",
				},
				Uid: c.Uid,
			}
			return
		}
		defer rows.Close()
		var userTradeOrdersJson models.UserTradeOrdersJson
		for rows.Next() {
			var userTradeOrders models.UserTradeOrders
			rows.Scan(&userTradeOrders.Id, &userTradeOrders.UserId, &userTradeOrders.BourseType, &userTradeOrders.SystemBoursesId, &userTradeOrders.Symbol, &userTradeOrders.StockSymbol, &userTradeOrders.StockName, &userTradeOrders.TradeType, &userTradeOrders.IsUpsDowns, &userTradeOrders.OrderType, &userTradeOrders.EntrustNum, &userTradeOrders.EntrustStatus, &userTradeOrders.LimitPrice, &userTradeOrders.MarketPrice, &userTradeOrders.StopPrice, &userTradeOrders.TakeProfitPrice, &userTradeOrders.OrderAmount, &userTradeOrders.OrderAmountSum, &userTradeOrders.EarnestMoney, &userTradeOrders.HandlingFee, &userTradeOrders.PositionPrice, &userTradeOrders.DealPrice, &userTradeOrders.EntrustOrderNumber, &userTradeOrders.StockHoldingsId, &userTradeOrders.EntrustTime, &userTradeOrders.UpdateTime)
			userTradeOrdersJson = models.ConvUserTradeOrdersNull2Json(userTradeOrders, userTradeOrdersJson)
			break
		}
		//获取用户限价订单哈希缓存 redis的哈希key=user_stock_trade_orders_data_+ client.UserId + "_" +systemBoursesId ...
		cacheOrdersKey := fmt.Sprintf("user_stock_trade_orders_data_%d_%d", c.UserId, userTradeOrdersForm.SystemBoursesId)
		marshal1, _ := json.Marshal(userTradeOrdersJson)
		redisDb.HSet(context.Background(), cacheOrdersKey, userTradeOrdersJson.Id, marshal1).Err()

		//写入用户持仓缓存 TODO redis的哈希key=user_stock_holdings_data_+ client.UserId + "_" + systemBoursesId ...
		sqlStr = "select id,userId,bourseType,systemBoursesId,symbol,stockSymbol,stockName,isUpsDowns,orderType,stockHoldingNum,stockHoldingType,limitPrice,marketPrice,stopPrice,takeProfitPrice,orderAmount,orderAmountSum,earnestMoney,handlingFee,positionPrice,dealPrice,buyingPrice,profitLossMoney,entrustOrderNumber,openPositionTime,updateTime from bo_user_stock_holdings where id=? and userId=? and stockHoldingType=?"
		rows, err = mysqlDb.Query(sqlStr, lastHoldingsId, c.UserId, 1)
		if err != nil {
			logger.Warn(fmt.Sprintf("写入缓存，查询写入的持仓数据失败 Err:[%s] Time:[%s]", err.Error(), time.Now().Format(time.DateTime)))
			c.Send <- &Message{
				ServersId: "post-place-order",
				Content: &Errors{
					Code: http.StatusInternalServerError,
					Msg:  "网络繁忙4",
				},
				Uid: c.Uid,
			}
			return
		}
		defer rows.Close()
		var userStockJson models.UserStockHoldingsJson
		for rows.Next() {
			var userStock models.UserStockHoldings
			rows.Scan(&userStock.Id, &userStock.UserId, &userStock.BourseType, &userStock.SystemBoursesId, &userStock.Symbol, &userStock.StockSymbol, &userStock.StockName, &userStock.IsUpsDowns, &userStock.OrderType, &userStock.StockHoldingNum, &userStock.StockHoldingType, &userStock.LimitPrice, &userStock.MarketPrice, &userStock.StopPrice, &userStock.TakeProfitPrice, &userStock.OrderAmount, &userStock.OrderAmountSum, &userStock.EarnestMoney, &userStock.HandlingFee, &userStock.PositionPrice, &userStock.DealPrice, &userStock.BuyingPrice, &userStock.ProfitLossMoney, &userStock.EntrustOrderNumber, &userStock.OpenPositionTime, &userStock.UpdateTime)
			userStockJson = models.ConvUserStockHoldingsNull2Json(userStock, userStockJson)
			break
		}

		//获取用户持写入仓哈希缓存 redis的哈希key=user_stock_holdings_data_+ client.UserId + "_" +systemBoursesId ...
		cacheKey := fmt.Sprintf("user_stock_holdings_data_%d_%d", c.UserId, userTradeOrdersForm.SystemBoursesId)
		marshal, _ := json.Marshal(userStockJson)
		err = redisDb.HSet(context.Background(), cacheKey, userStockJson.Id, marshal).Err()
		if err != nil {
			logger.Warn(fmt.Sprintf("写入缓存，写入限价订单缓存数据失败 Err:[%s] Time:[%s]", err.Error(), time.Now().Format(time.DateTime)))
			return
		}

	} else if userTradeOrdersForm.TradeType == 2 { //卖出操作不走这个接口
		c.Send <- &Message{
			ServersId: "post-place-order",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  "非法操作5",
			},
			Uid: c.Uid,
		}
		return
	}

	c.Send <- &Message{
		ServersId: "post-place-order",
		Content: &Success{
			Code: http.StatusOK,
			Msg:  "请求成功",
			Data: nil,
		},
		Uid: c.Uid,
	}
	//发送订阅通知 重新订阅行情
	redisDb.Publish(context.Background(), "stockSymbol_subscribe_", userTradeOrdersForm.StockSymbol)
	return
}

func (c *Client) PostWithdrawOrder(msg *SymbolMessage) {
	//redis初始化方式 ----- 撤单（买入--限价--未成交的订单）------
	//redisDb := inits.InitRedis()

	//mysql初始化方式
	mysqlDb := inits.InitMysql()
	defer mysqlDb.Close() //关闭数据库连接

	//参数赋值 stockSymbol
	tradeOrdersId, _ := strconv.ParseInt(msg.Content["tradeOrdersId"], 10, 64)
	bourseTypeInt, _ := strconv.Atoi(msg.Content["bourseType"])
	bourseType := int32(bourseTypeInt)
	systemBoursesId, _ := strconv.ParseInt(msg.Content["systemBoursesId"], 10, 64)

	connTx, err := mysqlDb.Begin()
	if err != nil {
		connTx.Rollback()
		c.Send <- &Message{
			ServersId: "post-withdraw-order",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  "网络繁忙",
			},
			Uid: c.Uid,
		}
		return
	}

	//查询订单信息
	sqlStr := "select id,userId,bourseType,systemBoursesId,symbol,stockSymbol,stockName,tradeType,isUpsDowns,orderType,entrustNum,entrustStatus,limitPrice,marketPrice,stopPrice,takeProfitPrice,orderAmount,orderAmountSum,earnestMoney,handlingFee,positionPrice,dealPrice,entrustOrderNumber,stockHoldingsId,entrustTime,updateTime from bo_user_trade_orders where id=? and userId=? and bourseType=? and systemBoursesId=? and entrustStatus=? limit 1"
	rows, err := mysqlDb.Query(sqlStr, tradeOrdersId, c.UserId, bourseType, systemBoursesId, 1)
	if err != nil {
		connTx.Rollback()
		c.Send <- &Message{
			ServersId: "post-withdraw-order",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  "网络繁忙",
			},
			Uid: c.Uid,
		}
		return
	}
	defer rows.Close()
	var userTradeOrdersJson models.UserTradeOrdersJson
	var userTradeOrders models.UserTradeOrders
	for rows.Next() {
		rows.Scan(&userTradeOrders.Id, &userTradeOrders.UserId, &userTradeOrders.BourseType, &userTradeOrders.SystemBoursesId, &userTradeOrders.Symbol, &userTradeOrders.StockSymbol, &userTradeOrders.StockName, &userTradeOrders.TradeType, &userTradeOrders.IsUpsDowns, &userTradeOrders.OrderType, &userTradeOrders.EntrustNum, &userTradeOrders.EntrustStatus, &userTradeOrders.LimitPrice, &userTradeOrders.MarketPrice, &userTradeOrders.StopPrice, &userTradeOrders.TakeProfitPrice, &userTradeOrders.OrderAmount, &userTradeOrders.OrderAmountSum, &userTradeOrders.EarnestMoney, &userTradeOrders.HandlingFee, &userTradeOrders.PositionPrice, &userTradeOrders.DealPrice, &userTradeOrders.EntrustOrderNumber, &userTradeOrders.StockHoldingsId, &userTradeOrders.EntrustTime, &userTradeOrders.UpdateTime)
		userTradeOrdersJson = models.ConvUserTradeOrdersNull2Json(userTradeOrders, userTradeOrdersJson)
		break
	}

	//判断订单是否存在
	if userTradeOrdersJson.Id <= 0 {
		c.Send <- &Message{
			ServersId: "post-withdraw-order",
			Content: Errors{
				Code: http.StatusInternalServerError,
				Msg:  "订单不存在",
			},
			Uid: c.Uid,
		}
		return
	}

	//退回资金账户 ==== 针对买入的要退回资金，卖出的不用退
	if userTradeOrdersJson.TradeType == 1 && userTradeOrdersJson.EntrustStatus == 1 {
		//获取根据交易类型用户资金账户信息-Redis查查不到-MySQL查
		//userFundAccountsKey := fmt.Sprintf("%s%v_%v", constant.UserFundAccountsKey, bourseType, c.UserId)
		//result, err := redisDb.Get(context.Background(), userFundAccountsKey).Result()

		var userFundAccountsJson models.UserFundAccountsJson
		sqlStr = "select id,userId,fundAccount,fundAccountType,money,unit,addTime,updateTime from bo_user_fund_accounts where userId=? and fundAccountType=? limit 1"
		rows, err = mysqlDb.Query(sqlStr, c.UserId, bourseType)
		if err != nil {
			logger.Warn(fmt.Sprintf("Time:[%s] Err:[%v]", time.Now().Format(time.DateTime), err.Error()))
			c.Send <- &Message{
				ServersId: "post-withdraw-order",
				Content: &Errors{
					Code: http.StatusInternalServerError,
					Msg:  "网络繁忙",
				},
				Uid: c.Uid,
			}
			return
		}
		defer rows.Close()
		for rows.Next() {
			var userFundAccounts models.UserFundAccounts
			rows.Scan(&userFundAccounts.Id, &userFundAccounts.UserId, &userFundAccounts.FundAccount, &userFundAccounts.FundAccountType, &userFundAccounts.Money, &userFundAccounts.Unit, &userFundAccounts.AddTime, &userFundAccounts.UpdateTime)
			userFundAccountsJson = models.ConvUserFundAccountsNull2Json(userFundAccounts, userFundAccountsJson)
			break
		}

		//5、开启事务--MySQL-修改用户资金账户余额--修改Redis中用户账户资金余额--(写入MQ队列)写入资金账户交易订单表--写入交易订单表
		//5.1 修改资金账户余额&修改Redis中用户账户资金余额(买入操作)
		userFundAccountServe := servers.UserFundAccountServe{
			Id:              userFundAccountsJson.Id,
			UserId:          userFundAccountsJson.UserId,
			FundAccount:     userFundAccountsJson.FundAccount,
			FundAccountType: userFundAccountsJson.FundAccountType,
			Operate:         1,
			OperateMoney:    userTradeOrdersJson.OrderAmount,
			OrderAmount:     userTradeOrdersJson.OrderAmount,
			OrderAmountSum:  userTradeOrdersJson.OrderAmountSum,
			EarnestMoney:    userTradeOrdersJson.EarnestMoney,
			HandlingFee:     userTradeOrdersJson.HandlingFee,
			OperationType:   1,
			OrderNo:         msg.Content["entrustOrderNumber"],
			Remark:          "撤单",
		}
		service, err := servers.UserFundAccountService(&userFundAccountServe, mysqlDb)
		if err != nil || service == false {
			connTx.Rollback()
			logger.Warn(fmt.Sprintf("Time:[%s] Err:[撤单退款失败. %v]", time.Now().Format(time.DateTime), err.Error()))
			c.Send <- &Message{
				ServersId: "post-withdraw-order",
				Content: &Errors{
					Code: http.StatusInternalServerError,
					Msg:  "撤单退款失败",
				},
				Uid: c.Uid,
			}
			return
		}
	}

	//修改订单状态
	sqlStr = "update bo_user_trade_orders set entrustStatus=?,updateTime=? where id=? and userId=? and bourseType=? and systemBoursesId=? and entrustStatus=?"
	exec, err := connTx.Exec(sqlStr, 3, time.Now().Format(time.DateTime), tradeOrdersId, c.UserId, bourseType, systemBoursesId, 1)
	if err != nil {
		connTx.Rollback()
		logger.Warn(fmt.Sprintf("Time:[%s] Err:[%v]", time.Now().Format(time.DateTime), err.Error()))
		c.Send <- &Message{
			ServersId: "post-withdraw-order",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  "网络繁忙",
			},
			Uid: c.Uid,
		}
		return
	}
	affected, err := exec.RowsAffected()
	if err != nil || affected != 1 {
		connTx.Rollback()
		logger.Warn(fmt.Sprintf("Time:[%s] Err:[%v]", time.Now().Format(time.DateTime), err.Error()))
		c.Send <- &Message{
			ServersId: "post-withdraw-order",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  "网络繁忙",
			},
			Uid: c.Uid,
		}
		return
	}
	//获取用户限价订单哈希缓存 redis的哈希key=user_stock_trade_orders_data_+ client.UserId + "_" +systemBoursesId ...
	cacheOrdersKey := fmt.Sprintf("user_stock_trade_orders_data_%d_%d", c.UserId, userTradeOrdersJson.SystemBoursesId)
	redisDb.HDel(context.Background(), cacheOrdersKey, fmt.Sprintf("%d", userTradeOrdersJson.Id)).Err()

	err = connTx.Commit()
	if err != nil {
		connTx.Rollback()
		logger.Warn(fmt.Sprintf("Time:[%s] Err:[%v]", time.Now().Format(time.DateTime), err.Error()))
		c.Send <- &Message{
			ServersId: "post-withdraw-order",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  "网络繁忙,提交事务失败",
			},
			Uid: c.Uid,
		}
		return
	}

	//撤单 ==== 针对买入且为限价交易的撤单移除集合，卖出的撤单不用移除集合
	/*if userTradeOrdersJson.TradeType == 1 && userTradeOrdersJson.OrderType == 1 && userTradeOrdersJson.EntrustStatus == 1 {
		// 移除限价有序集合 。。。 (ZADD stock_order_limit_price_交易所id  价格 订单ID)
		tradeOrdersKey := fmt.Sprintf("stock_order_limit_price_%v", userTradeOrdersJson.SystemBoursesId)
		orderMoney := userTradeOrdersJson.LimitPrice * global.PriceEnlargeMultiples
		redisDb.ZRem(context.Background(), tradeOrdersKey, redis.Z{Score: orderMoney, Member: userTradeOrdersJson.Id}).Err()
	}*/

	c.Send <- &Message{
		ServersId: "post-withdraw-order",
		Content: &Success{
			Code: http.StatusOK,
			Msg:  "撤单成功",
			Data: nil,
		},
		Uid: c.Uid,
	}
	return
}

func (c *Client) PostClosePositionOrder(msg *SymbolMessage) {
	//redis初始化方式 ----- 平仓(对持仓中的订单进行结算)--接口操作为市价平仓-----
	//redisDb := inits.InitRedis()

	//mysql初始化方式
	mysqlDb := inits.InitMysql()
	defer mysqlDb.Close() //关闭数据库连接

	// =========== RabbitMQ Simple模式 发送者 ===========
	//queueKey := "tradeOrderBuy_" //现货 spots 合约 contract
	//rabbitmq = inits.NewRabbitMQSimple("tradeOrderBuy_")

	//1、获取参数，并验证
	var userTradeOrdersForm UserTradeOrdersForm
	userTradeOrdersForm = ConvUserTradeOrdersForm2Msg(userTradeOrdersForm, msg)

	//1、查询该持仓信息
	sqlStr := "select id,userId,bourseType,systemBoursesId,symbol,stockSymbol,stockName,isUpsDowns,orderType,stockHoldingNum,stockHoldingType,limitPrice,marketPrice,stopPrice,takeProfitPrice,orderAmount,orderAmountSum,earnestMoney,handlingFee,positionPrice,dealPrice,buyingPrice,profitLossMoney,entrustOrderNumber,openPositionTime,updateTime from bo_user_stock_holdings where id=? and userId=? and bourseType=? and systemBoursesId=? and stockHoldingType=? and stockSymbol=?"
	rows, err := mysqlDb.Query(sqlStr, userTradeOrdersForm.StockHoldingsId, c.UserId, userTradeOrdersForm.BourseType, userTradeOrdersForm.SystemBoursesId, 1, userTradeOrdersForm.StockSymbol)
	if err != nil {
		logger.Warn(fmt.Sprintf("查询该持仓信息失败 Err:[%s] Time:[%s]", err.Error(), time.Now().Format(time.DateTime)))
		c.Send <- &Message{
			ServersId: "post-close-position-order",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  "网络繁忙",
			},
			Uid: c.Uid,
		}
		return
	}
	defer rows.Close()
	var userStockHoldingsJson models.UserStockHoldingsJson
	for rows.Next() {
		var userStockHoldings models.UserStockHoldings
		rows.Scan(&userStockHoldings.Id, &userStockHoldings.UserId, &userStockHoldings.BourseType, &userStockHoldings.SystemBoursesId, &userStockHoldings.Symbol, &userStockHoldings.StockSymbol, &userStockHoldings.StockName, &userStockHoldings.IsUpsDowns, &userStockHoldings.OrderType, &userStockHoldings.StockHoldingNum, &userStockHoldings.StockHoldingType, &userStockHoldings.LimitPrice, &userStockHoldings.MarketPrice, &userStockHoldings.StopPrice, &userStockHoldings.TakeProfitPrice, &userStockHoldings.OrderAmount, &userStockHoldings.OrderAmountSum, &userStockHoldings.EarnestMoney, &userStockHoldings.HandlingFee, &userStockHoldings.PositionPrice, &userStockHoldings.DealPrice, &userStockHoldings.BuyingPrice, &userStockHoldings.ProfitLossMoney, &userStockHoldings.EntrustOrderNumber, &userStockHoldings.OpenPositionTime, &userStockHoldings.UpdateTime)
		userStockHoldingsJson = models.ConvUserStockHoldingsNull2Json(userStockHoldings, userStockHoldingsJson)
		break
	}

	fmt.Printf("%+v", userStockHoldingsJson)
	//判断持仓是否存在
	if userStockHoldingsJson.Id <= 0 {
		c.Send <- &Message{
			ServersId: "post-close-position-order",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  "持仓不存在",
			},
			Uid: c.Uid,
		}
		return
	}

	var profitLossMoney float64
	var stockHoldingNumF64 float64
	if userStockHoldingsJson.IsUpsDowns == 1 {
		//计算盈亏：股票平仓盈利=(平仓价-开仓价)*数量 (这里按市价计算)
		stockHoldingNumF64, _ = strconv.ParseFloat(fmt.Sprintf("%d", userStockHoldingsJson.StockHoldingNum), 64)
		profitLossMoney = (userStockHoldingsJson.PositionPrice - userTradeOrdersForm.ClosingPrice) * stockHoldingNumF64
	} else if userStockHoldingsJson.IsUpsDowns == 2 {
		//计算盈亏：股票平仓盈利=(开仓价-平仓价)*数量 (这里按市价计算)
		stockHoldingNumF64, _ = strconv.ParseFloat(fmt.Sprintf("%d", userStockHoldingsJson.StockHoldingNum), 64)
		profitLossMoney = (userTradeOrdersForm.ClosingPrice - userStockHoldingsJson.PositionPrice) * stockHoldingNumF64
	}

	//查询对应资金账户信息---结算资金账户
	//获取根据交易类型用户资金账户信息-Redis查查不到-MySQL查
	//userFundAccountsKey := fmt.Sprintf("%s%v_%v", constant.UserFundAccountsKey, userStockHoldingsJson.SystemBoursesId, c.UserId)
	//result, err := redisDb.Get(context.Background(), userFundAccountsKey).Result()
	var userFundAccountsJson models.UserFundAccountsJson
	sqlStr = "select id,userId,fundAccount,fundAccountType,money,unit,addTime,updateTime from bo_user_fund_accounts where userId=? and fundAccountType=? limit 1"
	rows, err = mysqlDb.Query(sqlStr, c.UserId, userStockHoldingsJson.SystemBoursesId)
	if err != nil {
		c.Send <- &Message{
			ServersId: "post-close-position-order",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  "网络繁忙",
			},
			Uid: c.Uid,
		}
		return
	}
	//defer rows.Close()
	for rows.Next() {
		var userFundAccounts models.UserFundAccounts
		rows.Scan(&userFundAccounts.Id, &userFundAccounts.UserId, &userFundAccounts.FundAccount, &userFundAccounts.FundAccountType, &userFundAccounts.Money, &userFundAccounts.Unit, &userFundAccounts.AddTime, &userFundAccounts.UpdateTime)
		userFundAccountsJson = models.ConvUserFundAccountsNull2Json(userFundAccounts, userFundAccountsJson)
		break
	}
	if err := rows.Err(); err != nil {
		c.Send <- &Message{
			ServersId: "post-close-position-order",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  "网络繁忙",
			},
			Uid: c.Uid,
		}
		return
	}

	//计算操作金额
	operateMoney := userStockHoldingsJson.OrderAmount + profitLossMoney
	userFundAccountServe := servers.UserFundAccountServe{
		Id:              userFundAccountsJson.Id,
		UserId:          userFundAccountsJson.UserId,
		FundAccount:     userFundAccountsJson.FundAccount,
		FundAccountType: userFundAccountsJson.FundAccountType,
		Operate:         1,
		OperateMoney:    operateMoney,
		OrderAmount:     userStockHoldingsJson.OrderAmount,
		OrderAmountSum:  userStockHoldingsJson.OrderAmountSum,
		EarnestMoney:    0,
		HandlingFee:     0,
		OperationType:   4,
		OrderNo:         msg.Content["entrustOrderNumber"],
		Remark:          "平仓",
	}
	service, err := servers.UserFundAccountService(&userFundAccountServe, mysqlDb)
	if err != nil || service == false {
		logger.Warn(fmt.Sprintf("金额增减失败 Err:[%s] Time:[%s] 参数:[%+v]", err.Error(), time.Now().Format(time.DateTime), userFundAccountServe))
		c.Send <- &Message{
			ServersId: "post-close-position-order",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  "平仓失败",
			},
			Uid: c.Uid,
		}
		return
	}

	//修改持仓订单状态，并计算盈亏结果(结算)
	sqlStr = "update bo_user_stock_holdings set stockHoldingType=?,dealPrice=?,profitLossMoney=?,updateTime=? where id=? and userId=? and bourseType=? and systemBoursesId=? and stockSymbol=? and stockHoldingType=?"
	exec, err := mysqlDb.Exec(sqlStr, 2, userTradeOrdersForm.ClosingPrice, profitLossMoney, time.Now().Format(time.DateTime), userTradeOrdersForm.StockHoldingsId, c.UserId, userTradeOrdersForm.BourseType, userTradeOrdersForm.SystemBoursesId, userTradeOrdersForm.StockSymbol, 1)
	if err != nil {
		c.Send <- &Message{
			ServersId: "post-close-position-order",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  "网络繁忙",
			},
			Uid: c.Uid,
		}
		return
	}
	affected, err := exec.RowsAffected()
	if err != nil || affected != 1 {
		c.Send <- &Message{
			ServersId: "post-close-position-order",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  "网络繁忙",
			},
			Uid: c.Uid,
		}
		return
	}

	// 移除止损撮合集合
	tradeOrdersKey1 := fmt.Sprintf("stock_order_stop_price_%v", userStockHoldingsJson.SystemBoursesId)
	redisDb.ZRem(context.Background(), tradeOrdersKey1, userStockHoldingsJson.Id).Err()
	// 移除止盈撮合集合
	tradeOrdersKey2 := fmt.Sprintf("stock_order_take_profit_price_%v", userStockHoldingsJson.SystemBoursesId)
	redisDb.ZRem(context.Background(), tradeOrdersKey2, userStockHoldingsJson.Id).Err()

	//移除用户持仓列表
	cacheKey := fmt.Sprintf("user_stock_holdings_data_%d_%d", userStockHoldingsJson.UserId, userStockHoldingsJson.SystemBoursesId)
	redisDb.HDel(context.Background(), cacheKey, fmt.Sprintf("%d", userStockHoldingsJson.Id))

	c.Send <- &Message{
		ServersId: "post-close-position-order",
		Content: &Success{
			Code: http.StatusOK,
			Msg:  "请求成功",
			Data: nil,
		},
		Uid: c.Uid,
	}
	return
}

func (c *Client) GetUserUccountBalance(msg *SymbolMessage) {
	//mysql初始化方式 === 获取对应资金账户余额
	mysqlDb := inits.InitMysql()
	defer mysqlDb.Close() //关闭数据库连接

	//账户总余额  获取现货 获取合约 获取股票
	switch msg.Content["fundAccountType"] {
	case "1":
		//获取现货子账号
		sqlStr := "select id,fundAccountsId,userId,fundAccountSpot,fundAccountName,spotMoney,unit from bo_user_fund_account_spots where userId=? and fundAccountName=?"
		rows, err := mysqlDb.Query(sqlStr, c.UserId, msg.Content["fundAccountName"])
		if err != nil {
			c.Send <- &Message{
				ServersId: "get-user-account-balance",
				Content: &Errors{
					Code: http.StatusInternalServerError,
					Msg:  "网络繁忙",
				},
				Uid: c.Uid,
			}
			break
		}
		defer rows.Close()
		var spotsJson models.UserFundAccountSpotsJson
		for rows.Next() {
			var spots models.UserFundAccountSpots
			rows.Scan(&spots.Id, &spots.FundAccountsId, &spots.UserId, &spots.FundAccountSpot, &spots.FundAccountName, &spots.SpotMoney, &spots.Unit)
			spotsJson = models.ConvUserFundAccountSpotsNull2Json(spots, spotsJson)
			break
		}

		//发送响应消息
		c.Send <- &Message{
			ServersId: "get-user-account-balance",
			Content: &Success{
				Code: http.StatusOK,
				Msg:  "请求成功",
				Data: spotsJson,
			},
			Uid: c.Uid,
		}
	default:
		fmt.Println("非法传参")
		break
	}

	//获取现货子账号列表
	sqlStr := "select id,fundAccountsId,userId,fundAccountSpot,fundAccountName,spotMoney,unit from bo_user_fund_account_spots where userId=?"
	rows, err := mysqlDb.Query(sqlStr, c.UserId)
	if err != nil {
		logger.Warn(fmt.Sprintf("Err:[Query Sql failed.] SQl:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err.Error()))
		c.Send <- &Message{
			ServersId: "get-user-account-balance",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  "网络繁忙",
			},
			Uid: c.Uid,
		}
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
	}

	var userUccountBalanceList UserUccountBalanceList
	userUccountBalanceList.UserFundAccountsSpotsJson = c.GetFundAccountsFindClient(mysqlDb, 1)
	userUccountBalanceList.UserFundAccountsSwapsJson = c.GetFundAccountsFindClient(mysqlDb, 2)
	userUccountBalanceList.UserFundAccountsMalaysiaJson = c.GetFundAccountsFindClient(mysqlDb, 3)
	userUccountBalanceList.UserFundAccountsUSJson = c.GetFundAccountsFindClient(mysqlDb, 4)
	userUccountBalanceList.UserFundAccountsYenJson = c.GetFundAccountsFindClient(mysqlDb, 5)
	userUccountBalanceList.UserFundAccountsBrokerageJson = c.GetFundAccountsFindClient(mysqlDb, 6)
	userUccountBalanceList.SpotsJsonMap = spotsJsonMap

	return
}

func (c *Client) GetUserStockHoldingsNum(msg *SymbolMessage) {
	//mysql初始化方式 == 根据id获取持股数量
	mysqlDb := inits.InitMysql()
	defer mysqlDb.Close() //关闭数据库连接

	//根据持仓id获取持仓数量
	sqlStr := "select id,userId,bourseType,systemBoursesId,symbol,stockSymbol,stockName,isUpsDowns,orderType,stockHoldingNum,stockHoldingType,limitPrice,marketPrice,stopPrice,takeProfitPrice,orderAmount,orderAmountSum,earnestMoney,handlingFee,positionPrice,dealPrice,buyingPrice,profitLossMoney,entrustOrderNumber,openPositionTime,updateTime from bo_user_stock_holdings where userId=? and bourseType=? and systemBoursesId=? and stockHoldingType=? and stockSymbol=? limit 1"
	rows, err := mysqlDb.Query(sqlStr, c.Uid, msg.Content["bourseType"], msg.Content["systemBoursesId"], 1, msg.Content["stockSymbol"])
	if err != nil {
		logger.Warn(fmt.Sprintf("Exec SQL Failed:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err))
		//发送响应消息
		c.Send <- &Message{
			ServersId: "get-user-account-balance",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  "网络繁忙",
			},
			Uid: c.Uid,
		}
		return
	}
	defer rows.Close()
	var userStockHoldingsJson models.UserStockHoldingsJson
	for rows.Next() {
		var userStockHoldings models.UserStockHoldings
		rows.Scan(&userStockHoldings.Id, &userStockHoldings.UserId, &userStockHoldings.BourseType, &userStockHoldings.SystemBoursesId, &userStockHoldings.Symbol, &userStockHoldings.StockSymbol, &userStockHoldings.StockName, &userStockHoldings.IsUpsDowns, &userStockHoldings.OrderType, &userStockHoldings.StockHoldingNum, &userStockHoldings.StockHoldingType, &userStockHoldings.LimitPrice, &userStockHoldings.MarketPrice, &userStockHoldings.StopPrice, &userStockHoldings.TakeProfitPrice, &userStockHoldings.OrderAmount, &userStockHoldings.OrderAmountSum, &userStockHoldings.EarnestMoney, &userStockHoldings.HandlingFee, &userStockHoldings.PositionPrice, &userStockHoldings.DealPrice, &userStockHoldings.BuyingPrice, &userStockHoldings.ProfitLossMoney, &userStockHoldings.EntrustOrderNumber, &userStockHoldings.OpenPositionTime, &userStockHoldings.UpdateTime)
		userStockHoldingsJson = models.ConvUserStockHoldingsNull2Json(userStockHoldings, userStockHoldingsJson)
	}

	//发送响应消息
	c.Send <- &Message{
		ServersId: "get-user-stock-holdings-num",
		Content: &Success{
			Code: http.StatusOK,
			Msg:  "请求成功",
			Data: userStockHoldingsJson,
		},
		Uid: c.Uid,
	}
	return
}

func (c *Client) GetListUserTradePosition(msg *SymbolMessage) {
	//mysql初始化方式 == 根据id获取持股数量 -----获取用户交易持仓列表
	mysqlDb := inits.InitMysql()
	defer mysqlDb.Close() //关闭数据库连接

	//redisDb := inits.InitRedis()

	sqlStr := GetListUserTradePositionSql()
	rows, err := mysqlDb.Query(sqlStr, c.UserId, msg.Content["bourseType"], msg.Content["systemBoursesId"], 1)
	if err != nil {
		logger.Warn(fmt.Sprintf("Exec SQL Failed:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err))
		//发送响应消息
		c.Send <- &Message{
			ServersId: "get-user-account-balance",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  "网络繁忙",
			},
			Uid: c.Uid,
		}
		return
	}
	defer rows.Close()
	var userStockHoldingsJsonMap []*models.UserStockHoldingsJson
	//holdingsSpotMap := make(map[string]string, 0) //现货
	//holdingsSwapMap := make(map[string]string, 0) //合约

	for rows.Next() {
		var userStockHoldingsJson models.UserStockHoldingsJson
		var userStockHoldings models.UserStockHoldings
		rows.Scan(&userStockHoldings.Id, &userStockHoldings.UserId, &userStockHoldings.BourseType, &userStockHoldings.SystemBoursesId, &userStockHoldings.Symbol, &userStockHoldings.StockSymbol, &userStockHoldings.StockName, &userStockHoldings.IsUpsDowns, &userStockHoldings.OrderType, &userStockHoldings.StockHoldingNum, &userStockHoldings.StockHoldingType, &userStockHoldings.LimitPrice, &userStockHoldings.MarketPrice, &userStockHoldings.StopPrice, &userStockHoldings.TakeProfitPrice, &userStockHoldings.OrderAmount, &userStockHoldings.OrderAmountSum, &userStockHoldings.EarnestMoney, &userStockHoldings.HandlingFee, &userStockHoldings.PositionPrice, &userStockHoldings.DealPrice, &userStockHoldings.BuyingPrice, &userStockHoldings.ProfitLossMoney, &userStockHoldings.EntrustOrderNumber, &userStockHoldings.OpenPositionTime, &userStockHoldings.UpdateTime)
		userStockHoldingsJson = models.ConvUserStockHoldingsNull2Json(userStockHoldings, userStockHoldingsJson)
		userStockHoldingsJsonMap = append(userStockHoldingsJsonMap, &userStockHoldingsJson)

		//组装缓存key (用户持仓列表的key)
		cacheKey1 := fmt.Sprintf("user_stock_holdings_data_%d_%v", c.UserId, userStockHoldingsJson.SystemBoursesId)
		holdingsJson, _ := json.Marshal(&userStockHoldingsJson)
		redisDb.HSet(context.Background(), cacheKey1, userStockHoldingsJson.Id, holdingsJson).Err()
	}

	//发送响应消息
	c.Send <- &Message{
		ServersId: "get-list-user-trade-position",
		Content: &Success{
			Code: http.StatusOK,
			Msg:  "请求成功",
			Data: userStockHoldingsJsonMap,
		},
		Uid: c.Uid,
	}
	return
}

func (c *Client) GetListUserTradePendingWithdrawalClearing(msg *SymbolMessage) {
	//mysql初始化方式 == 根据id获取持股数量 -----获取用户交易撤单-挂单-结算列表
	mysqlDb := inits.InitMysql()
	defer mysqlDb.Close() //关闭数据库连接

	sqlStr := GetListUserTradePendingWithdrawalClearingSql()
	//1 挂单 Pending
	rows, err := mysqlDb.Query(sqlStr, c.UserId, msg.Content["bourseType"], msg.Content["systemBoursesId"], 1, 1)
	if err != nil {
		logger.Warn(fmt.Sprintf("Exec SQL Failed:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err))
		//发送响应消息
		c.Send <- &Message{
			ServersId: "get-user-account-balance",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  "网络繁忙",
			},
			Uid: c.Uid,
		}
		return
	}
	defer rows.Close()
	var pendingJsonMap []*models.UserTradeOrdersJson
	for rows.Next() {
		var pending models.UserTradeOrders
		var pendingJson models.UserTradeOrdersJson
		rows.Scan(&pending.Id, &pending.UserId, &pending.BourseType, &pending.SystemBoursesId, &pending.Symbol, &pending.StockSymbol, &pending.StockName, &pending.TradeType, &pending.IsUpsDowns, &pending.OrderType, &pending.EntrustNum, &pending.EntrustStatus, &pending.LimitPrice, &pending.MarketPrice, &pending.StopPrice, &pending.TakeProfitPrice, &pending.OrderAmount, &pending.OrderAmountSum, &pending.EarnestMoney, &pending.HandlingFee, &pending.PositionPrice, &pending.DealPrice, &pending.EntrustOrderNumber, &pending.StockHoldingsId, &pending.EntrustTime, &pending.UpdateTime)
		pendingJson = models.ConvUserTradeOrdersNull2Json(pending, pendingJson)
		pendingJsonMap = append(pendingJsonMap, &pendingJson)
	}

	//2 撤单 Withdrawal
	rows, err = mysqlDb.Query(sqlStr, c.UserId, msg.Content["bourseType"], msg.Content["systemBoursesId"], 1, 3)
	if err != nil {
		logger.Warn(fmt.Sprintf("Exec SQL Failed:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err))
		//发送响应消息
		c.Send <- &Message{
			ServersId: "get-user-account-balance",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  "网络繁忙",
			},
			Uid: c.Uid,
		}
		return
	}
	defer rows.Close()
	var withdrawalJsonMap []*models.UserTradeOrdersJson
	for rows.Next() {
		var withdrawal models.UserTradeOrders
		var withdrawalJson models.UserTradeOrdersJson
		rows.Scan(&withdrawal.Id, &withdrawal.UserId, &withdrawal.BourseType, &withdrawal.SystemBoursesId, &withdrawal.Symbol, &withdrawal.StockSymbol, &withdrawal.StockName, &withdrawal.TradeType, &withdrawal.IsUpsDowns, &withdrawal.OrderType, &withdrawal.EntrustNum, &withdrawal.EntrustStatus, &withdrawal.LimitPrice, &withdrawal.MarketPrice, &withdrawal.StopPrice, &withdrawal.TakeProfitPrice, &withdrawal.OrderAmount, &withdrawal.OrderAmountSum, &withdrawal.EarnestMoney, &withdrawal.HandlingFee, &withdrawal.PositionPrice, &withdrawal.DealPrice, &withdrawal.EntrustOrderNumber, &withdrawal.StockHoldingsId, &withdrawal.EntrustTime, &withdrawal.UpdateTime)
		withdrawalJson = models.ConvUserTradeOrdersNull2Json(withdrawal, withdrawalJson)
		withdrawalJsonMap = append(withdrawalJsonMap, &withdrawalJson)
	}

	//3 结算 Clearing
	rows, err = mysqlDb.Query(sqlStr, c.UserId, msg.Content["bourseType"], msg.Content["systemBoursesId"], 2, 3)
	if err != nil {
		logger.Warn(fmt.Sprintf("Exec SQL Failed:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err))
		//发送响应消息
		c.Send <- &Message{
			ServersId: "get-user-account-balance",
			Content: &Errors{
				Code: http.StatusInternalServerError,
				Msg:  "网络繁忙",
			},
			Uid: c.Uid,
		}
		return
	}
	defer rows.Close()
	var clearingJsonMap []*models.UserTradeOrdersJson
	for rows.Next() {
		var clearing models.UserTradeOrders
		var clearingJson models.UserTradeOrdersJson
		rows.Scan(&clearing.Id, &clearing.UserId, &clearing.BourseType, &clearing.SystemBoursesId, &clearing.Symbol, &clearing.StockSymbol, &clearing.StockName, &clearing.TradeType, &clearing.IsUpsDowns, &clearing.OrderType, &clearing.EntrustNum, &clearing.EntrustStatus, &clearing.LimitPrice, &clearing.MarketPrice, &clearing.StopPrice, &clearing.TakeProfitPrice, &clearing.OrderAmount, &clearing.OrderAmountSum, &clearing.EarnestMoney, &clearing.HandlingFee, &clearing.PositionPrice, &clearing.DealPrice, &clearing.EntrustOrderNumber, &clearing.StockHoldingsId, &clearing.EntrustTime, &clearing.UpdateTime)
		clearingJson = models.ConvUserTradeOrdersNull2Json(clearing, clearingJson)
		withdrawalJsonMap = append(clearingJsonMap, &clearingJson)
	}

	var pendingWithdrawalClearing UserTradeOrdersPendingWithdrawalClearing
	pendingWithdrawalClearing.Pending = pendingJsonMap
	pendingWithdrawalClearing.Withdrawal = withdrawalJsonMap
	pendingWithdrawalClearing.Clearing = clearingJsonMap

	//发送响应消息
	c.Send <- &Message{
		ServersId: "get-list-user-trade-pending-withdrawal-clearing",
		Content: &Success{
			Code: http.StatusOK,
			Msg:  "请求成功",
			Data: pendingWithdrawalClearing,
		},
		Uid: c.Uid,
	}
	return
}

// ===================================== SQL ==============================================

func GetListUserTradePositionSql() (sqlStr string) {
	sqlStr = "select id,userId,bourseType,systemBoursesId,symbol,stockSymbol,stockName,isUpsDowns,orderType,stockHoldingNum,stockHoldingType,limitPrice,marketPrice,stopPrice,takeProfitPrice,orderAmount,orderAmountSum,earnestMoney,handlingFee,positionPrice,dealPrice,buyingPrice,profitLossMoney,entrustOrderNumber,openPositionTime,updateTime from bo_user_stock_holdings where userId=? and bourseType=? and systemBoursesId=? and stockHoldingType=?"
	return
}

func GetListUserTradePendingWithdrawalClearingSql() (sqlStr string) {
	sqlStr = "select id,userId,bourseType,systemBoursesId,symbol,stockSymbol,stockName,tradeType,isUpsDowns,orderType,entrustNum,entrustStatus,limitPrice,marketPrice,stopPrice,takeProfitPrice,orderAmount,orderAmountSum,earnestMoney,handlingFee,positionPrice,dealPrice,entrustOrderNumber,stockHoldingsId,entrustTime,updateTime from bo_user_trade_orders where userId=? and bourseType=? and systemBoursesId=? and tradeType=? and entrustStatus=?"
	return
}

func (c *Client) GetFundAccountsFindClient(conn *sql.DB, fundAccountType int) *models.UserFundAccountsJson {
	sqlStr := "select id,userId,fundAccount,fundAccountType,money,unit,addTime,updateTime from bo_user_fund_accounts where userId=? and fundAccountType=? limit 1"
	rows, err := conn.Query(sqlStr, c.UserId, fundAccountType)
	if err != nil {
		logger.Warn(fmt.Sprintf("Err:[Query Sql failed.] SQl:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err.Error()))
		return nil
	}
	defer rows.Close()
	var userFundAccountsMalaysiaJson models.UserFundAccountsJson
	for rows.Next() {
		var userFundAccounts models.UserFundAccounts
		rows.Scan(&userFundAccounts.Id, &userFundAccounts.UserId, &userFundAccounts.FundAccount, &userFundAccounts.FundAccountType, &userFundAccounts.Money, &userFundAccounts.Unit)
		userFundAccountsMalaysiaJson = models.ConvUserFundAccountsNull2Json(userFundAccounts, userFundAccountsMalaysiaJson)
		break
	}
	return &userFundAccountsMalaysiaJson
}

func ConvUserTradeOrdersForm2Msg(userTradeOrdersForm UserTradeOrdersForm, msg *SymbolMessage) UserTradeOrdersForm {
	if len(msg.Content["bourseType"]) > 0 {
		bourseType, _ := strconv.Atoi(msg.Content["bourseType"])
		userTradeOrdersForm.BourseType = int32(bourseType)
	}

	if len(msg.Content["systemBoursesId"]) > 0 {
		systemBoursesId, _ := strconv.ParseInt(msg.Content["systemBoursesId"], 10, 64)
		userTradeOrdersForm.SystemBoursesId = systemBoursesId
	}
	if len(msg.Content["symbol"]) > 0 {
		userTradeOrdersForm.Symbol = msg.Content["symbol"]
	}
	if len(msg.Content["stockSymbol"]) > 0 {
		userTradeOrdersForm.StockSymbol = msg.Content["stockSymbol"]
	}
	if len(msg.Content["stockName"]) > 0 {
		userTradeOrdersForm.StockName = msg.Content["stockName"]
	}
	if len(msg.Content["tradeType"]) > 0 {
		tradeType, _ := strconv.Atoi(msg.Content["tradeType"])
		userTradeOrdersForm.TradeType = int32(tradeType)
	}
	if len(msg.Content["isUpsDowns"]) > 0 {
		isUpsDowns, _ := strconv.Atoi(msg.Content["isUpsDowns"])
		userTradeOrdersForm.IsUpsDowns = int32(isUpsDowns)
	}
	if len(msg.Content["orderType"]) > 0 {
		orderType, _ := strconv.Atoi(msg.Content["orderType"])
		userTradeOrdersForm.OrderType = int32(orderType)
	}
	if len(msg.Content["entrustNum"]) > 0 {
		entrustNum, _ := strconv.Atoi(msg.Content["entrustNum"])
		userTradeOrdersForm.EntrustNum = int32(entrustNum)
	}
	if len(msg.Content["limitPrice"]) > 0 {
		limitPrice, _ := strconv.ParseFloat(msg.Content["limitPrice"], 64)
		userTradeOrdersForm.LimitPrice = limitPrice
	}
	if len(msg.Content["marketPrice"]) > 0 {
		marketPrice, _ := strconv.ParseFloat(msg.Content["marketPrice"], 64)
		userTradeOrdersForm.MarketPrice = marketPrice
	}
	if len(msg.Content["stopPrice"]) > 0 {
		stopPrice, _ := strconv.ParseFloat(msg.Content["stopPrice"], 64)
		userTradeOrdersForm.StopPrice = stopPrice
	}
	if len(msg.Content["takeProfitPrice"]) > 0 {
		takeProfitPrice, _ := strconv.ParseFloat(msg.Content["takeProfitPrice"], 64)
		userTradeOrdersForm.TakeProfitPrice = takeProfitPrice
	}
	if len(msg.Content["orderAmount"]) > 0 {
		orderAmount, _ := strconv.ParseFloat(msg.Content["orderAmount"], 64)
		userTradeOrdersForm.OrderAmount = orderAmount
	}
	if len(msg.Content["orderAmountSum"]) > 0 {
		orderAmountSum, _ := strconv.ParseFloat(msg.Content["orderAmountSum"], 64)
		userTradeOrdersForm.OrderAmountSum = orderAmountSum
	}
	if len(msg.Content["earnestMoney"]) > 0 {
		earnestMoney, _ := strconv.ParseFloat(msg.Content["earnestMoney"], 64)
		userTradeOrdersForm.EarnestMoney = earnestMoney
	}
	if len(msg.Content["handlingFee"]) > 0 {
		handlingFee, _ := strconv.ParseFloat(msg.Content["handlingFee"], 64)
		userTradeOrdersForm.HandlingFee = handlingFee
	}
	if len(msg.Content["positionPrice"]) > 0 {
		positionPrice, _ := strconv.ParseFloat(msg.Content["positionPrice"], 64)
		userTradeOrdersForm.PositionPrice = positionPrice
	}
	if len(msg.Content["closingPrice"]) > 0 {
		closingPrice, _ := strconv.ParseFloat(msg.Content["closingPrice"], 64)
		userTradeOrdersForm.ClosingPrice = closingPrice
	}
	if len(msg.Content["stockHoldingsId"]) > 0 {
		stockHoldingsId, _ := strconv.ParseInt(msg.Content["stockHoldingsId"], 10, 64)
		userTradeOrdersForm.StockHoldingsId = stockHoldingsId
	}
	return userTradeOrdersForm
}

// GetInsertUserStockHoldingsSql 获取写入持仓订单SQL
func GetInsertUserStockHoldingsSql(userTradeOrdersJson UserTradeOrdersForm, userId int64, entrustOrderNumber string) string {
	addSql := fmt.Sprintf("insert into bo_user_stock_holdings (userId,bourseType,systemBoursesId,symbol,stockSymbol,stockName,isUpsDowns,orderType,stockHoldingNum,stockHoldingType,limitPrice,marketPrice,stopPrice,takeProfitPrice,orderAmount,orderAmountSum,earnestMoney,handlingFee,positionPrice,dealPrice,entrustOrderNumber,openPositionTime,updateTime) values (%d,%d,%d,'%s','%s','%s',%d,%d,%d,%d,%f,%f,%f,%f,%f,%f,%f,%f,%f,%d,'%s','%s','%s')", userId, userTradeOrdersJson.BourseType, userTradeOrdersJson.SystemBoursesId, userTradeOrdersJson.Symbol, userTradeOrdersJson.StockSymbol, userTradeOrdersJson.StockName, userTradeOrdersJson.IsUpsDowns, userTradeOrdersJson.OrderType, userTradeOrdersJson.EntrustNum, 1, userTradeOrdersJson.LimitPrice, userTradeOrdersJson.MarketPrice, userTradeOrdersJson.StopPrice, userTradeOrdersJson.TakeProfitPrice, userTradeOrdersJson.OrderAmount, userTradeOrdersJson.OrderAmountSum, userTradeOrdersJson.EarnestMoney, userTradeOrdersJson.HandlingFee, userTradeOrdersJson.PositionPrice, 0, entrustOrderNumber, time.Now().Format(time.DateTime), time.Now().Format(time.DateTime))
	return addSql
}
