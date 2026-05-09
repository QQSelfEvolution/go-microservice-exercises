// 小后 Day1练习 #1: Go net/http
// Go语言基础 - HTTP服务
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Response 响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// User 用户结构
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// homeHandler 首页
func homeHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head><title>小后的Go HTTP服务</title></head>
<body>
<h1>欢迎来到小后的Go服务器</h1>
<p>API端点:</p>
<ul>
<li><a href="/api/user/1">GET /api/user/:id</a></li>
<li><a href="/api/users">GET /api/users</a></li>
<li><a href="/health">健康检查</a></li>
</ul>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

// userHandler 获取单个用户
func userHandler(w http.ResponseWriter, r *http.Request) {
	// 模拟用户数据
	user := User{
		ID:    1,
		Name:  "小后",
		Email: "xiaohou@example.com",
	}
	
	resp := Response{
		Code:    0,
		Message: "success",
		Data:    user,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// usersHandler 获取用户列表
func usersHandler(w http.ResponseWriter, r *http.Request) {
	users := []User{
		{ID: 1, Name: "阿代码", Email: "acode@example.com"},
		{ID: 2, Name: "小匠", Email: "xiaojiang@example.com"},
		{ID: 3, Name: "小龙", Email: "xiaolong@example.com"},
		{ID: 4, Name: "小后", Email: "xiaohou@example.com"},
	}
	
	resp := Response{
		Code:    0,
		Message: "success",
		Data:    users,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// healthHandler 健康检查
func healthHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"uptime":    time.Since(startTime).String(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// timeHandler 时间服务
func timeHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"unix":    time.Now().Unix(),
		"iso":     time.Now().Format(time.RFC3339),
		"timezone": "Asia/Shanghai",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

var startTime = time.Now()

func main() {
	fmt.Println("=== 小后 Go语言学习 Day1 - HTTP服务 ===")
	
	// 注册路由
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/user/", userHandler)
	http.HandleFunc("/api/users", usersHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/time", timeHandler)
	
	port := ":8080
	fmt.Printf("🌐 服务器启动: http://localhost%s\n", port)
	fmt.Println("按 Ctrl+C 停止服务器")
	fmt.Println()
	fmt.Println("可用的端点:")
	fmt.Println("  GET /            - 首页")
	fmt.Println("  GET /api/user/1  - 获取用户")
	fmt.Println("  GET /api/users   - 用户列表")
	fmt.Println("  GET /health     - 健康检查")
	fmt.Println("  GET /time        - 当前时间")
	
	// 启动服务器
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
