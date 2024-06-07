package websocket

import (
	"Bourse/app/middleware/auth"
	"Bourse/app/servers/socket"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"net/http"
)

var (
	upGrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn    *sql.DB
	redisDb *redis.Client
)

func RunSocket(ctx *gin.Context) {
	//TODO 接收参数验证。。。
	//redisDb = inits.InitRedis()
	//UserInfo = global.GetUserInfo(redisDb)
	//fmt.Println(auth.TokenUserInfo.Uid)

	go socket.MsServer.Start()
	//go socket.MsServer.PushAccountSpotsSubscribe() //推送在线用户现货资金账户变动
	//go socket.MsServer.PushAccountSwapsSubscribe() //推送在线用户合约资金账户变动
	//go socket.MsServer.UserStockHoldingsPush() //推送持仓用户持仓数据变动
	go socket.MsServer.PushStocksUS() //根据美股行情，推送所有在线用户的美股资金账户、持仓数据变动

	ws, err := upGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		fmt.Println("升级ws服务失败")
		return
	}

	//建立client信息
	client := &socket.Client{
		Uid:    auth.TokenUserInfo.Uid,
		UserId: auth.TokenUserInfo.Id,
		Conn:   ws,
		Send:   make(chan *socket.Message),
	}

	//写入client注册管道
	socket.MsServer.Register <- client

	//协程处理读、写
	go client.Read()
	go client.Write()
}
