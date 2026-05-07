package tool

import (
	"context"
	"eino-quickstart/models"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type WeatherRequest struct {
	City string `json:"city" jsonschema:"required" jsonschema_description:"城市名称"`
}

type WeatherResponse struct {
	City    string `json:"city"`
	Temp    string `json:"temp"`
	Weather string `json:"weather"`
}

func getWeather(ctx context.Context, req *WeatherRequest) (*WeatherResponse, error) {
	// 这里用硬编码模拟，实际项目中你会去调天气 API
	mockData := map[string]WeatherResponse{
		"北京": {City: "北京", Temp: "22°C", Weather: "晴"},
		"上海": {City: "上海", Temp: "26°C", Weather: "多云"},
		"深圳": {City: "深圳", Temp: "30°C", Weather: "阵雨"},
	}

	if data, ok := mockData[req.City]; ok {
		return &data, nil
	}
	return &WeatherResponse{City: req.City, Temp: "未知", Weather: "未知"}, nil
}

func ToolDemo() {
	// ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	// defer cancel()
	// // 创建基础模型实例
	// cm, err := models.NewChatModel(ctx)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	ctx := context.Background()
	cm, err := models.NewChatModel(ctx)
	if err != nil {
		log.Fatal(err)
	}
	// 定义工具描述信息
	weatherTool := &schema.ToolInfo{
		Name: "get_weather",
		Desc: "查询指定城市的当前天气",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"city": {
				Type:     "string",
				Desc:     "城市名称，如：北京、上海",
				Required: true,
			},
		}),
	}
	// 用 WithTools 绑定工具，返回新实例
	cmWithTools, err := cm.WithTools([]*schema.ToolInfo{weatherTool})
	if err != nil {
		log.Fatal(err)
	}
	// 用绑定了工具的实例来调用
	messages := []*schema.Message{
		schema.SystemMessage("你是一个天气助手。你必须通过调用工具来查询天气，如果没有可用的工具，请直接告诉用户你无法查询实时天气信息，不要编造任何天气数据。"),
		schema.UserMessage("北京今天天气怎么样？"),
	}
	resp, err := cmWithTools.Generate(ctx, messages)
	if err != nil {
		log.Fatal(err)
	}
	// 检查模型是否发起了工具调用
	if len(resp.ToolCalls) > 0 {
		fmt.Println("模型请求调用工具：")
		for _, tc := range resp.ToolCalls {
			fmt.Printf("  工具名: %s\n", tc.Function.Name)
			fmt.Printf("  参数: %s\n", tc.Function.Arguments)
		}
	} else {
		fmt.Println("模型直接回复：", resp.Content)
	}
	// 原始的 cm 没有绑定工具，不受影响
	// 使用不同的 messages，明确告知模型它没有任何工具可用
	messagesNoTool := []*schema.Message{
		schema.SystemMessage("你是一个天气助手，但你没有任何工具可以使用。当用户询问天气时，你如果不知道，就要回答不知道，绝对不能编造天气信息。"),
		schema.UserMessage("北京今天天气怎么样？"),
	}
	resp2, _ := cm.Generate(ctx, messagesNoTool)
	fmt.Println("\n未绑定工具的实例回复：", resp2.Content)
}
func SingleToolParamDemo() {
	ctx := context.Background()

	// 用 NewTool 创建 InvokableTool
	weatherTool := utils.NewTool(
		&schema.ToolInfo{
			Name: "get_weather",
			Desc: "查询指定城市的实时天气信息，包括温度和天气状况",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"city": {
					Type:     schema.String,
					Desc:     "要查询天气的城市名称，如：北京、上海、深圳",
					Required: true,
				},
			}),
		},
		getWeather,
	)

	// 验证工具信息
	info, _ := weatherTool.Info(ctx)
	fmt.Printf("工具名: %s\n", info.Name)
	fmt.Printf("工具描述: %s\n", info.Desc)

	// 模拟模型生成的工具调用参数（JSON 字符串）
	args := `{"city": "北京"}`

	// 执行工具
	result, err := weatherTool.InvokableRun(ctx, args) //内部调用getWeather函数，并且根据返回的结构体定义序列化为JSON字符串返回
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("执行结果: %s\n", result)

	// 解析结果
	var resp WeatherResponse
	json.Unmarshal([]byte(result), &resp)
	fmt.Printf("城市: %s, 温度: %s, 天气: %s\n", resp.City, resp.Temp, resp.Weather)

}

