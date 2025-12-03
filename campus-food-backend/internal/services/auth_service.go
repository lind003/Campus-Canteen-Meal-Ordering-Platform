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
        INSERT INTO User (name, college, grade, phone, pwd, role, status,card_no) 
        VALUES (?, ?, ?, ?, ?, ?, 'active',?)
    `, req.Name, req.College, req.Grade, req.Phone, string(hashedPassword), req.Role, req.CardNo)

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

// 查原始资料
func (s *AuthService) GetProfile(userID int) (*models.User, error) {
	var u models.User
	err := s.db.QueryRow(`SELECT user_id,name,phone,college,grade,role,card_no FROM user WHERE user_id=?`, userID).
		Scan(&u.UserID, &u.Name, &u.Phone, &u.College, &u.Grade, &u.Role, &u.CardNo)
	if err == sql.ErrNoRows {
		return nil, errors.New("用户不存在")
	}
	return &u, err
}

// 改资料（含密码校验）
func (s *AuthService) UpdateProfile(userID int, req models.UpdateProfileRequest) error {
	// 先校验当前密码
	var hashedPwd string
	if err := s.db.QueryRow(`SELECT pwd FROM user WHERE user_id=?`, userID).Scan(&hashedPwd); err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(req.Password)); err != nil {
		return errors.New("密码错误，无法修改")
	}
	// 更新
	_, err := s.db.Exec(`UPDATE user SET name=?,phone=?,college=?,grade=?,card_no=? WHERE user_id=?`,
		req.Name, req.Phone, req.College, req.Grade, req.CardNo, userID)
	return err
}

// 改密码
func (s *AuthService) UpdatePassword(userID int, req models.UpdatePasswordRequest) error {
	var hashedPwd string
	if err := s.db.QueryRow(`SELECT pwd FROM user WHERE user_id=?`, userID).Scan(&hashedPwd); err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(req.OldPassword)); err != nil {
		return errors.New("原密码不正确")
	}
	newHash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	_, err := s.db.Exec(`UPDATE user SET pwd=? WHERE user_id=?`, string(newHash), userID)
	return err
}
