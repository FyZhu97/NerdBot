package main

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
)

var OneBotClient *http.Client

type OneBotTokenTransport struct {
	Token string
	Base  http.RoundTripper
}

func (t *OneBotTokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", "Bearer "+t.Token)
	return t.Base.RoundTrip(req)
}

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
	requestUrl := GlobalConfig.OneBot11.ServerUrl + "get_group_list"
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, err
	}
	resp, err := OneBotClient.Do(req)
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
	requestUrl := GlobalConfig.OneBot11.ServerUrl + "get_friend_list"
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, err
	}
	resp, err := OneBotClient.Do(req)
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
	requestUrl := GlobalConfig.OneBot11.ServerUrl + "get_login_info"
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return LoginInfo{}, err
	}
	resp, err := OneBotClient.Do(req)
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

func GetGroupMemberInfo(userId string, groupId string) (GroupMemberInfo, error) {
	requestUrl := GlobalConfig.OneBot11.ServerUrl + "get_group_member_info"
	req, err := http.NewRequest("GET", requestUrl, nil)
	q := req.URL.Query()
	q.Add("group_id", groupId)
	q.Add("user_id", userId)
	req.URL.RawQuery = q.Encode()
	if err != nil {
		return GroupMemberInfo{}, err
	}
	resp, err := OneBotClient.Do(req)
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
	requestUrl := GlobalConfig.OneBot11.ServerUrl + "get_group_info"
	req, err := http.NewRequest("GET", requestUrl, nil)
	q := req.URL.Query()
	q.Add("group_id", strconv.FormatInt(groupId, 10))
	req.URL.RawQuery = q.Encode()
	if err != nil {
		return GroupInfo{}, err
	}
	resp, err := OneBotClient.Do(req)
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
