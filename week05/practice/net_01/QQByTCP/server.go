package main

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

var (
	connNum   = make(chan struct{}, 100) //限制TCP客户端的连接数量
	user2user = make(map[string]net.Conn)
	userMu    sync.RWMutex
)

func main() {
	CreateSafeServer()
}
func CreateSafeServer() {
	tcplistener, err := net.ListenTCP("tcp", &net.TCPAddr{Port: 8000})
	if err != nil {
		fmt.Println("发生错误：", err)
	}
	fmt.Println("服务端已启动成功")
	for {
		conn, err := tcplistener.Accept()
		if err != nil {
			fmt.Println("发生错误：", err)
			continue
		}
		select {
		case connNum <- struct{}{}:
			fmt.Printf("客户端地址%v新连接已建立\n", conn.RemoteAddr())
			go handleConn(conn)
		default:
			fmt.Println("连接数已满，拒绝新连接")
			conn.Close()
		}
	}
}
func PrintAllOnlineUser(conn net.Conn) {
	conn.Write([]byte("当前在线用户：\n"))
	userMu.RLock()
	defer userMu.RUnlock()
	for userName, _ := range user2user {
		conn.Write([]byte("用户名：" + userName + "\t"))
	}
	conn.Write([]byte("\n"))
}
func handleConn(conn net.Conn) {
	var userName string
	var message = make([]byte, 1024)
	for {
		PrintAllOnlineUser(conn)
		conn.Write([]byte("请输入你的用户名：\n"))
		n, err := conn.Read(message)
		if err != nil {
			fmt.Println("发生错误：", err)
			continue
		}
		userName = string(message[:n])
		if userName == "" {
			conn.Write([]byte("用户名不能为空\n"))
			continue
		} else {
			userMu.Lock()
			if user2user[userName] != nil {
				conn.Write([]byte("用户名已存在\n"))
				userMu.Unlock()
				continue
			}
			user2user[userName] = conn
			userMu.Unlock()
			conn.Write([]byte("用户名注册成功\n")) //这个最好不要放在锁里面，不然如果阻塞，会导致锁长时间无法释放
			break
		}
	}
	var targetUserName string
	var err error
	targetUserName, err = choseUser(userName, conn)
	if err != nil {
		conn.Write([]byte(err.Error()))
		return
	}
	clear(message)
	for {
		userMu.RLock()
		var targetUserConn net.Conn = user2user[targetUserName]
		userMu.RUnlock()
		if targetUserConn == nil {
			conn.Write([]byte("目标用户已下线\n"))
			return
		}
		clear(message)
		conn.Write([]byte("请输入你想对目标用户" + targetUserName + "发送的消息：\n"))
		n, err := conn.Read(message)
		if err != nil {
			conn.Write([]byte("发生错误：" + err.Error()))
			continue
		}
		targetUserConn, err = checkOnline(targetUserName)
		if err != nil {
			conn.Write([]byte(err.Error()))
			break
		}
		nWrite, err := targetUserConn.Write(message[:n])
		if err != nil {
			conn.Write([]byte("发生错误：" + err.Error()))
			return
		}
		conn.Write([]byte("发送成功:" + string(message[:nWrite]) + "\n"))
	}
	conn.Close()
	fmt.Printf("客户端地址%v已断开连接\n", conn.RemoteAddr())
	userMu.Lock()
	delete(user2user, userName)
	userMu.Unlock()
	<-connNum
}
func checkOnline(userName string) (net.Conn, error) {
	userMu.RLock()
	defer userMu.RUnlock()
	if user2user[userName] == nil {
		return nil, errors.New("目标用户已下线")
	}
	return user2user[userName], nil
}
func choseUser(originUser string, conn net.Conn) (string, error) {
	var userName string
	var message = make([]byte, 1024)
	for {
		PrintAllOnlineUser(conn)
		conn.Write([]byte("请输入你想对哪个用户建立连接后发送消息：\n"))
		n, err := conn.Read(message)
		if err != nil {
			return "", errors.New("读取用户输入失败")
		}
		userName = string(message[:n])
		if userName == "" || userName == originUser {
			conn.Write([]byte("用户名不能为空或不能与自己建立连接\n"))
			continue
		} else {
			userMu.RLock()
			if user2user[userName] == nil {
				conn.Write([]byte("目标用户不存在或未注册\n"))
				userMu.RUnlock()
				continue
			}
			userMu.RUnlock()
			break
		}
	}
	return userName, nil
}
