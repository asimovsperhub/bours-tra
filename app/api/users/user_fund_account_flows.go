package users

import (
	"Bourse/app/models"
	"Bourse/common/global"
	"Bourse/common/inits"
	"Bourse/config"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// FundAccountFlowsForm 资金账户流水表
type FundAccountFlowsForm struct {
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

type UserFundAccountFlowsList struct {
	List     []*models.UserFundAccountFlowsJson `json:"list"`
	PageNum  int                                `json:"pageNum"`
	PageSize int                                `json:"pageSize"`
	Total    int32                              `json:"total"`
}

// GetFundAccountFlowsList 获取用户资金账户流水记录列表
func GetFundAccountFlowsList(ctx *gin.Context) {
	var msg any
	UserInfo = global.GetUserInfo(client)

	conn = inits.InitMysql()
	defer conn.Close()

	pageNum, _ := strconv.Atoi(ctx.DefaultQuery("pageNum", "1"))     //获取分页数量 pageNum
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "1"))   //获取当前页 pageSize total
	fundAccountType, _ := strconv.Atoi(ctx.Query("fundAccountType")) //获取分页数量 pageNum
	if fundAccountType <= 0 {
		code := 500
		msg = "参数错误"
		config.Errors(ctx, code, msg)
		return
	}

	//计算总数
	sqlStr := "select count(1) as total from bo_user_fund_account_flows where userId=? and fundAccountType=?"
	rows, err := conn.Query(sqlStr, UserInfo.Id, fundAccountType)
	if err != nil {
		code := 500
		msg = "系统内部错误"
		config.Errors(ctx, code, msg)
		return
	}
	defer rows.Close()
	var Total int32
	for rows.Next() {
		rows.Scan(&Total)
	}

	sqlStr = "select id,userId,fundAccountType,operationType,changeType,changeAmount,beforeAmount,afterAmount,unit,orderNo,remark,addTime from bo_user_fund_account_flows where userId=? and fundAccountType=? limit ? offset ?"
	rows, err = conn.Query(sqlStr, UserInfo.Id, fundAccountType, pageSize, (pageNum-1)*pageSize)
	if err != nil {
		code := 500
		msg = "系统内部错误"
		config.Errors(ctx, code, msg)
		return
	}
	defer rows.Close()

	var fundAccountFlowsJsonMap []*models.UserFundAccountFlowsJson
	for rows.Next() {
		var fundAccountFlows models.UserFundAccountFlows
		var undAccountFlowsJson models.UserFundAccountFlowsJson
		rows.Scan(&fundAccountFlows.Id, &fundAccountFlows.UserId, &fundAccountFlows.FundAccountType, &fundAccountFlows.OperationType, &fundAccountFlows.ChangeType, &fundAccountFlows.ChangeAmount, &fundAccountFlows.BeforeAmount, &fundAccountFlows.AfterAmount, &fundAccountFlows.Unit, &fundAccountFlows.OrderNo, &fundAccountFlows.Remark, &fundAccountFlows.AddTime)
		undAccountFlowsJson = models.ConvUserFundAccountFlowsNull2Json(fundAccountFlows, undAccountFlowsJson)
		fundAccountFlowsJsonMap = append(fundAccountFlowsJsonMap, &undAccountFlowsJson)
	}
	var fundAccountFlowsListList UserFundAccountFlowsList
	fundAccountFlowsListList.List = fundAccountFlowsJsonMap
	fundAccountFlowsListList.Total = Total
	fundAccountFlowsListList.PageNum = pageNum
	fundAccountFlowsListList.PageSize = pageSize
	config.Success(ctx, http.StatusOK, "请求成功", &fundAccountFlowsListList)
	return
}
