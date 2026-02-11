package models

import (
	"time"
)

// Message 消息模型
type Message struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	MessageType   string     `json:"messageType" gorm:"type:varchar(50);not null"` // 消息类型
	TargetExam    uint       `json:"targetExam"`                                   // 目标考试ID
	TargetClass   uint       `json:"targetClass"`                                  // 目标班级ID
	TargetStudent uint       `json:"targetStudent"`                                // 目标学生ID（用于直接消息）
	Title         string     `json:"title" gorm:"type:varchar(200);not null"`
	Content       string     `json:"content" gorm:"type:text;not null"`
	SendMethod    string     `json:"sendMethod" gorm:"type:varchar(20);not null"`      // immediate 或 scheduled
	SendTime      *time.Time `json:"sendTime"`                                         // 发送时间（定时消息）
	Status        string     `json:"status" gorm:"type:varchar(20);default:'pending'"` // pending, sent, failed, cancelled
	CreatedBy     uint       `json:"createdBy"`                                        // 创建者ID
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	SentAt        *time.Time `json:"sentAt"` // 实际发送时间
}
