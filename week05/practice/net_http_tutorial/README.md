package main

// ============================================================
// net/http 标准库教程（初学者版）
// ============================================================
// 本教程分为 12 个章节，每个章节一个独立的 main.go 文件。
// 按顺序学习即可，后一个章节会复用前面的知识点。
//
// 章节目录：
//   01_http_get          — 最简单的 HTTP 请求：http.Get
//   02_client_do         — 推荐写法：http.Client + client.Do
//   03_timeout           — 超时控制的两种方式
//   04_context_cancel    — 手动取消请求
//   05_post_request      — POST 请求：提交表单 / JSON 数据
//   06_request_headers   — 设置请求头（User-Agent, Authorization 等）
//   07_response_handling — 处理响应：状态码、响应头、流式读取
//   08_transport         — 自定义 Transport（连接池控制）
//   09_redirects         — 重定向处理
//   10_error_handling    — 错误分类与处理
//   11_httptest          — 用 httptest 写测试
//   12_http_server       — 最简单的 HTTP 服务端
//
// 前置知识：你需要了解 Go 的基础语法（函数、结构体、defer、go routine）
// 推荐先学完：time, context, io, errors 标准包
