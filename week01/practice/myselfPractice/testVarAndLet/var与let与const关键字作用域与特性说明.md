# var、let 与 const 关键字作用域与特性说明

## 作用域的基本概念

**作用域**是指变量或函数的可访问范围，决定了代码块中变量和函数的可见性。

在 JavaScript 中，主要有以下几种作用域：

- **全局作用域**：在整个代码中都可访问
- **函数作用域**：仅在函数内部可访问
- **块级作用域**：在 `{}` 块内可访问（ES6 引入）

## var 关键字

### 特性

1. **函数作用域**：var 声明的变量只在函数内部有效
2. **变量提升**：var 声明的变量会被提升到作用域顶部
3. **重复声明**：允许在同一作用域内重复声明同一变量
4. **全局对象挂载**：在全局作用域中声明的 var 变量会成为 window 对象的属性

### 代码示例

```javascript
// 变量提升示例
console.log(x); // 输出: undefined
var x = 5;
console.log(x); // 输出: 5

// 函数作用域示例
function test() {
    var y = 10;
    console.log(y); // 输出: 10
}
test();
console.log(y); // 报错: y is not defined

// 全局对象挂载示例
var globalVar = "Hello";
console.log(window.globalVar); // 输出: Hello
```

## let 关键字

### 特性

1. **块级作用域**：let 声明的变量在块内有效
2. **无变量提升**：let 声明的变量不会被提升，存在暂时性死区
3. **不可重复声明**：在同一作用域内不能重复声明同一变量
4. **全局对象不挂载**：在全局作用域中声明的 let 变量不会成为 window 对象的属性

### 代码示例

```javascript
// 块级作用域示例
if (true) {
    let x = 5;
    console.log(x); // 输出: 5
}
console.log(x); // 报错: x is not defined

// 暂时性死区示例
console.log(y); // 报错: Cannot access 'y' before initialization
let y = 10;

// 全局对象不挂载示例
let globalLet = "World";
console.log(window.globalLet); // 输出: undefined
```

## const 关键字

### 特性

1. **块级作用域**：const 声明的变量在块内有效
2. **无变量提升**：const 声明的变量不会被提升，存在暂时性死区
3. **不可重复声明**：在同一作用域内不能重复声明同一变量
4. **不可重新赋值**：const 声明的变量一旦赋值，不能重新赋值
5. **全局对象不挂载**：在全局作用域中声明的 const 变量不会成为 window 对象的属性
6. **对象可修改**：const 声明的对象，其引用不可变，但对象内部属性可以修改

### 代码示例

```javascript
// 不可重新赋值示例
const PI = 3.14;
PI = 3.14159; // 报错: Assignment to constant variable

// 对象可修改示例
const person = { name: "John" };
person.name = "Jane"; // 允许修改对象属性
console.log(person.name); // 输出: Jane

person = { name: "Tom" }; // 报错: Assignment to constant variable
```

## 三者对比

| 特性     | var   | let   | const |
| :----- | :---- | :---- | :---- |
| 作用域    | 函数作用域 | 块级作用域 | 块级作用域 |
| 变量提升   | 是     | 否     | 否     |
| 重复声明   | 允许    | 禁止    | 禁止    |
| 重新赋值   | 允许    | 允许    | 禁止    |
| 全局对象挂载 | 是     | 否     | 否     |
| 暂时性死区  | 无     | 有     | 有     |

## 最佳实践

1. **优先使用 const**：
   - 当变量值不需要改变时，使用 const
   - 提高代码可读性和可维护性
   - 防止意外修改
2. **需要重新赋值时使用 let**：
   - 当变量值需要在后续代码中改变时使用 let
   - 利用块级作用域减少变量泄露
3. **避免使用 var**：
   - var 的函数作用域可能导致变量泄露
   - 变量提升可能导致意外行为
   - 重复声明可能掩盖错误

## 实际应用场景

### 使用 const 的场景

- 声明常量（如 PI）
- 声明不需要修改的对象或数组
- 声明模块导入

### 使用 let 的场景

- 声明循环计数器
- 声明需要在条件语句中修改的变量
- 声明函数内部需要重新赋值的变量

### 避免使用 var 的场景

- 所有可以使用 let 或 const 的场景
- 全局变量声明
- 函数内部变量声明

## 代码示例：实际应用

```javascript
// 推荐用法
const API_URL = "https://api.example.com";
const users = [];

function fetchData() {
    let isLoading = true;
    const response = await fetch(API_URL);
    const data = await response.json();
    
    for (let i = 0; i < data.length; i++) {
        const user = data[i];
        users.push(user);
    }
    
    isLoading = false;
    return users;
}

// 不推荐用法
var count = 0;
function increment() {
    var count = 1; // 函数作用域，不会修改全局 count
    count++;
    console.log(count); // 输出: 2
}
increment();
console.log(count); // 输出: 0
```

## 总结

- **var**：函数作用域，变量提升，允许重复声明，全局变量挂载到 window
- **let**：块级作用域，无变量提升，禁止重复声明，全局变量不挂载到 window
- **const**：块级作用域，无变量提升，禁止重复声明，禁止重新赋值，全局变量不挂载到 window

在现代 JavaScript 开发中，应优先使用 const 和 let，避免使用 var，以提高代码的可读性、可维护性和安全性。
