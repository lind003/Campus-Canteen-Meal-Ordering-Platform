package handlers

import (
	"campus-food-backend/internal/models"
	"campus-food-backend/internal/services"
	"campus-food-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, "参数错误: "+err.Error())
		return
	}

	user, token, err := h.authService.Login(req)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"user":  user,
		"token": token,
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, "参数错误: "+err.Error())
		return
	}

	user, token, err := h.authService.Register(req)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "注册成功", gin.H{
		"user":  user,
		"token": token,
	})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetInt("user_id")

	// 从数据库获取完整的用户信息
	utils.Success(c, gin.H{
		"user_id": userID,
		"message": "获取用户信息成功",
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// 提取 token (Bearer <token>)
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token := authHeader[7:]
			h.authService.Logout(token)
		}
	}

	utils.SuccessWithMessage(c, "退出登录成功", nil)
}
