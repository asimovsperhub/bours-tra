package models

import "database/sql"

// SystemExchangeRates 汇率表
type SystemExchangeRates struct {
	Id           sql.NullInt64   `json:"id"`           //汇率表id
	SourceCode   sql.NullString  `json:"sourceCode"`   //原货币简码
	TargetCode   sql.NullString  `json:"targetCode"`   //目标货币简码
	ExchangeRate sql.NullFloat64 `json:"exchangeRate"` //汇率
	AddTime      sql.NullString  `json:"addTime"`      //添加时间
}

// SystemExchangeRatesJson 汇率表
type SystemExchangeRatesJson struct {
	Id           int64   `json:"id"`           //汇率表id
	SourceCode   string  `json:"sourceCode"`   //原货币简码
	TargetCode   string  `json:"targetCode"`   //目标货币简码
	ExchangeRate float64 `json:"exchangeRate"` //汇率
	AddTime      string  `json:"addTime"`      //添加时间
}
