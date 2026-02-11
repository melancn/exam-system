package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetResultsAnalysis 获取结果分析数据
func GetResultsAnalysis(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取统计信息
		var totalExams int64
		db.Model(&models.Exam{}).Count(&totalExams)

		var totalStudents int64
		db.Model(&models.User{}).Where("role = 1").Count(&totalStudents) // 1: student

		var totalResults int64
		db.Model(&models.ExamResult{}).Count(&totalResults)

		// 计算平均分
		var avgScore float64
		db.Model(&models.ExamResult{}).Select("COALESCE(AVG(score), 0)").Scan(&avgScore)

		// 计算及格率
		var passCount int64
		db.Model(&models.ExamResult{}).Where("score >= 60").Count(&passCount)
		var passRate float64
		if totalResults > 0 {
			passRate = float64(passCount) / float64(totalResults) * 100
		}

		response := gin.H{
			"totalExams":    totalExams,
			"totalStudents": totalStudents,
			"totalResults":  totalResults,
			"avgScore":      avgScore,
			"passRate":      passRate,
		}

		c.JSON(http.StatusOK, response)
	}
}

// ExportExamReport 导出考试分析报告
func ExportExamReport(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取统计信息
		var totalExams int64
		db.Model(&models.Exam{}).Count(&totalExams)

		var totalStudents int64
		db.Model(&models.User{}).Where("role = 1").Count(&totalStudents) // 1: student

		var totalResults int64
		db.Model(&models.ExamResult{}).Count(&totalResults)

		// 计算平均分
		var avgScore float64
		db.Model(&models.ExamResult{}).Select("COALESCE(AVG(score), 0)").Scan(&avgScore)

		// 计算及格率
		var passCount int64
		db.Model(&models.ExamResult{}).Where("score >= 60").Count(&passCount)
		var passRate float64
		if totalResults > 0 {
			passRate = float64(passCount) / float64(totalResults) * 100
		}

		// 获取分数段分布
		var scoreDistribution struct {
			Excellent int64 `json:"excellent"` // 90分以上
			Good      int64 `json:"good"`      // 80-89分
			Pass      int64 `json:"pass"`      // 60-79分
			Fail      int64 `json:"fail"`      // 不及格(<60)
		}

		db.Model(&models.ExamResult{}).Where("score >= 90").Count(&scoreDistribution.Excellent)
		db.Model(&models.ExamResult{}).Where("score >= 80 AND score < 90").Count(&scoreDistribution.Good)
		db.Model(&models.ExamResult{}).Where("score >= 60 AND score < 80").Count(&scoreDistribution.Pass)
		db.Model(&models.ExamResult{}).Where("score < 60").Count(&scoreDistribution.Fail)

		// 获取题目分析
		var questions []models.Question
		db.Find(&questions)

		var questionAnalysis []gin.H
		for _, question := range questions {
			// 计算题目正确率
			var totalCount int64
			var correctCount int64

			// 获取该题目相关的考试结果
			var results []models.ExamResult
			db.Where("exam_id = ?", question.ExamID).Find(&results)

			totalCount = int64(len(results))

			// 计算正确数量
			for _, result := range results {
				var answers []string
				if err := json.Unmarshal([]byte(result.Answers), &answers); err == nil {
					if len(answers) > 0 && answers[0] == question.Answer {
						correctCount++
					}
				}
			}

			var correctRate float64
			if totalCount > 0 {
				correctRate = float64(correctCount) / float64(totalCount) * 100
			}

			questionAnalysis = append(questionAnalysis, gin.H{
				"id":          question.ID,
				"content":     question.Content,
				"type":        question.Type,
				"score":       question.Score,
				"answer":      question.Answer,
				"correctRate": correctRate,
			})
		}

		// 获取学生成绩
		var studentResults []gin.H
		var results []models.ExamResult
		db.Preload("ExamAssignment").Preload("ExamAssignment.Exam").Find(&results)

		for _, result := range results {
			// 获取学生信息
			var student models.User
			db.Where("id = ? AND role = 1", result.StudentID).First(&student) // 1: student

			studentResults = append(studentResults, gin.H{
				"studentId":   student.StudentID,
				"studentName": student.Name,
				"classId":     student.ClassId,
				"examTitle":   result.ExamAssignment.Exam.Title,
				"score":       result.Score,
				"totalScore":  result.ExamAssignment.Exam.TotalScore,
				"timeUsed":    result.TimeUsed,
				"passed":      result.Score >= 60,
			})
		}

		// 获取班级统计
		var classes []models.Class
		db.Find(&classes)

		var classStatistics []gin.H
		for _, class := range classes {
			var classResults []models.ExamResult
			db.Joins("JOIN users ON exam_results.student_id = users.id").
				Where("users.class_id = ? AND users.role = 1", class.ID). // 1: student
				Find(&classResults)

			var totalScore float64
			var passCount, excellentCount int64
			var studentCount int64

			for _, result := range classResults {
				totalScore += float64(result.Score)
				studentCount++

				if result.Score >= 60 {
					passCount++
				}
				if result.Score >= 90 {
					excellentCount++
				}
			}

			var avgScore float64
			var passRate, excellentRate float64
			if studentCount > 0 {
				avgScore = totalScore / float64(studentCount)
				passRate = float64(passCount) / float64(studentCount) * 100
				excellentRate = float64(excellentCount) / float64(studentCount) * 100
			}

			classStatistics = append(classStatistics, gin.H{
				"className":     class.Name,
				"studentCount":  studentCount,
				"avgScore":      avgScore,
				"passRate":      passRate,
				"excellentRate": excellentRate,
			})
		}

		// 构建报告数据
		reportData := gin.H{
			"stats": gin.H{
				"totalExams":    totalExams,
				"totalStudents": totalStudents,
				"totalResults":  totalResults,
				"avgScore":      avgScore,
				"passRate":      passRate,
			},
			"scoreDistribution": scoreDistribution,
			"questions":         questionAnalysis,
			"studentResults":    studentResults,
			"classStatistics":   classStatistics,
		}

		c.JSON(http.StatusOK, reportData)
	}
}

