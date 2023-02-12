package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io"
	"net/http"
	"os"
)

var (
	globalConfig *Config
	QQClient     *http.Client
	AIClient     *http.Client
	Connection   *redis.Client
)

type QQTokenTransport struct {
	Token string
	Base  http.RoundTripper
}

type AITokenTransport struct {
	Token string
	Base  http.RoundTripper
}

type CqHttp struct {
	AccessToken      string `yaml:"accessToken"`
	ServerUrl        string `yaml:"serverUrl"`
	SelfId           int64  `yaml:"-"`
	HeartbeatTimeOut int    `yaml:"heartbeatTimeOut" comment:"heartbeat的超时时间"`
}

type OpenAI struct {
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

type Config struct {
	Port     string         `yaml:"port" comment:"监听端口，也即为cqhttp的上报端口"`
	AdminIds []int64        `yaml:"adminIds" comment:"管理员ID，目前仅支持QQ号"`
	CqHttp   CqHttp         `yaml:"cqHttp"`
	OpenAI   OpenAI         `yaml:"openAI"`
	Redis    RedisConfig    `yaml:"redisConfig"`
	Greeting GreetingConfig `yaml:"greeting"`
}

func (t *QQTokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", "Bearer "+t.Token)
	return t.Base.RoundTrip(req)
}

func (t *AITokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", "Bearer "+t.Token)
	return t.Base.RoundTrip(req)
}

func main() {
	var err error
	err = initGlobalConfig()
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
	globalConfig.CqHttp.SelfId = loginInfo.Data.UserId
	initRedis()
	defer func() {
		if err := Connection.FlushAll(context.Background()).Err(); err != nil {
			logrus.Fatalf("goredis - failed to flush: %v", err)
		}
		if err := Connection.Close(); err != nil {
			logrus.Fatalf("goredis - failed to communicate to redis-server: %v", err)
		}
	}()
	initOpenAI()
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	r := gin.Default()
	r.POST("/", reply)
	logrus.Info("listening to: ", globalConfig.Port)
	HeartbeatContinue()
	go HeartBeatMonitor()
	if globalConfig.Greeting.EnableGreeting {
		go DailyGreetings()
	}
	go DailyPromptsClear()
	err = r.Run(":" + globalConfig.Port)
	if err != nil {
		logrus.Error("listening port error:", err)
		return
	}
}

func initRedis() {
	Connection = redis.NewClient(&redis.Options{
		Addr:     globalConfig.Redis.Address,
		Password: globalConfig.Redis.Password,
		DB:       globalConfig.Redis.Database,
	})
	pong, err := Connection.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}
	logrus.Info("initiate GoRedis Client success: ", pong)
}

func initHTTPClients() {
	QQClient = &http.Client{
		Transport: &QQTokenTransport{
			Token: globalConfig.CqHttp.AccessToken,
			Base:  http.DefaultTransport,
		},
	}

	AIClient = &http.Client{
		Transport: &AITokenTransport{
			Token: globalConfig.OpenAI.APIKey,
			Base:  http.DefaultTransport,
		},
	}
}

func initGlobalConfig() error {
	globalConfig = &Config{
		Port:     "5701",
		AdminIds: []int64{123456},
		CqHttp: CqHttp{
			AccessToken:      "",
			ServerUrl:        "http://0.0.0.0:5700/",
			HeartbeatTimeOut: 20,
		},
		OpenAI: OpenAI{
			ChatAIUrl:            "https://api.openai.com/v1/completions",
			APIKey:               "YOUR_API_KEY",
			Model:                "text-davinci-003",
			ResponseMaxTokens:    500,
			GroupChatMaxTokens:   2000,
			PrivateChatMaxTokens: 4000,
			DefaultTemperature:   0.5,
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
		err = yaml.Unmarshal(yamlFile, globalConfig)
		if err != nil {
			return err
		}
		globalConfig.OpenAI.EnableGroupChat = make(map[int64]bool)
	} else {
		// Save default config to YAML file if it does not exist
		yamlData, err := yaml.Marshal(globalConfig)
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
