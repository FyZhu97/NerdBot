# NerdBot
一个简易的基于OpenAI chatGPT和cq_http的QQ聊天机器人（后续可能会增加微信聊天功能）

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
2. 安装、配置并启动cq_http服务  
参照<https://github.com/Mrs4s/go-cqhttp>
3. 安装Golang, version >= 1.18
4. `go build NerdBot`
5. 首次运行生成配置文件 config.yml。对其进行配置后重启服务即可。
## 管理员命令  
+ 管理员可以在QQ对话框中输入各类命令，目前包括：
    - `NerdBot group mode` //开启群聊模式，即记录所有群聊信息到prompts内，会消耗大量tokens
    - `NerdBot private mode` //默认模式，单对单的有上下文的对话
    - `NerdBot clear`      //清除与对话者的所有prompts，重新开始话题
    - `NerdBot set temperature [0 ~ 1]`   //设置temperature
## 作者的话  
欢迎积极参与开发与提issues。大佬轻喷。
