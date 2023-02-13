# NerdBot
一个简易的基于OpenAI chatGPT和Onebot 11标准的聊天机器人
## 使用步骤
1. 安装redis
    1. Update your system:  
`sudo yum update -y`
    2. Install the Redis package:  
`sudo yum install -y redis`
    3. Start the Redis service:  
`sudo systemctl start redis`
    4. Enable Redis to start at boot time:  
`sudo systemctl enable redis`
    5. Verify the Redis installation:  
`redis-cli ping`  
If Redis is running properly, you should receive a "PONG" response.
2. 安装、配置并启动符合OneBot11标准的登录转发服务，如cq_http  
参照 <https://github.com/Mrs4s/go-cqhttp>  
自行按照 Onebot11 标准开发请参照 <https://github.com/botuniverse/onebot-11>
3. 安装Golang, version >= 1.18
4. `go build NerdBot`
5. 首次运行生成配置文件 config.yaml。对其进行配置后再次启动服务即可。
## 用户命令
+ 任何用户都可执行的聊天窗口命令
    - `NerdBot clear`      //清除与对话者的所有prompts，重新开始话题
## 管理员命令  
+ 管理员可以在聊天窗口中输入各类命令，目前包括：
    - `NerdBot group mode` //开启群聊模式，即记录所有群聊信息到prompts内，会消耗大量tokens
    - `NerdBot private mode` //默认模式，单对单的有上下文的对话
    - `NerdBot set temperature [0 ~ 1]`   //设置temperature
## 作者的话  
欢迎积极参与开发与提issues。大佬轻喷。

# NerdBot
A simple chatbot based on the OpenAI chatGPT and Onebot 11 standards
## Use steps
1. Install redis
   1. Update your system:  
      `sudo yum update -y`
   2. Install the Redis package:  
      `sudo yum install -y redis`
   3. Start the Redis service:  
      `sudo systemctl start redis`
   4. Enable Redis to start at boot time:  
      `sudo systemctl enable redis`
   5. Verify the Redis installation:  
      `redis-cli ping`  
      If Redis is running properly, you should receive a "PONG" response.
2. Install, configure, and start the login and forwarding service that complies with the OneBot11 standard, for example, cq_http  
   With reference to the < https://github.com/Mrs4s/go-cqhttp >  
   If you want to develop on your own, please see the Onebot11 standards < https://github.com/botuniverse/onebot-11 >
3. Install Golang, version >= 1.18
4. `go build NerdBot`
5. Run and generate default configuration file "config.yaml" for the first time. Configure it and start the service again.
## User command
+ Chat window commands that any user can execute
- `NerdBot clear` // Clears all prompts with the user to restart the topic
## Administrator command
+ The administrator can enter various commands in the chat window, including:
- `NerdBot group mode` // Enabling group chat mode by logging all group chat information into prompts consumes a lot of tokens
- `NerdBot private mode` // Default mode, one-to-one conversation with context
- `NerdBot set temperature [0 ~ 1]` // Set temperature
## The author's words
Welcome to actively participate in the development and issues. 