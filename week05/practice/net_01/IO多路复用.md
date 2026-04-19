# select、poll、epoll —— 操作系统 I/O 多路复用详解

## 先回答一个问题：它们是什么？

一句话总结：**select、poll、epoll 是操作系统提供的三种 I/O 多路复用机制，用于让一个线程同时监控多个 socket 的状态。**

---

## 为什么需要 I/O 多路复用？

### 场景：一个服务器要同时服务 10000 个客户端

如果不用多路复用，有三种笨办法：

#### 笨办法 1：一个线程服务一个客户端

```
客户端 1 → 线程 1（阻塞等待数据）
客户端 2 → 线程 2（阻塞等待数据）
...
客户端 10000 → 线程 10000（阻塞等待数据）
```

- **问题**：创建 10000 个线程，每个线程占 1MB 栈 → 10GB 内存
- **问题**：线程切换开销巨大
- **结果**：服务器崩溃

#### 笨办法 2：一个线程轮流检查每个客户端（轮询）

```
while true {
    for conn in connections {
        data = conn.read()  // 如果没有数据，阻塞等待
    }
}
```

- **问题**：第一个客户端没数据就卡住了，后面的全被阻塞
- **结果**：延迟极高

#### 笨办法 3：非阻塞 + 轮询（忙等）

```
for conn in connections {
    conn.set_nonblocking()
}
while true {
    for conn in connections {
        data = conn.read()  // 没数据立即返回 EAGAIN
        if data == EAGAIN {
            continue  // 下一个
        }
    }
}
```

- **问题**：10000 个 socket 轮流问"有数据吗？没有。有数据吗？没有。..."
- **问题**：CPU 空转，100% 占用
- **结果**：CPU 被活活累死

---

## 正确答案：I/O 多路复用

```
while true {
    // 一次性告诉操作系统：帮我盯着这 10000 个 socket
    // 有任何一个"有数据可读"或"可以写了"，就叫醒我
    events = epoll_wait(epoll_fd, events, max_events, timeout)

    for event in events {
        conn = event.data.conn
        data = conn.read()  // 这里一定是有数据的，不会阻塞
        handle(data)
    }
}
```

**一个线程 + 一个系统调用 = 同时监控 10000 个连接。** 这就是 I/O 多路复用的本质。

---

## 三种机制的演进

### 第一代：select（1983 年，BSD Unix）

#### 工作原理

```
fd_set read_fds;
FD_ZERO(&read_fds);
FD_SET(socket1, &read_fds);
FD_SET(socket2, &read_fds);
// ... 把所有 socket 加入集合

select(max_fd + 1, &read_fds, NULL, NULL, &timeout);

// select 返回后，遍历所有 socket，看哪个在 read_fds 里
for (i = 0; i < max_fd; i++) {
    if FD_ISSET(i, &read_fds) {
        // 这个 socket 有数据了
    }
}
```

#### 三大缺陷

| 缺陷 | 说明 | 影响 |
|---|---|---|
| **fd 数量限制** | 用位图存储 fd，默认最多 1024 个（`FD_SETSIZE`） | 超过 1024 个连接直接不支持 |
| **每次调用都要全量拷贝** | 每次调用 select 都要把 1024 个 fd 从用户态拷贝到内核态 | 10000 个连接时，拷贝开销巨大 |
| **返回后需要遍历全部 fd** | select 只告诉你"有 fd 就绪了"，但不告诉你是哪个，你要遍历 1024 个一个个检查 | O(n) 时间复杂度，连接越多越慢 |

#### 类比

select 就像一个老师点名：
> "全班 1024 个同学，谁举手了？"
> 然后老师挨个看过去："张三？没举手。李四？没举手。... 王五？举手了！好，王五发言。"
> 每次点名都要从头开始挨个看。

---

### 第二代：poll（1997 年，SVR4）

#### 改进

poll 解决了 select 的第一个缺陷——**没有 1024 的数量限制**。

