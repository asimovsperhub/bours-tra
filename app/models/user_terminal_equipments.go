package models

import "database/sql"

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

func ConvUserTerminalEquipmentsNull2Json(v UserTerminalEquipments, p UserTerminalEquipmentsJson) UserTerminalEquipmentsJson {
	p.TerminalEquipmentId = v.TerminalEquipmentId.Int64
	p.UserId = v.UserId.Int64
	p.TerminalEquipmentName = v.TerminalEquipmentName.String
	p.LoginCountry = v.LoginCountry.String
	p.LoginCity = v.LoginCity.String
	p.LoginIp = v.LoginIp.Int64
	p.AddTime = v.AddTime.String
	return p
}
