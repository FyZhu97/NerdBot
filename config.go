package main

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
)

type OneBot11Config struct {
	AccessToken      string `yaml:"accessToken"`
	ServerUrl        string `yaml:"serverUrl"`
	SelfId           int64  `yaml:"-"`
	HeartbeatTimeOut int    `yaml:"heartbeatTimeOut" comment:"heartbeat的超时时间"`
}

type OpenAIConfig struct {
	ChatAIUrl            string         `yaml:"chatAIUrl" comment:"调用API的URL"`
	APIKey               string         `yaml:"APIKey"`
	Model                string         `yaml:"model"`
	ResponseMaxTokens    int            `yaml:"responseMaxTokens" comment:"AI回复内容的最大token数量"`
	GroupChatMaxTokens   int            `yaml:"groupChatMaxTokens" comment:"群聊模式下全部prompts的最大token数量"`
	PrivateChatMaxTokens int            `yaml:"privateChatMaxTokens" comment:"非群聊模式下全部prompts的最大token数量"`
	EnableGroupChat      map[int64]bool `yaml:"-"`
	DefaultTemperature   float64        `yaml:"defaultTemperature"`
	InitialPrompts       string         `yaml:"initialPrompts" comment:"初始化AI设定的prompts"`
	MinInterval          float64        `yaml:"minInterval" comment:"最短API调用间隔"`
}

type RedisConfig struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	Database int    `yaml:"database"`
}

type GreetingConfig struct {
	EnableGreeting  bool    `yaml:"enableGreeting"`
	GreetingMessage Message `yaml:"greetingMessage" comment:"打招呼信息，遵循Onebot的message标准"`
}

type ServerConfig struct {
	Address  string  `yaml:"address"`
	AdminIds []int64 `yaml:"adminIds" comment:"管理员帐号ID"`
}

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	OneBot11 OneBot11Config `yaml:"oneBot11"`
	AI       OpenAIConfig   `yaml:"openAI"`
	Redis    RedisConfig    `yaml:"redis"`
	Greeting GreetingConfig `yaml:"greeting"`
}

var GlobalConfig *Config

func InitGlobalConfig() error {
	GlobalConfig = &Config{
		Server: ServerConfig{
			Address:  "0.0.0.0:5701",
			AdminIds: []int64{123456},
		},
		OneBot11: OneBot11Config{
			AccessToken:      "",
			ServerUrl:        "http://0.0.0.0:5700/",
			HeartbeatTimeOut: 20,
		},
		AI: OpenAIConfig{
			ChatAIUrl:            "https://api.openai.com/v1/completions",
			APIKey:               "YOUR_API_KEY",
			Model:                "text-davinci-003",
			ResponseMaxTokens:    500,
			GroupChatMaxTokens:   2000,
			PrivateChatMaxTokens: 4000,
			DefaultTemperature:   0.9,
			InitialPrompts:       "",
			MinInterval:          1,
		},
		Redis: RedisConfig{
			Address:  "127.0.0.1:6379",
			Password: "",
			Database: 0,
		},
		Greeting: GreetingConfig{
			EnableGreeting: false,
			GreetingMessage: Message{
				Type: "text",
				Data: map[string]interface{}{
					"text": "",
				},
			},
		},
	}
	yamlFile, err := os.ReadFile("config.yaml")
	if err == nil {
		err = yaml.Unmarshal(yamlFile, GlobalConfig)
		if err != nil {
			return err
		}
		GlobalConfig.AI.EnableGroupChat = make(map[int64]bool)
	} else {
		// Save default config to YAML file if it does not exist
		yamlData, err := yaml.Marshal(GlobalConfig)
		if err != nil {
			return err
		}
		err = os.WriteFile("config.yaml", yamlData, 0644)
		if err != nil {
			return err
		}
		logrus.Warning("Config file does not exist: Creating a default one. Please edit configuration and restart.")
		os.Exit(0)
	}
	return nil
}
