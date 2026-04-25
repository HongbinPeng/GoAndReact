# Week05 练习目录

## 个人信息
- 姓名：彭鸿斌
- 学号：2024124379
- 学校：华中师范大学
这个目录用于存放我在学习 Go 标准库和并发编程时写下的练习代码。

本周为了完成“服务健康探测器”作业，我额外整理了一批“一个标准库一个目录”的学习示例，方便后续逐个运行、逐个理解。

## 标准库学习目录

- `encoding_json_01/`：学习 `encoding/json`
- `flag_01/`：学习 `flag`
- `os_01/`：学习 `os`
- `time_01/`：学习 `time`
- `fmt_01/`：学习 `fmt`
- `net_http_01/`：学习 `net/http`
- `net_01/`：学习 `net`
- `sync_01/`：学习 `sync`
- `io_01/`：学习 `io`
- `strings_01/`：学习 `strings`
- `context_01/`：学习 `context`
- `sort_01/`：学习 `sort`
- `text_tabwriter_01/`：学习 `text/tabwriter`
- `errors_01/`：学习 `errors`
- `path_filepath_01/`：学习 `path/filepath`
- `testing_01/`：学习 `testing`
- `net_http_httptest_01/`：学习 `net/http/httptest`

## 已有练习目录

- `arrayAndSlice/`
- `goroutine_01/`
- `library_01/`
- `map_02/`
- `slice_02/`
- `struct_01/`
## 可以深挖的内容
- `errors_01/`：学习 `errors`，这个可以深挖其底层是如何实现的，这里把指针方法和值方法的区别体现的很好，懂了这个再看errors.Is的实现会很有收获
errors.Is的实现方式是：
1. 先判断 err == target,很明显，这里的比较是先比较接口类型，再比较动态值，由于一般是指针方法集中实现了Error()方法，比较完接口类型后，再比较的是指针的指向，即指针值
2. 如果当前错误实现了自定义 Is(error) bool 方法，就调用它
3. 如果当前错误实现了 Unwrap() error，就继续向下找
4. 如果当前错误实现了 Unwrap() []error，就遍历多个子错误继续找
- `net_01/`：学习 `learnTcp`,`learnUdp`,这里学完标准包的常见功能后我实现了一个基于TCP的聊天功能,具体见 `week05\practice\net_01\QQByTCP`

