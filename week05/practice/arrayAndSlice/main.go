// arrayAndSlice main包
// 本文件通过代码实例 + 详细注释，讲解 Go 语言中数组（Array）与切片（Slice）的底层区别
package main

import (
	"fmt"
	"reflect"
	"unsafe"
)

// =============================================================================
// 第一部分：数组（Array）的底层原理
// =============================================================================
//
// 【数组是什么？】
// Go 中的数组是【值类型】，它的长度是类型的一部分。
// [3]int 和 [5]int 是两种完全不同的类型，编译期就确定了。
//
// 【内存布局】
// 数组在内存中就是一块【连续的、固定大小的】空间。
// 例如 var arr [3]int 在栈上直接分配 3 * 8 = 24 字节（64位机器上int占8字节）。
//
//  内存地址（低 → 高）：
//  +-----------+-----------+-----------+
//  | arr[0]=10 | arr[1]=20 | arr[2]=30 |
//  +-----------+-----------+-----------+
//
// 【核心特性】
// 1. 长度固定，创建后无法扩容或缩容
// 2. 赋值或传参时会发生【完整拷贝】（所有元素逐个复制）
// 3. 数组名直接指向数据本身（不是指针）
// 4. 长度是类型的一部分：[3]int != [5]int
// =============================================================================

func demonstrateArray() {
	fmt.Println("========== 数组（Array）演示 ==========")

	// 1. 数组声明：长度是类型的一部分
	var arr1 [3]int = [3]int{1, 2, 3}
	arr2 := [3]int{1, 2, 3}

	// 数组类型包含长度，[3]int 和 [4]int 是不同类型
	// 下面这行如果取消注释会编译错误：cannot use arr1 (type [3]int) as type [4]int
	// var arr3 [4]int = arr1
	fmt.Printf("arr1 类型: %T\n", arr1) // 输出: [3]int
	fmt.Printf("arr2 类型: %T\n", arr2) // 输出: [3]int
	_ = arr2

	// 2. 数组赋值是【值拷贝】—— 修改副本不影响原数组
	arrCopy := arr1 // 这里发生了完整的内存拷贝，3个int全部复制
	arrCopy[0] = 999
	fmt.Printf("arr1[0] = %d (未被修改)\n", arr1[0])  // 仍然是 1
	fmt.Printf("arrCopy[0] = %d (已被修改)\n", arrCopy[0]) // 变成 999

	// 3. 数组传参也是【值拷贝】—— 函数内部修改不影响外部
	fmt.Println("\n--- 数组传参演示 ---")
	fmt.Printf("调用 modifyArray 之前: arr1 = %v\n", arr1)
	modifyArray(arr1) // 传的是副本，函数内部修改的是拷贝出来的数组
	fmt.Printf("调用 modifyArray 之后: arr1 = %v (未被修改，因为传参是值拷贝)\n", arr1)

	// 4. 数组的内存地址：数组变量本身就是数据
	fmt.Println("\n--- 数组内存地址 ---")
	fmt.Printf("arr1 的地址:    %p\n", &arr1)     // 数组本身的地址
	fmt.Printf("arr1[0] 的地址: %p\n", &arr1[0])  // 第一个元素的地址 == 数组地址
	fmt.Printf("arr1[1] 的地址: %p\n", &arr1[1])  // 第二个元素地址 = arr1[0]地址 + 8字节
	fmt.Printf("arr1[2] 的地址: %p\n", &arr1[2])  // 第三个元素地址 = arr1[1]地址 + 8字节
	// 可以看到，元素地址是连续的，间隔正好是 int 的大小（8字节）

	// 5. 数组大小（编译期确定）
	fmt.Printf("\narr1 占用的内存大小: %d 字节 (3个int × %d字节)\n",
		unsafe.Sizeof(arr1), unsafe.Sizeof(arr1[0]))
}

// modifyArray 接收数组值，函数内部操作的是副本
// 注意参数类型是 [3]int，不是 []int，也不是 *[3]int
func modifyArray(arr [3]int) {
	arr[0] = 888
	fmt.Printf("  modifyArray 内部: arr = %v\n", arr)
}

