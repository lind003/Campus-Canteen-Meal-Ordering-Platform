package services

import (
	"campus-food-backend/internal/models"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"
)

type OrderService struct {
	db *sql.DB
}

// GetDB 获取数据库连接（供handler使用）
func (s *OrderService) GetDB() *sql.DB {
	return s.db
}

func NewOrderService(db *sql.DB) *OrderService {
	return &OrderService{db: db}
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(demanderID int, req models.CreateOrderRequest) (*models.Order, error) {
	fmt.Printf("=== CreateOrder 开始 ===\n")

	fmt.Printf("参数: demanderID=%d, canteenID=%d, merchantID=%d\n",
		demanderID, req.CanteenID, req.MerchantID)

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. 计算订单总金额
	var totalAmount float64
	for _, item := range req.OrderDetails {
		var dishPrice float64

		err := tx.QueryRow("SELECT price FROM dish WHERE dish_id = ?", item.DishID).Scan(&dishPrice)
		if err != nil {
			fmt.Printf("❌ 查询菜品价格错误: dish_id=%d, error=%v\n", item.DishID, err)
			return nil, fmt.Errorf("菜品不存在: %w", err)
		}
		fmt.Printf("✅ 菜品价格: dish_id=%d, price=%.2f\n", item.DishID, dishPrice)
		totalAmount += dishPrice * float64(item.Qty)
	}
	totalAmount += req.Tip
	fmt.Printf("订单总金额: %.2f (含小费: %.2f)\n", totalAmount, req.Tip)

	// 2. 解析时间字符串
	var fetchTime time.Time
	if req.FetchTime != "" {
		// 多种时间格式
		layouts := []string{
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006-01-02T15:04",
			"2006-01-02 15:04",
			time.RFC3339,
		}

		var parseErr error
		for _, layout := range layouts {
			fetchTime, parseErr = time.Parse(layout, req.FetchTime)
			if parseErr == nil {
				break
			}
		}

		if parseErr != nil {
			fmt.Printf("❌ 时间解析失败: %s, 错误: %v\n", req.FetchTime, parseErr)
			return nil, fmt.Errorf("时间格式错误: %s", req.FetchTime)
		}
	} else {
		// 默认取餐时间为1小时后
		fetchTime = time.Now().Add(time.Hour)
	}

	fmt.Printf("✅ 解析后的取餐时间: %v\n", fetchTime)

	// 3. 创建订单
	fmt.Printf("执行订单插入SQL...\n")
	result, err := tx.Exec(`
       INSERT INTO orders (
        demander_id, canteen_id, merchant_id, status, order_time, 
        fetch_time, tip, total_amount, delivery_address, contact_phone
    ) VALUES (?, ?, ?, 'pending', NOW(), ?, ?, ?, ?, ?)`,
		demanderID, req.CanteenID, req.MerchantID, fetchTime, req.Tip,
		totalAmount, req.DeliveryAddress, req.ContactPhone,
	)

	if err != nil {
		fmt.Printf("❌ 创建订单错误: %v\n", err)
		return nil, err
	}

	orderID, _ := result.LastInsertId()
	fmt.Printf("✅ 创建的订单ID: %d\n", orderID)

	// 创建订单明细
	for i, item := range req.OrderDetails {
		var dishPrice float64
		err := tx.QueryRow("SELECT price FROM dish WHERE dish_id = ?", item.DishID).Scan(&dishPrice)
		if err != nil {
			fmt.Printf("❌ 查询菜品价格错误: %v\n", err)
			return nil, err
		}

		fmt.Printf("插入订单明细 %d: order_id=%d, dish_id=%d, qty=%d, price=%.2f\n",
			i+1, orderID, item.DishID, item.Qty, dishPrice)

		sql := "INSERT INTO orderdetail (order_id, dish_id, qty, unit_price) VALUES (?, ?, ?, ?)"
		fmt.Printf("执行的SQL: %s\n", sql)
		fmt.Printf("参数: order_id=%d, dish_id=%d, qty=%d, unit_price=%.2f\n",
			orderID, item.DishID, item.Qty, dishPrice)

		_, err = tx.Exec(sql, orderID, item.DishID, item.Qty, dishPrice)

		if err != nil {
			fmt.Printf("❌ 创建订单明细错误: %v\n", err)
			fmt.Printf("完整错误信息: %+v\n", err)

			// 检查错误详细信息
			fmt.Printf("错误类型: %T\n", err)
			fmt.Printf("完整错误: %+v\n", err)

			// 如果是 MySQL 错误，显示详细信息
			if mysqlErr, ok := err.(*mysql.MySQLError); ok {
				fmt.Printf("MySQL错误代码: %d\n", mysqlErr.Number)
				fmt.Printf("MySQL错误信息: %s\n", mysqlErr.Message)
			} else {
				fmt.Printf("非MySQL错误: %v\n", err)
			}
			return nil, err
		}
		fmt.Printf("✅ 订单明细插入成功\n")
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		fmt.Printf("❌ 提交事务错误: %v\n", err)
		return nil, err
	}

	fmt.Printf("✅ 事务提交成功\n")

	// 返回创建的订单
	var order models.Order
	var dbFetchTime, finishTime *time.Time
	var tip, totalAmountDB *float64

	err = s.db.QueryRow(`
        SELECT 
            order_id, demander_id, runner_id, canteen_id, merchant_id, status,
            order_time, fetch_time, finish_time, tip, total_amount,
            delivery_address, contact_phone
        FROM orders WHERE order_id = ?`,
		orderID,
	).Scan(
		&order.OrderID, &order.DemanderID, &order.RunnerID, &order.CanteenID, &order.MerchantID,
		&order.Status, &order.OrderTime, &dbFetchTime, &finishTime, &tip, &totalAmountDB,
		&order.DeliveryAddress, &order.ContactPhone,
	)

	if err != nil {
		fmt.Printf("❌ 查询订单错误: %v\n", err)
		return nil, err
	}

	// 处理 NULL 值
	if dbFetchTime != nil {
		order.FetchTime = *dbFetchTime
	}
	if finishTime != nil {
		order.FinishTime = finishTime
	}
	if tip != nil {
		order.Tip = *tip
	}
	if totalAmountDB != nil {
		order.TotalAmount = *totalAmountDB
	}

	fmt.Printf("✅ 订单创建成功: order_id=%d, status=%s, total_amount=%.2f\n",
		order.OrderID, order.Status, order.TotalAmount)
	fmt.Printf("=== CreateOrder 结束 ===\n")
	return &order, nil
}

// TakeOrder 接单
func (s *OrderService) TakeOrder(orderID, runnerID int) error {
	result, err := s.db.Exec(`
        UPDATE orders 
        SET runner_id = ?, status = 'accepted' 
        WHERE order_id = ? AND status = 'pending'`,
		runnerID, orderID,
	)

	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("订单不存在或已被接单")
	}

	return nil
}

