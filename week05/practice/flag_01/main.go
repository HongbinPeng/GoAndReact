// flag_01 main 包
// 本文件通过代码实例 + 详细注释，讲解 Go 语言 flag 标准库的用法和底层原理
package main

import (
	"flag"
	"fmt"
	"time"
)

// 运行方式示例：
// go run main.go --config=custom.json --timeout=5 -v
// go run main.go -config custom.json -timeout 5          ← 短横线 + 空格也可以
// go run main.go -v extra-arg1 extra-arg2                ← 测试位置参数
//
// =============================================================================
// flag 标准库 —— 总览
// =============================================================================
//
// 【这个包是干什么的？】
// flag 是 Go 标准库提供的命令行参数解析工具。
//
// 当你运行一个 Go 程序时，可以在后面跟参数：
//   go run main.go --config=my.json --timeout=5 -v
//
// os.Args 会拿到原始的字符串切片：
//   ["main.go", "--config=my.json", "--timeout=5", "-v"]
//
// flag 包负责把这些字符串解析成结构化的 Go 变量（string、int、bool 等），
// 自动处理各种写法变体（--key=value、--key value、-key value）。
//
// 【为什么作业里必须用它？】
// 作业明确要求支持命令行参数：
//   --config   指定配置文件路径
//   --timeout  指定单次探测超时时间
//   -v         开启详细模式
//
// 如果不学 flag，你就得手动解析 os.Args 字符串，非常麻烦且容易出错。
//
// 【底层原理简述】
//
// flag 包内部维护了一个全局的 FlagSet（默认叫 CommandLine）：
//
//   type FlagSet struct {
//       name     string              // 名字，通常是程序名
//       parsed   bool                // 是否已经调用过 Parse()
//       actual   map[string]*Flag    // 命令行上实际传了的参数
//       formal   map[string]*Flag    // 所有注册过的参数定义
//       args     []string            // 没被 flag 消费的位置参数
//       // ...
//   }
//
//   type Flag struct {
//       Name     string  // 参数名，如 "config"
//       Usage    string  // 帮助信息，如 "配置文件路径"
//       Value    Value   // 存储值的接口（底层是 *stringValue、*intValue 等）
//       DefValue string  // 默认值的字符串表示
//   }
//
// 工作流程：
//   1. flag.String/Int/Bool → 在 formal map 中注册一个 Flag 定义
//   2. flag.Parse() → 遍历 os.Args[1:]，逐个匹配参数名，找到后把值写入对应的 Flag.Value
//   3. 没匹配到的参数 → 放入 args 切片（位置参数）
//   4. 遇到 -- 或 -h → 特殊处理（-- 之后的全部视为位置参数，-h 打印帮助）
//
// 【flag 和第三方库 cobra/pflag 的关系】
//
// flag 是 Go 标准库，简单够用，适合小型 CLI 工具。
// 但它有一些局限：
//   - 不支持短参数名（如 -c 代表 --config）
//   - 不支持子命令（如 git commit、git push 这种）
//   - 帮助信息格式比较简单
//
// 如果需要更强大的 CLI，可以用第三方库：
//   - github.com/spf13/cobra    → 支持子命令、自动生成帮助
//   - github.com/spf13/pflag    → flag 的增强版，支持短参数、POSIX 风格
//
// 作业用标准库 flag 就足够了。