// =============================================================================
// 第二部分：切片（Slice）的底层原理
// =============================================================================
//
// 【切片是什么？】
// 切片是【引用类型】（更准确地说，它是一个描述符/头结构体），
// 它本身不存储数据，而是指向底层数组。
//
// 【切片的底层结构体】（源码位置：runtime/slice.go）
//
// 	type slice struct {
// 	    array unsafe.Pointer // 指向底层数组的指针
// 	    len   int             // 当前切片中元素的个数（长度）
// 	    cap   int             // 底层数组的容量（从切片起始位置到数组末尾）
// 	}
//
// 这个结构体在 64 位系统上占 24 字节（3个8字节字段）。
//
// 【内存布局示例】
//
//  切片变量 s（栈上，24字节）          底层数组（堆上或栈上）
//  +-------------------+              +----+----+----+----+----+
//  | array: 0x12345000 | -----------> | 10 | 20 | 30 | 40 | 50 |
//  | len:    3         |              +----+----+----+----+----+
//  | cap:    5         |
//  +-------------------+
//
// s[0] 指向 10, s[1] 指向 20, s[2] 指向 30
// s[3] 和 s[4] 虽然存在但不能访问（因为 len=3），可以通过 append 扩容
//
// 【核心特性】
// 1. 切片是【描述符】，包含指针+长度+容量三个字段
// 2. 切片赋值只拷贝描述符（24字节），不拷贝底层数组
// 3. 多个切片可以【共享同一个底层数组】
// 4. 切片传参时拷贝的是描述符，但指向的是同一份数据
// 5. 切片长度不够时会自动扩容（分配新数组、拷贝数据、更新指针）
// 6. 切片长度 len 可以在 0 到 cap 之间通过切片操作灵活调整
// =============================================================================

