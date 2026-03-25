# Week01 课堂练习 - 学习收获总结

> **姓名**：彭鸿斌\
> **学校**：华中师范大学\
> **学号**：2024124379\
> **学习时间**：2026 年第一周

***

## 一、HTML 基础学习收获

### 1.1 HTML 文档结构

通过本周的学习，我深入理解了 HTML 文档的基本结构：

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>页面标题</title>
    <link rel="stylesheet" href="style.css">
</head>
<body>
    <!-- 页面内容 -->
    <script src="script.js"></script>
</body>
</html>
```

**关键知识点：**

- `<!DOCTYPE html>` 声明 HTML5 文档类型
- `<meta>` 标签设置字符编码和视口，确保移动端适配
- CSS 样式表在 `<head>` 中引入，确保页面渲染时就有样式
- JavaScript 脚本放在 `<body>` 末尾，确保 DOM 加载完成后再执行

### 1.2 常用 HTML 标签

我掌握了以下常用标签的使用：

- **结构标签**：`<div>`, `<ul>`, `<li>`, `<input>`, `<button>`
- **语义化标签**：理解语义化对 SEO 和可访问性的重要性
- **属性使用**：`id`, `class`, `placeholder`, `data-*` 等属性的应用

***

## 二、JavaScript 变量声明学习收获

### 2.1 var、let、const 的对比

这是我本周学习的重点内容，通过理论学习和实践练习，我深入理解了三个关键字的区别：

| 特性     | var   | let   | const |
| ------ | ----- | ----- | ----- |
| 作用域    | 函数作用域 | 块级作用域 | 块级作用域 |
| 变量提升   | ✅ 是   | ❌ 否   | ❌ 否   |
| 重复声明   | ✅ 允许  | ❌ 禁止  | ❌ 禁止  |
| 重新赋值   | ✅ 允许  | ✅ 允许  | ❌ 禁止  |
| 全局对象挂载 | ✅ 是   | ❌ 否   | ❌ 否   |
| 暂时性死区  | ❌ 无   | ✅ 有   | ✅ 有   |

### 2.2 实际应用经验

**使用 const 的场景：**

```javascript
// 声明常量
const API_URL = "https://api.example.com";

// 声明不需要修改的对象
const config = {
    timeout: 5000,
    retry: 3
};

// 数组引用不变，但可以修改内容
const tasks = [];
tasks.push({ id: 1, text: '学习 JavaScript' });
```

**使用 let 的场景：**

```javascript
// 循环计数器
for (let i = 0; i < 10; i++) {
    console.log(i);
}

// 需要重新赋值的变量
let isLoading = true;
isLoading = false;
```

**避免使用 var：**

```javascript
// ❌ 不推荐
var count = 0;

// ✅ 推荐
const count = 0;
let count = 0;
```

### 2.3 重要概念理解

**变量提升（Hoisting）：**

```javascript
// var 的变量提升
console.log(x); // undefined，不会报错
var x = 5;
console.log(x); // 5

// let/const 的暂时性死区
console.log(y); // ReferenceError: Cannot access 'y' before initialization
let y = 10;
```

**块级作用域：**

```javascript
if (true) {
    let x = 5;
    const y = 10;
    console.log(x, y); // 5, 10
}
console.log(x); // ReferenceError: x is not defined
console.log(y); // ReferenceError: y is not defined
```

**全局对象挂载：**

```javascript
var globalVar = "Hello";
console.log(window.globalVar); // "Hello"

let globalLet = "World";
console.log(window.globalLet); // undefined
```

***

## 三、DOM 操作学习收获

### 3.1 获取 DOM 元素

我学会了多种获取 DOM 元素的方法：

```javascript
// 通过 id 获取单个元素
const taskInput = document.getElementById('task-input');

// 通过 class 获取元素集合
const filterBtns = document.querySelectorAll('.filter-btn');

// 通过标签名获取
const buttons = document.getElementsByTagName('button');

// 通过选择器获取单个元素
const clearBtn = document.querySelector('.clear-btn');
```

### 3.2 事件监听

掌握了事件绑定的方法：

```javascript
// 键盘事件
taskInput.addEventListener('keypress', function(e) {
    if (e.key === 'Enter') {
        addTask(this.value);
    }
});

// 点击事件
filterBtns.forEach(btn => {
    btn.addEventListener('click', function() {
        // 处理逻辑
    });
});

