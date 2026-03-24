此目录存放本周课后作业，可以在此文件添加作业设计思路和流程图等

# 灵犀 AI 对话助手

## 个人基本信息

- **姓名**：彭鸿斌
- **学校**：华中师范大学
- **学号**：2024124379
- **项目位置**：`week01/homework/lingxi/`

## 项目简介

灵犀 AI 是一款基于阿里云 Qwen 大模型的原生 JavaScript 聊天助手应用，支持流式响应、图片上传、深度思考模式、主题切换等功能，无需构建即可在浏览器中运行。

## 开发任务索引

### ✅ 已完成功能清单

#### 1. 核心对话功能

- 实时 AI 对话（阿里云 Qwen API）
- 流式响应渲染（打字机效果）
- 多轮对话上下文（本地存储最近 4 轮）
- 停止生成（AbortController 中断请求）

#### 2. 用户体验优化

- 亮色/暗色主题切换（localStorage 持久化）
- 欢迎界面与聊天界面切换
- 快捷建议卡片（点击快速提问）
- Shift+Enter 换行，Enter 发送
- 自动滚动到最新消息

#### 3. 图片处理功能

- 图片上传（文件选择器）
- 剪贴板粘贴（Ctrl+V）
- Base64 编码传输
- 缩略图预览（文件名、大小显示）
- 移除已上传图片
- 图片模态框全屏查看

#### 4. 高级特性

- 深度思考模式（`enable_thinking` 参数）
- 思考过程可视化展示
- Markdown 渲染（Prism.js 代码高亮）
- Token 使用量统计显示
- API 配置模态框（自定义 API Key）

#### 5. 数据持久化

- 主题设置 localStorage
- API Key localStorage
- 聊天历史 localStorage
- 图片预览状态管理

## 核心技术实现

### 1. 流式响应架构

**技术栈**：Fetch API + SSE (Server-Sent Events) + AbortController

```javascript
// 流式响应核心流程
async function sendToAI() {
    abortController = new AbortController();
    const response = await fetch(BASE_URL, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${API_KEY}` },
        body: JSON.stringify({
            model: MODEL_NAME,
            messages: buildMessages(),
            stream: true,
            stream_options: { include_usage: true }
        }),
        signal: abortController.signal
    });
    
    const reader = response.body.getReader();
    while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        processChunk({ done, value }); // 解析 SSE 数据块
    }
}
```

**亮点**：

- 使用 `AbortController` 实现可中断的流式请求
- SSE 格式解析：`data: {json}\n\n`
- 三种内容类型处理：`reasoning_content`（思考）、`content`（回复）、`usage`（Token 统计）

### 2. 对话上下文管理

**技术栈**：localStorage + 滑动窗口算法

```javascript
const MAX_HISTORY = 4; // 最多保存 4 轮对话

function addToChatHistory(role, content) {
    const history = getChatHistory();
    history.push({ role, content });
    // 滑动窗口：只保留最近 N 轮（每轮 2 条消息：用户+AI）
    const trimmed = history.slice(-MAX_HISTORY * 2);
    localStorage.setItem('lingxi_chat_history', JSON.stringify(trimmed));
}
```

**亮点**：

- 自动限制历史长度，避免请求过大
- 重启页面后自动恢复对话历史
- 终止对话时自动清理不完整记录

### 3. 图片多模态处理

**技术栈**：FileReader + Base64 + 缩略图预览

```javascript
// 图片转 Base64
function handleImageUpload(file) {
    const reader = new FileReader();
    reader.onload = (e) => {
        currentImageBase64 = e.target.result.split(',')[1]; // 移除 data:image/png;base64,前缀
        currentImageFile = file;
        showThumbnail(file.name, file.size); // 显示缩略图
    };
    reader.readAsDataURL(file);
}

// 构建多模态消息
function buildMessages() {
    const content = [];
    if (currentImageBase64) {
        content.push({
            type: "image_url",
            image_url: { url: `data:image/png;base64,${currentImageBase64}` }
        });
    }
    content.push({ type: "text", text: userInput });
    return content;
}
```

**亮点**：

- 支持文件上传和剪贴板粘贴双模式
- 缩略图实时预览（文件名、大小）
- 阅读阿里云百炼官方API帮助文档，选用了Base64 编码直接嵌入 API 请求的方式进行AI对话请求

### 4. 主题切换与持久化

**技术栈**：CSS 类切换 + localStorage + 动态样式表

```javascript
function toggleTheme() {
    currentTheme = currentTheme === 'light' ? 'dark' : 'light';
    document.body.className = currentTheme;
    localStorage.setItem('theme', currentTheme);
    
    // 动态切换 Prism.js 代码高亮主题
    elements.prismTheme.href = PRISM_THEMES[currentTheme];
    updateThemeIcon();
}
```

**亮点**：

- 一键切换，全局生效
- 代码高亮主题同步切换
- 刷新页面后主题保持不变

### 5. 深度思考模式

**技术栈**：API 参数控制 + 双内容渲染

```javascript
// 请求参数
const payload = {
    model: MODEL_NAME,
    messages: buildMessages(),
    stream: true,
    enable_thinking: true // 启用深度思考
};

// 渲染逻辑
if (choice.delta.reasoning_content) {
    // 显示思考过程（灰色背景，单独区域）
    showThinkingContent(choice.delta.reasoning_content);
}
if (choice.delta.content) {
    // 显示实际回复（正常样式）
    showResponseContent(choice.delta.content);
}
```

**亮点**：

- 思考过程与最终回复分离显示
- 实时展示 AI 推理链路
- 提升透明度和可解释性

## 项目结构

```
week01/homework/lingxi/
├── index.html          # 单页面布局（欢迎页 + 聊天页）
├── css/
│   └── index.css       # 完整样式（亮色/暗色主题）
├── js/
│   └── index.js        # 所有逻辑（约 700 行）
├── assets/             # 静态资源（图标、图片）
└── README.md           # 项目文档
```

## 快速开始

1. 在浏览器中打开 `index.html`
2. 首次使用需配置 API Key（点击设置按钮）默认的模型是阿里云百炼平台的Qwen3-plus,
3. 默认使用localstorage存储EMB\_BASE\_URL，
4. 输入问题或上传图片，点击发送即可对话

## 技术亮点总结

1. **零依赖**：纯原生 JavaScript，无需构建工具
2. **流式体验**：实时打字机效果，支持思考过程展示
3. **多模态**：文本 + 图片混合输入
4. **持久化**：主题、API Key、4轮内的文本类聊天历史全部本地存储
5. **可中断**：可随时停止 AI 生成

