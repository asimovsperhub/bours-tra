package main

import (
	"Bourse/common/inits"
	"Bourse/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

type JsonData struct {
	Code int `json:"code"`
	Data []*struct {
		Code     string `json:"Code"`
		Name     string `json:"Name"`
		Country  string `json:"Country"`
		Exchange string `json:"Exchange"`
		Currency string `json:"Currency"`
		Type     string `json:"Type"`
		Isin     string `json:"Isin"`
	} `json:"data"`
	Message string `json:"message"`
}

func init() {
	//初始化配置方式
	config.LoadConfig()
	inits.InitLog()

}

func Import() {
	/*num := 0
	max := 0
	for i := 0; i < 50439; i++ {
		num = num + 1
		if i%100 == 0 && i > 0 {
			num = num + 1
			//fmt.Printf("  i=%d ---- num=%d  ", i, num)
			num = 0
			max = max + 1
		}

		if i == 50438 {
			max = max + 1
		}
	}
	fmt.Printf("max:%d", max)
	return*/
	mysqlDb := inits.InitMysql()
	defer mysqlDb.Close()
	stateTime := time.Now().Unix()
	jsonFile, err := ioutil.ReadFile("D:\\go\\src\\Bourse\\daemon\\daemon_socket_us_json_import\\etc\\socket-us.json")
	if err != nil {
		fmt.Println("打开文件失败")
		return
	}
	var jsonData JsonData
	err = json.Unmarshal(jsonFile, &jsonData)
	if err != nil {
		fmt.Println("解析JSON失败")
		return
	}
	//写入数据库 50439
	//num := len(jsonData.Data)
	//strSql := "insert into bo_system_stocks_us (code,name,country,exchange,currency,type,isin) values "
	//values := ""
	nums := 0
	//max := 0
	for i, data := range jsonData.Data {
		nums = nums + 1
		strSql := fmt.Sprintf("insert into bo_system_stocks_us (code,name,country,exchange,currency,type,isin) values ('%s','%s','%s','%s','%s','%s','%s')", data.Code, data.Name, data.Country, data.Exchange, data.Currency, data.Type, data.Isin)
		//strSql = strSql + values
		exec, err := mysqlDb.Exec(strSql)
		if err != nil {
			fmt.Printf("写入失败 Err:[%s] 第几行出错：%v", err.Error(), i+1)
			return
		}
		id, err := exec.LastInsertId()
		if err != nil {
			fmt.Printf("LastInsertId 写入失败 Err:[%s] LastInsertId:[%d] 第几行出错：%d", err.Error(), id, i+1)
			return
		}

		//if nums == 100 {
		//	values = fmt.Sprintf("%s('%s','%s','%s','%s','%s','%s','%s')", values, data.Code, data.Name, data.Country, data.Exchange, data.Currency, data.Type, data.Isin)
		//} else {
		//	values = fmt.Sprintf("%s('%s','%s','%s','%s','%s','%s','%s'),", values, data.Code, data.Name, data.Country, data.Exchange, data.Currency, data.Type, data.Isin)
		//}
		/*if i == 3 {
			strSql = strSql + values
			fmt.Println(strSql)
			return
		}*/

		//for i := 0; i < 50439; i++ {
		//num = num + 1
		/*if i%100 == 0 && i > 0 {
			nums = nums + 1
			strSql = strSql + values
			exec, err := mysqlDb.Exec(strSql)
			if err != nil {
				fmt.Printf("写入失败 Err:[%s] 第几行出错：%v", err.Error(), i+1)
				return
			}
			id, err := exec.LastInsertId()
			if err != nil || id != 1 {
				fmt.Printf("LastInsertId 写入失败 Err:[%s] LastInsertId:[%d] 第几行出错：%v", err.Error(), id, i+1)
				return
			}
			nums = 0
			max = max + 1
			values = ""
		}

		if i == 50438 {
			max = max + 1
			strSql = strSql + values
			exec, err := mysqlDb.Exec(strSql)
			if err != nil {
				fmt.Printf("写入失败 Err:[%s] SQL:[%s] 第几行出错：%v", err.Error(), strSql, i+1)
				return
			}
			id, err := exec.LastInsertId()
			if err != nil || id != 1 {
				fmt.Printf("LastInsertId 写入失败 Err:[%s] SQL:[%s] LastInsertId:[%d] 第几行出错：%v", err.Error(), strSql, id, i+1)
				return
			}
			values = ""
		}*/
		//}
	}

	endTime := time.Now().Unix()
	times := endTime - stateTime
	fmt.Printf("操作成功，共写入%d行数据。 Time:%d", nums, times)

	return
}

func main() {
	Import()
}
