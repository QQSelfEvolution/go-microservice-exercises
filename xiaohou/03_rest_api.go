// 小后 Day1练习 #3: REST API设计
// Go语言基础 - RESTful API设计
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ==================== 数据模型 ====================

// Task 任务
type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	Priority  int       `json:"priority"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ==================== 内存存储 ====================

type TaskStore struct {
	mu    sync.RWMutex
	tasks map[int]*Task
	nextID int
}

func NewTaskStore() *TaskStore {
	store := &TaskStore{
		tasks:  make(map[int]*Task),
		nextID: 1,
	}
	// 初始化一些示例数据
	store.tasks[1] = &Task{
		ID:        1,
		Title:     "学习Go语言",
		Content:   "完成REST API设计",
		Status:    "in_progress",
		Priority:  1,
		Tags:      []string{"学习", "Go"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	store.nextID = 2
	return store
}

// ==================== CRUD操作 ====================

// Create 创建任务
func (s *TaskStore) Create(task *Task) *Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	task.ID = s.nextID
	s.nextID++
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	
	s.tasks[task.ID] = task
	return task
}

// Get 获取单个任务
func (s *TaskStore) Get(id int) (*Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	task, exists := s.tasks[id]
	return task, exists
}

// List 获取所有任务
func (s *TaskStore) List() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// Update 更新任务
func (s *TaskStore) Update(id int, updates map[string]interface{}) (*Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	task, exists := s.tasks[id]
	if !exists {
		return nil, false
	}
	
	// 应用更新
	if title, ok := updates["title"]; ok {
		task.Title = title.(string)
	}
	if content, ok := updates["content"]; ok {
		task.Content = content.(string)
	}
	if status, ok := updates["status"]; ok {
		task.Status = status.(string)
	}
	if priority, ok := updates["priority"]; ok {
		task.Priority = int(priority.(float64))
	}
	if tags, ok := updates["tags"]; ok {
		task.Tags = tags.([]string)
	}
	task.UpdatedAt = time.Now()
	
	return task, true
}

// Delete 删除任务
func (s *TaskStore) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.tasks[id]; !exists {
		return false
	}
	delete(s.tasks, id)
	return true
}

// ==================== HTTP处理器 ====================

type TaskHandler struct {
	store *TaskStore
}

func NewTaskHandler(store *TaskStore) *TaskHandler {
	return &TaskHandler{store: store}
}

// writeJSON 写入JSON响应
func (h *TaskHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// CreateTask POST /tasks
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{
			Code:    405,
			Message: "只支持POST方法",
		})
		return
	}
	
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "读取请求失败",
		})
		return
	}
	
	var task Task
	if err := json.Unmarshal(body, &task); err != nil {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "无效的JSON格式",
		})
		return
	}
	
	created := h.store.Create(&task)
	h.writeJSON(w, http.StatusCreated, created)
}

// GetTask GET /tasks/:id
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{
			Code:    405,
			Message: "只支持GET方法",
		})
		return
	}
	
	// 解析ID
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "缺少任务ID",
		})
		return
	}
	
	var id int
	fmt.Sscanf(parts[2], "%d", &id)
	
	task, exists := h.store.Get(id)
	if !exists {
		h.writeJSON(w, http.StatusNotFound, ErrorResponse{
			Code:    404,
			Message: "任务不存在",
		})
		return
	}
	
	h.writeJSON(w, http.StatusOK, task)
}

// ListTasks GET /tasks
func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{
			Code:    405,
			Message: "只支持GET方法",
		})
		return
	}
	
	tasks := h.store.List()
	h.writeJSON(w, http.StatusOK, tasks)
}

// UpdateTask PUT /tasks/:id
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{
			Code:    405,
			Message: "只支持PUT方法",
		})
		return
	}
	
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "缺少任务ID",
		})
		return
	}
	
	var id int
	fmt.Sscanf(parts[2], "%d", &id)
	
	body, _ := io.ReadAll(r.Body)
	var updates map[string]interface{}
	json.Unmarshal(body, &updates)
	
	task, updated := h.store.Update(id, updates)
	if !updated {
		h.writeJSON(w, http.StatusNotFound, ErrorResponse{
			Code:    404,
			Message: "任务不存在",
		})
		return
	}
	
	h.writeJSON(w, http.StatusOK, task)
}

// DeleteTask DELETE /tasks/:id
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{
			Code:    405,
			Message: "只支持DELETE方法",
		})
		return
	}
	
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "缺少任务ID",
		})
		return
	}
	
	var id int
	fmt.Sscanf(parts[2], "%d", &id)
	
	if !h.store.Delete(id) {
		h.writeJSON(w, http.StatusNotFound, ErrorResponse{
			Code:    404,
			Message: "任务不存在",
		})
		return
	}
	
	h.writeJSON(w, http.StatusOK, map[string]string{"message": "删除成功"})
}

// ==================== 主函数 ====================

func main() {
	fmt.Println("=== 小后 Go语言学习 Day1 - REST API ===")
	
	store := NewTaskStore()
	handler := NewTaskHandler(store)
	
	// 注册路由
	http.HandleFunc("/tasks", handler.ListTasks)
	http.HandleFunc("/tasks/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetTask(w, r)
		case http.MethodPut:
			handler.UpdateTask(w, r)
		case http.MethodDelete:
			handler.DeleteTask(w, r)
		default:
			handler.writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{
				Code:    405,
				Message: "不支持的方法",
			})
		}
	})
	http.HandleFunc("/tasks", handler.CreateTask)
	
	port := ":8080
	fmt.Printf("REST API服务器启动: http://localhost%s\n", port)
	fmt.Println("\nAPI端点:")
	fmt.Println("  GET    /tasks       - 获取所有任务")
	fmt.Println("  POST   /tasks       - 创建任务")
	fmt.Println("  GET    /tasks/:id   - 获取单个任务")
	fmt.Println("  PUT    /tasks/:id   - 更新任务")
	fmt.Println("  DELETE /tasks/:id   - 删除任务")
	
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
