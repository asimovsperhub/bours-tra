package socket

import (
	"Bourse/app/middleware/auth"
	"Bourse/app/models"
	"Bourse/common/global"
	"Bourse/common/inits"
	"Bourse/config/constant"
	"context"
	"encoding/json"
	"fmt"
	"gitee.com/phper95/pkg/logger"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	MsServer           = NewServer()
	rabbitmqCloseOut   = inits.NewRabbitMQSimple(global.TradeorderqueueCloseout)   //平仓、强平、止损止盈处理队列名
	rabbitmqLimitPrice = inits.NewRabbitMQSimple(global.TradeorderqueueLimitprice) //撮合限价下单处理队列名
	redisDb            = inits.InitRedis()
	//mysqlDb            = inits.InitMysql()
)

type Server struct {
	mutex      *sync.Mutex        //互斥锁
	Clients    map[string]*Client //client信息
	Broadcast  chan *Message      //广播
	Register   chan *Client       //注册client
	Unregister chan *Client       //注销注册client
}

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
	Id                 int64   `json:"id"`                 //持仓股id
	SystemBoursesId    int64   `json:"systemBoursesId"`    //交易所股种id：1 现货 2 合约 3 马股 4 美股 5 日股
	IsUpsDowns         int32   `json:"IsUpsDowns"`         //交易方向：1 买涨(买入)   2 买跌(卖出)
	OrderType          int32   `json:"orderType"`          //交易方式：1 限价(需走撮合)   2 市价(下单直接成交)
	StockHoldingNum    int32   `json:"stockHoldingNum"`    //持仓数量
	LimitPrice         float64 `json:"limitPrice"`         //限价
	MarketPrice        float64 `json:"marketPrice"`        //市价
	NewMarketPrice     float64 `json:"NewMarketPrice"`     //最新市价
	BuyingPrice        float64 `json:"buyingPrice"`        //买入价
	OrderSumMoney      float64 `json:"orderSumMoney"`      //订单总金额
	ProfitLossMoney    float64 `json:"profitLossMoney"`    //订单盈亏
	ProfitLossSumMoney float64 `json:"ProfitLossSumMoney"` //总浮动盈亏
	//dayProfitLossSumMoney float64 `json:"dayProfitLossSumMoney"` //今日盈亏
	SumAssets       float64 `json:"sumAssets"`       //总资产
	AvailableAssets float64 `json:"availableAssets"` //可用资产
}

/*// MessageQuotesShareUs 美股消息
type MessageQuotesShareUs struct {
	S  string    `json:"s,omitempty"`  // 股票代码
	P  float64   `json:"p,omitempty"`  // 价格
	C  []float64 `json:"c,omitempty"`  // 条件，有关更多信息，请参阅贸易条件术语表
	V  int64     `json:"v,omitempty"`  // 交易量，代表在相应时间戳处交易的股票数量
	Dp bool      `json:"dp,omitempty"` // 暗池真/假
	Ms string    `json:"ms,omitempty"` // 市场状态，指示股票市场的当前状态（“开盘”、“收盘”、“延长交易时间”）
	T  int64     `json:"t,omitempty"`  // 以毫秒为单位的时间戳
	Av int64     `json:"av,omitempty"` // 今天累计交易量
	Op float64   `json:"op,omitempty"` // 今天正式开盘价格
	Vw float64   `json:"vw,omitempty"` // 即时报价的成交量加权平均价格
	Cl float64   `json:"cl,omitempty"` // 此聚合窗口的收盘价
	H  float64   `json:"h,omitempty"`  // 此聚合窗口的最高逐笔报价
	L  float64   `json:"l,omitempty"`  // 此聚合窗口的最低价格变动价格
	A  float64   `json:"a,omitempty"`  // 今天的成交量加权平均价格
	Z  int64     `json:"z,omitempty"`  // 此聚合窗口的平均交易规模
	Se int64     `json:"se,omitempty"` // 此聚合窗口的起始报价的时间戳（以 Unix 毫秒为单位）
}*/