// =============================================================================
// 第一部分：注册参数 —— 定义你的程序支持哪些命令行参数
// =============================================================================
//
// flag 提供了三种基本类型的参数注册函数：
//
//   flag.String(name, defaultValue, usage) → 返回 *string
//   flag.Int(name, defaultValue, usage)    → 返回 *int
//   flag.Bool(name, defaultValue, usage)   → 返回 *bool
//
// 还有更多类型：
//   flag.Uint / flag.Int64 / flag.Float64 / flag.Duration
//   flag.Var(value, name, usage)           → 自定义类型（实现 flag.Value 接口）
//
// 【每个函数的三个参数】
//
//   name         参数名（命令行里用的名字），如 "config"、"timeout"、"v"
//   defaultValue 默认值（命令行没传这个参数时用的值）
//   usage        帮助信息（-help 时显示的描述文字）
//
// 【为什么返回的是指针？】
//
// flag.String/Int/Bool 在内部创建了一个 Flag 对象，把指针存进去。
// Parse() 时会通过指针直接修改底层存储的值。
//
// 底层流程：
//   1. flag.String("config", "config.json", "...")
//      → 内部创建 &stringValue{value: "config.json"}
//      → 注册到 FlagSet.formal["config"]
//      → 返回这个指针给你
//
//   2. flag.Parse() 发现命令行传了 --config=custom.json
//      → 找到 formal["config"]
//      → 调用 Value.Set("custom.json") → 指针指向的值被改了
//
//   3. 你通过 *configPath 读到的就是更新后的值
//
// 如果返回的是值而不是指针，Parse() 改的是内部副本，你拿到的永远是最初的默认值。

// =============================================================================
// 第二部分：Parse() —— 真正的解析动作
// =============================================================================
//
// 【Parse() 做了什么？】
//
//   1. 遍历 os.Args[1:]（跳过程序名本身）
//   2. 对每个参数，尝试解析：
//      a. --key=value  → 拆成 key 和 value
//      b. --key value  → key 是参数名，下一个 arg 是值
//      c. -key value   → 同上（单横线也可以）
//      d. -key=value   → 同上
//      e. -xyz         → 特殊：如果 xyz 都是单字母 bool 参数，可以合并写
//   3. 匹配到注册的参数 → 写入对应指针
//   4. 没匹配到 → 放入位置参数列表（flag.Args()）
//   5. 遇到 -- → 之后的全部视为位置参数
//   6. 遇到 -h / -help / --help → 打印帮助信息并退出
//
// 【为什么必须先 Parse() 再读值？】
//
// flag.String/Int/Bool 只是"注册定义"，这时候参数还没从命令行读进来。
// 指针指向的还是默认值。
//
// 错误示例：
//
//   configPath := flag.String("config", "config.json", "...")
//   fmt.Println(*configPath)  // ← 在 Parse 之前读 → 永远是 "config.json"
//   flag.Parse()
//
// 正确示例：
//
//   configPath := flag.String("config", "config.json", "...")
//   flag.Parse()              // ← 先解析
//   fmt.Println(*configPath)  // ← 再读取 → 可能是用户传的 custom.json
//
// 【Parse() 只会执行一次】
//
// 多次调用 Parse() 不会报错，但第二次及以后的调用什么都不做（parsed 标志位已经设了）。
// 所以你可以在程序的任何地方调用 Parse()，通常放在 main 函数开头。

// =============================================================================
// 第三部分：位置参数（Positional Arguments）
// =============================================================================
//
// 位置参数是那些"没有被 flag 消费"的参数。
//
// 示例：
//   go run main.go --config=a.json file1.txt file2.txt
//
//   --config=a.json  → flag 消费了
//   file1.txt        → 没匹配到任何注册参数 → 位置参数
//   file2.txt        → 位置参数
//
//   flag.Args()  → ["file1.txt", "file2.txt"]
//   flag.Arg(0)  → "file1.txt"
//   flag.NArg()  → 2（位置参数的个数）
//
// 作业里你可能用不到位置参数，但了解一下没坏处。
// 常见用途：git commit -m "msg" file1.txt file2.txt 中的文件名就是位置参数。

