package handlers

import (
	"fmt"
	"log"
	"net/http"
	"server/models"
	"server/utils"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// WebSocket连接管理
type ExamTimerManager struct {
	connections map[string]*websocket.Conn
	mu          sync.RWMutex
	db          *gorm.DB
}

// TimerMessage 定义WebSocket消息结构
type TimerMessage struct {
	Type      string `json:"type"`      // 消息类型
	ExamID    uint   `json:"examId"`    // 考试ID
	StudentID uint   `json:"studentId"` // 学生ID
	TimeUsed  int    `json:"timeUsed"`  // 已用时间（秒）
	StartTime int64  `json:"startTime"` // 开始时间戳
	Message   string `json:"message"`   // 消息内容
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源
		},
	}
	timerManager *ExamTimerManager
)

// 初始化WebSocket管理器
func InitWebSocketManager(db *gorm.DB) {
	timerManager = &ExamTimerManager{
		connections: make(map[string]*websocket.Conn),
		db:          db,
	}
}

// WebSocket连接处理
func ExamTimerWebSocket(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 升级HTTP连接为WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("WebSocket upgrade failed: %v", err)
			return
		}
		defer conn.Close()

		// 处理WebSocket消息
		handleWebSocketMessages(conn, db)
	}
}

// 处理WebSocket消息
func handleWebSocketMessages(conn *websocket.Conn, db *gorm.DB) {
	var studentID uint
	var teacherID uint
	var connID string
	var authenticated bool
	var userType string // "student" 或 "teacher"

	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// 处理认证消息
		if !authenticated {
			if msgType, ok := msg["type"].(string); ok && msgType == "auth" {
				token, ok := msg["token"].(string)
				if !ok {
					conn.WriteJSON(gin.H{"error": "Invalid auth message"})
					continue
				}

				// 验证JWT token
				claims, err := utils.ParseToken(token)
				if err != nil {
					conn.WriteJSON(gin.H{"error": "Invalid token"})
					continue
				}

				// 设置用户ID和类型
				if claims.Role == 1 { // 1表示学生
					studentID = claims.UserID
					userType = "student"
					connID = generateConnectionID(studentID)
				} else if claims.Role == 2 { // 2表示教师
					teacherID = claims.UserID
					userType = "teacher"
					connID = generateTeacherConnectionID(teacherID)
				} else {
					conn.WriteJSON(gin.H{"error": "Unauthorized user role"})
					continue
				}

				authenticated = true

				// 添加到连接管理器
				timerManager.mu.Lock()
				timerManager.connections[connID] = conn
				timerManager.mu.Unlock()

				// 发送认证成功消息
				conn.WriteJSON(gin.H{
					"type":     "auth_success",
					"message":  "Authentication successful",
					"userId":   claims.UserID,
					"userType": userType,
					"role":     claims.Role,
				})

				log.Printf("%s %d authenticated successfully", userType, claims.UserID)
				continue
			}
			conn.WriteJSON(gin.H{"error": "Please authenticate first"})
			continue
		}

		// 处理认证后的消息
		_, ok := msg["type"].(string)
		if !ok {
			conn.WriteJSON(gin.H{"error": "Invalid message type"})
			continue
		}

		switch userType {
		case "student":
			handleStudentMessages(conn, msg, studentID, db)
		case "teacher":
			handleTeacherMessages(conn, msg, teacherID, db)
		default:
			conn.WriteJSON(gin.H{"error": "Unknown user type"})
		}
	}

	// 清理连接
	if authenticated && connID != "" {
		timerManager.mu.Lock()
		delete(timerManager.connections, connID)
		timerManager.mu.Unlock()
	}
}

// 处理学生消息
func handleStudentMessages(conn *websocket.Conn, msg map[string]interface{}, studentID uint, db *gorm.DB) {
	msgType, ok := msg["type"].(string)
	if !ok {
		conn.WriteJSON(gin.H{"error": "Invalid message type"})
		return
	}

	switch msgType {
	case "start":
		handleTimerStart(conn, msg, studentID, db)
	case "update":
		handleTimerUpdate(conn, msg, studentID, db)
	case "end":
		handleTimerEnd(conn, msg, studentID, db)
	default:
		conn.WriteJSON(gin.H{"error": "Unknown message type"})
	}
}