// jsonschema tag 里多个规则应该用英文逗号分隔，不是空格。
// enum 写法是 enum=xxx，不是 enum:xxx。
// 不需要写 type:enum，Op 本身是 string，JSON Schema 的 type: "string" 会自动推断出来；enum 是约束，不是 type。
type CalcRequest struct {
	A  float64 `json:"a" jsonschema:"required" jsonschema_description:"第一个数字"`
	B  float64 `json:"b" jsonschema:"required" jsonschema_description:"第二个数字"`
	Op string  `json:"op" jsonschema:"required,enum=add,enum=subtract,enum=multiply,enum=divide" jsonschema_description:"四则运算符，如：add、subtract、multiply、divide"`
}

type CalcResponse struct {
	Expression string  `json:"expression"`
	Result     float64 `json:"result"`
}

func calculate(ctx context.Context, req *CalcRequest) (*CalcResponse, error) {
	var result float64
	switch req.Op {
	case "add":
		result = req.A + req.B
	case "subtract":
		result = req.A - req.B
	case "multiply":
		result = req.A * req.B
	case "divide":
		result = req.A / req.B
	default:
		return nil, errors.New("不支持的操作符")
	}
	return &CalcResponse{
		Expression: fmt.Sprintf("%f %s %f", req.A, req.Op, req.B),
		Result:     result,
	}, nil
}
func MultiToolParamDemo() {
	ctx := context.Background()
	//定义计算工具
	calculateTool := utils.NewTool(
		&schema.ToolInfo{
			Name: "calculate",
			Desc: "计算两个数字的四则运算结果",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"a": {
					Type:     schema.Number,
					Desc:     "第一个数字",
					Required: true,
				},
				"b": {
					Type:     schema.Number,
					Desc:     "第二个数字",
					Required: true,
				},
				"op": {
					Type:     schema.String,
					Desc:     "四则运算符，如：add、subtract、multiply、divide",
					Required: true,
					Enum:     []string{"add", "subtract", "multiply", "divide"},
				},
			}),
		},
		calculate,
	)
	// 模拟调用
	result, err := calculateTool.InvokableRun(ctx, `{"a": 12.5, "b": 3.7, "op": "multiply"}`)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}

// 通过 struct tag 定义参数的描述、约束
type SearchRequest struct {
	Query    string `json:"query" jsonschema:"required" jsonschema_description:"搜索关键词"`
	MaxCount int    `json:"max_count" jsonschema_description:"最多返回的结果数量，默认5"`
	Language string `json:"language" jsonschema:"enum=zh,enum=en" jsonschema_description:"结果语言，zh为中文，en为英文"`
}

type SearchResult struct {
	Items []SearchItem `json:"items"`
	Total int          `json:"total"`
}

type SearchItem struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Summary string `json:"summary"`
}

