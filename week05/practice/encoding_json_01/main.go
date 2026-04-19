// encoding_json_01 main 包
// 本文件通过代码实例 + 详细注释，讲解 Go 语言 encoding/json 标准库的用法和底层原理
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// 运行方式：
// cd week05/practice/encoding_json_01
// go run main.go

// =============================================================================
// encoding/json 标准库 —— 总览
// =============================================================================
//
// 【这个包是干什么的？】
// encoding/json 是 Go 标准库提供的 JSON 编解码工具。它负责在 Go 的内存数据结构
//（结构体、切片、map 等）和 JSON 文本（[]byte）之间做双向转换。
//
// 【为什么作业里必须用它？】
// 监控器作业有两个核心需求：
// 1. 读取 config.json 配置文件 → 转成 Go 结构体，方便程序里用字段取值
// 2. 把探测结果（结构体切片）→ 转成 JSON 文本，按要求格式输出到终端或文件
//
// 【底层原理简述】
// json 包内部使用了 Go 的 reflect（反射）机制。Marshal/Unmarshal 时会：
//   1. 通过反射读取结构体的所有导出字段（首字母大写的字段）
//   2. 读取每个字段的 struct tag（`json:"..."`），决定 JSON 键名
//   3. 根据 Go 类型 → JSON 类型的映射规则，逐字段编码/解码
//   4. 内部有类型缓存（sync.Map），相同结构体不会每次重复反射，提升性能
//
// 【Go 类型 ↔ JSON 类型 对应关系】
//
//   Go 类型              →   JSON 类型
//   string             →   字符串 "hello"
//   int / float64      →   数字 42 / 3.14
//   bool               →   布尔 true / false
//   []byte             →   base64 编码的字符串
//   struct             →   对象 {"key": value, ...}
//   map[string]T       →   对象 {"key": value, ...}
//   slice/array        →   数组 [1, 2, 3]
//   nil                →   null
//
// =============================================================================

// =============================================================================
// 结构体定义 —— 对应监控器作业的数据模型
// =============================================================================

// Config 是整个配置文件的顶层结构体。
// 对应作业中 config.json 的顶层格式：{ "targets": [...] }
//
// struct tag 语法：`json:"字段名,选项"`
//   - 字段名：JSON 中的键名（小写下划线风格，是前端/配置的常见风格）
//   - 选项（可选）：omitempty、string 等
type Config struct {
	// Targets 是探测目标列表。
	// JSON 中的键名是 "targets"（小写），但 Go 字段必须大写开头才能被 json 包访问。
	// tag 里的 "targets" 告诉 json 包：JSON 里找 "targets" 这个键，填到 Targets 字段。
	Targets []Target `json:"targets"`
}

// Target 对应 config.json 中 targets 数组的每一个元素。
//
// 【struct tag 详解】
//
// `json:"name"`
//
//	最基本的形式。JSON 中必须有 "name" 这个键，值会填到 Name 字段。
//	如果 JSON 中没有这个键，Name 会保持零值（空字符串 ""）。
//
// `json:"retry_count,omitempty"`
//
//	分号分隔两部分：
//	- "retry_count" → JSON 键名（Go 里不能用下划线开头，所以用驼峰 RetryCount）
//	- omitempty     → Marshal（结构体→JSON）时，如果字段是零值，就不输出这个键
//	                  注意：Unmarshal（JSON→结构体）时 omitempty 不起作用！
//
// 【零值是什么？】
//
//	每种 Go 类型都有"零值"（声明变量但未赋值时的默认值）：
//	  string  → ""
//	  int     → 0
//	  bool    → false
//	  slice   → nil
//
// 【为什么作业里 RetryCount 用 int 而不是 *int？】
//
//	用 int：没写 retry_count → 0。但无法区分"没写"和"明确写了 0"。
//	用 *int：没写 retry_count → nil。写了 0 → &0。可以区分。
//	作业场景里，0 和"不写"效果一样（不重试），所以用 int 就够了。
type Target struct {
	Name       string `json:"name"`                  // 目标名称，如 "百度首页"
	Protocol   string `json:"protocol"`              // 探测协议："http" 或 "tcp"
	Address    string `json:"address"`               // 目标地址，如 "https://www.baidu.com"
	RetryCount int    `json:"retry_count,omitempty"` // 失败重试次数，0 表示不重试
	Contains   string `json:"contains,omitempty"`    // HTTP 探测时，期望响应体包含的关键词
}

