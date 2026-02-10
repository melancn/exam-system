package handlers

import (
	"fmt"
	"net/http"
	"server/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetAssignments 获取试卷分配列表（包含关联数据和分页）
func GetAssignments(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 分页参数
		page := c.DefaultQuery("page", "1")
		pageSize := c.DefaultQuery("pageSize", "10")
		keyword := c.DefaultQuery("keyword", "")

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
		query := db.Preload("Exam").Preload("Class")

		// 搜索条件
		if keyword != "" {
			search := "%" + keyword + "%"
			query = query.Where("exam_assignments.description LIKE ? OR exams.title LIKE ? OR classes.name LIKE ?",
				search, search, search)
		}

		// 获取总数
		var total int64
		if err := query.Model(&models.ExamAssignment{}).Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count assignments"})
			return
		}

		// 分页查询
		var assignments []models.ExamAssignment
		offset := (pageNum - 1) * pageSizeNum
		if err := query.Offset(offset).Limit(pageSizeNum).Find(&assignments).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignments"})
			return
		}

		// 构建响应数据
		var response []gin.H
		for _, assignment := range assignments {
			response = append(response, gin.H{
				"id":          strconv.FormatUint(uint64(assignment.ID), 10),
				"examId":      strconv.FormatUint(uint64(assignment.ExamID), 10),
				"examTitle":   assignment.Exam.Title,
				"classId":     strconv.FormatUint(uint64(assignment.ClassID), 10),
				"className":   assignment.Class.Name,
				"startTime":   assignment.StartTime,
				"endTime":     assignment.EndTime,
				"duration":    assignment.Duration,
				"passScore":   assignment.PassScore,
				"description": assignment.Description,
				"createdAt":   assignment.CreatedAt,
			})
		}

		// 确保返回空数组而不是null
		if response == nil {
			response = []gin.H{}
		}

		c.JSON(http.StatusOK, gin.H{
			"assignments": response,
			"pagination": gin.H{
				"page":       pageNum,
				"pageSize":   pageSizeNum,
				"total":      total,
				"totalPages": (total + int64(pageSizeNum) - 1) / int64(pageSizeNum),
			},
		})
	}
}

// CreateAssignment 创建试卷分配（支持批量创建）
func CreateAssignment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			ExamID      uint   `json:"examId" binding:"required"`
			ClassIDs    []uint `json:"classIds" binding:"required"`
			StartTime   string `json:"startTime" binding:"required"`
			EndTime     string `json:"endTime" binding:"required"`
			Duration    int    `json:"duration" binding:"required"`
			PassScore   int    `json:"passScore" binding:"required"`
			Description string `json:"description"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment data"})
			return
		}

		// 验证试卷是否存在
		var exam models.Exam
		if err := db.First(&exam, request.ExamID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Exam not found"})
			return
		}

		if exam.Status != "published" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Exam must be published"})
			return
		}

		var assignments []models.ExamAssignment
		var assignmentIds []uint

		// 开始事务
		tx := db.Begin()

		for _, classID := range request.ClassIDs {
			// 验证班级是否存在
			var class models.Class
			if err := tx.First(&class, classID).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"error": "Class not found: " + strconv.FormatUint(uint64(classID), 10)})
				return
			}

			assignment := models.ExamAssignment{
				ExamID:      request.ExamID,
				ClassID:     classID,
				StartTime:   request.StartTime,
				EndTime:     request.EndTime,
				Duration:    request.Duration,
				PassScore:   request.PassScore,
				Description: request.Description,
			}

			if err := tx.Create(&assignment).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create assignment"})
				return
			}

			assignments = append(assignments, assignment)
			assignmentIds = append(assignmentIds, assignment.ID)
		}

		tx.Commit()
		c.JSON(http.StatusCreated, gin.H{
			"assignments":   assignments,
			"assignmentIds": assignmentIds,
			"message":       "Assignments created successfully",
		})
	}
}

// UpdateAssignment 更新试卷分配
func UpdateAssignment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var assignment models.ExamAssignment
		if err := db.First(&assignment, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
			return
		}

		var request struct {
			StartTime   string `json:"startTime"`
			EndTime     string `json:"endTime"`
			Duration    int    `json:"duration"`
			PassScore   int    `json:"passScore"`
			Description string `json:"description"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assignment data"})
			return
		}

		// 只更新允许修改的字段
		updates := models.ExamAssignment{
			StartTime:   request.StartTime,
			EndTime:     request.EndTime,
			Duration:    request.Duration,
			PassScore:   request.PassScore,
			Description: request.Description,
		}

		if err := db.Model(&assignment).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update assignment"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Assignment updated successfully",
		})
	}
}

// DeleteAssignment 删除试卷分配
func DeleteAssignment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var assignment models.ExamAssignment
		if err := db.First(&assignment, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
			return
		}

		if err := db.Delete(&assignment).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete assignment"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Assignment deleted successfully",
		})
	}
}

