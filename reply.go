package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
)

type QQMessage struct {
	Interval      int64     `json:"interval"`
	MetaEventType string    `json:"meta_event_type"`
	PostType      string    `json:"post_type"`
	Message       []Message `json:"message"`
	RawMessage    string    `json:"raw_message"`
	//Sender        int64     `json:"sender"`
	SelfId      int64           `json:"self_id"`
	Time        int64           `json:"time"`
	Status      AppStatus       `json:"status"`
	UserId      int64           `json:"user_id"`
	TargetId    int64           `json:"target_id"`
	GroupId     int64           `json:"group_id"`
	MessageType string          `json:"message_type"`
	SubType     string          `json:"sub_type"`
	Font        int64           `json:"font"`
	MessageId   int64           `json:"message_id"`
	CqTypes     ReceivedCQTypes `json:"-"`
}

type AppStatus struct {
	AppEnabled     bool      `json:"app_enabled"`
	AppGood        bool      `json:"app_good"`
	AppInitialized bool      `json:"app_initialized"`
	Good           bool      `json:"good"`
	Online         bool      `json:"online"`
	PluginsGood    bool      `json:"plugins_good"`
	Stat           MsgStatus `json:"stat"`
}
type MsgStatus struct {
	PacketReceived  int64 `json:"packet_received"`
	PacketSent      int64 `json:"packet_sent"`
	PacketLost      int64 `json:"packet_lost"`
	MessageReceived int64 `json:"message_received"`
	MessageSent     int64 `json:"message_sent"`
	DisconnectTimes int64 `json:"disconnect_times"`
	LostTimes       int64 `json:"lost_times"`
	LastMessageTime int64 `json:"last_message_time"`
}
type ReceivedCQTypes struct {
	atSelf   bool
	hasImage bool
	hasJson  bool
	hasFace  bool
	hasReply bool
}

func reply(ctx *gin.Context) {
	var req QQMessage
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if req.MetaEventType == "heartbeat" {
		HeartbeatContinue()
		return
	}
	logrus.Info("Received message: ", req.Message)
	sender := SendMsgData{
		MessageType: req.MessageType,
		UserId:      req.UserId,
		GroupId:     req.GroupId,
		Message:     make([]Message, 0, 5),
		AutoEscape:  false,
		ReceivedMsg: req.RawMessage,
	}
	if strings.HasPrefix(req.RawMessage, "NerdBot ") {
		msg := req.ExecuteCommand()
		sender.Message = append(sender.Message, msg)
		err = sender.Send()
		if err != nil {
			logrus.Error(err)
		}
		return
	}
	cqMessage, remainText, types := ParseCQCode(req.RawMessage)
	req.CqTypes = types
	remainText = strings.Trim(remainText, " ")
	enableGroupChat, ok := globalConfig.OpenAI.EnableGroupChat[req.GroupId]
	var chatMode string
	enableAIReply := true
	if req.MessageType == "group" {
		if ok && enableGroupChat {
			chatMode = "group"
			if !req.CqTypes.atSelf {
				enableAIReply = false
			}
		} else if req.CqTypes.atSelf {
			chatMode = "private"
		}
	} else if req.MessageType == "private" {
		if cqMessage == nil {
			chatMode = "private"
		}
	}
	if chatMode == "private" {
		sender.Message = append(sender.Message, Message{
			Type: "reply",
			Data: map[string]interface{}{
				"id": req.MessageId,
			},
		})
	}
	if chatMode != "" {
		err = sender.AddAIPrompts(chatMode)
		if err != nil {
			logrus.Error("Add AI "+chatMode+" prompts error: ", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if enableAIReply {
			err = sender.AIChat(chatMode)
			if err != nil {
				logrus.Error("AI chat in "+chatMode+" error: ", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}
}

func (req QQMessage) ExecuteCommand() Message {
	var isAdmin = false
	var msg = Message{
		Type: "text",
		Data: map[string]interface{}{
			"text": "",
		},
	}

	enableGroupChat, ok := globalConfig.OpenAI.EnableGroupChat[req.GroupId]
	var id int64
	oriRec := OriRecord
	if req.MessageType == "group" && enableGroupChat {
		id = req.GroupId
		oriRec.Prompt = oriRec.Prompt + "AI在一个群聊内，作为一个群成员参与聊天。"
	} else {
		id = req.UserId
	}
	idStr := strconv.FormatInt(id, 10)

	remainText := strings.Replace(req.RawMessage, "NerdBot ", "", -1)
	remainText = strings.Trim(remainText, " ")
	for _, id := range globalConfig.AdminIds {
		if req.UserId == id {
			isAdmin = true
		}
	}

	if remainText == "clear" {
		err := StoreRecord(idStr, &oriRec)
		if err != nil {
			msg.Data["text"] = fmt.Sprintf("[错误]ID: %d 的上下文清除失败。", id)
		} else {
			msg.Data["text"] = fmt.Sprintf("[通知]ID: %d 的上下文已被清除。", id)
		}
		return msg
	}

	//all the command below need Auth
	if !isAdmin {
		msg.Data["text"] = "[错误]\n对不起，您没有权限执行该命令"
		return msg
	}
	if req.MessageType == "private" {

	} else if req.MessageType == "group" {
		if remainText == "group mode" {
			if !ok || !enableGroupChat {
				globalConfig.OpenAI.EnableGroupChat[req.GroupId] = true
				msg.Data["text"] = "[通知]\n群" + strconv.FormatInt(req.GroupId, 10) + "的群聊模式已开启，之后所有群聊文字信息" +
					"将以同一session供机器人进行分析。如需机器人进行回复，请在输入信息中@戴便机器人。\n注意: 此功能为实验性功能。另，群聊模式可能使用大量token，" +
					"请注意您的token使用量。"

			} else {
				msg.Data["text"] = "[错误]\n群" + strconv.FormatInt(req.GroupId, 10) + "的群聊模式已开启，无须重复操作。"
			}
			return msg
		} else if remainText == "private mode" {
			if ok && enableGroupChat {
				globalConfig.OpenAI.EnableGroupChat[req.GroupId] = false
				msg.Data["text"] = "[通知]\n群聊模式已关闭，机器人将恢复 1 vs 1 对话"
			} else {
				msg.Data["text"] = "[错误]群聊模式已经为关闭状态，无须操作"
			}
			return msg
		}
	}

	if strings.HasPrefix(remainText, "set temperature") {
		str := strings.Replace(remainText, "set temperature", "", -1)
		str = strings.Trim(str, " ")
		temp, err := strconv.ParseFloat(str, 10)
		if err != nil {
			logrus.Error(err)
		}
		if err != nil || temp < 0 || temp > 1 {
			msg.Data["text"] = "[错误]无效的temperature设置，值应该为0~1之间的小数"
			return msg
		}
		record, ok, err := RetrieveRecord(idStr)
		if err != nil {
			msg.Data["text"] = fmt.Sprintf("[错误]temperature参数设置失败:获取记录失败")
		} else {
			if !ok {
				record = &oriRec
			}
			record.Temperature = temp
			err = StoreRecord(idStr, record)
			if err != nil {
				msg.Data["text"] = fmt.Sprintf("[错误]temperature参数设置失败：存储记录失败")
			} else {
				msg.Data["text"] = fmt.Sprintf("[通知]新的temperature参数已生效: %f", temp)
			}
		}
		return msg
	}
	msg.Data["text"] = "[错误]未查询到相应指令"
	return msg
}
