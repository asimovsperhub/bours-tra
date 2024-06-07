package models

import "database/sql"

// SystemBourses 交易所股种表
type SystemBourses struct {
	Id           sql.NullInt64   `json:"id"`           //交易所股种id
	BourseName   sql.NullString  `json:"bourseName"`   //交易所股种名称
	BourseType   sql.NullInt32   `json:"bourseType"`   //所属类型：1 数字币 2 股票
	EarnestMoney sql.NullFloat64 `json:"earnestMoney"` //保证金
	HandlingFee  sql.NullFloat64 `json:"handlingFee"`  //手续费
	Sort         sql.NullInt32   `json:"sort"`         //排序
	Status       sql.NullInt32   `json:"status"`       //状态：0 关闭 1开启
}

// SystemBoursesJson 交易所股种表
type SystemBoursesJson struct {
	Id           int64   `json:"id"`           //交易所股种id
	BourseName   string  `json:"bourseName"`   //交易所股种名称
	BourseType   int32   `json:"bourseType"`   //所属类型：1 数字币 2 股票
	EarnestMoney float64 `json:"earnestMoney"` //保证金
	HandlingFee  float64 `json:"handlingFee"`  //手续费
	Sort         int32   `json:"sort"`         //排序
	Status       int32   `json:"status"`       //状态：0 关闭 1开启
}

func ConvSystemBoursesNull2Json(v SystemBourses, p SystemBoursesJson) SystemBoursesJson {
	p.Id = v.Id.Int64
	p.BourseName = v.BourseName.String
	p.BourseType = v.BourseType.Int32
	p.EarnestMoney = v.EarnestMoney.Float64
	p.HandlingFee = v.HandlingFee.Float64
	p.Sort = v.Sort.Int32
	p.Status = v.Status.Int32
	return p
}
