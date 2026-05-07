# eino schema/message.go 详解

## 文件概述

`message.go` 是 eino 框架中最核心的数据结构文件，定义了与大语言模型交互时所有消息相关的类型、接口和工具函数。它位于 `github.com/cloudwego/eino/schema` 包中。

---

## 一、核心枚举类型

### 1. FormatType — 模板渲染引擎

```go
type FormatType uint8

const (
    FString    FormatType = 0  // Python f-string 风格: "你好 {name}"
    GoTemplate FormatType = 1  // Go 标准库: "你好 {{.Name}}"
    Jinja2     FormatType = 2  // Python Jinja2: "你好 {{ name }}"
)
```

**作用**：决定模板中占位符的语法格式。

**底层实现**：
- `FString` → 使用 `pyfmt` 库（Python PEP-3101 实现）
- `GoTemplate` → 使用 Go 标准库 `text/template`
- `Jinja2` → 使用 `gonja` 库（Jinja2 的 Go 实现）

**安全限制**：Jinja2 引擎禁用了 `include`、`extends`、`import`、`from` 等关键字，防止模板注入攻击。

### 2. RoleType — 消息角色

```go
type RoleType string

const (
    Assistant RoleType = "assistant"  // AI 助手的回复
    User      RoleType = "user"       // 用户的输入
    System    RoleType = "system"     // 系统提示词
    Tool      RoleType = "tool"       // 工具调用结果
)
```

### 3. ChatMessagePartType — 消息内容类型

```go
const (
    ChatMessagePartTypeText      ChatMessagePartType = "text"       // 纯文本
    ChatMessagePartTypeImageURL  ChatMessagePartType = "image_url"  // 图片
    ChatMessagePartTypeAudioURL  ChatMessagePartType = "audio_url"  // 音频
    ChatMessagePartTypeVideoURL  ChatMessagePartType = "video_url"  // 视频
    ChatMessagePartTypeFileURL   ChatMessagePartType = "file_url"   // 文件
    ChatMessagePartTypeReasoning ChatMessagePartType = "reasoning"  // 推理过程
)
```

### 4. ToolPartType — 工具输出内容类型

```go
const (
    ToolPartTypeText  ToolPartType = "text"   // 文本
    ToolPartTypeImage ToolPartType = "image"  // 图片
    ToolPartTypeAudio ToolPartType = "audio"  // 音频
    ToolPartTypeVideo ToolPartType = "video"  // 视频
    ToolPartTypeFile  ToolPartType = "file"   // 文件
)
```

---

## 二、核心数据结构

### 1. Message — 消息（最核心的类型）

```go
type Message struct {
    Role RoleType `json:"role"`          // 消息角色

    Content string `json:"content"`      // 纯文本内容

    // 多模态输入内容（用户输入的图片、音频等）
    UserInputMultiContent []MessageInputPart `json:"user_input_multi_content,omitempty"`

    // 多模态输出内容（AI 生成的图片、音频等）
    AssistantGenMultiContent []MessageOutputPart `json:"assistant_output_multi_content,omitempty"`

    Name string `json:"name,omitempty"`          // 角色名称

    ToolCalls []ToolCall `json:"tool_calls,omitempty"`   // 工具调用（仅 Assistant 消息）
    ToolCallID string `json:"tool_call_id,omitempty"`    // 工具调用 ID（仅 Tool 消息）
    ToolName string `json:"tool_name,omitempty"`         // 工具名称（仅 Tool 消息）

    ResponseMeta *ResponseMeta `json:"response_meta,omitempty"`  // 响应元数据（token 统计等）

    ReasoningContent string `json:"reasoning_content,omitempty"`  // 模型推理过程

    Extra map[string]any `json:"extra,omitempty"`                 // 自定义额外信息
}
```

**使用示例**：

```go
// 用户消息
schema.UserMessage("你好，请介绍一下你自己")

// 系统消息
schema.SystemMessage("你是一个专业的 Go 语言顾问")

// 助手消息
schema.AssistantMessage("你好！我是 AI 助手", nil)

// 工具消息
schema.ToolMessage("工具执行结果", "call_123")

// 自定义消息
&schema.Message{
    Role:    schema.User,
    Content: "请分析这张图片",
    UserInputMultiContent: []schema.MessageInputPart{
        {Type: schema.ChatMessagePartTypeText, Text: "这是什么？"},
        {Type: schema.ChatMessagePartTypeImageURL, Image: &schema.MessageInputImage{
            MessagePartCommon: schema.MessagePartCommon{
                URL: strPtr("https://example.com/cat.jpg"),
            },
        }},
    },
}
```

