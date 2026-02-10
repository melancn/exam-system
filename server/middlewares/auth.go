package middlewares

import (
	"net/http"
	"server/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 通用认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取JWT token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		// 检查Bearer token格式
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}

		// 解析token并验证
		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// 设置用户信息到上下文中
		c.Set("user", gin.H{
			"id":       claims.UserID,
			"username": claims.Username,
			"role":     claims.Role,
			"isAdmin":  claims.IsAdmin,
		})

		c.Next()
	}
}

// StudentAuthMiddleware 学生认证中间件
func StudentAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		// 检查是否已认证
		user, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userMap := user.(gin.H)
		role := userMap["role"].(byte)

		// 检查是否为学生（角色为1）
		if role != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Student access required"})
			return
		}

		c.Next()
	}
}

// TeacherAuthMiddleware 教师认证中间件
func TeacherAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		// 检查是否已认证
		user, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userMap := user.(gin.H)
		role := userMap["role"].(byte)

		// 检查是否为教师（角色为2）
		if role != 2 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Teacher access required"})
			return
		}

		c.Next()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userMap := user.(gin.H)
		role := userMap["role"].(byte) // JWT claims中的数字类型
		isAdmin := userMap["isAdmin"].(byte)

		// 检查是否为管理员用户（角色为2且is_admin为true）
		if role != 2 || isAdmin != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Admin access required"})
			return
		}

		c.Next()
	}
}
