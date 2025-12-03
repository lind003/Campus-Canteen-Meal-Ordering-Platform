package models

import (
	"time"
)

type User struct {
	UserID    int       `json:"user_id" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	College   string    `json:"college" db:"college"`
	Grade     string    `json:"grade" db:"grade"`
	Phone     string    `json:"phone" db:"phone"`
	Password  string    `json:"-" db:"pwd"` // 不序列化到JSON
	Role      string    `json:"role" db:"role"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	CardNo    string    `json:"card_no" db:"card_no"`
}

type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	College  string `json:"college" binding:"required"`
	Grade    string `json:"grade" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required"`
	CardNo   string `json:"card_no"` // 校园卡号
}

// 修改资料
type UpdateProfileRequest struct {
	Name     string `json:"name" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	College  string `json:"college"`
	Grade    string `json:"grade"`
	CardNo   string `json:"card_no"`
	Password string `json:"password" binding:"required"` // 当前密码确认
}

// 修改密码
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
