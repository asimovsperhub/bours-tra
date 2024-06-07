package main

import (
	"Bourse/app/models"
	"Bourse/common/http_request"
	"Bourse/common/inits"
	"Bourse/config"
	"context"
	"encoding/json"
	"fmt"
	"gitee.com/phper95/pkg/logger"
	"time"
)

type PreviousClose struct {
	Code int `json:"code"`
	Data []struct {
		Code       string  `json:"Code"`
		ClosePrice float64 `json:"ClosePrice"`
	} `json:"data"`
	Message string `json:"message"`
}

func main() {
	httpUrl := "http://test.coincosmic.net/market/previous-close?stocksTicker="
	response, err := http_request.HttpGet(fmt.Sprintf("%s%s", httpUrl, "A"))
	if err != nil {
		logger.Warn(fmt.Sprintf("http请求最新价失败 Time:[%s] Err:[%s]", time.Now().Format(time.DateTime), err.Error()))
		return
	}
	var data PreviousClose
	json.Unmarshal(response, &data)
	fmt.Printf("%v===%+v", data.Code, data.Data)

	//Bearer ennM000Gt-kckgmhKjB_xEnarmxB1VM8-1686567228
	str := "Bearer ennM000Gt-kckgmhKjB_xEnarmxB1VM8-1686567228"
	fmt.Println(str[7:])

	return

	//初始化配置方式
	config.LoadConfig()
	inits.InitLog()

	mysqlDb := inits.InitMysql()
	defer mysqlDb.Close()
	redisDb := inits.InitRedis()

	sqlStr := "select id,userId,bourseType,systemBoursesId,symbol,stockSymbol,stockName,isUpsDowns,orderType,stockHoldingNum,stockHoldingType,limitPrice,marketPrice,stopPrice,takeProfitPrice,orderAmount,orderAmountSum,earnestMoney,handlingFee,positionPrice,dealPrice,buyingPrice,profitLossMoney,entrustOrderNumber,openPositionTime,updateTime from bo_user_stock_holdings where userId=? and systemBoursesId=? and stockHoldingType=?"
	rows, err := mysqlDb.Query(sqlStr, 1, 4, 1)
	if err != nil {
		logger.Warn(fmt.Sprintf("写入缓存，查询写入的持仓数据失败 Err:[%s] Time:[%s]", err.Error(), time.Now().Format(time.DateTime)))
		return
	}
	defer rows.Close()
	var userStockJson models.UserStockHoldingsJson
	for rows.Next() {
		var userStock models.UserStockHoldings
		rows.Scan(&userStock.Id, &userStock.UserId, &userStock.BourseType, &userStock.SystemBoursesId, &userStock.Symbol, &userStock.StockSymbol, &userStock.StockName, &userStock.IsUpsDowns, &userStock.OrderType, &userStock.StockHoldingNum, &userStock.StockHoldingType, &userStock.LimitPrice, &userStock.MarketPrice, &userStock.StopPrice, &userStock.TakeProfitPrice, &userStock.OrderAmount, &userStock.OrderAmountSum, &userStock.EarnestMoney, &userStock.HandlingFee, &userStock.PositionPrice, &userStock.DealPrice, &userStock.BuyingPrice, &userStock.ProfitLossMoney, &userStock.EntrustOrderNumber, &userStock.OpenPositionTime, &userStock.UpdateTime)
		userStockJson = models.ConvUserStockHoldingsNull2Json(userStock, userStockJson)
		break
	}
	//fmt.Printf("%+v", userStockJson)

	//获取用户持仓集合 redis的哈希key=user_stock_holdings_data_+ client.Uid + systemBoursesId ...
	cacheKey := "user_stock_holdings_data_" + "11111_" + fmt.Sprintf("%d", 4)
	marshal, _ := json.Marshal(userStockJson)
	//fmt.Println(marshal)
	_, err = redisDb.HSet(context.Background(), cacheKey, userStockJson.Id, marshal).Result()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

}