func demonstrateSlice() {
	fmt.Println("\n========== 切片（Slice）演示 ==========")

	// 1. 切片的创建方式
	// make([]T, len, cap) —— 这是最常用的方式
	// make 会在底层创建一个数组，然后返回指向该数组的切片描述符
	s1 := make([]int, 3, 5) // len=3, cap=5，底层数组有5个空间，但只能访问前3个
	fmt.Printf("s1: %v, len=%d, cap=%d\n", s1, len(s1), cap(s1))

	// 字面量创建 —— 此时 len == cap
	s2 := []int{10, 20, 30, 40, 50} // len=5, cap=5
	fmt.Printf("s2: %v, len=%d, cap=%d\n", s2, len(s2), cap(s2))

	// 从数组派生切片
	arr := [5]int{100, 200, 300, 400, 500}
	s3 := arr[1:4] // 从数组切出索引1到3（不含4）的元素
	fmt.Printf("s3: %v (从数组arr[1:4]切出), len=%d, cap=%d\n", s3, len(s3), cap(s3))
	// 注意：s3的cap=4而不是3！因为s3底层还是arr，从索引1到arr末尾还有4个元素

	// 2. 切片赋值是【描述符拷贝】，不是数据拷贝
	// 两个切片共享同一个底层数组
	fmt.Println("\n--- 切片赋值（共享底层数组）---")
	original := []int{1, 2, 3, 4, 5}
	copy := original // 只拷贝了24字节的描述符，array指针指向同一个底层数组

	fmt.Printf("original 描述符地址: %p\n", &original)
	fmt.Printf("copy     描述符地址: %p\n", &copy)
	// 描述符是两个不同的变量（地址不同）

	// 但底层数组是同一个
	fmt.Printf("original 底层数组地址: %p\n", &original[0])
	fmt.Printf("copy     底层数组地址: %p\n", &copy[0])
	// 两个切片的 array 指针相同 → 共享同一份数据

	// 修改 copy 会影响 original
	copy[0] = 999
	fmt.Printf("修改 copy[0]=999 后: original = %v\n", original) // original 也变了！

	// 3. 切片传参 —— 描述符拷贝，但数据共享
	fmt.Println("\n--- 切片传参演示 ---")
	data := []int{10, 20, 30}
	fmt.Printf("调用 modifySlice 之前: data = %v\n", data)
	modifySlice(data)
	fmt.Printf("调用 modifySlice 之后: data = %v (被修改了！因为共享底层数组)\n", data)

	// 4. 从数组切片：多个切片共享同一个数组
	fmt.Println("\n--- 多个切片共享底层数组 ---")
	baseArr := [6]int{0, 1, 2, 3, 4, 5}
	sA := baseArr[0:3] // [0,1,2], cap=6
	sB := baseArr[2:5] // [2,3,4], cap=4
	sC := baseArr[4:6] // [4,5],  cap=2

	fmt.Printf("sA: %v, cap=%d, 底层数组首地址: %p\n", sA, cap(sA), &sA[0])
	fmt.Printf("sB: %v, cap=%d, 底层数组首地址: %p\n", sB, cap(sB), &sB[0])
	fmt.Printf("sC: %v, cap=%d, 底层数组首地址: %p\n", sC, cap(sC), &sC[0])
	fmt.Printf("baseArr 首地址:         %p\n", &baseArr[0])
	// 可以看到 sA[0] 的地址 == baseArr[0] 的地址
	// sB[0] 的地址 == baseArr[2] 的地址
	// sC[0] 的地址 == baseArr[4] 的地址
	// 它们都指向同一个底层数组的不同位置！

	// 修改 sA 会影响 sB 中重叠的部分
	sA[2] = 222 // 修改了 baseArr[2]
	fmt.Printf("修改 sA[2]=222 后: sB = %v (sB[0] 也被影响了，因为 sB[0] == baseArr[2])\n", sB)

	// 5. 切片的扩容机制
	fmt.Println("\n--- 切片扩容（append）演示 ---")
	s := make([]int, 0, 2) // 初始: len=0, cap=2
	fmt.Printf("初始:     slice=%v, len=%d, cap=%d, 底层数组地址=%p\n",
		s, len(s), cap(s), getSliceAddr(s))

	s = append(s, 1, 2) // 刚好填满容量
	fmt.Printf("append后: slice=%v, len=%d, cap=%d, 底层数组地址=%p\n",
		s, len(s), cap(s), getSliceAddr(s))

	oldAddr := getSliceAddr(s)
	s = append(s, 3) // 容量不够了！触发扩容
	newAddr := getSliceAddr(s)
	fmt.Printf("再次append: slice=%v, len=%d, cap=%d, 底层数组地址=%p\n",
		s, len(s), cap(s), newAddr)

	if oldAddr != newAddr {
		fmt.Println("  → 底层数组地址变了！说明发生了扩容：分配了新数组，拷贝了旧数据")
	} else {
		fmt.Println("  → 底层数组地址没变")
	}
	// Go 的扩容策略（Go 1.18+）：
	//   cap < 256 时：扩容为原来的 2 倍
	//   cap >= 256 时：扩容为原来的约 1.25 倍
	// 这里 cap 从 2 变成 4（2倍），符合预期

	// 6. 切片不会自动缩容
	// 即使你 s = s[:1] 把长度缩小，底层数组仍然存在，cap 不变
	fmt.Println("\n--- 切片缩小区间（不释放底层数组）---")
	big := make([]int, 100) // len=100, cap=100
	small := big[:5]        // len=5, cap=100（注意cap还是100！）
	fmt.Printf("small: len=%d, cap=%d\n", len(small), cap(small))
	fmt.Println("small 仍然持有整个100个元素的底层数组，即使只用了前5个")

	// 7. nil 切片 vs 空切片
	fmt.Println("\n--- nil 切片 vs 空切片 ---")
	var nilSlice []int        // nil 切片：描述符全零，array=nil, len=0, cap=0
	emptySlice := []int{}     // 空切片：有底层数组（指向零值区域），len=0, cap=0
	emptyMake := make([]int, 0) // 也是空切片

	fmt.Printf("nilSlice:   ptr=%p, len=%d, cap=%d, isNil=%v\n",
		nilSlice, len(nilSlice), cap(nilSlice), nilSlice == nil)
	fmt.Printf("emptySlice: ptr=%p, len=%d, cap=%d, isNil=%v\n",
		emptySlice, len(emptySlice), cap(emptySlice), emptySlice == nil)
	fmt.Printf("emptyMake:  ptr=%p, len=%d, cap=%d, isNil=%v\n",
		emptyMake, len(emptyMake), cap(emptyMake), emptyMake == nil)
	// nil 切片的 array 指针是 nil
	// 空切片的 array 指针不为 nil（指向 runtime 的 zerobase）
}

