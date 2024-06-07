package models

import (
	"database/sql"
)

// UserFundAccounts 资金账户
type UserFundAccounts struct {
	Id              sql.NullInt64   `json:"id"`              //资金账户id
	UserId          sql.NullInt64   `json:"userId"`          //用户id
	FundAccountType sql.NullInt32   `json:"fundAccountType"` //资金账号类型：1 现货USDT(充值默认账户)  2 合约USDT 3 马股资产 4 美股资产 5 日股资产 6 佣金账户USDT
	FundAccount     sql.NullString  `json:"fundAccount"`     //资金账号
	Money           sql.NullFloat64 `json:"money"`           //账户金额
	Unit            sql.NullString  `json:"unit"`            //单位
	AddTime         sql.NullString  `json:"addTime"`         //创建时间
	UpdateTime      sql.NullString  `json:"updateTime"`      //更新时间
}

// UserFundAccountsJson 资金账户
type UserFundAccountsJson struct {
	Id              int64   `json:"id"`              //资金账户id
	UserId          int64   `json:"userId"`          //用户id
	FundAccountType int32   `json:"fundAccountType"` //资金账号类型：1 现货USDT(充值默认账户)  2 合约USDT 3 马股资产 4 美股资产 5 日股资产 6 佣金账户USDT
	FundAccount     string  `json:"fundAccount"`     //资金账号
	Money           float64 `json:"money"`           //账户金额
	Unit            string  `json:"unit"`            //单位
	AddTime         string  `json:"addTime"`         //创建时间
	UpdateTime      string  `json:"updateTime"`      //更新时间
}

func ConvUserFundAccountsNull2Json(v UserFundAccounts, p UserFundAccountsJson) UserFundAccountsJson {
	p.Id = v.Id.Int64
	p.UserId = v.UserId.Int64
	p.FundAccount = v.FundAccount.String
	p.FundAccountType = v.FundAccountType.Int32
	p.Money = v.Money.Float64
	p.Unit = v.Unit.String
	p.AddTime = v.AddTime.String
	p.UpdateTime = v.UpdateTime.String
	return p
}
