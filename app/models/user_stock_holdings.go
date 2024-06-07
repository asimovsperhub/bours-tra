package models

import (
	"database/sql"
	"fmt"
	"time"
)

// UserStockHoldings 用户持仓股表
type UserStockHoldings struct {
	Id                 sql.NullInt64   `json:"id"`                 //持仓股id
	UserId             sql.NullInt64   `json:"userId"`             //用户id
	BourseType         sql.NullInt32   `json:"bourseType"`         //交易所类型：1 数字币 2 股票
	SystemBoursesId    sql.NullInt64   `json:"systemBoursesId"`    //交易所股种id：1 现货 2 合约 3 马股 4 美股 5 日股
	Symbol             sql.NullString  `json:"symbol"`             //交易代码
	StockSymbol        sql.NullString  `json:"stockSymbol"`        //股票代码
	StockName          sql.NullString  `json:"stockName"`          //股票名字
	IsUpsDowns         sql.NullInt32   `json:"isUpsDowns"`         //是否买涨跌：1 买涨 2 买跌
	OrderType          sql.NullInt32   `json:"orderType"`          //交易方式：1 限价(需走撮合)   2 市价(下单直接成交)
	StockHoldingNum    sql.NullInt32   `json:"stockHoldingNum"`    //持仓数量
	StockHoldingType   sql.NullInt32   `json:"stockHoldingType"`   //持仓状态：1 持仓 2关闭
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
	BuyingPrice        sql.NullFloat64 `json:"buyingPrice"`        //开仓价(买入价)
	ProfitLossMoney    sql.NullFloat64 `json:"profitLossMoney"`    //平仓盈亏
	EntrustOrderNumber sql.NullString  `json:"entrustOrderNumber"` //委托订单号
	OpenPositionTime   sql.NullString  `json:"openPositionTime"`   //开仓时间
	UpdateTime         sql.NullString  `json:"updateTime"`         //更新时间
}

// UserStockHoldingsJson 用户持仓股表
type UserStockHoldingsJson struct {
	Id                 int64   `json:"id"`                 //持仓股id
	UserId             int64   `json:"userId"`             //用户id
	BourseType         int32   `json:"bourseType"`         //交易所类型：1 数字币 2 股票
	SystemBoursesId    int64   `json:"systemBoursesId"`    //交易所股种id：1 现货 2 合约 3 马股 4 美股 5 日股
	Symbol             string  `json:"symbol"`             //交易代码
	StockSymbol        string  `json:"stockSymbol"`        //股票代码
	StockName          string  `json:"stockName"`          //股票名字
	IsUpsDowns         int32   `json:"isUpsDowns"`         //是否买涨跌：1 买涨 2 买跌
	OrderType          int32   `json:"orderType"`          //交易方式：1 限价(需走撮合)   2 市价(下单直接成交)
	StockHoldingNum    int32   `json:"stockHoldingNum"`    //持仓数量
	StockHoldingType   int32   `json:"stockHoldingType"`   //持仓状态：1 持仓 2关闭(结算)
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
	BuyingPrice        float64 `json:"buyingPrice"`        //开仓价(买入价)
	ProfitLossMoney    float64 `json:"profitLossMoney"`    //平仓盈亏
	EntrustOrderNumber string  `json:"entrustOrderNumber"` //委托订单号
	OpenPositionTime   string  `json:"openPositionTime"`   //开仓时间
	UpdateTime         string  `json:"updateTime"`         //更新时间
}

func ConvUserStockHoldingsNull2Json(v UserStockHoldings, p UserStockHoldingsJson) UserStockHoldingsJson {
	openPositionTime, _ := time.Parse(time.RFC3339, v.OpenPositionTime.String)
	updateTime, _ := time.Parse(time.RFC3339, v.UpdateTime.String)
	p.Id = v.Id.Int64
	p.UserId = v.UserId.Int64
	p.BourseType = v.BourseType.Int32
	p.SystemBoursesId = v.SystemBoursesId.Int64
	p.Symbol = v.Symbol.String
	p.StockSymbol = v.StockSymbol.String
	p.StockName = v.StockName.String
	p.IsUpsDowns = v.IsUpsDowns.Int32
	p.OrderType = v.OrderType.Int32
	p.StockHoldingNum = v.StockHoldingNum.Int32
	p.StockHoldingType = v.StockHoldingType.Int32
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
	p.BuyingPrice = v.BuyingPrice.Float64
	p.ProfitLossMoney = v.ProfitLossMoney.Float64
	p.EntrustOrderNumber = v.EntrustOrderNumber.String
	p.OpenPositionTime = fmt.Sprintf("%s", openPositionTime.Format(time.DateTime))
	p.UpdateTime = fmt.Sprintf("%s", updateTime.Format(time.DateTime))
	return p
}
