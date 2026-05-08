package workflow

import (
	"context"
	"eino-quickstart/models"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func ChainDemo() {
	ctx := context.Background()
	chatModel, err := models.NewChatModel(ctx)
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}
	prompt := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage("你是一个专业的Go语言技术顾问，回答简洁准确。"),
		schema.UserMessage("Hello, {name}! how are you?"),
	)
	chain := compose.NewChain[map[string]any, *schema.Message]()
	chain.AppendChatTemplate(prompt).AppendChatModel(chatModel)
	runner, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("编译链失败: %v", err)
	}
	resp, err := runner.Invoke(ctx, map[string]any{"name": "张三"})
	if err != nil {
		log.Fatalf("运行链失败: %v", err)
	}
	fmt.Printf("模型回复：%s\n", resp.Content)
}

type TranslateRequest struct {
	OriginText string `json:"origin_text" jsonschema:"required" jsonschema_description:"需要翻译的原始内容"`
	TargetLang string `json:"target_lang" jsonschema:"required,enum=英文,enum=中文,enum=日文" jsonschema_description:"目标语言"`
}

type TranslateResponse struct {
	OriginText string `json:"origin_text"`
	TargetLang string `json:"target_lang"`
	Result     string `json:"result"`
}

func translateText(ctx context.Context, req *TranslateRequest) (*TranslateResponse, error) {
	text := strings.TrimSpace(req.OriginText)
	result := text

	switch req.TargetLang {
	case "英文":
		switch text {
		case "你好，世界", "你好世界":
			result = "Hello, world."
		case "今天天气很好":
			result = "The weather is nice today."
		default:
			result = "[模拟英文翻译] " + text
		}
	case "中文":
		switch strings.ToLower(text) {
		case "hello, world.", "hello world":
			result = "你好，世界"
		default:
			result = "[模拟中文翻译] " + text
		}
	case "日文":
		switch text {
		case "你好，世界", "你好世界":
			result = "こんにちは、世界。"
		default:
			result = "[模拟日文翻译] " + text
		}
	default:
		return nil, fmt.Errorf("不支持的目标语言: %s", req.TargetLang)
	}

	return &TranslateResponse{
		OriginText: req.OriginText,
		TargetLang: req.TargetLang,
		Result:     result,
	}, nil
}

