package models

import (
	"database/sql"
	"fmt"
	"time"
)

type UserFundAccountTransferBill struct {
	Id                      sql.NullInt64   `json:"id"`
	UserId                  sql.NullInt64   `json:"userId"`                  //用户ID
	TransferId              sql.NullInt64   `json:"transferId"`              //被划账户ID
	TransferAmount          sql.NullFloat64 `json:"transferAmount"`          //划转金额
	TransferFundAccountType sql.NullInt32   `json:"transferFundAccountType"` //被划转账户类型
	TransferUnit            sql.NullString  `json:"transferUnit"`            //被划转金额单位
	CollectId               sql.NullInt64   `json:"collectId"`               //收款账户ID
	CollectAmount           sql.NullFloat64 `json:"collectAmount"`           //划转收款金额
	CollectFundAccountType  sql.NullInt32   `json:"collectFundAccountType"`  //收款账户类型
	CollectUnit             sql.NullString  `json:"collectUnit"`             //收款金额单位
	OrderNo                 sql.NullString  `json:"orderNo"`                 //划转订单号
	ExchangeRate            sql.NullFloat64 `json:"exchangeRate"`            //汇率
	Status                  sql.NullInt32   `json:"status"`                  //划转状态
	AddTime                 sql.NullString  `json:"addTime"`                 //划转时间
}

type UserFundAccountTransferBillJson struct {
	Id                      int64   `json:"id"`
	UserId                  int64   `json:"userId"`                  //用户ID
	TransferId              int64   `json:"transferId"`              //被划账户ID
	TransferAmount          float64 `json:"transferAmount"`          //划转金额
	TransferFundAccountType int32   `json:"transferFundAccountType"` //被划转账户类型
	TransferUnit            string  `json:"transferUnit"`            //被划转金额单位
	CollectId               int64   `json:"collectId"`               //收款账户ID
	CollectAmount           float64 `json:"collectAmount"`           //划转收款金额
	CollectFundAccountType  int32   `json:"collectFundAccountType"`  //收款账户类型
	CollectUnit             string  `json:"collectUnit"`             //收款金额单位
	OrderNo                 string  `json:"orderNo"`                 //划转订单号
	ExchangeRate            float64 `json:"exchangeRate"`            //汇率
	Status                  int32   `json:"status"`                  //划转状态
	AddTime                 string  `json:"addTime"`                 //划转时间
}

func ConvUserFundAccountTransferBillNull2Json(v UserFundAccountTransferBill, p UserFundAccountTransferBillJson) UserFundAccountTransferBillJson {
	addTime, _ := time.Parse(time.RFC3339, v.AddTime.String)
	p.Id = v.Id.Int64
	p.UserId = v.UserId.Int64
	p.TransferId = v.TransferId.Int64
	p.TransferAmount = v.TransferAmount.Float64
	p.TransferFundAccountType = v.TransferFundAccountType.Int32
	p.TransferUnit = v.TransferUnit.String
	p.CollectId = v.CollectId.Int64
	p.CollectAmount = v.CollectAmount.Float64
	p.CollectFundAccountType = v.CollectFundAccountType.Int32
	p.CollectUnit = v.CollectUnit.String
	p.OrderNo = v.OrderNo.String
	p.ExchangeRate = v.ExchangeRate.Float64
	p.Status = v.Status.Int32
	p.AddTime = fmt.Sprintf("%s", addTime.Format(time.DateTime))
	return p
}
