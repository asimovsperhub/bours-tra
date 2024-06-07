package models

import "database/sql"

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
	Id            int64  `json:"id"`            //userId
	Uid           string `json:"uid"`           //UID
	CountryCode   string `json:"countryCode"`   //国家代码
	PhoneNumber   int64  `json:"phoneNumber"`   //手机号
	Email         string `json:"email"`         //电子邮箱
	LoginPassword string `json:"loginPassword"` //登陆密码
	TradePassword string `json:"tradePassword"` //交易密码
	Surname       string `json:"surname"`       //姓
	Name          string `json:"name"`          //名
	Sex           int32  `json:"sex"`           //性别：1男 2女 0未设置
	Birthday      string `json:"birthday"`      //出生日期
	Country       string `json:"country"`       //国家
	NickName      string `json:"nickName"`      //昵称
	Avatar        string `json:"avatar"`        //头像
	IsRealName    int32  `json:"isRealName"`    //是否已实名认证：0未认证 1已认证
	InviteCode    string `json:"inviteCode"`    //邀请码
	LastLoginTime string `json:"lastLoginTime"` //最近一次登陆时间
	Status        int32  `json:"status"`        //状态：1 启用 2禁用 3 黑名单
	AccessToken   string `json:"accessToken"`   //登陆令牌
	AddTime       string `json:"addTime"`       //创建时间
	UpdateTime    string `json:"updateTime"`    //更新时间
}

func ConvUsersNull2Json(v Users, p UsersJson) UsersJson {
	p.Id = v.Id.Int64
	p.Uid = v.Uid.String
	p.CountryCode = v.CountryCode.String
	p.PhoneNumber = v.PhoneNumber.Int64
	p.Email = v.Email.String
	p.LoginPassword = v.LoginPassword.String
	p.TradePassword = v.TradePassword.String
	p.Surname = v.Surname.String
	p.Name = v.Name.String
	p.Sex = v.Sex.Int32
	p.Birthday = v.Birthday.String
	p.Country = v.Country.String
	p.NickName = v.NickName.String
	p.Avatar = v.Avatar.String
	p.IsRealName = v.IsRealName.Int32
	p.InviteCode = v.InviteCode.String
	p.LastLoginTime = v.LastLoginTime.String
	p.Status = v.Status.Int32
	p.AccessToken = v.AccessToken.String
	p.AddTime = v.AddTime.String
	p.UpdateTime = v.UpdateTime.String
	return p
}
