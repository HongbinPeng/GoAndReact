# Eino 框架学习教程

> 基于本项目 `week06/practice/eino_quick_start` 和本机源码：
>
> - `github.com/cloudwego/eino@v0.8.13`
> - `github.com/cloudwego/eino-ext/components/model/openai@v0.1.13`
>
> 这份教程的目标不是只会跑 demo，而是看懂 Eino 的核心抽象：Message、ChatModel、Prompt、Tool，以及它们如何组合成一个 LLM 应用。

## 1. 项目结构

当前 quick start 项目结构大致如下：

```text
eino_quick_start
├── go.mod
├── go.sum
├── main.go
├── models
│   └── modelDemo.go
├── promptsFormat
│   ├── promptdemo.go
│   └── message详解.md
└── tool
    └── toolDemo.go
```

各目录职责：

| 路径 | 作用 |
| --- | --- |
| `main.go` | 程序入口，演示普通对话、流式对话、工具调用 |
| `models/modelDemo.go` | 封装模型初始化逻辑 |
| `promptsFormat/promptdemo.go` | 演示 Prompt Template、多轮历史、图片消息、Chain |
| `tool/toolDemo.go` | 演示 ToolInfo、WithTools、InvokableTool、JSON 参数转换 |

当前依赖的核心版本：

```go
github.com/cloudwego/eino v0.8.13
github.com/cloudwego/eino-ext/components/model/openai v0.1.13
```

## 2. Eino 的整体设计

Eino 可以理解成一个 Go 语言的 LLM 应用编排框架。它把大模型应用拆成几个稳定组件：

| 组件 | 源码包 | 作用 |
| --- | --- | --- |
| Message | `github.com/cloudwego/eino/schema` | 表示 system/user/assistant/tool 消息 |
| ChatModel | `github.com/cloudwego/eino/components/model` | 统一模型调用接口 |
| Prompt | `github.com/cloudwego/eino/components/prompt` | 把变量格式化为消息列表 |
| Tool | `github.com/cloudwego/eino/components/tool` | 描述和执行工具 |
| Compose | `github.com/cloudwego/eino/compose` | 把 Prompt、Model、Tool 等组件串成 Chain 或 Graph |
| Ext | `github.com/cloudwego/eino-ext/...` | 第三方模型、向量库、组件适配层 |

最核心的调用链通常是：

```text
用户输入
  ↓
Prompt Template 格式化
  ↓
[]*schema.Message
  ↓
ChatModel.Generate / ChatModel.Stream
  ↓
模型返回 *schema.Message
  ↓
如果有 ToolCall，则执行工具
  ↓
把工具结果作为 tool message 再交给模型
  ↓
最终回答
```

你现在的 demo 已经覆盖了前三个核心主题：

1. 普通模型调用：`Generate`
2. 流式模型调用：`Stream`
3. 工具调用：`WithTools` 和 `InvokableRun`

## 3. ChatModel：模型调用抽象

### 3.1 接口源码

Eino 的模型接口在：

```text
github.com/cloudwego/eino@v0.8.13/components/model/interface.go
```

核心接口：

```go
type BaseChatModel interface {
    Generate(ctx context.Context, input []*schema.Message, opts ...Option) (*schema.Message, error)
    Stream(ctx context.Context, input []*schema.Message, opts ...Option) (*schema.StreamReader[*schema.Message], error)
}
```

它说明了两种模型调用方式：

| 方法 | 特点 |
| --- | --- |
| `Generate` | 一次性返回完整回复 |
| `Stream` | 返回流式 reader，逐块读取模型输出 |

输入统一是：

```go
[]*schema.Message
```

输出统一是：

```go
*schema.Message
```

这就是 Eino 的一个关键设计：不管底层是 OpenAI、DashScope、Azure 还是其他模型，业务代码都尽量只面对统一接口。

### 3.2 ToolCallingChatModel

源码里还有一个重要接口：

```go
type ToolCallingChatModel interface {
    BaseChatModel
    WithTools(tools []*schema.ToolInfo) (ToolCallingChatModel, error)
}
```

