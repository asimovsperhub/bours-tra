package trade_order

import (
	"Bourse/app/models"
	"Bourse/app/servers"
	"Bourse/common/global"
	"Bourse/common/inits"
	"Bourse/common/ordernum"
	"Bourse/common/translators"
	"Bourse/config"
	"Bourse/config/constant"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"gitee.com/phper95/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"net/http"
	"time"
)

var (
	conn     *sql.DB
	client   *redis.Client
	rabbitmq *inits.RabbitMQ

	UserInfo *constant.UserInfoStruct
	httpUrl  string = "http://test.coincosmic.net/market/previous-close?stocksTicker="
)

const (
	TableUserTradeOrders    = "bo_user_trade_orders"
	TableUserFundAccounts   = "bo_user_fund_accounts"
	TableUserOptionalStocks = "bo_user_optional_stocks"
	TableUserStockHoldings  = "bo_user_stock_holdings"
)

// PreviousClose 闭盘价结构体
type PreviousClose struct {
	Code int `json:"code"`
	Data []struct {
		Code       string  `json:"Code"`
		ClosePrice float64 `json:"ClosePrice"`
	} `json:"data"`
	Message string `json:"message"`
}

// UserTradeOrdersForm 交易订单表-add入参
type UserTradeOrdersForm struct {
	BourseType         int32   `json:"bourseType" form:"bourseType" binding:"required"`   //所属类型：1 现货 2 合约 3 马股 4 美股 5 日股
	SystemBoursesId    int64   `json:"systemBoursesId" form:"systemBoursesId"`            //交易所股种id
	Symbol             string  `json:"symbol" form:"symbol"`                              //交易代码
	StockSymbol        string  `json:"stockSymbol" form:"stockSymbol"`                    //股票代码
	StockName          string  `json:"stockName" form:"stockName"`                        //股票名字
	TradeType          int32   `json:"tradeType" form:"tradeType" binding:"required"`     //交易类型：1买入 2 卖出
	IsUpsDowns         int32   `json:"isUpsDowns" form:"isUpsDowns" binding:"required"`   //是否买涨跌：1 买涨 2 买跌
	OrderType          int32   `json:"orderType" form:"orderType"`                        //订单类型：0 未知 1 限价 2 市价 3 止损止盈
	EntrustNum         int32   `json:"entrustNum" form:"entrustNum"`                      //委托数量
	LimitPrice         float64 `json:"limitPrice" form:"limitPrice"`                      //限价
	MarketPrice        float64 `json:"marketPrice" form:"marketPrice"`                    //市价
	OrderAmount        float64 `json:"orderAmount" form:"orderAmount" binding:"required"` //订单金额
	LatestPrice        float64 `json:"latestPrice" form:"latestPrice"`                    //最新价
	PositionPrice      float64 `json:"positionPrice" form:"positionPrice"`                //持仓价
	EarnestMoney       float64 `json:"earnestMoney" form:"earnestMoney"`                  //保证金
	HandlingFee        float64 `json:"handlingFee" form:"handlingFee"`                    //手续费
	StopPrice          float64 `json:"stopPrice" form:"stopPrice"`                        //止损价
	TakeProfitPrice    float64 `json:"takeProfitPrice" form:"takeProfitPrice"`            //止盈价
	EntrustOrderNumber string  `json:"entrustOrderNumber" form:"entrustOrderNumber"`      //委托订单号
}

// UserOptionalStocksAddForm 自选股表-add入参
type UserOptionalStocksAddForm struct {
	SystemBoursesId int64  `json:"systemBoursesId" form:"systemBoursesId" binding:"required"` //交易所股种id
	BourseType      int32  `json:"bourseType" form:"bourseType" binding:"required"`           //所属类型：1 数字币 2 股票 (冗余)
	StockCode       string `json:"stockCode" form:"stockCode" binding:"required"`             //股票代码
}

// UserOptionalStocksListForm 自选股表-list入参
type UserOptionalStocksListForm struct {
	SystemBoursesId int64 `json:"systemBoursesId" form:"systemBoursesId"` //交易所股种id
	BourseType      int32 `json:"bourseType" form:"bourseType"`           //所属类型：1 数字币 2 股票 (冗余)
}