```c
struct pollfd fds[10000];
fds[0].fd = socket1;
fds[0].events = POLLIN;
fds[1].fd = socket2;
fds[1].events = POLLIN;
// ... 任意数量

poll(fds, 10000, timeout);

// 同样需要遍历所有 fd 检查
for (i = 0; i < 10000; i++) {
    if (fds[i].revents & POLLIN) {
        // 这个 socket 有数据了
    }
}
```

#### 但仍然存在的缺陷

| 缺陷 | 说明 |
|---|---|
| **仍然需要遍历全部 fd** | poll 返回后你还是不知道哪些 fd 就绪了，只能挨个检查 `revents` |
| **每次调用仍然要全量拷贝** | 10000 个 pollfd 结构体每次都要从用户态拷贝到内核态 |

#### 类比

poll 就像老师说：
> "全班同学（不限人数），谁举手了？"
> 然后还是挨个看过去："1号？没有。2号？没有。3号？没有。..."
> 虽然不限人数，但还是得一个个问。

---

### 第三代：epoll（2002 年，Linux 2.6）

#### 革命性的改进

epoll 解决了 select/poll 的所有缺陷，是真正的"高并发神器"。

```c
// 第一步：创建一个 epoll 实例（只做一次）
int epoll_fd = epoll_create(1);

// 第二步：把 socket 注册到 epoll 中（连接建立时做一次）
struct epoll_event event;
event.events = EPOLLIN;  // 监听可读事件
event.data.fd = socket;
epoll_ctl(epoll_fd, EPOLL_CTL_ADD, socket, &event);

// 第三步：等待事件（每次循环调用）
struct epoll_event events[1024];
int n = epoll_wait(epoll_fd, events, 1024, timeout);

// 第四步：只处理就绪的 fd（n 个，不是全部）
for (i = 0; i < n; i++) {
    int fd = events[i].data.fd;
    // 直接处理，不需要遍历全部 fd
    read(fd, buffer, sizeof(buffer));
}
```

#### epoll 的三大优势

| 优势 | 说明 | 对比 |
|---|---|---|
| **无 fd 数量限制** | 只受系统文件描述符上限限制（通常几十万） | select 限 1024 |
| **不需要全量拷贝** | fd 只在 `epoll_ctl` 时拷贝一次，`epoll_wait` 不需要再拷贝 | select/poll 每次调用都拷贝 |
| **只返回就绪的 fd** | `epoll_wait` 直接返回就绪的 fd 列表，不用遍历全部 | select/poll 返回后要 O(n) 遍历 |

#### 类比

epoll 就像老师配了一个智能系统：
> 系统自动记录谁举手了。老师不用挨个问，直接看系统显示的名单："3号、7号、15号举手了。"
> 老师直接叫这三个人发言，其他人不用管。

---

## epoll 的内部机制（简版）

epoll 之所以快，是因为它内部维护了一个**红黑树 + 就绪链表**：

```
epoll 实例（内核中）：
┌─────────────────────────────────────┐
│  红黑树（存储所有注册的 socket）      │
│  ┌───┐                              │
│  │ 5 │  ← socket fd 5              │
│  ├───┤                              │
│  │12 │  ← socket fd 12             │
│  ├───┤                              │
│  │23 │  ← socket fd 23             │
│  └───┘                              │
│                                     │
│  就绪链表（存储已就绪的 socket）      │
│  ┌──────┐  ┌──────┐  ┌──────      │
│  │ fd 5 │→│fd 23 │→│ NULL │      │
│  └──────┘  └──────┘  └──────┘      │
│         ↑                           │
│    中断回调函数触发时加入            │
└─────────────────────────────────────┘
```

工作流程：

```
1. epoll_ctl(ADD, fd)  → 把 fd 插入红黑树 + 注册中断回调
2. 网络数据到达        → 网卡触发硬件中断
3. 内核中断处理程序    → 调用回调函数，把 fd 加入就绪链表
4. epoll_wait()        → 直接从就绪链表取 fd 返回（O(1)！）
```

**关键**：epoll_wait 的复杂度是 O(1)（返回的就绪 fd 数量），不是 O(n)（总 fd 数量）。

