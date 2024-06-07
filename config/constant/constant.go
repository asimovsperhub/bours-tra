package constant

import "Bourse/app/models"

const (
	HeatBeat = "heatbeat"
	PONG     = "pong"
	// UserInfoUidKey “user_info_phoneNumber_”+UserId
	UserInfoUidKey = "user_info_"

	// UsdtsDefaultErcKey 用户ERC默认地址 “usdts_default_erc_”+UserId
	UsdtsDefaultErcKey = "usdts_default_erc_"

	// UsdtsDefaultTrcKey 用户TRC默认地址 “usdts_default_trc_”+UserId
	UsdtsDefaultTrcKey = "usdts_default_trc_"

	// UserFundAccountsKey “user_fund_accounts_”+账户类型枚举(1~6)+“_”+UserId  资金账号类型：1 现货资产 2 合约资产 3 马股资产 4 美股资产 5 日股资产 6 佣金账户
	UserFundAccountsKey = "user_fund_accounts_"

	//BankcardsDefaultKey "bankcards_"+UserId
	BankcardsDefaultKey = "default_bankcards_"

	// UserOptionalStocks 自选股 "user_optional_stocks_"+UserId
	UserOptionalStocks = "user_optional_stocks_"

	// SALTMD5 MD5的加密盐
	SALTMD5 = "asd%jobs"
	OK      = "请求成功"
)

type UserInfoStruct struct {
	Id                     int64                                `json:"id"`                     //userId
	Uid                    string                               `json:"uid"`                    //UID
	CountryCode            string                               `json:"countryCode"`            //国家代码
	PhoneNumber            int64                                `json:"phoneNumber"`            //手机号
	Email                  string                               `json:"email"`                  //电子邮箱
	LoginPassword          string                               `json:"loginPassword"`          //登陆密码
	TradePassword          string                               `json:"tradePassword"`          //交易密码
	Surname                string                               `json:"surname"`                //姓
	Name                   string                               `json:"name"`                   //名
	Sex                    int32                                `json:"sex"`                    //性别：1男 2女 0未设置
	Birthday               string                               `json:"birthday"`               //出生日期
	Country                string                               `json:"country"`                //国家
	NickName               string                               `json:"nickName"`               //昵称
	Avatar                 string                               `json:"avatar"`                 //头像
	IsRealName             int32                                `json:"isRealName"`             //是否已实名认证：0未认证 1已认证
	InviteCode             string                               `json:"inviteCode"`             //邀请码
	LastLoginTime          string                               `json:"lastLoginTime"`          //最近一次登陆时间
	Status                 int32                                `json:"status"`                 //状态：1 启用 2禁用 3 黑名单
	AccessToken            string                               `json:"accessToken"`            //登陆令牌
	AddTime                string                               `json:"addTime"`                //创建时间
	UpdateTime             string                               `json:"updateTime"`             //更新时间
	UserTerminalEquipments []*models.UserTerminalEquipmentsJson `json:"userTerminalEquipments"` //用户终端设备
	UserFundAccounts       []*models.UserFundAccountsJson       `json:"userFundAccounts"`       //资金账户
}

type Messages struct {
	Uid         string `json:"uid"`         // 发送消息用户uid
	Content     string `json:"content"`     // 文本消息内容
	ApiType     int32  `json:"apiType"`     // 接口类型：1.文字
	Type        string `json:"type"`        // 消息传输类型：1 文本 如果是心跳消息，该内容为pong
	RespondCode string `json:"respondCode"` //回应code
}