`WithTools` 的作用是把工具说明绑定到模型实例上。注意它返回的是一个新实例：

```go
cmWithTools, err := cm.WithTools([]*schema.ToolInfo{weatherTool})
```

这比旧的 `BindTools` 更安全，因为它不会修改原来的 `cm`。

你的 `toolDemo.go` 里正好演示了这一点：

```go
cmWithTools, err := cm.WithTools([]*schema.ToolInfo{weatherTool})
```

然后：

```go
resp, err := cmWithTools.Generate(ctx, messages)
```

此时模型知道自己可以调用 `get_weather` 工具。

而原始的 `cm` 没有绑定工具：

```go
resp2, _ := cm.Generate(ctx, messagesNoTool)
```

它不会知道有 `get_weather` 这个工具。

## 4. OpenAI 适配层：为什么可以调用 DashScope

你的模型初始化在：

```go
models/modelDemo.go
```

代码：

```go
chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
    BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
    APIKey:  "...",
    Model:   "qwen3.6-plus",
})
```

这里使用的是：

```go
github.com/cloudwego/eino-ext/components/model/openai
```

虽然包名叫 `openai`，但它支持 OpenAI-compatible API。DashScope 的兼容模式就是这种接口形式，所以可以通过 `BaseURL` 指向 DashScope。

### 4.1 ChatModelConfig

源码位置：

```text
github.com/cloudwego/eino-ext/components/model/openai@v0.1.13/chatmodel.go
```

常见配置：

| 字段 | 作用 |
| --- | --- |
| `APIKey` | 模型服务密钥 |
| `BaseURL` | API 基础地址，兼容服务需要配置 |
| `Model` | 模型名称 |
| `Timeout` | 请求超时时间 |
| `Temperature` | 随机性 |
| `TopP` | nucleus sampling |
| `MaxTokens` | 最大输出 token，旧字段 |
| `MaxCompletionTokens` | 最大 completion token，新字段 |
| `ResponseFormat` | 结构化输出 |
| `ExtraFields` | 传递厂商特有字段 |

建议把 API Key 放到环境变量里，而不是硬编码在源码中：

```go
APIKey: os.Getenv("DASHSCOPE_API_KEY")
```

### 4.2 Generate 的源码链路

OpenAI 适配层里的 `Generate` 实现很薄：

```go
func (cm *ChatModel) Generate(ctx context.Context, in []*schema.Message, opts ...model.Option) (*schema.Message, error) {
    ctx = callbacks.EnsureRunInfo(ctx, cm.GetType(), components.ComponentOfChatModel)
    out, err := cm.cli.Generate(ctx, in, opts...)
    if err != nil {
        return nil, convOrigAPIError(err)
    }
    return out, nil
}
```

它主要做三件事：

1. 给 callback 系统补充运行信息。
2. 调用底层 `cm.cli.Generate`。
3. 把底层 API 错误转换成 Eino ext 的错误类型。

所以你写：

```go
resp, err := chatModel.Generate(ctx, history)
```

本质上就是把 `[]*schema.Message` 转成 OpenAI-compatible 请求，发给模型服务，再把响应转回 `*schema.Message`。

## 5. Message：大模型输入输出的统一载体

Eino 里聊天消息使用：

```go
schema.Message
```

常见构造函数：

```go
schema.SystemMessage("你是一个专业的Go语言技术顾问，回答简洁准确。")
schema.UserMessage("什么是 goroutine？")
schema.AssistantMessage(resp.Content, nil)
```

### 5.1 常见角色

| Role | 含义 |
| --- | --- |
| `system` | 系统指令，控制模型身份、规则、风格 |
| `user` | 用户输入 |
| `assistant` | 模型回复 |
| `tool` | 工具执行结果 |

### 5.2 多轮对话

你在 `main.go` 里维护了历史：

```go
history := []*schema.Message{
    schema.SystemMessage("你是一个专业的Go语言技术顾问，回答简洁准确。"),
}
```