// GetExamResultsAnalysis 获取考试结果分析详情
func GetExamResultsAnalysis(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 分页参数
		page := c.DefaultQuery("page", "1")
		pageSize := c.DefaultQuery("pageSize", "10")
		examID := c.DefaultQuery("examId", "")
		classID := c.DefaultQuery("classId", "")
		scoreFilter := c.DefaultQuery("scoreFilter", "")
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
		query := db.Model(&models.ExamResult{}).
			Joins("JOIN exam_assignments ON exam_results.exam_assignment_id = exam_assignments.id").
			Joins("JOIN exams ON exam_assignments.exam_id = exams.id").
			Joins("JOIN users ON exam_results.student_id = users.id").
			Joins("JOIN classes ON users.class_id = classes.id").
			Select("exam_results.*, exams.title as exam_title, users.name as student_name, users.student_id, classes.name as class_name").
			Where("users.role = 1") // 1: student

		// 过滤条件
		if examID != "" {
			query = query.Where("exam_assignments.exam_id = ?", examID)
		}
		if classID != "" {
			query = query.Where("users.class_id = ?", classID)
		}
		if scoreFilter != "" {
			switch scoreFilter {
			case "excellent":
				query = query.Where("exam_results.score >= 90")
			case "good":
				query = query.Where("exam_results.score >= 80 AND exam_results.score < 90")
			case "pass":
				query = query.Where("exam_results.score >= 60 AND exam_results.score < 80")
			case "fail":
				query = query.Where("exam_results.score < 60")
			}
		}
		if keyword != "" {
			search := "%" + keyword + "%"
			query = query.Where("users.name LIKE ? OR users.student_id LIKE ?", search, search)
		}

		// 获取总数
		var total int64
		if err := query.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count results"})
			return
		}

		// 分页查询
		var results []gin.H
		offset := (pageNum - 1) * pageSizeNum
		if err := query.Offset(offset).Limit(pageSizeNum).Order("exam_results.score DESC").Scan(&results).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch results"})
			return
		}

		// 确保返回空数组而不是null
		if results == nil {
			results = []gin.H{}
		}

		c.JSON(http.StatusOK, gin.H{
			"data": results,
			"pagination": gin.H{
				"page":       pageNum,
				"pageSize":   pageSizeNum,
				"total":      total,
				"totalPages": (total + int64(pageSizeNum) - 1) / int64(pageSizeNum),
			},
		})
	}
}

// GetScoreDistribution 获取成绩分布数据
func GetScoreDistribution(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rangeSize := c.DefaultQuery("range", "10") // 5, 10, 20

		var rangeInt int
		if r, err := strconv.Atoi(rangeSize); err == nil && (r == 5 || r == 10 || r == 20) {
			rangeInt = r
		} else {
			rangeInt = 10
		}

		// 获取成绩分布
		var distributions []gin.H
		for i := 0; i < 100; i += rangeInt {
			start := i
			end := i + rangeInt - 1
			if end > 100 {
				end = 100
			}

			var count int64
			db.Model(&models.ExamResult{}).Where("score >= ? AND score <= ?", start, end).Count(&count)

			distributions = append(distributions, gin.H{
				"range": fmt.Sprintf("%d-%d", start, end),
				"count": count,
				"start": start,
				"end":   end,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"distribution": distributions,
			"range":        rangeInt,
		})
	}
}

