/**
 * 灵犀 AI 对话助手
 * 实现真实的 AI 对话功能，支持深色/浅色主题切换、文件上传、流式响应、Markdown 渲染等
 */

// ==================== 全局变量 ====================

// 主题相关
let currentTheme = localStorage.getItem('theme') || 'light';
// API 相关
let API_KEY = '';
let BASE_URL = '';
let MODEL_NAME = '';
let isStreaming = false;
let abortController = null;
// 对话历史
const MAX_HISTORY = 4;

// DOM 元素
const elements = {
    themeToggle: document.getElementById('themeToggle'),//主题切换按钮
    welcomeSection: document.getElementById('welcomeSection'),//欢迎界面
    chatSection: document.getElementById('chatSection'),//对话界面
    chatMessages: document.getElementById('chatMessages'),//对话界面的div
    messageInput: document.getElementById('messageInput'),//输入框
    sendBtn: document.getElementById('sendBtn'),//发送按钮
    clearChat: document.getElementById('clearChat'),//清除对话按钮
    fileInput: document.getElementById('fileInput'),//文件上传按钮
    imageModal: document.getElementById('imageModal'),//图片预览模态框
    previewImage: document.getElementById('previewImage'),//图片预览模态框中的图片
    closeModal: document.querySelector('.close-modal'),//图片预览模态框中的关闭按钮
    suggestionCards: document.querySelectorAll('.suggestion-card')//建议卡片
};

// 发送按钮状态
let isSendBtnStopping = false;

// ==================== 初始化 ====================

function init() {
    // 加载主题
    loadTheme();
    
    // 加载 API 配置
    loadAPIConfig();
    
    // 绑定事件
    bindEvents();
}

// ==================== 主题相关 ====================

function loadTheme() {
    try {
        const theme = localStorage.getItem('theme');
        if (theme) {
            currentTheme = theme;
        }
    } catch (e) {
        console.warn('无法读取主题设置:', e);
    }
    document.body.className = currentTheme;
    updateThemeIcon();
}

function toggleTheme() {
    currentTheme = currentTheme === 'light' ? 'dark' : 'light';
    document.body.className = currentTheme;
    try {
        localStorage.setItem('theme', currentTheme);
    } catch (e) {
        console.warn('无法保存主题设置:', e);
    }
    updateThemeIcon();
}

function updateThemeIcon() {
    const icon = elements.themeToggle.querySelector('svg');
    if (currentTheme === 'dark') {
        icon.innerHTML = '<circle cx="12" cy="12" r="5"></circle><line x1="12" y1="1" x2="12" y2="3"></line><line x1="12" y1="21" x2="12" y2="23"></line><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line><line x1="1" y1="12" x2="3" y2="12"></line><line x1="21" y1="12" x2="23" y2="12"></line><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line>';
    } else {
        icon.innerHTML = '<path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path>';
    }
}

// ==================== API 配置管理 ====================

function loadAPIConfig() {
    // 尝试从 localStorage 加载
    try {
        API_KEY = localStorage.getItem('LINGXI_API_KEY') || '';
        BASE_URL = localStorage.getItem('EMB_BASE_URL') || '';
        MODEL_NAME = localStorage.getItem('EMB_MODEL_NAME') || '';
    } catch (e) {
        console.warn('无法从 localStorage 加载 API 配置:', e);
        API_KEY = '';
        BASE_URL = '';
        MODEL_NAME = '';
    }
    
    // 如果没有，使用默认配置（直接从 .env 文件内容获取）
    if (!API_KEY || !BASE_URL || !MODEL_NAME) {
        // 直接设置默认配置
        API_KEY = 'sk-e4dcae9060904558a2f64d6eb12249e0';
        BASE_URL = 'https://dashscope.aliyuncs.com/compatible-mode/v1';
        MODEL_NAME = 'qwen3.5-plus';
        
        // 保存到 localStorage
        try {
            localStorage.setItem('LINGXI_API_KEY', API_KEY);
            localStorage.setItem('EMB_BASE_URL', BASE_URL);
            localStorage.setItem('EMB_MODEL_NAME', MODEL_NAME);
        } catch (e) {
            console.warn('无法保存 API 配置到 localStorage:', e);
        }
    }
}