用户每输入一次，就追加 user message：

```go
history = append(history, schema.UserMessage(input))
```

模型回复后，再追加 assistant message：

```go
history = append(history, schema.AssistantMessage(resp.Content, nil))
```

这就是多轮对话的本质：每次请求都把必要的历史消息重新发给模型。

模型本身不会自动记住你本地程序里的历史，历史是你通过 `[]*schema.Message` 传进去的。

## 6. Generate：普通非流式调用

你在 `main.go` 里的 `MultiGenerateQuestion` 演示了普通对话：

```go
resp, err := chatModel.Generate(ctx, history, model.WithMaxTokens(1024))
```

流程：

```text
读取用户输入
  ↓
append UserMessage
  ↓
Generate(ctx, history)
  ↓
打印 resp.Content
  ↓
append AssistantMessage
```

适合场景：

| 场景 | 是否适合 Generate |
| --- | --- |
| 短回答 | 适合 |
| 后台任务 | 适合 |
| 不需要边生成边展示 | 适合 |
| 长文本实时展示 | 更适合 Stream |

## 7. Stream：流式调用

你在 `main.go` 里的 `MultiStreamQuestion` 演示了流式输出：

```go
stream, err := chatModel.Stream(ctx, history)
```

然后循环读取：

```go
for {
    chunk, err := stream.Recv()
    if errors.Is(err, io.EOF) {
        break
    }
    if err != nil {
        log.Fatalf("读取流数据失败: %v", err)
    }
    fmt.Print(chunk.Content)
    sb.WriteString(chunk.Content)
}
stream.Close()
```

关键点：

1. `Stream` 返回的是 `*schema.StreamReader[*schema.Message]`。
2. 每次 `Recv()` 得到一个消息片段。
3. 遇到 `io.EOF` 表示流结束。
4. 最后要 `Close()`。
5. 如果要继续多轮对话，需要把流式片段拼成完整回复，再放进 history。

你这里用 `strings.Builder` 做拼接：

```go
var sb strings.Builder
sb.WriteString(chunk.Content)
history = append(history, schema.AssistantMessage(sb.String(), nil))
```

这是正确思路。

## 8. Prompt Template：把变量变成消息列表

Prompt 相关源码：

```text
github.com/cloudwego/eino@v0.8.13/components/prompt/chat_template.go
```

核心创建方式：

```go
template := prompt.FromMessages(schema.FString,
    schema.SystemMessage("你是一个{role}。"),
    schema.MessagesPlaceholder("history_key", false),
    &schema.Message{
        Role:    schema.User,
        Content: "请帮我{task}。",
    },
)
```

这里的 `schema.FString` 表示使用类似 Python f-string 的 `{变量名}` 占位格式。

格式化：

```go
messages, err := template.Format(ctx, map[string]any{
    "role": "专业的助手",
    "task": "写一首诗",
    "history_key": []*schema.Message{
        schema.UserMessage("告诉我油画是什么?"),
        schema.AssistantMessage("油画是xxx", nil),
    },
})
```

结果是：

```go
[]*schema.Message
```

也就是说，Prompt Template 的职责不是直接调用模型，而是生成模型需要的消息数组。

### 8.1 MessagesPlaceholder

你在 `PromptDemo` 里用了：

```go
schema.MessagesPlaceholder("history", false)
```

它表示这里可以插入一段历史消息。

模板：

```go
template := prompt.FromMessages(schema.FString,
    schema.SystemMessage("你是一个Go语言教学助手。回答简洁明了。"),
    schema.MessagesPlaceholder("history", false),
    schema.UserMessage("{input}"),
)
```

变量：

```go
messages, _ := template.Format(ctx, map[string]any{
    "history": trimmed,
    "input":   input,
})
```

最终会形成：

```text
system: 你是一个Go语言教学助手。回答简洁明了。
历史 user/assistant 消息...
user: 当前 input
```

### 8.2 控制历史长度

你的 `trimHistory` 很重要：

