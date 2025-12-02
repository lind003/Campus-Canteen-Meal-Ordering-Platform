package services

import (
	"campus-food-backend/internal/models"
	"database/sql"
	"fmt"
	"time"
)

type ComplaintService struct {
	db *sql.DB
}

func NewComplaintService(db *sql.DB) *ComplaintService {
	return &ComplaintService{db: db}
}

func (s *ComplaintService) CreateComplaint(complainerID int, req models.CreateComplaintRequest) (*models.Complaint, error) {
	fmt.Printf("CreateComplaint - 开始处理投诉: complainerID=%d, orderID=%d, type=%s\n",
		complainerID, req.OrderID, req.Type)

	// 检查订单是否存在且属于该用户
	var orderDemanderID int
	err := s.db.QueryRow("SELECT demander_id FROM orders WHERE order_id = ?", req.OrderID).Scan(&orderDemanderID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("CreateComplaint - 订单不存在: orderID=%d\n", req.OrderID)
			return nil, fmt.Errorf("订单不存在")
		}
		fmt.Printf("CreateComplaint - 查询订单错误: %v\n", err)
		return nil, err
	}

	fmt.Printf("CreateComplaint - 订单需求者ID: %d, 投诉者ID: %d\n", orderDemanderID, complainerID)

	// 检查是否是该用户的订单
	if orderDemanderID != complainerID {
		fmt.Printf("CreateComplaint - 权限错误: 用户 %d 试图投诉订单 %d (属于用户 %d)\n",
			complainerID, req.OrderID, orderDemanderID)
		return nil, fmt.Errorf("只能投诉自己的订单")
	}

	// 检查是否已经投诉过该订单
	var existingComplaint int
	s.db.QueryRow("SELECT COUNT(*) FROM complaint WHERE order_id = ? AND complainter_id = ?", req.OrderID, complainerID).Scan(&existingComplaint)
	if existingComplaint > 0 {
		fmt.Printf("CreateComplaint - 重复投诉: orderID=%d, complainerID=%d\n", req.OrderID, complainerID)
		return nil, fmt.Errorf("已经投诉过该订单")
	}

	fmt.Printf("CreateComplaint - 创建投诉记录\n")

	//管理员ID
	adminID := 4

	// 创建投诉
	result, err := s.db.Exec(`
        INSERT INTO complaint 
        (order_id, complainter_id, admin_id, type, reason)
        VALUES (?, ?, ?, ?, ?)
    `, req.OrderID, complainerID, adminID, req.Type, req.Reason)

	if err != nil {
		fmt.Printf("CreateComplaint - 插入投诉记录错误: %v\n", err)
		return nil, err
	}

	complaintID, _ := result.LastInsertId()

	// 返回创建的投诉
	var complaint models.Complaint
	err = s.db.QueryRow(`
        SELECT complaint_id, order_id, complainter_id, admin_id, type, reason, 
               admin_feedback, handle_time
        FROM complaint WHERE complaint_id = ?
    `, complaintID).Scan(
		&complaint.ComplaintID, &complaint.OrderID, &complaint.ComplainerID, &complaint.AdminID,
		&complaint.Type, &complaint.Reason, &complaint.AdminFeedback, &complaint.HandleTime,
	)

	if err != nil {
		fmt.Printf("CreateComplaint - 查询投诉记录错误: %v\n", err)
		return nil, err
	}

	fmt.Printf("CreateComplaint - 投诉创建成功: complaintID=%d\n", complaint.ComplaintID)
	return &complaint, nil
}

func (s *ComplaintService) GetComplaints() ([]models.Complaint, error) {
	rows, err := s.db.Query(`
        SELECT c.complaint_id, c.order_id, c.complainter_id, c.admin_id, c.type, c.reason,
               c.admin_feedback, c.status, c.handle_time, c.created_at,
               u.name as complainer_name, u.phone as complainer_phone,
               a.name as admin_name, o.order_time
        FROM Complaint c
        LEFT JOIN User u ON c.complainter_id = u.user_id
        LEFT JOIN User a ON c.admin_id = a.user_id
        LEFT JOIN Orders o ON c.order_id = o.order_id
        ORDER BY c.created_at DESC
    `)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var complaints []models.Complaint
	for rows.Next() {
		var complaint models.Complaint
		var complainerName, complainerPhone string
		var adminName *string
		var orderTime time.Time

		err := rows.Scan(
			&complaint.ComplaintID, &complaint.OrderID, &complaint.ComplainerID, &complaint.AdminID,
			&complaint.Type, &complaint.Reason, &complaint.AdminFeedback, &complaint.Status,
			&complaint.HandleTime, &complaint.CreatedAt,
			&complainerName, &complainerPhone, &adminName, &orderTime,
		)
		if err != nil {
			return nil, err
		}

		complaints = append(complaints, complaint)
	}

	return complaints, nil
}

func (s *ComplaintService) HandleComplaint(complaintID, adminID int, feedback string) error {
	result, err := s.db.Exec(`
        UPDATE Complaint 
        SET admin_id = ?, admin_feedback = ?, status = 'resolved', handle_time = NOW()
        WHERE complaint_id = ? AND status = 'pending'
    `, adminID, feedback, complaintID)

	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("投诉不存在或已被处理")
	}

	return nil
}
