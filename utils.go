package main

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

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
			if message.Type == "at" && message.Data["qq"] == strconv.FormatInt(GlobalConfig.OneBot11.SelfId, 10) || message.Data["qq"] == "all" {
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