// ProbeResult 对应一次探测的结果。
// 作业要求输出格式类似：
//
//	{
//	  "name": "百度首页",
//	  "ok": true,
//	  "latency": "80ms",
//	  "error": "连接超时"   ← 只有失败时才出现（此时omitempty 生效）
//	}
type ProbeResult struct {
	Name    string `json:"name"`
	OK      bool   `json:"ok"`
	Latency string `json:"latency"`
	Error   string `json:"error,omitempty"` // 探测成功时 error 是 ""（零值），omitempty 会跳过不输出
}

// =============================================================================
// 第一部分：Unmarshal —— JSON 文本 → Go 结构体
// =============================================================================
//
// 【Unmarshal 的函数签名】
//   func Unmarshal(data []byte, v any) error
//
//   - data：JSON 字节切片（通常从 os.ReadFile 读文件得到）
//   - v：    接收结果的变量，**必须传指针**（&cfg），因为 Unmarshal 需要修改它
//   - 返回：解析失败时的错误（JSON 格式不对、类型不匹配等）
//
// 【底层原理：Unmarshal 是怎么把 JSON 填进结构体的？】
//
//   1. json 包内部有一个 JSON 词法分析器（scanner），逐个读取 JSON token
//   2. 遇到 "{" → 开始解析对象，查找结构体中对应 JSON 键的字段
//   3. 查找匹配规则（优先级从高到低）：
//      a. 完全匹配 tag 名：json:"name" → JSON 中找 "name"
//      b. 忽略大小写匹配：JSON "Name" → Go 字段 Name
//      c. 不匹配 → 跳过该 JSON 键
//   4. 找到字段后，根据 JSON 值的类型做类型转换：
//      - JSON 字符串 → Go string
//      - JSON 数字 → Go int/float64（会自动转换，但如果 JSON 是 3.14 而 Go 是 int，会报错）
//      - JSON 数组 → Go slice（需要元素类型匹配）
//      - JSON null → Go 零值
//   5. 遇到 "}" → 对象结束，返回
//
// 【为什么 Unmarshal 第二个参数必须是指针？】
//
//   func Unmarshal(data []byte, v any) error
//   这里的 v 是 any（即 interface{}），Unmarshal 内部通过反射修改 v 指向的值。
//   如果你传 cfg（值），Unmarshal 修改的是副本，原始 cfg 不受影响。
//   传 &cfg（指针），Unmarshal 通过指针找到原始变量，写入数据。
//   这和 fmt.Scan(&x) 必须传 &x 的道理一样。

