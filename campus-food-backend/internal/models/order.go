package models

import "time"

type Order struct {
	OrderID         int        `json:"order_id" db:"order_id"`
	DemanderID      int        `json:"demander_id" db:"demander_id"`
	RunnerID        *int       `json:"runner_id" db:"runner_id"`
	CanteenID       int        `json:"canteen_id" db:"canteen_id"`
	MerchantID      int        `json:"merchant_id" db:"merchant_id"`
	Status          string     `json:"status" db:"status"`
	OrderTime       time.Time  `json:"order_time" db:"order_time"`
	FetchTime       time.Time  `json:"fetch_time" db:"fetch_time"`
	FinishTime      *time.Time `json:"finish_time" db:"finish_time"`
	Tip             float64    `json:"tip" db:"tip"`
	TotalAmount     float64    `json:"total_amount" db:"total_amount"`
	DeliveryAddress string     `json:"delivery_address" db:"delivery_address"`
	ContactPhone    string     `json:"contact_phone" db:"contact_phone"`
	MerchantDesc    string     `json:"merchant_desc,omitempty" db:"merchant_desc"`   //商家描述
	MerchantHours   string     `json:"merchant_hours,omitempty" db:"business_hours"` // 营业时间

	// 扩展字段（用于查询结果）
	DemanderName  string `json:"demander_name,omitempty"`
	DemanderPhone string `json:"demander_phone,omitempty"`
	RunnerName    string `json:"runner_name,omitempty"`
	CanteenName   string `json:"canteen_name,omitempty"`
	MerchantName  string `json:"merchant_name,omitempty"`
}

type OrderDetail struct {
	DetailID  int     `json:"detail_id" db:"detail_id"`
	OrderID   int     `json:"order_id" db:"order_id"`
	DishID    int     `json:"dish_id" db:"dish_id"`
	Qty       int     `json:"qty" db:"qty"`
	UnitPrice float64 `json:"unit_price" db:"unit_price"`
}

// OrderDetailWithDish 订单明细包含菜品信息
type OrderDetailWithDish struct {
	DetailID      int     `json:"detail_id" db:"detail_id"`
	OrderID       int     `json:"order_id" db:"order_id"`
	DishID        int     `json:"dish_id" db:"dish_id"`
	Qty           int     `json:"qty" db:"qty"`
	UnitPrice     float64 `json:"unit_price" db:"unit_price"`
	DishName      string  `json:"dish_name" db:"dish_name"`
	OriginalPrice float64 `json:"original_price" db:"original_price"`
	DishDesc      string  `json:"dish_desc,omitempty" db:"dish_desc"` // 新增：菜品描述
	Category      string  `json:"category,omitempty" db:"category"`   // 新增：菜品类别
	SpicyLevel    int     `json:"spicy_level" db:"spicy_level"`       // 新增：辣度
}

// OrderWithDetails 订单包含详情信息
type OrderWithDetails struct {
	Order   Order                 `json:"order"`
	Details []OrderDetailWithDish `json:"details"`
}

type CreateOrderRequest struct {
	CanteenID       int              `json:"canteen_id" binding:"required"`
	MerchantID      int              `json:"merchant_id" binding:"required"`
	Tip             float64          `json:"tip"`
	FetchTime       string           `json:"fetch_time" binding:"required"` // 改为字符串，前端处理格式
	DeliveryAddress string           `json:"delivery_address" binding:"required"`
	ContactPhone    string           `json:"contact_phone" binding:"required"`
	OrderDetails    []OrderDetailReq `json:"order_details" binding:"required"`
}

type OrderDetailReq struct {
	DishID int `json:"dish_id" binding:"required"`
	Qty    int `json:"qty" binding:"required,min=1"`
}

// 更新订单状态请求
type UpdateOrderStatusRequest struct {
	OrderID int    `json:"order_id" binding:"required"`
	Status  string `json:"status" binding:"required"`
	Reason  string `json:"reason,omitempty"` // 状态变更原因
}

/*// 订单状态历史
type OrderStatusHistory struct {
	HistoryID    int       `json:"history_id" db:"history_id"`
	OrderID      int       `json:"order_id" db:"order_id"`
	OldStatus    string    `json:"old_status" db:"old_status"`
	NewStatus    string    `json:"new_status" db:"new_status"`
	ChangedBy    int       `json:"changed_by" db:"changed_by"`
	ChangeReason string    `json:"change_reason,omitempty" db:"change_reason"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}*/

// 菜品信息
type Dish struct {
	DishID      int     `json:"dish_id" db:"dish_id"`
	MerchantID  int     `json:"merchant_id" db:"merchant_id"`
	DishName    string  `json:"dish_name" db:"dish_name"`
	Price       float64 `json:"price" db:"price"`
	Category    string  `json:"category" db:"category"`
	Description string  `json:"description" db:"description"`
	ImageURL    string  `json:"image_url" db:"image_url"`
	IsAvailable bool    `json:"is_available" db:"is_available"`
	SpicyLevel  int     `json:"spicy_level" db:"spicy_level"`
	CreatedAt   string  `json:"created_at" db:"created_at"`
}

// ：商家信息结构体
type Merchant struct {
	MerchantID    int    `json:"merchant_id" db:"merchant_id"`
	CanteenID     int    `json:"canteen_id" db:"canteen_id"`
	MerchantName  string `json:"merchant_name" db:"merchant_name"`
	WindowNo      string `json:"window_no" db:"window_no"`
	Phone         string `json:"phone" db:"mer_phone"`
	Description   string `json:"description" db:"description"`
	BusinessHours string `json:"business_hours" db:"business_hours"`
	//Address       string `json:"address" db:"address"`
}

// ：订单查询参数
type OrderQueryParams struct {
	Status   string `form:"status"`
	UserID   int    `form:"user_id"`
	Role     string `form:"role"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// 订单状态常量
const (
	OrderStatusPending    = "pending"    // 待接单
	OrderStatusAccepted   = "accepted"   // 已接单
	OrderStatusProcessing = "processing" // 处理中
	OrderStatusCompleted  = "completed"  // 已完成
	OrderStatusCancelled  = "cancelled"  // 已取消
)
