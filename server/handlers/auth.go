package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"server/models"
	"server/utils"
)

func Login(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginReq struct {
			Username  string `json:"username"`
			Password  string `json:"password"`
			IsTeacher bool   `json:"isTeacher"`
		}

		if err := c.ShouldBindJSON(&loginReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		var user models.User
		if err := db.Where("username = ?", loginReq.Username).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials1"})
			return
		}

		// 验证用户角色是否匹配
		if loginReq.IsTeacher && user.Role == 1 { // 1: student
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials2"})
			return
		}
		if !loginReq.IsTeacher && user.Role != 1 { // 1: student
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials3"})
			return
		}

		// 使用SHA256验证密码
		if !models.CheckPassword(loginReq.Password, user.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// 记录登录日志
		loginLog := models.LoginLog{
			Username:  user.Username,
			IP:        c.ClientIP(),
			UserAgent: c.GetHeader("User-Agent"),
			CreatedAt: time.Now().UTC(),
		}
		db.Create(&loginLog)

		// 生成JWT token
		token, err := utils.GenerateToken(user.ID, user.Username, user.Role, user.IsAdmin)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
			"token":   token,
			"user": gin.H{
				"id":      user.ID,
				"name":    user.Name,
				"role":    user.Role,
				"isAdmin": user.IsAdmin,
			},
		})
	}
}
