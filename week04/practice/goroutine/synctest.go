package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================
// 一、sync.Mutex —— 互斥锁
// ============================================================
// 作用：同一时刻只允许一个 goroutine 访问被保护的共享资源
// 类比：卫生间的门锁，进去一个人后别人只能排队
//
// 核心方法：
//   Lock()   —— 加锁，如果锁已被占用则阻塞等待
//   Unlock() —— 解锁，释放锁让其他 goroutine 竞争
//
// 注意：
//   - 加锁和解锁必须成对出现，通常用 defer unlock() 保证不会遗漏
//   - 同一个 goroutine 不能对一个普通的 Mutex 重复加锁（会死锁）
//   - 未加锁就 Unlock() 会 panic

func mutexExample() {
	fmt.Println("========== sync.Mutex ==========")

	// 共享变量，多个 goroutine 会同时修改它
	var count int
	var mu sync.Mutex

	// WaitGroup 用于等待所有 goroutine 完成（后面会详细讲）
	var wg sync.WaitGroup

	// 启动 5 个 goroutine，每个都执行 100 次 count++
	for i := 0; i < 5; i++ {
		wg.Add(1) // 登记一个待完成的 goroutine
		go func(id int) {
			defer wg.Done() // goroutine 结束时会调用，表示"我做完了"
			for j := 0; j < 100; j++ {
				mu.Lock()   // 加锁：进入临界区
				count++     // 临界区：被保护的操作
				mu.Unlock() // 解锁：离开临界区
			}
			fmt.Printf("  goroutine %d 完成\n", id)
		}(i)
	}

	wg.Wait() // 阻塞等待，直到所有 wg.Done() 都被调用
	fmt.Printf("  count = %d（预期 500，有锁保护所以一定是 500）\n", count)
}

// ============================================================
// 二、sync.RWMutex —— 读写互斥锁
// ============================================================
// 作用：区分"读"和"写"两种场景，提高并发性能
//
// 规则：
//   - 多个 goroutine 可以 同时 读（RLock 之间不互斥）
//   - 写操作（Lock）与任何读/写操作都互斥
//   - 写操作有优先权，防止"写饥饿"
//
// 适用场景：读多写少的情况（比如缓存、配置）

func rwMutexExample() {
	fmt.Println("========== sync.RWMutex ==========")

	var mu sync.RWMutex
	var data = "初始数据"
	var wg sync.WaitGroup

	// --- 写 goroutine（只有一个，需要独占锁）---
	wg.Add(1)
	go func() {
		defer wg.Done()
		mu.Lock() // 写锁：独占，阻塞所有读和写
		fmt.Println("  [写] 获取写锁，正在修改数据...")
		time.Sleep(100 * time.Millisecond) // 模拟写操作耗时
		data = "新数据"
		mu.Unlock()
		fmt.Println("  [写] 释放写锁")
	}()

	// --- 读 goroutine（多个，可以并发读）---
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			time.Sleep(50 * time.Millisecond) // 等写 goroutine 先启动
			mu.RLock()                        // 读锁：多个读可以同时持有
			fmt.Printf("  [读%d] 读到: %s\n", id, data)
			time.Sleep(50 * time.Millisecond)
			mu.RUnlock()
		}(i)
	}

	wg.Wait()
}

// ============================================================
// 三、sync.WaitGroup —— 等待一组 goroutine 完成
// ============================================================
// 作用：主 goroutine 等待其他 goroutine 全部结束后再继续
// 类比：老师等全班同学都交卷了才收卷
//
// 核心方法：
//   Add(n)  —— 增加 n 个待完成的计数
//   Done()  —— 减少 1 个计数（等价于 Add(-1)）
//   Wait()  —— 阻塞等待，直到计数变为 0
//
// 注意：
//   - Add 必须在 goroutine 启动前调用（在主 goroutine 中调用）
//   - Done 必须在 goroutine 内部调用，通常 defer Done()
//   - Wait 在需要等待的地方调用

func waitGroupExample() {
	fmt.Println("========== sync.WaitGroup ==========")

	var wg sync.WaitGroup

	urls := []string{"https://a.com", "https://b.com", "https://c.com"}

	for _, url := range urls {
		wg.Add(1) // 每启动一个 goroutine 就 Add(1)
		go func(u string) {
			defer wg.Done() // 保证 goroutine 结束时一定会 Done
			// 模拟网络请求
			time.Sleep(200 * time.Millisecond)
			fmt.Printf("  已访问: %s\n", u)
		}(url)
	}

	fmt.Println("  等待所有请求完成...")
	wg.Wait() // 阻塞在这里，直到所有 Done 都被调用
	fmt.Println("  全部完成！")
}

// ============================================================
// 四、sync.Once —— 确保某段代码只执行一次
// ============================================================
// 作用：无论多少个 goroutine 同时调用，函数只执行一次
// 典型场景：单例模式、全局配置初始化、数据库连接池创建
//
// 核心方法：
//   Do(func()) —— 第一次调用时执行传入的函数，之后再调用什么都不做
//
// 原理：内部用 Mutex + done 标志实现，保证线程安全

