package global

import (
	"Bourse/app/middleware/auth"
	"Bourse/common/inits"
	"Bourse/config/constant"
	"context"
	"encoding/json"
	"fmt"
	"gitee.com/phper95/pkg/logger"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	rabbitmq = inits.NewRabbitMQSimple("tradeOrderQueue_")
)

const (
	PriceEnlargeMultiples            float64 = 100000                               //价格放大倍数
	MatchingRange                    float64 = 2                                    //浮动范围
	TradeOrderQueueKey                       = "tradeOrderQueue_"                   //撮合订单队列名
	TradeorderqueueCloseout                  = "tradeOrderQueue_CloseOut"           //平仓、强平、止损止盈处理队列名
	TradeOrderQueueCloseoutHSetKey           = "tradeOrderQueue_CloseOut_HSetKey"   //平仓、强平、止损止盈处理队列通讯数据zset持久化key
	TradeorderqueueLimitprice                = "tradeOrderQueue_LimitPrice"         //撮合限价下单处理队列名
	TradeOrderQueueLimitpriceHSetKey         = "tradeOrderQueue_Limitprice_HSetKey" //撮合限价下单处理队列通讯数据zset持久化key
)

// TradeOrderHoldingsMQMsg 订单交易/持仓平的MQ生产-消费者的消息通讯结构体
type TradeOrderHoldingsMQMsg struct {
	QueueHSetKey    string  `json:"queueHSetKey"`    //处理队列通讯数据zset持久化key
	TimeNow         string  `json:"timeNow"`         //时间戳
	TradeOrderId    int64   `json:"tradeOrderId"`    //订单id
	StockHoldingsId int64   `json:"stockHoldingsId"` //持仓数组id
	DealPrice       float64 `json:"dealPrice"`       //成交价
}

func GetUserInfo(client *redis.Client) *constant.UserInfoStruct {
	//获取userInfo缓存
	userInfoUidKey := fmt.Sprintf("%s%d", constant.UserInfoUidKey, auth.TokenUserInfo.Id)
	result, err := client.Get(context.Background(), userInfoUidKey).Result()
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed:[Redis Get %s fail] Time:[%s] Err:[%v]", userInfoUidKey, time.Now().Format(time.DateTime), err))
		return nil
	}
	var UserInfo *constant.UserInfoStruct
	err = json.Unmarshal([]byte(result), &UserInfo)
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed:[%s] Time:[%s] Err:[%v]", "json解析失败!", time.Now().Format(time.DateTime), err))
		return nil
	}

	return UserInfo
}

type ZsetStockOrderStruct struct {
	UserId          int64   `json:"userId,omitempty"` //用户ID
	Uid             string  `json:"uid,omitempty"`    //用户uid
	TradeOrderId    int64   `json:"tradeOrderId"`     //交易订单id
	SystemBoursesId int64   `json:"systemBoursesId"`  //交易所股种id：1 现货 2 合约 3 马股 4 美股 5 日股
	Symbol          string  `json:"symbol"`           //交易代码
	StockSymbol     string  `json:"stockSymbol"`      //股票代码
	TradeType       int32   `json:"tradeType"`        //操作类型：1买入   2 平仓(卖出，必须持仓有这个订单)
	IsUpsDowns      int32   `json:"isUpsDowns"`       //交易方向：1 买涨(买入)   2 买跌(卖出)
	OrderType       int32   `json:"orderType"`        //交易方式：1 限价(需走撮合)   2 市价(下单直接成交)
	LimitPrice      float64 `json:"limitPrice"`       //限价
	MarketPrice     float64 `json:"marketPrice"`      //市价
	StopPrice       float64 `json:"stopPrice"`        //止损价(设置)
	TakeProfitPrice float64 `json:"takeProfitPrice"`  //止盈价(设置)
	DealPrice       float64 `json:"dealPrice"`        //成交价
	ClientConn      *redis.Client
}