// GetOrders 获取订单列表
func (s *OrderService) GetOrders(userID int, status string) ([]models.Order, error) {
	log.Println("=== GetOrders 开始 ===")

	query := `
        SELECT
    	o.order_id, o.demander_id, o.runner_id, o.canteen_id, o.merchant_id,
    	o.status, o.order_time, o.fetch_time, o.finish_time, o.tip, o.total_amount,
    	o.delivery_address, o.contact_phone,

    	u.name  AS demander_name,
    	u.phone AS demander_phone,
		u.card_no AS demander_card,
    	r.name  AS runner_name,
    	r.phone AS runner_phone,
    	r.card_no AS runner_card,

    	c.canteen_name, m.merchant_name
		FROM orders o
		LEFT JOIN user u ON o.demander_id = u.user_id
		LEFT JOIN user r ON o.runner_id   = r.user_id      -- 新增
		LEFT JOIN canteen c ON o.canteen_id   = c.canteen_id
		LEFT JOIN merchant m ON o.merchant_id = m.merchant_id
		WHERE 1=1
    `

	var args []interface{}

	if status != "all" && status != "" {
		query += " AND o.status = ?"
		args = append(args, status)
	}

	query += " ORDER BY o.order_time DESC"

	log.Printf("执行查询: %s\n", query)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		log.Printf("❌ 查询错误: %v\n", err)
		return nil, fmt.Errorf("数据库查询失败: %v", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		var runnerID *int
		var fetchTime, finishTime *time.Time
		var tip, totalAmount *float64
		var demanderName, demanderPhone, demanderCard, runnerName, runnerPhone, runnerCard, canteenName, merchantName *string

		err := rows.Scan(
			&order.OrderID, &order.DemanderID, &runnerID, &order.CanteenID, &order.MerchantID,
			&order.Status, &order.OrderTime, &fetchTime, &finishTime,
			&tip, &totalAmount, &order.DeliveryAddress, &order.ContactPhone,
			&demanderName, &demanderPhone, &demanderCard,
			&runnerName, &runnerPhone, &runnerCard,
			&canteenName, &merchantName,
		)
		if err != nil {
			log.Printf("❌ 数据扫描错误: %v\n", err)
			return nil, fmt.Errorf("数据扫描失败: %v", err)
		}

		order.RunnerID = runnerID

		// 处理 NULL 时间
		if fetchTime != nil {
			order.FetchTime = *fetchTime
		}
		if finishTime != nil {
			order.FinishTime = finishTime
		}

		// 处理 NULL 金额
		if tip != nil {
			order.Tip = *tip
		}
		if totalAmount != nil {
			order.TotalAmount = *totalAmount
		}

		// 处理扩展字段
		if demanderName != nil {
			order.DemanderName = *demanderName
		}
		if demanderPhone != nil {
			order.DemanderPhone = *demanderPhone
		}
		if runnerName != nil {
			order.RunnerName = *runnerName
		}
		if canteenName != nil {
			order.CanteenName = *canteenName
		}
		if merchantName != nil {
			order.MerchantName = *merchantName
		}
		if demanderCard != nil {
			order.DemanderCard = *demanderCard
		}
		if runnerName != nil {
			order.RunnerName = *runnerName
		}
		if runnerPhone != nil {
			order.RunnerPhone = *runnerPhone
		}
		if runnerCard != nil {
			order.RunnerCard = *runnerCard
		}
		orders = append(orders, order)
		log.Printf("✅ 扫描成功: order_id=%d, status=%s, amount=%.2f\n",
			order.OrderID, order.Status, order.TotalAmount)
	}

	if err = rows.Err(); err != nil {
		log.Printf("❌ 行遍历错误: %v\n", err)
		return nil, fmt.Errorf("行遍历失败: %v", err)
	}

	log.Printf("✅ 成功获取 %d 个订单\n", len(orders))
	return orders, nil
}