// GetClassComparison 获取班级对比数据
func GetClassComparison(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		metric := c.DefaultQuery("metric", "avg") // avg, pass, excellent

		var classes []models.Class
		if err := db.Find(&classes).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch classes"})
			return
		}

		var comparisons []gin.H
		for _, class := range classes {
			var result gin.H

			switch metric {
			case "avg":
				var avgScore float64
				db.Model(&models.ExamResult{}).
					Joins("JOIN users ON exam_results.student_id = users.id").
					Where("users.class_id = ? AND users.role = 1", class.ID). // 1: student
					Select("COALESCE(AVG(score), 0)").
					Scan(&avgScore)
				result = gin.H{
					"classId":   class.ID,
					"className": class.Name,
					"value":     avgScore,
				}
			case "pass":
				var totalCount, passCount int64
				db.Model(&models.ExamResult{}).
					Joins("JOIN users ON exam_results.student_id = users.id").
					Where("users.class_id = ? AND users.role = 1", class.ID). // 1: student
					Count(&totalCount)
				db.Model(&models.ExamResult{}).
					Joins("JOIN users ON exam_results.student_id = users.id").
					Where("users.class_id = ? AND users.role = 1 AND score >= 60", class.ID). // 1: student
					Count(&passCount)
				var passRate float64
				if totalCount > 0 {
					passRate = float64(passCount) / float64(totalCount) * 100
				}
				result = gin.H{
					"classId":   class.ID,
					"className": class.Name,
					"value":     passRate,
				}
			case "excellent":
				var totalCount, excellentCount int64
				db.Model(&models.ExamResult{}).
					Joins("JOIN users ON exam_results.student_id = users.id").
					Where("users.class_id = ? AND users.role = 1", class.ID). // 1: student
					Count(&totalCount)
				db.Model(&models.ExamResult{}).
					Joins("JOIN users ON exam_results.student_id = users.id").
					Where("users.class_id = ? AND users.role = 1 AND score >= 90", class.ID). // 1: student
					Count(&excellentCount)
				var excellentRate float64
				if totalCount > 0 {
					excellentRate = float64(excellentCount) / float64(totalCount) * 100
				}
				result = gin.H{
					"classId":   class.ID,
					"className": class.Name,
					"value":     excellentRate,
				}
			default:
				result = gin.H{
					"classId":   class.ID,
					"className": class.Name,
					"value":     0,
				}
			}

			comparisons = append(comparisons, result)
		}

		c.JSON(http.StatusOK, gin.H{
			"comparisons": comparisons,
			"metric":      metric,
		})
	}
}

// GetExamDetail 获取考试详情和题目分析
func GetExamDetail(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		resultID := c.Param("id")

		var result models.ExamResult
		if err := db.Preload("ExamAssignment").First(&result, resultID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Result not found"})
			return
		}

		// 获取学生信息
		var student models.User
		db.Where("id = ? AND role = 1", result.StudentID).First(&student) // 1: student

		// 获取考试信息
		var exam models.Exam
		db.First(&exam, result.ExamAssignment.ExamID)

		// 获取题目信息
		var questions []models.Question
		db.Where("exam_id = ?", result.ExamAssignment.ExamID).Find(&questions)

		// 解析学生答案
		var studentAnswers []ans
		if err := json.Unmarshal([]byte(result.Answers), &studentAnswers); err != nil {
			studentAnswers = []ans{}
		}

		// 构建题目分析
		var questionDetails []gin.H
		for _, studentAnswer := range studentAnswers {
			// 找到对应的题目
			var question models.Question
			for _, q := range questions {
				if q.ID == studentAnswer.QuestionID {
					question = q
					break
				}
			}

			if question.ID == 0 {
				continue // 题目不存在
			}

			correct := false

			if question.Type == "single" {
				// 单选题：使用 Answer 字段
				correct = studentAnswer.Answer == question.Answer
			} else if question.Type == "fill" {
				// 填空题：使用 Answers 字段
				if len(studentAnswer.Answers) > 0 {
					for k, answer := range studentAnswer.Answers {
						for _, correctAnswer := range question.Answers[k].Options {
							if answer == correctAnswer {
								correct = true
								break
							}
						}
						if correct {
							break
						}
					}
				}
			}

			questionDetails = append(questionDetails, gin.H{
				"id":            question.ID,
				"content":       question.Content,
				"type":          question.Type,
				"score":         question.Score,
				"studentAnswer": studentAnswer,
				"correctAnswer": question.Answer,
				"correct":       correct,
			})
		}

		response := gin.H{
			"id":              result.ID,
			"examId":          result.ExamAssignment.ExamID,
			"examTitle":       exam.Title,
			"studentId":       student.StudentID,
			"studentName":     student.Name,
			"classId":         student.ClassId,
			"score":           result.Score,
			"totalScore":      exam.TotalScore,
			"timeUsed":        result.TimeUsed,
			"submitTime":      result.CreatedAt,
			"passed":          result.Score >= 60,
			"questionDetails": questionDetails,
		}

		c.JSON(http.StatusOK, response)
	}
}