### 2. ToolCall — 工具调用

```go
type ToolCall struct {
    Index    *int         `json:"index,omitempty"`     // 多个工具调用时的索引
    ID       string       `json:"id"`                  // 工具调用唯一 ID
    Type     string       `json:"type"`                // 类型，默认 "function"
    Function FunctionCall `json:"function"`            // 具体函数调用
    Extra    map[string]any `json:"extra,omitempty"`   // 额外信息
}

type FunctionCall struct {
    Name      string `json:"name,omitempty"`      // 函数名
    Arguments string `json:"arguments,omitempty"` // JSON 格式的参数
}
```

### 3. ResponseMeta — 响应元数据

```go
type ResponseMeta struct {
    FinishReason string     `json:"finish_reason,omitempty"`  // 停止原因: "stop", "length", "tool_calls"
    Usage        *TokenUsage `json:"usage,omitempty"`         // Token 使用统计
    LogProbs     *LogProbs  `json:"logprobs,omitempty"`       // 日志概率
}

type TokenUsage struct {
    PromptTokens         int `json:"prompt_tokens"`           // 输入 token 数
    CompletionTokens     int `json:"completion_tokens"`       // 输出 token 数
    TotalTokens          int `json:"total_tokens"`            // 总 token 数
    PromptTokenDetails   PromptTokenDetails `json:"prompt_token_details"`
    CompletionTokensDetails CompletionTokensDetails `json:"completion_token_details"`
}
```

### 4. MessagesTemplate — 消息模板接口

```go
type MessagesTemplate interface {
    Format(ctx context.Context, vs map[string]any, formatType FormatType) ([]*Message, error)
}
```

实现了这个接口的类型有：
- `*Message` — 普通消息（支持模板变量替换）
- `messagesPlaceholder` — 消息占位符（动态插入消息列表）

---

## 三、消息占位符

```go
// 创建一个消息占位符
placeholder := schema.MessagesPlaceholder("history", false)

// 使用场景
chatTemplate := prompt.FromMessages(
    schema.FString,
    schema.SystemMessage("你是一个助手"),
    schema.MessagesPlaceholder("history", false),  // ← 这里会被 params 中的 history 替换
    schema.UserMessage("当前问题: {question}"),
)

msgs, _ := chatTemplate.Format(ctx, map[string]any{
    "history": []*schema.Message{
        {Role: "user", Content: "你好"},
        {Role: "assistant", Content: "你好，有什么可以帮你的？"},
    },
    "question": "Go 的 defer 是什么？",
})
```

---

## 四、流式消息合并（关键功能）

LLM 流式输出时，会返回多个 `Message` 碎片，需要合并成完整消息。

### ConcatMessages — 合并消息切片

```go
// 流式接收
msgs := []*schema.Message{}
for {
    msg, err := stream.Recv()
    if errors.Is(err, io.EOF) {
        break
    }
    msgs = append(msgs, msg)
}

// 合并
fullMessage, err := schema.ConcatMessages(msgs)
```

**合并规则**：
| 字段 | 合并方式 |
|------|----------|
| `Content` | 字符串拼接 |
| `ReasoningContent` | 字符串拼接 |
| `ToolCalls` | 按 Index 分组后合并 |
| `MultiContent` | 同类型碎片拼接 |
| `Role` / `Name` | 必须一致，否则报错 |
| `ResponseMeta.Usage` | 取最大值 |
| `FinishReason` | 取最后一个有效值 |

### ConcatMessageStream — 从流读取器直接合并

```go
fullMessage, err := schema.ConcatMessageStream(stream)
// 等价于：Recv() 循环 + ConcatMessages()
```

---

## 五、工具结果处理

### ToolResult — 工具输出结果

```go
type ToolResult struct {
    Parts []ToolOutputPart `json:"parts,omitempty"`
}
```

### ConcatToolResults — 合并工具输出碎片

```go
// 流式工具输出合并
mergedResult, err := schema.ConcatToolResults(toolChunks)
```

