package main

import (
	"Bourse/app/models"
	"Bourse/common/inits"
	"Bourse/common/translators"
	"Bourse/config"
	"Bourse/config/constant"
	"database/sql"
	"encoding/json"
	"fmt"
	"gitee.com/phper95/pkg/logger"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	conn   *sql.DB
	client *redis.Client

	UserInfo *constant.UserInfoStruct

	file string
)

const (
	TableUserTradeOrders    = "bo_user_trade_orders"
	TableUserFundAccounts   = "bo_user_fund_accounts"
	TableUserOptionalStocks = "bo_user_optional_stocks"
	TableUserStockHoldings  = "bo_user_stock_holdings"
)

// UserTradeOrdersForm 交易订单表-add入参
type UserTradeOrdersForm struct {
	BourseType         int32   `json:"bourseType"`         //所属类型：1 现货 2 合约 3 马股 4 美股 5 日股
	SystemBoursesId    int64   `json:"systemBoursesId"`    //交易所股种id
	Symbol             string  `json:"symbol"`             //交易代码
	StockSymbol        string  `json:"stockSymbol"`        //股票代码
	StockName          string  `json:"stockName"`          //股票名字
	TradeType          int32   `json:"tradeType"`          //交易类型：1买入 2 卖出
	OrderType          int32   `json:"orderType"`          //订单类型：0 未知 1 限价 2 市价 3 止损止盈
	EntrustNum         int32   `json:"entrustNum"`         //委托数量
	LimitPrice         float64 `json:"limitPrice"`         //限价
	MarketPrice        float64 `json:"marketPrice"`        //市价
	OrderAmount        float64 `json:"orderAmount"`        //订单金额
	LatestPrice        float64 `json:"latestPrice"`        //最新价
	PositionPrice      float64 `json:"positionPrice"`      //持仓价
	EarnestMoney       float64 `json:"earnestMoney"`       //保证金
	HandlingFee        float64 `json:"handlingFee"`        //手续费
	StopPrice          float64 `json:"stopPrice"`          //止损价
	TakeProfitPrice    float64 `json:"takeProfitPrice"`    //止盈价
	EntrustOrderNumber string  `json:"entrustOrderNumber"` //委托订单号
}

func init() {
	//初始化配置方式
	config.LoadConfig()
	inits.InitLog()

	//redis初始化方式
	client = inits.InitRedis()

	file = "daemon_users_init_redis"

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
		true,
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
			//实现我们要处理的逻辑函数
			tradeOrder(d.Body)
			logger.Warn(fmt.Sprintf("file:%s  Received a message:[%s] Time:[%s]", file, d.Body, time.Now().Format(time.DateTime)))
			//logger.Warn(fmt.Sprintf("Received a message:%s", d.Body))
			//fmt.Println(d.Body)

			d.Ack(false)
		}

	}()
	logger.Warn(fmt.Sprintf("【*】warting for messages, To exit press CCTRAL+C"))
	<-forever
}

func tradeOrder(body []byte) {
	//解析消息
	var msg models.UserTradeOrdersJson
	err := json.Unmarshal(body, &msg)
	if err != nil {
		logger.Warn(fmt.Sprintf("file:%s Time:[%s] Err:[%v]", file, time.Now().Format(time.DateTime), err))
		return
	}
	/*id                 //交易订单id
	userId             //用户id
	bourseType         //所属类型：1 现货 2 合约 3 马股 4 美股 5 日股
	systemBoursesId    //交易所股种id
	symbol             //交易代码
	stockSymbol        //股票代码
	stockName          //股票名字
	tradeType          //交易类型：1买入 2 卖出
	isUpsDowns         //是否买涨跌：1 买涨 2 买跌
	orderType          //订单类型：0 未知 1 限价 2 市价 3 止损止盈
	entrustNum         //委托数量
	entrustStatus      //委托状态：1 待成交 2  成功 3 撤单
	limitPrice         //限价
	marketPrice        //市价
	orderAmount        //订单金额
	latestPrice        //最新价
	positionPrice      //持仓价
	earnestMoney       //保证金
	handlingFee        //手续费
	stopPrice          //止损价
	takeProfitPrice    //止盈价
	entrustOrderNumber //委托订单号
	entrustTime        //委托时间
	updateTime         //更新时间（平仓时间）*/

	//mysql初始化方式
	conn = inits.InitMysql()
	defer conn.Close() //关闭数据库连接

	switch msg.BourseType {
	case 1: //数据货币--现货
		//1、写入交易订单表
		//写入数据库
		sqlStr := GetTradeOrdersSql(msg)
		exec, err := conn.Exec(sqlStr)
		if err != nil {
			fmt.Printf("插入失败 SQL:[%s] Err:[%s]", sqlStr, err.Error())
			return
		}
		lastInsertId, err := exec.LastInsertId()
		if err != nil || lastInsertId <= 0 {
			fmt.Printf("获取LastInsertId失败 Err:[%s]", err.Error())
			return
		}
		//写入有序集合价格

	//2、写入用户资金账户对应订单表
	//3、写入对应资金账户流水表

	case 2: //数据货币--合约

	case 3: //股票--马股

	case 4: //股票--美股

	case 5: //股票--日股

	default:
		logger.Warn(fmt.Sprintf("file:%s Time:[%s] Err:[非法操作]", file, time.Now().Format(time.DateTime)))
		break
	}

	return
}

