package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func recordclientip(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("收到请求：%s\n", r.RemoteAddr+r.URL.RequestURI())
	rsp := "你好啊：" + r.RemoteAddr + r.URL.RequestURI()
	fmt.Printf("响应内容：%s\n", rsp)
	w.Write([]byte(rsp))

}
func returnHtml(w http.ResponseWriter, r *http.Request) {
	rsp := `<!DOCTYPE html>
	<html>
	<head>
		<title>HTML Response</title>
	</head>
	<body>
		<h1>Hello, World!</h1>
		<p>This is an HTML response.</p>
	</body>
	</html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(rsp))
}
func receiveJson(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("收到请求：%s\n", r.RemoteAddr+r.URL.RequestURI())
	// r.
}
func receiveQuery(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("收到请求：%s\n", r.RemoteAddr+r.URL.RequestURI())
	// r.URL.Query()["username"]
	fmt.Printf("用户名：%s\n", r.URL.Query().Get("username"))
	fmt.Printf("密码：%s\n", r.URL.Query().Get("password"))
}
func receiveForm(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("收到请求：%s\n", r.RemoteAddr+r.URL.RequestURI())
	err := r.ParseForm()
	if err != nil {
		fmt.Printf("解析表单失败：%v\n", err)
	}
	fmt.Printf("查询参数：%s\n", r.URL.Query().Get("username"))
	fmt.Printf("查询参数：%s\n", r.URL.Query().Get("password"))
	time.Sleep(5 * time.Second)
	fmt.Printf("用户名：%s\n", r.Form.Get("username"))
	fmt.Printf("密码：%s\n", r.Form.Get("password"))
}
func reveiveMultipartForm(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // 10 MB//这个函数会将请求体中的 multipart/form-data 数据解析到 r.MultipartForm 中，最大内存限制为 10 MB。超过这个限制的部分会被存储在临时文件中。
	file, header, err := r.FormFile("file")
	if err != nil {
		fmt.Printf("解析文件失败：%v\n", err)
	}
	fmt.Printf("文件名：%s\n", header.Filename)
	fmt.Printf("文件大小：%dKB\n", header.Size>>10)
	curfile, err := os.Create(header.Filename)
	if err != nil {
		fmt.Printf("创建文件失败：%v\n", err)
	}
	_, err = io.Copy(curfile, file)
	if err != nil {
		fmt.Printf("复制文件失败：%v\n", err)
	}
	defer curfile.Close()
	defer file.Close()
	username := r.FormValue("username")
	fmt.Printf("用户名：%s\n", username)
	w.WriteHeader(http.StatusOK)
}

func main() {
	//创建服务器实例
	server := &http.Server{
		Addr:    ":8080",
		Handler: nil,
	}
	//使用默认的全局ServeMux注册路由
	// http.HandleFunc("/", recordclientip)
	// http.HandleFunc("/hello", returnHtml)
	// http.HandleFunc("/json", receiveJson)
	// http.HandleFunc("/form", receiveForm)
	// http.HandleFunc("/query", receiveQuery)
	// http.HandleFunc("/multipart", reveiveMultipartForm)
	//使用指定的ServeMux注册路由
	mux := http.NewServeMux()
	mux.HandleFunc("/", recordclientip)
	mux.HandleFunc("/hello", returnHtml)
	mux.HandleFunc("/json", receiveJson)
	mux.HandleFunc("/form", receiveForm)
	mux.HandleFunc("/query", receiveQuery)
	mux.HandleFunc("/multipart", reveiveMultipartForm)
	server.Handler = mux //将ServeMux注册到服务器实例中
	// 启动服务器（非阻塞）
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 创建可取消的context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听context取消信号
	go func() {
		<-ctx.Done()
		fmt.Println("收到关闭信号，正在优雅关闭...")
		// 优雅关闭服务器
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("强制关闭服务器: %v", err)
		}
		fmt.Println("服务器已优雅关闭")
	}()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	fmt.Println("收到终止信号")
	// 优雅关闭服务器
	cancel()
	time.Sleep(10 * time.Second) //等待服务器完成关闭
	fmt.Println("服务器已关闭")
}
