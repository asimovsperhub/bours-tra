package models

import (
	"database/sql"
	"fmt"
	"time"
)

// UserTradeOrders 交易订单表
type UserTradeOrders struct {
	Id                 sql.NullInt64   `json:"id"`                 //交易订单id
	UserId             sql.NullInt64   `json:"userId"`             //用户id
	BourseType         sql.NullInt32   `json:"bourseType"`         //交易所类型：1 数字币 2 股票
	SystemBoursesId    sql.NullInt64   `json:"systemBoursesId"`    //交易所股种id：1 现货 2 合约 3 马股 4 美股 5 日股
	Symbol             sql.NullString  `json:"symbol"`             //交易代码
	StockSymbol        sql.NullString  `json:"stockSymbol"`        //股票代码
	StockName          sql.NullString  `json:"stockName"`          //股票名字
	TradeType          sql.NullInt32   `json:"tradeType"`          //操作类型：1买入   2 平仓(卖出，必须持仓有这个订单)
	IsUpsDowns         sql.NullInt32   `json:"isUpsDowns"`         //交易方向：1 买涨(买入)   2 买跌(卖出)
	OrderType          sql.NullInt32   `json:"orderType"`          //交易方式：1 限价(需走撮合)   2 市价(下单直接成交)
	EntrustNum         sql.NullInt32   `json:"entrustNum"`         //委托数量
	EntrustStatus      sql.NullInt32   `json:"entrustStatus"`      //委托状态：1 待成交 2  成功 3 撤单
	LimitPrice         sql.NullFloat64 `json:"limitPrice"`         //限价
	MarketPrice        sql.NullFloat64 `json:"marketPrice"`        //市价
	StopPrice          sql.NullFloat64 `json:"stopPrice"`          //止损价(设置)
	TakeProfitPrice    sql.NullFloat64 `json:"takeProfitPrice"`    //止盈价(设置)
	OrderAmount        sql.NullFloat64 `json:"orderAmount"`        //订单金额
	OrderAmountSum     sql.NullFloat64 `json:"orderAmountSum"`     //订单总金额(包含保证金或手续费)
	EarnestMoney       sql.NullFloat64 `json:"earnestMoney"`       //保证金
	HandlingFee        sql.NullFloat64 `json:"handlingFee"`        //手续费
	PositionPrice      sql.NullFloat64 `json:"positionPrice"`      //持仓价
	DealPrice          sql.NullFloat64 `json:"dealPrice"`          //成交价
	EntrustOrderNumber sql.NullString  `json:"entrustOrderNumber"` //委托订单号
	StockHoldingsId    sql.NullInt64   `json:"stockHoldingsId"`    //持仓id
	EntrustTime        sql.NullString  `json:"entrustTime"`        //委托时间
	UpdateTime         sql.NullString  `json:"updateTime"`         //更新时间（平仓时间）
}

// UserTradeOrdersJson 交易订单表
type UserTradeOrdersJson struct {
	Id                 int64   `json:"id"`                 //交易订单id
	UserId             int64   `json:"userId"`             //用户id
	BourseType         int32   `json:"bourseType"`         //交易所类型：1 数字币 2 股票
	SystemBoursesId    int64   `json:"systemBoursesId"`    //交易所股种id：1 现货 2 合约 3 马股 4 美股 5 日股
	Symbol             string  `json:"symbol"`             //交易代码
	StockSymbol        string  `json:"stockSymbol"`        //股票代码
	StockName          string  `json:"stockName"`          //股票名字
	TradeType          int32   `json:"tradeType"`          //操作类型：1买入   2 平仓(卖出，必须持仓有这个订单)
	IsUpsDowns         int32   `json:"isUpsDowns"`         //交易方向：1 买涨(买入)   2 买跌(卖出)
	OrderType          int32   `json:"orderType"`          //交易方式：1 限价(需走撮合)   2 市价(下单直接成交)
	EntrustNum         int32   `json:"entrustNum"`         //委托数量
	EntrustStatus      int32   `json:"entrustStatus"`      //委托状态：1 待成交 2  成功 3 撤单
	LimitPrice         float64 `json:"limitPrice"`         //限价
	MarketPrice        float64 `json:"marketPrice"`        //市价
	StopPrice          float64 `json:"stopPrice"`          //止损价(设置)
	TakeProfitPrice    float64 `json:"takeProfitPrice"`    //止盈价(设置)
	OrderAmount        float64 `json:"orderAmount"`        //订单金额
	OrderAmountSum     float64 `json:"orderAmountSum"`     //订单总金额(包含保证金或手续费)
	EarnestMoney       float64 `json:"earnestMoney"`       //保证金
	HandlingFee        float64 `json:"handlingFee"`        //手续费
	PositionPrice      float64 `json:"positionPrice"`      //持仓价
	DealPrice          float64 `json:"dealPrice"`          //成交价
	EntrustOrderNumber string  `json:"entrustOrderNumber"` //委托订单号
	StockHoldingsId    int64   `json:"stockHoldingsId"`    //持仓id
	EntrustTime        string  `json:"entrustTime"`        //委托时间
	UpdateTime         string  `json:"updateTime"`         //更新时间（平仓时间）
}