```go
func trimHistory(history []*schema.Message, maxRounds int) []*schema.Message {
    maxMessages := maxRounds * 2
    if len(history) <= maxMessages {
        return history
    }
    return history[len(history)-maxMessages:]
}
```

为什么要裁剪历史？

1. 模型上下文长度有限。
2. 历史越多，请求越慢。
3. token 越多，成本越高。
4. 很早的上下文可能会干扰当前问题。

在教学 demo 中保留最近 3 轮是合理的。

## 9. 多模态消息

你在 `PromptMultiType` 中构造了图片输入：

```go
&schema.Message{
    Role: schema.User,
    UserInputMultiContent: []schema.MessageInputPart{
        {
            Type: schema.ChatMessagePartTypeText,
            Text: "下面这些图片是什么？",
        },
        {
            Type: schema.ChatMessagePartTypeImageURL,
            Image: &schema.MessageInputImage{
                MessagePartCommon: schema.MessagePartCommon{
                    URL:      Ptr("https://...jpeg"),
                    MIMEType: "image/jpeg",
                },
            },
        },
    },
}
```

普通文本消息使用：

```go
Content: "文本"
```

多模态消息则使用：

```go
UserInputMultiContent: []schema.MessageInputPart{...}
```

注意：多模态能不能成功，取决于你使用的模型是否支持图片输入。不是所有 OpenAI-compatible 模型都支持视觉输入。

## 10. ToolInfo：给模型看的工具说明书

工具相关源码：

```text
github.com/cloudwego/eino@v0.8.13/schema/tool.go
github.com/cloudwego/eino@v0.8.13/components/tool/interface.go
```

你的工具信息：

```go
weatherTool := &schema.ToolInfo{
    Name: "get_weather",
    Desc: "查询指定城市的当前天气",
    ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
        "city": {
            Type:     schema.String,
            Desc:     "城市名称，如：北京、上海",
            Required: true,
        },
    }),
}
```

`ToolInfo` 的作用是告诉模型：

| 字段 | 含义 |
| --- | --- |
| `Name` | 工具名，模型调用时会使用这个名字 |
| `Desc` | 工具用途，模型根据描述判断什么时候调用 |
| `ParamsOneOf` | 工具参数 schema |

这不是给 Go 编译器看的，而是给模型看的。

### 10.1 ParamsOneOf

源码里 `ToolInfo` 使用：

```go
*ParamsOneOf
```

它支持两种参数描述方式：

1. `schema.NewParamsOneOfByParams`
2. `schema.NewParamsOneOfByJSONSchema`

你现在用的是轻量方式：

```go
schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
    "city": {
        Type:     schema.String,
        Desc:     "要查询天气的城市名称，如：北京、上海、深圳",
        Required: true,
    },
})
```

这会被转换成 JSON Schema，再传给模型。

模型看到的大概意思是：

```json
{
  "type": "object",
  "properties": {
    "city": {
      "type": "string",
      "description": "要查询天气的城市名称，如：北京、上海、深圳"
    }
  },
  "required": ["city"]
}
```

### 10.2 为什么参数叫 city

你的工具函数是：

```go
func getWeather(ctx context.Context, req *WeatherRequest) (*WeatherResponse, error)
```

请求结构体：

```go
type WeatherRequest struct {
    City string `json:"city" jsonschema:"required" jsonschema_description:"城市名称"`
}
```

`city` 对应的是 JSON 字段名，不是 Go 函数参数名。

模型调用工具时会生成：

```json
{"city":"北京"}
```

Eino 内部再把它反序列化为：

```go
&WeatherRequest{
    City: "北京",
}
```

所以这几个地方要一致：

```text
ToolInfo.ParamsOneOf 里的 "city"
WeatherRequest 的 json:"city"
模型生成的 {"city":"北京"}
```

如果 ToolInfo 写成 `"location"`，但结构体还是 `json:"city"`，模型就可能生成：

```json
{"location":"北京"}
```

这样 `req.City` 就拿不到值。