// MessageQuotesShareUs 美股消息
type MessageQuotesShareUs struct {
	S  string   `json:"s,omitempty"`  // 股票代码
	P  string   `json:"p,omitempty"`  // 价格
	C  []string `json:"c,omitempty"`  // 条件，有关更多信息，请参阅贸易条件术语表
	V  int      `json:"v,omitempty"`  // 交易量，代表在相应时间戳处交易的股票数量
	Dp bool     `json:"dp,omitempty"` // 暗池真/假
	Ms string   `json:"ms,omitempty"` // 市场状态，指示股票市场的当前状态（“开盘”、“收盘”、“延长交易时间”）
	T  int64    `json:"t,omitempty"`  // 以毫秒为单位的时间戳
	Av int      `json:"av,omitempty"` // 今天累计交易量
	Op string   `json:"op,omitempty"` // 今天正式开盘价格
	Vw string   `json:"vw,omitempty"` // 即时报价的成交量加权平均价格
	Cl string   `json:"cl,omitempty"` // 此聚合窗口的收盘价
	H  string   `json:"h,omitempty"`  // 此聚合窗口的最高逐笔报价
	L  string   `json:"l,omitempty"`  // 此聚合窗口的最低价格变动价格
	A  string   `json:"a,omitempty"`  // 今天的成交量加权平均价格
	Z  int      `json:"z,omitempty"`  // 此聚合窗口的平均交易规模
	Se int64    `json:"se,omitempty"` // 此聚合窗口的起始报价的时间戳（以 Unix 毫秒为单位）
}

func NewServer() *Server {
	return &Server{
		mutex:      &sync.Mutex{},
		Clients:    make(map[string]*Client),
		Broadcast:  make(chan *Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// Start 协程处理管道消息
func (s *Server) Start() {
	fmt.Println("start server ...")
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("系统内部错误")
		}
	}()
	for {
		select {
		case conn, ok := <-s.Register: //处理注册client
			if !ok {
				break
			}
			fmt.Printf("用户登录，注册client Uid:%v ", conn.Uid)
			s.Clients[conn.Uid] = conn //用用户uid做通讯录
			//发送建立连接成功消息
			conn.Send <- &Message{
				Content: fmt.Sprintf("建立连接成功 Uid:%v ", conn.Uid),
				Uid:     auth.TokenUserInfo.Uid,
			}
			break
		case conn, ok := <-s.Unregister:
			if !ok {
				break
			}
			fmt.Printf("用户断开连接，注销client Uid:%v ", conn.Uid)
			//判断用户client是否存在通讯录中（TODO 改用Redis存 ...）
			//s.mutex.Lock()
			if _, ok := s.Clients[conn.Uid]; ok {
				close(conn.Send)
				delete(s.Clients, conn.Uid)
			}
			//s.mutex.Unlock()
			break
		case message, ok := <-s.Broadcast:
			if !ok {
				break
			}
			//s.mutex.Lock()
			client, ok := s.Clients[message.Uid]
			if ok {
				client.Send <- message
			}
			break
			//s.mutex.Unlock()
		}
	}
}

// PushAccountSpotsSubscribe 根据行情订阅，推送所有在线用户的现货资金账户变动
func (s *Server) PushAccountSpotsSubscribe() {
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

			content2 := content

			//给在线用户推送资金账户变动消息
			for _, client := range s.Clients {
				//TODO 浮动盈亏&累计盈亏 。。。
				//计算资金账户余额变动
				str := strings.Split(content.ServersId, ".")[1]
				userFundAccountSpotsKey := fmt.Sprintf("user_fund_account_spots_%v", client.UserId)
				spotMoney, _ := redisDb.HGet(context.Background(), userFundAccountSpotsKey, str).Result()

				//修改接口标识
				content.ServersId = "account.spots." + content.ServersId
				content.Content.Ch = "account.spots." + content.Content.Ch

				//计算实时价
				floatLastSize, _ := strconv.ParseFloat(content.Content.Tick.LastSize, 64)
				floatSpotMoney, _ := strconv.ParseFloat(spotMoney, 64)
				money := (floatLastSize - floatSpotMoney) + floatSpotMoney
				if floatSpotMoney == 0 {
					money = 0
				}
				content.Content.Tick.LastSize = fmt.Sprintf("%f", money)
				sendMsg := &Message{
					ServersId: content.ServersId,
					Content:   content.Content,
					Symbol:    content.Symbol,
					Uid:       client.Uid,
				}

				//将新余额写入Redis
				redisDb.HSet(context.Background(), userFundAccountSpotsKey, client.UserId, money).Err()

				//发送客户端
				MsServer.Broadcast <- sendMsg

				// ================= 计算主账户总资产 ========================

				//推送资金账户余额--现货
				userFundAccountSpotsNumKey := fmt.Sprintf("user_fund_account_1_%v", client.UserId)

				//取改用户所有账户金额
				hValsArr, _ := redisDb.HVals(context.Background(), userFundAccountSpotsNumKey).Result()
				var spotsMoneyNum float64
				for _, hValsStr := range hValsArr {
					float, _ := strconv.ParseFloat(hValsStr, 64)
					spotsMoneyNum += float
				}

				//修改接口标识
				content2.ServersId = "account.spots.num.hvals"
				content2.Content.Ch = "account.spots.num.hvals"

				moneyNum := (floatLastSize - spotsMoneyNum) + spotsMoneyNum
				//计算实时价
				if spotsMoneyNum == 0 {
					moneyNum = 0
				}

				content2.Content.Tick.Close = fmt.Sprintf("%f", moneyNum)
				sendNumMsg := &Message{
					ServersId: content2.ServersId,
					Content:   content2.Content,
					Symbol:    content2.Symbol,
					Uid:       client.Uid,
				}

				MsServer.Broadcast <- sendNumMsg
			}
		}
	}
	return
}