function setAPIConfig(key, baseURL, modelName) {
    API_KEY = key;
    BASE_URL = baseURL;
    MODEL_NAME = modelName;
    localStorage.setItem('LINGXI_API_KEY', key);
    localStorage.setItem('EMB_BASE_URL', baseURL);
    localStorage.setItem('EMB_MODEL_NAME', modelName);
}

// ==================== 事件绑定 ====================

function bindEvents() {
    // 主题切换
    elements.themeToggle.addEventListener('click', toggleTheme);
    
    // 消息输入框
    elements.messageInput.addEventListener('input', handleInputChange);
    elements.messageInput.addEventListener('keypress', handleKeyPress);
    
    // 发送按钮
    elements.sendBtn.addEventListener('click', sendMessage);
    
    // 清除对话
    elements.clearChat.addEventListener('click', clearChat);
    
    // 文件上传
    elements.fileInput.addEventListener('change', handleFileUpload);
    
    // 图片预览模态框
    elements.closeModal.addEventListener('click', () => {
        elements.imageModal.style.display = 'none';
    });
    
    // 点击模态框外部关闭
    window.addEventListener('click', (e) => {
        if (e.target === elements.imageModal) {
            elements.imageModal.style.display = 'none';
        }
    });
    
    // 快捷建议卡片
    elements.suggestionCards.forEach(card => {
        card.addEventListener('click', () => {
            const prompt = card.dataset.prompt;
            elements.messageInput.value = prompt;
            sendMessage();
        });
    });
}

// ==================== 输入处理 ====================

function handleInputChange() {
    elements.sendBtn.disabled = !elements.messageInput.value.trim();
    // 自动调整 textarea 高度
    elements.messageInput.style.height = 'auto';
    elements.messageInput.style.height = Math.min(elements.messageInput.scrollHeight, 200) + 'px';
}

function handleKeyPress(e) {
    // Enter 发送消息，Shift+Enter 换行
    if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        sendMessage();
    }
}

// ==================== 消息处理 ====================

function sendMessage() {
    // 如果发送按钮当前是停止状态，不执行发送操作
    if (isSendBtnStopping) return;
    
    const message = elements.messageInput.value.trim();
    if (!message) return;
    
    // 显示用户消息
    addMessage('user', message);
    
    // 清空输入框
    elements.messageInput.value = '';
    elements.sendBtn.disabled = true;
    // 重置输入框高度
    elements.messageInput.style.height = 'auto';
    
    // 隐藏欢迎页面，显示对话页面
    elements.welcomeSection.style.display = 'none';
    elements.chatSection.style.display = 'flex';
    
    // 滚动到底部
    scrollToBottom();
    
    // 发送请求到 AI
    sendToAI(message);
}

function addMessage(type, content) {
    const messageDiv = document.createElement('div');
    messageDiv.className = `message ${type}-message`;
    
    if (type === 'ai') {
        messageDiv.innerHTML = `
            <div class="message-avatar">灵</div>
            <div class="message-content">
                <div class="reasoning-content"></div>
                <div class="actual-content">${content}</div>
            </div>
        `;
    } else {
        messageDiv.innerHTML = `
            <div class="message-content">${content}</div>
            <div class="message-avatar">我</div>
        `;
    }
    
    elements.chatMessages.appendChild(messageDiv);
    scrollToBottom();
    
    return messageDiv;
}

