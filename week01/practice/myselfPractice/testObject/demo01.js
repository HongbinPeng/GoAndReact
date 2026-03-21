//创建对象
// 对象字面量（推荐）
let user = {
  name: '张三',
  age: 25,
  isAdmin: false,
};
// 构造函数
let user2 = new Object();
user2.name = '李四';
user2.age=25
// 访问对象属性
// 点运算符
user.name;     // "张三"
// 方括号运算符（可以使用变量和任意字符串作为键）
user['name'];  // "张三"
let key = 'age';
user[key];     // 25
// 访问不存在的属性返回 undefined
user.email;    // undefined


// 添加、修改、删除属性
user.email = 'zhangsan@example.com';  // 添加
user.age = 26;                         // 修改
// delete user.isAdmin;                   // 删除
console.log(user);


// in 运算符
'name' in user;   // true
'email' in user;  // true
console.log('name' in user);
console.log('email' in user);

// 与 undefined 比较（不完全可靠，因为属性值可能就是 undefined）
user.name !== undefined;  // true

//遍历对象
// for...in
for (let key in user) {
  console.log(`${key}: ${user[key]}`);
}
//遍历对象
// Object.keys / values / entries（ES2017）
Object.keys(user);    // ["name", "age", "city"]
Object.values(user);  // ["张三", 25, "北京"]
Object.entries(user);  // [["name","张三"], ["age",25], ["city","北京"]]

// 结合解构遍历
for (let [key, value] of Object.entries(user)) {
  console.log(`${key}: ${value}`);
}
