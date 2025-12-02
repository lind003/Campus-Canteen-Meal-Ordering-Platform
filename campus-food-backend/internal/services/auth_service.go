package services

import (
	"campus-food-backend/internal/models"
	"campus-food-backend/internal/utils"
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db *sql.DB
}

func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Register(req models.RegisterRequest) (*models.User, string, error) {
	// 检查手机号是否已存在
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM User WHERE phone = ?", req.Phone).Scan(&count)
	if err != nil {
		return nil, "", err
	}
	if count > 0 {
		return nil, "", errors.New("手机号已注册")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	// 创建用户
	result, err := s.db.Exec(`
        INSERT INTO User (name, college, grade, phone, pwd, role, status) 
        VALUES (?, ?, ?, ?, ?, ?, 'active')
    `, req.Name, req.College, req.Grade, req.Phone, string(hashedPassword), req.Role)

	if err != nil {
		return nil, "", err
	}

	userID, _ := result.LastInsertId()

	user := &models.User{
		UserID:  int(userID),
		Name:    req.Name,
		College: req.College,
		Grade:   req.Grade,
		Phone:   req.Phone,
		Role:    req.Role,
		Status:  "active",
	}

	// 生成令牌
	token, err := utils.GenerateJWTToken(user.UserID, user.Role)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) Login(req models.LoginRequest) (*models.User, string, error) {
	var user models.User
	err := s.db.QueryRow(`
        SELECT user_id, name, college, grade, phone, pwd, role, status 
        FROM User WHERE phone = ? AND status = 'active'
    `, req.Phone).Scan(
		&user.UserID, &user.Name, &user.College, &user.Grade,
		&user.Phone, &user.Password, &user.Role, &user.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", errors.New("用户不存在或账号被冻结")
		}
		return nil, "", err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, "", errors.New("密码错误")
	}

	// 清除密码字段
	user.Password = ""

	// 生成令牌
	token, err := utils.GenerateJWTToken(user.UserID, user.Role)
	if err != nil {
		return nil, "", err
	}

	return &user, token, nil
}

func (s *AuthService) Logout(token string) error {

	return nil
}
