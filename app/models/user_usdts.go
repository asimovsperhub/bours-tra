package models

import "database/sql"

// UserUsdts 用户USDT地址表
type UserUsdts struct {
	UsdtId         sql.NullInt64  `json:"usdtId"`         //USDT地ID
	FundAccountsId sql.NullInt64  `json:"fundAccountsId"` //资金账户id
	UserId         sql.NullInt64  `json:"userId"`         //用户表ID
	AddreType      sql.NullInt32  `json:"addreType"`      //地址类型：1 ERC 2 TRC
	UsdtAddre      sql.NullString `json:"usdtAddre"`      //usdt地址
	UsdtTitle      sql.NullString `json:"usdtTitle"`      //usdt备注名称
	DefaultCard    sql.NullInt32  `json:"defaultCard"`    //是否默认：0 否  1 默认
	AddTime        sql.NullString `json:"addTime"`        //添加时间
}

// UserUsdtsJson 用户USDT地址表
type UserUsdtsJson struct {
	UsdtId         int64  `json:"usdtId"`         //USDT地ID
	FundAccountsId int64  `json:"fundAccountsId"` //资金账户id
	UserId         int64  `json:"userId"`         //用户表ID
	AddreType      int32  `json:"addreType"`      //地址类型：1 ERC 2 TRC
	UsdtAddre      string `json:"usdtAddre"`      //usdt地址
	UsdtTitle      string `json:"usdtTitle"`      //usdt备注名称
	DefaultCard    int32  `json:"defaultCard"`    //是否默认：0 否  1 默认
	AddTime        string `json:"addTime"`        //添加时间
}