func searchWeb(ctx context.Context, req *SearchRequest) (*SearchResult, error) {
	// 模拟搜索逻辑
	maxCount := req.MaxCount
	if maxCount <= 0 {
		maxCount = 5
	}

	items := []SearchItem{
		{Title: "Go语言官方文档", URL: "https://go.dev/doc/", Summary: "Go编程语言官方文档和教程"},
		{Title: "Eino框架指南", URL: "https://cloudwego.io/docs/eino", Summary: "字节跳动开源的Go语言LLM应用开发框架"},
		{Title: "Go并发编程实战", URL: "https://example.com/go-concurrency", Summary: "深入讲解goroutine和channel的用法"},
	}

	// 简单过滤
	filtered := make([]SearchItem, 0)
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Title+item.Summary), strings.ToLower(req.Query)) {
			filtered = append(filtered, item)
		}
		if len(filtered) >= maxCount {
			break
		}
	}

	return &SearchResult{Items: filtered, Total: len(filtered)}, nil
}
func InferToolParams() {
	// InferTool 从函数签名和 struct tag 自动推断工具信息
	searchTool, err := utils.InferTool("web_search", "搜索互联网上的信息，返回相关网页的标题、链接和摘要", searchWeb)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// 查看自动推断出的工具信息
	info, _ := searchTool.Info(ctx)
	infoJSON, _ := json.MarshalIndent(info, "", "  ")
	fmt.Println("自动推断的工具信息：")
	fmt.Println(string(infoJSON))

	fmt.Println()

	// 执行工具
	_ = time.Now() // 占位，实际项目中可能用于日志
	result, err := searchTool.InvokableRun(ctx, `{"query": "Go", "max_count": 2, "language": "zh"}`)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("搜索结果：")
	fmt.Println(result)
}

type CityTimeRequest struct {
	City string `json:"city" jsonschema:"required enum=北京,东京,伦敦,纽约" jsonschema_description:"城市名称"`
}

type CityTimeResponse struct {
	City string `json:"city"`
	Time string `json:"time"`
	Zone string `json:"zone"`
}