// PushAccountSwapsSubscribe 根据行情订阅，推送所有在线用户的合约资金账户变动
func (s *Server) PushAccountSwapsSubscribe() {
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
			content2 := content

			//给在线用户推送资金账户变动消息
			for _, client := range s.Clients {
				//计算资金账户余额变动
				str := strings.Split(content.ServersId, ".")[1]
				userFundAccountSwapsKey := fmt.Sprintf("user_fund_account_swaps_%v", client.UserId)
				swapMoney, _ := redisDb.HGet(context.Background(), userFundAccountSwapsKey, str).Result()

				//修改接口标识
				content.ServersId = "account.swaps." + content.ServersId
				content.Content.Ch = "account.swaps." + content.Content.Ch

				//计算实时价
				floatClose, _ := strconv.ParseFloat(content.Content.Tick.Close, 64)
				floatSwapMoney, _ := strconv.ParseFloat(swapMoney, 64)
				money := (floatClose - floatSwapMoney) + floatSwapMoney
				if floatSwapMoney == 0 {
					money = 0
				}
				content.Content.Tick.Close = fmt.Sprintf("%f", money)
				sendMsg := &Message{
					ServersId: content.ServersId,
					Content:   content.Content,
					Symbol:    content.Symbol,
					Uid:       client.Uid,
				}

				//将新余额写入Redis
				redisDb.HSet(context.Background(), userFundAccountSwapsKey, client.UserId, money).Err()

				//发送客户端
				MsServer.Broadcast <- sendMsg

				//================= 计算主账户总资产 =====================

				//推送资金账户余额--现货
				userFundAccountSwapsNumKey := fmt.Sprintf("user_fund_account_2_%v", client.UserId)

				//取改用户所有账户金额
				hValsArr, _ := redisDb.HVals(context.Background(), userFundAccountSwapsNumKey).Result()
				var swapMoneyNum float64
				for _, hValsStr := range hValsArr {
					float, _ := strconv.ParseFloat(hValsStr, 64)
					swapMoneyNum += float
				}

				//修改接口标识
				content2.ServersId = "account.swaps.num.hvals"
				content2.Content.Ch = "account.swaps.num.hvals"

				//计算实时价
				moneyNum := (floatClose - swapMoneyNum) + swapMoneyNum
				if swapMoneyNum == 0 {
					moneyNum = 0
				}
				content.Content.Tick.Close = fmt.Sprintf("%f", moneyNum)
				sendNumMsg := &Message{
					ServersId: content2.ServersId,
					Content:   content2.Content,
					Symbol:    content2.Symbol,
					Uid:       client.Uid,
				}

				MsServer.Broadcast <- sendNumMsg
			}

		}
	}
	return
}

// UserStockHoldingsPush 用户持仓数据推送
func (s *Server) UserStockHoldingsPush() {
	//给在线用户推送资金账户变动消息
	/*for _, client := range s.Clients {
		time.Sleep(2 * time.Second) //延时2秒
		//获取用户持仓缓存数据
		go s.GetUserStockholdingsRedisData(client, "1", "account.spots.stock_holdings_data") //现货
		go s.GetUserStockholdingsRedisData(client, "2", "account.spots.swap_holdings_data")  //合约
	}*/
	return
}