## 11. Tool 接口：Info 和 InvokableRun

源码位置：

```text
github.com/cloudwego/eino@v0.8.13/components/tool/interface.go
```

核心接口：

```go
type BaseTool interface {
    Info(ctx context.Context) (*schema.ToolInfo, error)
}

type InvokableTool interface {
    BaseTool
    InvokableRun(ctx context.Context, argumentsInJSON string, opts ...Option) (string, error)
}
```

### 11.1 Info 是什么

`Info(ctx)` 返回工具元信息：

```go
info, _ := weatherTool.Info(ctx)
fmt.Printf("工具名: %s\n", info.Name)
fmt.Printf("工具描述: %s\n", info.Desc)
```

它不会执行工具，只是返回工具说明书。

模型需要通过 `Info` 或 `ToolInfo` 知道工具怎么用：

```text
工具名是什么
工具描述是什么
参数有哪些
哪些参数必填
```

### 11.2 InvokableRun 是什么

`InvokableRun` 才是真正执行工具：

```go
args := `{"city": "北京"}`
result, err := weatherTool.InvokableRun(ctx, args)
```

输入：

```go
argumentsInJSON string
```

输出：

```go
string
```

在普通工具里，这个 string 通常也是 JSON 字符串。

## 12. utils.NewTool：Go 函数如何变成 Eino Tool

源码位置：

```text
github.com/cloudwego/eino@v0.8.13/components/tool/utils/invokable_func.go
```

你现在的代码：

```go
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
```

`NewTool` 的签名：

```go
func NewTool[T, D any](desc *schema.ToolInfo, i InvokeFunc[T, D], opts ...Option) tool.InvokableTool
```

`InvokeFunc` 的定义：

```go
type InvokeFunc[T, D any] func(ctx context.Context, input T) (output D, err error)
```

你的函数：

```go
func getWeather(ctx context.Context, req *WeatherRequest) (*WeatherResponse, error)
```

所以泛型会推导成：

```go
T = *WeatherRequest
D = *WeatherResponse
```

也就是说，Eino 包装出来的是一个：

```go
invokableTool[*WeatherRequest, *WeatherResponse]
```

## 13. 工具调用内部实现

`InvokableRun` 的核心逻辑：

```go
func (i *invokableTool[T, D]) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (output string, err error) {
    var inst T

    inst = generic.NewInstance[T]()

    err = sonic.UnmarshalString(arguments, &inst)
    if err != nil {
        return "", err
    }

    resp, err := i.Fn(ctx, inst, opts...)
    if err != nil {
        return "", err
    }

    output, err = marshalString(resp)
    if err != nil {
        return "", err
    }

    return output, nil
}
```

你的工具调用链：

```text
args := `{"city":"北京"}`
  ↓
sonic.UnmarshalString(arguments, &inst)
  ↓
inst = &WeatherRequest{City: "北京"}
  ↓
i.Fn(ctx, inst)
  ↓
getWeather(ctx, req)
  ↓
resp = &WeatherResponse{City:"北京", Temp:"22°C", Weather:"晴"}
  ↓
marshalString(resp)
  ↓
result = `{"city":"北京","temp":"22°C","weather":"晴"}`
```

### 13.1 结构体指针如何变成 JSON 字符串

源码位置：

```text
github.com/cloudwego/eino@v0.8.13/components/tool/utils/common.go
```

实现：

```go
func marshalString(resp any) (string, error) {
    if rs, ok := resp.(string); ok {
        return rs, nil
    }
    return sonic.MarshalString(resp)
}
```

所以：

| 工具函数返回值 | InvokableRun 返回 |
| --- | --- |
| `string` | 原样返回 |
| `struct` | JSON 字符串 |
| `*struct` | JSON 字符串 |
| `map` | JSON 字符串 |
| `slice` | JSON 字符串 |

你的 `*WeatherResponse` 会被 `sonic.MarshalString` 转成：

```json
{"city":"北京","temp":"22°C","weather":"晴"}
```

这就是为什么后面可以：