func demonstrateUnmarshal() {
	fmt.Println("========== 1. JSON -> 结构体（Unmarshal）==========")

	// rawJSON 模拟从 config.json 读取的内容。
	// 实际作业中你会用 os.ReadFile("config.json") 得到 []byte。
	//
	// 注意：JSON 的键名是下划线风格（retry_count），而 Go 字段是驼峰（RetryCount），
	// 靠 struct tag `json:"retry_count"` 做映射。
	rawJSON := []byte(`{
  "targets": [
    {
      "name": "百度首页",
      "protocol": "http",
      "address": "https://www.baidu.com"
    },
    {
      "name": "Raw 文本文件",
      "protocol": "http",
      "address": "https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt",
      "retry_count": 2,
      "contains": "Go"
    }
  ]
}`)

	// 声明一个 Config 类型的变量。
	// 此时 cfg.Targets 是 nil（切片的零值）。
	var cfg Config

	// Unmarshal 把 rawJSON 解析后填入 cfg。
	// 注意第二个参数是 &cfg（指针），不是 cfg（值）。
	//
	// 解析过程：
	//   1. 看到外层 { → 开始解析对象
	//   2. 看到 "targets" → 在 Config 结构体中查找 json tag 为 "targets" 的字段 → 找到 Targets
	//   3. Targets 的类型是 []Target → 看到 JSON 数组 [...] → 逐个解析数组元素
	//   4. 数组第一个元素 { ... } → 在 Target 结构体中逐字段匹配：
	//      - "name" → json:"name" → Name 字段 → 填入 "百度首页"
	//      - "protocol" → json:"protocol" → Protocol 字段 → 填入 "http"
	//      - "address" → json:"address" → Address 字段 → 填入 URL
	//      - 没有 "retry_count" → RetryCount 保持零值 0
	//      - 没有 "contains" → Contains 保持零值 ""
	//   5. 数组第二个元素同理，但这次有 retry_count 和 contains，会被正确填入
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		// 常见错误：
		//   - JSON 格式错误（少了引号、逗号、括号）
		//   - JSON 中的数字类型和 Go 字段类型不匹配（如 JSON 是字符串 "42"，Go 是 int）
		//   - JSON 中数组元素的类型和 Go 切片元素类型不匹配
		fmt.Println("Unmarshal 失败：", err)
		return
	}

	// 打印解析结果
	fmt.Printf("配置中共有 %d 个目标\n", len(cfg.Targets))
	for i, target := range cfg.Targets {
		// %+v 会打印结构体字段名：{Name:百度首页 Protocol:http Address:https://... RetryCount:0 Contains:}
		fmt.Printf("第 %d 个目标：%+v\n", i+1, target)

		// 演示：作业中你会这样访问字段
		//   url := target.Address
		//   protocol := target.Protocol
		//   retries := target.RetryCount  // 第一个目标是 0，第二个是 2
	}
}

// =============================================================================
// 第二部分：Marshal —— Go 结构体 → JSON 文本
// =============================================================================
//
// 【Marshal 的函数签名】Marshal的中文是”序列化“
//   func Marshal(v any) ([]byte, error)
//
//   - v：任意 Go 值（结构体、切片、map 等）
//   - 返回：JSON 字节切片 + 错误
//
// 【MarshalIndent 的函数签名】
//   func MarshalIndent(v any, prefix, indent string) ([]byte, error)
//
//   - prefix：每行 JSON 的前缀（通常是 ""）
//   - indent：缩进字符串（通常是 "  " 两个空格，或 "\t" 制表符）
//   - 返回：格式化后的 JSON 字节切片 + 错误
//
// 【底层原理：Marshal 是怎么把结构体变成 JSON 的？】
//
//   1. 通过反射读取结构体的所有导出字段（首字母大写的）
//   2. 对每个字段：
//      a. 读取 struct tag，确定 JSON 键名
//      b. 如果字段值是零值且有 omitempty 选项 → 跳过该字段（不输出）
//      c. 否则，将 Go 值编码为 JSON 值：
//         - string → "hello"（内部做 JSON 转义：引号→\"，换行→\n 等）
//         - int → 42
//         - bool → true/false
//         - nil → null
//         - slice → [元素1, 元素2, ...]
//         - struct → {"键1": 值1, "键2": 值2, ...}
//   3. 用 json.Encoder 内部缓冲区拼接最终 JSON 字符串
//   4. 返回 []byte
//
// 【常见错误】
//   - 结构体字段全小写 → 反射看不到 → 输出 {}（空对象）
//   - 循环引用（结构体字段间接引用自身）→ 无限递归 → panic
//   - map 的 key 不是 string/int/float/bool → 报错（JSON 对象键必须是字符串）