// CompleteOrder 完成订单
func (s *OrderService) CompleteOrder(orderID, runnerID int) error {
	result, err := s.db.Exec(`
        UPDATE orders 
        SET status = 'completed', finish_time = NOW()
        WHERE order_id = ? AND runner_id = ? AND status = 'accepted'`,
		orderID, runnerID,
	)

	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("订单不存在或无法完成")
	}

	return nil
}

// GetOrderByID 根据ID获取订单
func (s *OrderService) GetOrderByID(orderID int) (*models.Order, error) {
	var order models.Order
	err := s.db.QueryRow(`
        SELECT 
            order_id, demander_id, runner_id, canteen_id, merchant_id, status,
            order_time, fetch_time, finish_time, tip, total_amount,
            delivery_address, contact_phone
        FROM orders WHERE order_id = ?`,
		orderID,
	).Scan(
		&order.OrderID, &order.DemanderID, &order.RunnerID, &order.CanteenID, &order.MerchantID,
		&order.Status, &order.OrderTime, &order.FetchTime, &order.FinishTime, &order.Tip,
		&order.TotalAmount, &order.DeliveryAddress, &order.ContactPhone,
	)

	if err != nil {
		return nil, err
	}

	return &order, nil
}

// GetOrderDetails 获取订单详情及明细
func (s *OrderService) GetOrderDetails(orderID int) (*models.OrderWithDetails, error) {
	// 获取订单基本信息
	var order models.Order
	var demanderName, canteenName, merchantName, merchantDesc, merchantHours *string
	var runnerID *int
	var runnerName *string

	err := s.db.QueryRow(`
		SELECT 
			o.order_id, o.demander_id, o.runner_id, o.canteen_id, o.merchant_id,
			o.status, o.order_time, o.fetch_time, o.finish_time, o.tip, o.total_amount,
			o.delivery_address, o.contact_phone,
			u.name as demander_name, u.phone as demander_phone,
			r.name as runner_name,
			c.canteen_name, 
			m.merchant_name, m.description as merchant_desc, m.business_hours as merchant_hours
		FROM orders o
		LEFT JOIN user u ON o.demander_id = u.user_id        
		LEFT JOIN user r ON o.runner_id = r.user_id          
		LEFT JOIN canteen c ON o.canteen_id = c.canteen_id   
		LEFT JOIN merchant m ON o.merchant_id = m.merchant_id 
		WHERE o.order_id = ?
	`, orderID).Scan(
		&order.OrderID, &order.DemanderID, &runnerID, &order.CanteenID, &order.MerchantID,
		&order.Status, &order.OrderTime, &order.FetchTime, &order.FinishTime,
		&order.Tip, &order.TotalAmount, &order.DeliveryAddress, &order.ContactPhone,
		&demanderName, &order.DemanderPhone,
		&runnerName,
		&canteenName, &merchantName, &merchantDesc, &merchantHours,
	)

	if err != nil {
		return nil, err
	}

	order.RunnerID = runnerID

	// 处理扩展字段
	if demanderName != nil {
		order.DemanderName = *demanderName
	}
	if canteenName != nil {
		order.CanteenName = *canteenName
	}
	if merchantName != nil {
		order.MerchantName = *merchantName
	}
	if merchantDesc != nil {
		order.MerchantDesc = *merchantDesc
	}
	if merchantHours != nil {
		order.MerchantHours = *merchantHours
	}
	if runnerName != nil {
		order.RunnerName = *runnerName
	}

	// 获取订单明细，包含菜品信息
	rows, err := s.db.Query(`
		SELECT 
			od.detail_id, od.order_id, od.dish_id, od.qty, od.unit_price,
			d.dish_name, d.price as original_price,
			d.description as dish_desc, d.category, d.spicy_level
		FROM orderdetail od
		JOIN dish d ON od.dish_id = d.dish_id
		WHERE od.order_id = ?
	`, orderID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orderDetails []models.OrderDetailWithDish
	for rows.Next() {
		var detail models.OrderDetailWithDish
		var dishDesc, category *string
		var spicyLevel *int

		err := rows.Scan(
			&detail.DetailID, &detail.OrderID, &detail.DishID, &detail.Qty, &detail.UnitPrice,
			&detail.DishName, &detail.OriginalPrice,
			&dishDesc, &category, &spicyLevel,
		)
		if err != nil {
			return nil, err
		}

		// 处理可能为NULL的字段
		if dishDesc != nil {
			detail.DishDesc = *dishDesc
		}
		if category != nil {
			detail.Category = *category
		}
		if spicyLevel != nil {
			detail.SpicyLevel = *spicyLevel
		}

		orderDetails = append(orderDetails, detail)
	}

	orderWithDetails := &models.OrderWithDetails{
		Order:   order,
		Details: orderDetails,
	}

	return orderWithDetails, nil
}

// UpdateOrderStatus 更新订单状态
func (s *OrderService) UpdateOrderStatus(userID int, req models.UpdateOrderStatusRequest) error {
	fmt.Printf("UpdateOrderStatus - 开始更新订单状态: userID=%d, orderID=%d, newStatus=%s\n",
		userID, req.OrderID, req.Status)

	// 检查订单是否存在
	var currentStatus string
	var demanderID int
	var runnerID *int
	var merchantID int

	err := s.db.QueryRow(`
        SELECT status, demander_id, runner_id, merchant_id 
        FROM orders WHERE order_id = ?
    `, req.OrderID).Scan(&currentStatus, &demanderID, &runnerID, &merchantID)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("订单不存在")
		}
		return err
	}

	fmt.Printf("UpdateOrderStatus - 当前状态: %s, 需求者: %d, 商家: %d\n",
		currentStatus, demanderID, merchantID)

	runnerIDValue := 0
	if runnerID != nil {
		runnerIDValue = *runnerID
		fmt.Printf("UpdateOrderStatus - 跑腿员: %d\n", runnerIDValue)
	}

	// 检查用户权限
	userRole, err := s.getUserRole(userID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败")
	}

	if !s.hasPermissionToUpdate(userID, userRole, demanderID, runnerIDValue, merchantID, currentStatus, req.Status) {
		return fmt.Errorf("您没有权限执行此操作")
	}

	// 检查状态流转是否合法
	if !s.isValidStatusTransition(currentStatus, req.Status, userRole) {
		return fmt.Errorf("不允许从 %s 状态变更为 %s 状态", currentStatus, req.Status)
	}

	// 特殊处理：如果是跑腿员接单，需要设置runner_id
	var result sql.Result
	if req.Status == models.OrderStatusAccepted && userRole == "runner" {
		result, err = s.db.Exec(`
            UPDATE orders SET runner_id = ?, status = ? 
            WHERE order_id = ? AND status = 'pending'
        `, userID, req.Status, req.OrderID)
	} else {
		result, err = s.db.Exec(`
            UPDATE orders SET status = ? WHERE order_id = ?
        `, req.Status, req.OrderID)
	}

	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("订单状态更新失败")
	}

	// 如果是完成订单，更新时间字段
	if req.Status == models.OrderStatusCompleted {
		_, err = s.db.Exec(`
            UPDATE orders SET finish_time = NOW() WHERE order_id = ?
        `, req.OrderID)
		if err != nil {
			return err
		}
	}

	fmt.Printf("UpdateOrderStatus - 订单状态更新成功: orderID=%d, %s -> %s\n",
		req.OrderID, currentStatus, req.Status)

	return nil
}