---

## 三种机制对比总结

| 特性 | select | poll | epoll |
|---|---|---|---|
| **最大 fd 数** | 1024（编译时固定） | 无限制 | 无限制（受系统限制） |
| **fd 拷贝开销** | 每次调用全量拷贝 | 每次调用全量拷贝 | 只在注册时拷贝一次 |
| **就绪 fd 查找** | O(n) 遍历全部 | O(n) 遍历全部 | O(1) 直接取就绪链表 |
| **触发模式** | 水平触发（LT） | 水平触发（LT） | 水平触发（LT）+ 边缘触发（ET） |
| **适用场景** | 连接少且活跃 | 连接少且活跃 | 连接多、大部分不活跃 |
| **操作系统** | 跨平台（Unix/Windows） | 跨平台（Unix） | 仅 Linux |

---

## Go runtime 是怎么用的

Go 在不同操作系统上自动选择最适合的多路复用机制：

| 操作系统 | Go 使用的机制 | 实现文件 |
|---|---|---|
| Linux | **epoll** | `runtime/netpoll_epoll.go` |
| macOS / BSD | **kqueue** | `runtime/netpoll_kqueue.go` |
| Windows | **IOCP** | `runtime/netpoll_iocp.go` |
| Solaris | **event ports** | `runtime/netpoll_solaris.go` |

### Go 中的 epoll 使用流程

```
你写的代码:
    conn, err := net.DialTimeout("tcp", addr, timeout)

Go runtime 内部做的事:
    1. 创建 socket，设为非阻塞模式
    2. 发起 connect() → 返回 EINPROGRESS（正在连接）
    3. 把 socket fd 注册到 epoll（epoll_ctl ADD）
    4. 把当前 goroutine 挂起（放入等待队列）
    5. 其他 goroutine 继续执行
    6. TCP 握手完成 → 内核触发 epoll 事件
    7. netpoller 收到事件 → 唤醒挂起的 goroutine
    8. goroutine 恢复执行，conn 返回
```

你写的代码看起来是同步阻塞的，但 runtime 在底层用 **epoll + 非阻塞 socket + goroutine 调度** 把它变成了异步的。

---

## 水平触发（LT）vs 边缘触发（ET）

这是 epoll 特有的概念，理解它对深入理解网络编程很重要。

### 水平触发（Level-Triggered，默认模式）

```
socket 有数据可读 → epoll_wait 通知你
→ 你读了一部分，还剩一些数据没读
→ 下次调用 epoll_wait → 仍然通知你（因为"还有数据"这个条件持续成立）
→ 一直通知，直到你把所有数据读完
```

就像门铃：只要门开着（条件成立），门铃就一直响。

**优点**：不容易漏事件，即使你这次没处理完，下次还会提醒你。

### 边缘触发（Edge-Triggered，需要手动开启）

```
socket 有数据可读 → epoll_wait 通知你一次
→ 你读了一部分，还剩一些数据没读
→ 下次调用 epoll_wait → 不通知你了（因为"新数据到达"这个事件已经发生过了）
→ 直到有新数据再次到达，才会再通知
```

就像敲门：只在你敲门的那一瞬间通知你，敲完就不管了。

**优点**：减少重复通知，性能更高。

**缺点**：必须一次性把所有数据读完（循环 read 直到 EAGAIN），否则剩余数据会被漏掉。

**Go 用的是水平触发（LT）**，所以不用担心漏事件的问题。

---

## 总结：一句话理解三者的关系

| 机制 | 一句话 |
|---|---|
| **select** | 老前辈，有 1024 个 fd 的上限，每次全量拷贝 + 遍历 |
| **poll** | 选了前辈的改进版，去掉了 1024 限制，但仍然是全量拷贝 + 遍历 |
| **epoll** | 革命者，只注册一次、只返回就绪的、O(1) 效率，高并发的基石 |

**select → poll → epoll** 是一条性能不断优化的演进路线。现代高性能服务器（Nginx、Redis、Go runtime）全部使用 epoll（或等价机制）作为底层 I/O 模型。