func demonstrateMarshal() {
	fmt.Println("\n========== 2. 结构体 -> JSON（Marshal）==========")

	// 模拟一次探测后的结果列表。
	// 作业中你会在并发探测完成后，把所有 goroutine 返回的 ProbeResult 放进一个切片。
	results := []ProbeResult{
		{
			// 探测成功的情况：Error 是零值 ""，omitempty 会让它在 JSON 中不出现
			Name:    "百度首页",
			OK:      true,
			Latency: "80ms",
			// Error: ""  ← 零值，不写也可以，omitempty 会跳过
		},
		{
			// 探测失败的情况：Error 有值，会正常输出
			Name:    "无效服务",
			OK:      false,
			Latency: "3s",
			Error:   "dial timeout",
		},
	}

	// MarshalIndent 生成带缩进的 JSON。
	// 参数解释：
	//   results → 要序列化的 Go 值
	//   ""      → 每行前缀（不用前缀就传空字符串）
	//   "  "    → 缩进字符串（两个空格，也可以用 "\t"）
	//
	// 输出效果：
	//   [
	//     {
	//       "name": "百度首页",
	//       "ok": true,
	//       "latency": "80ms"
	//     },
	//     {
	//       "name": "无效服务",
	//       "ok": false,
	//       "latency": "3s",
	//       "error": "dial timeout"    ← 注意这个字段出现了，因为它不是零值
	//     }
	//   ]
	data, err := json.MarshalIndent(results, "", "  ")
	//这里的MashakIndent()三个参数如果为data := json.MarshalIndent(results, ">>>", "\t")
	/*
		那么就打印：
		>>> {
		>>>   "name": "百度首页",
		>>>   "ok": true,
		>>>   "latency": "80ms"
		>>> },
		>>> {
		>>>   "name": "无效服务",
		>>>   "ok": false,
		>>>   "latency": "3s",
		>>>   "error": "dial timeout"    ← 注意这个字段出现了，因为它不是零值
		>>> }
		>>> ]
	*/
	if err != nil {
		fmt.Println("MarshalIndent 失败：", err)
		return
	}

	// data 是 []byte，用 string(data) 转成字符串打印。
	// 作业中你可以用 os.WriteFile("report.json", data, 0644) 写入文件。
	fmt.Println(string(data))

	// --- 对比：不加缩进的 Marshal ---
	fmt.Println("\n--- 对比：不加缩进的 Marshal ---")
	compactData, _ := json.Marshal(results)
	fmt.Println(string(compactData))
	// 输出（一行）：[{"name":"百度首页","ok":true,"latency":"80ms"},{"name":"无效服务","ok":false,"latency":"3s","error":"dial timeout"}]
	// 适合 API 响应（省带宽），不适合人工阅读。
}

// =============================================================================
// 第三部分：常见陷阱与注意事项
// =============================================================================
//
// 这些是作业中最容易踩的坑，理解了能少走很多弯路。

