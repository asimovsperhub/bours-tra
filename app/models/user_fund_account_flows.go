package models

import (
	"database/sql"
	"fmt"
	"time"
)

// UserFundAccountFlows 资金账户流水表
type UserFundAccountFlows struct {
	Id              sql.NullInt64   `json:"id"`              //流水id
	UserId          sql.NullInt64   `json:"userId"`          //用户id
	FundAccountType sql.NullInt32   `json:"fundAccountType"` //资金账号类型：1 现货USDT(充值默认账户)  2 合约USDT 3 马股资产 4 美股资产 5 日股资产 6 佣金账户USDT
	OperationType   sql.NullInt32   `json:"operationType"`   //操作类型: 1 购股 2 购股保证金 3 购股手续费 4 持仓结算 5 充值 6 提款 7 团队分佣 8 持仓结算手续费 9 平台操作
	ChangeType      sql.NullInt32   `json:"changeType"`      //变动反向：1 增加  2 扣减
	ChangeAmount    sql.NullFloat64 `json:"changeAmount"`    //变动金额
	BeforeAmount    sql.NullFloat64 `json:"beforeAmount"`    //变动前金额
	AfterAmount     sql.NullFloat64 `json:"afterAmount"`     //变动后金额
	Unit            sql.NullString  `json:"unit"`            //金额单位
	OrderNo         sql.NullString  `json:"orderNo"`         //操作对应单号（没有的不填）
	Remark          sql.NullString  `json:"remark"`          //操作备注
	AddTime         sql.NullString  `json:"addTime"`         //创建时间
}

// UserFundAccountFlowsJson 资金账户流水表
type UserFundAccountFlowsJson struct {
	Id              int64   `json:"id"`              //流水id
	UserId          int64   `json:"userId"`          //用户id
	FundAccountType int32   `json:"fundAccountType"` //资金账号类型：1 现货USDT(充值默认账户)  2 合约USDT 3 马股资产 4 美股资产 5 日股资产 6 佣金账户USDT
	OperationType   int32   `json:"operationType"`   //操作类型: 1 购股 2 购股保证金 3 购股手续费 4 持仓结算 5 充值 6 提款 7 团队分佣 8 持仓结算手续费 9 平台操作
	ChangeType      int32   `json:"changeType"`      //变动反向：1 增加  2 扣减
	ChangeAmount    float64 `json:"changeAmount"`    //变动金额
	BeforeAmount    float64 `json:"beforeAmount"`    //变动前金额
	AfterAmount     float64 `json:"afterAmount"`     //变动后金额
	Unit            string  `json:"unit"`            //金额单位
	OrderNo         string  `json:"orderNo"`         //操作对应单号（没有的不填）
	Remark          string  `json:"remark"`          //操作备注
	AddTime         string  `json:"addTime"`         //创建时间
}

func ConvUserFundAccountFlowsNull2Json(v UserFundAccountFlows, p UserFundAccountFlowsJson) UserFundAccountFlowsJson {
	addTime, _ := time.Parse(time.RFC3339, v.AddTime.String)
	p.Id = v.Id.Int64
	p.UserId = v.UserId.Int64
	p.FundAccountType = v.FundAccountType.Int32
	p.OperationType = v.OperationType.Int32
	p.ChangeType = v.ChangeType.Int32
	p.ChangeAmount = v.ChangeAmount.Float64
	p.BeforeAmount = v.BeforeAmount.Float64
	p.AfterAmount = v.AfterAmount.Float64
	p.Unit = v.Unit.String
	p.OrderNo = v.OrderNo.String
	p.Remark = v.Remark.String
	p.AddTime = fmt.Sprintf("%s", addTime.Format(time.DateTime))
	return p
}
