package user_fund_accounts

const (
	TableUserFundAccounts = "bo_user_fund_accounts"
	UserFundAccountsField = "id,userId,fundAccount,fundAccountType,money,unit"
)

func GetListUserFundAccountsSql() (strSql string) {
	strSql = "select " + UserFundAccountsField + " from " + TableUserFundAccounts + " where userId=? order by fundAccountType asc"
	return
}