func demonstrateCommonPitfalls() {
	fmt.Println("\n========== 3. 常见陷阱 ==========")

	// ---------- 陷阱 1：字段首字母小写，Unmarshal 读不进去 ----------
	//
	// json 包内部用 reflect 访问结构体字段，只能看到【导出字段】（首字母大写）。
	// 如果写成小写 name，Unmarshal 会"看不到"这个字段，JSON 中的 "name" 值会被丢弃。
	//
	// ❌ 错误写法：
	//   type Target struct {
	//       name string `json:"name"`  // 小写 n → 无法被 json 包访问
	//   }
	//
	// ✅ 正确写法：
	//   type Target struct {
	//       Name string `json:"name"`  // 大写 N → 可被访问，tag 决定 JSON 键名
	//   }

	fmt.Println("1. 字段必须大写开头，json.Unmarshal 才能写入。")
	fmt.Println("   原因：json 包用 reflect 反射，只能访问导出字段（首字母大写）。")

	// ---------- 陷阱 2：json tag 写错，字段对不上 ----------
	//
	// tag 是反引号 ` 包裹的字符串，不是双引号 "。
	// tag 中的键名必须和 JSON 中的键名完全一致（区分大小写）。
	//
	// ❌ 错误：tag 用双引号
	//   Name string "json:\"name\""  ← 编译能通过，但 json 包不会识别
	//
	// ❌ 错误：tag 键名和 JSON 不匹配
	//   Name string `json:"Name"`    ← JSON 里是 "name"（小写），对不上 → 读不到
	//
	// ✅ 正确：
	//   Name string `json:"name"`    ← JSON 中找 "name"，填到 Name 字段

	fmt.Println("2. json tag 名必须和 JSON 里的字段名完全对应（区分大小写）。")
	fmt.Println("   tag 必须用反引号 ` 包裹，格式为 `json:\"键名\"`。")

	// ---------- 陷阱 3：omitempty 只影响 Marshal，不影响 Unmarshal ----------
	//
	// omitempty 的作用时机：
	//   - Marshal（结构体 → JSON）：字段是零值时，不输出该 JSON 键 ✅
	//   - Unmarshal（JSON → 结构体）：不起任何作用！
	//
	// 也就是说：
	//   - JSON 中没有 "retry_count" → Unmarshal 后 RetryCount = 0（int 的零值）
	//   - 结构体中 RetryCount = 0 → Marshal 时不输出 "retry_count" 这个键
	//
	// 这是一个"不对称"行为，容易让人困惑。

	fmt.Println("3. omitempty 只在 Marshal（结构体→JSON）时生效，Unmarshal 时无视它。")
	fmt.Println("   它的意思是：序列化时，零值字段不输出 JSON 键。")

	// ---------- 陷阱 4：JSON 缺字段 → Go 字段保持零值 ----------
	//
	// 当 JSON 中没有某个键时，Unmarshal 不会报错，对应字段保持其类型的零值：
	//   string → ""
	//   int    → 0
	//   bool   → false
	//
	// 这就带来一个问题：无法区分"JSON 没写这个字段"和"JSON 明确写了 0"。
	//
	// 示例：
	//   JSON: { "retry_count": 0 }    → RetryCount = 0
	//   JSON: { }                      → RetryCount = 0  ← 看起来一样！
	//
	// 如果业务上需要区分这两种情况，用指针类型：
	//   RetryCount *int `json:"retry_count,omitempty"`
	//   JSON: { "retry_count": 0 }    → RetryCount = &0（非 nil 指针，指向 0）
	//   JSON: { }                      → RetryCount = nil（空指针）
	//
	// 但指针类型用起来麻烦（每次取值都要判 nil），作业场景里不需要这么精细。

	fmt.Println("4. JSON 中没有 retry_count 时，RetryCount 会保持 int 的零值 0。")
	fmt.Println("   无法区分 '没写' 和 '明确写了 0'。如需区分，改用 *int。")

	// ---------- 陷阱 5：json.Number 解决大数字精度丢失 ----------
	//
	// 当 Unmarshal 到 interface{} 或 map[string]interface{} 时，
	// JSON 数字默认被解析为 float64。大整数会丢失精度：
	//
	//   JSON: { "id": 9007199254740993 }  ← 超过 float64 精度范围
	//   Go:   id = 9007199254740992       ← 精度丢失！
	//
	// 解决方法：使用 json.Decoder 并开启 UseNumber：
	//
	//   dec := json.NewDecoder(reader)
	//   dec.UseNumber()
	//   dec.Decode(&v)
	//
	// 这样数字会被解析为 json.Number 类型（底层是 string），不会丢失精度。
	// 作业中用结构体直接 Unmarshal 不会有这个问题（int 字段会正确解析）。

	fmt.Println("5. 大整数精度问题：Unmarshal 到 map 时数字会变成 float64，可能丢精度。")
	fmt.Println("   用 json.Decoder + UseNumber() 可以解决，但作业中用结构体不需要担心。")

	// ---------- 陷阱 6：循环引用导致 Marshal 无限递归 ----------
	//
	// 如果结构体字段间接引用自身，Marshal 会无限递归直到栈溢出（panic）。
	//
	// ❌ 危险：
	//   type Node struct {
	//       Name     string `json:"name"`
	//       Children []Node `json:"children"`
	//   }
	//   n := Node{Name: "root"}
	//   n.Children = []Node{n}  // 子节点包含父节点 → 循环引用！
	//   json.Marshal(n)         // panic: 无限递归
	//
	// 解决方法：避免在结构体中创建循环引用，或使用自定义 MarshalJSON。

	fmt.Println("6. 循环引用会导致 Marshal 无限递归 panic，注意避免结构体自引用。")

	// ---------- 陷阱 7：JSON 键名大小写不敏感匹配的坑 ----------
	//
	// Unmarshal 匹配字段的规则是：
	//   1. 优先精确匹配 tag 名（区分大小写）
	//   2. 如果没有 tag 或 tag 不匹配，尝试忽略大小写匹配字段名
	//
	// 这意味着：
	//   JSON: { "NAME": "test" }
	//   Go:   type T struct { Name string }  ← 没有 tag
	//   → 能匹配成功！（"NAME" 忽略大小写匹配 "Name"）
	//
	// 但这不是好事——依赖隐式匹配容易出 bug。最佳实践：
	//   始终给字段加 json tag，不依赖大小写不敏感的 fallback 行为。

	fmt.Println("7. Unmarshal 会做大小写不敏感的 fallback 匹配，但最好始终显式写 tag。")
}