// UserStockHoldingsListForm 用户持仓股表-list入参
type UserStockHoldingsListForm struct {
	/*Id                 int64   `json:"id"`                                                        //持仓股id
	UserId             int64   `json:"userId"`                                                    //用户id*/
	BourseType       int32 `json:"bourseType" form:"bourseType" binding:"required"`           //所属类型：1 数字币 2 股票
	SystemBoursesId  int64 `json:"systemBoursesId" form:"systemBoursesId" binding:"required"` //交易所股种id
	StockHoldingType int32 `json:"stockHoldingType" form:"stockHoldingType"`                  //持仓状态：1 持仓 2关闭（数量为0时关闭）
	/*Symbol             string  `json:"symbol"`                                                    //交易代码
	StockSymbol        string  `json:"stockSymbol"`                                               //股票代码
	StockName          string  `json:"stockName"`                                                 //股票名字
	StockHoldingNum    int32   `json:"stockHoldingNum"`                                           //数量
	IsUpsDowns         int32   `json:"isUpsDowns"`                                                //是否买涨跌：1 买涨 2 买跌
	ProfitAndLoss      float64 `json:"profitAndLoss"`                                             //盈亏
	LimitPrice         float64 `json:"limitPrice"`                                                //限价
	MarketPrice        float64 `json:"marketPrice"`                                               //市价
	OrderAmount        float64 `json:"orderAmount"`                                               //订单金额
	LatestPrice        float64 `json:"latestPrice"`                                               //最新价
	PositionPrice      float64 `json:"positionPrice"`                                             //持仓价
	EarnestMoney       float64 `json:"earnestMoney"`                                              //保证金
	HandlingFee        float64 `json:"handlingFee"`                                               //手续费
	StopPrice          float64 `json:"stopPrice"`                                                 //止损价
	TakeProfitPrice    float64 `json:"takeProfitPrice"`                                           //止盈价
	EntrustOrderNumber string  `json:"entrustOrderNumber"`                                        //委托订单号
	OpenPositionTime   string  `json:"openPositionTime"`                                          //开仓时间
	UpdateTime         string  `json:"updateTime"`                                                //更新时间*/
}

// UserTradeOrdersListForm 交易订单表-list入参
type UserTradeOrdersListForm struct {
	BourseType         int32   `json:"bourseType" form:"bourseType" binding:"required"`           //所属类型：1 现货 2 合约 3 马股 4 美股 5 日股
	SystemBoursesId    int64   `json:"systemBoursesId" form:"systemBoursesId" binding:"required"` //交易所股种id
	Symbol             string  `json:"symbol" form:"symbol"`                                      //交易代码
	StockSymbol        string  `json:"stockSymbol" form:"stockSymbol"`                            //股票代码
	StockName          string  `json:"stockName" form:"stockName"`                                //股票名字
	TradeType          int32   `json:"tradeType" form:"tradeType"`                                //交易类型：1买入 2 卖出
	IsUpsDowns         int32   `json:"isUpsDowns" form:"isUpsDowns"`                              //是否买涨跌：1 买涨 2 买跌
	OrderType          int32   `json:"orderType" form:"orderType"`                                //订单类型：0 未知 1 限价 2 市价 3 止损止盈
	EntrustNum         int32   `json:"entrustNum" form:"entrustNum"`                              //委托数量
	LimitPrice         float64 `json:"limitPrice" form:"limitPrice"`                              //限价
	MarketPrice        float64 `json:"marketPrice" form:"marketPrice"`                            //市价
	OrderAmount        float64 `json:"orderAmount" form:"orderAmount"`                            //订单金额
	LatestPrice        float64 `json:"latestPrice" form:"latestPrice"`                            //最新价
	PositionPrice      float64 `json:"positionPrice" form:"positionPrice"`                        //持仓价
	EarnestMoney       float64 `json:"earnestMoney" form:"earnestMoney"`                          //保证金
	HandlingFee        float64 `json:"handlingFee" form:"handlingFee"`                            //手续费
	StopPrice          float64 `json:"stopPrice" form:"stopPrice"`                                //止损价
	TakeProfitPrice    float64 `json:"takeProfitPrice" form:"takeProfitPrice"`                    //止盈价
	EntrustOrderNumber string  `json:"entrustOrderNumber" form:"entrustOrderNumber"`              //委托订单号
	Page               int32   `json:"page" form:"page"`                                          //当前页
	PageSize           int32   `json:"pageSize" form:"pageSize"`                                  //每页数量
	//TotalNum           int32   `json:"totalNum"`           //当前页
}

// TradeOrdersForm 根据条件查询自选股入参
type TradeOrdersForm struct {
	BourseType       int32  `json:"bourseType" form:"bourseType"`             //所属类型：1 现货 2 合约 3 马股 4 美股 5 日股
	SystemBoursesId  int64  `json:"systemBoursesId" form:"systemBoursesId"`   //交易所股种id
	StockSymbol      string `json:"stockSymbol" form:"stockSymbol"`           //股票代码
	OptionalStocksId int64  `json:"optionalStocksId" form:"optionalStocksId"` //自选股id
}

// UserTradeOrdersPendingWithdrawalClearing 交易订单表-list出参
type UserTradeOrdersPendingWithdrawalClearing struct {
	Pending    []*models.UserTradeOrdersJson `json:"pending"`
	Withdrawal []*models.UserTradeOrdersJson `json:"withdrawal"`
	//Clearing   []*models.UserTradeOrdersJson `json:"clearing"`
}

// TakeProfitAndStopLossForm 设置止盈止损价
type TakeProfitAndStopLossForm struct {
	BourseType      int32   `json:"bourseType" form:"bourseType" binding:"required"`           //所属类型：1 现货 2 合约 3 马股 4 美股 5 日股
	SystemBoursesId int64   `json:"systemBoursesId" form:"systemBoursesId" binding:"required"` //交易所股种id
	Symbol          string  `json:"symbol" form:"symbol" binding:"required"`                   //交易代码
	StockSymbol     string  `json:"stockSymbol" form:"stockSymbol" binding:"required"`         //股票代码
	StopPrice       float64 `json:"stopPrice" form:"stopPrice" binding:"required"`             //止损价
	TakeProfitPrice float64 `json:"takeProfitPrice" form:"takeProfitPrice" binding:"required"` //止盈价
	StockHoldingsId int64   `json:"stockHoldingsId" form:"buyingPrice" binding:"required"`     //持仓id
}

