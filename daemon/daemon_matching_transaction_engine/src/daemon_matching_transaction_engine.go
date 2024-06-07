package main

import (
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

// =============== 撮合交易引擎脚本 ====================

// SymbolMessage 聚合行情类信息用
type SymbolMessage struct {
	Type    string            `json:"type"`              // 消息类型
	Symbol  string            `json:"symbol,omitempty"`  // 订阅接口内容
	Content map[string]string `json:"content,omitempty"` // 发送内容
	/*ServersId string `json:"serversId,omitempty"` // 服务ID标识
	Content   any    `json:"content,omitempty"`   // 发送内容
	Symbol    string `json:"symbol,omitempty"`    // 订阅类型*/
}

type Message struct {
	ServersId string `json:"serversId,omitempty"` // Service ID (used to receive the client push ID of the third-party push server)
	Sender    string `json:"sender,omitempty"`    // sender
	Recipient string `json:"recipient,omitempty"` // recipient
	Content   any    `json:"content,omitempty"`   // send content
	Symbol    string `json:"symbol,omitempty"`    // Subscription type
	Uid       string `json:"uid,omitempty"`
}

// GetSpotsMessage 本平台的行情数据结构--现货
type GetSpotsMessage struct {
	ServersId string `json:"serversId"`
	Content   struct {
		Ch   string `json:"ch"`
		Ts   int64  `json:"ts"`
		Tick struct {
			Open      string `json:"open"`
			High      string `json:"high"`
			Low       string `json:"low"`
			Close     string `json:"close"`
			Amount    string `json:"amount"`
			Count     int    `json:"count"`
			Bid       string `json:"bid"`
			BidSize   string `json:"bidSize"`
			Ask       string `json:"ask"`
			AskSize   string `json:"askSize"`
			LastPrice string `json:"lastPrice"`
			LastSize  string `json:"lastSize"`
		} `json:"Tick"`
		Data interface{} `json:"Data"`
	} `json:"content"`
	Symbol string `json:"symbol "`
}

// GetSwapsMessage 本平台的行情数据结构--合约
type GetSwapsMessage struct {
	ServersId string `json:"serversId"`
	Content   struct {
		Ch   string `json:"ch"`
		Ts   int64  `json:"ts"`
		Tick struct {
			Id            int         `json:"id"`
			Mrid          int64       `json:"mrid"`
			Open          string      `json:"open"`
			Close         string      `json:"close"`
			High          string      `json:"high"`
			Low           string      `json:"low"`
			Amount        string      `json:"amount"`
			Vol           string      `json:"vol"`
			TradeTurnover string      `json:"trade_turnover"`
			Count         int         `json:"count"`
			Asks          interface{} `json:"asks"`
			Bids          interface{} `json:"bids"`
		} `json:"Tick"`
	} `json:"content"`
	Symbol string `json:"symbol"`
}

// StockHoldingsJson 用户持仓股表
type StockHoldingsJson struct {
	Id              int64   `json:"id,omitempty"`              //持仓股id
	SystemBoursesId int64   `json:"systemBoursesId,omitempty"` //交易所股种id
	ProfitAndLoss   float64 `json:"profitAndLoss,omitempty"`   //盈亏
	LimitPrice      float64 `json:"limitPrice,omitempty"`      //限价
	MarketPrice     float64 `json:"marketPrice,omitempty"`     //市价
	OrderAmount     float64 `json:"orderAmount,omitempty"`     //订单金额
	LatestPrice     float64 `json:"latestPrice,omitempty"`     //最新价
	PositionPrice   float64 `json:"positionPrice,omitempty"`   //持仓价
}

func init() {
	//初始化配置方式
	config.LoadConfig()
	inits.InitLog()

	//redis初始化方式
	redisDb = inits.InitRedis()

	file = "daemon_matching_transaction_engine"

	//多语言
	if err := translators.InitTranslators("zh"); err != nil {
		fmt.Sprintf("init Translator failed,err:%v\n", err)
		logger.Warn(fmt.Sprintf("file:%s  Failed:[init Translator failed] Time:[%s] Err:[%v]", file, time.Now().Format(time.DateTime), err))
	}
}

// AccountSpotsSubscribe 获取现货行情
func AccountSpotsSubscribe() {
	subscribeSpotsMap := make(map[int]string, 0)
	subscribeSpotsMap[1] = "market.btcusdt.ticker"
	subscribeSpotsMap[2] = "market.ethusdt.ticker"
	subscribeSpotsMap[3] = "market.ltcusdt.ticker"
	subscribeSpotsMap[4] = "market.dogeusdt.ticker"
	subscribeSpotsMap[5] = "market.shibusdt.ticker"
	subscribeSpotsMap[6] = "market.xrpusdt.ticker"
	subscribeSpotsMap[7] = "market.linkusdt.ticker"
	subscribeSpotsMap[8] = "market.trxusdt.ticker"
	subscribeSpotsMap[9] = "market.dotusdt.ticker"
	subscribeSpotsMap[10] = "market.adausdt.ticker"
	subscribeSpotsMap[11] = "market.eosusdt.ticker"
	subscribeSpotsMap[12] = "market.bchusdt.ticker"

	for _, subscribe := range subscribeSpotsMap {
		pubSub := redisDb.Subscribe(context.Background(), subscribe) //msg.Symbol
		defer pubSub.Close()
		_, err := pubSub.Receive(context.Background())
		if err != nil {
			fmt.Println("无法从控件 PubSub 接收")
			return
		}
		ch := pubSub.Channel()
		for chMsg := range ch {
			var content GetSpotsMessage
			json.Unmarshal([]byte(chMsg.Payload), &content)
			// TODO (ZADD stock_order_交易所id_买卖方式_是否买涨_订单类型  价格 订单ID)
			//tick := content.Content.Tick
			//lastSize, _ := strconv.ParseFloat(tick.LastSize, 64)
			//lastSizeMax := int64((lastSize + global.MatchingRange) * global.PriceEnlargeMultiples)
			//lastSizeMin := int64((lastSize - global.MatchingRange) * global.PriceEnlargeMultiples)

			//获取符合范围内的有序集合成员 ZRANGE 递增排序  ZREVRANGE 递减排序  止盈止损取交集(在监控对应资金账户的盈亏逻辑中，进行执行写入队列处理)
			//组装key
			//1、现货1-买入1-买涨1-市价2  现货1-卖出2-买涨1-市价2  现货1-买入1-买涨1-限价1  现货1-买入1-买涨1-市价2
			//"stock_order_1_1_1_2     stock_order_1_2_1_2  stock_order_1_1_1_1   stock_order_1_1_1_2"
			//result1, _ := redisDb.ZRevRange(context.Background(), "stock_order_1_1_1_2", lastSizeMin, lastSizeMax).Result()
			//global.ZRevRangeQueue(redisDb, result1, lastSize)
			//
			//result2, _ := redisDb.ZRange(context.Background(), "stock_order_1_2_1_2", lastSizeMin, lastSizeMax).Result()
			//global.ZRevRangeQueue(redisDb, result2, lastSize)
			//
			//result3, _ := redisDb.ZRevRange(context.Background(), "stock_order_1_1_1_1", lastSizeMin, lastSizeMax).Result()
			//global.ZRevRangeQueue(redisDb, result3, lastSize)
			//
			//result4, _ := redisDb.ZRange(context.Background(), "stock_order_1_1_1_2", lastSizeMin, lastSizeMax).Result()
			//global.ZRevRangeQueue(redisDb, result4, lastSize)

			//2、合约2-买入1-止损止盈3-买涨1  合约2-卖出2-止损止盈3-买涨1  合约2-买入1-止损止盈3-买跌2  合约2-卖出2-止损止盈3-买跌2
			//"stock_order_2_1_3_1        stock_order_2_2_3_1      stock_order_2_1_3_2      stock_order_2_2_3_2"

			//3、美股4-买入1-止损止盈3-买涨1  美股4-卖出2-止损止盈3-买涨1  美股4-买入1-止损止盈3-买跌2  美股4-卖出2-止损止盈3-买跌2 美股4-买入1-买涨1-市价2  美股4-卖出2-买涨1-市价2  美股4-买入2-买涨1-限价1  美股4-买入1-买涨1-市价2
			//"stock_order_4_1_3_1        stock_order_4_2_3_1      stock_order_4_1_3_2      stock_order_4_2_3_2     stock_order_4_1_1_2   stock_order_4_2_1_2   stock_order_4_2_1_1  stock_order_4_1_1_2"
			//result5, _ := redisDb.ZRange(context.Background(), "stock_order_4_1_1_2", lastSizeMin, lastSizeMax).Result()
			//result6, _ := redisDb.ZRange(context.Background(), "stock_order_4_2_1_2", lastSizeMin, lastSizeMax).Result()
			//result7, _ := redisDb.ZRange(context.Background(), "stock_order_4_2_1_1", lastSizeMin, lastSizeMax).Result()
			//result8, _ := redisDb.ZRange(context.Background(), "stock_order_4_1_1_2", lastSizeMin, lastSizeMax).Result()

		}
	}
	return
}

// AccountSwapsSubscribe 获取合约行情
func AccountSwapsSubscribe() {
	subscribeSwapsMap := make(map[int]string, 0)
	subscribeSwapsMap[1] = "market.BTC-USDT.detail"
	subscribeSwapsMap[2] = "market.ETH-USDT.detail"
	subscribeSwapsMap[3] = "market.LTC-USDT.detail"
	subscribeSwapsMap[4] = "market.DOGE-USDT.detail"
	subscribeSwapsMap[5] = "market.SHIB-USDT.detail"
	subscribeSwapsMap[6] = "market.XRP-USDT.detail"
	subscribeSwapsMap[7] = "market.LINK-USDT.detail"
	subscribeSwapsMap[8] = "market.TRX-USDT.detail"
	subscribeSwapsMap[9] = "market.DOT-USDT.detail"
	subscribeSwapsMap[10] = "market.ADA-USDT.detail"
	subscribeSwapsMap[11] = "market.EOS-USDT.detail"
	subscribeSwapsMap[12] = "market.BCH-USDT.detail"

	for _, subscribe := range subscribeSwapsMap {
		pubSub := redisDb.Subscribe(context.Background(), subscribe) //msg.Symbol
		defer pubSub.Close()
		_, err := pubSub.Receive(context.Background())
		if err != nil {
			fmt.Println("无法从控件 PubSub 接收")
			return
		}
		ch := pubSub.Channel()
		for chMsg := range ch {
			var content GetSwapsMessage
			json.Unmarshal([]byte(chMsg.Payload), &content)

		}
	}
	return
}

func main() {
	//撮合交易引擎启动
	fmt.Println("The matching transaction engine starts ...")
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("系统内部错误")
			//重新拉起服务
			go AccountSpotsSubscribe() //现货
			go AccountSwapsSubscribe() //合约
		}
	}()

	//1、拉取行情（现货、合约、美股、马股）--- 2、获取订单集合(买集合、卖集合) ---- 3、匹配价 ---- 4、修改数据状态 ---- 5、推送消息
	go AccountSpotsSubscribe() //现货
	go AccountSwapsSubscribe() //合约

	fmt.Println("The matching transaction engine end ...") //撮合交易引擎启动
}
