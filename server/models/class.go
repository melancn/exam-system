package models

import (
	"gorm.io/gorm"
)

// Class 班级模型
type Class struct {
	gorm.Model
	Name           string `gorm:"not null;unique"` // 班级名称
	Major          string `gorm:"not null"`        // 专业名称
	Description    string // 班级描述
	TeacherID      uint   `gorm:"not null"` // 所属教师ID
	EnrollmentYear int    // 入学年份
}

// ExamAssignment 试卷分配模型
type ExamAssignment struct {
	gorm.Model
	ExamID      uint   `json:"examId" gorm:"not null"`
	ClassID     uint   `json:"classId" gorm:"not null"`
	StartTime   string `json:"startTime" gorm:"not null"`
	EndTime     string `json:"endTime" gorm:"not null"`
	Duration    int    `json:"duration" gorm:"not null"` // 分钟
	PassScore   int    `json:"passScore" gorm:"not null"`
	Description string `json:"description"`

	Exam  Exam  `gorm:"foreignKey:ExamID"`
	Class Class `gorm:"foreignKey:ClassID"`
}
