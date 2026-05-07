package promptsFormat

import (
	"context"
	"eino-quickstart/models"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func Ptr[T any](v T) *T { return &v }

// 保留最近 maxRounds 轮对话（每轮 = 1条用户消息 + 1条助手消息）
func trimHistory(history []*schema.Message, maxRounds int) []*schema.Message {
	maxMessages := maxRounds * 2
	if len(history) <= maxMessages {
		return history
	}
	return history[len(history)-maxMessages:]
}
func PromptPlaceHolder() {
	ctx := context.Background()
	cm, err := models.NewChatModel(ctx)
	if err != nil {
		log.Fatal(err)
	}
	// 创建模板
	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个{role}。"),
		schema.MessagesPlaceholder("history_key", false),
		&schema.Message{
			Role:    schema.User,
			Content: "请帮我{task}。",
		},
	)
	// 准备变量
	variables := map[string]any{
		"role":        "专业的助手",
		"task":        "写一首诗",
		"history_key": []*schema.Message{{Role: schema.User, Content: "告诉我油画是什么?"}, {Role: schema.Assistant, Content: "油画是xxx"}},
	}
	// 格式化模板
	messages, err := template.Format(ctx, variables)
	if err != nil {
		log.Fatal(err)
	}
	// 打印格式化后的消息
	cm.Generate(ctx, messages, openai.WithExtraFields(map[string]any{"temperature": 0.7}))
}

func PromptDemo() {
	ctx := context.Background()
	cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
		APIKey:  os.Getenv("DASHSCOPE_API_KEY"),
		Model:   "qwen-plus",
	})
	if err != nil {
		log.Fatal(err)
	}
	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个Go语言教学助手。回答简洁明了。"),
		schema.MessagesPlaceholder("history", false),
		schema.UserMessage("{input}"),
	)
	// 模拟多轮对话
	conversations := []string{
		"什么是 goroutine？",
		"它和线程有什么区别？",
		"那 channel 呢？",
		"channel 有缓冲和无缓冲的区别是什么？",
		"前面说的 goroutine 调度器是怎么工作的？",
	}
	history := make([]*schema.Message, 0)
	for _, input := range conversations {
		// 只保留最近 3 轮对话
		trimmed := trimHistory(history, 3)
		messages, _ := template.Format(ctx, map[string]any{
			"history": trimmed,
			"input":   input,
		})
		resp, err := cm.Generate(ctx, messages)
		if err != nil {
			log.Printf("调用失败: %v", err)
			continue
		}
		fmt.Printf("用户: %s\n", input)
		fmt.Printf("助手: %s\n\n", resp.Content)
		history = append(history, schema.UserMessage(input))
		history = append(history, schema.AssistantMessage(resp.Content, nil))
	}

	fmt.Printf("实际历史消息数: %d，发送给模型的最多: %d\n", len(history), 3*2)
}
func PromptMultiType() {
	ctx := context.Background()
	cm, err := models.NewChatModel(ctx)
	if err != nil {
		log.Fatal(err)
	}
	var history []*schema.Message = []*schema.Message{
		schema.SystemMessage("你是一个AI助手。回答简洁明了。"),
		schema.UserMessage("你能介绍一下Go语言吗？"),
		schema.AssistantMessage("Go语言是一种静态类型、编译型语言，由Google开发。它的设计目标是简单、高效和可维护。", nil),
		&schema.Message{
			Role: schema.User,
			UserInputMultiContent: []schema.MessageInputPart{
				{Type: schema.ChatMessagePartTypeText,
					Text: "下面这些图片是什么？"},
				{
					Type: schema.ChatMessagePartTypeImageURL,
					Image: &schema.MessageInputImage{
						MessagePartCommon: schema.MessagePartCommon{
							URL:      Ptr("https://c-ssl.duitang.com/uploads/blog/202405/28/xDS6boyS2wmn20.jpeg"),
							MIMEType: "image/jpeg",
						},
					},
				},
				{
					Type: schema.ChatMessagePartTypeImageURL,
					Image: &schema.MessageInputImage{
						MessagePartCommon: schema.MessagePartCommon{
							URL:      Ptr("https://www.keaitupian.cn/cjpic/frombd/2/253/3207721921/3451553052.jpg"),
							MIMEType: "image/jpg",
						},
					},
				},
			},
		},
	}
	resp, err := cm.Generate(ctx, history)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("助手: %s\n", resp.Content)
}
func PromptInChain() {
	ctx := context.Background()
	// 创建模板
	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个{role}。"),
		schema.MessagesPlaceholder("history_key", false),
		&schema.Message{
			Role:    schema.User,
			Content: "请帮我{task}。",
		},
	)

	// 准备变量
	variables := map[string]any{
		"role":        "专业的助手",
		"task":        "写一首诗",
		"history_key": []*schema.Message{{Role: schema.User, Content: "告诉我油画是什么?"}, {Role: schema.Assistant, Content: "油画是xxx"}},
	}

	cm, err := models.NewChatModel(ctx)
	if err != nil {
		log.Fatal(err)
	}
	// 在 Chain 中使用
	chain := compose.NewChain[map[string]any, []*schema.Message]()
	chain.AppendChatTemplate(template)
	chain.AppendChatModel(cm)
	// 编译并运行
	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Fatal(err)
	}
	result, err := runnable.Invoke(ctx, variables)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("助手: %s\n", result[0].Content)

	// 在 Graph 中使用
	graph := compose.NewGraph[map[string]any, []*schema.Message]()
	graph.AddChatTemplateNode("template_node", template)
}
