package handlers

import (
	"encoding/json"
	"net/http"
	"server/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetStudentExams 获取学生可参加的考试列表
func GetStudentExams(db *gorm.DB) gin.HandlerFunc {
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

		// 从上下文获取用户信息
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		userMap := user.(gin.H)
		studentID := userMap["id"].(uint)

		// 获取学生所在班级
		var student models.User
		if err := db.First(&student, studentID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch student"})
			return
		}

		// 构建查询：获取学生所在班级的考试分配
		query := db.Preload("Class").Preload("Exam.Questions").InnerJoins("Exam")

		// 添加班级条件
		if student.ClassId > 0 {
			query = query.Where("class_id = ? and Exam.status='published'", student.ClassId)
		} else {
			// 如果学生没有班级，返回空结果
			c.JSON(http.StatusOK, gin.H{
				"exams": []gin.H{},
				"pagination": gin.H{
					"page":       pageNum,
					"pageSize":   pageSizeNum,
					"total":      0,
					"totalPages": 0,
				},
			})
			return
		}

		// 搜索条件
		if keyword != "" {
			search := "%" + keyword + "%"
			query = query.Where("exams.title LIKE ? OR exams.description LIKE ? OR classes.name LIKE ?",
				search, search, search)
		}

		// 获取总数
		var total int64
		if err := query.Model(&models.ExamAssignment{}).Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count exams"})
			return
		}

		// 分页查询
		var assignments []models.ExamAssignment
		offset := (pageNum - 1) * pageSizeNum
		if err := query.Offset(offset).Limit(pageSizeNum).Find(&assignments).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch exam assignments"})
			return
		}

		// 构建响应数据
		var exams []gin.H
		for _, assignment := range assignments {
			// 计算考试状态
			status := getExamStatus(assignment.StartTime, assignment.EndTime)

			exams = append(exams, gin.H{
				"id":            assignment.Exam.ID,
				"title":         assignment.Exam.Title,
				"description":   assignment.Exam.Description,
				"duration":      assignment.Duration,
				"totalScore":    assignment.Exam.TotalScore,
				"questionCount": len(assignment.Exam.Questions), // 需要预加载Questions
				"startTime":     assignment.StartTime,
				"endTime":       assignment.EndTime,
				"passScore":     assignment.PassScore,
				"assignmentId":  assignment.ID,
				"className":     assignment.Class.Name,
				"major":         assignment.Class.Major,
				"status":        status,
			})
		}

		// 确保返回空数组而不是null
		if exams == nil {
			exams = []gin.H{}
		}

		c.JSON(http.StatusOK, gin.H{
			"exams": exams,
			"pagination": gin.H{
				"page":       pageNum,
				"pageSize":   pageSizeNum,
				"total":      total,
				"totalPages": (total + int64(pageSizeNum) - 1) / int64(pageSizeNum),
			},
		})
	}
}

type ans struct {
	QuestionID uint     `json:"questionId"`
	Type       string   `json:"type"` // "single" or "fill"
	Answer     string   `json:"answer"`
	Answers    []string `json:"answers"`
}