// GetUserStockholdingsRedisData 获取用户持仓缓存数据 TODO 废弃 。。。
func (s *Server) GetUserStockholdingsRedisData(client *Client, systemBoursesId string, serversId string) {
	//组装缓存key
	cacheKey := "user_stock_holdings_data_" + client.Uid + systemBoursesId
	//返回key数据
	hValsArr, _ := redisDb.HVals(context.Background(), cacheKey).Result()
	if len(hValsArr) > 0 {
		for _, hValsStr := range hValsArr {
			var stockHoldingsJson StockHoldingsJson
			json.Unmarshal([]byte(hValsStr), stockHoldingsJson)
			//TODO 计算价格 盈和亏怎么计算。。。

			//2、合约2-买入1-止损止盈3-买涨1  合约2-卖出2-止损止盈3-买涨1  合约2-买入1-止损止盈3-买跌2  合约2-卖出2-止损止盈3-买跌2
			//"stock_order_2_1_3_1        stock_order_2_2_3_1      stock_order_2_1_3_2      stock_order_2_2_3_2"

			/*result1, _ := redisDb.ZRevRange(context.Background(), "stock_order_2_1_3_1", lastSizeMin, lastSizeMax).Result()
			global.ZRevRangeQueue(redisDb, result1)

			result2, _ := redisDb.ZRange(context.Background(), "stock_order_2_2_3_1", lastSizeMin, lastSizeMax).Result()
			global.ZRevRangeQueue(redisDb, result2)

			result3, _ := redisDb.ZRevRange(context.Background(), "stock_order_2_1_3_2", lastSizeMin, lastSizeMax).Result()
			global.ZRevRangeQueue(redisDb, result3)

			result4, _ := redisDb.ZRange(context.Background(), "stock_order_2_2_3_2", lastSizeMin, lastSizeMax).Result()
			global.ZRevRangeQueue(redisDb, result4)*/

			//发送消息
			sendMsg := &Message{
				ServersId: serversId,
				Content:   stockHoldingsJson,
				Uid:       client.Uid,
			}
			MsServer.Broadcast <- sendMsg
		}
	}
	return
}

// PushStocksUS 根据美股行情，推送所有在线用户的美股资金账户变动
func (s *Server) PushStocksUS() {
	//连接美股行情ws
	conn := s.NewQuotesShareWssUS()

	//拉取持仓数据，订阅
	go s.GetUserStockHoldingsList(conn)

	//读ws
	for {
		time.Sleep(time.Second * 1)
		go s.ReadMessageQuotesShareWssUS(conn)
	}

	//写ws
	//s.WriteMessageQuotesShareWssUS(conn)
	return
}

// NewQuotesShareWssUS 建立连接美股行情
func (s *Server) NewQuotesShareWssUS() *websocket.Conn {
	//连接美股行情ws
	url := "ws://34.126.106.7:8855/quotes-share-wss"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		logger.Warn(fmt.Sprintf("Err:链接wss服务器失败 Time:[%s] Err:[%v]", time.Now().Format(time.DateTime), err.Error()))
		return nil
	}
	return conn
}