func onceExample() {
	fmt.Println("========== sync.Once ==========")

	var once sync.Once
	var config string

	// 模拟初始化函数（只应该执行一次）
	initConfig := func() {
		fmt.Println("  初始化配置（只打印一次）...")
		time.Sleep(100 * time.Millisecond)
		config = "生产环境配置"
	}

	var wg sync.WaitGroup

	// 10 个 goroutine 同时尝试初始化
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("  goroutine %d 尝试初始化\n", id)
			once.Do(initConfig) // 只有第一个到达的 goroutine 会执行 initConfig
			fmt.Printf("  goroutine %d 获取到配置: %s\n", id, config)
		}(i)
	}

	wg.Wait()
}

// ============================================================
// 五、sync.Map —— 并发安全的 map
// ============================================================
// 作用：线程安全的 map，不需要额外加锁
//
// 适用场景：
//   - 大量并发读，少量写
//   - 不同 goroutine 操作不同的 key（key 不重叠）
// 注意：普通场景推荐用 `map + Mutex`，性能更好
//
// 核心方法：
//   Store(key, value)  —— 写入
//   Load(key)          —— 读取，返回 (value, ok)
//   LoadOrStore(key, value) —— 如果 key 存在就返回已有值，否则存入新值
//   Delete(key)        —— 删除
//   Range(func(key, value)) —— 遍历

func mapExample() {
	fmt.Println("========== sync.Map ==========")

	var m sync.Map

	// Store: 写入
	m.Store("name", "张三")
	m.Store("age", 25)

	// Load: 读取
	if v, ok := m.Load("name"); ok {
		fmt.Printf("  name = %v\n", v)
	}

	// LoadOrStore: 存在就返回已有值，不存在就写入
	v, loaded := m.LoadOrStore("age", 30) // "age" 已存在，返回 25
	fmt.Printf("  age = %v, 是否新写入: %v\n", v, loaded)

	v, loaded = m.LoadOrStore("city", "北京") // "city" 不存在，写入 "北京"
	fmt.Printf("  city = %v, 是否新写入: %v\n", v, loaded)

	// Range: 遍历（注意：遍历时可能不反映最新状态）
	fmt.Println("  遍历所有键值对:")
	m.Range(func(key, value any) bool {
		fmt.Printf("    %v = %v\n", key, value)
		return true // 返回 true 继续遍历，返回 false 停止
	})

	// Delete: 删除
	m.Delete("name")
	if _, ok := m.Load("name"); !ok {
		fmt.Println("  name 已被删除")
	}
}

// ============================================================
// 六、sync.Pool —— 对象池
// ============================================================
// 作用：缓存和复用对象，减少 GC（垃圾回收）压力
// 典型场景：频繁创建/销毁临时对象，比如 fmt 的内部 buffer
//
// 注意：
//   - Pool 中的对象随时可能被 GC 清除，不能依赖它存储持久数据
//   - 它是"性能优化"工具，不是"数据正确性"工具
//
// 核心方法：
//   New: 当池中没有可用对象时，用这个函数创建一个
//   Get() —— 从池中获取一个对象（如果没有就调用 New）
//   Put() —— 把对象放回池中（用完之后归还）

func poolExample() {
	fmt.Println("========== sync.Pool ==========")

	// 创建一个对象池
	var pool = sync.Pool{
		New: func() any {
			fmt.Println("    [Pool] 创建新对象")
			return make([]byte, 1024) // 每个对象是一个 1KB 的 buffer
		},
	}

	// 第一次获取：池是空的，会调用 New 创建
	buf1 := pool.Get().([]byte)
	fmt.Printf("  获取 buf1，长度=%d\n", len(buf1))

	// 用完之后归还
	pool.Put(buf1)
	fmt.Println("  归还 buf1 到池中")

	// 第二次获取：池中有刚才归还的 buf1，直接复用，不调用 New
	buf2 := pool.Get().([]byte)
	fmt.Printf("  获取 buf2，长度=%d（复用了 buf1，没有创建新对象）\n", len(buf2))

	// 如果不归还，下次 Get 就会创建新的
	pool.Put(buf2)
	buf3 := pool.Get().([]byte)
	buf4 := pool.Get().([]byte) // 池中只有一个，第二个会调用 New
	_ = buf3
	_ = buf4
}

// ============================================================
// 七、sync.Cond —— 条件变量
// ============================================================
// 作用：让一组 goroutine 等待某个条件满足后再继续执行
// 类比：游乐园的排队区，工作人员喊"可以进去了"大家才一起走
//
// 核心方法：
//   Wait()   —— 释放锁并进入等待，直到被唤醒
//   Signal() —— 唤醒一个等待的 goroutine
//   Broadcast() —— 唤醒所有等待的 goroutine
//
// 注意：
//   - Wait() 必须在持有锁的情况下调用
//   - 通常配合某个条件变量使用，用 for 循环检查条件（防止虚假唤醒）

