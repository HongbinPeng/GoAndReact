# 前端开发学习总结

## 一、React 基础

### 1.1 组件化思想

React 的核心理念是**组件化**。将 UI 拆分为独立的、可复用的组件，每个组件只关注自身的逻辑和渲染。

```jsx
function Welcome({ name }) {
  return <h1>Hello, {name}</h1>;
}
```

### 1.2 Hooks 的使用

React Hooks 让函数组件也能拥有状态和生命周期：

- **useState** — 管理组件状态
- **useEffect** — 处理副作用（数据请求、订阅等）
- **useContext** — 跨组件共享数据
- **useRef** — 访问 DOM 或保存可变值
- **useMemo / useCallback** — 性能优化

```jsx
const [count, setCount] = useState(0);

useEffect(() => {
  document.title = `点击了 ${count} 次`;
}, [count]);
```

### 1.3 路由管理

使用 React Router 实现前端路由：

```jsx
import { BrowserRouter, Routes, Route } from 'react-router-dom';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/about" element={<About />} />
      </Routes>
    </BrowserRouter>
  );
}
```

---

## 附图

![flower](assets/flower.png)

---

## 二、Tailwind CSS

### 2.1 实用优先的 CSS 框架

Tailwind CSS 采用原子化 CSS 思想，通过组合小的工具类来构建 UI：

```html
<button class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded">
  按钮
</button>
```

### 2.2 响应式设计

Tailwind 内置响应式断点前缀：

| 前缀 | 最小宽度 | 说明 |
|------|---------|------|
| `sm:` | 640px | 小屏幕 |
| `md:` | 768px | 中等屏幕 |
| `lg:` | 1024px | 大屏幕 |
| `xl:` | 1280px | 超大屏幕 |

### 2.3 常用类名速查

- **布局**: `flex`, `grid`, `block`, `hidden`
- **间距**: `p-4`, `m-2`, `space-x-4`, `gap-6`
- **尺寸**: `w-full`, `h-screen`, `max-w-lg`
- **颜色**: `bg-white`, `text-gray-700`, `border-blue-500`
- **圆角**: `rounded`, `rounded-lg`, `rounded-full`
- **阴影**: `shadow`, `shadow-md`, `shadow-lg`

---

## 三、HTTP 请求与 API 交互

### 3.1 Axios 基础用法

```js
import axios from 'axios';

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
});

// 请求拦截器 — 添加认证令牌
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});
```

### 3.2 常见请求模式

```js
// GET 请求 — 获取列表
const { data } = await api.get('/courses', { params: { page: 1, pageSize: 10 } });

// POST 请求 — 创建资源
await api.post('/courses', { name: '新课程', description: '描述' });

// PUT 请求 — 更新资源
await api.put('/courses/1', { name: '更新后的名称' });

// DELETE 请求 — 删除资源
await api.delete('/courses/1');
```

---

## 四、项目实战要点

### 4.1 项目结构建议

```
src/
├── api/          # API 请求封装
├── components/   # 可复用组件
├── context/      # React Context
├── layouts/      # 布局组件
├── pages/        # 页面组件
├── App.jsx       # 根组件
├── main.jsx      # 入口文件
└── index.css     # 全局样式
```

### 4.2 开发流程

1. 分析需求，确定页面和功能模块
2. 搭建项目骨架（路由、布局、API 封装）
3. 逐个实现页面功能
4. 联调接口，处理异常情况
5. 优化体验（加载态、空状态、错误提示）

### 4.3 调试技巧

- 使用浏览器 DevTools 的 Network 面板查看请求
- 使用 React Developer Tools 检查组件状态
- `console.log` 配合断点调试定位问题
- 善用 VS Code 的搜索和代码跳转功能

---

> 💡 **提示**：动手实践是最好的学习方式。遇到问题时先阅读官方文档，然后尝试自己解决，最后再寻求帮助。