func ChainWithToolDemo() {
	ctx := context.Background()

	chatModel, err := models.NewChatModel(ctx)
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}

	translateTool, err := utils.InferTool(
		"translate_text",
		"翻译文本到指定目标语言。只有用户明确要求翻译时才使用这个工具。",
		translateText,
	)
	if err != nil {
		log.Fatalf("创建翻译工具失败: %v", err)
	}

	toolInfo, err := translateTool.Info(ctx)
	if err != nil {
		log.Fatalf("获取工具信息失败: %v", err)
	}

	chatModelWithTools, err := chatModel.WithTools([]*schema.ToolInfo{toolInfo})
	if err != nil {
		log.Fatalf("绑定工具失败: %v", err)
	}

	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{translateTool},
		ToolArgumentsHandler: func(ctx context.Context, name, arguments string) (string, error) {
			fmt.Printf("[ToolArgumentsHandler] 工具名: %s\n", name)
			fmt.Printf("[ToolArgumentsHandler] 参数: %s\n", arguments)
			return arguments, nil
		},
		ToolCallMiddlewares: []compose.ToolMiddleware{
			{
				Invokable: func(next compose.InvokableToolEndpoint) compose.InvokableToolEndpoint {
					return func(ctx context.Context, input *compose.ToolInput) (*compose.ToolOutput, error) {
						fmt.Printf("[ToolCall] 执行工具: %s, args=%s\n", input.Name, input.Arguments)
						out, err := next(ctx, input)
						if err != nil {
							return nil, err
						}
						fmt.Printf("[ToolCall] 工具结果: %s\n", out.Result)
						return out, nil
					}
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("创建 ToolsNode 失败: %v", err)
	}

	template := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage("你是一个翻译助手。用户要求翻译时，必须调用 translate_text 工具，不要自己直接翻译。"),
		schema.UserMessage("{input}"),
	)

	chain := compose.NewChain[map[string]any, []*schema.Message]()
	chain.
		AppendChatTemplate(template).
		AppendChatModel(chatModelWithTools).
		AppendToolsNode(toolsNode)

	runner, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("编译工具链失败: %v", err)
	}
	results, err := runner.Invoke(ctx, map[string]any{
		"input": "请把“你好，世界”翻译成英文",
	})
	if err != nil {
		log.Fatalf("运行工具链失败: %v", err)
	}

	for _, msg := range results {
		fmt.Printf("工具消息 Role: %s\n", msg.Role)
		fmt.Printf("工具调用 ID: %s\n", msg.ToolCallID)
		fmt.Printf("工具返回内容: %s\n", msg.Content)
	}
}
func BranchDemo() {
	ctx := context.Background()

	model, err := models.NewChatModel(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// 分类器：用 Lambda 做简单的关键词分类
	classifier := compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		question, ok := input["question"].(string)
		if !ok {
			return nil, fmt.Errorf("question 必须是 string")
		}

		if strings.Contains(question, "代码") || strings.Contains(question, "编程") || strings.Contains(question, "bug") {
			input["category"] = "code"
		} else if strings.Contains(question, "部署") || strings.Contains(question, "运维") || strings.Contains(question, "服务器") {
			input["category"] = "ops"
		} else {
			input["category"] = "general"
		}
		return input, nil
	})

	// 三个不同角色的 Prompt 模板
	codeTpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个资深Go语言开发者，擅长代码审查和问题排查，请回答用户的编程问题。"),
		schema.UserMessage("{question}"),
	)

	opsTpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个运维专家，精通Linux、Docker和K8s，请回答用户的运维问题。"),
		schema.UserMessage("{question}"),
	)

	generalTpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个友好的技术助手，请简洁地回答用户的问题。"),
		schema.UserMessage("{question}"),
	)

	// 构建 Graph
	graph := compose.NewGraph[map[string]any, *schema.Message]()

	// 添加节点
	_ = graph.AddLambdaNode("classifier", classifier)
	_ = graph.AddChatTemplateNode("code_tpl", codeTpl)
	_ = graph.AddChatTemplateNode("ops_tpl", opsTpl)
	_ = graph.AddChatTemplateNode("general_tpl", generalTpl)
	_ = graph.AddChatModelNode("model", model)

	// 定义条件路由
	_ = graph.AddEdge(compose.START, "classifier")
	_ = graph.AddBranch("classifier", compose.NewGraphBranch(
		// 条件函数：根据分类结果决定走哪个分支
		func(ctx context.Context, input map[string]any) (string, error) {
			category := input["category"].(string)
			return category + "_tpl", nil
		},
		// 分支映射：声明所有可能的下游节点
		map[string]bool{
			"code_tpl":    true,
			"ops_tpl":     true,
			"general_tpl": true,
		},
	))

	// 三条分支最终都汇聚到同一个模型节点
	_ = graph.AddEdge("code_tpl", "model")
	_ = graph.AddEdge("ops_tpl", "model")
	_ = graph.AddEdge("general_tpl", "model")
	_ = graph.AddEdge("model", compose.END)

	// 编译并运行
	runner, err := graph.Compile(ctx)
	if err != nil {
		log.Fatal("编译失败:", err)
	}

	// 测试不同类型的问题
	questions := []string{
		"Go代码里怎么避免goroutine泄漏？",
		"Docker容器部署时端口映射不生效怎么办？",
		"推荐几本学习分布式系统的书？",
	}

	for _, q := range questions {
		result, err := runner.Invoke(ctx, map[string]any{"question": q})
		if err != nil {
			log.Printf("问题: %s, 错误: %v\n", q, err)
			continue
		}
		fmt.Printf("问题: %s\n回答: %s\n\n", q, result.Content)
	}
}
func ParallDemo() {
	ctx := context.Background()

	model, err := models.NewChatModel(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// 两个不同视角的模板
	techTpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个技术架构师，请从技术可行性角度分析这个需求，用一两句话概括。"),
		schema.UserMessage("{requirement}"),
	)
	productTpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个产品经理，请从用户价值角度分析这个需求，用一两句话概括。"),
		schema.UserMessage("{requirement}"),
	)

	// 构建两条子链
	techChain := compose.NewChain[map[string]any, *schema.Message]()
	techChain.AppendChatTemplate(techTpl).AppendChatModel(model)

	productChain := compose.NewChain[map[string]any, *schema.Message]()
	productChain.AppendChatTemplate(productTpl).AppendChatModel(model)

	// 构建并行节点
	parallel := compose.NewParallel()
	parallel.AddGraph("tech", techChain)
	parallel.AddGraph("product", productChain)

	// 主链：并行执行 → 合并结果
	chain := compose.NewChain[map[string]any, string]()
	chain.
		AppendParallel(parallel).
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, results map[string]any) (string, error) {
			techResult := results["tech"].(*schema.Message)
			productResult := results["product"].(*schema.Message)
			return fmt.Sprintf("【技术视角】%s\n\n【产品视角】%s",
				techResult.Content, productResult.Content), nil
		}))

	runner, err := chain.Compile(ctx)
	if err != nil {
		log.Fatal("编译失败:", err)
	}

	result, err := runner.Invoke(ctx, map[string]any{
		"requirement": "为电商App添加AI智能客服功能",
	})
	if err != nil {
		log.Fatal("运行失败:", err)
	}

	fmt.Println(result)
}
func StreamChainDemo() {
	ctx := context.Background()

	model, err := models.NewChatModel(ctx)
	if err != nil {
		log.Fatal(err)
	}

	tpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个Go语言教学助手，请详细解释用户的问题。"),
		schema.UserMessage("{question}"),
	)

	chain := compose.NewChain[map[string]any, *schema.Message]()
	chain.AppendChatTemplate(tpl).AppendChatModel(model)

	runner, err := chain.Compile(ctx)
	if err != nil {
		log.Fatal(err)
	}

	input := map[string]any{"question": "Go的interface是怎么实现的？"}

	// 方式一：Invoke 同步调用
	fmt.Println("=== Invoke 同步调用 ===")
	result, err := runner.Invoke(ctx, input)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Content)

	// 方式二：Stream 流式调用
	fmt.Println("\n=== Stream 流式调用 ===")
	stream, err := runner.Stream(ctx, input)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		fmt.Print(chunk.Content) // 逐块输出，不换行
	}
	fmt.Println()
}
func GraphDemo() {
	ctx := context.Background()
	model, err := models.NewChatModel(ctx)
	if err != nil {
		log.Fatal(err)
	}
	classifier := compose.InvokableLambda[map[string]any, map[string]any](
		func(ctx context.Context, input map[string]any) (map[string]any, error) {
			question, ok := input["question"].(string)
			if !ok {
				return nil, fmt.Errorf("question 必须是 string")
			}
			if strings.Contains(question, "代码") || strings.Contains(question, "编程") || strings.Contains(question, "bug") {
				input["category"] = "code"
			} else if strings.Contains(question, "操作") || strings.Contains(question, "部署") || strings.Contains(question, "配置") {
				input["category"] = "ops"
			} else {
				input["category"] = "general"
			}
			return input, nil
		},
	)
	codeTpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是前后端全栈程序员，你需要根据用户的问题，给出相应的解决方案。"),
		schema.UserMessage("{question}"),
	)
	opsTpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个运维专家，你需要根据用户的问题，给出相应的解决方案。"),
		schema.UserMessage("{question}"),
	)
	generalTpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个通用助手，你需要根据用户的问题，给出相应的解决方案。"),
		schema.UserMessage("{question}"),
	)
	graph := compose.NewGraph[map[string]any, string]()
	graph.AddLambdaNode("classifier", classifier)
	graph.AddChatTemplateNode("code", codeTpl)
	graph.AddChatTemplateNode("ops", opsTpl)
	graph.AddChatTemplateNode("general", generalTpl)
	graph.AddChatModelNode("model", model)
	_ = graph.AddLambdaNode("formatter", compose.InvokableLambda(
		func(ctx context.Context, msg *schema.Message) (string, error) {
			return fmt.Sprintf("[AI助手] %s", msg.Content), nil
		},
	))
	// 连接编排
	_ = graph.AddEdge(compose.START, "classifier")

	// 条件路由
	_ = graph.AddBranch("classifier", compose.NewGraphBranch(
		func(ctx context.Context, input map[string]any) (string, error) {
			return input["category"].(string), nil
		},
		map[string]bool{
			"code":    true,
			"ops":     true,
			"general": true,
		},
	))

	// 汇聚到模型
	_ = graph.AddEdge("code", "model")
	_ = graph.AddEdge("ops", "model")
	_ = graph.AddEdge("general", "model")
	_ = graph.AddEdge("model", "formatter")
	_ = graph.AddEdge("formatter", compose.END)

	// 编译
	runner, err := graph.Compile(ctx)
	if err != nil {
		log.Fatal("编译失败:", err)
	}

	// 测试多个问题
	questions := []string{
		"我该如何部署这个前端项目？",
		"go的代码如何写一个和函数",
		"新手程序员应该先学前端还是后端？",
	}

	for _, q := range questions {
		result, err := runner.Invoke(ctx, map[string]any{"question": q})
		if err != nil {
			log.Printf("错误: %v\n", err)
			continue
		}
		fmt.Printf("问题: %s\n%s\n\n", q, result)
	}
}
