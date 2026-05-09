// 小后 Day1练习 #2: JSON处理
// Go语言基础 - JSON序列化与反序列化
package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// ==================== 结构体定义 ====================

// Product 商品
type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Price       float64   `json:"price"`
	Category    string    `json:"category,omitempty"`
	Stock       int       `json:"stock"`
	Tags        []string  `json:"tags"`
	IsAvailable bool      `json:"is_available"`
	CreatedAt   time.Time `json:"created_at"`
}

// Order 订单
type Order struct {
	OrderID   string            `json:"order_id"`
	UserID    int               `json:"user_id"`
	Products  []OrderItem       `json:"products"`
	TotalPrice float64          `json:"total_price"`
	Status    string            `json:"status"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

// OrderItem 订单项
type OrderItem struct {
	ProductID int     `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// APIResponse API响应
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ==================== JSON操作函数 ====================

// marshalPretty 格式化JSON输出
func marshalPretty(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("JSON错误: %v", err)
	}
	return string(data)
}

// marshalCompact 压缩JSON输出
func marshalCompact(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("JSON错误: %v", err)
	}
	return string(data)
}

// unmarshalJSON 反序列化
func unmarshalJSON(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}

// ==================== 演示函数 ====================

func main() {
	fmt.Println("=== 小后 Go语言学习 Day1 - JSON处理 ===")
	
	// 1. 结构体转JSON
	fmt.Println("\n--- 1. 结构体转JSON ---")
	product := Product{
		ID:          1,
		Name:        "机械键盘",
		Price:       299.00,
		Category:    "外设",
		Stock:       100,
		Tags:        []string{"机械", "RGB", "青轴"},
		IsAvailable: true,
		CreatedAt:   time.Now(),
	}
	
	jsonStr := marshalPretty(product)
	fmt.Println("商品JSON:")
	fmt.Println(jsonStr)
	
	// 2. JSON转结构体
	fmt.Println("\n--- 2. JSON转结构体 ---")
	jsonData := `{
  "id": 2,
  "name": "无线鼠标",
  "price": 129.50,
  "category": "外设",
  "stock": 50,
  "tags": ["无线", "静音"],
  "is_available": true
}`
	
	var mouse Product
	if err := unmarshalJSON(jsonData, &mouse); err != nil {
		fmt.Printf("解析错误: %v\n", err)
	} else {
		fmt.Printf("解析成功: ID=%d, Name=%s, Price=%.2f\n", 
			mouse.ID, mouse.Name, mouse.Price)
	}
	
	// 3. 嵌套结构体
	fmt.Println("\n--- 3. 嵌套结构体 ---")
	order := Order{
		OrderID:    "ORD20260509001",
		UserID:     100,
		Products: []OrderItem{
			{ProductID: 1, Name: "机械键盘", Quantity: 1, Price: 299.00},
			{ProductID: 2, Name: "无线鼠标", Quantity: 2, Price: 259.00},
		},
		TotalPrice: 558.00,
		Status:    "pending",
		Extra: map[string]interface{}{
			"shipping_address": "北京市朝阳区",
			"remark":            "请小心轻放",
		},
	}
	
	orderJSON := marshalPretty(order)
	fmt.Println("订单JSON:")
	fmt.Println(orderJSON)
	
	// 4. API响应封装
	fmt.Println("\n--- 4. API响应封装 ---")
	response := APIResponse{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"total_users":  1000,
			"active_users": 856,
		},
	}
	fmt.Println("API响应:")
	fmt.Println(marshalPretty(response))
	
	// 5. 动态JSON处理
	fmt.Println("\n--- 5. 动态JSON (map) ---")
	var dynamicData map[string]interface{}
	dynamicJSON := `{
  "type": "notification",
  "content": {
    "title": "系统通知",
    "body": "这是一条通知内容",
    "priority": 1
  },
  "timestamp": 1715241600
}`
	
	if err := json.Unmarshal([]byte(dynamicJSON), &dynamicData); err != nil {
		fmt.Printf("解析错误: %v\n", err)
	} else {
		fmt.Printf("类型: %v\n", dynamicData["type"])
		content := dynamicData["content"].(map[string]interface{})
		fmt.Printf("标题: %v\n", content["title"])
	}
	
	// 6. 压缩输出
	fmt.Println("\n--- 6. 压缩JSON输出 ---")
	fmt.Printf("压缩后: %s\n", marshalCompact(product))
	
	fmt.Println("\n=== Day1练习2完成 ===")
}