**合并规则**：
- 文本碎片 → 拼接
- 非文本碎片（图片、音频等）→ 不合并，每种类型只能在一个 chunk 中出现

### ToMessageInputParts — 将工具结果转为消息输入

```go
toolResult := &schema.ToolResult{
    Parts: []schema.ToolOutputPart{
        {Type: schema.ToolPartTypeText, Text: "搜索结果：北京是中国首都"},
    },
}

// 转为消息输入格式，作为下一轮对话的上下文
inputParts, _ := toolResult.ToMessageInputParts()
```

---

## 六、多模态支持类型层次

```
MessageInputPart（用户多模态输入）
├── Text（文本）
├── Image → MessageInputImage（图片输入）
├── Audio → MessageInputAudio（音频输入）
├── Video → MessageInputVideo（视频输入）
── File  → MessageInputFile（文件输入）

MessageOutputPart（AI 多模态输出）
├── Text（文本）
── Image → MessageOutputImage（图片输出）
├── Audio → MessageOutputAudio（音频输出）
├── Video → MessageOutputVideo（视频输出）
└── Reasoning → MessageOutputReasoning（推理过程）

ToolOutputPart（工具输出）
├── Text（文本）
├── Image → ToolOutputImage（图片）
├── Audio → ToolOutputAudio（音频）
├── Video → ToolOutputVideo（视频）
└── File  → ToolOutputFile（文件）
```

---

## 七、LogProbs — 日志概率

```go
type LogProbs struct {
    Content []LogProb `json:"content"`
}

type LogProb struct {
    Token       string       `json:"token"`        // token 文本
    LogProb     float64      `json:"logprob"`      // 对数概率
    Bytes       []int64      `json:"bytes,omitempty"`  // UTF-8 字节表示
    TopLogProbs []TopLogProb `json:"top_logprobs"` // 最可能的 token 列表
}
```

用于分析模型生成每个 token 时的置信度，常用于：
- 调试模型输出质量
- 检测模型不确定性的部分
- 安全审计（检测可能的幻觉输出）

---

## 八、完整使用流程示例

```go
package main

import (
    "context"
    "fmt"

    "github.com/cloudwego/eino/components/prompt"
    "github.com/cloudwego/eino/schema"
)

func main() {
    ctx := context.Background()

    // 1. 创建消息模板
    chatTemplate := prompt.FromMessages(
        schema.FString,
        schema.SystemMessage("你是一个{role}，回答简洁。"),
        schema.MessagesPlaceholder("history", true),
        schema.UserMessage("问题：{question}"),
    )

    // 2. 渲染模板
    msgs, err := chatTemplate.Format(ctx, map[string]any{
        "role": "Go 语言专家",
        "history": []*schema.Message{
            schema.UserMessage("你好"),
            schema.AssistantMessage("你好！请问有什么可以帮你的？", nil),
        },
        "question": "解释 defer 的作用",
    })

    // 3. 发送给模型（此处省略模型调用）
    // response, err := chatModel.Generate(ctx, msgs)

    // 4. 流式输出时合并消息
    // fullMsg, err := schema.ConcatMessageStream(stream)

    fmt.Printf("渲染后的消息数量: %d\n", len(msgs))
    for _, msg := range msgs {
        fmt.Printf("%s: %s\n", msg.Role, msg.Content)
    }
}
```

---

## 九、文件结构总结

| 分类 | 主要类型/函数 |
|------|---------------|
| **枚举类型** | `FormatType`, `RoleType`, `ChatMessagePartType`, `ToolPartType` |
| **核心结构** | `Message`, `ToolCall`, `ResponseMeta`, `TokenUsage` |
| **模板系统** | `MessagesTemplate` 接口, `MessagesPlaceholder`, `Message.Format()` |
| **多模态** | `MessageInputPart`, `MessageOutputPart`, `ToolOutputPart` |
| **流式合并** | `ConcatMessages`, `ConcatMessageStream`, `ConcatToolResults` |
| **工具结果** | `ToolResult`, `ToMessageInputParts()` |
| **概率分析** | `LogProbs`, `LogProb`, `TopLogProb` |
| **辅助函数** | `SystemMessage()`, `UserMessage()`, `AssistantMessage()`, `ToolMessage()` |
