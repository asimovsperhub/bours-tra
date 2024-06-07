package routers

import (
	"Bourse/app/api/trade_order"
	"Bourse/app/api/user_fund_accounts"
	"Bourse/app/api/user_usdts"
	"Bourse/app/api/users"
	"Bourse/app/api/websocket"
	"Bourse/app/middleware/auth"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()

	//登陆注册
	//注册--邮箱
	//r.POST("/order-wss/users/register", users.PostUserRegister)

	r.Use(auth.Cors(), auth.Auth(), auth.Recovery) //门面 r.Use(auth.CrossDomain(), auth.Auth())

	//用户
	//获取用户设备信息
	r.GET("/order-wss/users/get-user-terminal-equipments-list", users.GetUserTerminalEquipmentsList)

	//user-usdts
	//新增用户USDT地址
	r.POST("/order-wss/users/create-user-usdts", user_usdts.CreateUserUsdts)
	//删除用户USDT地址（软删）
	r.DELETE("/order-wss/users/delete-user-usdts", user_usdts.DeleteUserUsdts)
	//根据usdtId获取用户USDT地址
	r.GET("/order-wss/users/get-user-usdts", user_usdts.GetUserUsdts)
	//获取用户USDT地址列表
	r.GET("/order-wss/users/get-list-user-usdts", user_usdts.GetListUserUsdts)

	//withdraw-money 获取用户提款信息
	r.GET("/order-wss/users/get-withdraw-money-info", users.GetWithdrawMoneyInfo)

	//获取用户资金账户列表
	r.GET("/order-wss/users/get-user-accounts-balance-list", users.GetUserAccountsBalanceList)
	//获取用户资金账户流水记录列表
	r.GET("/order-wss/users/get-fund-account-flows-list", users.GetFundAccountFlowsList)
	//用户账户资金划转
	r.POST("/order-wss/users/post-fund-account-transfer", users.PostFundAccountTransfer)
	//获取用户划转记录
	r.GET("/order-wss/users/get-fund-account-transfer-list", users.GetFundAccountTransferList)
	//获取汇率与计算划转金额
	r.GET("/order-wss/users/count-exchange-rates", users.CountExchangeRates)

	//user-fund-accounts 获取用户资金账户列表
	r.GET("/order-wss/users/get-list-user-fund-accounts", user_fund_accounts.GetListUserFundAccounts)

	//trade-order
	//设置持仓止损止盈
	r.POST("/order-wss/trade-order/update-take-profit-and-stop-loss-settings", trade_order.UpdateTakeProfitAndStopLossSettings)
	//数字币&股票交易下单(买入卖出)
	r.POST("/order-wss/trade-order/create-trade-order", trade_order.CreateTradeOrder)
	//获取自选股列表
	r.GET("/order-wss/trade-order/get-list-optional-stock", trade_order.GetListOptionalStock)
	//根据id获取自选股
	r.GET("/order-wss/trade-order/get-optional-stock", trade_order.GetOptionalStock)
	//添加自选股
	r.POST("/order-wss/trade-order/add-optional-stock", trade_order.AddOptionalStock)
	//删除自选股
	r.DELETE("/order-wss/trade-order/delete-optional-stock", trade_order.DeleteOptionalStock)
	//获取用户交易持仓列表
	r.GET("/order-wss/trade-order/get-list-user-trade-position", trade_order.GetListUserTradePosition)
	//获取用户交易挂单-撤单-结算列表
	r.GET("/order-wss/trade-order/get-list-pending-withdrawal-clearing", trade_order.GetListUserTradePendingWithdrawalClearing)
	//获取总浮动盈亏&历史累计盈亏&订单总金额
	r.GET("/order-wss/trade-order/get-count-profitLos-and-accumulative", trade_order.GetCountProfitLosAndAccumulative)

	//订单wss
	r.GET("/order-wss/wss", websocket.RunSocket)

	return r
}