// 获取用户角色
func (s *OrderService) getUserRole(userID int) (string, error) {
	var roleID int
	err := s.db.QueryRow("SELECT role_id FROM user WHERE user_id = ?", userID).Scan(&roleID)
	if err != nil {
		return "", err
	}

	// 根据role_id返回角色名称
	switch roleID {
	case 1:
		return "admin", nil
	case 2:
		return "demander", nil
	case 3:
		return "runner", nil
	case 4:
		return "merchant", nil
	default:
		return "user", nil
	}
}

// 检查用户权限
func (s *OrderService) hasPermissionToUpdate(userID int, userRole string,
	demanderID, runnerID, merchantID int, currentStatus, newStatus string) bool {

	switch userRole {
	case "admin":
		return true // 管理员有所有权限
	case "demander":
		// 需求者只能取消自己的待接单订单
		return userID == demanderID &&
			currentStatus == models.OrderStatusPending &&
			newStatus == models.OrderStatusCancelled
	case "runner":
		// 跑腿员可以接单、完成订单等
		if newStatus == models.OrderStatusAccepted {
			// 接单：只有pending状态的订单可以接
			return currentStatus == models.OrderStatusPending
		} else if newStatus == models.OrderStatusCompleted {
			// 完成订单：只有自己接的订单可以完成
			return userID == runnerID && currentStatus == models.OrderStatusAccepted
		}
		return false
	case "merchant":
		// 商家可以确认订单准备完成
		return userID == merchantID && newStatus == models.OrderStatusProcessing
	default:
		return false
	}
}