// =============================================================================
// 第四部分：额外实战技巧（作业扩展用）
// =============================================================================

// demonstrateCustomMarshal 演示自定义 JSON 序列化。
// 当默认行为不满足需求时，可以实现 json.Marshaler / json.Unmarshaler 接口。
func demonstrateCustomMarshal() {
	fmt.Println("\n========== 4. 自定义 JSON 序列化 ==========")

	// --- 场景：把 ProbeResult 的 Latency 从 "80ms" 改成只输出数字毫秒 ---
	//
	// 默认情况下，Latency 是 string 类型，直接按字符串输出。
	// 如果想让 JSON 中 latency 是数字（如 80），有两种做法：
	//
	// 做法 A：改用 int 类型（最简单）
	//   type ProbeResult struct {
	//       LatencyMS int `json:"latency_ms"`
	//   }
	//
	// 做法 B：实现 json.Marshaler 接口（灵活，适合复杂逻辑）
	//   type Latency struct { ms int }
	//   func (l Latency) MarshalJSON() ([]byte, error) {
	//       return []byte(fmt.Sprintf("%d", l.ms)), nil
	//   }
	//
	// json.Marshal 会先检查类型是否实现了 Marshaler 接口，
	// 实现了就调用自定义方法，而不是默认的反射逻辑。
	//
	// 同样的，Unmarshal 会检查 Unmarshaler 接口。

	fmt.Println("自定义序列化：实现 json.Marshaler 接口可以完全控制字段的 JSON 输出。")
	fmt.Println("json.Marshal 会优先调用自定义的 MarshalJSON() 方法。")
}

// demonstrateJSONStream 演示流式 JSON 编解码。
// 适合处理大文件或网络流，不需要把整个 JSON 加载到内存。
func demonstrateJSONStream() {
	fmt.Println("\n========== 5. 流式 JSON 编解码（Encoder/Decoder）==========")

	// 【Encoder 和 Decoder 是什么？】
	//
	// json.Marshal/Unmarshal 是一次性把整个 []byte 读完/写完。
	// 如果 JSON 文件很大（几十 MB），全部加载到内存可能浪费。
	//
	// json.Encoder 和 json.Decoder 是流式 API：
	//   - Encoder 写入 io.Writer（如文件、网络连接），边生成边输出
	//   - Decoder 从 io.Reader（如文件、网络连接）读取，边读边解析
	//
	// --- Decoder 示例（读大文件）---
	//
	file, _ := os.Open("large-config.json")
	defer file.Close()
	dec := json.NewDecoder(file)
	var cfg Config
	dec.Decode(&cfg) // 流式解析，不需要把整个文件读进内存
	//
	// --- Encoder 示例（写 JSON Lines）---
	//
	//   file, _ := os.Create("results.jsonl")
	//   defer file.Close()
	//   enc := json.NewEncoder(file)
	//   for _, result := range results {
	//       enc.Encode(result)  // 每行一个 JSON 对象
	//   }
	//
	// 输出格式（JSON Lines）：
	//   {"name":"百度首页","ok":true,"latency":"80ms"}
	//   {"name":"无效服务","ok":false,"latency":"3s","error":"dial timeout"}

	fmt.Println("流式 API 适合大文件：")
	fmt.Println("  json.NewDecoder(reader).Decode(&v)  → 边读边解析，省内存")
	fmt.Println("  json.NewEncoder(writer).Encode(v)   → 边生成边写入，适合 JSON Lines")
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	demonstrateUnmarshal()
	demonstrateMarshal()
	demonstrateCommonPitfalls()
	demonstrateCustomMarshal()
	demonstrateJSONStream()
}
