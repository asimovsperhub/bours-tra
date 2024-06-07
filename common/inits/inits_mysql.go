package inits

import (
	"Bourse/config"
	"database/sql"
	"fmt"
	"gitee.com/phper95/pkg/logger"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

func InitMysql() *sql.DB {
	conf := config.Cfg.Mysql
	dns := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local&timeout=30s", conf.User, conf.Password, conf.Host, conf.DBName)
	db, err := sql.Open("mysql", dns)
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed:[数据库连接错误!] dns:[%s] Time:[%s] Err:[%v]", dns, time.Now().Format(time.DateTime), err))
		return nil
	}
	//defer db.Close()

	err = db.Ping()
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed:[数据库连接错误!] dns:[%s] Time:[%s] Err:[%v]", dns, time.Now().Format(time.DateTime), err))
		return nil
	}
	db.SetConnMaxLifetime(time.Second * 30)
	// 设置连接池中的最大闲置连接数，0 表示不会保留闲置。
	db.SetMaxIdleConns(20)
	db.SetMaxOpenConns(20)
	return db
}
