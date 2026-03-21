console.log("AboutWindowObject 浏览器窗口宽度为：" + window.innerWidth); // 浏览器窗口宽度
console.log("AboutWindowObject 当前页面 URL 为：" + window.location.href); // 当前页面 URL
console.log("AboutWindowObject 当前页面路径为：" + window.location.pathname); // 当前页面路径
window.sessionStorage.setItem("username", "penghongbin");
var carName="奔驰";
console.log("AboutWindowObject 我有一辆" + window.carName);//var声明的变量会成为window对象的属性
let carName2="沃尔沃";
console.log("AboutWindowObject 我有一辆" + window.carName2);//let声明的变量不会成为window对象的属性

var x=10
function test1(){
    console.log("test1函数内部的x"+x);
    x=20;
    var y=10; 
}
test1();
console.log("函数外的全局函数x:"+x);


console.log("未声明的变量y:"+y);//var声明的变量可以在声明前使用，但是是undefined
// let 关键字定义的变量则不可以在使用后声明，也就是变量需要先声明再使用。
var y=10;


//学习js对象
let person={
    name:"penghongbin",
    age:20,
    sex:"男",
    getage : function(){
        return this.age;
    },
    getname: function(){
        return this.name;
    }
}
console.log("person对象的name属性为："+person.name);
console.log("person对象的age属性为："+person.getage());