// 处理教师消息
func handleTeacherMessages(conn *websocket.Conn, msg map[string]interface{}, teacherID uint, db *gorm.DB) {
	msgType, ok := msg["type"].(string)
	if !ok {
		conn.WriteJSON(gin.H{"error": "Invalid message type"})
		return
	}

	switch msgType {
	case "get_exam_status":
		handleGetExamStatus(conn, msg, teacherID, db)
	case "get_student_status":
		handleGetStudentStatus(conn, msg, teacherID, db)
	case "broadcast":
		handleBroadcastMessage(conn, msg, teacherID, db)
	case "pause":
		handlePauseExam(conn, msg, teacherID, db)
	case "resume":
		handleResumeExam(conn, msg, teacherID, db)
	default:
		conn.WriteJSON(gin.H{"error": "Unknown message type"})
	}
}

// 处理获取考试状态
func handleGetExamStatus(conn *websocket.Conn, msg map[string]interface{}, teacherID uint, db *gorm.DB) {
	examID, ok := msg["examId"].(float64)
	if !ok {
		conn.WriteJSON(gin.H{"error": "Invalid examId"})
		return
	}

	var timers []models.ExamTimer
	if err := db.Where("exam_id = ? AND is_active = ?", uint(examID), true).Find(&timers).Error; err != nil {
		conn.WriteJSON(gin.H{"error": "Failed to get exam status"})
		return
	}

	// 计算每个学生的实时时间
	for i := range timers {
		currentTime := time.Now().Unix()
		timers[i].TimeUsed = int(currentTime - timers[i].StartTime)
	}

	conn.WriteJSON(gin.H{
		"type":    "exam_status",
		"examId":  uint(examID),
		"timers":  timers,
		"count":   len(timers),
		"message": "Exam status retrieved successfully",
	})
}

// 处理获取学生状态
func handleGetStudentStatus(conn *websocket.Conn, msg map[string]interface{}, teacherID uint, db *gorm.DB) {
	studentID, ok := msg["studentId"].(float64)
	if !ok {
		conn.WriteJSON(gin.H{"error": "Invalid studentId"})
		return
	}

	var timers []models.ExamTimer
	if err := db.Where("student_id = ?", uint(studentID)).Order("start_time DESC").Find(&timers).Error; err != nil {
		conn.WriteJSON(gin.H{"error": "Failed to get student status"})
		return
	}

	conn.WriteJSON(gin.H{
		"type":      "student_status",
		"studentId": uint(studentID),
		"timers":    timers,
		"count":     len(timers),
		"message":   "Student status retrieved successfully",
	})
}

// 处理广播消息
func handleBroadcastMessage(conn *websocket.Conn, msg map[string]interface{}, teacherID uint, db *gorm.DB) {
	examID, ok := msg["examId"].(float64)
	if !ok {
		conn.WriteJSON(gin.H{"error": "Invalid examId"})
		return
	}

	message, ok := msg["message"].(string)
	if !ok {
		conn.WriteJSON(gin.H{"error": "Invalid message"})
		return
	}

	// 广播消息给所有连接的学生
	broadcastToStudents(uint(examID), gin.H{
		"type":      "broadcast",
		"message":   message,
		"timestamp": time.Now().Unix(),
	})

	conn.WriteJSON(gin.H{
		"type":    "broadcast_ack",
		"examId":  uint(examID),
		"message": "Message broadcasted successfully",
	})
}

// 处理暂停考试
func handlePauseExam(conn *websocket.Conn, msg map[string]interface{}, teacherID uint, db *gorm.DB) {
	examID, ok := msg["examId"].(float64)
	if !ok {
		conn.WriteJSON(gin.H{"error": "Invalid examId"})
		return
	}

	// 更新数据库中的计时器状态
	if err := db.Model(&models.ExamTimer{}).Where("exam_id = ? AND is_active = ?", uint(examID), true).Update("is_active", false).Error; err != nil {
		conn.WriteJSON(gin.H{"error": "Failed to pause exam"})
		return
	}

	// 通知所有连接的学生考试已暂停
	broadcastToStudents(uint(examID), gin.H{
		"type":      "pause",
		"message":   "Exam paused by teacher",
		"timestamp": time.Now().Unix(),
	})

	conn.WriteJSON(gin.H{
		"type":    "pause_ack",
		"examId":  uint(examID),
		"message": "Exam paused successfully",
	})
}

// 处理恢复考试
func handleResumeExam(conn *websocket.Conn, msg map[string]interface{}, teacherID uint, db *gorm.DB) {
	examID, ok := msg["examId"].(float64)
	if !ok {
		conn.WriteJSON(gin.H{"error": "Invalid examId"})
		return
	}

	// 更新数据库中的计时器状态
	if err := db.Model(&models.ExamTimer{}).Where("exam_id = ? AND is_active = ?", uint(examID), false).Update("is_active", true).Error; err != nil {
		conn.WriteJSON(gin.H{"error": "Failed to resume exam"})
		return
	}

	// 通知所有连接的学生考试已恢复
	broadcastToStudents(uint(examID), gin.H{
		"type":      "resume",
		"message":   "Exam resumed by teacher",
		"timestamp": time.Now().Unix(),
	})

	conn.WriteJSON(gin.H{
		"type":    "resume_ack",
		"examId":  uint(examID),
		"message": "Exam resumed successfully",
	})
}