function sendToAI(message) {
    if (!API_KEY || !BASE_URL || !MODEL_NAME) {
        // 如果没有 API 配置，使用模拟响应
        simulateAIResponse(message);
        return;
    }
    
    // 显示 AI 正在输入
    const aiMessageDiv = addMessage('ai', '<span class="typing">正在思考...</span>');
    
    // 创建中止控制器
    abortController = new AbortController();
    isStreaming = true;
    
    // 更改发送按钮为停止按钮
    elements.sendBtn.innerHTML = '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="6" y="4" width="4" height="16"></rect><rect x="14" y="4" width="4" height="16"></rect></svg>';
    elements.sendBtn.disabled = false;
    isSendBtnStopping = true;
    
    // 保存原始发送按钮点击事件
    const originalSendBtnClick = elements.sendBtn.onclick;
    
    // 添加停止功能
    elements.sendBtn.onclick = () => {
        if (abortController) {
            abortController.abort();
            isStreaming = false;
            // 恢复发送按钮
            restoreSendButton();
            // 只移除 AI 消息，保留用户消息
            const messages = elements.chatMessages.children;
            if (messages.length >= 1) {
                // 只移除最后一条消息（AI消息）
                elements.chatMessages.removeChild(messages[messages.length - 1]);
            }
        }
    };
    
    // 获取对话历史
    const chatHistory = getChatHistory();
    
    // 构建消息数组
    const messages = [
        {
            role: 'system',
            content: '你是一个智能对话助手，名为灵犀。请使用友好、自然的语言回答用户的问题。'
        }
    ];
    
    // 添加对话历史作为上下文
    chatHistory.forEach(item => {
        messages.push(item);
    });
    
    // 添加当前用户消息
    messages.push({
        role: 'user',
        content: message
    });
    
    // 调用阿里云百炼 OpenAI 兼容 API
    fetch(`${BASE_URL}/chat/completions`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${API_KEY}`
        },
        body: JSON.stringify({
            model: MODEL_NAME,
            messages: messages,
            stream: true
        }),
        signal: abortController.signal
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('API 请求失败');
        }
        
        const reader = response.body.getReader();
        let aiResponse = '';
        let currentReasoning = '';
        let isFirstChunk = true;
        
        function processChunk({ done, value }) {
            if (done) {
                isStreaming = false;
                // 恢复发送按钮
                restoreSendButton();
                // 显示最终回答
                aiMessageDiv.querySelector('.actual-content').innerHTML = renderMarkdown(aiResponse);
                // 确保思考过程可见
                const reasoningContent = aiMessageDiv.querySelector('.reasoning-content');
                if (reasoningContent.textContent.trim()) {
                    reasoningContent.classList.add('has-content');
                }
                addCopyButtons(aiMessageDiv);
                // 将用户消息和 AI 响应添加到对话历史
                addToChatHistory('user', message);
                addToChatHistory('assistant', aiResponse);
                return;
            }
            
            if (!isStreaming) {
                reader.cancel();
                return;
            }
            
            // 解析 SSE 格式的响应
            const chunk = new TextDecoder().decode(value);
            const lines = chunk.split('\n');
            
            let hasContentUpdate = false;
            let hasReasoningUpdate = false;
            
            lines.forEach(line => {
                line = line.trim();
                if (line.startsWith('data: ')) {
                    const data = line.substring(6);
                    if (data === '[DONE]') return;
                    
                    try {
                        const event = JSON.parse(data);
                        if (event.choices && event.choices[0]) {
                            const choice = event.choices[0];
                            
                            // 处理思考过程
                            if (choice.delta && choice.delta.reasoning_content) {
                                currentReasoning += choice.delta.reasoning_content;
                                hasReasoningUpdate = true;
                                isFirstChunk = false;
                            }
                            
                            // 处理实际内容
                            if (choice.delta && choice.delta.content) {
                                aiResponse += choice.delta.content;
                                hasContentUpdate = true;
                                isFirstChunk = false;
                            }
                        }
                    } catch (e) {
                        console.error('解析响应失败:', e);
                    }
                }
            });
            
            // 实时更新思考过程，不需要 markdown 渲染
            if (hasReasoningUpdate) {
                // 简单处理，去除可能的 markdown 标记
                const cleanedReasoning = currentReasoning.replace(/[\*`]/g, '');
                const reasoningContent = aiMessageDiv.querySelector('.reasoning-content');
                reasoningContent.textContent = cleanedReasoning;
                // 添加 has-content 类来显示思考过程区域
                reasoningContent.classList.add('has-content');
                // 滚动到思考内容底部
                reasoningContent.scrollTop = reasoningContent.scrollHeight;
            }
            
            // 批量更新 DOM，减少重绘
            if (hasContentUpdate) {
                aiMessageDiv.querySelector('.actual-content').innerHTML = renderMarkdown(aiResponse);
                scrollToBottom();
            }
            
            reader.read().then(processChunk);
        }
        
        return reader.read().then(processChunk);
    })
    .catch(error => {
        if (error.name !== 'AbortError') {
            console.error('API 调用失败:', error);
            aiMessageDiv.querySelector('.actual-content').textContent = '抱歉，我暂时无法回答你的问题。请稍后再试。';
            // 恢复发送按钮
            restoreSendButton();
            // 将用户消息和 AI 错误响应添加到对话历史
            addToChatHistory('user', message);
            addToChatHistory('assistant', '抱歉，我暂时无法回答你的问题。请稍后再试。');
        } else {
            // 如果是用户中止，只移除 AI 消息，保留用户消息
            const messages = elements.chatMessages.children;
            if (messages.length >= 1) {
                // 只移除最后一条消息（AI消息）
                elements.chatMessages.removeChild(messages[messages.length - 1]);
            }
        }
        isStreaming = false;
    });
}