// SubmitExam 提交考试结果
func SubmitExam(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		examID := c.Param("id")

		var request struct {
			Answers  []ans `json:"answers"`
			TimeUsed int   `json:"timeUsed"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exam result"})
			return
		}

		// 从上下文获取用户信息
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		userMap := user.(gin.H)
		studentID := userMap["id"].(uint)

		var assignment models.ExamAssignment
		if err := db.Where("id = ?", examID).First(&assignment).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "No permission to access this exam"})
			return
		}

		// 验证考试是否存在
		var exam models.Exam
		if err := db.First(&exam, assignment.ExamID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Exam not found"})
			return
		}

		// 计算分数（根据正确答案计算）
		score := calculateScore(request.Answers, assignment.ExamID, db)
		b, err := json.Marshal(request.Answers)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to serialize answers"})
			return
		}

		// 创建考试结果
		result := models.ExamResult{
			ExamAssignmentID: assignment.ID,
			StudentID:        studentID,
			Score:            score,
			Answers:          string(b),
			TimeUsed:         request.TimeUsed,
		}

		if err := db.Create(&result).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit exam"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"score":      score,
			"totalScore": exam.TotalScore,
			"message":    "Exam submitted successfully",
		})
	}
}

// GetExamResults 获取学生考试结果
func GetExamResults(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 分页参数
		page := c.DefaultQuery("page", "1")
		pageSize := c.DefaultQuery("pageSize", "10")

		pageNum := 1
		pageSizeNum := 10

		// 验证分页参数
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			pageNum = p
		}
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			pageSizeNum = ps
		}

		// 从上下文获取用户信息
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		userMap := user.(gin.H)
		studentID := userMap["id"].(uint)

		// 构建查询
		query := db.Preload("ExamAssignment").Preload("ExamAssignment.Exam").Where("student_id = ?", studentID)

		// 获取总数
		var total int64
		if err := query.Model(&models.ExamResult{}).Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count results"})
			return
		}

		// 分页查询
		var results []models.ExamResult
		offset := (pageNum - 1) * pageSizeNum
		if err := query.Offset(offset).Limit(pageSizeNum).Find(&results).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch results" + err.Error()})
			return
		}

		// 构建响应数据
		var response []gin.H
		for _, result := range results {
			response = append(response, gin.H{
				"id":               result.ID,
				"examAssignmentId": result.ExamAssignmentID,
				"examTitle":        result.ExamAssignment.Exam.Title,
				"score":            result.Score,
				"totalScore":       result.ExamAssignment.Exam.TotalScore,
				"timeUsed":         result.TimeUsed,
				"submitTime":       result.CreatedAt,
				"passed":           result.Score >= 60,
			})
		}

		// 确保返回空数组而不是null
		if response == nil {
			response = []gin.H{}
		}

		c.JSON(http.StatusOK, gin.H{
			"results": response,
			"pagination": gin.H{
				"total": total,
			},
		})
	}
}

// GetExamDetailsForStudent 获取试卷详情（学生端）
func GetExamDetailsForStudent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		examID := c.Param("id")

		// 从上下文获取用户信息
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		userMap := user.(gin.H)
		var userInfo models.User
		db.Where("id=?", userMap["id"]).First(&userInfo)

		// 验证学生是否已参加考试
		var result models.ExamResult
		if err := db.Where("exam_assignment_id = ? AND student_id = ?", examID, userMap["id"]).First(&result).Error; err == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You have not taken this exam"})
			return
		}

		// 验证学生是否有权限访问该试卷
		var assignment models.ExamAssignment
		if err := db.Where("id = ? AND class_id = ?", examID, userInfo.ClassId).First(&assignment).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "No permission to access this exam"})
			return
		}

		// 获取试卷信息
		var exam models.Exam
		if err := db.First(&exam, assignment.ExamID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
			return
		}

		// 获取试卷相关问题
		var questions []models.Question
		if err := db.Where("exam_id = ?", assignment.ExamID).Find(&questions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch questions"})
			return
		}

		// 构建响应数据，确保字段名与前端一致
		var questionResponses []gin.H
		for _, question := range questions {
			questionResponse := gin.H{
				"id":      question.ID,
				"type":    question.Type,
				"content": question.Content,
				"score":   question.Score,
			}

			// 根据题目类型添加特定字段
			if question.Type == "single" {
				questionResponse["options"] = question.Options
			} else if question.Type == "fill" {
				// 填空题需要配置数据，但不返回答案
				questionResponse["placeholder"] = question.Placeholder

				// 解析填空题的详细配置
				if len(question.Answers) > 0 {
					// 只返回配置，不返回具体的答案选项
					var configs []gin.H
					for _, config := range question.Answers {
						configs = append(configs, gin.H{
							"type": config.Type,
						})
					}
					questionResponse["answerConfigs"] = configs
				}
			}

			questionResponses = append(questionResponses, questionResponse)
		}

		response := gin.H{
			"id":         assignment.ExamID,
			"title":      exam.Title,
			"duration":   assignment.Duration,
			"totalScore": exam.TotalScore,
			"questions":  questionResponses,
		}
		c.JSON(http.StatusOK, response)
	}
}

// GetExamResult 获取单个考试结果详情
func GetExamResult(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		resultID := c.Param("resultId")

		// 从上下文获取用户信息
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		userMap := user.(gin.H)
		studentID := userMap["id"].(uint)

		var result models.ExamResult
		if err := db.Preload("ExamAssignment").Preload("ExamAssignment.Exam").Where("id = ? AND student_id = ?", resultID, studentID).First(&result).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Result not found"})
			return
		}

		response := gin.H{
			"id":         result.ID,
			"examId":     result.ExamAssignment.ID,
			"examTitle":  result.ExamAssignment.Exam.Title,
			"score":      result.Score,
			"totalScore": result.ExamAssignment.Exam.TotalScore,
			"timeUsed":   result.TimeUsed,
			"submitTime": result.CreatedAt,
			"passed":     result.Score >= 60,
			"answers":    result.Answers,
		}

		c.JSON(http.StatusOK, response)
	}
}

// GetStudentProfile 获取学生个人信息
func GetStudentProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户信息
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		userMap := user.(gin.H)
		studentID := userMap["id"].(uint)

		// 获取学生信息
		var student models.User
		if err := db.First(&student, studentID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch student info"})
			return
		}

		// 获取班级信息
		var classInfo models.Class
		if student.ClassId > 0 {
			db.First(&classInfo, student.ClassId)
		}

		// 获取学习统计
		var totalExams int64
		db.Model(&models.ExamResult{}).Where("student_id = ?", studentID).Count(&totalExams)

		var Score struct {
			a float64
			b float64
			c int64
		}
		db.Model(&models.ExamResult{}).Where("student_id = ?", studentID).Select("COALESCE(AVG(score), 0) a, COALESCE(MAX(score), 0) b, sum(iif(score >= 60,1,0))").Scan(&Score)

		// 构建响应数据
		response := gin.H{
			"studentId":      student.StudentID,
			"name":           student.Name,
			"class":          classInfo.Name,
			"major":          student.Major,
			"enrollmentDate": classInfo.EnrollmentYear,
			"phone":          "13800138000", // 临时使用默认电话
			"stats": gin.H{
				"totalExams":   totalExams,
				"avgScore":     Score.a,
				"highestScore": Score.b,
				"passedExams":  Score.c,
			},
		}

		c.JSON(http.StatusOK, response)
	}
}

// GetStudents 获取所有学生列表
func GetStudents(db *gorm.DB) gin.HandlerFunc {
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
		query := db.Model(&models.User{}).Where("role = 1") // 1: student

		// 搜索条件
		if keyword != "" {
			search := "%" + keyword + "%"
			query = query.Where("name LIKE ? OR username LIKE ? OR student_id LIKE ?", search, search, search)
		}

		// 获取总数
		var total int64
		if err := query.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count students"})
			return
		}

		// 分页查询
		var students []models.User
		offset := (pageNum - 1) * pageSizeNum
		if err := query.Offset(offset).Limit(pageSizeNum).Find(&students).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch students"})
			return
		}

		// 构建响应数据
		var response []gin.H
		for _, student := range students {
			// 获取班级信息
			var classInfo models.Class
			db.First(&classInfo, student.ClassId)

			response = append(response, gin.H{
				"id":         student.ID,
				"username":   student.Username,
				"studentId":  student.StudentID,
				"name":       student.Name,
				"classId":    student.ClassId,
				"className":  classInfo.Name,
				"major":      student.Major,
				"createTime": student.CreatedAt,
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

// GetTeachers 获取教师列表（用于前端选择）
func GetTeachers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var teachers []models.User
		if err := db.Where("role = 2").Find(&teachers).Error; err != nil { // 2: teacher
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch teachers"})
			return
		}

		// 构建响应数据，统一字段格式
		var response []gin.H
		for _, teacher := range teachers {
			response = append(response, gin.H{
				"id":         teacher.ID,
				"username":   teacher.Username,
				"name":       teacher.Name,
				"role":       teacher.Role,
				"status":     teacher.Status,
				"createTime": teacher.CreatedAt,
			})
		}

		// 确保返回空数组而不是null
		if response == nil {
			response = []gin.H{}
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  response,
			"total": len(response),
		})
	}
}

// CreateStudent 创建新学生
func CreateStudent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			Username  string `json:"username" binding:"required"`
			StudentID string `json:"studentId"`
			Name      string `json:"name" binding:"required"`
			ClassId   int    `json:"classId" binding:"required"`
			Major     string `json:"major"`
			Phone     string `json:"phone"`
			Password  string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student data"})
			return
		}

		// 检查用户名是否已存在
		var existingUser models.User
		if err := db.Where("username = ? AND role = 1", request.Username).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
			return
		}

		// 检查学号是否已存在（如果提供了学号）
		if request.StudentID != "" {
			var existingStudent models.User
			if err := db.Where("student_id = ? AND role = 1", request.StudentID).First(&existingStudent).Error; err == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Student ID already exists"})
				return
			}
		}

		student := models.User{
			Username:  request.Username,
			Password:  models.HashPassword(request.Password),
			Role:      1, // 1: student
			Name:      request.Name,
			StudentID: request.StudentID,
			ClassId:   request.ClassId,
			Major:     request.Major,
			Phone:     request.Phone,
			Status:    1, // 学生状态默认为1
		}

		if err := db.Create(&student).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create student"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":        student.ID,
			"username":  student.Username,
			"studentId": student.StudentID,
			"name":      student.Name,
			"classId":   student.ClassId,
			"major":     student.Major,
			"phone":     student.Phone,
			"message":   "Student created successfully",
		})
	}
}

// UpdateStudent 更新学生信息
func UpdateStudent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var student models.User
		if err := db.Where("id = ? AND role = 1", id).First(&student).Error; err != nil { // 1: student
			c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
			return
		}

		var request struct {
			Name     string `json:"name"`
			ClassId  int    `json:"classId"`
			Major    string `json:"major"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student data"})
			return
		}

		if request.Name != "" {
			student.Name = request.Name
		}
		if request.ClassId != 0 {
			student.ClassId = request.ClassId
		}
		if request.Major != "" {
			student.Major = request.Major
		}
		if request.Password != "" {
			student.Password = models.HashPassword(request.Password)
		}

		if err := db.Save(&student).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update student"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Student updated successfully",
		})
	}
}

