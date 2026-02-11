package models

import "gorm.io/gorm"

// ExamTimer 考试计时器
type ExamTimer struct {
	gorm.Model
	ExamID    uint  `gorm:"not null;index:idx_exam_id" json:"examId"` // 考试ID
	StudentID uint  `gorm:"not null" json:"studentId"`                // 学生ID
	StartTime int64 `gorm:"not null" json:"startTime"`                // 开始时间
	TimeUsed  int   `gorm:"default:0" json:"timeUsed"`                // 已用时间
	IsActive  bool  `gorm:"default:true" json:"isActive"`             // 是否活跃
}

// TableName 指定表名为 exam_timers
func (ExamTimer) TableName() string {
	return "exam_timers"
}
