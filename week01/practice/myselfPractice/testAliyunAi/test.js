const input = document.getElementById("input");
const btn = document.getElementById("btn");
const outputText = document.getElementById("outputText");
const aimaodel={
    model:"qwen3.5-plus",
    apiKey:"sk-e4dcae9060904558a2f64d6eb12249e0",
    baseUrl:"https://dashscope.aliyuncs.com/compatible-mode/v1",
}

console.log('DOM元素加载完成');
console.log('input:', input);
console.log('btn:', btn);
console.log('outputText:', outputText);

// 绑定发送按钮事件
btn.addEventListener('click', send);
console.log('绑定了点击事件');

// 按Enter键发送
input.addEventListener('keypress', function(e) {
    if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        send();
    }
});
console.log('绑定了键盘事件');

// Markdown 解析函数 - 使用 marked.js
function renderMarkdown(text) {
    // 使用 marked.js 解析 Markdown
    return marked.parse(text);
}

// 添加复制按钮
function addCopyButtons() {
    const preElements = document.querySelectorAll('pre');
    preElements.forEach(pre => {
        // 检查是否已有复制按钮
        if (!pre.querySelector('.copy-btn')) {
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
                }).catch(err => {
                    console.error('复制失败:', err);
                    copyBtn.textContent = '复制失败';
                    setTimeout(() => {
                        copyBtn.textContent = '复制';
                    }, 2000);
                });
            });
            
            pre.appendChild(copyBtn);
        }
    });
}

function send() {
    console.log('开始发送请求');
    const text = input.value;
    if(!text){
        alert("请输入内容");
        return;
    }
    
    // 清空输出
    outputText.innerHTML = "正在生成...";
    // 禁用按钮
    btn.disabled = true;
    btn.textContent = "生成中...";
    
    console.log('请求URL:', `${aimaodel.baseUrl}/chat/completions`);
    console.log('请求数据:', JSON.stringify({
        model: aimaodel.model,
        messages: [
            { role: "system", content: "You are a helpful assistant." },
            { role: "user", content: text }
        ],
        stream: true
    }));
    
    // 发起流式请求
    fetch(`${aimaodel.baseUrl}/chat/completions`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${aimaodel.apiKey}`
        },
        body: JSON.stringify({
            model: aimaodel.model,
            messages: [
                { role: "system", content: "You are a helpful assistant." },
                { role: "user", content: text }
            ],
            stream: true
        })
    })
    .then(response => {
        console.log('响应状态:', response.status);
        console.log('响应头:', response.headers);
        
        if (!response.ok) {
            throw new Error('API请求失败');
        }
        
        console.log('开始处理流式响应');
        const reader = response.body.getReader();
        let result = '';
        
        function processChunk({ done, value }) {
            if (done) {
                console.log('流式响应结束');
                // 完成处理
                btn.disabled = false;
                btn.textContent = "发送";
                return;
            }
            
            // 解析SSE格式的响应
            const chunk = new TextDecoder().decode(value); // 解码为字符串
            console.log('收到数据块:', chunk);
            const lines = chunk.split('\n');
            
            lines.forEach(line => {
                line = line.trim();
                if (line.startsWith('data: ')) {
                    const data = line.substring(6);
                    console.log('解析数据:', data);
                    if (data === '[DONE]') {
                        console.log('收到结束标记');
                        return;
                    }
                    
                    try {
                        const event = JSON.parse(data);
                        console.log('解析事件:', event);
                        if (event.choices && event.choices[0] && event.choices[0].delta && event.choices[0].delta.content) {
                            const content = event.choices[0].delta.content;
                            result += content;
                            console.log('更新结果:', result);
                            // 使用Markdown解析
                            const renderedContent = renderMarkdown(result);
                            outputText.innerHTML = renderedContent;
                            // 初始化Prism语法高亮
                            if (typeof Prism !== 'undefined') {
                                Prism.highlightAll();
                            }
                            // 添加复制按钮
                            addCopyButtons();
                        }
                    } catch (e) {
                        console.error('解析响应失败:', e);
                    }
                }
            });
            
            reader.read().then(processChunk);
        }
        
        return reader.read().then(processChunk);
    })
    .catch(error => {
        console.error('请求失败:', error);
        outputText.textContent = '抱歉，请求失败，请稍后再试。';
        btn.disabled = false;
        btn.textContent = "发送";
    });
}
