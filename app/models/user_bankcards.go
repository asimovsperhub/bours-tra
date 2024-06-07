package models

import "database/sql"

// UserBankcards 用户银行卡表
type UserBankcards struct {
	BankCardId       sql.NullInt64  `json:"bankCardId"`       //银行卡id
	UserId           sql.NullInt64  `json:"userId"`           //用户表id
	ActualName       sql.NullString `json:"actualName"`       //真实姓名
	BankCountryCode  sql.NullString `json:"bankCountryCode"`  //国家代码
	BankCountryName  sql.NullString `json:"bankCountryName"`  //国家名称
	BankName         sql.NullString `json:"bankName"`         //银行名称
	BankCode         sql.NullString `json:"bankCode"`         //银行代码
	BankPhoneNumber  sql.NullString `json:"bankPhoneNumber"`  //手机号
	BankCardNumber   sql.NullString `json:"bankCardNumber"`   //卡号
	BankCardEmail    sql.NullString `json:"bankCardEmail"`    //邮箱
	DefaultBankcards sql.NullInt32  `json:"defaultBankcards"` //是否默认：1.默认，0.不默认
	AddTime          sql.NullString `json:"addTime"`          //创建时间
	UpdateTime       sql.NullString `json:"updateTime"`       //更新时间
}

// UserBankcardsJson 用户银行卡表
type UserBankcardsJson struct {
	BankCardId       int64  `json:"bankCardId"`       //银行卡id
	UserId           int64  `json:"userId"`           //用户表id
	ActualName       string `json:"actualName"`       //真实姓名
	BankCountryCode  string `json:"bankCountryCode"`  //国家代码
	BankCountryName  string `json:"bankCountryName"`  //国家名称
	BankName         string `json:"bankName"`         //银行名称
	BankCode         string `json:"bankCode"`         //银行代码
	BankPhoneNumber  string `json:"bankPhoneNumber"`  //手机号
	BankCardNumber   string `json:"bankCardNumber"`   //卡号
	BankCardEmail    string `json:"bankCardEmail"`    //邮箱
	DefaultBankcards int32  `json:"defaultBankcards"` //是否默认：1.默认，0.不默认
	AddTime          string `json:"addTime"`          //创建时间
	UpdateTime       string `json:"updateTime"`       //更新时间
}

func ConvUserBankcardsNull2Json(v UserBankcards, p UserBankcardsJson) UserBankcardsJson {
	p.BankCardId = v.BankCardId.Int64
	p.UserId = v.UserId.Int64
	p.ActualName = v.ActualName.String
	p.BankCountryCode = v.BankCountryCode.String
	p.BankCountryName = v.BankCountryName.String
	p.BankName = v.BankName.String
	p.BankCode = v.BankCode.String
	p.BankPhoneNumber = v.BankPhoneNumber.String
	p.BankCardNumber = v.BankCardNumber.String
	p.BankCardEmail = v.BankCardEmail.String
	p.DefaultBankcards = v.DefaultBankcards.Int32
	p.AddTime = v.AddTime.String
	p.UpdateTime = v.UpdateTime.String
	return p
}
