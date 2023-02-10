package main

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type FriendInfoData struct {
	UserId   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Remark   string `json:"remark"`
}

type GroupInfoData struct {
	GroupId         int64  `json:"group_id"`
	GroupName       string `json:"group_name"`
	GroupMemo       string `json:"group_memo"`
	GroupCreateTime uint32 `json:"group_create_time"`
	GroupLevel      uint32 `json:"group_level"`
	MemberCount     int32  `json:"member_count"`
	MaxMemberCount  int32  `json:"max_member_count"`
}

type FriendList struct {
	Retcode int64            `json:"retcode"`
	Status  string           `json:"status"`
	Data    []FriendInfoData `json:"data"`
}

type GroupList struct {
	Retcode int64           `json:"retcode"`
	Status  string          `json:"status"`
	Data    []GroupInfoData `json:"data"`
}
type GroupInfo struct {
	Retcode int64         `json:"retcode"`
	Status  string        `json:"status"`
	Data    GroupInfoData `json:"data"`
}
type LoginInfo struct {
	Retcode int64  `json:"retcode"`
	Status  string `json:"status"`
	Data    struct {
		UserId   int64  `json:"user_id"`
		Nickname string `json:"nickname"`
	} `json:"data"`
}
type GroupMemberInfo struct {
	Retcode int64  `json:"retcode"`
	Status  string `json:"status"`
	Data    struct {
		UserId   int64  `json:"user_id"`
		Nickname string `json:"nickname"`
		Card     string `json:"card"`
	} `json:"data"`
}

func GetGroupList() ([]GroupInfoData, error) {
	requestUrl := globalConfig.CqHttp.ServerUrl + "get_group_list"
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, err
	}
	resp, err := QQClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	logrus.Info("Get group list success: " + string(body))
	var respData GroupList
	err = json.Unmarshal(body, &respData)
	return respData.Data, err
}

func GetFriendList() ([]FriendInfoData, error) {
	requestUrl := globalConfig.CqHttp.ServerUrl + "get_friend_list"
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, err
	}
	resp, err := QQClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	logrus.Info("Get friend list success: " + string(body))
	var respData FriendList
	err = json.Unmarshal(body, &respData)
	return respData.Data, err
}

func GetLoginInfo() (LoginInfo, error) {
	requestUrl := globalConfig.CqHttp.ServerUrl + "get_login_info"
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return LoginInfo{}, err
	}
	resp, err := QQClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return LoginInfo{}, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return LoginInfo{}, err
	}
	logrus.Info("Get login info success: " + string(body))
	var respData LoginInfo
	err = json.Unmarshal(body, &respData)
	return respData, err
}

func GetGroupMemberInfo(userId int64, groupId int64) (GroupMemberInfo, error) {
	requestUrl := globalConfig.CqHttp.ServerUrl + "get_group_member_info"
	req, err := http.NewRequest("GET", requestUrl, nil)
	q := req.URL.Query()
	q.Add("group_id", strconv.FormatInt(groupId, 10))
	q.Add("user_id", strconv.FormatInt(userId, 10))
	req.URL.RawQuery = q.Encode()
	if err != nil {
		return GroupMemberInfo{}, err
	}
	resp, err := QQClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return GroupMemberInfo{}, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GroupMemberInfo{}, err
	}
	logrus.Info("Get group member info success: " + string(body))
	var respData GroupMemberInfo
	err = json.Unmarshal(body, &respData)
	return respData, err
}

func GetGroupInfo(groupId int64) (GroupInfo, error) {
	requestUrl := globalConfig.CqHttp.ServerUrl + "get_group_info"
	req, err := http.NewRequest("GET", requestUrl, nil)
	q := req.URL.Query()
	q.Add("group_id", strconv.FormatInt(groupId, 10))
	req.URL.RawQuery = q.Encode()
	if err != nil {
		return GroupInfo{}, err
	}
	resp, err := QQClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return GroupInfo{}, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GroupInfo{}, err
	}
	logrus.Info("Get group info success: " + string(body))
	var respData GroupInfo
	err = json.Unmarshal(body, &respData)
	return respData, err
}

func SetTime(hour, min, second int) (d time.Duration) {
	now := time.Now()
	setTime := time.Date(now.Year(), now.Month(), now.Day(), hour, min, second, 0, now.Location())
	d = setTime.Sub(now)
	if d > 0 {
		return
	}
	return d + time.Hour*24
}
func ParseCQCode(cqCode string) ([]Message, string, ReceivedCQTypes) {
	types := ReceivedCQTypes{
		atSelf:   false,
		hasImage: false,
		hasJson:  false,
		hasFace:  false,
		hasReply: false,
	}
	re := regexp.MustCompile(`(\[CQ:([^,\]]+)(?:,([^\]]+))*\])`)
	matches := re.FindAllStringSubmatch(cqCode, -1)
	if len(matches) > 0 {
		var messages []Message
		var parsedCodes []string
		for _, match := range matches {
			message := Message{
				Type: match[2],
				Data: make(map[string]interface{}),
			}
			params := strings.Split(match[3], ",")
			for _, param := range params {
				kv := strings.Split(param, "=")
				if len(kv) == 2 {
					message.Data[kv[0]] = kv[1]
				}
			}
			if message.Type == "at" && message.Data["qq"] == strconv.FormatInt(globalConfig.CqHttp.SelfId, 10) || message.Data["qq"] == "all" {
				types.atSelf = true
			} else if message.Type == "json" {
				types.hasJson = true
			} else if message.Type == "reply" {
				types.hasReply = true
			} else if message.Type == "face" {
				types.hasFace = true
			} else if message.Type == "Image" {
				types.hasImage = true
			}
			messages = append(messages, message)
			parsedCodes = append(parsedCodes, match[1])
			cqCode = strings.Replace(cqCode, match[1], "", 1)
		}
		return messages, cqCode, types
	}
	return nil, cqCode, types
}
func StoreRecord(key string, record *Record) error {
	// Convert the Record struct to JSON
	recordJSON, err := json.Marshal(record)
	if err != nil {
		return err
	}

	// Store the JSON in Redis
	err = Connection.Set(context.Background(), key, recordJSON, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func RetrieveRecord(key string) (*Record, bool, error) {
	exists, err := Connection.Exists(context.Background(), key).Result()
	if err != nil {
		return nil, false, err
	}

	if exists == 0 {
		// The key does not exist
		return nil, false, nil
	}
	// Get the stored JSON from Redis
	recordJSON, err := Connection.Get(context.Background(), key).Bytes()
	if err != nil {
		return nil, false, err
	}

	// Convert the JSON to a Record struct
	var record Record
	err = json.Unmarshal(recordJSON, &record)
	if err != nil {
		return nil, false, err
	}

	return &record, true, err
}
