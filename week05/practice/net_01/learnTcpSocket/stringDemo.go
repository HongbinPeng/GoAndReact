package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	// ================================================================
	// 1. string ↔ []byte ↔ []rune 互相转换
	// ================================================================
	fmt.Println("========== 1. 类型互相转换 ==========")

	s := "Hello\t你好"
	fmt.Printf("原始字符串: %q\n", s) //%q代表什么含义？%q表示字符串的原始值，不解析转义字符，包括引号，%v表示字符串的字面值

	// string → []byte（拷贝底层数据）
	b := []byte(s)
	fmt.Printf("string → []byte: %v\n", b)
	// 英文字母对应ASCII码，中文每个字占3个byte

	// string → []rune（按UTF-8解码成Unicode码位）
	r := []rune(s)
	fmt.Printf("string → []rune: %v\n", r)
	// 每个字符一个rune，中文"你"=20320，"好"=22909

	// []byte → string
	s1 := string(b)
	fmt.Printf("[]byte → string: %s\n", s1)

	// []rune → string
	s2 := string(r)
	fmt.Printf("[]rune → string: %s\n", s2)

	// []byte → []rune（中间经过string转换）
	r2 := []rune(string(b))
	fmt.Printf("[]byte → []rune: %v\n", r2)

	// []rune → []byte（中间经过string转换）
	b2 := []byte(string(r))
	fmt.Printf("[]rune → []byte: %v\n", b2)

	// ================================================================
	// 2. 修改 []byte 不会影响原 string（拷贝关系）
	// ================================================================
	fmt.Println("\n========== 2. 修改 []byte 不影响原 string ==========")

	b3 := []byte("hello")
	s3 := string(b3)
	fmt.Printf("初始: s3=%s, b3=%s\n", s3, b3)

	b3[0] = 'H'
	fmt.Printf("修改b3[0]='H'后: s3=%s, b3=%s\n", s3, b3)
	// s3 不受影响，因为 string(b3) 是一次拷贝

	// 反过来也一样
	s4 := "hello"
	b4 := []byte(s4)
	fmt.Printf("初始: s4=%s, b4=%s\n", s4, b4)

	b4[0] = 'H'
	fmt.Printf("修改b4[0]='H'后: s4=%s, b4=%s\n", s4, b4)
	// s4 仍然不受影响

	// ================================================================
	// 3. 字符串本身不可变（不能直接修改 s[0]）
	// ================================================================
	fmt.Println("\n========== 3. 字符串不可变 ==========")

	s5 := "hello"
	fmt.Printf("原始: %s\n", s5)
	// s5[0] = 'H'  ← 编译错误：cannot assign to s5[0]

	// 想"修改"字符串只能创建新的
	s5 = "H" + s5[1:]
	fmt.Printf("拼接新字符串后: %s\n", s5)

	// 实用技巧：先转[]byte，修改后再转回string
	s6 := "hello"
	b6 := []byte(s6)
	b6[0] = 'H'
	s6 = string(b6)
	fmt.Printf("通过[]byte修改后: %s\n", s6)

	// ================================================================
	// 4. 两种遍历方式：按byte遍历 vs 按rune遍历
	// ================================================================
	fmt.Println("\n========== 4. 遍历方式对比 ==========")

	s7 := "Hi你"

	fmt.Println("--- 按 byte 遍历（s7[i] 返回单个字节）---")
	for i := 0; i < len(s7); i++ {
		fmt.Printf("  byte[%d] = 0x%02x (%d)\n", i, s7[i], s7[i])
	}
	// '你' 占3个byte，所以会看到3行输出

	fmt.Println("--- 按 rune 遍历（for range 自动按UTF-8解码）---")
	for i, ch := range s7 {
		fmt.Printf("  rune[字节偏移%d] = '%c' (U+%04X)\n", i, ch, ch)
	}
	// '你' 只出现1次，注意索引 i 是字节偏移量，不是第几个字符

	// ================================================================
	// 5. 长度对比：len() 返回字节数，不是字符数
	// ================================================================
	fmt.Println("\n========== 5. 长度对比 ==========")

	cases := []string{"Hello", "你好世界", "Go你好🎉"}
	for _, cs := range cases {
		fmt.Printf("字符串: %q\n", cs)
		fmt.Printf("  len(s)                = %d  (UTF-8字节数)\n", len(cs))
		fmt.Printf("  len([]rune(s))        = %d  (字符数/码位数)\n", len([]rune(cs)))
		fmt.Printf("  utf8.RuneCountInString = %d  (同上，但不分配新内存)\n",
			utf8.RuneCountInString(cs))
		fmt.Println()
	}
	// 关键：中文每个字符占3字节，emoji占4字节

	// ================================================================
	// 6. UTF-8 编码细节：每个字符占几个字节
	// ================================================================
	fmt.Println("========== 6. UTF-8编码细节 ==========")

	s8 := "A你😀"
	fmt.Printf("字符串: %q\n", s8)
	fmt.Printf("总字节数: %d, 总字符数: %d\n", len(s8), utf8.RuneCountInString(s8))

	for _, ch := range s8 {
		encoded := string(ch)
		fmt.Printf("  字符 '%c' (U+%04X) → UTF-8占 %d 字节: % x\n",
			ch, ch, len(encoded), encoded)
	}
	// 英文1字节，中文3字节，emoji 4字节

	// ================================================================
	// 7. fmt.Printf 占位符详解
	// ================================================================
	fmt.Println("\n========== 7. fmt.Printf 占位符详解 ==========")

	// ------------------------------
	// 7.1 通用占位符（适用所有类型）
	// ------------------------------
	fmt.Println("--- %v / %+v / %#v / %T ---")

	type Person struct {
		Name string
		Age  int
	}
	p := Person{Name: "张三", Age: 25}

	fmt.Printf("%%v  默认格式: %v\n", p)    // {张三 25}
	fmt.Printf("%%+v 带字段名: %+v\n", p)   // {Name:张三 Age:25}
	fmt.Printf("%%#v Go语法格式: %#v\n", p) // main.Person{Name:"张三", Age:25}
	fmt.Printf("%%T  类型: %T\n", p)      // main.Person

	// %v 对不同类型的表现
	fmt.Printf("int:    %%v = %v\n", 42)                     // 42
	fmt.Printf("float:  %%v = %v\n", 3.14)                   // 3.14
	fmt.Printf("bool:   %%v = %v\n", true)                   // true
	fmt.Printf("string: %%v = %v\n", "hello")                // hello
	fmt.Printf("slice:  %%v = %v\n", []int{1, 2})            // [1 2]
	fmt.Printf("map:    %%v = %v\n", map[string]int{"a": 1}) // map[a:1]

	// ------------------------------
	// 7.2 整数占位符
	// ------------------------------
	fmt.Println("\n--- %d / %b / %o / %x / %X / %c ---")

	n := 65
	fmt.Printf("%%d  十进制: %d\n", n)      // 65
	fmt.Printf("%%b  二进制: %b\n", n)      // 1000001
	fmt.Printf("%%o  八进制: %o\n", n)      // 101
	fmt.Printf("%%x  十六进制(小写): %x\n", n) // 41
	fmt.Printf("%%X  十六进制(大写): %X\n", n) // 41
	fmt.Printf("%%c  对应字符: %c\n", n)     // A（ASCII 65）

	// 不同进制的对比
	fmt.Println()
	for _, num := range []int{0, 10, 255, 65535} {
		fmt.Printf("十进制=%6d → 二进制=%16b  八进制=%4o  十六进制=%4x\n", num, num, num, num)
	}

	// ------------------------------
	// 7.3 浮点数占位符
	// ------------------------------
	fmt.Println("\n--- %f / %e / %E / %g / %G ---")

	f := 123.456789
	fmt.Printf("%%f  定点小数: %f\n", f)               // 123.456789
	fmt.Printf("%%.2f 保留2位: %.2f\n", f)            // 123.46（四舍五入）
	fmt.Printf("%%e  科学计数(小写): %e\n", f)           // 1.234568e+02
	fmt.Printf("%%E  科学计数(大写): %E\n", f)           // 1.234568E+02
	fmt.Printf("%%g  自动选择(默认6位有效数字): %g\n", f)     // 123.457
	fmt.Printf("%%.10g 自动选择(10位有效数字): %.10g\n", f) // 123.456789

	// ------------------------------
	// 7.4 字符串占位符
	// ------------------------------
	fmt.Println("\n--- %s / %q / %x / %X ---")

	str := "Hello\t世界\n"
	fmt.Printf("%%s  字符串: [%s]\n", str) // [Hello 世界
	//                                            ]  （\t和\n被实际执行）
	fmt.Printf("%%q  带引号+转义: %q\n", str)  // "Hello\t世界\n"（能看到\t\n字符本身）
	fmt.Printf("%%x  十六进制字节: %x\n", str)  // 48656c6c6f09e4b896e7958c0a
	fmt.Printf("%% X  带空格分隔: % X\n", str) // 48 65 6C 6C 6F 09 E4 B8 96 E7 95 8C 0A

	// ------------------------------
	// 7.5 指针占位符 %p
	// ------------------------------
	fmt.Println("\n--- %p ---")

	x := 42
	ptr := &x
	fmt.Printf("变量值: %d\n", x)         // 42
	fmt.Printf("变量地址: %p\n", &x)       // 0xc000012340（具体值每次运行不同）
	fmt.Printf("指针值(存的地址): %p\n", ptr) // 同上，指针里存的就是 &x
	fmt.Printf("解引用: %d\n", *ptr)      // 42

	// 切片底层数组地址
	slice := []int{1, 2, 3}
	fmt.Printf("切片底层数组首地址: %p\n", &slice[0]) // 0xc000xxx

	// ------------------------------
	// 7.6 布尔占位符 %t
	// ------------------------------
	fmt.Println("\n--- %t ---")

	fmt.Printf("%%t true: %t\n", true)         // true
	fmt.Printf("%%t false: %t\n", false)       // false
	fmt.Printf("%%t 表达式: 3 > 2 = %t\n", 3 > 2) // true

	// ------------------------------
	// 7.7 宽度和精度控制
	// ------------------------------
	fmt.Println("\n--- 宽度（对齐）和精度（小数位数）---")

	// %5d  表示最少占5个字符宽度，右对齐，不足补空格
	fmt.Printf("[%5d]\n", 42)    // [   42]
	fmt.Printf("[%5d]\n", 12345) // [12345]
	fmt.Printf("[%05d]\n", 42)   // [00042]（用0补齐）
	fmt.Printf("[%-5d]\n", 42)   // [42   ]（左对齐，加负号）

	// %.2f  表示保留2位小数
	fmt.Printf("%.0f\n", 3.14) // 3
	fmt.Printf("%.1f\n", 3.14) // 3.1
	fmt.Printf("%.2f\n", 3.14) // 3.14
	fmt.Printf("%.6f\n", 3.14) // 3.140000

	// %10.2f  总宽度10，小数2位
	fmt.Printf("[%10.2f]\n", 3.14) // [      3.14]

	// 字符串截断
	fmt.Printf("[%.3s]\n", "Hello") // [Hel]
	fmt.Printf("[%10s]\n", "Hi")    // [        Hi]
	fmt.Printf("[%-10s]\n", "Hi")   // [Hi        ]

	// ------------------------------
	// 7.8 特殊占位符 %%
	// ------------------------------
	fmt.Println("\n--- %% (打印百分号本身) ---")

	fmt.Printf("完成率: 85%%\n") // 完成率: 85%
	fmt.Printf("折扣: 9.5折\n")

	// ------------------------------
	// 7.9 占位符速查表
	// ------------------------------
	fmt.Println("\n--- 占位符速查表 ---")
	fmt.Println("  %v   默认格式（自动推断类型）")
	fmt.Println("  %+v  结构体带字段名")
	fmt.Println("  %#v  Go语法格式（可直接复制用）")
	fmt.Println("  %T   类型名")
	fmt.Println("  %d   十进制整数")
	fmt.Println("  %b   二进制")
	fmt.Println("  %o   八进制")
	fmt.Println("  %x   十六进制（小写）")
	fmt.Println("  %X   十六进制（大写）")
	fmt.Println("  %c   字符（ASCII/Unicode码位）")
	fmt.Println("  %q   带引号的字符串（转义可见）")
	fmt.Println("  %s   字符串内容")
	fmt.Println("  %f   浮点数")
	fmt.Println("  %e   科学计数法")
	fmt.Println("  %t   布尔值")
	fmt.Println("  %p   指针地址")
	fmt.Println("  %%   百分号本身")
	fmt.Println()
	fmt.Println("  宽度: %5d  右对齐占5位")
	fmt.Println("        %-5d 左对齐占5位")
	fmt.Println("        %05d 用0补齐5位")
	fmt.Println("  精度: %.2f  保留2位小数")
	fmt.Println("        %.3s  字符串截断3字符")
	fmt.Println("        %10.2f 总宽10，小数2位")
}
