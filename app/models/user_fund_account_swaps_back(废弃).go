package models

import "database/sql"

// UserFundAccountSwaps 资金子账户-合约
type UserFundAccountSwaps_back struct {
	Id              sql.NullInt64   `json:"id"`              //资金账户id
	UserId          sql.NullInt64   `json:"userId"`          //用户id
	FundAccountSwap sql.NullString  `json:"fundAccountSwap"` //资金账号
	FundAccountType sql.NullInt32   `json:"fundAccountType"` //资金账号类型：1 现货资产 2 合约资产
	FundAccountName sql.NullString  `json:"fundAccountName"` //资金账户名称-对应币名
	SwapMoney       sql.NullFloat64 `json:"swapMoney"`       //账户金额
	Unit            sql.NullString  `json:"unit"`            //单位
	AddTime         sql.NullString  `json:"addTime"`         //创建时间
	UpdateTime      sql.NullString  `json:"updateTime"`      //更新时间
}

// UserFundAccountSwapsJson 资金子账户-合约
type UserFundAccountSwapsJson_back struct {
	Id              int64   `json:"id"`              //资金账户id
	UserId          int64   `json:"userId"`          //用户id
	FundAccountSwap string  `json:"fundAccountSwap"` //资金账号
	FundAccountType int32   `json:"fundAccountType"` //资金账号类型：1 现货资产 2 合约资产
	FundAccountName string  `json:"fundAccountName"` //资金账户名称-对应币名
	SwapMoney       float64 `json:"swapMoney"`       //账户金额
	Unit            string  `json:"unit"`            //单位
	AddTime         string  `json:"addTime"`         //创建时间
	UpdateTime      string  `json:"updateTime"`      //更新时间
}

/*func ConvUserFundAccountSwapsNull2Json_back(v UserFundAccountSwaps, p UserFundAccountSwapsJson) UserFundAccountSwapsJson {
	p.Id = v.Id.Int64
	p.UserId = v.UserId.Int64
	p.FundAccountSwap = v.FundAccountSwap.String
	p.FundAccountType = v.FundAccountType.Int32
	p.FundAccountName = v.FundAccountName.String
	p.SwapMoney = v.SwapMoney.Float64
	p.Unit = v.Unit.String
	p.AddTime = v.AddTime.String
	p.UpdateTime = v.UpdateTime.String
	return p
}*/