```go
var resp WeatherResponse
json.Unmarshal([]byte(result), &resp)
```

因为 `result` 已经是 JSON 字符串了。

## 14. NewTool 和 InferTool 的区别

源码中还有：

```go
func InferTool[T, D any](toolName, toolDesc string, i InvokeFunc[T, D], opts ...Option) (tool.InvokableTool, error)
```

注释里说明：

```text
InferTool creates an InvokableTool by inferring the parameter JSON schema from the fields and tags of the input type T.
```

也就是说：

| 方法 | 特点 |
| --- | --- |
| `utils.NewTool` | 你手写 `ToolInfo` 和参数 schema |
| `utils.InferTool` | 根据入参结构体 tag 自动生成 ToolInfo 参数 schema |

你的当前写法：

```go
weatherTool := utils.NewTool(toolInfo, getWeather)
```

需要手动保证：

```text
ToolInfo 里的 city
WeatherRequest 里的 json:"city"
```

保持一致。

如果改成 `InferTool`，可以减少重复：

```go
weatherTool, err := utils.InferTool(
    "get_weather",
    "查询指定城市的实时天气信息，包括温度和天气状况",
    getWeather,
)
if err != nil {
    log.Fatal(err)
}
```

这时 Eino 会根据：

```go
type WeatherRequest struct {
    City string `json:"city" jsonschema:"required" jsonschema_description:"城市名称"`
}
```

自动生成参数 schema。

学习阶段建议你两种都试一下：

1. 先用 `NewTool`，理解 ToolInfo 到 JSON Schema 的过程。
2. 再用 `InferTool`，理解结构体 tag 如何自动生成 schema。

## 15. WithTools 只负责让模型“知道工具”，不负责执行工具

这是一个非常容易混淆的点。

你的 `ToolDemo` 里：

```go
cmWithTools, err := cm.WithTools([]*schema.ToolInfo{weatherTool})
resp, err := cmWithTools.Generate(ctx, messages)
```

这一步只会让模型返回工具调用意图，例如：

```text
工具名: get_weather
参数: {"city":"北京"}
```

它不会自动调用你的本地 `getWeather` 函数。

也就是说：

```go
WithTools([]*schema.ToolInfo{...})
```

给模型的是工具说明书，不是工具实现。

真正执行工具需要：

```go
weatherTool.InvokableRun(ctx, args)
```

或者使用 Eino 的 `compose.ToolsNode` 这类编排节点，把“模型产生 tool call -> 执行本地工具 -> 把结果交回模型”串起来。

简单理解：

```text
ToolInfo: 告诉模型有什么工具
InvokableTool: 本地真正能执行的工具
ToolsNode: 自动执行工具调用的编排节点
```

## 16. 完整工具调用闭环

一个完整 Tool Calling 应用通常是：

```text
1. 定义 Go 请求结构体
2. 定义 Go 响应结构体
3. 编写工具函数
4. 用 NewTool 或 InferTool 包装成 InvokableTool
5. 调用 tool.Info(ctx) 得到 ToolInfo
6. 用 cm.WithTools([]*schema.ToolInfo{info}) 绑定给模型
7. 模型 Generate 返回 ToolCalls
8. 根据 ToolCall 名称找到本地 InvokableTool
9. 调用 InvokableRun(ctx, arguments)
10. 把工具结果作为 tool message 追加进 messages
11. 再调用模型生成最终自然语言回复
```

伪代码：

```go
invokable := utils.NewTool(toolInfo, getWeather)
info, _ := invokable.Info(ctx)

cmWithTools, _ := cm.WithTools([]*schema.ToolInfo{info})

resp, _ := cmWithTools.Generate(ctx, messages)

for _, tc := range resp.ToolCalls {
    if tc.Function.Name == "get_weather" {
        result, _ := invokable.InvokableRun(ctx, tc.Function.Arguments)

        messages = append(messages, resp)
        messages = append(messages, schema.ToolMessage(result, tc.ID))
    }
}

finalResp, _ := cmWithTools.Generate(ctx, messages)
fmt.Println(finalResp.Content)
```

