package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var AIClient *http.Client

type AITokenTransport struct {
	Token string
	Base  http.RoundTripper
}

func (t *AITokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", "Bearer "+t.Token)
	return t.Base.RoundTrip(req)
}

type AIRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

type AIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Text         string      `json:"text"`
		Index        int         `json:"index"`
		LogProbs     interface{} `json:"logprobs"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type Record struct {
	Prompt      string    `json:"prompt"`
	TotalTokens int       `json:"totalTokens"`
	LastRequest time.Time `json:"lastRequest"`
	Temperature float64   `json:"temperature"`
}

func (data SendMsgData) AIChat(mode string) error {
	req := AIRequest{
		Model:       GlobalConfig.AI.Model,
		Prompt:      "",
		MaxTokens:   GlobalConfig.AI.ResponseMaxTokens,
		Temperature: GlobalConfig.AI.DefaultTemperature,
	}
	var id int64
	if mode == "private" {
		id = data.UserId
	} else if mode == "group" {
		id = data.GroupId
	}
	idStr := strconv.FormatInt(id, 10)
	promptRecord, err := RetrieveOrDefaultRecord(idStr)
	if err != nil {
		return fmt.Errorf("retrieve record error: %s", err)
	}
	req.Temperature = promptRecord.Temperature
	promptRecord.Prompt = promptRecord.Prompt + "\n戴便机器人:"
	req.Prompt = promptRecord.Prompt
	AIResp, err := req.GetAIResponseWithRetries(3)
	if err != nil {
		return err
	}
	respText := AIResp.Choices[0].Text
	respText = strings.Replace(respText, "戴便机器人:", "", -1)
	data.Message = append(data.Message, Message{
		Type: "text",
		Data: map[string]interface{}{
			"text": strings.Trim(respText, "\n"),
		},
	})
	err = data.Send()
	if err != nil {
		return err
	}
	promptRecord.Prompt = promptRecord.Prompt + respText
	promptRecord.TotalTokens += AIResp.Usage.TotalTokens
	promptRecord.LastRequest = time.Now()
	err = StoreRecord(idStr, promptRecord)
	return err
}

func (data SendMsgData) AddAIPrompts(mode string) error {
	var userName string
	var id int64
	var maxTokens int
	var groupPrompt = ""
	if mode == "group" {
		memberInfo, err := GetGroupMemberInfo(data.UserId, data.GroupId)
		if err != nil {
			logrus.Error("get group member info fail: ", err)
		}
		if memberInfo.Data.Card != "" {
			userName = memberInfo.Data.Card
		} else {
			userName = memberInfo.Data.Nickname
		}
		id = data.GroupId
		maxTokens = GlobalConfig.AI.GroupChatMaxTokens
		groupPrompt = "AI在一个群聊内，作为一个群成员参与聊天。"
	} else if mode == "private" {
		userName = "用户"
		id = data.UserId
		maxTokens = GlobalConfig.AI.PrivateChatMaxTokens
	} else {
		return errors.New("invalid mode")
	}
	idStr := strconv.FormatInt(id, 10)
	promptRecord, err := RetrieveOrDefaultRecord(idStr)
	if err != nil {
		return fmt.Errorf("retrieve record error: %s", err)
	}
	if promptRecord.Prompt == GlobalConfig.AI.InitialPrompts {
		promptRecord.Prompt = groupPrompt + promptRecord.Prompt
	}
	if time.Now().Sub(promptRecord.LastRequest).Seconds() < GlobalConfig.AI.MinInterval {
		data.Message = append(data.Message, Message{
			Type: "text",
			Data: map[string]interface{}{
				"text": "别急，让我仔细想想[发送频率过快]",
			},
		})
		err := data.Send()
		return err
	}
	if promptRecord.TotalTokens > maxTokens {
		data.Message = append(data.Message, Message{
			Type: "text",
			Data: map[string]interface{}{
				"text": "这个话题聊得太深入了，我们换个话题吧[上下文已清空]",
			},
		})
		DeleteRecord(idStr)
		err = data.Send()
		return err
	}
	promptRecord.Prompt = promptRecord.Prompt + "\n" + userName + ":" + data.ReceivedMsg
	err = StoreRecord(idStr, promptRecord)
	return err
}

func (reqBody *AIRequest) DoAIRequest() (AIResponse, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return AIResponse{}, err
	}
	req, err := http.NewRequest("POST", GlobalConfig.AI.ChatAIUrl, bytes.NewBuffer(body))
	if err != nil {
		return AIResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := AIClient.Do(req)
	if err != nil {
		return AIResponse{}, err
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return AIResponse{}, err
	}
	var AIResp AIResponse
	err = json.Unmarshal(responseBody, &AIResp)
	if err != nil {
		return AIResponse{}, err
	}

	return AIResp, nil
}
func (reqBody *AIRequest) GetAIResponseWithRetries(maxRetries int) (AIResponse, error) {
	var result AIResponse
	var err error
	for i := 0; i < maxRetries; i++ {
		result, err = reqBody.DoAIRequest()
		if err != nil {
			return AIResponse{}, err
		}
		if len(result.Choices) > 0 {
			return result, nil
		}
		time.Sleep(time.Second)
	}
	return AIResponse{}, fmt.Errorf("no successful response after %d retries", maxRetries)
}

func DailyPromptsClear() {
	t := time.NewTimer(SetTime(7, 0, 0))
	defer t.Stop()
	for {
		select {
		case <-t.C:
			t.Reset(time.Hour * 24)
			Connection.FlushAll(context.Background())
		}
	}
}