// GetClasses 获取班级列表（用于前端选择）
func GetClasses(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var classes []models.Class
		if err := db.Find(&classes).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch classes"})
			return
		}

		var response []gin.H
		for _, class := range classes {
			// 获取班级学生数量
			var studentCount int64
			db.Model(&models.User{}).Where("class_id = ? AND role = 1", class.ID).Count(&studentCount)

			response = append(response, gin.H{
				"id":             class.ID,
				"name":           class.Name,
				"description":    class.Description,
				"teacherId":      class.TeacherID,
				"major":          class.Major,
				"enrollmentYear": class.EnrollmentYear,
				"studentCount":   studentCount,
			})
		}

		// 确保返回空数组而不是null
		if response == nil {
			response = []gin.H{}
		}

		c.JSON(http.StatusOK, gin.H{
			"classes": response,
			"total":   len(response),
		})
	}
}

// CreateClass 创建班级
func CreateClass(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			Name           string `json:"name" binding:"required"`
			Description    string `json:"description"`
			TeacherID      *uint  `json:"teacherId"`      // 改为指针，允许nil
			Major          string `json:"major"`          // 新增字段
			EnrollmentYear int    `json:"enrollmentYear"` // 新增字段，使用int类型
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid class data"})
			return
		}

		// 验证教师是否存在（如果提供了teacherId）
		var teacher *models.User
		if request.TeacherID != nil && *request.TeacherID > 0 {
			if err := db.First(&teacher, *request.TeacherID).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Teacher not found"})
				return
			}
		}

		// 检查班级名称是否已存在
		var existingClass models.Class
		if err := db.Where("name = ?", request.Name).First(&existingClass).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Class name already exists"})
			return
		}

		class := models.Class{
			Name:           request.Name,
			Description:    request.Description,
			Major:          request.Major,
			EnrollmentYear: request.EnrollmentYear,
		}

		// 只有当teacherId有效时才设置
		if request.TeacherID != nil && *request.TeacherID > 0 {
			class.TeacherID = *request.TeacherID
		}

		if err := db.Create(&class).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create class"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":             class.ID,
			"name":           class.Name,
			"description":    class.Description,
			"teacherId":      class.TeacherID,
			"major":          class.Major,
			"enrollmentYear": class.EnrollmentYear,
			"message":        "Class created successfully",
		})
	}
}

// GetClassDetails 获取班级详情
func GetClassDetails(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var class models.Class
		if err := db.First(&class, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Class not found"})
			return
		}

		// 获取教师信息
		var teacher models.User
		db.First(&teacher, class.TeacherID)

		// 获取学生数量
		var studentCount int64
		db.Model(&models.User{}).Where("class_id = ? AND role = 1", id).Count(&studentCount)

		response := gin.H{
			"id":             class.ID,
			"name":           class.Name,
			"description":    class.Description,
			"teacherId":      class.TeacherID,
			"teacherName":    teacher.Name,
			"major":          class.Major,
			"enrollmentYear": class.EnrollmentYear,
			"studentCount":   studentCount,
			"createdAt":      class.CreatedAt,
			"updatedAt":      class.UpdatedAt,
		}

		c.JSON(http.StatusOK, response)
	}
}