// ZAddStockOrder 写入有序集合价格 用于撮合订单
func ZAddStockOrder(zsetData *ZsetStockOrderStruct) bool {
	// TODO 写入有序集合价格 /止损止盈集合 。。。 (ZADD stock_order_交易所id_买卖方式_是否买涨_订单类型  价格 订单ID)
	tradeOrdersKey := fmt.Sprintf("stock_order_%v_%v_%v_%v", zsetData.SystemBoursesId, zsetData.TradeType, zsetData.IsUpsDowns, zsetData.OrderType)
	switch zsetData.OrderType {
	case 1: //限价
		orderMoney := zsetData.LimitPrice * PriceEnlargeMultiples
		err := zsetData.ClientConn.ZAdd(context.Background(), tradeOrdersKey, redis.Z{Score: orderMoney, Member: zsetData.TradeOrderId}).Err()
		if err != nil {
			return false
		}
	case 2: //市价
		orderMoney := zsetData.MarketPrice * PriceEnlargeMultiples
		err := zsetData.ClientConn.ZAdd(context.Background(), tradeOrdersKey, redis.Z{Score: orderMoney, Member: zsetData.TradeOrderId}).Err()
		if err != nil {
			return false
		}
	case 3: //止损止盈  缓存key追加loss/profit区分,key带上用户uid (ZADD stock_order_用户uid_交易所id_买卖方式_是否买涨_订单类型  价格 订单ID)
		tradeOrdersKey2 := fmt.Sprintf("stock_order_%v_%v_%v_%v_%v", zsetData.Uid, zsetData.SystemBoursesId, zsetData.TradeType, zsetData.IsUpsDowns, zsetData.OrderType)
		tradeOrdersLossKey := fmt.Sprintf("%s_%s", tradeOrdersKey2, "loss")
		tradeOrdersProfitKey := fmt.Sprintf("%s_%s", tradeOrdersKey2, "profit")
		//止损
		err := zsetData.ClientConn.ZAdd(context.Background(), tradeOrdersLossKey, redis.Z{Score: zsetData.StopPrice * PriceEnlargeMultiples, Member: zsetData.TradeOrderId}).Err()
		if err != nil {
			return false
		}
		//止盈
		err = zsetData.ClientConn.ZAdd(context.Background(), tradeOrdersProfitKey, redis.Z{Score: zsetData.TakeProfitPrice * PriceEnlargeMultiples, Member: zsetData.TradeOrderId}).Err()
		if err != nil {
			return false
		}
	}
	return true
}

// ZremStockOrder 从有序集合移除 用于撮合订单
func ZremStockOrder(zsetData *ZsetStockOrderStruct) bool {
	// TODO 移除有序集合价格 /止损止盈集合 。。。 (ZREM stock_order_交易所id_买卖方式_是否买涨_订单类型  价格 订单ID)
	tradeOrdersKey := fmt.Sprintf("stock_order_%v_%v_%v_%v", zsetData.SystemBoursesId, zsetData.TradeType, zsetData.IsUpsDowns, zsetData.OrderType)
	switch zsetData.OrderType {
	case 1: //限价
		err := zsetData.ClientConn.ZRem(context.Background(), tradeOrdersKey, zsetData.TradeOrderId).Err()
		if err != nil {
			return false
		}
	case 2: //市价
		err := zsetData.ClientConn.ZRem(context.Background(), tradeOrdersKey, zsetData.TradeOrderId).Err()
		if err != nil {
			return false
		}
	case 3: //止损止盈  缓存key追加loss/profit区分
		tradeOrdersLossKey := fmt.Sprintf("%s_%s", tradeOrdersKey, "loss")
		tradeOrdersProfitKey := fmt.Sprintf("%s_%s", tradeOrdersKey, "profit")
		//止损
		err := zsetData.ClientConn.ZRem(context.Background(), tradeOrdersLossKey, zsetData.TradeOrderId).Err()
		if err != nil {
			return false
		}
		//止盈
		err = zsetData.ClientConn.ZRem(context.Background(), tradeOrdersProfitKey, zsetData.TradeOrderId).Err()
		if err != nil {
			return false
		}
	}
	return true
}

// ZRevRangeQueue 将撮合出来的订单，写入MQ队列
func ZRevRangeQueue(redisDb *redis.Client, result []string, dealPrice float64) {
	if len(result) > 0 {
		for _, s := range result {
			//数据持久化到Redis有序集合中 key=tradeOrderMQ 时间戳 订单id
			score := float64(time.Now().Unix())
			err := redisDb.ZAdd(context.Background(), TradeOrderQueueKey, redis.Z{
				Score:  score,
				Member: s,
			}).Err()
			if err != nil { //重试
				time.Sleep(2 * time.Second)
				redisDb.ZAdd(context.Background(), TradeOrderQueueKey, redis.Z{
					Score:  score,
					Member: s,
				}).Err()
			}
			tradeOrderMQMsg := TradeOrderHoldingsMQMsg{
				QueueHSetKey: TradeOrderQueueKey,
				TimeNow:      "score",
				TradeOrderId: 111111111,
				DealPrice:    dealPrice,
			}
			//5.2 (写入MQ队列)处理交易订单表
			marshal, _ := json.Marshal(&tradeOrderMQMsg)
			//发送队列消息
			rabbitmq.PublishSimple(string(marshal))
		}
	}
	return
}