// 检查状态流转是否合法
func (s *OrderService) isValidStatusTransition(oldStatus, newStatus, userRole string) bool {
	// 简化的状态流转规则
	transitions := map[string][]string{
		models.OrderStatusPending: {
			models.OrderStatusAccepted,  // 跑腿员接单
			models.OrderStatusCancelled, // 需求者取消
		},
		models.OrderStatusAccepted: {
			models.OrderStatusCompleted, // 跑腿员完成配送
			models.OrderStatusCancelled, // 管理员取消
		},
		models.OrderStatusCompleted: {}, // 已完成不能变更
		models.OrderStatusCancelled: {}, // 已取消不能变更
	}

	allowed, ok := transitions[oldStatus]
	if !ok {
		return false
	}

	for _, status := range allowed {
		if status == newStatus {
			return true
		}
	}

	return false
}

// CancelOrder 取消订单（单独接口，方便前端调用）
func (s *OrderService) CancelOrder(userID, orderID int, reason string) error {
	req := models.UpdateOrderStatusRequest{
		OrderID: orderID,
		Status:  models.OrderStatusCancelled,
		Reason:  reason,
	}
	return s.UpdateOrderStatus(userID, req)
}

// GetAvailableStatusOptions 获取可用的状态选项
func (s *OrderService) GetAvailableStatusOptions(userID, orderID int) ([]string, error) {
	// 获取订单当前状态
	var currentStatus string
	err := s.db.QueryRow("SELECT status FROM orders WHERE order_id = ?", orderID).Scan(&currentStatus)
	if err != nil {
		return nil, err
	}

	// 获取用户角色
	userRole, err := s.getUserRole(userID)
	if err != nil {
		return nil, err
	}

	// 简化的状态流转
	allTransitions := map[string][]string{
		models.OrderStatusPending: {
			models.OrderStatusAccepted, models.OrderStatusCancelled,
		},
		models.OrderStatusAccepted: {
			models.OrderStatusCompleted, models.OrderStatusCancelled,
		},
	}

	available := allTransitions[currentStatus]
	var filtered []string

	for _, status := range available {
		// 简单的权限检查
		if s.isStatusAllowedForRole(status, userRole, currentStatus, userID, orderID) {
			filtered = append(filtered, status)
		}
	}

	return filtered, nil
}

