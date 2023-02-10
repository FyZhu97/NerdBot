package main

import (
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

func DailyGreetings() {
	t := time.NewTimer(SetTime(7, 0, 0))
	defer t.Stop()
	for {
		select {
		case <-t.C:
			t.Reset(time.Hour * 24)
			SendGreetings()
		}
	}
}

func SendGreetings() {
	friendInfos, err := GetFriendList()
	if err != nil {
		logrus.Error("[DailyGreetings]get friend list fail: " + err.Error())
	}
	groupInfos, err := GetGroupList()
	if err != nil {
		logrus.Error("[DailyGreetings]get group list fail: " + err.Error())
	}
	sender := SendMsgData{
		Message:    []Message{globalConfig.Greeting.GreetingMessage},
		AutoEscape: false,
	}
	for _, friendInfo := range friendInfos {
		sender.MessageType = "private"
		sender.UserId = friendInfo.UserId
		err := sender.Send()
		if err != nil {
			logrus.Error("[DailyGreetings]send greeting to friend " + strconv.FormatInt(sender.UserId, 10) + " fail: " + err.Error())
		}
	}
	for _, groupInfo := range groupInfos {
		sender.MessageType = "group"
		sender.GroupId = groupInfo.GroupId
		err := sender.Send()
		if err != nil {
			logrus.Error("[DailyGreetings]send greeting to group " + strconv.FormatInt(sender.UserId, 10) + " fail: " + err.Error())
		}
	}
}
