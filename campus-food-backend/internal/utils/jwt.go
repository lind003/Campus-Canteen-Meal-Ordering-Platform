package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var jwtSecret = []byte("campus-food-secret-key-2024")

type Claims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	Exp    int64  `json:"exp"`
	Iat    int64  `json:"iat"`
}

// GenerateJWTToken 生成 JWT token
func GenerateJWTToken(userID int, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := Claims{
		UserID: userID,
		Role:   role,
		Exp:    expirationTime.Unix(),
		Iat:    time.Now().Unix(),
	}

	fmt.Printf("GenerateJWTToken - 生成 token for user_id=%d, role=%s\n", userID, role)

	// 编码 header
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	headerJSON, _ := json.Marshal(header)
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)

	// 编码 payload
	payloadJSON, _ := json.Marshal(claims)
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// 创建签名
	signatureInput := headerEncoded + "." + payloadEncoded
	h := hmac.New(sha256.New, jwtSecret)
	h.Write([]byte(signatureInput))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	// 组合 JWT
	token := signatureInput + "." + signature

	fmt.Printf("GenerateJWTToken - 生成的 token: %s\n", token)
	return token, nil
}

// ValidateJWTToken 验证 JWT token
func ValidateJWTToken(tokenString string) (*Claims, error) {
	fmt.Printf("ValidateJWTToken - 验证 token: %s\n", tokenString)

	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		fmt.Printf("ValidateJWTToken - 错误: token 格式错误，部分数: %d\n", len(parts))
		return nil, errors.New("invalid token format")
	}

	// 验证签名
	signatureInput := parts[0] + "." + parts[1]
	h := hmac.New(sha256.New, jwtSecret)
	h.Write([]byte(signatureInput))
	expectedSignature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	fmt.Printf("ValidateJWTToken - 计算签名: %s\n", expectedSignature)
	fmt.Printf("ValidateJWTToken - 实际签名: %s\n", parts[2])

	if parts[2] != expectedSignature {
		fmt.Printf("ValidateJWTToken - 错误: 签名不匹配\n")
		return nil, errors.New("invalid signature")
	}

	// 解码 payload
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		fmt.Printf("ValidateJWTToken - 错误: 解码 payload 失败: %v\n", err)
		return nil, err
	}

	var claims Claims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		fmt.Printf("ValidateJWTToken - 错误: 解析 claims 失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("ValidateJWTToken - 解析的 claims: user_id=%d, role=%s, exp=%d\n",
		claims.UserID, claims.Role, claims.Exp)

	// 检查过期时间
	currentTime := time.Now().Unix()
	fmt.Printf("ValidateJWTToken - 当前时间: %d, 过期时间: %d\n", currentTime, claims.Exp)

	if currentTime > claims.Exp {
		fmt.Printf("ValidateJWTToken - 错误: token 已过期\n")
		return nil, errors.New("token expired")
	}

	fmt.Printf("ValidateJWTToken - token 验证成功\n")
	return &claims, nil
}

// InitJWT 初始化 JWT 配置
func InitJWT(secret string) {
	if secret != "" {
		jwtSecret = []byte(secret)
		fmt.Println("InitJWT - JWT 密钥已设置")
	} else {
		fmt.Println("InitJWT - 使用默认 JWT 密钥")
	}
}