func (s *OrderService) isStatusAllowedForRole(status, userRole, currentStatus string,
	userID, orderID int) bool {

	// 获取订单信息用于权限检查
	var demanderID int
	var runnerID *int
	var merchantID int

	err := s.db.QueryRow(`
        SELECT demander_id, runner_id, merchant_id 
        FROM orders WHERE order_id = ?
    `, orderID).Scan(&demanderID, &runnerID, &merchantID)

	if err != nil {
		return false
	}

	runnerIDValue := 0
	if runnerID != nil {
		runnerIDValue = *runnerID
	}

	// 检查权限
	return s.hasPermissionToUpdate(userID, userRole, demanderID, runnerIDValue, merchantID, currentStatus, status)
}

// GetMyRunningOrders 获取当前用户进行中的订单
func (s *OrderService) GetMyRunningOrders(userID int) ([]models.Order, error) {
	// 获取用户角色
	userRole, err := s.getUserRole(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败")
	}

	var query string
	var args []interface{}

	if userRole == "demander" || userRole == "user" {
		// 需求者：查询自己发布的进行中订单
		query = `
			SELECT o.*, u.name as demander_name, u.phone as demander_phone,
				r.name as runner_name, c.canteen_name, m.merchant_name
			FROM orders o
			LEFT JOIN user u ON o.demander_id = u.user_id
			LEFT JOIN user r ON o.runner_id = r.user_id
			LEFT JOIN canteen c ON o.canteen_id = c.canteen_id
			LEFT JOIN merchant m ON o.merchant_id = m.merchant_id
			WHERE o.demander_id = ? AND o.status IN ('pending', 'accepted')
			ORDER BY o.order_time DESC
		`
		args = append(args, userID)
	} else if userRole == "runner" {
		// 跑腿员：查询自己接的进行中订单
		query = `
			SELECT o.*, u.name as demander_name, u.phone as demander_phone,
				r.name as runner_name, c.canteen_name, m.merchant_name
			FROM orders o
			LEFT JOIN user u ON o.demander_id = u.user_id
			LEFT JOIN user r ON o.runner_id = r.user_id
			LEFT JOIN canteen c ON o.canteen_id = c.canteen_id
			LEFT JOIN merchant m ON o.merchant_id = m.merchant_id
			WHERE o.runner_id = ? AND o.status IN ('accepted')
			ORDER BY o.order_time DESC
		`
		args = append(args, userID)
	} else {
		return nil, fmt.Errorf("当前角色无法查询进行中订单")
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		var runnerID *int
		var fetchTime, finishTime *time.Time
		var tip, totalAmount *float64
		var demanderName, demanderPhone, runnerName, canteenName, merchantName *string

		err := rows.Scan(
			&order.OrderID, &order.DemanderID, &runnerID, &order.CanteenID, &order.MerchantID,
			&order.Status, &order.OrderTime, &fetchTime, &finishTime,
			&tip, &totalAmount, &order.DeliveryAddress, &order.ContactPhone,
			&demanderName, &demanderPhone, &runnerName, &canteenName, &merchantName,
		)
		if err != nil {
			continue
		}

		order.RunnerID = runnerID

		if fetchTime != nil {
			order.FetchTime = *fetchTime
		}
		if finishTime != nil {
			order.FinishTime = finishTime
		}
		if tip != nil {
			order.Tip = *tip
		}
		if totalAmount != nil {
			order.TotalAmount = *totalAmount
		}
		if demanderName != nil {
			order.DemanderName = *demanderName
		}
		if demanderPhone != nil {
			order.DemanderPhone = *demanderPhone
		}
		if runnerName != nil {
			order.RunnerName = *runnerName
		}
		if canteenName != nil {
			order.CanteenName = *canteenName
		}
		if merchantName != nil {
			order.MerchantName = *merchantName
		}

		orders = append(orders, order)
	}

	return orders, nil
}