func getCityTime(ctx context.Context, req *CityTimeRequest) (*CityTimeResponse, error) {
	zones := map[string]string{
		"北京": "Asia/Shanghai (UTC+8)",
		"东京": "Asia/Tokyo (UTC+9)",
		"伦敦": "Europe/London (UTC+0)",
		"纽约": "America/New_York (UTC-5)",
	}
	zone := zones[req.City]
	if zone == "" {
		zone = "未知时区"
	}
	return &CityTimeResponse{
		City: req.City,
		Time: "2025-06-01 14:30:00",
		Zone: zone,
	}, nil
}
func ToolNodeDemo() { //ToolNode节点用于执行工具，ToolInfo只负责定义工具
	ctx := context.Background()

	// 创建工具
	timeTool, _ := utils.InferTool("get_city_time", "查询指定城市的当前时间和时区信息", getCityTime)
	/*
		type ToolsNodeConfig struct {
			// 工具列表
			Tools []tool.BaseTool

			// 模型幻觉处理：当模型调用了一个不存在的工具时怎么办
			UnknownToolsHandler func(ctx context.Context, name, input string) (string, error)

			// 是否按顺序执行工具（默认并行执行）
			ExecuteSequentially bool

			// 工具参数预处理
			ToolArgumentsHandler func(ctx context.Context, name, arguments string) (string, error)
		}*/
	// 	ExecuteSequentially控制多个工具的执行顺序。默认情况下，如果模型一次请求调用了多个工具(比如同时查北京和上海的天气)，
	// ToolsNode会并行执行它们，提高效率。如果你的工具之间有依赖关系(比如第二个工具的参数依赖第一个工具的结果)，就需要设为true 让它们按顺序执行。
	// UnknownToolsHandler用来兜底模型的"幻觉"。大模型偶尔会编造一个不存在的工具名来调用--如果你不设这个handler，
	// ToolsNode会直接报错;设了之后可以优雅地返回一个提示信息，告诉模型"这个工具不存在"，让它换个思路。
	// ToolArgumentsHandler可以在工具执行前对参数做预处理。比如模型有时候会把数字写成字符串、日期格式不统一，你可以在这里做一次清洗和标准化。
	// 创建 ToolsNode，不会自己思考要不要调用工具，也不会自己决定调用哪个工具；决策在大模型那里，执行在 ToolsNode 这里。
	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{timeTool},
	})
	if err != nil {
		log.Fatal(err)
	}

	// 模拟模型返回了一条带有工具调用请求的消息
	modelOutput := &schema.Message{
		Role: schema.Assistant,
		ToolCalls: []schema.ToolCall{
			{
				ID: "call_001",
				Function: schema.FunctionCall{
					Name:      "get_city_time",
					Arguments: `{"city": "东京"}`,
				},
			},
			{
				ID: "call_002",
				Function: schema.FunctionCall{
					Name:      "get_city_time",
					Arguments: `{"city": "伦敦"}`,
				},
			},
		},
	}

	// ToolsNode 执行工具调用
	results, err := toolsNode.Invoke(ctx, modelOutput)
	if err != nil {
		log.Fatal(err)
	}

	for _, msg := range results {
		fmt.Printf("角色: %s\n", msg.Role)
		fmt.Printf("工具调用ID: %s\n", msg.ToolCallID)
		fmt.Printf("结果: %s\n", msg.Content)
	}
}
func RealToolNodeDemo() {
	ctx := context.Background()
	chatModel, err := models.NewChatModel(ctx)
	if err != nil {
		log.Fatal(err)
	}
	//定义工具
	/*type ToolInfo struct {
		// The unique name of the tool that clearly communicates its purpose.
		Name string
		// Used to tell the model how/when/why to use the tool.
		// You can provide few-shot examples as a part of the description.
		Desc string
		// Extra is the extra information for the tool.
		Extra map[string]any

		// The parameters the functions accepts (different models may require different parameter types).
		// can be described in two ways:
		//  - use params: schema.NewParamsOneOfByParams(params)
		//  - use jsonschema: schema.NewParamsOneOfByJSONSchema(jsonschema)
		// If is nil, signals that the tool does not need any input parameter
		*ParamsOneOf
	}*/
	timeTool, _ := utils.InferTool("get_city_time", "查询指定城市的当前时间和时区信息", getCityTime)
	calculateTool, _ := utils.InferTool("calculate", "计算两个数字的四则运算,需要计算时，严格调用此工具", calculate)

	timeInfo, _ := timeTool.Info(ctx)           //转为ToolInfo类型
	calculateInfo, _ := calculateTool.Info(ctx) //转为ToolInfo类型
	toolInfos := []*schema.ToolInfo{timeInfo, calculateInfo}

	chatModelTools, err := chatModel.WithTools(toolInfos) //返回带工具的ChatModel模型实例
	if err != nil {
		log.Fatal(err)
	}
	//创建执行的ToolNode
	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{timeTool, calculateTool},
	})
	if err != nil {
		log.Fatal(err)
	}
	//构建消息
	history := []*schema.Message{schema.SystemMessage("你是一个智能助手，可以调用工具查询指定地点的时间和计算结果。"),
		schema.UserMessage("我昨天在上海买了一个商品，商品100元，我手里80元，请问我需要借多少钱才能买两个商品，同时我昨天位于哪个时区？")}
	//调用模型
	resp, err := chatModelTools.Generate(ctx, history) //限制输出的token数量，避免输出过长
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("模型回复：%s\n", resp.String())
	//调用工具
	if len(resp.ToolCalls) > 0 {
		fmt.Println("模型请求调用工具：")
		for _, toolCall := range resp.ToolCalls {
			fmt.Printf("工具调用ID: %s\n", toolCall.ID)
			fmt.Printf("工具调用函数: %s\n", toolCall.Function.Name)
			fmt.Printf("工具调用参数: %s\n", toolCall.Function.Arguments)
		}
		toolsResults, err := toolsNode.Invoke(ctx, resp)
		if err != nil {
			log.Fatal(err)
		}
		for _, msg := range toolsResults {
			fmt.Printf("角色: %s\n", msg.Role)
			fmt.Printf("工具调用ID: %s\n", msg.ToolCallID)
			fmt.Printf("结果: %s\n", msg.Content)
		}
		history = append(history, resp)            // 模型的工具调用消息
		history = append(history, toolsResults...) // 工具执行结果
		// 9. 第二轮：模型根据工具结果生成最终回答
		finalResp, err := chatModel.Generate(ctx, history)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("助手: %s\n", finalResp.Content)
	}
	fmt.Println(resp.Content)
}