具体 `ToolMessage` 的构造函数名字可以根据你当前 Eino 版本的 `schema/message.go` 查看。不同版本可能略有差异，但思想不变：工具结果必须带上对应的 tool call id，模型才能知道这是哪次工具调用的结果。

## 17. Chain：把组件串起来

你在 `PromptInChain` 里用了：

```go
chain := compose.NewChain[map[string]any, []*schema.Message]()
chain.AppendChatTemplate(template)
chain.AppendChatModel(cm)
```

这段表达的是：

```text
map[string]any 变量
  ↓
ChatTemplate
  ↓
[]*schema.Message
  ↓
ChatModel
  ↓
模型输出
```

然后编译：

```go
runnable, err := chain.Compile(ctx)
```

运行：

```go
result, err := runnable.Invoke(ctx, variables)
```

Chain 适合线性流程，比如：

```text
Prompt -> Model
Prompt -> Model -> Parser
Retriever -> Prompt -> Model
```

如果流程有分支、并发、多节点依赖，就更适合 Graph。

## 18. Graph：更复杂的流程编排

你在 `PromptInChain` 里有一个 Graph 的开头：

```go
graph := compose.NewGraph[map[string]any, []*schema.Message]()
graph.AddChatTemplateNode("template_node", template)
```

Graph 的思想是给每个组件一个节点名，然后用边连接节点。

适合场景：

| 场景 | 是否适合 Graph |
| --- | --- |
| 单纯 Prompt -> Model | Chain 更简单 |
| 多个工具节点 | Graph 更合适 |
| 条件分支 | Graph 更合适 |
| RAG 检索 + 重排 + 生成 | Graph 更合适 |
| Agent 多步骤执行 | Graph 更合适 |

你当前阶段可以先掌握 Chain，再学 Graph。

## 19. 学习 Eino 源码的推荐路线

建议按这个顺序看源码：

### 19.1 schema 层

先看：

```text
schema/message.go
schema/tool.go
schema/stream.go
```

重点理解：

| 文件 | 重点 |
| --- | --- |
| `message.go` | 消息角色、文本消息、多模态消息、ToolCall |
| `tool.go` | ToolInfo、ParameterInfo、ParamsOneOf、JSON Schema |
| `stream.go` | StreamReader 的 Recv/Close |

### 19.2 components 接口层

再看：

```text
components/model/interface.go
components/tool/interface.go
components/prompt/chat_template.go
```

重点理解：

| 接口 | 核心问题 |
| --- | --- |
| `BaseChatModel` | 模型统一怎么调用 |
| `ToolCallingChatModel` | 工具说明怎么绑定 |
| `BaseTool` | 工具信息怎么暴露 |
| `InvokableTool` | 工具怎么执行 |
| `DefaultChatTemplate` | 变量怎么格式化成消息 |

### 19.3 utils 工具层

看：

```text
components/tool/utils/invokable_func.go
components/tool/utils/common.go
```

重点理解：

| 函数 | 作用 |
| --- | --- |
| `NewTool` | 手写 ToolInfo + Go 函数 -> InvokableTool |
| `InferTool` | Go struct tag -> ToolInfo + InvokableTool |
| `InvokableRun` | JSON 参数 -> Go 入参 -> Go 输出 -> JSON 字符串 |
| `marshalString` | 工具输出如何转成 string |

### 19.4 ext 适配层

看：

```text
eino-ext/components/model/openai/chatmodel.go
eino-ext/components/model/openai/option.go
```

重点理解：

| 内容 | 作用 |
| --- | --- |
| `ChatModelConfig` | 模型配置 |
| `NewChatModel` | 创建模型实例 |
| `Generate` | 调用底层 client |
| `Stream` | 调用底层 streaming |
| `WithTools` | 返回带工具说明的新模型实例 |
| `WithExtraFields` | 传厂商特有参数 |

## 20. 常见易错点

### 20.1 WithTools 不等于执行工具

