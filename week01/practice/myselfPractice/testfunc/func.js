//箭头函数
// 完整写法
const add = (a, b) => {
  return a + b;
};
document.getElementById("p1").innerHTML="add(1,2)="+add(1,2);
// 单表达式可省略 {} 和 return
const add1 = (a, b) => a + b;
document.getElementById("p1").innerHTML+="<br>add1(1,2)="+add1(1,2);
// 单参数可省略 ()
const double = n => n * 2;
// 无参数必须写 ()
const greet = () => 'Hello!';
// 返回对象字面量需要用 () 包裹
const makeUser = (name, age) => ({ name, age });


//回调函数
function ask(question, yes, no) {
  if (confirm(question)) {
    yes();
  } else {
    no();
  }
}

ask(
  '你同意吗？',
  () => console.log('同意了'),
  () => console.log('拒绝了')
);

// 箭头函数的 this 指向定义时的上下文，而不是调用时的上下文
const obj = {
  name: 'Alice',
  sayName: () => {
    console.log(this.name); // undefined
  }
};
obj.sayName();

(function () {
    var x = "Hello!!";      // 我将调用自己
})();