// ReadMessageQuotesShareWssUS 读取ws美股行情
func (s *Server) ReadMessageQuotesShareWssUS(conn *websocket.Conn) {
	_, msg, err := conn.ReadMessage()
	if err != nil {
		logger.Warn(fmt.Sprintf("Err:US行情读取失败 Time:[%s] Err:[%v]", time.Now().Format(time.DateTime), err.Error()))
		return
	}

	//fmt.Printf("========= %+v =========", string(msg))

	var msgUs MessageQuotesShareUs

	/*msgUs := MessageQuotesShareUs{
		S:  "AAA",
		P:  "26.00",
		C:  nil,
		V:  0,
		Dp: false,
		Ms: "",
		T:  0,
		Av: 0,
		Op: "",
		Vw: "",
		Cl: "",
		H:  "",
		L:  "",
		A:  "",
		Z:  0,
		Se: 0,
	}*/

	//解析消息
	json.Unmarshal(msg, &msgUs)
	pFloat64, _ := strconv.ParseFloat(msgUs.P, 64)

	//fmt.Printf("=========最新价 %v =========", msgUs.P)

	//给在线用户推送美股变动消息
	for _, client := range s.Clients {
		//获取用户持仓列表 TODO redis的哈希key=user_stock_holdings_data_+ client.Uid + systemBoursesId ...
		cacheKey := fmt.Sprintf("user_stock_holdings_data_%d_4", client.UserId)
		//返回key数据
		HVals, _ := redisDb.HVals(context.Background(), cacheKey).Result()

		var stockHoldingsJsonMap []*StockHoldingsJson
		var msgStockHoldingsJson StockHoldingsJson
		var profitLossSumMoney float64 //总浮动盈亏
		var orderSumMoney float64      //订单总金额
		if len(HVals) > 0 {
			// ================ 推送持仓数据 ==================
			for _, HValsStr := range HVals {
				var stockHoldingsJson models.UserStockHoldingsJson
				json.Unmarshal([]byte(HValsStr), &stockHoldingsJson)

				if stockHoldingsJson.StockSymbol != msgUs.S || stockHoldingsJson.Id == 0 { //只处理对应股的数据
					//fmt.Printf(" ==== %s ------ %s ======= ", stockHoldingsJson.StockSymbol, msgUs.S)
					continue
				}
				var profitLossMoney float64                                              //浮动盈亏
				msgStockHoldingsJson.Id = stockHoldingsJson.Id                           //持仓股id
				msgStockHoldingsJson.SystemBoursesId = stockHoldingsJson.SystemBoursesId //交易所股种id：1 现货 2 合约 3 马股 4 美股 5 日股
				msgStockHoldingsJson.IsUpsDowns = stockHoldingsJson.IsUpsDowns           //是否买涨跌：1 买涨 2 买跌
				msgStockHoldingsJson.OrderType = stockHoldingsJson.OrderType             //交易方式：1 限价(需走撮合)   2 市价(下单直接成交)
				msgStockHoldingsJson.StockHoldingNum = stockHoldingsJson.StockHoldingNum //持仓数量
				msgStockHoldingsJson.LimitPrice = stockHoldingsJson.LimitPrice           //限价
				msgStockHoldingsJson.MarketPrice = stockHoldingsJson.MarketPrice         //市价
				msgStockHoldingsJson.BuyingPrice = stockHoldingsJson.BuyingPrice         //开仓价(买入价)
				msgStockHoldingsJson.NewMarketPrice = pFloat64                           //最新市价
				orderSumMoney = orderSumMoney + stockHoldingsJson.OrderAmount            //计算订单总金额 TODO 计算错误....

				// ============ 盈亏计算 ================
				var stockHoldingNumF64 float64
				if stockHoldingsJson.IsUpsDowns == 1 {
					//计算盈亏：股票平仓盈利=(平仓价-开仓价)*数量 (这里按市价计算)
					stockHoldingNumF64, _ = strconv.ParseFloat(fmt.Sprintf("%d", stockHoldingsJson.StockHoldingNum), 64)
					profitLossMoney = (pFloat64 - stockHoldingsJson.PositionPrice) * stockHoldingNumF64
				} else if stockHoldingsJson.IsUpsDowns == 2 {
					//计算盈亏：股票平仓盈利=(开仓价-平仓价)*数量 (这里按市价计算)
					stockHoldingNumF64, _ = strconv.ParseFloat(fmt.Sprintf("%d", stockHoldingsJson.StockHoldingNum), 64)
					profitLossMoney = (stockHoldingsJson.PositionPrice - pFloat64) * stockHoldingNumF64

				}
				profitLossSumMoney = profitLossSumMoney + profitLossMoney    //计算总浮动盈亏
				msgStockHoldingsJson.ProfitLossSumMoney = profitLossSumMoney //总的浮动盈亏
				msgStockHoldingsJson.MarketPrice = pFloat64                  //最新行情
				msgStockHoldingsJson.ProfitLossMoney = profitLossMoney       //订单盈亏
				msgStockHoldingsJson.OrderSumMoney = orderSumMoney           //订单总金额
				stockHoldingsJsonMap = append(stockHoldingsJsonMap, &msgStockHoldingsJson)

				// =============== 缓存用户盈亏&总盈亏 TODO 脚本写入数据库 。。。===========
				//缓存盈亏zset stockHoldings_profitLossMoneykey_+systemBoursesId (盈亏 msgStockHoldingsJson.Id)
				profitLossMoneykey := fmt.Sprintf("stockHoldings_profitLossMoneykey_%d", stockHoldingsJson.SystemBoursesId)
				err = redisDb.ZAdd(context.Background(), profitLossMoneykey, redis.Z{
					Score:  profitLossMoney,
					Member: stockHoldingsJson.Id,
				}).Err()
				if err != nil {
					logger.Warn("==== 缓存盈亏失败", zap.String(profitLossMoneykey, fmt.Sprintf("==== 缓存失败 Score:[%f] Member:[%d]", profitLossMoney, msgStockHoldingsJson.Id)))
					return
				}
				//缓存用户总盈亏zset stockHoldings_profitLossMoneykey_Count_+systemBoursesId (总盈亏 client.UserId)
				profitLossMoneyCountkey := fmt.Sprintf("stockHoldings_profitLossMoneykey_Count_%d", stockHoldingsJson.SystemBoursesId)
				err = redisDb.ZAdd(context.Background(), profitLossMoneyCountkey, redis.Z{
					Score:  profitLossSumMoney,
					Member: client.UserId,
				}).Err()
				if err != nil {
					logger.Warn("==== 缓存用户总盈亏失败", zap.String(profitLossMoneyCountkey, fmt.Sprintf("==== 缓存失败 Score:[%f] Member:[%d]", profitLossMoney, client.UserId)))
					return
				}

				// ====================== 发送消息 ---- 推送持仓数据 ================
				//组装发送消息结构体
				sendMessage1 := &Message{
					ServersId: "account.share.us.stock_holdings_data",
					Content:   msgStockHoldingsJson,
					Symbol:    msgUs.Ms,
					Uid:       client.Uid,
				}
				MsServer.Broadcast <- sendMessage1

				//================= 获取美股资金账户缓存   账户总资产、总资产、可用资产、订单总金额、总浮动盈亏=============
				userFundAccountsKey4 := fmt.Sprintf("%s4_%v", constant.UserFundAccountsKey, client.UserId)
				result, err := redisDb.Get(context.Background(), userFundAccountsKey4).Result()
				if err != nil {
					continue
				}
				var userFundAccountsJson models.UserFundAccountsJson
				json.Unmarshal([]byte(result), &userFundAccountsJson)
				msgStockHoldingsJson.SumAssets = userFundAccountsJson.Money + msgStockHoldingsJson.ProfitLossSumMoney + msgStockHoldingsJson.OrderSumMoney //计算用户总资产 = 账户金额 + 总浮动盈亏 + 总订单金额
				msgStockHoldingsJson.AvailableAssets = userFundAccountsJson.Money                                                                          //可用资产是不动的

				//发送消息 ---- 推送美股资金账户数据 account.swaps.num.hvals
				//组装发送消息结构体
				sendMessage := &Message{
					ServersId: "account.share.us.num.hvals",
					Content:   msgStockHoldingsJson,
					Symbol:    msgUs.Ms,
					Uid:       client.Uid,
				}
				MsServer.Broadcast <- sendMessage

				// ================= 监听止盈止损  =================
				queueCloseoutHSetKey := fmt.Sprintf("%s_%d", global.TradeOrderQueueCloseoutHSetKey, msgStockHoldingsJson.SystemBoursesId)
				//判断是否已在处理队列中
				dataStockHoldingsId, _ := redisDb.HGet(context.Background(), queueCloseoutHSetKey, fmt.Sprintf("%d", stockHoldingsJson.Id)).Result()
				fmt.Printf("=========dataStockHoldingsId=%s ======", dataStockHoldingsId)
				if len(dataStockHoldingsId) > 0 {
					continue
				}

				//买涨的情况下
				if stockHoldingsJson.IsUpsDowns == 1 && stockHoldingsJson.StopPrice >= pFloat64 { //止损
					//发送消息 ----止盈：盈亏>=止盈价 推送平仓数据，写入平仓队列
					//组装发送MQ消息结构体
					tradeOrderHoldingsMQMsg := global.TradeOrderHoldingsMQMsg{
						QueueHSetKey:    queueCloseoutHSetKey,
						TimeNow:         fmt.Sprintf("%d", time.Now().Unix()),
						StockHoldingsId: stockHoldingsJson.Id,
						DealPrice:       stockHoldingsJson.StopPrice,
					}
					marshal, _ := json.Marshal(tradeOrderHoldingsMQMsg)
					rabbitmqCloseOut.PublishSimple(string(marshal))

					//将消息记录进消费队列的zset持久化
					redisDb.HSet(context.Background(), queueCloseoutHSetKey, stockHoldingsJson.Id, marshal)
				} else if stockHoldingsJson.IsUpsDowns == 1 && stockHoldingsJson.TakeProfitPrice <= pFloat64 { //止盈
					//发送消息 ----止盈：盈亏>=止盈价 推送平仓数据，写入平仓队列
					//组装发送MQ消息结构体
					tradeOrderHoldingsMQMsg := global.TradeOrderHoldingsMQMsg{
						QueueHSetKey:    queueCloseoutHSetKey,
						TimeNow:         fmt.Sprintf("%d", time.Now().Unix()),
						StockHoldingsId: stockHoldingsJson.Id,
						DealPrice:       stockHoldingsJson.TakeProfitPrice,
					}
					marshal, _ := json.Marshal(tradeOrderHoldingsMQMsg)
					rabbitmqCloseOut.PublishSimple(string(marshal))

					//将消息记录进消费队列的zset持久化
					redisDb.HSet(context.Background(), queueCloseoutHSetKey, stockHoldingsJson.Id, marshal)
				}

				//买跌的情况下
				if stockHoldingsJson.IsUpsDowns == 2 && stockHoldingsJson.StopPrice <= pFloat64 { //止损
					//发送消息 ----止盈：盈亏>=止盈价 推送平仓数据，写入平仓队列
					//组装发送MQ消息结构体
					tradeOrderHoldingsMQMsg := global.TradeOrderHoldingsMQMsg{
						QueueHSetKey:    queueCloseoutHSetKey,
						TimeNow:         fmt.Sprintf("%d", time.Now().Unix()),
						StockHoldingsId: stockHoldingsJson.Id,
						DealPrice:       stockHoldingsJson.StopPrice,
					}
					marshal, _ := json.Marshal(tradeOrderHoldingsMQMsg)
					rabbitmqCloseOut.PublishSimple(string(marshal))

					//将消息记录进消费队列的zset持久化
					redisDb.HSet(context.Background(), queueCloseoutHSetKey, stockHoldingsJson.Id, marshal)
				} else if stockHoldingsJson.IsUpsDowns == 2 && stockHoldingsJson.TakeProfitPrice >= pFloat64 { //止盈
					//发送消息 ----止盈：盈亏>=止盈价 推送平仓数据，写入平仓队列
					//组装发送MQ消息结构体
					tradeOrderHoldingsMQMsg := global.TradeOrderHoldingsMQMsg{
						QueueHSetKey:    queueCloseoutHSetKey,
						TimeNow:         fmt.Sprintf("%d", time.Now().Unix()),
						StockHoldingsId: stockHoldingsJson.Id,
						DealPrice:       stockHoldingsJson.TakeProfitPrice,
					}
					marshal, _ := json.Marshal(tradeOrderHoldingsMQMsg)
					rabbitmqCloseOut.PublishSimple(string(marshal))

					//将消息记录进消费队列的zset持久化
					redisDb.HSet(context.Background(), queueCloseoutHSetKey, stockHoldingsJson.Id, marshal)
				}

				// ================= 监听强平 TODO 订单金额永远不会被亏没 ...======================
				/*if (stockHoldingsJson.OrderAmount - profitLossMoney) >= (stockHoldingsJson.OrderAmount * 0.2) {
					//发送消息 ---- 盈亏>=订单金额*0.8 推送强平数据，写入平仓队列
					//组装发送MQ消息结构体
					tradeOrderHoldingsMQMsg := global.TradeOrderHoldingsMQMsg{
						QueueHSetKey:    queueCloseoutHSetKey,
						TimeNow:         fmt.Sprintf("%d", time.Now().Unix()),
						StockHoldingsId: stockHoldingsJson.Id,
						DealPrice:       pFloat64,
					}
					marshal, _ := json.Marshal(tradeOrderHoldingsMQMsg)
					//组装发送消息结构体
					rabbitmqCloseOut.PublishSimple(string(marshal))

					//将消息记录进消费队列的zset持久化
					redisDb.HSet(context.Background(), queueCloseoutHSetKey, stockHoldingsJson.Id, marshal)
				}*/
			}
		}

		// ================= 监听限价交易撮合 =================
		//获取用户持仓列表 TODO redis的哈希key=user_stock_trade_orders_data_+ client.Uid + systemBoursesId ...
		cacheOrdersKey := fmt.Sprintf("user_stock_trade_orders_data_%d_4", client.UserId)
		//返回key数据
		HValsOrders, _ := redisDb.HVals(context.Background(), cacheOrdersKey).Result()
		if len(HValsOrders) > 0 {
			for _, orderHas := range HValsOrders {
				var userTradeOrdersJson models.UserTradeOrdersJson
				json.Unmarshal([]byte(orderHas), &userTradeOrdersJson)
				//fmt.Printf("======= %s ==orderHas== %+v =====", msgUs.S, orderHas)
				if userTradeOrdersJson.StockSymbol != msgUs.S { //只处理对应股的数据
					//fmt.Printf(" ==== %s ------ %s ======= ", stockHoldingsJson.StockSymbol, msgUs.S)
					continue
				}
				queueLimitpriceHSetKey := fmt.Sprintf("%s_%d", global.TradeOrderQueueLimitpriceHSetKey, userTradeOrdersJson.SystemBoursesId)
				//判断是否已在处理队列中
				if userTradeOrdersJson.Id > 0 {
					dataStockHoldingsId, _ := redisDb.HGet(context.Background(), queueLimitpriceHSetKey, fmt.Sprintf("%d", userTradeOrdersJson.Id)).Result()
					if len(dataStockHoldingsId) > 0 {
						continue
					}
				}

				if userTradeOrdersJson.IsUpsDowns == 1 && userTradeOrdersJson.LimitPrice >= pFloat64 { //买涨的情况
					//组装发送MQ消息结构体
					tradeOrderHoldingsMQMsg := global.TradeOrderHoldingsMQMsg{
						QueueHSetKey: queueLimitpriceHSetKey,
						TimeNow:      fmt.Sprintf("%d", time.Now().Unix()),
						TradeOrderId: userTradeOrdersJson.Id,
						DealPrice:    userTradeOrdersJson.LimitPrice, // TODO +0.02还是-0.02 ...
					}
					tradeOrderHoldingsMQMsg.TradeOrderId = userTradeOrdersJson.Id

					marshal, _ := json.Marshal(tradeOrderHoldingsMQMsg)
					rabbitmqLimitPrice.PublishSimple(string(marshal))
					fmt.Printf("  =====限价成交匹配中111111====queueLimitpriceHSetKey:[%s] ID:[%d] IsUpsDowns:[%d] LimitPrice:[%f] pFloat64:[%f] msg:[%s] ====  ", queueLimitpriceHSetKey, userTradeOrdersJson.Id, userTradeOrdersJson.IsUpsDowns, userTradeOrdersJson.LimitPrice, pFloat64, string(marshal))
					//将消息记录进消费队列的zset持久化
					err = redisDb.HSet(context.Background(), queueLimitpriceHSetKey, userTradeOrdersJson.Id, marshal).Err()
					if err != nil {
						logger.Warn(fmt.Sprintf("将消息记录进消费队列的zset持久化失败 userTradeOrdersJson.Id:[%d] Time:[%s] Err:[%v]", userTradeOrdersJson.Id, time.Now().Format(time.DateTime), err))
						break
					}
					//fmt.Println(454545454545)

				}
				if userTradeOrdersJson.IsUpsDowns == 2 && userTradeOrdersJson.LimitPrice <= pFloat64 { //买跌的情况
					fmt.Printf("  =====限价成交匹配中22222====queueLimitpriceHSetKey:[%s] ID:[%d] IsUpsDowns:[%d] LimitPrice:[%f] pFloat64:[%f] ====  ", queueLimitpriceHSetKey, userTradeOrdersJson.Id, userTradeOrdersJson.IsUpsDowns, userTradeOrdersJson.LimitPrice, pFloat64)
					//组装发送MQ消息结构体
					tradeOrderHoldingsMQMsg := global.TradeOrderHoldingsMQMsg{
						QueueHSetKey: queueLimitpriceHSetKey,
						TimeNow:      fmt.Sprintf("%d", time.Now().Unix()),
						TradeOrderId: userTradeOrdersJson.Id,
						DealPrice:    userTradeOrdersJson.LimitPrice, // TODO +0.02还是-0.02 ...
					}
					tradeOrderHoldingsMQMsg.TradeOrderId = userTradeOrdersJson.Id

					marshal, _ := json.Marshal(tradeOrderHoldingsMQMsg)
					rabbitmqLimitPrice.PublishSimple(string(marshal))
					//将消息记录进消费队列的zset持久化
					err = redisDb.HSet(context.Background(), queueLimitpriceHSetKey, userTradeOrdersJson.Id, marshal).Err()
					if err != nil {
						logger.Warn(fmt.Sprintf("将消息记录进消费队列的zset持久化失败 userTradeOrdersJson.Id:[%d] Time:[%s] Err:[%v]", userTradeOrdersJson.Id, time.Now().Format(time.DateTime), err))
						break
					}
					//fmt.Println(67676767676)
				}
			}
		}
	}
	return
}

