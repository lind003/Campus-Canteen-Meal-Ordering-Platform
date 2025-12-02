package models

import "time"

type Complaint struct {
	ComplaintID   int        `json:"complaint_id" db:"complaint_id"`
	OrderID       int        `json:"order_id" db:"order_id"`
	ComplainerID  int        `json:"complainer_id" db:"complainter_id"` //数据库是complainter
	AdminID       *int       `json:"admin_id" db:"admin_id"`
	Type          string     `json:"type" db:"type"`
	Reason        string     `json:"reason" db:"reason"`
	AdminFeedback *string    `json:"admin_feedback" db:"admin_feedback"`
	Status        string     `json:"status" db:"status"`
	HandleTime    *time.Time `json:"handle_time" db:"handle_time"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

type CreateComplaintRequest struct {
	OrderID int    `json:"order_id" binding:"required"`
	Type    string `json:"type" binding:"required"`
	Reason  string `json:"reason" binding:"required"`
}