func init() {
	//初始化配置方式
	config.LoadConfig()
	inits.InitLog()

	//多语言
	if err := translators.InitTranslators("zh"); err != nil {
		fmt.Sprintf("init Translator failed,err:%v\n", err)
	}
}

// UpdateTakeProfitAndStopLossSettings 修改止损止盈接口
func UpdateTakeProfitAndStopLossSettings(ctx *gin.Context) {
	//获取表单数据,并验证参数
	var takeProfitAndStopLossForm TakeProfitAndStopLossForm
	var msg any
	//redis初始化方式
	client = inits.InitRedis()
	UserInfo = global.GetUserInfo(client)

	//mysql初始化方式
	conn = inits.InitMysql()
	defer conn.Close() //关闭数据库连接

	//1、获取参数，并验证
	if errs := ctx.ShouldBindJSON(&takeProfitAndStopLossForm); errs != nil {
		code := 500
		err, ok := errs.(validator.ValidationErrors)
		msg = err.Translate(translators.Trans)
		if !ok {
			msg = errs.Error()
		}
		logger.Warn(fmt.Sprintf("Time:[%s] Err:[%v]", time.Now().Format(time.DateTime), err))
		config.Errors(ctx, code, msg)
		return
	}

	//查询持仓股是否存在
	sqlStr := "select id,userId,bourseType,systemBoursesId,symbol,stockSymbol,stockName,isUpsDowns,orderType,stockHoldingNum,stockHoldingType,limitPrice,marketPrice,stopPrice,takeProfitPrice,orderAmount,orderAmountSum,earnestMoney,handlingFee,positionPrice,dealPrice,buyingPrice,profitLossMoney,entrustOrderNumber,openPositionTime,updateTime from " + TableUserStockHoldings + " where id=? and userId=? and bourseType=? and systemBoursesId=? and stockHoldingType=?"
	rows, err := conn.Query(sqlStr, takeProfitAndStopLossForm.StockHoldingsId, UserInfo.Id, takeProfitAndStopLossForm.BourseType, takeProfitAndStopLossForm.SystemBoursesId, 1)
	if err != nil {
		logger.Warn(fmt.Sprintf("Time:[%s] Err:[%v]", time.Now().Format(time.DateTime), err.Error()))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	defer rows.Close()
	var userStockJson models.UserStockHoldingsJson
	for rows.Next() {
		var userStock models.UserStockHoldings
		rows.Scan(&userStock.Id, &userStock.UserId, &userStock.BourseType, &userStock.SystemBoursesId, &userStock.Symbol, &userStock.StockSymbol, &userStock.StockName, &userStock.IsUpsDowns, &userStock.OrderType, &userStock.StockHoldingNum, &userStock.StockHoldingType, &userStock.LimitPrice, &userStock.MarketPrice, &userStock.StopPrice, &userStock.TakeProfitPrice, &userStock.OrderAmount, &userStock.OrderAmountSum, &userStock.EarnestMoney, &userStock.HandlingFee, &userStock.PositionPrice, &userStock.DealPrice, &userStock.BuyingPrice, &userStock.ProfitLossMoney, &userStock.EntrustOrderNumber, &userStock.OpenPositionTime, &userStock.UpdateTime)
		userStockJson = models.ConvUserStockHoldingsNull2Json(userStock, userStockJson)
		fmt.Println(userStock.Id.Int64)
		break
	}
	//判断持仓是否存在
	if userStockJson.Id <= 0 {
		logger.Warn(fmt.Sprintf("该持仓数据不存在,id:[%d] Time:[%s]", takeProfitAndStopLossForm.StockHoldingsId, time.Now().Format(time.DateTime)))
		code := 500
		msg = "该持仓数据不存在"
		config.Errors(ctx, code, msg)
		return
	}

	//修改持股的止盈止损设置
	sqlStr = "update bo_user_stock_holdings set stopPrice=?,takeProfitPrice=?,updateTime=? where id=? and userId=? and stockHoldingType=?"
	exec, err := conn.Exec(sqlStr, takeProfitAndStopLossForm.StopPrice, takeProfitAndStopLossForm.TakeProfitPrice, time.Now().Format(time.DateTime), userStockJson.Id, UserInfo.Id, 1)
	if err != nil {
		logger.Warn(fmt.Sprintf("修改持股的止盈止损设置失败,id:[%d] Time:[%s] Err:[%v]", takeProfitAndStopLossForm.StockHoldingsId, time.Now().Format(time.DateTime), err.Error()))
		code := 500
		msg = "修改失败"
		config.Errors(ctx, code, msg)
		return
	}
	affected, err := exec.RowsAffected()
	if err != nil || affected != 1 {
		logger.Warn(fmt.Sprintf("修改持股的止盈止损设置失败,id:[%d] Time:[%s] affected:[%d]", takeProfitAndStopLossForm.StockHoldingsId, time.Now().Format(time.DateTime), affected))
		code := 500
		msg = "修改失败"
		config.Errors(ctx, code, msg)
		return
	}

	//修改改止盈止损的集合 获取用户持仓列表 redis的哈希key=user_stock_holdings_data_+ client.Uid + systemBoursesId ...
	userStockJson.StopPrice = takeProfitAndStopLossForm.StopPrice
	userStockJson.TakeProfitPrice = takeProfitAndStopLossForm.TakeProfitPrice
	userStockJson.UpdateTime = time.Now().Format(time.DateTime)
	marshal, _ := json.Marshal(userStockJson)
	cacheKey := "user_stock_holdings_data_" + UserInfo.Uid + fmt.Sprintf("_%d", takeProfitAndStopLossForm.SystemBoursesId)
	client.HSet(context.Background(), cacheKey, userStockJson.Id, marshal)

	// TODO 设置到止损止盈集合中 止损设置 (ZADD stock_order_stop_price_交易所id  价格 订单ID) ...
	tradeOrdersKey1 := fmt.Sprintf("stock_order_stop_price_%v", userStockJson.SystemBoursesId)
	orderMoney1 := userStockJson.StopPrice * global.PriceEnlargeMultiples
	client.ZAdd(context.Background(), tradeOrdersKey1, redis.Z{Score: orderMoney1, Member: userStockJson.Id}).Err()

	//  止赢设置 (ZADD stock_order_stop_price_交易所id  价格 订单ID)
	tradeOrdersKey2 := fmt.Sprintf("stock_order_take_profit_price_%v", userStockJson.SystemBoursesId)
	orderMoney2 := userStockJson.StopPrice * global.PriceEnlargeMultiples
	client.ZAdd(context.Background(), tradeOrdersKey2, redis.Z{Score: orderMoney2, Member: userStockJson.Id}).Err()

	config.Success(ctx, 200, "设置成功", nil)
	return
}

// CreateTradeOrder 数字币&股票交易下单(买入卖出)
func CreateTradeOrder(ctx *gin.Context) {
	//获取表单数据,并验证参数
	var userTradeOrdersForm UserTradeOrdersForm
	var msg any

	//redis初始化方式
	client = inits.InitRedis()
	UserInfo = global.GetUserInfo(client)

	//mysql初始化方式
	conn = inits.InitMysql()
	defer conn.Close() //关闭数据库连接

	// =========== RabbitMQ Simple模式 发送者 ===========
	queueKey := "tradeOrderBuy_" //现货 spots 合约 contract
	rabbitmq = inits.NewRabbitMQSimple(queueKey)

	//1、获取参数，并验证
	if errs := ctx.ShouldBindJSON(&userTradeOrdersForm); errs != nil {
		code := 500
		err, ok := errs.(validator.ValidationErrors)
		msg = err.Translate(translators.Trans)
		if !ok {
			msg = errs.Error()
		}
		logger.Warn(fmt.Sprintf("Time:[%s] Err:[%v]", time.Now().Format(time.DateTime), err))
		config.Errors(ctx, code, msg)
		return
	}

	//2、查询检验股票代码相关信息

	//3、获取根据交易类型用户资金账户信息-Redis查查不到-MySQL查
	userFundAccountsKey := fmt.Sprintf("%s%v_%v", constant.UserFundAccountsKey, userTradeOrdersForm.BourseType, UserInfo.Id)
	result, err := client.Get(context.Background(), userFundAccountsKey).Result()
	var userFundAccountsJson models.UserFundAccountsJson
	if err != nil || result == "" {
		//缓存没有进数据库取
		sqlStr := GetUserFundAccountsFirstSql()
		rows, err := conn.Query(sqlStr, UserInfo.Id, userTradeOrdersForm.BourseType)
		if err != nil {
			logger.Warn(fmt.Sprintf("Time:[%s] Err:[%v]", time.Now().Format(time.DateTime), err.Error()))
			code := 500
			msg = "网络繁忙"
			config.Errors(ctx, code, msg)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var userFundAccounts models.UserFundAccounts
			rows.Scan(&userFundAccounts.Id, &userFundAccounts.UserId, &userFundAccounts.FundAccount, &userFundAccounts.FundAccountType, &userFundAccounts.Money, &userFundAccounts.Unit)
			userFundAccountsJson = models.ConvUserFundAccountsNull2Json(userFundAccounts, userFundAccountsJson)
			break
		}
	} else {
		err = json.Unmarshal([]byte(result), &userFundAccountsJson)
		if err != nil {
			logger.Warn(fmt.Sprintf("Time:[%s] Err:[%v]", time.Now().Format(time.DateTime), err.Error()))
			code := 500
			msg = "网络繁忙"
			config.Errors(ctx, code, msg)
			return
		}
	}

	//4、判断用户账户余额是否充足
	if userFundAccountsJson.Money < userTradeOrdersForm.OrderAmount {
		logger.Warn(fmt.Sprintf("Time:[%s] Err:[用户账户资金不足 Money:%v OrderAmount:%v ]", time.Now().Format(time.DateTime), userFundAccountsJson.Money, userTradeOrdersForm.OrderAmount))
		code := 500
		msg = "用户账户资金不足"
		config.Errors(ctx, code, msg)
		return
	}

	//5、开启事务--MySQL-修改用户资金账户余额--修改Redis中用户账户资金余额--(写入MQ队列)写入资金账户交易订单表--写入交易订单表
	//5.1 修改资金账户余额&修改Redis中用户账户资金余额(买入操作)
	if userTradeOrdersForm.TradeType == 1 {
		userFundAccountServe := servers.UserFundAccountServe{
			Id:              userFundAccountsJson.Id,
			UserId:          userFundAccountsJson.UserId,
			FundAccount:     userFundAccountsJson.FundAccount,
			FundAccountType: userFundAccountsJson.FundAccountType,
			Operate:         2,
			OperateMoney:    userTradeOrdersForm.OrderAmount,
		}
		service, err := servers.UserFundAccountService(&userFundAccountServe, conn)

		if err != nil || service == false {
			logger.Warn(fmt.Sprintf("Time:[%s] Err:[下单失败. %v]", time.Now().Format(time.DateTime), err.Error()))
			code := 500
			msg = "下单失败"
			config.Errors(ctx, code, msg)
			return
		}
	}

	//5.2 (写入MQ队列)写入资金账户交易订单表--写入交易订单表
	marshal, err := json.Marshal(&userTradeOrdersForm)
	if err != nil {
		logger.Warn(fmt.Sprintf("Time:[%s] Err:[下单失败. %v]", time.Now().Format(time.DateTime), err.Error()))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	//发送队列消息
	rabbitmq.PublishSimple(string(marshal))

	config.Success(ctx, 200, "ok", ordernum.GetOrderNo())
	return
}

// GetListOptionalStock 获取自选股列表
func GetListOptionalStock(ctx *gin.Context) {
	//获取表单数据,并验证参数
	var listForm UserOptionalStocksListForm
	var msg any

	//redis初始化方式
	client = inits.InitRedis()
	UserInfo = global.GetUserInfo(client)

	//mysql初始化方式
	conn = inits.InitMysql()
	defer conn.Close() //关闭数据库连接

	//1、获取参数，并验证
	if errs := ctx.ShouldBind(&listForm); errs != nil {
		code := 500
		err, ok := errs.(validator.ValidationErrors)
		msg = err.Translate(translators.Trans)
		if !ok {
			msg = errs.Error()
		}
		logger.Warn(fmt.Sprintf("Time:[%s] Err:[%v]", time.Now().Format(time.DateTime), err))
		config.Errors(ctx, code, msg)
		return
	}

	sqlStr := GetListOptionalStockSql()
	rows, err := conn.Query(sqlStr, UserInfo.Id, listForm.SystemBoursesId, listForm.BourseType)
	if err != nil {
		logger.Warn(fmt.Sprintf("Exec SQL Failed:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	defer rows.Close()
	var userOptionalStocksJsonMap []*models.UserOptionalStocksJson
	for rows.Next() {
		var userOptionalStocks models.UserOptionalStocks
		var userOptionalStocksJson models.UserOptionalStocksJson
		rows.Scan(&userOptionalStocks.Id, &userOptionalStocks.UserId, &userOptionalStocks.SystemBoursesId, &userOptionalStocks.BourseType, &userOptionalStocks.StockCode, &userOptionalStocks.Icon, &userOptionalStocks.Unit, &userOptionalStocks.FullName, &userOptionalStocksJson.BeforeClose, &userOptionalStocksJson.YesterdayClose)
		userOptionalStocksJson = models.ConvUserOptionalStocksNull2Json(userOptionalStocks, userOptionalStocksJson)
		userOptionalStocksJsonMap = append(userOptionalStocksJsonMap, &userOptionalStocksJson)
	}

	config.Success(ctx, http.StatusOK, msg, &userOptionalStocksJsonMap)
	return
}

func GetOptionalStock(ctx *gin.Context) {
	//获取表单数据,并验证参数
	var tradeOrdersForm TradeOrdersForm
	var msg any

	//redis初始化方式
	client = inits.InitRedis()
	UserInfo = global.GetUserInfo(client)

	//mysql初始化方式
	conn = inits.InitMysql()
	defer conn.Close() //关闭数据库连接

	//1、获取参数，并验证
	if errs := ctx.ShouldBind(&tradeOrdersForm); errs != nil {
		code := 500
		err, ok := errs.(validator.ValidationErrors)
		msg = "网络繁忙"
		if !ok {
			msg = errs.Error()
		}
		logger.Warn(fmt.Sprintf("Time:[%s] Err:[%v]", time.Now().Format(time.DateTime), err))
		config.Errors(ctx, code, msg)
		return
	}

	sqlStr := "select id,userId,systemBoursesId,bourseType,stockCode,icon,unit,fullName,beforeClose,yesterdayClose from bo_user_optional_stocks where "
	if tradeOrdersForm.OptionalStocksId > 0 {
		sqlStr = fmt.Sprintf("%sid=%d", sqlStr, tradeOrdersForm.OptionalStocksId)
	} else {
		sqlStr = fmt.Sprintf("%suserId=%d and stockCode='%s' and bourseType=%d and systemBoursesId=%d", sqlStr, UserInfo.Id, tradeOrdersForm.StockSymbol, tradeOrdersForm.BourseType, tradeOrdersForm.SystemBoursesId)
	}

	rows, err := conn.Query(sqlStr)
	if err != nil {
		logger.Warn(fmt.Sprintf("Exec SQL Failed:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	defer rows.Close()
	var userOptionalStocksJson models.UserOptionalStocksJson
	for rows.Next() {
		var userOptionalStocks models.UserOptionalStocks
		rows.Scan(&userOptionalStocks.Id, &userOptionalStocks.UserId, &userOptionalStocks.SystemBoursesId, &userOptionalStocks.BourseType, &userOptionalStocks.StockCode, &userOptionalStocks.Icon, &userOptionalStocks.Unit, &userOptionalStocks.FullName, &userOptionalStocksJson.BeforeClose, &userOptionalStocksJson.YesterdayClose)
		userOptionalStocksJson = models.ConvUserOptionalStocksNull2Json(userOptionalStocks, userOptionalStocksJson)
		break
	}

	var data any
	if userOptionalStocksJson.Id == 0 {
		data = nil
	} else {
		data = userOptionalStocksJson
	}

	config.Success(ctx, http.StatusOK, msg, data)
	return
}

// AddOptionalStock 添加自选股
func AddOptionalStock(ctx *gin.Context) {
	//获取表单数据,并验证参数
	var userOptionalStocksAddForm UserOptionalStocksAddForm
	var msg any

	//redis初始化方式
	client = inits.InitRedis()
	UserInfo = global.GetUserInfo(client)

	//mysql初始化方式
	conn = inits.InitMysql()
	defer conn.Close() //关闭数据库连接

	//1、获取参数，并验证
	if errs := ctx.ShouldBindJSON(&userOptionalStocksAddForm); errs != nil {
		code := 500
		err, ok := errs.(validator.ValidationErrors)
		msg = "网络繁忙"
		if !ok {
			msg = errs.Error()
		}
		logger.Warn(fmt.Sprintf("Time:[%s] Err:[%v]", time.Now().Format(time.DateTime), err))
		config.Errors(ctx, code, msg)
		return
	}

	sqlStr := GetAddOptionalStockSql()
	exec, err := conn.Exec(sqlStr, UserInfo.Id, userOptionalStocksAddForm.SystemBoursesId, userOptionalStocksAddForm.BourseType, userOptionalStocksAddForm.StockCode)
	if err != nil {
		logger.Warn(fmt.Sprintf("Exec SQL Failed:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	LastId, err := exec.LastInsertId()
	if err != nil || LastId <= 0 {
		logger.Warn(fmt.Sprintf("Exec SQL Failed:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	config.Success(ctx, http.StatusOK, msg, LastId)
	return
}

// DeleteOptionalStock 删除自选股列表
func DeleteOptionalStock(ctx *gin.Context) {
	//获取表单数据,并验证参数
	var msg any

	//redis初始化方式
	client = inits.InitRedis()
	UserInfo = global.GetUserInfo(client)

	//mysql初始化方式
	conn = inits.InitMysql()
	defer conn.Close() //关闭数据库连接

	//1、获取参数，并验证
	systemBoursesId := ctx.Query("optionalStocksId")
	if len(systemBoursesId) == 0 {
		code := 500
		msg = "参数错误，请重新提交"
		config.Errors(ctx, code, msg)
		return
	}

	sqlStr := GetDeleteOptionalStockSql()
	exec, err := conn.Exec(sqlStr, systemBoursesId, UserInfo.Id)
	if err != nil {
		logger.Warn(fmt.Sprintf("Exec SQL Failed:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	affected, err := exec.RowsAffected()
	if err != nil || affected != 1 {
		logger.Warn(fmt.Sprintf("Exec SQL Failed:[%s] Time:[%s] Err:[%v] RowsAffected：%v", sqlStr, time.Now().Format(time.DateTime), err, affected))
		code := 500
		msg = "删除失败"
		config.Errors(ctx, code, msg)
		return
	}

	config.Success(ctx, http.StatusOK, msg, &affected)
	return
}

// GetListUserTradePosition 获取用户交易持仓列表
func GetListUserTradePosition(ctx *gin.Context) {
	//获取表单数据,并验证参数
	var listForm UserStockHoldingsListForm
	var msg any

	//redis初始化方式
	client = inits.InitRedis()
	UserInfo = global.GetUserInfo(client)

	//mysql初始化方式
	conn = inits.InitMysql()
	defer conn.Close() //关闭数据库连接

	//1、获取参数，并验证
	if errs := ctx.ShouldBind(&listForm); errs != nil {
		code := 500
		err, ok := errs.(validator.ValidationErrors)
		msg = err.Translate(translators.Trans)
		if !ok {
			msg = errs.Error()
		}
		logger.Warn(fmt.Sprintf("Time:[%s] Err:[%v]", time.Now().Format(time.DateTime), err))
		config.Errors(ctx, code, msg)
		return
	}
	fmt.Printf("%+v", listForm)
	sqlStr := GetListUserTradePositionSql()
	rows, err := conn.Query(sqlStr, UserInfo.Id, listForm.BourseType, listForm.SystemBoursesId, listForm.StockHoldingType)
	if err != nil {
		logger.Warn(fmt.Sprintf("Exec SQL Failed:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	defer rows.Close()
	var userStockHoldingsJsonMap []*models.UserStockHoldingsJson
	//var previousClose PreviousClose
	for rows.Next() {
		var userStockHoldingsJson models.UserStockHoldingsJson
		var userStockHoldings models.UserStockHoldings
		rows.Scan(&userStockHoldings.Id, &userStockHoldings.UserId, &userStockHoldings.BourseType, &userStockHoldings.SystemBoursesId, &userStockHoldings.Symbol, &userStockHoldings.StockSymbol, &userStockHoldings.StockName, &userStockHoldings.IsUpsDowns, &userStockHoldings.OrderType, &userStockHoldings.StockHoldingNum, &userStockHoldings.StockHoldingType, &userStockHoldings.LimitPrice, &userStockHoldings.MarketPrice, &userStockHoldings.StopPrice, &userStockHoldings.TakeProfitPrice, &userStockHoldings.OrderAmount, &userStockHoldings.OrderAmountSum, &userStockHoldings.EarnestMoney, &userStockHoldings.HandlingFee, &userStockHoldings.PositionPrice, &userStockHoldings.DealPrice, &userStockHoldings.BuyingPrice, &userStockHoldings.ProfitLossMoney, &userStockHoldings.EntrustOrderNumber, &userStockHoldings.OpenPositionTime, &userStockHoldings.UpdateTime)
		userStockHoldingsJson = models.ConvUserStockHoldingsNull2Json(userStockHoldings, userStockHoldingsJson)
		//替换最新价并计算盈亏
		/*response, err := http_request.HttpGet(fmt.Sprintf("%s%s", httpUrl, userStockHoldingsJson.StockSymbol))
		if err != nil {
			logger.Warn(fmt.Sprintf("http请求最新价失败 Time:[%s] Err:[%s]", time.Now().Format(time.DateTime), err.Error()))
			return
		}
		json.Unmarshal(response, &previousClose)
		for _, datum := range previousClose.Data {
			if datum.Code == userStockHoldingsJson.StockSymbol {
				//最新价
				userStockHoldingsJson.MarketPrice = datum.ClosePrice
				//计算盈亏
				var stockHoldingNumF64 float64
				stockHoldingNumF64, _ = strconv.ParseFloat(fmt.Sprintf("%d", userStockHoldingsJson.StockHoldingNum), 64)
				if userStockHoldingsJson.IsUpsDowns == 1 { //买涨
					userStockHoldingsJson.ProfitLossMoney = (userStockHoldingsJson.PositionPrice - datum.ClosePrice) * stockHoldingNumF64
				} else if userStockHoldingsJson.IsUpsDowns == 2 { //买跌
					userStockHoldingsJson.ProfitLossMoney = (datum.ClosePrice - userStockHoldingsJson.PositionPrice) * stockHoldingNumF64
				}
			}
		}*/
		userStockHoldingsJsonMap = append(userStockHoldingsJsonMap, &userStockHoldingsJson)
	}

	config.Success(ctx, http.StatusOK, msg, &userStockHoldingsJsonMap)
	return
}

// GetListUserTradePendingWithdrawalClearing 获取用户交易挂单-撤单列表
func GetListUserTradePendingWithdrawalClearing(ctx *gin.Context) {
	//获取表单数据,并验证参数
	var listForm UserTradeOrdersListForm
	var msg any

	//redis初始化方式
	client = inits.InitRedis()
	UserInfo = global.GetUserInfo(client)

	//mysql初始化方式
	conn = inits.InitMysql()
	defer conn.Close() //关闭数据库连接

	//1、获取参数，并验证
	if errs := ctx.ShouldBind(&listForm); errs != nil {
		code := 500
		err, ok := errs.(validator.ValidationErrors)
		msg = err.Translate(translators.Trans)
		if !ok {
			msg = errs.Error()
		}
		logger.Warn(fmt.Sprintf("Time:[%s] Err:[%v]", time.Now().Format(time.DateTime), err))
		config.Errors(ctx, code, msg)
		return
	}

	sqlStr := GetListUserTradePendingWithdrawalClearingSql()
	//1 挂单 Pending
	rows, err := conn.Query(sqlStr, UserInfo.Id, listForm.BourseType, listForm.SystemBoursesId, 1, 1)
	if err != nil {
		logger.Warn(fmt.Sprintf("Exec SQL Failed:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	defer rows.Close()
	var pendingJsonMap []*models.UserTradeOrdersJson
	//var previousClose PreviousClose
	for rows.Next() {
		var pending models.UserTradeOrders
		var pendingJson models.UserTradeOrdersJson
		rows.Scan(&pending.Id, &pending.UserId, &pending.BourseType, &pending.SystemBoursesId, &pending.Symbol, &pending.StockSymbol, &pending.StockName, &pending.TradeType, &pending.IsUpsDowns, &pending.OrderType, &pending.EntrustNum, &pending.EntrustStatus, &pending.LimitPrice, &pending.MarketPrice, &pending.StopPrice, &pending.TakeProfitPrice, &pending.OrderAmount, &pending.OrderAmountSum, &pending.EarnestMoney, &pending.HandlingFee, &pending.PositionPrice, &pending.DealPrice, &pending.EntrustOrderNumber, &pending.StockHoldingsId, &pending.EntrustTime, &pending.UpdateTime)
		pendingJson = models.ConvUserTradeOrdersNull2Json(pending, pendingJson)
		//替换最新价并计算盈亏
		/*response, err := http_request.HttpGet(fmt.Sprintf("%s%s", httpUrl, pendingJson.StockSymbol))
		if err != nil {
			logger.Warn(fmt.Sprintf("http请求最新价失败 Time:[%s] Err:[%s]", time.Now().Format(time.DateTime), err.Error()))
			return
		}
		json.Unmarshal(response, &previousClose)
		for _, datum := range previousClose.Data {
			if datum.Code == pendingJson.StockSymbol {
				//最新价
				pendingJson.MarketPrice = datum.ClosePrice
			}
		}*/
		pendingJsonMap = append(pendingJsonMap, &pendingJson)
	}

	//2 撤单 Withdrawal
	rows, err = conn.Query(sqlStr, UserInfo.Id, listForm.BourseType, listForm.SystemBoursesId, 1, 3)
	if err != nil {
		logger.Warn(fmt.Sprintf("Exec SQL Failed:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	defer rows.Close()
	var withdrawalJsonMap []*models.UserTradeOrdersJson
	for rows.Next() {
		var withdrawal models.UserTradeOrders
		var withdrawalJson models.UserTradeOrdersJson
		rows.Scan(&withdrawal.Id, &withdrawal.UserId, &withdrawal.BourseType, &withdrawal.SystemBoursesId, &withdrawal.Symbol, &withdrawal.StockSymbol, &withdrawal.StockName, &withdrawal.TradeType, &withdrawal.IsUpsDowns, &withdrawal.OrderType, &withdrawal.EntrustNum, &withdrawal.EntrustStatus, &withdrawal.LimitPrice, &withdrawal.MarketPrice, &withdrawal.StopPrice, &withdrawal.TakeProfitPrice, &withdrawal.OrderAmount, &withdrawal.OrderAmountSum, &withdrawal.EarnestMoney, &withdrawal.HandlingFee, &withdrawal.PositionPrice, &withdrawal.DealPrice, &withdrawal.EntrustOrderNumber, &withdrawal.StockHoldingsId, &withdrawal.EntrustTime, &withdrawal.UpdateTime)
		withdrawalJson = models.ConvUserTradeOrdersNull2Json(withdrawal, withdrawalJson)
		//替换最新价并计算盈亏
		/*response, err := http_request.HttpGet(fmt.Sprintf("%s%s", httpUrl, withdrawalJson.StockSymbol))
		if err != nil {
			logger.Warn(fmt.Sprintf("http请求最新价失败 Time:[%s] Err:[%s]", time.Now().Format(time.DateTime), err.Error()))
			return
		}
		json.Unmarshal(response, &previousClose)
		for _, datum := range previousClose.Data {
			if datum.Code == withdrawalJson.StockSymbol {
				//最新价
				withdrawalJson.MarketPrice = datum.ClosePrice
			}
		}*/
		withdrawalJsonMap = append(withdrawalJsonMap, &withdrawalJson)
	}

	var pendingWithdrawalClearing UserTradeOrdersPendingWithdrawalClearing
	pendingWithdrawalClearing.Pending = pendingJsonMap
	pendingWithdrawalClearing.Withdrawal = withdrawalJsonMap

	// TODO 请求行情获取闭盘价 获取最新价&计数每支盈亏 ...

	config.Success(ctx, http.StatusOK, msg, &pendingWithdrawalClearing)
	return
}

// GetCountProfitLosAndAccumulative 获取总浮动盈亏&历史累计盈亏
func GetCountProfitLosAndAccumulative(ctx *gin.Context) {
	var msg string
	//redis初始化方式
	client = inits.InitRedis()
	UserInfo = global.GetUserInfo(client)

	//mysql初始化方式
	conn = inits.InitMysql()
	defer conn.Close() //关闭数据库连接
	//获取累计盈亏
	sqlStr := "select count(profitLossMoney) as profitLossMoneyCount,count(orderAmount) as orderAmountCount from bo_user_stock_holdings where userId=? and stockHoldingType=?"
	rows, err := conn.Query(sqlStr, UserInfo.Id, 2)
	if err != nil {
		logger.Warn(fmt.Sprintf("Exec SQL Failed:[%s] Time:[%s] Err:[%v]", sqlStr, time.Now().Format(time.DateTime), err.Error()))
		code := 500
		msg = "网络繁忙"
		config.Errors(ctx, code, msg)
		return
	}
	defer rows.Close()
	var profitLossMoneyCount float64
	var orderAmountCount float64
	for rows.Next() {
		rows.Scan(&profitLossMoneyCount, &orderAmountCount)
	}

	//获取总浮动盈亏 TODO 目前只计算了美股,新增股累加 ...
	profitLossMoneyCountkey := fmt.Sprintf("stockHoldings_profitLossMoneykey_Count_%d", 4)
	accumulativeCount, _ := client.ZScore(context.Background(), profitLossMoneyCountkey, fmt.Sprintf("%d", UserInfo.Id)).Result()

	data := make(map[string]float64)
	data["profitLossCount"] = profitLossMoneyCount
	data["accumulativeCount"] = accumulativeCount
	data["orderAmountCount"] = orderAmountCount
	config.Success(ctx, http.StatusOK, msg, data)
	return
}