// 广播消息给指定考试的所有学生
func broadcastToStudents(examID uint, message gin.H) {
	timerManager.mu.RLock()
	defer timerManager.mu.RUnlock()

	for connID, conn := range timerManager.connections {
		if isStudentConnection(connID) {
			// 这里可以添加逻辑来检查学生是否参加指定的考试
			// 暂时广播给所有学生连接
			if err := conn.WriteJSON(message); err != nil {
				log.Printf("Failed to send message to student %s: %v", connID, err)
			}
		}
	}
}

// 广播消息给所有教师
func broadcastToTeachers(message gin.H) {
	timerManager.mu.RLock()
	defer timerManager.mu.RUnlock()

	for connID, conn := range timerManager.connections {
		if isTeacherConnection(connID) {
			if err := conn.WriteJSON(message); err != nil {
				log.Printf("Failed to send message to teacher %s: %v", connID, err)
			}
		}
	}
}

// 检查连接ID是否为学生连接
func isStudentConnection(connID string) bool {
	return len(connID) > 8 && connID[:8] == "student_"
}

// 检查连接ID是否为教师连接
func isTeacherConnection(connID string) bool {
	return len(connID) > 8 && connID[:8] == "teacher_"
}

// 生成教师连接ID
func generateTeacherConnectionID(teacherID uint) string {
	return "teacher_" + strconv.FormatUint(uint64(teacherID), 10)
}

// 生成连接ID
func generateConnectionID(studentID uint) string {
	return "student_" + strconv.FormatUint(uint64(studentID), 10)
}

// 处理计时开始
func handleTimerStart(conn *websocket.Conn, msg map[string]interface{}, studentID uint, db *gorm.DB) {
	examID, ok := msg["examId"].(float64)
	if !ok {
		conn.WriteJSON(gin.H{"error": "Invalid examId"})
		return
	}

	// 记录开始时间
	timer := models.ExamTimer{
		ExamID:    uint(examID),
		StudentID: studentID,
		StartTime: time.Now().Unix(),
		TimeUsed:  0,
		IsActive:  true,
	}

	// 保存到数据库
	if err := db.Create(&timer).Error; err != nil {
		conn.WriteJSON(gin.H{"error": "Failed to start timer"})
		return
	}

	// 发送确认消息
	conn.WriteJSON(gin.H{
		"type":      "start_ack",
		"examId":    uint(examID),
		"startTime": timer.StartTime,
		"message":   "Timer started successfully",
	})

	// 通知教师端有学生开始考试
	notifyTeacherStart(uint(examID), studentID, timer.StartTime)
}

// 处理计时更新
func handleTimerUpdate(conn *websocket.Conn, msg map[string]interface{}, studentID uint, db *gorm.DB) {
	examID, ok := msg["examId"].(float64)
	if !ok {
		conn.WriteJSON(gin.H{"error": "Invalid examId"})
		return
	}

	timeUsed, ok := msg["timeUsed"].(float64)
	if !ok {
		conn.WriteJSON(gin.H{"error": "Invalid timeUsed"})
		return
	}

	// 更新数据库中的时间
	var timer models.ExamTimer
	if err := db.Where("exam_id = ? AND student_id = ? AND is_active = ?",
		uint(examID), studentID, true).First(&timer).Error; err != nil {
		conn.WriteJSON(gin.H{"error": "Timer not found"})
		return
	}

	timer.TimeUsed = int(timeUsed)
	if err := db.Save(&timer).Error; err != nil {
		conn.WriteJSON(gin.H{"error": "Failed to update timer"})
		return
	}

	// 发送确认消息
	conn.WriteJSON(gin.H{
		"type":     "update_ack",
		"examId":   uint(examID),
		"timeUsed": timer.TimeUsed,
		"message":  "Timer updated successfully",
	})

	// 通知教师端时间更新
	notifyTeacherUpdate(uint(examID), studentID, timer.TimeUsed)
}

