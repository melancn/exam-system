package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JWT密钥，在生产环境中应该从环境变量中读取
var jwtSecret = []byte("exam_system_secret_key")

// Claims 定义JWT的声明结构
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     byte   `json:"role"`
	IsAdmin  byte   `json:"is_admin"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT token
func GenerateToken(userID uint, username string, role byte, isAdmin byte) (string, error) {
	// 设置token过期时间（7天）
	expirationTime := time.Now().Add(7 * 24 * time.Hour)

	// 创建声明
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		IsAdmin:  isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "exam_system",
		},
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名token
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken 解析并验证JWT token
func ParseToken(tokenString string) (*Claims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	// 验证token是否有效
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken 刷新token（重新生成token）
func RefreshToken(tokenString string) (string, error) {
	// 先验证原token
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 如果token即将过期（剩余时间小于1天），则重新生成
	if time.Until(claims.ExpiresAt.Time) < 24*time.Hour {
		return GenerateToken(claims.UserID, claims.Username, claims.Role, claims.IsAdmin)
	}

	// 否则返回原token
	return tokenString, nil
}
