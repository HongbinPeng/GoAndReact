# Eino Callback Guide

## 1. 回调是什么

可以把 Eino 的回调理解成：

“给 Prompt / Model / Tool / Lambda / Graph / Chain / Workflow 这些组件加生命周期监听器”

它不是业务逻辑本身，而是一层观测能力。最常见的用途有：

- 打印节点执行日志
- 统计每个节点耗时
- 统一记录错误
- 观察 Graph / Chain 的真实执行顺序
- 做 tracing、埋点、监控

一句话说，回调是为了“看见这条链路是怎么跑的”。

## 2. 最常用的 3 个时机

### `OnStart`

组件真正开始执行前触发。

适合做：

- 记录“哪个节点开始了”
- 记录开始时间
- 在 `context` 里放一些轻量状态

### `OnEnd`

组件成功执行后触发。

适合做：

- 记录“哪个节点结束了”
- 统计耗时
- 读取输出做日志或指标

### `OnError`

组件执行失败时触发。

适合做：

- 打印错误日志
- 告警上报
- trace 标记

注意：`OnEnd` 和 `OnError` 是互斥的。成功走 `OnEnd`，失败走 `OnError`。

## 3. `RunInfo` 是什么

回调函数里会收到：

```go
info *callbacks.RunInfo
```

它表示“是谁触发了这次回调”。

最重要的字段有：

- `info.Name`
  业务层给节点起的名字。通常来自 `compose.WithNodeName(...)`。
- `info.Type`
  组件实现类型，比如某个 ChatModel 可能是 `OpenAI`。
- `info.Component`
  组件类别，比如 `Lambda`、`Prompt`、`ChatModel`、`Tool`、`Chain`。

通常调试时最常看的是 `Name` 和 `Component`。

## 4. `CallbackInput` / `CallbackOutput` 怎么理解

回调里还会拿到：

```go
input callbacks.CallbackInput
output callbacks.CallbackOutput
```

它们表示组件级别的输入输出。

例如：

- 对 `ChatTemplate` 来说，输入可能是 `map[string]any`，输出可能是 `[]*schema.Message`
- 对 `ChatModel` 来说，输入可能是 `[]*schema.Message`，输出可能是 `*schema.Message`
- 对 `Lambda` 来说，输入输出就是你自己定义的类型

因为 callbacks 包要统一适配不同组件，所以这里用的是通用类型。

如果你想精确读取某类组件的数据，通常要用对应组件包提供的 `ConvCallbackInput` / `ConvCallbackOutput` 去安全断言。

## 5. `HandlerBuilder` 怎么用

最常见的写法就是你 demo 里的这种：

```go
handler := callbacks.NewHandlerBuilder().
    OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
        log.Printf("[开始] 组件=%s 名称=%s 类型=%s", info.Component, info.Name, info.Type)
        return context.WithValue(ctx, "start_time_"+info.Name, time.Now())
    }).
    OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
        if start, ok := ctx.Value("start_time_" + info.Name).(time.Time); ok {
            log.Printf("[结束] 组件=%s 名称=%s 耗时=%v", info.Component, info.Name, time.Since(start))
        }
        return ctx
    }).
    OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
        log.Printf("[错误] 组件=%s 名称=%s 错误=%v", info.Component, info.Name, err)
        return ctx
    }).
    Build()
```

它的含义是：

- `OnStartFn`：注册开始时做什么
- `OnEndFn`：注册成功结束时做什么
- `OnErrorFn`：注册失败时做什么
- `Build()`：构造出真正的 `callbacks.Handler`

## 6. `context` 在回调里是干什么的

回调体系里，`context` 最常见的作用有两个：

1. 把回调处理器从外层传到各个组件
2. 在同一个 handler 的 `OnStart -> OnEnd / OnError` 之间传递临时状态

你 demo 里的这段就是典型例子：

```go
return context.WithValue(ctx, "start_time_"+info.Name, time.Now())
```

然后在 `OnEnd` 再取出来：

```go
start, ok := ctx.Value("start_time_" + info.Name).(time.Time)
```

这就实现了“开始时记时间，结束时算耗时”。

注意几点：

- 这种状态传递适合同一个 handler 内部使用
- 不要依赖不同 handler 之间的顺序
- 不要把大对象塞进 `context`

## 7. `WithCallbacks(handler)` 怎么生效

真正把回调挂到一次执行上的，是这句：

```go
result, err := runnable.Invoke(ctx, input, compose.WithCallbacks(handler))
```

意思是：

“这一次 Invoke，请带上这个 handler 一起执行”

所以回调不是写了就全局生效，而是按“这次调用”注入进去。

如果你有多个 handler，也可以一起传：

```go
compose.WithCallbacks(h1, h2, h3)
```

## 8. 你这个 demo 的执行顺序

`callback_demo.go` 里的这条 Chain 是：

```text
string
  ↓
Lambda(消息构建)
  ↓
ChatModel(通义千问)
  ↓
*schema.Message
```

所以日志大致会体现出：

1. Chain 开始
2. `消息构建` 开始
3. `消息构建` 结束
4. `通义千问` 开始
5. `通义千问` 结束
6. Chain 结束

如果中间报错，就会看到对应节点进入 `OnError`。

## 9. 什么时候适合用回调

很适合：

- 想知道节点有没有执行
- 想知道执行顺序
- 想统计耗时
- 想统一做模型日志
- 想挂 tracing / metrics

不太适合：

- 在回调里偷偷修改业务输入输出
- 把核心业务逻辑写进回调
- 依赖多个 handler 的顺序做复杂状态流转

## 10. 对应到这个 demo

你的 [callback_demo.go](E:/AllCode/VueCode/penghongbin/week06/practice/eino_quick_start/call_back/callback_demo.go) 里这段：

```go
handler := callbacks.NewHandlerBuilder().
    OnStartFn(...).
    OnEndFn(...).
    OnErrorFn(...).
    Build()
```

做的事情非常典型：

- `OnStart`：打印开始日志，记录开始时间
- `OnEnd`：打印结束日志，计算耗时
- `OnError`：打印错误日志

这就是 Eino 回调最常见、也最值得先掌握的第一种用法。