// modifySlice 接收切片，虽然描述符是拷贝的，但指向同一底层数组
func modifySlice(s []int) {
	s[0] = 888
	fmt.Printf("  modifySlice 内部: s = %v\n", s)
}

// getSliceAddr 获取切片底层数组的首地址
// 通过 reflect.SliceHeader 来访问切片的内部字段
func getSliceAddr(s []int) unsafe.Pointer {
	// reflect.SliceHeader 是切片内部结构的公开表示
	// 它和 runtime 中的 slice 结构体布局一致
	// Data 字段是 uintptr 类型，需要转为 unsafe.Pointer
	return unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&s)).Data)
}

// =============================================================================
// 第三部分：数组 vs 切片 核心对比总结
// =============================================================================
//
// +-------------------+----------------------------------+----------------------------------+
// | 特性              | 数组 [N]T                        | 切片 []T                         |
// +-------------------+----------------------------------+----------------------------------+
// | 类型本质          | 值类型                           | 引用类型（描述符结构体）         |
// | 长度              | 编译期固定，是类型的一部分       | 运行时可变                       |
// | 内存布局          | 数据直接存在变量中               | 描述符(ptr+len+cap)+底层数组    |
// | 赋值/传参         | 完整拷贝所有元素（深拷贝）       | 只拷贝描述符24字节（浅拷贝）     |
// | 修改影响          | 副本修改不影响原数组             | 修改底层数组会影响所有共享的切片 |
// | 扩容              | 不支持                           | 支持，append 自动扩容            |
// | 长度 vs 容量      | 只有长度                         | 有 len（长度）和 cap（容量）     |
// | nil 值            | 不存在 nil 数组                  | 存在 nil 切片                    |
// | 比较              | 可以用 == 比较（长度+元素值）    | 只能用 reflect.DeepEqual 比较   |
// | 函数返回          | 返回数组会拷贝数据               | 返回切片只拷贝描述符             |
// | 性能              | 大数组传参拷贝开销大             | 传参开销小（仅24字节）           |
// | 使用场景          | 固定大小、需要值语义的场景       | 动态大小、绝大多数日常场景       |
// +-------------------+----------------------------------+----------------------------------+
//
// 【底层结构对比图】
//
// 数组 var arr [3]int（64位系统）：
//
//   arr 变量本身（24字节，直接在栈上）
//   +----+----+----+
//   | 10 | 20 | 30 |   ← 数据就是变量本身
//   +----+----+----+
//
// 切片 s := []int{10, 20, 30}（64位系统）：
//
//   s 变量本身（24字节，在栈上）        底层数组（在堆上或栈上）
//   +-------------------+              +----+----+----+
//   | array: 0x1000 ----|------------> | 10 | 20 | 30 |
//   | len:    3         |              +----+----+----+
//   | cap:    3         |
//   +-------------------+
//
// 【关键理解：切片是数组的"窗口"】
// 切片不拥有数据，它只是一个"窗口"，让你能看到和操作底层数组的一部分。
// 当你 s[0:2] 时，你只是调整了窗口的大小和位置，底层数组不变。
// 当你 append 超过 cap 时，系统会换一个更大的底层数组，窗口指向新位置。