// 恢复发送按钮状态
function restoreSendButton() {
    elements.sendBtn.innerHTML = '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="22" y1="2" x2="11" y2="13"></line><polygon points="22 2 15 22 11 13 2 9 22 2"></polygon></svg>';
    elements.sendBtn.disabled = !elements.messageInput.value.trim();
    isSendBtnStopping = false;
}

function simulateAIResponse(message) {
    // 模拟 AI 响应，用于演示
    const aiMessageDiv = addMessage('ai', '<span class="typing">正在思考...</span>');
    
    // 模拟响应内容
    const responses = {
        '你好': '你好！我是灵犀，很高兴为你服务。请问有什么我可以帮助你的吗？',
        '今天天气如何': '抱歉，我无法获取实时天气信息。不过你可以通过天气应用或搜索引擎查询当地的天气情况。',
        '写一篇环保短文': '《绿色生活，从点滴做起》\n\n在这个快节奏的时代，环保似乎成了一个宏大的命题，让许多人望而却步。但实际上，环保并非遥不可及的目标，它存在于我们日常生活的每一个细节中。\n\n早晨起床，用可重复使用的水杯接水，而不是选择一次性塑料杯；上班路上，选择公共交通或骑行，减少碳排放；午餐时间，自带餐具，拒绝一次性筷子和餐盒；下班回家，将垃圾分类投放，让资源得到有效回收。\n\n这些看似微小的举动，汇聚起来就是一股强大的力量。当我们每个人都从自身做起，从点滴小事做起，我们的地球将会变得更加美好。\n\n环保不是口号，而是一种生活态度。让我们携手共进，为了更蓝的天空、更清的河水、更绿的大地，共同努力。',
        '如何学习编程': '学习编程是一个循序渐进的过程，以下是一些建议：\n\n1. **选择合适的编程语言**：对于初学者，Python 是一个不错的选择，它语法简单，易于理解。\n\n2. **建立基础概念**：了解变量、数据类型、控制流、函数等基本概念。\n\n3. **实践是关键**：通过实际项目练习，将理论知识转化为实际技能。\n\n4. **利用在线资源**：如 Codecademy、Coursera、YouTube 等平台都有优质的编程教程。\n\n5. **加入社区**：与其他学习者交流，分享经验，解决问题。\n\n6. **保持耐心**：编程学习中遇到困难是正常的，坚持下去，你会逐渐掌握它。\n\n记住，编程是一项技能，需要不断练习和积累。祝你学习顺利！',
        '推荐一些好看的电影': '当然！以下是几部不同类型的优秀电影推荐：\n\n**剧情片**：\n- 《肖申克的救赎》（1994）\n- 《楚门的世界》（1998）\n- 《活着》（1994）\n\n**科幻片**：\n- 《星际穿越》（2014）\n- 《银翼杀手2049》（2017）\n- 《阿凡达》（2009）\n\n**动画电影**：\n- 《千与千寻》（2001）\n- 《寻梦环游记》（2017）\n- 《疯狂动物城》（2016）\n\n**喜剧片**：\n- 《怦然心动》（2010）\n- 《三傻大闹宝莱坞》（2009）\n- 《白日梦想家》（2013）\n\n希望你能找到喜欢的电影！'
    };
    
    // 查找匹配的响应
    let response = responses[message] || '感谢你的问题！这是一个模拟响应。在实际应用中，我会使用真实的 AI 模型来回答你的问题。';
    
    // 模拟打字效果
    let index = 0;
    const typingSpeed = 50; // 打字速度（毫秒）
    
    const interval = setInterval(() => {
        if (index < response.length) {
            const currentText = response.substring(0, index + 1);
            aiMessageDiv.querySelector('.message-content').innerHTML = renderMarkdown(currentText);
            scrollToBottom();
            index++;
        } else {
            clearInterval(interval);
            addCopyButtons(aiMessageDiv);
            // 将用户消息和 AI 响应添加到对话历史
            addToChatHistory('user', message);
            addToChatHistory('assistant', response);
        }
    }, typingSpeed);
}

