package models

import "database/sql"

// UserFundAccountSpots 资金子账户-现货
type UserFundAccountSpots struct {
	Id              sql.NullInt64   `json:"id"`              //资金账户id
	FundAccountsId  sql.NullInt64   `json:"fundAccountsId"`  //用户USDT现货资金表id
	UserId          sql.NullInt64   `json:"userId"`          //用户id
	FundAccountSpot sql.NullString  `json:"fundAccountSpot"` //资金账号
	FundAccountName sql.NullString  `json:"fundAccountName"` //资金账户名称-对应币名
	SpotMoney       sql.NullFloat64 `json:"spotMoney"`       //账户金额(与对应持仓表对应)
	Unit            sql.NullString  `json:"unit"`            //单位
	AddTime         sql.NullString  `json:"addTime"`         //创建时间
	UpdateTime      sql.NullString  `json:"updateTime"`      //更新时间
}

// UserFundAccountSpotsJson 资金子账户-现货
type UserFundAccountSpotsJson struct {
	Id              int64   `json:"id"`              //资金账户id
	FundAccountsId  int64   `json:"fundAccountsId"`  //用户USDT现货资金表id
	UserId          int64   `json:"userId"`          //用户id
	FundAccountSpot string  `json:"fundAccountSpot"` //资金账号
	FundAccountName string  `json:"fundAccountName"` //资金账户名称-对应币名
	SpotMoney       float64 `json:"spotMoney"`       //账户金额(与对应持仓表对应)
	Unit            string  `json:"unit"`            //单位
	AddTime         string  `json:"addTime"`         //创建时间
	UpdateTime      string  `json:"updateTime"`      //更新时间
}

func ConvUserFundAccountSpotsNull2Json(v UserFundAccountSpots, p UserFundAccountSpotsJson) UserFundAccountSpotsJson {
	p.Id = v.Id.Int64
	p.UserId = v.UserId.Int64
	p.FundAccountSpot = v.FundAccountSpot.String
	p.FundAccountName = v.FundAccountName.String
	p.SpotMoney = v.SpotMoney.Float64
	p.Unit = v.Unit.String
	p.AddTime = v.AddTime.String
	p.UpdateTime = v.UpdateTime.String
	return p
}