// 处理计时结束
func handleTimerEnd(conn *websocket.Conn, msg map[string]interface{}, studentID uint, db *gorm.DB) {
	examID, ok := msg["examId"].(float64)
	if !ok {
		conn.WriteJSON(gin.H{"error": "Invalid examId"})
		return
	}

	timeUsed, ok := msg["timeUsed"].(float64)
	if !ok {
		conn.WriteJSON(gin.H{"error": "Invalid timeUsed"})
		return
	}

	// 更新数据库
	var timer models.ExamTimer
	if err := db.Where("exam_id = ? AND student_id = ? AND is_active = ?",
		uint(examID), studentID, true).First(&timer).Error; err != nil {
		conn.WriteJSON(gin.H{"error": "Timer not found"})
		return
	}

	timer.TimeUsed = int(timeUsed)
	timer.IsActive = false
	if err := db.Save(&timer).Error; err != nil {
		conn.WriteJSON(gin.H{"error": "Failed to end timer"})
		return
	}

	// 发送确认消息
	conn.WriteJSON(gin.H{
		"type":     "end_ack",
		"examId":   uint(examID),
		"timeUsed": timer.TimeUsed,
		"message":  "Timer ended successfully",
	})

	// 通知教师端考试结束
	notifyTeacherEnd(uint(examID), studentID, timer.TimeUsed)
}

// 通知教师端学生开始考试
func notifyTeacherStart(examID, studentID uint, startTime int64) {
	// 构建通知消息
	message := gin.H{
		"type":      "student_start",
		"examId":    examID,
		"studentId": studentID,
		"startTime": startTime,
		"timestamp": time.Now().Unix(),
	}

	// 广播给所有教师连接
	broadcastToTeachers(message)

	log.Printf("Student %d started exam %d at %d", studentID, examID, startTime)
}

// 通知教师端时间更新
func notifyTeacherUpdate(examID, studentID uint, timeUsed int) {
	// 构建通知消息
	message := gin.H{
		"type":      "update",
		"examId":    examID,
		"studentId": studentID,
		"timeUsed":  timeUsed,
		"timestamp": time.Now().Unix(),
	}

	// 广播给所有教师连接
	broadcastToTeachers(message)

	log.Printf("Student %d exam %d time updated: %d seconds", studentID, examID, timeUsed)
}

// 通知教师端考试结束
func notifyTeacherEnd(examID, studentID uint, timeUsed int) {
	// 构建通知消息
	message := gin.H{
		"type":      "student_end",
		"examId":    examID,
		"studentId": studentID,
		"timeUsed":  timeUsed,
		"timestamp": time.Now().Unix(),
	}

	// 广播给所有教师连接
	broadcastToTeachers(message)

	log.Printf("Student %d finished exam %d in %d seconds", studentID, examID, timeUsed)
}

// 获取考试实时状态
func GetExamRealTimeStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		examID := c.Param("id")

		var timers []models.ExamTimer
		if err := db.Where("exam_id = ? AND is_active = ?", examID, true).Find(&timers).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get exam status"})
			return
		}

		// 计算每个学生的实时时间
		for i := range timers {
			currentTime := time.Now().Unix()
			timers[i].TimeUsed = int(currentTime - timers[i].StartTime)
		}

		c.JSON(http.StatusOK, gin.H{
			"examId": examID,
			"timers": timers,
			"count":  len(timers),
		})
	}
}

// 获取学生考试历史
func GetStudentExamHistory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		studentID := c.Param("id")

		var timers []models.ExamTimer
		if err := db.Where("student_id = ?", studentID).Order("start_time DESC").Find(&timers).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get exam history"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"studentId": studentID,
			"timers":    timers,
			"count":     len(timers),
		})
	}
}

// GetOnlineUsers 获取在线用户列表
func GetOnlineUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		timerManager.mu.RLock()
		defer timerManager.mu.RUnlock()

		var students []uint
		var teachers []uint

		for connID := range timerManager.connections {
			if isStudentConnection(connID) {
				// 从连接ID中提取学生ID
				idStr := connID[8:]
				if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
					students = append(students, uint(id))
				}
			} else if isTeacherConnection(connID) {
				// 从连接ID中提取教师ID
				idStr := connID[8:]
				if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
					teachers = append(teachers, uint(id))
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"students": students,
			"teachers": teachers,
			"total":    len(timerManager.connections),
		})
	}
}

// SendNotification 发送通知给指定用户
func SendNotification(userID uint, userType string, message gin.H) error {
	var connID string
	if userType == "student" {
		connID = generateConnectionID(userID)
	} else if userType == "teacher" {
		connID = generateTeacherConnectionID(userID)
	} else {
		return fmt.Errorf("invalid user type")
	}

	timerManager.mu.RLock()
	conn, exists := timerManager.connections[connID]
	timerManager.mu.RUnlock()

	if !exists {
		return fmt.Errorf("user not connected")
	}

	return conn.WriteJSON(message)
}