func condExample() {
	fmt.Println("========== sync.Cond ==========")

	var mu sync.Mutex
	// 条件变量需要绑定到一个 Locker（Mutex 或 RWMutex）上
	cond := sync.NewCond(&mu)
	var ready bool // 条件标志

	// 启动 3 个等待的 goroutine
	for i := 0; i < 3; i++ {
		go func(id int) {
			mu.Lock() // Wait 需要持有锁
			fmt.Printf("  goroutine %d 开始等待...\n", id)
			for !ready { // 必须用 for 循环检查，不能用 if
				cond.Wait() // 释放锁并阻塞，被唤醒后重新获取锁
			}
			fmt.Printf("  goroutine %d 收到信号，继续执行！\n", id)
			mu.Unlock()
		}(i)
	}

	// 主 goroutine 等一会儿后发出信号
	time.Sleep(500 * time.Millisecond)
	mu.Lock()
	ready = true // 先修改条件
	mu.Unlock()
	fmt.Println("  条件已满足，广播信号...")
	cond.Broadcast() // 唤醒所有等待的 goroutine

	time.Sleep(200 * time.Millisecond) // 等所有 goroutine 打印完
}

// ============================================================
// 八、sync/atomic —— 原子操作
// ============================================================
// 作用：对单个变量进行无锁的原子读写
// 适用场景：简单的计数器、标志位等，比 Mutex 更高效
//
// 常用操作：
//   AddInt64 / AddUint64 —— 原子加法
//   LoadInt64 / StoreInt64 —— 原子读/写
//   CompareAndSwapInt64 —— CAS 操作（比较并交换）

func atomicExample() {
	fmt.Println("========== sync/atomic ==========")

	var counter int64 // 注意：原子操作要求变量地址对齐，一般用 int64
	var wg sync.WaitGroup

	// 5 个 goroutine 各加 100 次
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				atomic.AddInt64(&counter, 1) // 原子加 1，不需要加锁
			}
		}()
	}

	wg.Wait()
	fmt.Printf("  counter = %d（预期 500，原子操作保证结果正确）\n", counter)

	// Load 和 Store 示例
	var flag int64
	atomic.StoreInt64(&flag, 1)                          // 原子写入
	fmt.Printf("  flag = %d\n", atomic.LoadInt64(&flag)) // 原子读取
}

// ============================================================
// 九、Mutex 死锁演示
// ============================================================
// 死锁：多个 goroutine 互相等待对方释放锁，导致全部卡死
// Go 运行时能检测死锁并报错：fatal error: all goroutines are asleep - deadlock!
//
// 常见死锁场景：
//   1. 同一 goroutine 对 Mutex 重复 Lock
//   2. 两个 goroutine 以相反顺序获取两把锁
//   3. channel 操作导致互相等待
//
// 这里不实际运行死锁代码，避免程序卡死，仅做注释说明
// 取消下面函数的注释并运行可以观察死锁错误

func deadlockExample() {
	fmt.Println("========== 死锁演示 ==========")

	// --- 场景1：同一 goroutine 重复加锁（必死锁）---
	/*
		var mu sync.Mutex
		mu.Lock()
		mu.Lock() // 卡死！同一个 Mutex 不能被同一个 goroutine 重复加锁
		mu.Unlock()
		mu.Unlock()
	*/

	// --- 场景2：两把锁，交叉获取（必死锁）---
	/*
		var mu1, mu2 sync.Mutex

		// goroutine A: 先锁 mu1，再锁 mu2
		go func() {
			mu1.Lock()
			time.Sleep(100 * time.Millisecond) // 让 B 有机会先锁 mu2
			mu2.Lock() // 等待 B 释放 mu2 —— 但 B 在等 A 释放 mu1
			fmt.Println("A 完成")
			mu2.Unlock()
			mu1.Unlock()
		}()

		// goroutine B: 先锁 mu2，再锁 mu1
		go func() {
			mu2.Lock()
			time.Sleep(100 * time.Millisecond) // 让 A 有机会先锁 mu1
			mu1.Lock() // 等待 A 释放 mu1 —— 但 A 在等 B 释放 mu2
			fmt.Println("B 完成")
			mu1.Unlock()
			mu2.Unlock()
		}()

		time.Sleep(time.Second) // 等待死锁发生
	*/

	fmt.Println("  死锁代码已注释，取消注释可查看效果")
}

// ============================================================
// 主函数：依次运行所有示例
// ============================================================
func main() {
	waitGroupExample()
	fmt.Println()

	mutexExample()
	fmt.Println()

	rwMutexExample()
	fmt.Println()

	onceExample()
	fmt.Println()

	atomicExample()
	fmt.Println()

	mapExample()
	fmt.Println()

	poolExample()
	fmt.Println()

	condExample()
	fmt.Println()

	deadlockExample()
}
