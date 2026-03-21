<br />

# HTML 中 Script 标签的执行机制

## 基本执行机制

HTML 中的 <script> 标签默认是 同步执行 的，具体执行时机如下：

### 1. 常规 Script 标签

```HTML
<script src="script.js"></script>
```

- 执行时机 ：当浏览器解析到 <script> 标签时，会立即停止 HTML 解析，下载并执行脚本
- 执行顺序 ：按照在 HTML 中出现的顺序依次执行
- 阻塞特性 ：会阻塞 HTML 解析，直到脚本执行完成

### 2. 带有 async 属性的 Script 标签

```HTML
<script async src="script.js"></script>
```

- 执行时机 ：脚本下载完成后立即执行（异步下载，不阻塞 HTML 解析）
- 执行顺序 ：不保证按照在 HTML 中出现的顺序执行，谁先下载完成谁先执行
- 适用场景 ：独立的脚本，不依赖其他脚本，如统计代码、广告脚本

### 3. 带有 defer 属性的 Script 标签

```
<script defer src="script.js"></script>
```

- 执行时机 ：脚本下载完成后，等待 HTML 解析完成，在 DOMContentLoaded 事件触发前执行
- 执行顺序 ：按照在 HTML 中出现的顺序执行
- 适用场景 ：需要操作 DOM 的脚本，依赖其他脚本的脚本

## 执行流程详解

### 浏览器解析 HTML 的过程

1. 解析 HTML ：浏览器从上到下解析 HTML 文档
2. 遇到 Script 标签 ：
   - 无属性 ：停止解析 → 下载脚本 → 执行脚本 → 继续解析 HTML
   - async ：继续解析 HTML → 并行下载脚本 → 下载完成后立即执行（可能中断 HTML 解析）
   - defer ：继续解析 HTML → 并行下载脚本 → HTML 解析完成后按顺序执行
3. 解析完成 ：触发 DOMContentLoaded 事件
4. 所有资源加载完成 ：触发 load 事件

## 实际应用场景

### 1. 内联脚本

```HTML
<script>
    console.log('内联脚本执行');
</script>
```

- 立即执*行，*阻塞 HTML 解析
- *适用于需要在 HTML 解析过程*中执行的脚本

### 2. 外部脚本

```HTML
<script src="app.js"></script>
```

- 下载后执行，阻塞 HTML 解析
- 适用于需要在页面加载初期执行的脚本

### 3. 异步脚本

```HTML
<script async src="analytics.js"></script>
```

- 不阻塞 HTML 解析，下载完成后立即执行
- 适用于独立的、不依赖其他脚本的功能

### 4. 延迟脚本

```HTML
<script defer src="app.js"></script>
```

- 不阻塞 HTML 解析，HTML 解析完成后执行
- 适用于需要操作 DOM 的脚本

## 最佳实践

1. 将脚本放在底部 ：在 </body> 标签前添加脚本，避免阻塞页面渲染
2. 使用 defer ：对于需要操作 DOM 的脚本，使用 defer 属性
3. 使用 async ：对于独立的脚本，使用 async 属性
4. 动态加载 ：对于非关键脚本，使用动态创建 script 标签的方式加载