// UpdateClass 更新班级信息
func UpdateClass(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var class models.Class
		if err := db.First(&class, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Class not found"})
			return
		}

		var request struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			TeacherID   uint   `json:"teacherId"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid class data"})
			return
		}

		// 如果修改了名称，检查是否已存在
		if request.Name != "" && request.Name != class.Name {
			var existingClass models.Class
			if err := db.Where("name = ?", request.Name).First(&existingClass).Error; err == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Class name already exists"})
				return
			}
			class.Name = request.Name
		}

		if request.Description != "" {
			class.Description = request.Description
		}

		if request.TeacherID != 0 {
			// 验证新教师是否存在
			var teacher models.User
			if err := db.First(&teacher, request.TeacherID).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Teacher not found"})
				return
			}
			class.TeacherID = request.TeacherID
		}

		if err := db.Save(&class).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update class"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Class updated successfully",
		})
	}
}

// DeleteClass 删除班级
func DeleteClass(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var class models.Class
		if err := db.First(&class, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Class not found"})
			return
		}

		// 检查班级是否有学生
		var studentCount int64
		db.Model(&models.User{}).Where("class_id = ? AND role = 1", id).Count(&studentCount)
		if studentCount > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete class with students"})
			return
		}

		if err := db.Delete(&class).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete class"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Class deleted successfully",
		})
	}
}

// GetClassStudents 获取班级学生列表
func GetClassStudents(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// 验证班级是否存在
		var class models.Class
		if err := db.First(&class, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Class not found"})
			return
		}

		var students []models.User
		if err := db.Where("class_id = ? AND role = 1", id).Find(&students).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch students"})
			return
		}

		var studentList []gin.H
		for _, student := range students {
			studentList = append(studentList, gin.H{
				"id":        student.ID,
				"studentId": student.StudentID,
				"name":      student.Name,
				"classId":   student.ClassId,
				"major":     student.Major,
				"joinTime":  student.CreatedAt,
			})
		}

		// 确保返回空数组而不是null
		if studentList == nil {
			studentList = []gin.H{}
		}

		c.JSON(http.StatusOK, gin.H{
			"classId":   class.ID,
			"className": class.Name,
			"students":  studentList,
			"total":     len(studentList),
		})
	}
}

// GetClassStatistics 获取班级统计信息
func GetClassStatistics(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		classId := c.Query("classId")

		// 如果提供了classId，只获取指定班级的统计信息
		if classId != "" {
			var class models.Class
			if err := db.First(&class, classId).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Class not found"})
				return
			}

			statistics := calculateClassStatistics(db, class)
			c.JSON(http.StatusOK, statistics)
			return
		}

		// 否则获取所有班级的统计信息
		var classes []models.Class
		if err := db.Find(&classes).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch classes"})
			return
		}

		var statistics []gin.H
		for _, class := range classes {
			classStats := calculateClassStatistics(db, class)
			statistics = append(statistics, classStats)
		}

		// 确保返回空数组而不是null
		if statistics == nil {
			statistics = []gin.H{}
		}

		c.JSON(http.StatusOK, statistics)
	}
}

// calculateClassStatistics 计算单个班级的统计信息
func calculateClassStatistics(db *gorm.DB, class models.Class) gin.H {
	// 获取学生数量
	var studentCount int64
	db.Model(&models.User{}).Where("class_id = ? AND role = 1", class.ID).Count(&studentCount)

	// 如果没有学生，返回默认值
	if studentCount == 0 {
		return gin.H{
			"classId":       class.ID,
			"className":     class.Name,
			"studentCount":  0,
			"avgScore":      0,
			"passRate":      "0%",
			"excellentRate": "0%",
			"examCount":     0,
		}
	}

	// 获取班级学生ID列表
	var students []models.User
	db.Where("class_id = ? AND role = 1", class.ID).Find(&students)

	var studentIds []uint
	for _, student := range students {
		studentIds = append(studentIds, student.ID)
	}

	// 获取班级相关的考试分配
	var examAssignments []models.ExamAssignment
	db.Where("class_id = ?", class.ID).Find(&examAssignments)

	var assignmentIds []uint
	for _, assignment := range examAssignments {
		assignmentIds = append(assignmentIds, assignment.ID)
	}

	// 如果没有考试分配，返回默认值
	if len(assignmentIds) == 0 {
		return gin.H{
			"classId":       class.ID,
			"className":     class.Name,
			"studentCount":  studentCount,
			"avgScore":      0,
			"passRate":      "0%",
			"excellentRate": "0%",
			"examCount":     0,
		}
	}

	// 获取班级学生的考试结果
	var examResults []models.ExamResult
	db.Where("exam_assignment_id IN ? AND student_id IN ?", assignmentIds, studentIds).Find(&examResults)

	// 如果没有考试结果，返回默认值
	if len(examResults) == 0 {
		return gin.H{
			"classId":       class.ID,
			"className":     class.Name,
			"studentCount":  studentCount,
			"avgScore":      0,
			"passRate":      "0%",
			"excellentRate": "0%",
			"examCount":     len(assignmentIds),
		}
	}

	// 计算统计数据
	var totalScore int
	var passCount int
	var excellentCount int

	for _, result := range examResults {
		totalScore += result.Score

		// 获取对应的考试分配以获取及格分数
		var assignment models.ExamAssignment
		db.First(&assignment, result.ExamAssignmentID)

		if result.Score >= assignment.PassScore {
			passCount++
		}

		// 优秀率：分数达到总分90%以上
		if assignment.ExamID > 0 {
			var exam models.Exam
			db.First(&exam, assignment.ExamID)
			excellentThreshold := exam.TotalScore * 9 / 10 // 90%
			if result.Score >= excellentThreshold {
				excellentCount++
			}
		}
	}

	avgScore := float64(totalScore) / float64(len(examResults))
	passRate := float64(passCount) / float64(len(examResults)) * 100
	excellentRate := float64(excellentCount) / float64(len(examResults)) * 100

	return gin.H{
		"classId":       class.ID,
		"className":     class.Name,
		"studentCount":  studentCount,
		"avgScore":      avgScore,
		"passRate":      fmt.Sprintf("%.1f%%", passRate),
		"excellentRate": fmt.Sprintf("%.1f%%", excellentRate),
		"examCount":     len(assignmentIds),
	}
}