func GetTradeOrdersSql(ordersJson models.UserTradeOrdersJson) string {
	//sqlStr := "insert into bo_user_trade_orders (userId,bourseType,systemBoursesId,symbol,stockSymbol,stockName,tradeType,isUpsDowns,orderType,entrustNum,entrustStatus,limitPrice,marketPrice,orderAmount,latestPrice,positionPrice,earnestMoney,handlingFee,stopPrice,takeProfitPrice,entrustOrderNumber,entrustTime) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
	sqlStr := "insert into bo_user_trade_orders (userId,entrustStatus,entrustTime"
	values := fmt.Sprintf(" values (%v,%v,%v", ordersJson.UserId, 1, time.Now().Format(time.DateTime))
	if ordersJson.BourseType >= 1 && ordersJson.BourseType <= 5 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "bourseType")
		values = fmt.Sprintf("%s,%v", values, ordersJson.BourseType)
	}
	if ordersJson.SystemBoursesId > 0 && ordersJson.SystemBoursesId <= 5 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "systemBoursesId")
		values = fmt.Sprintf("%s,%v", values, ordersJson.SystemBoursesId)
	}
	if len(ordersJson.Symbol) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "symbol")
		values = fmt.Sprintf("%s,%s", values, ordersJson.Symbol)
	}
	if len(ordersJson.StockSymbol) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "stockSymbol")
		values = fmt.Sprintf("%s,%v", values, ordersJson.StockSymbol)
	}
	if len(ordersJson.StockName) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "stockName")
		values = fmt.Sprintf("%s,%v", values, ordersJson.StockName)
	}
	if ordersJson.TradeType >= 1 && ordersJson.TradeType <= 2 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "tradeType")
		values = fmt.Sprintf("%s,%v", values, ordersJson.TradeType)
	}
	if ordersJson.IsUpsDowns >= 1 && ordersJson.IsUpsDowns <= 2 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "isUpsDowns")
		values = fmt.Sprintf("%s,%v", values, ordersJson.BourseType)
	}
	if ordersJson.OrderType >= 1 && ordersJson.OrderType <= 3 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "orderType")
		values = fmt.Sprintf("%s,%v", values, ordersJson.OrderType)
	}
	if ordersJson.EntrustNum >= 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "entrustNum")
		values = fmt.Sprintf("%s,%v", values, ordersJson.EntrustNum)
	}
	if ordersJson.LimitPrice > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "limitPrice")
		values = fmt.Sprintf("%s,%v", values, ordersJson.LimitPrice)
	}
	if ordersJson.MarketPrice > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "marketPrice")
		values = fmt.Sprintf("%s,%v", values, ordersJson.MarketPrice)
	}
	if ordersJson.OrderAmount > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "orderAmount")
		values = fmt.Sprintf("%s,%v", values, ordersJson.OrderAmount)
	}
	/*	if ordersJson.LatestPrice > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "latestPrice")
		values = fmt.Sprintf("%s,%v", values, ordersJson.LatestPrice)
	}*/
	if ordersJson.PositionPrice > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "positionPrice")
		values = fmt.Sprintf("%s,%v", values, ordersJson.PositionPrice)
	}
	if ordersJson.EarnestMoney > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "earnestMoney")
		values = fmt.Sprintf("%s,%v", values, ordersJson.EarnestMoney)
	}
	if ordersJson.HandlingFee > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "handlingFee")
		values = fmt.Sprintf("%s,%v", values, ordersJson.HandlingFee)
	}
	if ordersJson.StopPrice > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "stopPrice")
		values = fmt.Sprintf("%s,%v", values, ordersJson.StopPrice)
	}
	if ordersJson.TakeProfitPrice > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "takeProfitPrice")
		values = fmt.Sprintf("%s,%v", values, ordersJson.TakeProfitPrice)
	}
	if len(ordersJson.EntrustOrderNumber) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "entrustOrderNumber")
		values = fmt.Sprintf("%s,%v", values, ordersJson.EntrustOrderNumber)
	}
	return sqlStr + values
}

func main() {
	queueKey := "tradeOrderBuy_" //现货 spots 合约 contract
	tradeOrderConsumeSimpleMq(queueKey)
}
