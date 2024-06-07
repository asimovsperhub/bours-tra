package main

import (
	"Bourse/app/models"
	"Bourse/common/inits"
	"Bourse/common/translators"
	"Bourse/config"
	"Bourse/config/constant"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"gitee.com/phper95/pkg/logger"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

var (
	conn   *sql.DB
	client *redis.Client

	UserInfo *constant.UserInfoStruct

	file string
)

const (
	TableUsers = "bo_users"
	UsersField = "id,uid,countryCode,phoneNumber,email,loginPassword,tradePassword,surname,name,sex,birthday,country,nickName,avatar,isRealName,inviteCode,lastLoginTime,status,accessToken,addTime,updateTime"

	TableUserTerminalEquipments = "bo_user_terminal_equipments"
	UserTerminalEquipmentsField = "terminalEquipmentId,userId,terminalEquipmentName,loginCountry,loginCity,loginIp,addTime"

	TableUserUsdts = "bo_user_usdts"
	UserUsdtsField = "usdtId,userId,addreType,usdtAddre,usdtTitle,defaultCard,addTime"

	TableUserFundAccounts = "bo_user_fund_accounts"
	UserFundAccountsField = "id,userId,fundAccountType,fundAccount,money,unit,addTime,updateTime"

	TableUserBankcards = "bo_user_bankcards"
	UserBankcardsField = "bankCardId,userId,actualName,bankCountryCode,bankCountryName,bankName,bankCode,bankPhoneNumber,bankCardNumber,bankCardEmail,defaultBankcards,addTime,updateTime"
)

// Users 用户表
type Users struct {
	Id            sql.NullInt64  `json:"id"`            //userId
	Uid           sql.NullString `json:"uid"`           //UID
	CountryCode   sql.NullString `json:"countryCode"`   //国家代码
	PhoneNumber   sql.NullInt64  `json:"phoneNumber"`   //手机号
	Email         sql.NullString `json:"email"`         //电子邮箱
	LoginPassword sql.NullString `json:"loginPassword"` //登陆密码
	TradePassword sql.NullString `json:"tradePassword"` //交易密码
	Surname       sql.NullString `json:"surname"`       //姓
	Name          sql.NullString `json:"name"`          //名
	Sex           sql.NullInt32  `json:"sex"`           //性别：1男 2女 0未设置
	Birthday      sql.NullString `json:"birthday"`      //出生日期
	Country       sql.NullString `json:"country"`       //国家
	NickName      sql.NullString `json:"nickName"`      //昵称
	Avatar        sql.NullString `json:"avatar"`        //头像
	IsRealName    sql.NullInt32  `json:"isRealName"`    //是否已实名认证：0未认证 1已认证
	InviteCode    sql.NullString `json:"inviteCode"`    //邀请码
	LastLoginTime sql.NullString `json:"lastLoginTime"` //最近一次登陆时间
	Status        sql.NullInt32  `json:"status"`        //状态：1 启用 2禁用 3 黑名单
	AccessToken   sql.NullString `json:"accessToken"`   //登陆令牌
	AddTime       sql.NullString `json:"addTime"`       //创建时间
	UpdateTime    sql.NullString `json:"updateTime"`    //更新时间
}

// UsersJson 用户表
type UsersJson struct {
	Id                     int64                         `json:"id"`                     //userId
	Uid                    string                        `json:"uid"`                    //UID
	CountryCode            string                        `json:"countryCode"`            //国家代码
	PhoneNumber            int64                         `json:"phoneNumber"`            //手机号
	Email                  string                        `json:"email"`                  //电子邮箱
	LoginPassword          string                        `json:"loginPassword"`          //登陆密码
	TradePassword          string                        `json:"tradePassword"`          //交易密码
	Surname                string                        `json:"surname"`                //姓
	Name                   string                        `json:"name"`                   //名
	Sex                    int32                         `json:"sex"`                    //性别：1男 2女 0未设置
	Birthday               string                        `json:"birthday"`               //出生日期
	Country                string                        `json:"country"`                //国家
	NickName               string                        `json:"nickName"`               //昵称
	Avatar                 string                        `json:"avatar"`                 //头像
	IsRealName             int32                         `json:"isRealName"`             //是否已实名认证：0未认证 1已认证
	InviteCode             string                        `json:"inviteCode"`             //邀请码
	LastLoginTime          string                        `json:"lastLoginTime"`          //最近一次登陆时间
	Status                 int32                         `json:"status"`                 //状态：1 启用 2禁用 3 黑名单
	AccessToken            string                        `json:"accessToken"`            //登陆令牌
	AddTime                string                        `json:"addTime"`                //创建时间
	UpdateTime             string                        `json:"updateTime"`             //更新时间
	UserTerminalEquipments []*UserTerminalEquipmentsJson `json:"userTerminalEquipments"` //用户终端设备
	UserFundAccounts       []*UserFundAccountsJson       `json:"userFundAccounts"`       //资金账户
}

// UserTerminalEquipments 用户终端设备表
type UserTerminalEquipments struct {
	TerminalEquipmentId   sql.NullInt64  `json:"terminalEquipmentId"`   //终端设备表id
	UserId                sql.NullInt64  `json:"userId"`                //用户表id
	TerminalEquipmentName sql.NullString `json:"terminalEquipmentName"` //终端设备名称
	LoginCountry          sql.NullString `json:"loginCountry"`          //登录国家
	LoginCity             sql.NullString `json:"loginCity"`             //登录城市
	LoginIp               sql.NullInt64  `json:"loginIp"`               //登录IP
	AddTime               sql.NullString `json:"addTime"`               //创建时间
}

// UserTerminalEquipmentsJson 用户终端设备表
type UserTerminalEquipmentsJson struct {
	TerminalEquipmentId   int64  `json:"terminalEquipmentId"`   //终端设备表id
	UserId                int64  `json:"userId"`                //用户表id
	TerminalEquipmentName string `json:"terminalEquipmentName"` //终端设备名称
	LoginCountry          string `json:"loginCountry"`          //登录国家
	LoginCity             string `json:"loginCity"`             //登录城市
	LoginIp               int64  `json:"loginIp"`               //登录IP
	AddTime               string `json:"addTime"`               //创建时间
}

// UserUsdts 用户USDT地址表
type UserUsdts struct {
	UsdtId      sql.NullInt64  `json:"usdtId"`      //USDT地ID
	UserId      sql.NullInt64  `json:"userId"`      //用户表ID
	AddreType   sql.NullInt32  `json:"addreType"`   //地址类型：1 ERC 2 TRC
	UsdtAddre   sql.NullString `json:"usdtAddre"`   //usdt地址
	UsdtTitle   sql.NullString `json:"usdtTitle"`   //usdt备注名称
	DefaultCard sql.NullInt32  `json:"defaultCard"` //是否默认：0 否  1 默认
	AddTime     sql.NullString `json:"addTime"`     //添加时间
}

// UserUsdtsJson 用户USDT地址表
type UserUsdtsJson struct {
	UsdtId      int64  `json:"usdtId"`      //USDT地ID
	UserId      int64  `json:"userId"`      //用户表ID
	AddreType   int32  `json:"addreType"`   //地址类型：1 ERC 2 TRC
	UsdtAddre   string `json:"usdtAddre"`   //usdt地址
	UsdtTitle   string `json:"usdtTitle"`   //usdt备注名称
	DefaultCard int32  `json:"defaultCard"` //是否默认：0 否  1 默认
	AddTime     string `json:"addTime"`     //添加时间
}

// UserFundAccounts 资金账户
type UserFundAccounts struct {
	Id              sql.NullInt64   `json:"id"`              //资金账户id
	UserId          sql.NullInt64   `json:"userId"`          //用户id
	FundAccount     sql.NullString  `json:"fundAccount"`     //资金账号
	FundAccountType sql.NullInt32   `json:"fundAccountType"` //资金账号类型：1 现货资产 2 合约资产 3 马股资产 4 美股资产 5 日股资产 6 佣金账户
	Money           sql.NullFloat64 `json:"money"`           //账户金额
	Unit            sql.NullString  `json:"unit"`            //单位
	AddTime         sql.NullString  `json:"addTime"`         //创建时间
	UpdateTime      sql.NullString  `json:"updateTime"`      //更新时间
}

// UserFundAccountsJson 资金账户
type UserFundAccountsJson struct {
	Id              int64   `json:"id"`              //资金账户id
	UserId          int64   `json:"userId"`          //用户id
	FundAccount     string  `json:"fundAccount"`     //资金账号
	FundAccountType int32   `json:"fundAccountType"` //资金账号类型：1 现货资产 2 合约资产 3 马股资产 4 美股资产 5 日股资产 6 佣金账户
	Money           float64 `json:"money"`           //账户金额
	Unit            string  `json:"unit"`            //单位
	AddTime         string  `json:"addTime"`         //创建时间
	UpdateTime      string  `json:"updateTime"`      //更新时间
}

// UserBankcards 用户银行卡表
type UserBankcards struct {
	BankCardId       sql.NullInt64  `json:"bankCardId"`       //银行卡id
	UserId           sql.NullInt64  `json:"userId"`           //用户表id
	ActualName       sql.NullString `json:"actualName"`       //真实姓名
	BankCountryCode  sql.NullString `json:"bankCountryCode"`  //国家代码
	BankCountryName  sql.NullString `json:"bankCountryName"`  //国家名称
	BankName         sql.NullString `json:"bankName"`         //银行名称
	BankCode         sql.NullString `json:"bankCode"`         //银行代码
	BankPhoneNumber  sql.NullString `json:"bankPhoneNumber"`  //手机号
	BankCardNumber   sql.NullString `json:"bankCardNumber"`   //卡号
	BankCardEmail    sql.NullString `json:"bankCardEmail"`    //邮箱
	DefaultBankcards sql.NullInt32  `json:"defaultBankcards"` //是否默认：1.默认，0.不默认
	AddTime          sql.NullString `json:"addTime"`          //创建时间
	UpdateTime       sql.NullString `json:"updateTime"`       //更新时间
}

// UserBankcardsJson 用户银行卡表
type UserBankcardsJson struct {
	BankCardId       int64  `json:"bankCardId"`       //银行卡id
	UserId           int64  `json:"userId"`           //用户表id
	ActualName       string `json:"actualName"`       //真实姓名
	BankCountryCode  string `json:"bankCountryCode"`  //国家代码
	BankCountryName  string `json:"bankCountryName"`  //国家名称
	BankName         string `json:"bankName"`         //银行名称
	BankCode         string `json:"bankCode"`         //银行代码
	BankPhoneNumber  string `json:"bankPhoneNumber"`  //手机号
	BankCardNumber   string `json:"bankCardNumber"`   //卡号
	BankCardEmail    string `json:"bankCardEmail"`    //邮箱
	DefaultBankcards int32  `json:"defaultBankcards"` //是否默认：1.默认，0.不默认
	AddTime          string `json:"addTime"`          //创建时间
	UpdateTime       string `json:"updateTime"`       //更新时间
}

func init() {
	//初始化配置方式
	config.LoadConfig()
	inits.InitLog()

	//redis初始化方式
	client = inits.InitRedis()

	file = "daemon_users_init_redis"

	//多语言
	if err := translators.InitTranslators("zh"); err != nil {
		fmt.Sprintf("init Translator failed,err:%v\n", err)
		logger.Warn(fmt.Sprintf("file:%s  Failed:[init Translator failed] Time:[%s] Err:[%v]", file, time.Now().Format(time.DateTime), err))
	}
}

func initCache() {
	//mysql初始化方式
	conn = inits.InitMysql()
	defer conn.Close() //关闭数据库连接

	//获取users总数
	sqlStr := "select count(1) as count from " + TableUsers + " where deleteTime is null"
	rows, err := conn.Query(sqlStr)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	count := 0
	for rows.Next() {
		rows.Scan(&count)
		break
	}
	defer rows.Close()

	for i := 0; i < count; i++ {
		sqlStr = "select id,uid,countryCode,phoneNumber,email,loginPassword,tradePassword,surname,name,sex,birthday,country,nickName,avatar,isRealName,inviteCode,lastLoginTime,status,accessToken,addTime,updateTime from " + TableUsers + " where id>? and deleteTime is null order by addTime asc limit ?"
		rows, err = conn.Query(sqlStr, i*1000, 1000)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer rows.Close()

		var usersMap []*UsersJson
		for rows.Next() {
			var user Users
			var userJson UsersJson
			rows.Scan(&user.Id, &user.Uid, &user.CountryCode, &user.PhoneNumber, &user.Email, &user.LoginPassword, &user.TradePassword, &user.Surname, &user.Name, &user.Sex, &user.Birthday, &user.Country, &user.NickName, &user.Avatar, &user.IsRealName, &user.InviteCode, &user.LastLoginTime, &user.Status, &user.AccessToken, &user.AddTime, &user.UpdateTime)
			userJson.Id = user.Id.Int64
			userJson.Uid = user.Uid.String
			userJson.CountryCode = user.CountryCode.String
			userJson.PhoneNumber = user.PhoneNumber.Int64
			userJson.Email = user.Email.String
			userJson.LoginPassword = user.LoginPassword.String
			userJson.TradePassword = user.TradePassword.String
			userJson.Surname = user.Surname.String
			userJson.Name = user.Name.String
			userJson.Sex = user.Sex.Int32
			userJson.Birthday = user.Birthday.String
			userJson.Country = user.Country.String
			userJson.NickName = user.NickName.String
			userJson.Avatar = user.Avatar.String
			userJson.IsRealName = user.IsRealName.Int32
			userJson.InviteCode = user.InviteCode.String
			userJson.LastLoginTime = user.LastLoginTime.String
			userJson.Status = user.Status.Int32
			userJson.AccessToken = user.AccessToken.String
			userJson.AddTime = user.AddTime.String
			userJson.UpdateTime = user.UpdateTime.String
			usersMap = append(usersMap, &userJson)
		}
		//无数据跳出
		if len(usersMap) == 0 {
			return
		}

		for i2, users := range usersMap {
			//缓存默认设置
			//设置用户USDT地址缓存-TRC、ERC
			usdtsDefaultErcKey := fmt.Sprintf("%s%v", constant.UsdtsDefaultErcKey, users.Id)
			usdtsDefaultTrcKey := fmt.Sprintf("%s%v", constant.UsdtsDefaultTrcKey, users.Id)

			userFundAccountsKey1 := fmt.Sprintf("%s1_%v", constant.UserFundAccountsKey, users.Id)
			userFundAccountsKey2 := fmt.Sprintf("%s2_%v", constant.UserFundAccountsKey, users.Id)
			userFundAccountsKey3 := fmt.Sprintf("%s3_%v", constant.UserFundAccountsKey, users.Id)
			userFundAccountsKey4 := fmt.Sprintf("%s4_%v", constant.UserFundAccountsKey, users.Id)
			userFundAccountsKey5 := fmt.Sprintf("%s5_%v", constant.UserFundAccountsKey, users.Id)
			userFundAccountsKey6 := fmt.Sprintf("%s6_%v", constant.UserFundAccountsKey, users.Id)

			bankcardsDefaultKey := fmt.Sprintf("%s%v", constant.BankcardsDefaultKey, users.Id)

			err = client.MSet(context.Background(), usdtsDefaultErcKey, nil, usdtsDefaultTrcKey, nil, userFundAccountsKey1, nil, userFundAccountsKey2, nil, userFundAccountsKey3, nil, userFundAccountsKey4, nil, userFundAccountsKey5, nil, userFundAccountsKey6, nil, bankcardsDefaultKey, nil).Err()
			if err != nil {
				fmt.Sprintf("缓存设置失败。 userId:%d", users.Id)
				fmt.Println(err.Error())
				return
			}

			//获取用户终端设备表
			sqlStr = "select " + UserTerminalEquipmentsField + " from " + TableUserTerminalEquipments + " where userId=?"
			rows, err = conn.Query(sqlStr, users.Id)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			defer rows.Close()
			for rows.Next() {
				var userTE UserTerminalEquipments
				var userTEJson UserTerminalEquipmentsJson
				rows.Scan(&userTE.TerminalEquipmentId, &userTE.UserId, &userTE.TerminalEquipmentName, &userTE.LoginCountry, &userTE.LoginCity, &userTE.LoginIp, &userTE.AddTime)
				userTEJson.TerminalEquipmentId = userTE.TerminalEquipmentId.Int64
				userTEJson.UserId = userTE.UserId.Int64
				userTEJson.TerminalEquipmentName = userTE.TerminalEquipmentName.String
				userTEJson.LoginCountry = userTE.LoginCountry.String
				userTEJson.LoginCity = userTE.LoginCity.String
				userTEJson.LoginIp = userTE.LoginIp.Int64
				userTEJson.AddTime = userTE.AddTime.String
				//赋值
				if userTE.TerminalEquipmentId.Int64 > 0 {
					usersMap[i2].UserTerminalEquipments = append(usersMap[i2].UserTerminalEquipments, &userTEJson)
				} else {
					usersMap[i2].UserTerminalEquipments = nil
				}
			}

			//获取用户USDT地址
			sqlStr = "select " + UserUsdtsField + " from " + TableUserUsdts + " where userId=? and deleteTime is null"
			rows, err = conn.Query(sqlStr, users.Id)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			defer rows.Close()
			for rows.Next() {
				var userUsdts UserUsdts
				var userUsdtsJson UserUsdtsJson
				rows.Scan(&userUsdts.UsdtId, &userUsdts.UserId, &userUsdts.AddreType, &userUsdts.UsdtAddre, &userUsdts.UsdtTitle, &userUsdts.DefaultCard, &userUsdts.AddTime)
				userUsdtsJson.UsdtId = userUsdts.UsdtId.Int64
				userUsdtsJson.UserId = userUsdts.UserId.Int64
				userUsdtsJson.AddreType = userUsdts.AddreType.Int32
				userUsdtsJson.UsdtAddre = userUsdts.UsdtAddre.String
				userUsdtsJson.UsdtTitle = userUsdts.UsdtTitle.String
				userUsdtsJson.DefaultCard = userUsdts.DefaultCard.Int32
				userUsdtsJson.AddTime = userUsdts.AddTime.String

				if userUsdts.AddreType.Int32 == 1 && userUsdts.DefaultCard.Int32 == 1 {
					//设置用户USDT地址缓存-TRC
					userUsdtsMarshal, _ := json.Marshal(&userUsdtsJson)
					usdtsDefaultErcKey = fmt.Sprintf("%s%v", constant.UsdtsDefaultErcKey, users.Id)
					err = client.Set(context.Background(), usdtsDefaultErcKey, userUsdtsMarshal, 0).Err()
					if err != nil {
						fmt.Sprintf("缓存设置失败。 userId:%d", users.Id)
						fmt.Println(err.Error())
						return
					}
				}

				if userUsdts.AddreType.Int32 == 2 && userUsdts.DefaultCard.Int32 == 2 {
					//设置用户USDT地址缓存-ERC
					userUsdtsMarshal, _ := json.Marshal(&userUsdtsJson)
					usdtsDefaultTrcKey = fmt.Sprintf("%s%v", constant.UsdtsDefaultTrcKey, users.Id)
					err = client.Set(context.Background(), usdtsDefaultTrcKey, userUsdtsMarshal, 0).Err()
					if err != nil {
						fmt.Sprintf("缓存设置失败。 userId:%d", users.Id)
						fmt.Println(err.Error())
						return
					}
				}
			}

			//资金账户
			sqlStr = "select " + UserFundAccountsField + " from " + TableUserFundAccounts + " where userId=?"
			rows, err = conn.Query(sqlStr, users.Id)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			defer rows.Close()
			for rows.Next() {
				var userFundAccounts UserFundAccounts
				var userFundAccountsJson UserFundAccountsJson
				rows.Scan(&userFundAccounts.Id, &userFundAccounts.UserId, &userFundAccounts.FundAccountType, &userFundAccounts.FundAccount, &userFundAccounts.Money, &userFundAccounts.Unit, &userFundAccounts.AddTime, &userFundAccounts.UpdateTime)
				userFundAccountsJson.Id = userFundAccounts.Id.Int64
				userFundAccountsJson.UserId = userFundAccounts.UserId.Int64
				userFundAccountsJson.FundAccount = userFundAccounts.FundAccount.String
				userFundAccountsJson.FundAccountType = userFundAccounts.FundAccountType.Int32
				userFundAccountsJson.Money = userFundAccounts.Money.Float64
				userFundAccountsJson.Unit = userFundAccounts.Unit.String
				userFundAccountsJson.AddTime = userFundAccounts.AddTime.String
				userFundAccountsJson.UpdateTime = userFundAccounts.UpdateTime.String
				//赋值
				if userFundAccounts.Id.Int64 > 0 {
					usersMap[i2].UserFundAccounts = append(usersMap[i2].UserFundAccounts, &userFundAccountsJson)
				} else {
					usersMap[i2].UserFundAccounts = nil
				}

				//设置用户资金账户缓存
				switch int(userFundAccounts.FundAccountType.Int32) {
				case 1:
					userFundAccountsMarshal, _ := json.Marshal(&userFundAccountsJson)
					fmt.Println(userFundAccountsKey1, userFundAccountsMarshal)
					err = client.Set(context.Background(), userFundAccountsKey1, userFundAccountsMarshal, 0).Err()
					if err != nil {
						fmt.Sprintf("缓存设置失败。 userId:%d", users.Id)
						fmt.Println(err.Error())
						return
					}
				case 2:
					userFundAccountsMarshal, _ := json.Marshal(&userFundAccountsJson)
					err = client.Set(context.Background(), userFundAccountsKey2, userFundAccountsMarshal, 0).Err()
					if err != nil {
						fmt.Sprintf("缓存设置失败。 userId:%d", users.Id)
						fmt.Println(err.Error())
						return
					}
				case 3:
					userFundAccountsMarshal, _ := json.Marshal(&userFundAccountsJson)
					err = client.Set(context.Background(), userFundAccountsKey3, userFundAccountsMarshal, 0).Err()
					if err != nil {
						fmt.Sprintf("缓存设置失败。 userId:%d", users.Id)
						fmt.Println(err.Error())
						return
					}
				case 4:
					userFundAccountsMarshal, _ := json.Marshal(&userFundAccountsJson)
					err = client.Set(context.Background(), userFundAccountsKey4, userFundAccountsMarshal, 0).Err()
					if err != nil {
						fmt.Sprintf("缓存设置失败。 userId:%d", users.Id)
						fmt.Println(err.Error())
						return
					}
				case 5:
					userFundAccountsMarshal, _ := json.Marshal(&userFundAccountsJson)
					err = client.Set(context.Background(), userFundAccountsKey5, userFundAccountsMarshal, 0).Err()
					if err != nil {
						fmt.Sprintf("缓存设置失败。 userId:%d", users.Id)
						fmt.Println(err.Error())
						return
					}
				case 6:
					userFundAccountsMarshal, _ := json.Marshal(&userFundAccountsJson)
					err = client.Set(context.Background(), userFundAccountsKey6, userFundAccountsMarshal, 0).Err()
					if err != nil {
						fmt.Sprintf("缓存设置失败。 userId:%d", users.Id)
						fmt.Println(err.Error())
						return
					}
				}
			}

			//用户银行卡表
			sqlStr = "select " + UserBankcardsField + " from " + TableUserBankcards + " where userId=? and deleteTime is null"
			rows, err = conn.Query(sqlStr, users.Id)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			defer rows.Close()
			for rows.Next() {
				var userBankcards UserBankcards
				var userBankcardsJson UserBankcardsJson
				rows.Scan(&userBankcards.BankCardId, &userBankcards.UserId, &userBankcards.ActualName, &userBankcards.BankCountryCode, &userBankcards.BankCountryName, &userBankcards.BankName, &userBankcards.BankCode, &userBankcards.BankPhoneNumber, &userBankcards.BankCardNumber, &userBankcards.BankCardEmail, &userBankcards.DefaultBankcards, &userBankcards.AddTime, &userBankcards.UpdateTime)

				userBankcardsJson.BankCardId = userBankcards.BankCardId.Int64
				userBankcardsJson.UserId = userBankcards.UserId.Int64
				userBankcardsJson.ActualName = userBankcards.ActualName.String
				userBankcardsJson.BankCountryCode = userBankcards.BankCountryCode.String
				userBankcardsJson.BankCountryName = userBankcards.BankCountryName.String
				userBankcardsJson.BankName = userBankcards.BankName.String
				userBankcardsJson.BankCode = userBankcards.BankCode.String
				userBankcardsJson.BankPhoneNumber = userBankcards.BankPhoneNumber.String
				userBankcardsJson.BankCardNumber = userBankcards.BankCardNumber.String
				userBankcardsJson.BankCardEmail = userBankcards.BankCardEmail.String
				userBankcardsJson.DefaultBankcards = userBankcards.DefaultBankcards.Int32
				userBankcardsJson.AddTime = userBankcards.AddTime.String
				userBankcardsJson.UpdateTime = userBankcards.UpdateTime.String

				if userBankcards.DefaultBankcards.Int32 == 1 {
					userBankcardsMarshal, _ := json.Marshal(&userBankcardsJson)
					err = client.Set(context.Background(), bankcardsDefaultKey, userBankcardsMarshal, 0).Err()
					if err != nil {
						fmt.Sprintf("缓存设置失败。 userId:%d", users.Id)
						fmt.Println(err.Error())
						return
					}
				}
			}

			//1、设置userInfo缓存
			userInfo, _ := json.Marshal(usersMap[i2])
			userInfoKey := fmt.Sprintf("%s%v", constant.UserInfoUidKey, usersMap[i2].Id)
			err = client.Set(context.Background(), userInfoKey, userInfo, 0).Err()
			if err != nil {
				fmt.Sprintf("缓存设置失败。 userId:%d", users.Id)
				fmt.Println(err.Error())
				return
			}
		}
	}

}

// 资金子账号-现货-热更新脚本
func initCacheSpots() {
	//mysql初始化方式
	conn = inits.InitMysql()
	defer conn.Close() //关闭数据库连接

	client = inits.InitRedis()
	fmt.Println(66666)
	//获取现货子账号
	sqlStr := "select id,fundAccountsId,userId,fundAccountSpot,fundAccountName,spotMoney,unit from bo_user_fund_account_spots where userId=?"
	rows, err := conn.Query(sqlStr, 1)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var spotsJsonMap []*models.UserFundAccountSpotsJson
	for rows.Next() {
		var spotsJson models.UserFundAccountSpotsJson
		var spots models.UserFundAccountSpots
		rows.Scan(&spots.Id, &spots.FundAccountsId, &spots.UserId, &spots.FundAccountSpot, &spots.FundAccountName, &spots.SpotMoney, &spots.Unit)
		spotsJson = models.ConvUserFundAccountSpotsNull2Json(spots, spotsJson)
		spotsJsonMap = append(spotsJsonMap, &spotsJson)
		userFundAccountSpotsKey := fmt.Sprintf("user_fund_account_spots_%v", spotsJson.UserId) //spotsJson.UserId
		err := client.HSet(context.Background(), userFundAccountSpotsKey, fmt.Sprintf("%susdt", spotsJson.FundAccountName), spotsJson.SpotMoney).Err()
		if err != nil {
			fmt.Printf("提交失败 Err:%s", err.Error())
			return
		}
	}
	defer rows.Close()
	fmt.Println("设置成功")
}

func redisZadd() {
	spotsKey := fmt.Sprintf("user_fund_account_swaps_%v", 1)
	//err := client.ZAdd(context.Background(), spotsKey, redis.Z{Score: 0, Member: "1"}, redis.Z{Score: 0, Member: "2"}, redis.Z{Score: 0, Member: "3"}, redis.Z{Score: 0, Member: "4"}, redis.Z{Score: 0, Member: "5"}, redis.Z{Score: 0, Member: "6"}, redis.Z{Score: 0, Member: "7"}, redis.Z{Score: 0, Member: "8"}, redis.Z{Score: 0, Member: "9"}, redis.Z{Score: 0, Member: "10"}, redis.Z{Score: 0, Member: "11"}, redis.Z{Score: 0, Member: "12"}).Err()
	//if err != nil {
	//	fmt.Println("zadd失败")
	//	return [1 0 0 6.3154 0 0 0 0 0 0 0 0]
	//}
	//fmt.Println("zadd成功")
	result, _ := client.HVals(context.Background(), spotsKey).Result()
	var sum float64
	for _, str := range result {
		float, _ := strconv.ParseFloat(str, 64)
		sum += float
	}
	fmt.Println(sum)
}

func main() {
	//将user相关表放进缓存脚本
	initCache()
	// 资金子账号-现货-热更新脚本
	initCacheSpots()
}
