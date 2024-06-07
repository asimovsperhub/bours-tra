package trade_order

func GetUserFundAccountsFirstSql() (sqlStr string) {
	sqlStr = "select id,userId,fundAccountType,fundAccount,money,unit,addTime,updateTime from " + TableUserFundAccounts + " where userId=? and fundAccountType=? limit 1"
	return
}

func GetAddOptionalStockSql() (sqlStr string) {
	sqlStr = "insert into " + TableUserOptionalStocks + " (userId,systemBoursesId,bourseType,stockCode) values(?,?,?,?)"
	return
}

func GetListOptionalStockSql() (sqlStr string) {
	sqlStr = "select id,userId,systemBoursesId,bourseType,stockCode,icon,unit,fullName,beforeClose,yesterdayClose from " + TableUserOptionalStocks + " where userId=? and systemBoursesId=? and bourseType=?"
	return
}

func GetDeleteOptionalStockSql() (sqlStr string) {
	sqlStr = "delete from " + TableUserOptionalStocks + " where id=? and userId=?"
	return
}

func GetListUserTradePositionSql() (sqlStr string) {
	sqlStr = "select id,userId,bourseType,systemBoursesId,symbol,stockSymbol,stockName,isUpsDowns,orderType,stockHoldingNum,stockHoldingType,limitPrice,marketPrice,stopPrice,takeProfitPrice,orderAmount,orderAmountSum,earnestMoney,handlingFee,positionPrice,dealPrice,buyingPrice,profitLossMoney,entrustOrderNumber,openPositionTime,updateTime from " + TableUserStockHoldings + " where userId=? and bourseType=? and systemBoursesId=? and stockHoldingType=?"
	return
}

func GetListUserTradePendingWithdrawalClearingSql() (sqlStr string) {
	sqlStr = "select id,userId,bourseType,systemBoursesId,symbol,stockSymbol,stockName,tradeType,isUpsDowns,orderType,entrustNum,entrustStatus,limitPrice,marketPrice,stopPrice,takeProfitPrice,orderAmount,orderAmountSum,earnestMoney,handlingFee,positionPrice,dealPrice,entrustOrderNumber,stockHoldingsId,entrustTime,updateTime from " + TableUserTradeOrders + " where userId=? and bourseType=? and systemBoursesId=? and tradeType=? and entrustStatus=?"
	return
}
