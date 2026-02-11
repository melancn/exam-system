package handlers

import (
	"net/http"
	"server/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetExams 获取所有试卷列表
func GetExams(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 分页参数
		page := c.DefaultQuery("page", "1")
		pageSize := c.DefaultQuery("pageSize", "10")
		keyword := c.DefaultQuery("keyword", "")
		status := c.DefaultQuery("status", "")

		pageNum := 1
		pageSizeNum := 10

		// 验证分页参数
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			pageNum = p
		}
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			pageSizeNum = ps
		}

		// 构建查询
		query := db.Model(&models.Exam{})

		// 搜索条件
		if keyword != "" {
			search := "%" + keyword + "%"
			query = query.Where("title LIKE ?", search)
		}

		// 状态过滤
		if status != "" {
			query = query.Where("status = ?", status)
		}

		// 获取总数
		var total int64
		if err := query.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count exams"})
			return
		}

		// 分页查询
		var exams []models.Exam
		offset := (pageNum - 1) * pageSizeNum
		if err := query.Offset(offset).Limit(pageSizeNum).Find(&exams).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch exams"})
			return
		}

		// 构建响应数据
		var response []gin.H
		for _, exam := range exams {
			// 获取题目数量
			var questionCount int64
			db.Model(&models.Question{}).Where("exam_id = ?", exam.ID).Count(&questionCount)

			// 获取单选题和填空题数量
			var singleChoiceCount int64
			db.Model(&models.Question{}).Where("exam_id = ? AND type = 'single'", exam.ID).Count(&singleChoiceCount)

			var fillBlankCount int64
			db.Model(&models.Question{}).Where("exam_id = ? AND type = 'fill'", exam.ID).Count(&fillBlankCount)

			response = append(response, gin.H{
				"id":                exam.ID,
				"title":             exam.Title,
				"description":       exam.Description,
				"totalScore":        exam.TotalScore,
				"questionCount":     questionCount,
				"singleChoiceCount": singleChoiceCount,
				"fillBlankCount":    fillBlankCount,
				"status":            exam.Status,
				"createTime":        exam.CreatedAt,
				"updateTime":        exam.UpdatedAt,
			})
		}

		// 确保返回空数组而不是null
		if response == nil {
			response = []gin.H{}
		}

		c.JSON(http.StatusOK, gin.H{
			"data": response,
			"pagination": gin.H{
				"page":       pageNum,
				"pageSize":   pageSizeNum,
				"total":      total,
				"totalPages": (total + int64(pageSizeNum) - 1) / int64(pageSizeNum),
			},
		})
	}
}

// GetExamDetails 获取试卷详情
func GetExamDetails(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var exam models.Exam
		if err := db.First(&exam, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
			return
		}

		// 获取试卷相关问题
		var questions []models.Question
		if err := db.Where("exam_id = ?", id).Find(&questions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch questions"})
			return
		}

		// 构建响应数据，确保字段名与前端一致
		var questionResponses []gin.H
		for _, question := range questions {
			questionResponse := gin.H{
				"id":            question.ID,
				"type":          question.Type,
				"content":       question.Content,
				"score":         question.Score,
				"correctAnswer": question.Answer,
				"placeholder":   question.Placeholder,
			}

			// 根据题目类型添加特定字段
			if question.Type == "single" {
				// 单选题需要选项数据
				questionResponse["options"] = question.Options
			} else if question.Type == "fill" {
				// 填空题需要答案数据
				questionResponse["answers"] = question.Answers
			}

			questionResponses = append(questionResponses, questionResponse)
		}

		response := gin.H{
			"id":          exam.ID,
			"title":       exam.Title,
			"description": exam.Description,
			"totalScore":  exam.TotalScore,
			"status":      exam.Status,
			"questions":   questionResponses,
		}
		c.JSON(http.StatusOK, response)
	}
}

// CreateExam 创建新试卷
func CreateExam(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			Exam      models.Exam       `json:"exam"`
			Questions []models.Question `json:"questions"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		// 开始事务
		tx := db.Begin()

		// 创建试卷
		if err := tx.Create(&request.Exam).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create exam"})
			return
		}

		// 创建问题
		for _, item := range request.Questions {
			item.ExamID = request.Exam.ID
			// 处理填空题的详细配置
			if item.Type == "fill" && len(item.Answers) == 0 {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid fill-in answers format"})
				return
			}
			if err := tx.Create(&item).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create questions"})
				return
			}
		}

		tx.Commit()
		c.JSON(http.StatusCreated, gin.H{
			"exam":      request.Exam,
			"questions": request.Questions,
			"message":   "Exam created successfully",
		})
	}
}

// UpdateExam 更新试卷
func UpdateExam(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var request struct {
			Exam      models.Exam       `json:"exam"`
			Questions []models.Question `json:"questions"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		var existingExam models.Exam
		if err := db.First(&existingExam, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
			return
		}

		// 开始事务
		tx := db.Begin()

		// 更新试卷
		if err := tx.Model(&existingExam).Updates(request.Exam).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update exam"})
			return
		}

		// 删除旧问题
		if err := tx.Where("exam_id = ?", id).Delete(&models.Question{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete old questions"})
			return
		}

		// 创建新问题
		for i := range request.Questions {
			request.Questions[i].ExamID = existingExam.ID
			// 确保填空题字段有默认值
			if err := tx.Create(&request.Questions[i]).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create questions"})
				return
			}
		}

		tx.Commit()
		c.JSON(http.StatusOK, gin.H{
			"message": "Exam updated successfully",
		})
	}
}

// UpdateExamStatus 更新试卷状态
func UpdateExamStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var request struct {
			Status string `json:"status" binding:"required,oneof=published draft archived"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		var exam models.Exam
		if err := db.First(&exam, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
			return
		}

		// 更新试卷状态
		if err := db.Model(&exam).Update("status", request.Status).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update exam status"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Exam status updated successfully",
			"data": gin.H{
				"id":     exam.ID,
				"title":  exam.Title,
				"status": request.Status,
			},
		})
	}
}

// DeleteExam 删除试卷
func DeleteExam(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var exam models.Exam
		if err := db.First(&exam, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
			return
		}

		// 开始事务
		tx := db.Begin()

		// 删除相关问题
		if err := tx.Where("exam_id = ?", id).Delete(&models.Question{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete questions"})
			return
		}

		// 删除试卷
		if err := tx.Delete(&exam).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete exam"})
			return
		}

		tx.Commit()
		c.JSON(http.StatusOK, gin.H{
			"message": "Exam deleted successfully",
		})
	}
}