// 事件委托（高效方式）
taskList.addEventListener('click', function(e) {
    const target = e.target;
    if (target.type === 'checkbox') {
        toggleTaskStatus(taskId);
    }
});
```

### 3.3 DOM 操作技巧

```javascript
// 创建元素
const li = document.createElement('li');
li.className = 'task-item';

// 设置 HTML 内容
li.innerHTML = `
    <div class="task-content">
        <span>${task.text}</span>
    </div>
`;

// 添加到 DOM
taskList.appendChild(li);

// 清空内容
taskList.innerHTML = '';

// 操作类名
btn.classList.add('active');
btn.classList.remove('active');
```

***

## 四、项目实战经验

### 4.1 TODO 列表项目

通过完成 TODO 列表项目，我综合运用了所学知识：

**实现的功能：**

1. ✅ 添加任务（Enter 键快速添加）
2. ✅ 标记任务完成/未完成
3. ✅ 编辑任务内容
4. ✅ 删除任务
5. ✅ 过滤任务（所有/进行中/已完成）
6. ✅ 清空所有任务
7. ✅ 数据持久化（localStorage）

**核心技术点：**

**数据持久化：**

```javascript
// 读取本地存储
let tasks = JSON.parse(localStorage.getItem('tasks')) || [];

// 保存到本地存储
localStorage.setItem('tasks', JSON.stringify(tasks));
```

**数组操作：**

```javascript
// 添加任务
tasks.push(newTask);

// 查找任务
const task = tasks.find(task => task.id === id);

// 过滤任务
const filteredTasks = tasks.filter(task => {
    if (currentFilter === 'active') return !task.completed;
    return true;
});

// 删除任务
tasks = tasks.filter(task => task.id !== id);
```

**唯一 ID 生成：**

```javascript
const newTask = {
    id: Date.now(), // 使用时间戳作为唯一 ID
    text: text,
    completed: false
};
```

### 4.2 代码规范意识

通过练习，我养成了良好的编码习惯：

1. **使用 const 和 let**：避免使用 var
2. **注释清晰**：关键代码都有详细注释
3. **函数命名**：使用动词 + 名词的命名方式（如 `addTask`, `renderTasks`）
4. **代码分组**：变量定义、DOM 获取、函数定义分组管理
5. **错误处理**：使用 confirm 确认重要操作

***

## 五、遇到的问题与解决方案

### 5.1 问题 1：DOM 元素获取为 null

**问题描述：**

```javascript
const taskInput = document.getElementById('task-input');
console.log(taskInput); // null
```

**原因分析：**
脚本在 HTML 元素加载前就执行了

**解决方案：**
将 `<script>` 标签放在 `</body>` 前面

### 5.2 问题 2：事件委托中的 target 判断

**问题描述：**
点击按钮的子元素（如图标 SVG）时，无法正确获取到按钮元素

**解决方案：**

```javascript
// 使用 closest 方法查找最近的祖先元素
const taskId = parseInt(target.closest('[data-id]').dataset.id);

// 判断点击的是哪种按钮
if (target.closest('.edit')) {
    editTask(taskId);
}
```

### 5.3 问题 3：对象数组的修改

**问题描述：**
修改对象属性后，界面没有更新

**解决方案：**

```javascript
// 找到对象并直接修改属性
const task = tasks.find(task => task.id === id);
if (task) {
    task.completed = !task.completed;
    renderTasks(); // 重新渲染
}
```

***

## 六、学习心得

通过本周的学习，我深刻体会到：光看理论是不够的，必须动手写代码才能真正理解，同时调试的工具使用也很重要，浏览器开发者工具调试在前端代码样式调整、网络请求分析方面作用很大，另外写注释的过程也是梳理思路的过程，把大问题拆成小功能逐个实现，同时要设计合理的数据结构，比如先使用选择器将需要绑定响应函数的元素和调整样式的元素柔和在同一个对象中，这样代码可读性更强，也能更方便的实现界面渲染

<br />

***

## 八、总结

第一周的学习让我从零开始，逐步掌握了：

✅ **HTML 基础**：文档结构、常用标签、属性使用\
✅ **JavaScript 变量**：var/let/const 的区别和应用场景\
✅ **DOM 操作**：元素获取、事件绑定、DOM 增删改\
✅ **项目实战**：TODO 列表的完整实现\
✅ **调试技巧**：浏览器开发者工具的使用\
✅ **代码规范**：良好的编码习惯和注释规范

这为后续的学习打下了坚实的基础。我将继续保持学习热情，不断提升自己的前端开发能力！

***

**感谢老师的悉心指导！**

*彭鸿斌*\
*2026 年第一周*
