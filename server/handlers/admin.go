package handlers

import (
	"net/http"
	"server/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateTeacher(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password"`
			Name     string `json:"name" binding:"required"`
			Role     string `json:"role" binding:"required"`
			IsAdmin  bool   `json:"isAdmin"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid teacher data"})
			return
		}

		if request.Role != "teacher" && request.Role != "admin" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
			return
		}

		// 如果密码为空，设置默认密码
		if request.Password == "" {
			request.Password = "123456"
		}

		// 密码需要哈希加密
		passwordHash := models.HashPassword(request.Password)

		teacher := models.User{
			Username: request.Username,
			Password: passwordHash,
			Role:     2, // 2: teacher
			Name:     request.Name,
			IsAdmin:  0, // 默认不是管理员
			Status:   1, // 1: active
		}

		if err := db.Create(&teacher).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create teacher"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":       teacher.ID,
			"username": teacher.Username,
			"name":     teacher.Name,
			"role":     teacher.Role,
			"isAdmin":  teacher.IsAdmin,
			"status":   teacher.Status,
			"message":  "Teacher created successfully",
		})
	}
}

func UpdateTeacher(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var teacher models.User
		if err := db.Where("id = ? AND role = 2", id).First(&teacher).Error; err != nil { // 2: teacher
			c.JSON(http.StatusNotFound, gin.H{"error": "Teacher not found"})
			return
		}

		var updateData struct {
			Name    string `json:"name"`
			IsAdmin int    `json:"isAdmin"`
			Status  int    `json:"status"`
		}

		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid teacher data"})
			return
		}

		// 更新字段
		if updateData.Name != "" {
			teacher.Name = updateData.Name
		}

		if err := db.Save(&teacher).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update teacher"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":       teacher.ID,
			"username": teacher.Username,
			"name":     teacher.Name,
			"role":     teacher.Role,
			"isAdmin":  teacher.IsAdmin,
			"status":   teacher.Status,
			"message":  "Teacher updated successfully",
		})
	}
}

func DeleteTeacher(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var teacher models.User
		if err := db.Where("id = ? AND role = 2", id).First(&teacher).Error; err != nil { // 2: teacher
			c.JSON(http.StatusNotFound, gin.H{"error": "Teacher not found"})
			return
		}

		// 检查是否是最后一个管理员
		if teacher.IsAdmin == 1 {
			var adminCount int64
			db.Model(&models.User{}).Where("role = 2 AND is_admin = ?", 1).Count(&adminCount) // 2: teacher
			if adminCount <= 1 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete the last admin"})
				return
			}
		}

		if err := db.Delete(&teacher).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete teacher"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Teacher deleted successfully"})
	}
}
