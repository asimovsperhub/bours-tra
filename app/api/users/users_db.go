package users

const (
	TableUserBankcards = "bo_user_bankcards"
	TableUserUsdts     = "bo_user_usdts"
	UserBankcardsField = "bankCardId,bankCardNumber"
	//UserBankcardsField = "actualName,bankCountryCode,bankCountryName,bankName,bankCode,bankPhoneNumber,bankCardNumber,bankCardEmail,defaultBankcards"
	UserUsdtsField = "addreType,usdtAddre,usdtTitle,defaultCard"
)

func GetUserBankcardsSql() (sqlStr string) {
	sqlStr = "select " + UserBankcardsField + " from " + TableUserBankcards + " where userId= ? and defaultBankcards=? and deleteTime is null limit 1"
	return
}

func GetUserUsdtsDefaultSql() (sqlStr string) {
	sqlStr = "select " + UserUsdtsField + " from " + TableUserUsdts + " where userId= ? and defaultCard=? order by addreType asc limit 2"
	return
}