// =============================================================================
// 第四部分：支持的命令行写法变体
// =============================================================================
//
// flag 支持的参数写法非常灵活，下面这些写法效果完全一样：
//
//   --config=config.json    ← 双横线 + 等号（最常见）
//   --config config.json    ← 双横线 + 空格
//   -config=config.json     ← 单横线 + 等号
//   -config config.json     ← 单横线 + 空格
//
// Bool 参数更特殊，下面这些写法效果一样：
//
//   -v                      ← 传了 → true
//   -v=true                 ← 显式写 true → true
//   -v=false                ← 显式写 false → false（覆盖默认值）
//   （不传 -v）              ← 用默认值 false
//
// 注意：flag 不支持短参数别名！
//   -c 代表 --config  → 不行！除非你注册了一个叫 "c" 的参数
//   如果你想要短参数别名，得用 github.com/spf13/pflag
//
// 【特殊参数：--】
//
// -- 之后的所有参数都视为位置参数，不再做 flag 解析。
//
//   go run main.go --config=a.json -- file1.txt --fake-flag
//
//   --config=a.json   → 正常解析
//   --                → 分隔符，之后的全是位置参数
//   file1.txt         → 位置参数
//   --fake-flag       → 也是位置参数（不会被当成 flag 解析报错）
//
// 这个特性在你需要传递包含 - 开头的字符串给子程序时很有用。

// =============================================================================
// 第五部分：自定义参数类型（flag.Value 接口）
// =============================================================================
//
// flag.Var 允许你注册自定义类型的参数。
// 只需要实现 flag.Value 接口：
//
//   type Value interface {
//       String() string   // 返回当前值的字符串表示
//       Set(string) error // 从字符串解析并设置值
//   }
//
// 示例：支持 "1h30m" 这种时间格式的参数：
//
//   type DurationValue time.Duration
//
//   func (d *DurationValue) String() string {
//       return time.Duration(*d).String()
//   }
//   func (d *DurationValue) Set(s string) error {
//       v, err := time.ParseDuration(s)
//       if err != nil { return err }
//       *d = DurationValue(v)
//       return nil
//   }
//
//   var timeout DurationValue
//   flag.Var(&timeout, "timeout", "超时时间（如 5s、1m30s）")
//   flag.Parse()
//   // 现在可以这样用：--timeout=5s  --timeout=1m30s
//
// 作业里用 flag.Int 就够了，自定义类型了解一下就行。

// =============================================================================
// 演示代码
// =============================================================================

