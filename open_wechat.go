package main

import (
	"fmt"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/officialaccount/config"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"github.com/sirupsen/logrus"
	"net/http"
)

type OpenWechatHandler struct {
}

func (h OpenWechatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wc := wechat.NewWechat()
	memory := cache.NewMemory()
	cfg := &config.Config{
		AppID:          GlobalConfig.OpenWechat.AppID,
		AppSecret:      GlobalConfig.OpenWechat.AppSecret,
		Token:          GlobalConfig.OpenWechat.Token,
		EncodingAESKey: GlobalConfig.OpenWechat.EncodingAESKey,
		Cache:          memory,
	}
	officialAccount := wc.GetOfficialAccount(cfg)

	// 传入request和responseWriter
	server := officialAccount.GetServer(r, w)
	// 设置接收消息的处理方法
	server.SetMessageHandler(handleMessage)

	// 处理消息接收以及回复
	err := server.Serve()
	if err != nil {
		fmt.Println(err)
		return
	}
	// 发送回复的消息
	server.Send()
}
func handleMessage(msg message.MixMessage) *message.Reply {
	data := SendMsgData{
		MessageType: "private",
		UserId:      string(msg.FromUserName),
		GroupId:     "",
		Message:     make([]Message, 0, 5),
		AutoEscape:  false,
		ReceivedMsg: msg.Content,
	}
	err := data.AddAIPrompts("private")
	if err != nil {
		logrus.Errorf("add AI prompts error: %s", err)
		return nil
	}
	err = data.AIChat("private")
	if err != nil {
		logrus.Errorf("AI chat error: %s", err)
		return nil
	}

	openWechatReply := message.Reply{
		MsgType: message.MsgTypeText,
		MsgData: message.NewText(fmt.Sprintf("%v", data.Message[0].Data["text"])),
	}

	return &openWechatReply
}
