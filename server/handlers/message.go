package handlers

import (
	"log"
	"net/http"
	"server/models"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MessageScheduler 定时消息调度器
type MessageScheduler struct {
	db       *gorm.DB
	mu       sync.Mutex
	ticker   *time.Ticker
	stopChan chan struct{}
}

var scheduler *MessageScheduler

// InitMessageScheduler 初始化消息调度器
func InitMessageScheduler(db *gorm.DB) {
	scheduler = &MessageScheduler{
		db:       db,
		ticker:   time.NewTicker(10 * time.Second), // 每10秒检查一次待发送消息
		stopChan: make(chan struct{}),
	}

	// 启动调度器
	go scheduler.run()

	// 立即执行一次检查，处理之前可能错过的消息
	go scheduler.checkPendingMessages()
}

// StopMessageScheduler 停止消息调度器
func StopMessageScheduler() {
	if scheduler != nil {
		scheduler.ticker.Stop()
		close(scheduler.stopChan)
	}
}

// run 运行调度器
func (s *MessageScheduler) run() {
	for {
		select {
		case <-s.ticker.C:
			s.checkPendingMessages()
		case <-s.stopChan:
			return
		}
	}
}

// checkPendingMessages 检查并发送待发送的消息
func (s *MessageScheduler) checkPendingMessages() {
	var messages []models.Message
	now := time.Now()

	// 查找所有待发送且到达发送时间的消息
	if err := s.db.Where("status = ? AND send_method = ? AND send_time <= ?", "pending", "scheduled", now).
		Find(&messages).Error; err != nil {
		log.Printf("Failed to fetch pending messages: %v", err)
		return
	}

	for _, msg := range messages {
		s.sendMessage(&msg)
	}
}

// sendMessage 发送消息
func (s *MessageScheduler) sendMessage(msg *models.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	sentAt := now

	// 构建WebSocket消息
	wsMessage := gin.H{
		"type":        "broadcast",
		"messageId":   msg.ID,
		"messageType": msg.MessageType,
		"title":       msg.Title,
		"content":     msg.Content,
		"timestamp":   now.Unix(),
	}

	// 根据目标类型发送消息
	var success bool
	if msg.TargetStudent > 0 {
		// 发送给指定学生
		if err := SendNotification(msg.TargetStudent, "student", wsMessage); err != nil {
			log.Printf("Failed to send message to student %d: %v", msg.TargetStudent, err)
			success = false
		} else {
			success = true
		}
	} else if msg.TargetExam > 0 {
		// 发送给指定考试的所有学生
		broadcastToStudents(msg.TargetExam, wsMessage)
		success = true
	} else {
		// 发送给所有学生
		broadcastToAllStudents(wsMessage)
		success = true
	}

	// 更新消息状态
	status := "sent"
	if !success {
		status = "failed"
	}

	if err := s.db.Model(msg).Updates(map[string]interface{}{
		"status":  status,
		"sent_at": &sentAt,
	}).Error; err != nil {
		log.Printf("Failed to update message status: %v", err)
	}

	log.Printf("Message %d sent: status=%s", msg.ID, status)
}

// broadcastToAllStudents 广播消息给所有学生
func broadcastToAllStudents(message gin.H) {
	if timerManager == nil {
		return
	}

	timerManager.mu.RLock()
	defer timerManager.mu.RUnlock()

	for connID, conn := range timerManager.connections {
		if isStudentConnection(connID) {
			if err := conn.WriteJSON(message); err != nil {
				log.Printf("Failed to send message to student %s: %v", connID, err)
			}
		}
	}
}

// CreateMessage 创建消息
func CreateMessage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			MessageType   string `json:"messageType" binding:"required"`
			TargetExam    uint   `json:"targetExam"`
			TargetClass   uint   `json:"targetClass"`
			TargetStudent uint   `json:"targetStudent"`
			Title         string `json:"title" binding:"required"`
			Content       string `json:"content" binding:"required"`
			SendMethod    string `json:"sendMethod" binding:"required"`
			SendTime      string `json:"sendTime"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 验证发送方式
		if req.SendMethod != "immediate" && req.SendMethod != "scheduled" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid send method"})
			return
		}

		// 如果是定时发送，验证发送时间
		var sendTime *time.Time
		if req.SendMethod == "scheduled" {
			if req.SendTime == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Send time is required for scheduled messages"})
				return
			}

			parsedTime, err := time.Parse("2006-01-02 15:04:05", req.SendTime)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid send time format"})
				return
			}

			if parsedTime.Before(time.Now()) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Send time must be in the future"})
				return
			}

			sendTime = &parsedTime
		}

		// 获取创建者ID（从JWT token中）
		createdBy, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		message := models.Message{
			MessageType:   req.MessageType,
			TargetExam:    req.TargetExam,
			TargetClass:   req.TargetClass,
			TargetStudent: req.TargetStudent,
			Title:         req.Title,
			Content:       req.Content,
			SendMethod:    req.SendMethod,
			SendTime:      sendTime,
			Status:        "pending",
			CreatedBy:     createdBy.(uint),
		}

		if err := db.Create(&message).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
			return
		}

		// 如果是立即发送，立即发送
		if req.SendMethod == "immediate" {
			if scheduler != nil {
				go scheduler.sendMessage(&message)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Message created successfully",
			"id":      message.ID,
		})
	}
}

// GetMessages 获取消息列表
func GetMessages(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取分页参数
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 10
		}

		// 获取筛选参数
		messageType := c.Query("messageType")
		status := c.Query("status")
		keyword := c.Query("keyword")

		// 构建查询
		query := db.Model(&models.Message{})

		if messageType != "" {
			query = query.Where("message_type = ?", messageType)
		}

		if status != "" {
			query = query.Where("status = ?", status)
		}

		if keyword != "" {
			query = query.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
		}

		// 获取总数
		var total int64
		if err := query.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count messages"})
			return
		}

		// 获取消息列表
		var messages []models.Message
		offset := (page - 1) * pageSize
		if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&messages).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"messages": messages,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		})
	}
}

// GetMessage 获取单个消息详情
func GetMessage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
			return
		}

		var message models.Message
		if err := db.First(&message, uint(id)).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": message,
		})
	}
}

// CancelMessage 取消消息
func CancelMessage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
			return
		}

		var message models.Message
		if err := db.First(&message, uint(id)).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
			return
		}

		// 只能取消待发送的消息
		if message.Status != "pending" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Can only cancel pending messages"})
			return
		}

		if err := db.Model(&message).Update("status", "cancelled").Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel message"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Message cancelled successfully",
		})
	}
}

// DeleteMessage 删除消息
func DeleteMessage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
			return
		}

		if err := db.Delete(&models.Message{}, uint(id)).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete message"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Message deleted successfully",
		})
	}
}
