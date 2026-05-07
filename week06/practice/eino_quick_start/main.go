package main

import (
	"bufio"
	"context"
	"eino-quickstart/tool"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

func main() {
	// MultiStreamQuestion()
	// MultiGenerateQuestion()
	// tool.ToolDemo()
	// promptsFormat.PromptMultiType()
	// tool.MultiToolParamDemo()
	// tool.InferToolParams()
	// tool.ToolNodeDemo()
	tool.RealToolNodeDemo()
	// react_agent.ReactAgentDemo()
}
func MultiGenerateQuestion() {
	ctx := context.Background()
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
		APIKey:  "sk-e4dcae9060904558a2f64d6eb12249e0",
		Model:   "qwen3.6-plus",
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}
	history := []*schema.Message{
		schema.SystemMessage("你是一个专业的Go语言技术顾问，回答简洁准确。"),
	}
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("开始对话（输入 quit 退出）：")
	for {
		ok := scanner.Scan()
		if !ok {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "quit" {
			break
		}
		history = append(history, schema.UserMessage(input))
		resp, err := chatModel.Generate(ctx, history, model.WithMaxTokens(1024)) //限制输出的token数量，避免输出过长
		if err != nil {
			log.Fatalf("调用模型失败: %v", err)
		}
		fmt.Printf("模型回复：%s\n", resp.Content)
		history = append(history, schema.AssistantMessage(resp.Content, nil))
	}

}
func MultiStreamQuestion() {
	ctx := context.Background()
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
		APIKey:  "sk-e4dcae9060904558a2f64d6eb12249e0",
		Model:   "qwen-plus",
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}
	history := []*schema.Message{
		schema.SystemMessage("你是一个专业的Go语言技术顾问，回答简洁准确。"),
	}
	scanner := bufio.NewScanner(os.Stdin)
	var sb strings.Builder
	fmt.Println("开始对话（输入 quit 退出）：")
	for {
		ok := scanner.Scan()
		if !ok {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "quit" {
			break
		}
		history = append(history, schema.UserMessage(input))
		stream, err := chatModel.Stream(ctx, history)
		if err != nil {
			log.Fatalf("调用模型失败: %v", err)
		}
		fmt.Println("模型回复（流式）：")
		// 循环读取流式数据块
		for {
			chunk, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				// 流结束
				break
			}
			if err != nil {
				log.Fatalf("读取流数据失败: %v", err)
			}
			// 每收到一块就立即输出，不换行
			fmt.Print(chunk.Content)
			sb.WriteString(chunk.Content)
		}
		stream.Close()
		history = append(history, schema.AssistantMessage(sb.String(), nil))
		sb.Reset()    // 清空 StringBuilder
		fmt.Println() // 最后换行
	}
}
