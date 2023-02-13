package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

func main() {
	var err error
	err = InitGlobalConfig()
	if err != nil {
		logrus.Error("initiate global config fail: ", err)
		return
	}
	logrus.Info("initiate global config success")
	initHTTPClients()
	loginInfo, err := GetLoginInfo()
	if err != nil {
		logrus.Error("initiate login info fail. Please check whether cqhttp is running. Error: ", err)
		return
	}
	GlobalConfig.OneBot11.SelfId = loginInfo.Data.UserId
	InitRedis()
	defer func() {
		if err := Connection.FlushAll(context.Background()).Err(); err != nil {
			logrus.Fatalf("goredis - failed to flush: %v", err)
		}
		if err := Connection.Close(); err != nil {
			logrus.Fatalf("goredis - failed to communicate to redis-server: %v", err)
		}
	}()
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	r := gin.Default()
	r.POST("/", reply)
	if GlobalConfig.OneBot11.HeartbeatTimeOut > 0 {
		HeartbeatContinue()
		go HeartBeatMonitor()
	}
	if GlobalConfig.Greeting.EnableGreeting {
		go DailyGreetings()
	}
	go DailyPromptsClear()
	logrus.Info("listening to: ", GlobalConfig.Server.Address)
	err = r.Run(GlobalConfig.Server.Address)
	if err != nil {
		logrus.Error("listening port error:", err)
		return
	}
}

func initHTTPClients() {
	OneBotClient = &http.Client{
		Transport: &OneBotTokenTransport{
			Token: GlobalConfig.OneBot11.AccessToken,
			Base:  http.DefaultTransport,
		},
	}
	AIClient = &http.Client{
		Transport: &AITokenTransport{
			Token: GlobalConfig.AI.APIKey,
			Base:  http.DefaultTransport,
		},
	}
}
