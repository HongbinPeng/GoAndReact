package main

// ============================================================
// context 标准库教程（初学者版）
// ============================================================
// 本教程分为 10 个章节，每个章节一个独立的 main.go 文件。
// 按顺序学习即可，后一个章节会复用前面的知识点。
//
// 章节目录：
//   01_what_is_context     — context 是什么？为什么需要它？
//   02_background_todo     — context.Background() 和 context.TODO()
//   03_with_cancel         — WithCancel：手动取消 goroutine
//   04_with_timeout        — WithTimeout：超时自动取消
//   05_with_deadline       — WithDeadline：在指定时刻取消
//   06_with_value          — WithValue：在 context 中传递数据
//   07_context_tree        — context 树和级联取消
//   08_http_and_context    — context 在 HTTP 请求中的实际应用
//   09_goroutine_and_context — context 管理多个 goroutine
//   10_best_practices      — 最佳实践和常见陷阱
//
// 前置知识：你需要了解 Go 的 goroutine、channel、defer