`WithTools` 只是把工具 schema 给模型，模型会返回 ToolCall。

本地执行工具还需要 `InvokableRun` 或 `ToolsNode`。

### 20.2 ToolInfo 参数名要和 json tag 一致

正确：

```go
ParamsOneOf: map[string]*schema.ParameterInfo{
    "city": {...},
}

type WeatherRequest struct {
    City string `json:"city"`
}
```

错误：

```go
ParamsOneOf: map[string]*schema.ParameterInfo{
    "location": {...},
}

type WeatherRequest struct {
    City string `json:"city"`
}
```

### 20.3 InvokableRun 的输入输出都是 string

输入：

```go
args := `{"city":"北京"}`
```

输出：

```go
result := `{"city":"北京","temp":"22°C","weather":"晴"}`
```

如果你想在 Go 代码里继续使用结构体，要自己反序列化：

```go
var resp WeatherResponse
json.Unmarshal([]byte(result), &resp)
```

### 20.4 流式输出要拼接成完整 assistant message

流式读取时，每个 chunk 只是片段。要放进 history，需要拼完整：

```go
var sb strings.Builder
sb.WriteString(chunk.Content)
history = append(history, schema.AssistantMessage(sb.String(), nil))
```

### 20.5 API Key 不建议硬编码

建议：

```go
APIKey: os.Getenv("DASHSCOPE_API_KEY")
```

运行前设置环境变量。

## 21. 推荐练习

### 练习 1：把 Weather Tool 改成 InferTool

目标：理解结构体 tag 自动生成 schema。

```go
weatherTool, err := utils.InferTool(
    "get_weather",
    "查询指定城市的实时天气信息，包括温度和天气状况",
    getWeather,
)
if err != nil {
    log.Fatal(err)
}
```

然后打印：

```go
info, _ := weatherTool.Info(ctx)
fmt.Printf("%+v\n", info)
```

观察 ToolInfo 是否自动生成。

### 练习 2：增加一个日期参数

把请求结构体改成：

```go
type WeatherRequest struct {
    City string `json:"city" jsonschema:"required" jsonschema_description:"城市名称"`
    Date string `json:"date" jsonschema_description:"日期，如 today、tomorrow"`
}
```

然后让模型生成：

```json
{"city":"北京","date":"today"}
```

### 练习 3：完成工具调用闭环

当前 `ToolDemo` 只打印了模型要调用的工具：

```go
fmt.Printf("参数: %s\n", tc.Function.Arguments)
```

你可以继续：

1. 根据工具名找到本地 `weatherTool`。
2. 调用 `InvokableRun`。
3. 把结果作为 tool message 追加回 messages。
4. 再调用模型生成自然语言最终回答。

这是从“工具声明”走向“真正 Agent”的关键一步。

### 练习 4：把 PromptInChain 跑通

你已经写了：

```go
chain := compose.NewChain[map[string]any, []*schema.Message]()
chain.AppendChatTemplate(template)
chain.AppendChatModel(cm)
```

可以重点观察：

```text
输入类型 map[string]any
中间类型 []*schema.Message
输出类型是什么
```

如果编译报泛型类型错误，就顺着 `AppendChatModel` 的类型签名看 Chain 的输入输出约束。

## 22. 一句话总结

Eino 的核心思想是：用 `schema.Message` 统一大模型上下文，用 `ChatModel` 统一模型调用，用 `Prompt` 生成消息，用 `ToolInfo` 告诉模型工具能力，用 `InvokableTool` 在本地执行工具，再用 `compose` 把这些组件编排成完整应用。

你现在最该吃透的一条线是：

```text
WeatherRequest / WeatherResponse
  ↓
getWeather
  ↓
utils.NewTool
  ↓
Info 得到 ToolInfo
  ↓
WithTools 绑定给模型
  ↓
模型返回 ToolCall JSON
  ↓
InvokableRun 执行本地函数
  ↓
返回 JSON 字符串
```

把这条线搞明白，Eino 的 Tool Calling 就已经入门了。
