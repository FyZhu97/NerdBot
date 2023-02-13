package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type SendMsgData struct {
	MessageType string    `json:"message_type"`
	UserId      int64     `json:"user_id"`
	GroupId     int64     `json:"group_id"`
	Message     []Message `json:"message"`
	AutoEscape  bool      `json:"auto_escape"`
	ReceivedMsg string    `json:"-"`
}

type Message struct {
	Type string                 `json:"type" yaml:"type"`
	Data map[string]interface{} `json:"data" yaml:"data"`
}

func (data SendMsgData) Send() error {
	replyUrl := GlobalConfig.OneBot11.ServerUrl + "send_msg"
	bytesData, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	var body = bytes.NewReader(bytesData)
	req, err := http.NewRequest("POST", replyUrl, body)
	defer req.Body.Close()
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	resp, err := OneBotClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New("[Sender] reply to " + data.MessageType + " " + strconv.FormatInt(data.UserId, 10) + " error: " + strconv.FormatInt(int64(resp.StatusCode), 10))
	}
	logrus.Info("[Sender]Send message success: ", data.Message)
	return nil
}
