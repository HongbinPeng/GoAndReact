package react_agent

import (
	"context"
	"eino-quickstart/models"
	"fmt"
	"log"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

type WeatherRequest struct {
	City string `json:"city"`
}

type WeatherResponse struct {
	City    string `json:"city"`
	Temp    string `json:"temp"`
	Weather string `json:"weather"`
}

func getWeather(ctx context.Context, req *WeatherRequest) (*WeatherResponse, error) {
	// fmt.Printf("[getWeather] 收到参数: %+v\n", req)
	mockData := map[string]WeatherResponse{
		"北京": {City: "北京", Temp: "22°C", Weather: "晴"},
		"上海": {City: "上海", Temp: "26°C", Weather: "多云"},
	}
	if data, ok := mockData[req.City]; ok {
		return &data, nil
	}
	return &WeatherResponse{City: req.City, Temp: "未知", Weather: "未知"}, nil
}

func ReactAgentDemo() {
	ctx := context.Background()

	chatModel, err := models.NewChatModel(ctx)
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}

	weatherTool := utils.NewTool(
		&schema.ToolInfo{
			Name: "get_weather",
			Desc: "查询指定城市的实时天气信息",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"city": {Type: schema.String, Desc: "城市名称", Required: true},
			}),
		},
		getWeather,
	)

	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{weatherTool},
			ToolArgumentsHandler: func(ctx context.Context, name, arguments string) (string, error) {
				fmt.Printf("[ToolArgumentsHandler] 工具名: %s\n", name)
				fmt.Printf("[ToolArgumentsHandler] 原始参数: %s\n", arguments)

				// 不修改参数，原样返回
				return arguments, nil
			},
		},
		MessageModifier: func(ctx context.Context, input []*schema.Message) []*schema.Message {
			messages := make([]*schema.Message, 0, len(input)+1)
			messages = append(messages, schema.SystemMessage("你是一个天气助手，回答简洁。"))
			messages = append(messages, input...)
			return messages
		},
	})
	if err != nil {
		log.Fatalf("创建 Agent 失败: %v", err)
	}

	// 维护对话历史
	history := make([]*schema.Message, 0)

	// 第一轮对话
	history = append(history, schema.UserMessage("北京天气怎么样？"))
	answer1, err := agent.Generate(ctx, history)
	if err != nil {
		log.Fatalf("执行失败: %v", err)
	}
	fmt.Println("第一轮回答:", answer1.Content)
	history = append(history, answer1) // 把 Agent 的回答也加入历史

	// 第二轮对话——追问，不需要重复说城市
	history = append(history, schema.UserMessage("那上海呢？"))
	answer2, err := agent.Generate(ctx, history)
	if err != nil {
		log.Fatalf("执行失败: %v", err)
	}
	fmt.Println("第二轮回答:", answer2.Content)
	history = append(history, answer2)

	// 第三轮对话——基于前两轮的对比问题
	history = append(history, schema.UserMessage("哪个城市更热？"))
	answer3, err := agent.Generate(ctx, history)
	if err != nil {
		log.Fatalf("执行失败: %v", err)
	}
	fmt.Println("第三轮回答:", answer3.Content)
}
