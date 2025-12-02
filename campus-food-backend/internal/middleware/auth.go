package middleware

import (
	"campus-food-backend/internal/utils"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		// 添加调试信息
		fmt.Printf("AuthMiddleware - 请求路径: %s\n", c.Request.URL.Path)
		fmt.Printf("AuthMiddleware - Authorization Header: %s\n", authHeader)

		if authHeader == "" {
			fmt.Println("AuthMiddleware - 错误: 缺少认证令牌")
			utils.Unauthorized(c, "缺少认证令牌")
			c.Abort()
			return
		}

		// 提取 token (Bearer <token>)
		if len(authHeader) <= 7 || !strings.HasPrefix(authHeader, "Bearer ") {
			fmt.Printf("AuthMiddleware - 错误: 令牌格式错误, header: %s\n", authHeader)
			utils.Unauthorized(c, "令牌格式错误")
			c.Abort()
			return
		}

		tokenString := authHeader[7:]
		fmt.Printf("AuthMiddleware - 提取的 Token: %s\n", tokenString)

		claims, err := utils.ValidateJWTToken(tokenString)
		if err != nil {
			fmt.Printf("AuthMiddleware - Token 验证失败: %v\n", err)
			utils.Unauthorized(c, "令牌无效或已过期")
			c.Abort()
			return
		}

		fmt.Printf("AuthMiddleware - Token 验证成功: user_id=%d, role=%s\n", claims.UserID, claims.Role)

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先进行认证检查
		authHeader := c.GetHeader("Authorization")
		// 添加调试信息
		fmt.Printf("AdminMiddleware - 请求路径: %s\n", c.Request.URL.Path)
		fmt.Printf("AdminMiddleware - Authorization Header: %s\n", authHeader)

		if authHeader == "" {
			fmt.Println("AdminMiddleware - 错误: 缺少认证令牌")
			utils.Unauthorized(c, "缺少认证令牌")
			c.Abort()
			return
		}

		if len(authHeader) <= 7 || !strings.HasPrefix(authHeader, "Bearer ") {
			fmt.Printf("AdminMiddleware - 错误: 令牌格式错误, header: %s\n", authHeader)
			utils.Unauthorized(c, "令牌格式错误")
			c.Abort()
			return
		}

		tokenString := authHeader[7:]
		fmt.Printf("AdminMiddleware - 提取的 Token: %s\n", tokenString)

		claims, err := utils.ValidateJWTToken(tokenString)
		if err != nil {
			fmt.Printf("AdminMiddleware - Token 验证失败: %v\n", err)
			utils.Unauthorized(c, "令牌无效或已过期")
			c.Abort()
			return
		}

		fmt.Printf("AdminMiddleware - Token 验证成功: user_id=%d, role=%s\n", claims.UserID, claims.Role)

		// 检查是否为管理员角色
		if claims.Role != "admin" {
			fmt.Printf("AdminMiddleware - 错误: 需要管理员权限，当前角色: %s\n", claims.Role)
			utils.Unauthorized(c, "需要管理员权限")
			c.Abort()
			return
		}

		fmt.Println("AdminMiddleware - 管理员权限验证通过")

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}