// GetAvailableOrders 获取可接订单（跑腿员用）
func (s *OrderService) GetAvailableOrders(userID int) ([]models.Order, error) {
	// 检查用户是否是跑腿员
	userRole, err := s.getUserRole(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败")
	}

	if userRole != "runner" {
		return nil, fmt.Errorf("只有跑腿员可以查看可接订单")
	}

	// 查询所有待接单的订单
	rows, err := s.db.Query(`
		SELECT o.*, u.name as demander_name, u.phone as demander_phone,
			c.canteen_name, m.merchant_name, m.description as merchant_desc
		FROM orders o
		LEFT JOIN user u ON o.demander_id = u.user_id
		LEFT JOIN canteen c ON o.canteen_id = c.canteen_id
		LEFT JOIN merchant m ON o.merchant_id = m.merchant_id
		WHERE o.status = 'pending'
		ORDER BY o.order_time DESC
	`)

	if err != nil {
		return nil, fmt.Errorf("查询订单失败: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		var runnerID *int
		var fetchTime, finishTime *time.Time
		var tip, totalAmount *float64
		var demanderName, demanderPhone, canteenName, merchantName, merchantDesc *string

		err := rows.Scan(
			&order.OrderID, &order.DemanderID, &runnerID, &order.CanteenID, &order.MerchantID,
			&order.Status, &order.OrderTime, &fetchTime, &finishTime,
			&tip, &totalAmount, &order.DeliveryAddress, &order.ContactPhone,
			&demanderName, &demanderPhone, &canteenName, &merchantName, &merchantDesc,
		)
		if err != nil {
			continue
		}

		order.RunnerID = runnerID

		if fetchTime != nil {
			order.FetchTime = *fetchTime
		}
		if finishTime != nil {
			order.FinishTime = finishTime
		}
		if tip != nil {
			order.Tip = *tip
		}
		if totalAmount != nil {
			order.TotalAmount = *totalAmount
		}
		if demanderName != nil {
			order.DemanderName = *demanderName
		}
		if demanderPhone != nil {
			order.DemanderPhone = *demanderPhone
		}
		if canteenName != nil {
			order.CanteenName = *canteenName
		}
		if merchantName != nil {
			order.MerchantName = *merchantName
		}
		if merchantDesc != nil {
			order.MerchantDesc = *merchantDesc
		}

		orders = append(orders, order)
	}

	return orders, nil
}

// GetOrderStats 获取订单统计
func (s *OrderService) GetOrderStats(userID int) (map[string]interface{}, error) {
	userRole, err := s.getUserRole(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败")
	}

	var query string
	var args []interface{}

	stats := make(map[string]interface{})

	if userRole == "demander" {
		// 需求者统计
		query = `
			SELECT 
				COUNT(*) as total_orders,
				SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending_orders,
				SUM(CASE WHEN status = 'accepted' THEN 1 ELSE 0 END) as accepted_orders,
				SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed_orders,
				SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END) as cancelled_orders,
				SUM(CASE WHEN status = 'completed' THEN total_amount ELSE 0 END) as total_spent,
				AVG(CASE WHEN status = 'completed' THEN tip ELSE NULL END) as average_tip
			FROM orders WHERE demander_id = ?
		`
		args = append(args, userID)
	} else if userRole == "runner" {
		// 跑腿员统计
		query = `
			SELECT 
				COUNT(*) as total_orders,
				SUM(CASE WHEN status = 'accepted' THEN 1 ELSE 0 END) as accepted_orders,
				SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed_orders,
				SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END) as cancelled_orders,
				SUM(CASE WHEN status = 'completed' THEN tip ELSE 0 END) as total_earned,
				AVG(CASE WHEN status = 'completed' THEN tip ELSE NULL END) as average_tip
			FROM orders WHERE runner_id = ?
		`
		args = append(args, userID)
	} else if userRole == "admin" {
		// 管理员统计所有订单
		query = `
			SELECT 
				COUNT(*) as total_orders,
				SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending_orders,
				SUM(CASE WHEN status = 'accepted' THEN 1 ELSE 0 END) as accepted_orders,
				SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed_orders,
				SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END) as cancelled_orders
			FROM orders
		`
	} else {
		return nil, fmt.Errorf("当前角色无法查看统计信息")
	}

	// 执行查询
	var total, pending, accepted, completed, cancelled int
	var totalSpent, totalEarned, avgTip float64

	if userRole == "demander" {
		err = s.db.QueryRow(query, args...).Scan(
			&total, &pending, &accepted, &completed, &cancelled, &totalSpent, &avgTip,
		)
		if err == nil {
			stats["total_orders"] = total
			stats["pending_orders"] = pending
			stats["accepted_orders"] = accepted
			stats["completed_orders"] = completed
			stats["cancelled_orders"] = cancelled
			stats["total_spent"] = totalSpent
			stats["average_tip"] = avgTip
		}
	} else if userRole == "runner" {
		err = s.db.QueryRow(query, args...).Scan(
			&total, &accepted, &completed, &cancelled, &totalEarned, &avgTip,
		)
		if err == nil {
			stats["total_orders"] = total
			stats["accepted_orders"] = accepted
			stats["completed_orders"] = completed
			stats["cancelled_orders"] = cancelled
			stats["total_earned"] = totalEarned
			stats["average_tip"] = avgTip
		}
	} else if userRole == "admin" {
		err = s.db.QueryRow(query).Scan(&total, &pending, &accepted, &completed, &cancelled)
		if err == nil {
			stats["total_orders"] = total
			stats["pending_orders"] = pending
			stats["accepted_orders"] = accepted
			stats["completed_orders"] = completed
			stats["cancelled_orders"] = cancelled
		}
	}

	if err != nil {
		return nil, fmt.Errorf("获取统计信息失败: %w", err)
	}

	return stats, nil
}