// WriteMessageQuotesShareWssUS 发送ws美股行情
func (s *Server) WriteMessageQuotesShareWssUS(conn *websocket.Conn) {
	//获取用户美股持仓数据  redis订阅key=SubscribeQuotesShareWssUS
	pubSub := redisDb.Subscribe(context.Background(), "SubscribeQuotesShareWssUS") //msg.Symbol
	defer pubSub.Close()
	_, err := pubSub.Receive(context.Background())
	if err != nil {
		fmt.Println("无法从控件 PubSub 接收")
		return
	}
	ch := pubSub.Channel()
	for chMsg := range ch {
		//订阅行情股票
		w, err := conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write([]byte(chMsg.Payload))
		if err := w.Close(); err != nil {
			return
		}
	}
	return
}

// GetUserStockHoldingsList 获取全量用户持仓列表
func (s *Server) GetUserStockHoldingsList(conn *websocket.Conn) {
	mysqlDb := inits.InitMysql()
	defer mysqlDb.Close()
	sqlStr := "select distinct stockSymbol from bo_user_stock_holdings where stockHoldingType=1"
	rows, err := mysqlDb.Query(sqlStr)
	if err != nil {
		fmt.Println("获取持仓数据失败")
		return
	}
	defer rows.Close()
	var stockHoldings string
	for rows.Next() {
		rows.Scan(&stockHoldings)
		subscribe := fmt.Sprintf("{\"type\":\"subscribe\",\"symbol\":\"%v\"}", stockHoldings)
		w, err := conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write([]byte(subscribe))
		if err := w.Close(); err != nil {
			return
		}
	}
	//获取redis订阅消息
	pubSub := redisDb.Subscribe(context.Background(), "stockSymbol_subscribe_")
	defer func() {
		pubSub.Close()
	}()
	_, err = pubSub.Receive(context.Background())
	if err != nil {
		fmt.Println("获取订阅失败")
		return
	}

	ch := pubSub.Channel()
	for msg := range ch {
		subscribe := fmt.Sprintf("{\"type\":\"subscribe\",\"symbol\":\"%v\"}", msg.Payload)
		w, err := conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write([]byte(subscribe))
		if err := w.Close(); err != nil {
			return
		}
	}

	return
}