// DeleteStudent 删除学生
func DeleteStudent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var student models.User
		if err := db.Where("id = ? AND role = 1", id).First(&student).Error; err != nil { // 1: student
			c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
			return
		}

		if err := db.Delete(&student).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete student"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Student deleted successfully",
		})
	}
}

// ImportStudents 批量导入学生
func ImportStudents(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var students []models.User
		if err := c.ShouldBindJSON(&students); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student data"})
			return
		}

		// 验证数据
		for i, student := range students {
			if student.Username == "" || student.Name == "" || student.ClassId == 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Invalid student data",
					"message": "学生账号、姓名、班级不能为空",
					"index":   i,
				})
				return
			}

			// 检查用户名是否已存在
			var existingUser models.User
			if err := db.Where("username = ? AND role = 1", student.Username).First(&existingUser).Error; err == nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":    "Username already exists",
					"message":  "学生账号已存在",
					"username": student.Username,
					"index":    i,
				})
				return
			}

			// 检查学号是否已存在（如果提供了学号）
			if student.StudentID != "" {
				var existingStudent models.User
				if err := db.Where("student_id = ? AND role = 1", student.StudentID).First(&existingStudent).Error; err == nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"error":     "Student ID already exists",
						"message":   "学号已存在",
						"studentId": student.StudentID,
						"index":     i,
					})
					return
				}
			}
		}

		// 转换为学生用户格式
		var userStudents []models.User
		password := models.HashPassword("123456") // 默认密码
		for _, student := range students {
			userStudents = append(userStudents, models.User{
				Username:  student.Username,
				Password:  password, // 默认密码
				Role:      1,        // 1: student
				Name:      student.Name,
				StudentID: student.StudentID,
				ClassId:   student.ClassId,
				Major:     student.Major,
				Phone:     student.Phone,
				Status:    1, // 学生状态默认为1
			})
		}

		// 批量插入
		if err := db.Create(&userStudents).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import students"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Students imported successfully",
			"count":   len(userStudents),
		})
	}
}

