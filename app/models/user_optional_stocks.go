package models

import "database/sql"

// UserOptionalStocks 自选股表
type UserOptionalStocks struct {
	Id              sql.NullInt64   `json:"id"`              //自选股id
	UserId          sql.NullInt64   `json:"userId"`          //用户id
	SystemBoursesId sql.NullInt64   `json:"systemBoursesId"` //交易所股种id
	BourseType      sql.NullInt32   `json:"bourseType"`      //所属类型：1 数字币 2 股票 (冗余)
	StockCode       sql.NullString  `json:"stockCode"`       //股票代码
	Icon            sql.NullString  `json:"icon"`            //图标
	Unit            sql.NullString  `json:"unit"`            //单位
	FullName        sql.NullString  `json:"fullName"`        //全称
	BeforeClose     sql.NullFloat64 `json:"beforeClose"`     //昨天的最新价
	YesterdayClose  sql.NullFloat64 `json:"yesterdayClose"`  //前天的最新价
}

// UserOptionalStocksJson 自选股表
type UserOptionalStocksJson struct {
	Id              int64   `json:"id"`              //自选股id
	UserId          int64   `json:"userId"`          //用户id
	SystemBoursesId int64   `json:"systemBoursesId"` //交易所股种id
	BourseType      int32   `json:"bourseType"`      //所属类型：1 数字币 2 股票 (冗余)
	StockCode       string  `json:"stockCode"`       //股票代码
	Icon            string  `json:"icon"`            //图标
	Unit            string  `json:"unit"`            //单位
	FullName        string  `json:"fullName"`        //全称
	BeforeClose     float64 `json:"beforeClose"`     //昨天的最新价
	YesterdayClose  float64 `json:"yesterdayClose"`  //前天的最新价
}

func ConvUserOptionalStocksNull2Json(v UserOptionalStocks, p UserOptionalStocksJson) UserOptionalStocksJson {
	p.Id = v.Id.Int64
	p.UserId = v.UserId.Int64
	p.SystemBoursesId = v.SystemBoursesId.Int64
	p.BourseType = v.BourseType.Int32
	p.StockCode = v.StockCode.String
	p.Icon = v.Icon.String
	p.Unit = v.Unit.String
	p.FullName = v.FullName.String
	p.BeforeClose = v.BeforeClose.Float64
	p.YesterdayClose = v.YesterdayClose.Float64
	return p
}