func demonstrateComparison() {
	fmt.Println("\n========== 数组 vs 切片 对比演示 ==========")

	// 对比1：传参行为差异
	fmt.Println("--- 传参行为对比 ---")

	arr := [3]int{1, 2, 3}
	sli := []int{1, 2, 3}

	fmt.Printf("传参前: arr=%v, sli=%v\n", arr, sli)
	modifyArray(arr)  // 数组：值拷贝，原数组不受影响
	modifySlice(sli)  // 切片：共享数据，原切片受影响
	fmt.Printf("传参后: arr=%v (不变), sli=%v (被修改)\n", arr, sli)

	// 对比2：赋值行为差异
	fmt.Println("\n--- 赋值行为对比 ---")

	arr1 := [3]int{10, 20, 30}
	arr2 := arr1 // 数组赋值：完整拷贝数据
	arr2[0] = 99
	fmt.Printf("数组: arr1=%v, arr2=%v (arr1不受影响)\n", arr1, arr2)

	sl1 := []int{10, 20, 30}
	sl2 := sl1 // 切片赋值：只拷贝描述符，共享数据
	sl2[0] = 99
	fmt.Printf("切片: sl1=%v, sl2=%v (sl1也被影响了)\n", sl1, sl2)

	// 对比3：长度与容量
	fmt.Println("\n--- 长度与容量对比 ---")

	arr3 := [5]int{1, 2, 3, 4, 5}
	sl3 := make([]int, 3, 5)

	fmt.Printf("数组: len=%d (数组没有cap概念，len永远等于声明大小)\n", len(arr3))
	fmt.Printf("切片: len=%d, cap=%d (len可以小于cap)\n", len(sl3), cap(sl3))

	// 对比4：相等性比较
	fmt.Println("\n--- 相等性比较对比 ---")

	a1 := [3]int{1, 2, 3}
	a2 := [3]int{1, 2, 3}
	fmt.Printf("数组 a1 == a2: %v (数组支持 == 运算符)\n", a1 == a2)

	sx := []int{1, 2, 3}
	sy := []int{1, 2, 3}
	// 切片不支持 == 运算符，下面这行会编译错误：
	// fmt.Println(sx == sy)
	// 必须使用 reflect.DeepEqual
	fmt.Printf("切片 sx == sy: %v (使用 reflect.DeepEqual)\n", reflect.DeepEqual(sx, sy))

	// 对比5：从数组切片的"窗口"效果
	fmt.Println("\n--- 切片作为数组的窗口 ---")
	base := [5]int{10, 20, 30, 40, 50}
	window1 := base[0:3] // 窗口看到 [10,20,30], 但 cap=5（能看到底层数组剩余部分）
	window2 := base[2:5] // 窗口看到 [30,40,50], cap=3

	fmt.Printf("base:     %v\n", base)
	fmt.Printf("window1:  %v, cap=%d\n", window1, cap(window1))
	fmt.Printf("window2:  %v, cap=%d\n", window2, cap(window2))

	// 通过 window1 修改 base[2]
	window1[2] = 333
	fmt.Printf("修改 window1[2]=333 后: base=%v, window2=%v\n", base, window2)
	// window2[0] 也变了！因为 window1[2] 和 window2[0] 都指向 base[2]
}

// =============================================================================
// 第四部分：常见陷阱与注意事项
// =============================================================================

func demonstratePitfalls() {
	fmt.Println("\n========== 常见陷阱演示 ==========")

	// 陷阱1：range 遍历时修改切片元素
	// range 返回的是值的拷贝，不是引用
	fmt.Println("--- 陷阱1：range 遍历时值拷贝 ---")
	nums := []int{1, 2, 3}
	for i, v := range nums {
		v = v * 10 // 这里修改的是 v 的拷贝，不是 nums[i]
		_ = i
	}
	fmt.Printf("range 修改 v 后: nums = %v (未改变)\n", nums)

	// 正确做法：直接通过索引修改
	for i := range nums {
		nums[i] = nums[i] * 10
	}
	fmt.Printf("通过索引修改后: nums = %v (已改变)\n", nums)

	// 陷阱2：append 后旧切片仍然指向旧数组
	fmt.Println("\n--- 陷阱2：append 扩容后旧引用 ---")
	old := make([]int, 2, 2)
	old[0] = 1
	old[1] = 2
	oldAddr := getSliceAddr(old)

	// 复制一份描述符
	ref := old

	// append 触发扩容
	old = append(old, 3, 4, 5)
	newAddr := getSliceAddr(old)

	fmt.Printf("old 扩容前地址: %p, 扩容后地址: %p\n", oldAddr, newAddr)
	fmt.Printf("ref 仍然指向旧地址: %p\n", getSliceAddr(ref))
	fmt.Printf("old=%v, ref=%v (ref 不受扩容影响，仍然看旧数组)\n", old, ref)

	// 陷阱3：大数组切片导致内存泄漏
	// 当你从一个大数组或大切片中取一小段，切片仍然持有整个底层数组的引用
	// 这可能导致大量无用数据无法被 GC 回收
	fmt.Println("\n--- 陷阱3：切片持有大数组导致内存泄漏 ---")
	bigData := make([]byte, 1024*1024) // 1MB
	for i := range bigData {
		bigData[i] = byte(i % 256)
	}
	smallSlice := bigData[:10] // 只取10个字节，但底层1MB仍然被引用
	_ = smallSlice
	fmt.Println("smallSlice 只用了10字节，但底层1MB数组无法被GC回收")
	fmt.Println("解决方案：使用 copy 将需要的数据复制到新的切片中")
	safe := make([]byte, 10)
	copy(safe, bigData[:10])
	// 现在 bigData 可以被 GC 回收（如果没有其他引用）
	_ = safe

	// 陷阱4：nil 切片和空切片的行为差异
	fmt.Println("\n--- 陷阱4：nil 切片 vs 空切片的 JSON 序列化差异 ---")
	var nilSl []int
	emptySl := []int{}

	fmt.Printf("nil 切片:  %v\n", nilSl) // 输出: []
	fmt.Printf("空切片:    %v\n", emptySl) // 输出: []
	// fmt.Printf 看起来一样，但 JSON 序列化不同：
	// nil 切片 → null
	// 空切片 → []
	// 这在 API 响应中可能导致前端解析错误
}