// ExportStudents 导出学生列表
func ExportStudents(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var students []models.User
		if err := db.Where("role = 1").Find(&students).Error; err != nil { // 1: student
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch students"})
			return
		}

		// 转换为导出格式
		var exportData []gin.H
		for _, student := range students {
			// 获取班级信息
			var classInfo models.Class
			db.First(&classInfo, student.ClassId)

			exportData = append(exportData, gin.H{
				"username":   student.Username,
				"studentId":  student.StudentID,
				"name":       student.Name,
				"className":  classInfo.Name,
				"major":      student.Major,
				"phone":      student.Phone,
				"createTime": student.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}

		c.JSON(http.StatusOK, exportData)
	}
}

// calculateScore 计算考试分数
func calculateScore(answers []ans, examID uint, db *gorm.DB) int {
	var totalScore int

	// 获取试卷的所有题目
	var questions []models.Question
	if err := db.Where("exam_id = ?", examID).Find(&questions).Error; err != nil {
		return 0
	}

	// 创建题目ID到题目的映射
	questionMap := make(map[uint]models.Question)
	for _, q := range questions {
		questionMap[q.ID] = q
	}

	// 计算分数
	for _, answer := range answers {
		question, exists := questionMap[answer.QuestionID]
		if !exists {
			continue // 题目不存在
		}

		if question.Type == "single" {
			// 单选题评分
			if answer.Answer == question.Answer {
				totalScore += question.Score
			}
		} else if question.Type == "fill" {
			// 填空题评分
			if c := checkFillAnswerArray(answer.Answers, question.Answers); c > 0 {
				totalScore += question.Score * c / len(question.Answers)
			}
		}
	}

	return totalScore
}

// checkFillAnswerArray 检查填空题数组答案是否正确
func checkFillAnswerArray(studentAnswer []string, correctAnswers models.Answers) int {
	// 如果学生没有作答，返回false
	if len(studentAnswer) == 0 {
		return 0
	}

	var c int
	for k, answer := range studentAnswer {
		for _, correctAnswer := range correctAnswers[k].Options {
			if answer == correctAnswer {
				c++
				break
			}
		}
	}
	return c
}

// getExamStatus 根据开始时间和结束时间计算考试状态
func getExamStatus(startTime, endTime string) string {
	// 解析开始时间
	start, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		// 如果解析失败，尝试其他常见格式
		start, err = time.Parse("2006-01-02 15:04:05", startTime)
		if err != nil {
			return "未知"
		}
	}

	// 解析结束时间
	end, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		// 如果解析失败，尝试其他常见格式
		end, err = time.Parse("2006-01-02 15:04:05", endTime)
		if err != nil {
			return "未知"
		}
	}

	// 获取当前时间
	now := time.Now()

	// 比较时间
	if now.Before(start) {
		return "未开始"
	} else if now.After(end) {
		return "已结束"
	} else {
		return "进行中"
	}
}
