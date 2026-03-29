/** @typedef {{ id: string; title: string; options: string[]; correctIndex: number }} Question */

/** @type {Question[]} */
export const questions = [
  {
    id: 'q1',
    title: 'HTML5 中 canvas 元素的用途是什么？',
    options: ['绘制图形', '播放音频', '用于存储数据', '显示视频'],
    correctIndex: 0,
  },
  {
    id: 'q2',
    title: 'JavaScript 中用于声明常量的关键字是？',
    options: ['var', 'let', 'const', 'function'],
    correctIndex: 2,
  },
  {
    id: 'q3',
    title: 'CSS 中用于设置元素外边距的属性是？',
    options: ['padding', 'margin', 'border', 'width'],
    correctIndex: 1,
  },
  {
    id: 'q4',
    title: 'React 中用于在函数组件内保存状态的内置 Hook 是？',
    options: ['useEffect', 'useState', 'useMemo', 'useRef'],
    correctIndex: 1,
  },
  {
    id: 'q5',
    title: 'HTTP 协议中，常用于获取资源的请求方法是？',
    options: ['POST', 'PUT', 'GET', 'DELETE'],
    correctIndex: 2,
  },
]
