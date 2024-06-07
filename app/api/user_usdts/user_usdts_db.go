package user_usdts

const (
	TableUserUsdts = "bo_user_usdts"
	UserUsdtsField = "usdtId,userId,addreType,usdtAddre,usdtTitle,defaultCard"
)

func GetUserUsdtsDefaultCardSql() (sqlStr string) {
	sqlStr = "select " + UserUsdtsField + " from " + TableUserUsdts + " where userId=? and addreType=? and defaultCard=1 and deleteTime is null limit 1"
	return
}

func GetCreateInUpdateUserUsdtsSql() (sqlStr string) {
	sqlStr = "update " + TableUserUsdts + " set defaultCard=? where usdtId=? and userId=?"
	return
}

func GetCreateUserUsdtsSql() (sqlStr string) {
	sqlStr = "insert into " + TableUserUsdts + "(userId,addreType,usdtAddre,usdtTitle,defaultCard,addTime) values (?,?,?,?,?,?)"
	return
}

func GetDeleteUserUsdtsSql() (sqlStr string) {
	sqlStr = "update " + TableUserUsdts + " set deleteTime=? where usdtId=? and userId=?"
	return
}

func GetUserUsdtsSql() (sqlStr string) {
	sqlStr = "select " + UserUsdtsField + " from " + TableUserUsdts + " where usdtId=? and userId=? and defaultCard=? and deleteTime is null limit 1"
	return
}

func GetUserAllUsdtsSql() (sqlStr string) {
	sqlStr = "select " + UserUsdtsField + " from " + TableUserUsdts + " where userId=? and deleteTime is null"
	return
}
