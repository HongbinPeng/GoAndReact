/**
 * TODO 列表应用 JavaScript 脚本
 * 实现任务的添加、编辑、删除、过滤等功能
 */

// ==================== 变量定义 ====================

// 任务数据数组，存储所有任务对象
// 每个任务对象包含：id（唯一标识）、text（任务内容）、completed（完成状态）
// 首先尝试从浏览器本地存储获取数据，如果不存在则使用默认数据
let tasks = JSON.parse(localStorage.getItem('tasks')) || [
    { id: 1, text: '吃饭', completed: false },
    { id: 2, text: '睡觉', completed: false },
    { id: 3, text: '打豆豆', completed: false }
];

// 当前过滤状态，用于记录用户选择的过滤类型
// 可选值：'all'（所有）、'active'（进行中）、'completed'（已完成）
let currentFilter = 'all';

// ==================== DOM 元素获取 ====================

// 获取任务输入框元素，用于监听用户输入
const taskInput = document.getElementById('task-input');

// 获取任务列表容器元素，用于动态渲染任务
const taskList = document.getElementById('task-list');

// 获取所有过滤按钮（NodeList，类似数组）
const filterBtns = document.querySelectorAll('.filter-btn');

// 获取清空按钮元素
const clearBtn = document.querySelector('.clear-btn');


// ==================== 函数定义 ====================

/**
 * 初始化函数
 * 在页面加载时调用，设置初始状态并绑定事件
 */
function init() {
    // 首次渲染任务列表
    renderTasks();
    // 绑定所有事件监听器
    bindEvents();
}

/**
 * 渲染任务列表函数
 * 根据当前过滤状态筛选任务，并生成 HTML 显示在页面上
 */
function renderTasks() {
    // 根据 currentFilter 筛选任务
    const filteredTasks = tasks.filter(task => {
        // 如果过滤条件是"进行中"，返回未完成的任务
        if (currentFilter === 'active') return !task.completed;
        // 如果过滤条件是"已完成"，返回已完成的任务
        if (currentFilter === 'completed') return task.completed;
        // 否则返回所有任务
        return true;
    });

    // 清空任务列表容器
    taskList.innerHTML = '';

    // 遍历筛选后的任务，为每个任务创建 HTML 元素
    filteredTasks.forEach(task => {
        // 创建 li 元素
        const li = document.createElement('li');
        // 添加类名
        li.className = 'task-item';

        // 设置任务项的 HTML 内容
        // 使用模板字符串拼接 HTML
        // 如果任务已完成，添加 'completed' 类名，并默认勾选复选框
        li.innerHTML = `
            <div class="task-content">
                <input type="checkbox" ${task.completed ? 'checked' : ''} data-id="${task.id}">
                <span class="task-text ${task.completed ? 'completed' : ''}">${task.text}</span>
            </div>
            <div class="task-actions">
                <button class="action-btn edit" data-id="${task.id}">
                    <!-- 编辑图标 SVG -->
                    <svg width="14" height="14" viewBox="0 0 14 14" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <path d="M11.6667 1.16667L2.33333 10.5M1.16667 12.8333L3.5 10.5M11.6667 1.16667V3.5M11.6667 1.16667H9.33333" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                    </svg>
                    编辑
                </button>
                <button class="action-btn delete" data-id="${task.id}">
                    <!-- 删除图标 SVG -->
                    <svg width="14" height="14" viewBox="0 0 14 14" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <path d="M12.25 3.5H1.75M9.66666 3.5V1.75C9.66666 1.47386 9.55554 1.21071 9.36394 1.01912C9.17235 0.827527 8.9092 0.716406 8.63333 0.716406H5.36666C5.09079 0.716406 4.82764 0.827527 4.63606 1.01912C4.44446 1.21071 4.33333 1.47386 4.33333 1.75V3.5M9.66666 3.5V12.25C9.66666 12.5261 9.55554 12.7893 9.36394 12.9809C9.17235 13.1725 8.9092 13.2836 8.63333 13.2836H5.36666C5.09079 13.2836 4.82764 13.1725 4.63606 12.9809C4.44446 12.7893 4.33333 12.5261 4.33333 12.25V3.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                    </svg>
                    删除
                </button>
            </div>
            <button class="more-btn">...</button>
        `;

        // 将创建的任务项添加到任务列表容器中
        taskList.appendChild(li);
    });

    // 将任务数据保存到浏览器本地存储，实现数据持久化
    localStorage.setItem('tasks', JSON.stringify(tasks));
}

