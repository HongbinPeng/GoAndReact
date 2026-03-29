# React.createElement()

```jsx
import React from "react";
import { createRoot } from "react-dom/client";

// 虚拟 DOM
const title = React.createElement("h1", { title: "🤠" }, "Hello World");

const root = createRoot(document.getElementById("root"));
// 把虚拟 DOM 转成真实 DOM 渲染到 root
root.render(title);
```

# 不够声明式

```jsx
import React from "react";
import { createRoot } from "react-dom/client";

// 麻烦，不够声明式
const title = React.createElement(
  "div",
  null,
  React.createElement(
    "ul",
    null,
    React.createElement("li", null, "Vue"),
    React.createElement("li", null, "React"),
    React.createElement("li", null, "Angular"),
  ),
);

const root = createRoot(document.getElementById("root"));
root.render(title);
```

# 解决上面的问题（JSX）

```jsx
import React from "react";
import { createRoot } from "react-dom/client";

// 内部会转换成 React.createElement() 这种形式
const title = (
  <div>
    <ul>
      <li>React</li>
      <li>Vue</li>
      <li>Angular</li>
    </ul>
  </div>
);

const root = createRoot(document.getElementById("root"));
root.render(title);
```

# JSX 的注意点

```jsx
import React from "react";
import { createRoot } from "react-dom/client";

const o = { name: "ifer" };

const arr = ["react", "vue", "angular"];

const title = (
  <div style={{ color: "red" }}>
    {/* 不能直接写对象 */}
    {/* <p>{o}</p> */}

    {/* key 要唯一，是为了在 VDOM Diff 的时候快速找到变更的那一个，按需更新 */}
    <ul>
      {arr.map((item) => (
        <li key={item}>{item}</li>
      ))}
    </ul>
  </div>
);

const root = createRoot(document.getElementById("root"));
root.render(title);
```

# 第一个组件

```jsx
import React from "react";
import { createRoot } from "react-dom/client";

function Hello() {
  return <div>这是第一个函数组件</div>;
}

const root = createRoot(document.getElementById("root"));
// 下面才是当做组件使用的，Hello 组件里面能使用组件的特性，例如里面能写 Hooks
root.render(<Hello/>);

// 下面是当做普通函数调用的，Hello 组件里面不能使用组件的特性，例如里面不能写 Hooks
// root.render(Hello());
```

# 状态的“不可变性”

不要直接修改原数据，而是永远要产生一份新数据 ，然后通过 setState() 用新的数据覆盖原数据。

```jsx
import React, { useState } from "react";

const App = () => {
  const [obj, setObj] = useState({
    count: 0,
  });
  const handleClick = () => {
    // Error
    // obj.count++;
    // Object.is(a, b)
    // setObj(obj);
    // Right
    setObj({
      ...obj,
      count: obj.count + 1,
    })
  };
  return (
    <div>
      <p>{obj.count}</p>
      <button onClick={handleClick}>click</button>
    </div>
  );
};

export default App;
```

# useState 的优化

```jsx
import React, { useState } from "react";

export default function App() {
  const [count, setCount] = useState(() => {
    // 大量计算
    let defaultCount = 0;
    for (let i = 0; i < 1000; i++) {
      defaultCount += i;
    }
    return defaultCount;
  });
  return (
    <div>
      <p>You clicked {count} times</p>
      <button onClick={() => setCount(count + 1)}>Click me</button>
    </div>
  );
}
```

# 受控和非受控（主要针对表单元素来说的）

受控的方式收集数据

```jsx
import React, { useState } from "react";

export default function App() {
  const [username, setUsername] = useState("");
  const changeText = (e) => {
    setUsername(e.target.value);
  };
  return (
    <ul>
      <li>
        <label htmlFor="username">用户名</label>
        <input id="username" type="text" value={username} onChange={changeText} />
      </li>
    </ul>
  );
}
```

非受控的方式收集表单数据

```jsx
import React, { useRef } from "react";

export default function App() {
  // Step1
  const inputRef = useRef(null);
  const handleChange = () => {
    // Step3
    console.log(inputRef.current.value);
    console.log(document.querySelector('input').value)
  };
  return (
    <div>
      {/* Step2 */}
      <input ref={inputRef} type="text" placeholder="输入内容" onChange={handleChange} />
    </div>
  );
}
```

# 单项数据流

1. 数据从哪儿来的就在哪改
2. 祖先的数据变化了，后代使用数据的地方也会更新


# 组件通信

父子（props）、子父（方法回调）、兄弟（状态提升）、跨层级组件通信（React.createContext 和 useContext）

# useEffect

```jsx
useEffect(() => {
  // 副作用回调函数
  return () => {
    // 清理函数
  }
}, 依赖项)

副作用回调函数执行时机：
1. 依赖项不写：初始化的时候会执行；任何状态变化的时候会执行
2. 依赖性为空数组：初始化的时候会执行
3. 依赖性为 [age, num]：初始化的时候会执行；只有 age 和 num 状态变化的时候执行


清理函数干啥的？清理定时器、解绑事件
清理函数的执行时机：
1. 组件卸载的时候执行
2. 下一次副作用回调函数执行的时候会执行
```

# 倒计时案例

```jsx
import React, { useState } from "react";

const App = () => {
  const [count, setCount] = useState(10);
  setInterval(() => {
    // !一个闭包
    setCount(count - 1);
  }, 1000);
  return (
    <div>
      <p>{count}</p>
    </div>
  );
};

export default App;
```

```jsx
import React, { useState } from "react";

const App = () => {
  const [count, setCount] = useState(10);
  let timer = null;
  const handleClick = () => {
    // !每次点击都是一个闭包
    clearInterval(timer);
    timer = setInterval(() => {
      // 注意点
      setCount((count) => count - 1);
    }, 1000);
  };
  return (
    <div>
      <p>{count}</p>
      <button onClick={handleClick}>点击按钮开始倒计时</button>
    </div>
  );
};

export default App;
```


```jsx
import React, { useState, useRef } from "react";


// !当前场景点击能不能清除定时器保证只有一个？如果不能为什么？如果能可能有什么问题？
// let timer = null

const App = () => {
  const [count, setCount] = useState(10);
  // 多次渲染共享同一个数据
  const timerRef = useRef(null)
  const handleClick = () => {
    clearInterval(timerRef.current)
    timerRef.current = setInterval(() => {
      // 注意点
      setCount((count) => count - 1);
    }, 1000);
  }
  return <div>
    <p>{count}</p>
    <button onClick={handleClick}>点击按钮开始倒计时</button>
  </div>;
};

export default App;
```

# useCallback 和 React.memo

App.jsx

```jsx
import React, { useState, useCallback } from 'react'
import Test from './Test'
export default function App() {
  const [count, setCount] = useState(1)
  const handleClick = useCallback(() => {
    console.log('handleClick')
  }, [])
  return (
    <div>
      <p>{count}</p>
      <button onClick={() => setCount(count + 1)}>+1</button>
      <Test handleClick={handleClick} />
    </div>
  )
}
```

Test.jsx

```jsx
import React from 'react'

export default React.memo(function Test() {
  console.log('Test')
  return (
    <div>Test</div>
  )
})
```