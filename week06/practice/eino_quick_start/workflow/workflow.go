package workflow

import (
	"context"
	"eino-quickstart/models"
	"fmt"
	"log"
	"math/rand/v2"
	"strings"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func WorkflowDemo() {
	ctx := context.Background()

	// Workflow 的整体输入类型。
	//
	// 下面 NewWorkflow[Input, map[string]any]() 的第一个泛型参数就是 Input，
	// 所以 runner.Invoke(ctx, Input{...}) 传进来的 START 节点数据就是这个结构体。
	//
	// 注意：Workflow 的字段映射使用的是 Go 结构体字段名，也就是 Text、Keyword，
	// 不是 json tag。字段必须导出，也就是首字母大写，否则反射取不到。
	type Input struct {
		Text    string
		Keyword string
	}

	// 节点1：统计字符数。
	//
	// 这个 Lambda 的输入类型是 string，输出类型是 int：
	//
	//   string -> int
	//
	// 因为它只需要整段文本，所以后面会用 FromField("Text") 把 START.Text
	// 单独取出来，作为这个节点的完整输入。
	charCounter := func(ctx context.Context, text string) (int, error) {
		return len([]rune(text)), nil
	}

	// 节点2：统计关键词出现次数。
	//
	// 这个 Lambda 的输入类型不是 string，而是 KeywordInput：
	//
	//   KeywordInput -> int
	//
	// 所以后面必须把上游数据组装成 KeywordInput 结构体。
	// MapFields("Text", "FullText") 的意思就是：
	//
	//   上游的 Text 字段 -> 当前节点输入结构体的 FullText 字段
	//
	// MapFields("Keyword", "Word") 的意思就是：
	//
	//   上游的 Keyword 字段 -> 当前节点输入结构体的 Word 字段
	type KeywordInput struct {
		FullText string
		Word     string
	}
	keywordCounter := func(ctx context.Context, input KeywordInput) (int, error) {
		return strings.Count(input.FullText, input.Word), nil
	}

	// 创建 Workflow。
	//
	// 第一个泛型参数 Input：整个 Workflow 的入参类型。
	// 第二个泛型参数 map[string]any：整个 Workflow 的最终出参类型。
	//
	// 可以把 START 理解成一个虚拟节点，它的输出就是 runner.Invoke 传进来的 Input。
	// 可以把 END 理解成另一个虚拟节点，它的输入会被组装成最终返回值 map[string]any。
	wf := compose.NewWorkflow[Input, map[string]any]()

	// 添加字符计数节点。
	//
	// AddLambdaNode("char_count", ...) 创建一个名为 char_count 的节点。
	//
	// AddInput(compose.START, compose.FromField("Text")) 建立两件事：
	//
	// 1. 执行依赖：
	//      char_count 要等 START 数据准备好之后执行。
	//
	// 2. 数据映射：
	//      从 START 的输出 Input 中取 Text 字段，把这个字段值作为 char_count 的完整输入。
	//
	// 伪代码等价于：
	//
	//   start := Input{Text: "...", Keyword: "Go"}
	//   charCounter(ctx, start.Text)
	//
	// FromField("Text") 适合这种场景：下游节点只需要上游的一个字段，
	// 并且这个字段值本身就是下游节点的完整输入。
	wf.AddLambdaNode("char_count", compose.InvokableLambda(charCounter)).
		AddInput(compose.START, compose.FromField("Text"))

	// 添加关键词计数节点。
	//
	// keyword_count 节点的函数签名是：
	//
	//   func(ctx context.Context, input KeywordInput) (int, error)
	//
	// 所以这个节点运行前，Eino 需要先构造一个 KeywordInput。
	//
	// AddInput 里的两个 MapFields 会先产生一个中间 map：
	//
	//   map[string]any{
	//       "FullText": START.Text,
	//       "Word":     START.Keyword,
	//   }
	//
	// 然后 Eino 再根据 keywordCounter 的入参类型 KeywordInput，
	// 用反射把这个 map 转成：
	//
	//   KeywordInput{
	//       FullText: START.Text,
	//       Word:     START.Keyword,
	//   }
	//
	// 最后调用：
	//
	//   keywordCounter(ctx, KeywordInput{...})
	//
	// 这就是你觉得“莫名其妙自动绑定”的地方：它不是 Go 自动做的，
	// 而是 Eino 在 Workflow 编译和运行时用 FieldMapping + reflect 做的。
	wf.AddLambdaNode("keyword_count", compose.InvokableLambda(keywordCounter)).
		AddInput(compose.START,
			// START 的 Text 字段 → KeywordInput 的 FullText 字段
			compose.MapFields("Text", "FullText"),
			// START 的 Keyword 字段 → KeywordInput 的 Word 字段
			compose.MapFields("Keyword", "Word"),
		)

	// END 节点从两个计算节点收集结果。
	//
	// 这里 Workflow 的最终输出类型是 map[string]any，所以 END 会被组装成一个 map。
	//
	// AddInput("char_count", compose.ToField("char_count")) 表示：
	//
	//   把 char_count 节点的完整输出 int 放到最终 map 的 "char_count" key 里。
	//
	// AddInput("keyword_count", compose.ToField("keyword_count")) 表示：
	//
	//   把 keyword_count 节点的完整输出 int 放到最终 map 的 "keyword_count" key 里。
	//
	// 伪代码等价于：
	//
	//   result := map[string]any{
	//       "char_count":    charCountOutput,
	//       "keyword_count": keywordCountOutput,
	//   }
	//
	// ToField("xxx") 适合这种场景：把上游节点的整个输出塞进下游输入的某个字段或 map key。
	wf.End().
		AddInput("char_count", compose.ToField("char_count")).
		AddInput("keyword_count", compose.ToField("keyword_count"))

	// 编译并运行。
	//
	// Compile 阶段会检查这些字段映射是否合法：
	//
	// - START 有没有 Text、Keyword 字段
	// - keyword_count 的输入类型 KeywordInput 有没有 FullText、Word 字段
	// - 字段类型能不能赋值，比如 string -> string 可以，string -> int 不行
	//
	// 运行时才会真正根据映射取值、组装节点输入、执行节点。
	runner, err := wf.Compile(ctx)
	if err != nil {
		log.Fatal("编译失败:", err)
	}

	// Invoke 的入参就是 Workflow 的整体输入 Input。
	//
	// 从这一刻开始，START 的输出就等于这个 Input 值：
	//
	//   START = Input{
	//       Text:    "...",
	//       Keyword: "Go",
	//   }
	//
	// 后面的 AddInput 都是在声明：每个节点从 START 或其他节点输出里取什么数据。
	result, err := runner.Invoke(ctx, Input{
		Text:    "Go语言是一门简洁高效的编程语言，Go的并发模型是Go最大的亮点之一。",
		Keyword: "Go",
	})
	if err != nil {
		log.Fatal("运行失败:", err)
	}

	fmt.Printf("字符总数: %v\n", result["char_count"])
	fmt.Printf("关键词出现次数: %v\n", result["keyword_count"])
}

/*
字段注入 / 字段映射方法速查

这些方法都用在：

	currentNode.AddInput(fromNodeKey, mappings...)

它们描述的是：

	“如何把上游节点 fromNodeKey 的输出，组装成当前节点的输入”

也就是说，永远站在“当前节点输入”这个视角来理解：

	上游输出 ----字段映射----> 当前节点输入

1. compose.FromField("Text")

语义：

	当前节点的完整输入 = 上游输出的 Text 字段

适用场景：

	下游节点只需要上游的某一个字段，并且这个字段值本身就是下游节点的完整输入。

例子：

	wf.AddLambdaNode("char_count", compose.InvokableLambda(charCounter)).
		AddInput(compose.START, compose.FromField("Text"))

伪代码等价于：

	charCounter(ctx, start.Text)

--------------------------------------------------

2. compose.ToField("char_count")

语义：

	当前节点输入的 char_count 字段 / map key = 上游节点的完整输出

适用场景：

	想把某个上游节点的整个输出，塞到当前节点输入的某个字段里。

例子：

	wf.End().
		AddInput("char_count", compose.ToField("char_count"))

伪代码等价于：

	result["char_count"] = charCountOutput

--------------------------------------------------

3. compose.MapFields("Text", "FullText")

语义：

	当前节点输入的 FullText 字段 = 上游输出的 Text 字段

适用场景：

	上游和下游都只取部分字段，并且字段名还不一样。

例子：

	wf.AddLambdaNode("keyword_count", compose.InvokableLambda(keywordCounter)).
		AddInput(compose.START,
			compose.MapFields("Text", "FullText"),
			compose.MapFields("Keyword", "Word"),
		)

伪代码等价于：

	keywordInput := KeywordInput{
		FullText: start.Text,
		Word:     start.Keyword,
	}

--------------------------------------------------
4. compose.FromFieldPath(compose.FieldPath{"User", "Profile", "Name"})
语义：

	当前节点的完整输入 = 上游输出的嵌套字段 User.Profile.Name

适用场景：

	上游是嵌套结构体或 map，只想取其中一条深层路径作为当前节点完整输入。

伪代码等价于：

	currentInput = upstream.User.Profile.Name

--------------------------------------------------

5. compose.ToFieldPath(compose.FieldPath{"Result", "Name"})

语义：

	当前节点输入的嵌套字段 Result.Name = 上游节点的完整输出

适用场景：

	想把上游整个输出写到当前节点输入的某个嵌套字段里。

伪代码等价于：

	currentInput.Result.Name = upstreamOutput

--------------------------------------------------

 6. compose.MapFieldPaths(
    compose.FieldPath{"User", "Profile", "Name"},
    compose.FieldPath{"Author", "Name"},

)

语义：

	当前节点输入的 Author.Name = 上游输出的 User.Profile.Name

适用场景：

	上游和下游都是嵌套结构，并且要做“深层字段 -> 深层字段”的映射。

--------------------------------------------------

7. compose.ToField(..., compose.WithCustomExtractor(...))

语义：

	先用自定义函数从上游输出中提取/计算值，再把这个值写入当前节点输入指定字段。

适用场景：

	简单字段映射不够，想做拼接、格式转换、特殊提取时使用。

--------------------------------------------------

一句话记忆：

	FromXxx：从上游哪里取
	ToXxx：放到当前节点输入哪里
	MapXxx：从上游哪里取，并放到当前节点输入哪里

再简化一层：

	FromField("A")        => currentInput = upstream.A
	ToField("B")          => currentInput.B = upstream
	MapFields("A", "B")   => currentInput.B = upstream.A

注意事项：

1. 这里用的是 Go 结构体字段名，不是 json tag。
2. 字段必须导出，也就是首字母大写，否则反射拿不到。
3. 类型必须可赋值，比如 string -> string 可以，string -> int 不行。
4. Workflow 只是声明依赖和映射；真正的字段读取、组装和写入发生在 Compile/Run 时，由 Eino 内部通过 FieldMapping + reflect 完成。
*/
func WorkflowMapFiledsPathDemo() {
	ctx := context.Background()

	// 输入包含嵌套的 Message 结构
	type Input struct {
		*schema.Message
		Lang string
	}

	// 翻译函数：只需要文本内容和目标语言
	type TranslateInput struct {
		Text       string
		TargetLang string
	}
	translator := func(ctx context.Context, input TranslateInput) (string, error) {
		// 这里用简单的拼接模拟翻译结果
		return fmt.Sprintf("[%s] %s", input.TargetLang, input.Text), nil
	}

	wf := compose.NewWorkflow[Input, map[string]any]()

	wf.AddLambdaNode("translate", compose.InvokableLambda(translator)).
		AddInput(compose.START,
			// 从嵌套路径 Message.Content 映射到 Text,代表Message.Content这个字段的值
			compose.MapFieldPaths([]string{"Message", "Content"}, []string{"Text"}),
			// 顶层字段映射
			compose.MapFields("Lang", "TargetLang"),
		)

	wf.End().AddInput("translate", compose.ToField("result"))

	runner, err := wf.Compile(ctx)
	if err != nil {
		log.Fatal("编译失败:", err)
	}

	result, err := runner.Invoke(ctx, Input{
		Message: &schema.Message{
			Role:    schema.User,
			Content: "今天天气真好",
		},
		Lang: "English",
	})
	if err != nil {
		log.Fatal("运行失败:", err)
	}

	fmt.Println(result["result"])
}
func SetStartValueDemo() {
	ctx := context.Background()

	type BidInput struct {
		Price  float64
		Budget float64
	}

	// 竞价函数：根据当前价格和预算出价
	bidder := func(ctx context.Context, in BidInput) (float64, error) {
		if in.Price >= in.Budget {
			return in.Budget, nil
		}
		return in.Price + rand.Float64()*(in.Budget-in.Price), nil
	}

	wf := compose.NewWorkflow[float64, map[string]float64]()

	// 竞价者1：预算3.0
	wf.AddLambdaNode("bidder1", compose.InvokableLambda(bidder)).
		AddInput(compose.START, compose.ToField("Price")).
		SetStaticValue([]string{"Budget"}, 3.0) // 静态设置预算

	// 竞价者2：预算5.0
	wf.AddLambdaNode("bidder2", compose.InvokableLambda(bidder)).
		AddInput(compose.START, compose.ToField("Price")).
		SetStaticValue([]string{"Budget"}, 5.0) // 不同的预算

	wf.End().
		AddInput("bidder1", compose.ToField("bidder1")).
		AddInput("bidder2", compose.ToField("bidder2"))

	runner, err := wf.Compile(ctx)
	if err != nil {
		log.Fatal("编译失败:", err)
	}

	result, err := runner.Invoke(ctx, 2.0) // 起始价 2.0
	if err != nil {
		log.Fatal("运行失败:", err)
	}

	fmt.Printf("竞价者1出价: %.2f（预算3.0）\n", result["bidder1"])
	fmt.Printf("竞价者2出价: %.2f（预算5.0）\n", result["bidder2"])
}
func ParallelWorkflowDemo() {
	ctx := context.Background()

	// 创建模型
	model, err := models.NewChatModel(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// 定义输入
	type AnalysisInput struct {
		Text string
		TopN string // 提取关键词的数量，用字符串方便模板渲染
	}

	// 情感分析 Prompt 模板
	sentimentTpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个情感分析专家。请分析以下文本的情感倾向，输出格式：情感倾向（正面/负面/中性）+ 一句话理由。"),
		schema.UserMessage("{text}"),
	)

	// 关键词提取 Prompt 模板
	keywordTpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个关键词提取专家。请从以下文本中提取{top_n}个最重要的关键词，用逗号分隔输出。"),
		schema.UserMessage("{text}"),
	)

	// 情感分析链：模板 → 模型
	sentimentChain := compose.NewChain[map[string]any, *schema.Message]()
	sentimentChain.AppendChatTemplate(sentimentTpl).AppendChatModel(model)

	// 关键词提取链：模板 → 模型
	keywordChain := compose.NewChain[map[string]any, *schema.Message]()
	keywordChain.AppendChatTemplate(keywordTpl).AppendChatModel(model)

	// 构建 Workflow
	wf := compose.NewWorkflow[AnalysisInput, string]()

	// 情感分析节点：只需要 Text 字段，映射到模板变量 text
	wf.AddGraphNode("sentiment", sentimentChain).
		AddInput(compose.START, compose.MapFields("Text", "text"))

	// 关键词提取节点：需要 Text 和 TopN 两个字段
	wf.AddGraphNode("keywords", keywordChain).
		AddInput(compose.START,
			compose.MapFields("Text", "text"),
			compose.MapFields("TopN", "top_n"),
		)

	// 合并结果的 Lambda
	wf.AddLambdaNode("merge", compose.InvokableLambda(
		func(ctx context.Context, results map[string]any) (string, error) {
			sentiment := results["sentiment"].(*schema.Message)
			keywords := results["keywords"].(*schema.Message)
			return fmt.Sprintf("=== 文本分析报告 ===\n\n【情感分析】\n%s\n\n【关键词提取】\n%s",
				sentiment.Content, keywords.Content), nil
		},
	)).
		AddInput("sentiment", compose.ToField("sentiment")).
		AddInput("keywords", compose.ToField("keywords"))

	wf.End().AddInput("merge")

	// 编译并运行
	runner, err := wf.Compile(ctx)
	if err != nil {
		log.Fatal("编译失败:", err)
	}

	result, err := runner.Invoke(ctx, AnalysisInput{
		Text: "这款新出的Go语言框架Eino让我眼前一亮，它把大模型应用开发中最头疼的编排问题解决得很优雅，API设计简洁又不失灵活性，字节跳动内部的实战验证也让人放心，唯一的遗憾是文档还不够丰富，社区生态还在成长期。",
		TopN: "5",
	})
	if err != nil {
		log.Fatal("运行失败:", err)
	}

	fmt.Println(result)
}