/**
 * 绑定事件函数
 * 为页面中的各种元素绑定事件监听器
 */
function bindEvents() {
    // 为输入框绑定键盘按下事件
    taskInput.addEventListener('keypress', function(e) {
        // 检测是否按下的是回车键（Enter）
        // 并且输入框内容不为空（去除首尾空格后）
        if (e.key === 'Enter' && this.value.trim() !== '') {
            // 调用添加任务函数
            addTask(this.value.trim());
            // 清空输入框
            this.value = '';
        }
    });

    // 为每个过滤按钮绑定点击事件
    filterBtns.forEach(btn => {
        btn.addEventListener('click', function() {
            // 移除所有按钮的 active 类
            filterBtns.forEach(b => b.classList.remove('active'));
            // 为当前点击的按钮添加 active 类
            this.classList.add('active');
            // 获取按钮上 data-filter 属性的值，更新当前过滤状态
            currentFilter = this.dataset.filter;
            // 重新渲染任务列表
            renderTasks();
        });
    });

    // 为清空按钮绑定点击事件
    clearBtn.addEventListener('click', function() {
        // 弹出确认对话框
        if (confirm('确定要清空所有任务吗？')) {
            // 清空任务数组
            tasks = [];
            // 重新渲染任务列表
            renderTasks();
        }
    });

    // 为任务列表绑定点击事件（事件委托）
    // 这种方式比给每个任务项单独绑定事件更高效
    taskList.addEventListener('click', function(e) {
        // 获取事件目标元素
        const target = e.target;
        // 查找最近的带有 data-id 属性的祖先元素，获取任务 ID
        // parseInt 将字符串转换为整数
        const taskId = parseInt(target.closest('[data-id]').dataset.id);

        // 判断点击的是哪种元素，执行相应的操作
        if (target.type === 'checkbox') {
            // 点击的是复选框，切换任务完成状态
            toggleTaskStatus(taskId);
        } else if (target.closest('.edit')) {
            // 点击的是编辑按钮，编辑任务
            editTask(taskId);
        } else if (target.closest('.delete')) {
            // 点击的是删除按钮，删除任务
            deleteTask(taskId);
        }
    });
}

/**
 * 添加任务函数
 * @param {string} text - 任务内容
 */
function addTask(text) {
    // 创建新任务对象
    const newTask = {
        // 使用时间戳作为唯一 ID
        id: Date.now(),
        // 任务内容
        text: text,
        // 默认未完成
        completed: false
    };

    // 将新任务添加到任务数组中
    tasks.push(newTask);

    // 重新渲染任务列表
    renderTasks();
}

/**
 * 切换任务完成状态函数
 * @param {number} id - 任务 ID
 */
function toggleTaskStatus(id) {
    // 在任务数组中查找对应 ID 的任务
    const task = tasks.find(task => task.id === id);

    // 如果找到任务
    if (task) {
        // 取反切换完成状态
        task.completed = !task.completed;
        // 重新渲染任务列表
        renderTasks();
    }
}

/**
 * 编辑任务函数
 * @param {number} id - 任务 ID
 */
function editTask(id) {
    // 在任务数组中查找对应 ID 的任务
    const task = tasks.find(task => task.id === id);

    // 如果找到任务
    if (task) {
        // 弹出输入框，第二个参数是输入框的默认显示值
        const newText = prompt('请输入新的任务内容：', task.text);

        // 如果用户点击了确定且输入不为空
        if (newText !== null && newText.trim() !== '') {
            // 更新任务内容（去除首尾空格）
            task.text = newText.trim();
            // 重新渲染任务列表
            renderTasks();
        }
    }
}

/**
 * 删除任务函数
 * @param {number} id - 任务 ID
 */
function deleteTask(id) {
    // 弹出确认对话框
    if (confirm('确定要删除这个任务吗？')) {
        // 使用 filter 方法过滤掉要删除的任务，返回剩余任务
        tasks = tasks.filter(task => task.id !== id);
        // 重新渲染任务列表
        renderTasks();
    }
}


// ==================== 应用初始化 ====================

// 调用初始化函数，启动应用
init();