// ==================== Markdown 渲染 ====================

function renderMarkdown(text) {
    // 使用 marked.js 解析 Markdown
    const html = marked.parse(text);
    // 延迟执行Prism高亮，确保DOM已经更新
    setTimeout(() => {
        Prism.highlightAll();
    }, 0);
    return html;
}

function addCopyButtons(messageDiv) {
    // 为代码块添加复制按钮
    const preElements = messageDiv.querySelectorAll('pre');
    preElements.forEach(pre => {
        const copyBtn = document.createElement('button');
        copyBtn.className = 'copy-btn';
        copyBtn.textContent = '复制';
        
        copyBtn.addEventListener('click', () => {
            const code = pre.querySelector('code').textContent;
            navigator.clipboard.writeText(code).then(() => {
                copyBtn.textContent = '已复制';
                setTimeout(() => {
                    copyBtn.textContent = '复制';
                }, 2000);
            });
        });
        
        pre.appendChild(copyBtn);
    });
}

// ==================== 图片上传 ====================

function handleFileUpload(e) {
    const file = e.target.files[0];
    if (!file) return;
    
    // 检查文件类型
    if (!file.type.startsWith('image/')) {
        alert('请上传图片文件！');
        return;
    }
    
    // 预览图片
    const reader = new FileReader();
    reader.onload = function(e) {
        elements.previewImage.src = e.target.result;
        elements.imageModal.style.display = 'flex';
    };
    reader.readAsDataURL(file);
}

// ==================== 对话管理 ====================

function getChatHistory() {
    try {
        const history = localStorage.getItem('lingxi_chat_history');
        return history ? JSON.parse(history) : [];
    } catch (e) {
        console.warn('无法读取对话历史:', e);
        return [];
    }
}

function saveChatHistory(history) {
    // 只保存最近的 MAX_HISTORY 轮对话
    const trimmedHistory = history.slice(-MAX_HISTORY * 2); // 每轮对话包含用户和 AI 两条消息
    try {
        localStorage.setItem('lingxi_chat_history', JSON.stringify(trimmedHistory));
    } catch (e) {
        console.warn('无法保存对话历史:', e);
    }
}

function addToChatHistory(role, content) {
    const history = getChatHistory();
    history.push({ role, content });
    saveChatHistory(history);
}

function clearChat() {
    if (confirm('确定要清除所有对话吗？')) {
        elements.chatMessages.innerHTML = '';
        elements.welcomeSection.style.display = 'flex';
        elements.chatSection.style.display = 'none';
        try {
            localStorage.removeItem('lingxi_chat_history');
        } catch (e) {
            console.warn('无法清除对话历史:', e);
        }
    }
}

function scrollToBottom() {
    elements.chatMessages.scrollTop = elements.chatMessages.scrollHeight;
}

// ==================== 初始化应用 ====================

init();