func ConvUserTradeOrdersNull2Json(v UserTradeOrders, p UserTradeOrdersJson) UserTradeOrdersJson {
	p.Id = v.Id.Int64
	p.UserId = v.UserId.Int64
	p.BourseType = v.BourseType.Int32
	p.SystemBoursesId = v.SystemBoursesId.Int64
	p.Symbol = v.Symbol.String
	p.StockSymbol = v.StockSymbol.String
	p.StockName = v.StockName.String
	p.TradeType = v.TradeType.Int32
	p.IsUpsDowns = v.IsUpsDowns.Int32
	p.OrderType = v.OrderType.Int32
	p.EntrustNum = v.EntrustNum.Int32
	p.EntrustStatus = v.EntrustStatus.Int32
	p.LimitPrice = v.LimitPrice.Float64
	p.MarketPrice = v.MarketPrice.Float64
	p.StopPrice = v.StopPrice.Float64
	p.TakeProfitPrice = v.TakeProfitPrice.Float64
	p.OrderAmount = v.OrderAmount.Float64
	p.OrderAmountSum = v.OrderAmountSum.Float64
	p.EarnestMoney = v.EarnestMoney.Float64
	p.HandlingFee = v.HandlingFee.Float64
	p.PositionPrice = v.PositionPrice.Float64
	p.DealPrice = v.DealPrice.Float64
	p.EntrustOrderNumber = v.EntrustOrderNumber.String
	p.StockHoldingsId = v.StockHoldingsId.Int64
	p.EntrustTime = v.EntrustTime.String
	p.UpdateTime = v.UpdateTime.String
	return p
}

// GetWssTradeOrdersSql ws写入订单SQL
func GetWssTradeOrdersSql(ordersJson map[string]string, UserId int64, entrustStatus int) string {
	sqlStr := "insert into bo_user_trade_orders (userId,entrustStatus,entrustOrderNumber,entrustTime"
	values := fmt.Sprintf(" values (%v,%v,'%s','%v'", UserId, entrustStatus, ordersJson["entrustOrderNumber"], time.Now().Format(time.DateTime))
	if len(ordersJson["bourseType"]) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "bourseType")
		values = fmt.Sprintf("%s,%v", values, ordersJson["bourseType"])
	}
	if len(ordersJson["systemBoursesId"]) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "systemBoursesId")
		values = fmt.Sprintf("%s,%v", values, ordersJson["systemBoursesId"])
	}
	if len(ordersJson["symbol"]) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "symbol")
		values = fmt.Sprintf("%s,'%s'", values, ordersJson["symbol"])
	}
	if len(ordersJson["stockSymbol"]) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "stockSymbol")
		values = fmt.Sprintf("%s,'%v'", values, ordersJson["stockSymbol"])
	}
	if len(ordersJson["stockName"]) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "stockName")
		values = fmt.Sprintf("%s,'%v'", values, ordersJson["stockName"])
	}
	if len(ordersJson["tradeType"]) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "tradeType")
		values = fmt.Sprintf("%s,%v", values, ordersJson["tradeType"])
	}
	if len(ordersJson["isUpsDowns"]) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "isUpsDowns")
		values = fmt.Sprintf("%s,%v", values, ordersJson["isUpsDowns"])
	}
	if len(ordersJson["orderType"]) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "orderType")
		values = fmt.Sprintf("%s,%v", values, ordersJson["orderType"])
	}
	if len(ordersJson["entrustNum"]) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "entrustNum")
		values = fmt.Sprintf("%s,%v", values, ordersJson["entrustNum"])
	}
	if len(ordersJson["limitPrice"]) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "limitPrice")
		values = fmt.Sprintf("%s,%v", values, ordersJson["limitPrice"])
	}
	if len(ordersJson["marketPrice"]) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "marketPrice")
		values = fmt.Sprintf("%s,%v", values, ordersJson["marketPrice"])
	}
	if len(ordersJson["stopPrice"]) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "stopPrice")
		values = fmt.Sprintf("%s,%v", values, ordersJson["stopPrice"])
	}
	if len(ordersJson["takeProfitPrice"]) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "takeProfitPrice")
		values = fmt.Sprintf("%s,%v", values, ordersJson["takeProfitPrice"])
	}
	if len(ordersJson["orderAmount"]) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "orderAmount")
		values = fmt.Sprintf("%s,%v", values, ordersJson["orderAmount"])
	}
	if len(ordersJson["orderAmountSum"]) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "orderAmountSum")
		values = fmt.Sprintf("%s,%v", values, ordersJson["orderAmountSum"])
	}
	if len(ordersJson["earnestMoney"]) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "earnestMoney")
		values = fmt.Sprintf("%s,%v", values, ordersJson["earnestMoney"])
	}
	if len(ordersJson["handlingFee"]) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "handlingFee")
		values = fmt.Sprintf("%s,%v", values, ordersJson["handlingFee"])
	}
	if len(ordersJson["positionPrice"]) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "positionPrice")
		values = fmt.Sprintf("%s,%v", values, ordersJson["positionPrice"])
	}
	if len(ordersJson["dealPrice"]) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "dealPrice")
		values = fmt.Sprintf("%s,%v", values, ordersJson["dealPrice"])
	}
	if len(ordersJson["stockHoldingsId"]) > 0 {
		sqlStr = fmt.Sprintf("%s,%s", sqlStr, "stockHoldingsId")
		values = fmt.Sprintf("%s,%v", values, ordersJson["stockHoldingsId"])
	}
	sqlStr = sqlStr + ")" + values + ")"
	return sqlStr
}