// =============================================================================
// 第五部分：深入——切片扩容策略源码分析
// =============================================================================
//
// Go 1.18+ 的切片扩容逻辑（简化版，源自 runtime/slice.go）：
//
// 	growslice(oldCap, newLen) → newCap:
//
// 	  if newLen > 2 * oldCap:
// 	      newCap = newLen  // 新长度超过2倍旧容量，直接按需分配
//
// 	  threshold = 256
// 	  if oldCap < threshold:
// 	      newCap = oldCap * 2  // 小切片：翻倍
// 	  else:
// 	      // 大切片：每次增长约 25%
// 	      for newCap < newLen:
// 	          if oldCap < threshold:
// 	              newCap = oldCap * 2
// 	          else:
// 	              newCap += (newCap + 3*threshold) / 4  // ≈ 1.25x
//
// 为什么这样设计？
// - 小切片翻倍： amortize（摊销）分配和拷贝成本
// - 大切片 1.25x：避免一次性分配过多内存，减少浪费
//
// 扩容时会发生什么？
// 1. runtime.mallocgc 分配新内存（新底层数组）
// 2. memmove 将旧数组数据拷贝到新数组
// 3. 旧数组如果没有其他引用，会被 GC 回收
// 4. 切片描述符的 array 指针更新为指向新数组

func demonstrateGrowthStrategy() {
	fmt.Println("\n========== 切片扩容策略演示 ==========")

	// 观察小切片（cap < 256）的翻倍增长
	s := make([]int, 0, 1)
	fmt.Println("--- 小切片扩容（cap < 256，每次翻倍）---")
	fmt.Printf("初始: len=%d, cap=%d\n", len(s), cap(s))

	for i := 1; i <= 10; i++ {
		oldCap := cap(s)
		s = append(s, i)
		if cap(s) != oldCap {
			fmt.Printf("  append(%d) 触发扩容: cap %d → %d (%.1fx)\n",
				i, oldCap, cap(s), float64(cap(s))/float64(oldCap))
		}
	}

	// 观察大切片的增长
	fmt.Println("\n--- 大切片扩容（cap >= 256，约1.25倍）---")
	large := make([]int, 0, 256)
	for i := 1; i <= 300; i++ {
		oldCap := cap(large)
		large = append(large, i)
		if cap(large) != oldCap {
			fmt.Printf("  append(%d) 触发扩容: cap %d → %d (%.2fx)\n",
				i, oldCap, cap(large), float64(cap(large))/float64(oldCap))
		}
	}
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	// 第一部分：数组的底层原理
	demonstrateArray()

	// 第二部分：切片的底层原理
	demonstrateSlice()

	// 第三部分：数组 vs 切片对比
	demonstrateComparison()

	// 第四部分：常见陷阱
	demonstratePitfalls()

	// 第五部分：切片扩容策略
	demonstrateGrowthStrategy()
}