func main() {
	// ------------------------------------------------------------------------
	// 步骤 1：注册参数（定义你的程序支持哪些命令行选项）
	// ------------------------------------------------------------------------
	//
	// 这三个调用只是在 FlagSet 的 formal map 里登记定义，
	// 此时参数值还是默认值，还没从命令行读进来。
	//
	// 每个调用做三件事：
	//   1. 创建内部 Flag 对象（含 name、usage、defValue）
	//   2. 创建一个 Value 对象（stringValue/intValue/boolValue）存放值
	//   3. 返回 Value 的指针给你
	//
	// 注册之后，-help 会自动显示这些参数的用法说明。

	configPath := flag.String("config", "config.json", "配置文件路径")
	//   参数名: "config"
	//   默认值: "config.json"
	//   帮助信息: "配置文件路径"
	//   返回值: *string（指向内部存储）
	//
	// 命令行写法: --config=path/to/file.json
	//           -config path/to/file.json

	timeoutSeconds := flag.Int("timeout", 3, "单次探测超时时间，单位秒")
	//   参数名: "timeout"
	//   默认值: 3
	//   帮助信息: "单次探测超时时间，单位秒"
	//   返回值: *int
	//
	// 命令行写法: --timeout=5
	//           -timeout 5
	//
	// 注意：这里返回的是 int（秒），作业中需要转成 time.Duration。

	verbose := flag.Bool("v", false, "是否打印详细日志")
	//   参数名: "v"
	//   默认值: false
	//   帮助信息: "是否打印详细日志"
	//   返回值: *bool
	//
	// 命令行写法: -v          → true
	//           -v=false     → false（显式关闭）
	//           （不传 -v）   → false（默认值）

	// ------------------------------------------------------------------------
	// 步骤 2：解析命令行参数（从 os.Args 读值）
	// ------------------------------------------------------------------------
	//
	// Parse() 是真正的"干活"函数。它遍历 os.Args[1:]，
	// 把命令行上传的参数值写入步骤 1 创建的指针指向的存储位置。
	//
	// 如果没有调用 Parse()，上面三个指针读出来的全是默认值。

	flag.Parse()

	// ------------------------------------------------------------------------
	// 步骤 3：使用解析后的参数值
	// ------------------------------------------------------------------------
	//
	// 注意：configPath、timeoutSeconds、verbose 都是指针类型，
	// 读取值时必须解引用（加 *）。
	//
	// 常见错误：忘记加 *，打印出来的是指针地址而不是值。

	// timeoutSeconds 是 *int（秒），需要转成 time.Duration 才能传给 HTTP 客户端。
	// 转换过程：
	//   *timeoutSeconds    → 解引用拿到 int 值，如 5
	//   time.Duration(5)   → 转成 time.Duration 类型，即 5ns（纳秒！）
	//   * time.Second      → 乘以 1 秒 → 5s（5 秒）
	//
	// 为什么要这么绕？因为 time.Duration 的底层单位是纳秒（int64）。
	// time.Duration(5) = 5 纳秒，不是 5 秒。
	// 必须乘以 time.Second 才能得到 5 秒。
	timeout := time.Duration(*timeoutSeconds) * time.Second

	fmt.Println("========== flag 标准库演示 ==========")
	fmt.Printf("配置文件路径：%s\n", *configPath)
	fmt.Printf("超时时间：%v\n", timeout)
	fmt.Printf("是否详细模式：%v\n", *verbose)

	// ------------------------------------------------------------------------
	// 步骤 4：查看位置参数（没被 flag 消费的参数）
	// ------------------------------------------------------------------------
	//
	// flag.Args() 返回 []string，包含所有没匹配到注册参数的命令行参数。
	//
	// 示例：go run main.go -v file1.txt file2.txt
	//   -v         → 被 flag 消费（verbose = true）
	//   file1.txt  → 没匹配到任何参数 → 位置参数
	//   file2.txt  → 位置参数
	//   flag.Args() → ["file1.txt", "file2.txt"]

	if len(flag.Args()) > 0 {
		fmt.Println("\n剩余位置参数：", flag.Args())
	} else {
		fmt.Println("\n没有额外的位置参数。")
	}

	// ------------------------------------------------------------------------
	// 额外演示：查看 -help 自动生成的帮助信息
	// ------------------------------------------------------------------------
	//
	// 运行 go run main.go -help 或 go run main.go --help 会看到：
	//
	//   Usage of ...:
	//     -config string
	//           配置文件路径 (default "config.json")
	//     -timeout int
	//           单次探测超时时间，单位秒 (default 3)
	//     -v    是否打印详细日志
	//
	// flag 自动帮我们格式化了帮助信息，包括：
	//   - 参数名和类型
	//   - 帮助描述
	//   - 默认值（非 bool 类型会显示）
	//
	// 作业里不需要自己处理 -help，flag 包自动处理。

	// ------------------------------------------------------------------------
	// 作业中完整的参数解析模式
	// ------------------------------------------------------------------------
	//
	// 你会在 main 函数开头这样写：
	//
	//   func main() {
	//       configPath := flag.String("config", "config.json", "配置文件路径")
	//       timeoutSec := flag.Int("timeout", 3, "单次探测超时时间（秒）")
	//       verbose := flag.Bool("v", false, "是否开启详细模式")
	//       flag.Parse()
	//
	//       timeout := time.Duration(*timeoutSec) * time.Second
	//
	//       // 验证参数合法性
	//       if *timeoutSec <= 0 {
	//           fmt.Fprintln(os.Stderr, "错误：超时时间必须大于 0")
	//           os.Exit(1)
	//       }
	//
	//       // 加载配置
	//       cfg, err := loadConfig(*configPath)
	//       if err != nil {
	//           fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
	//           os.Exit(1)
	//       }
	//
	//       // 开始探测
	//       runProbe(cfg.Targets, timeout, *verbose)
	//   }
}