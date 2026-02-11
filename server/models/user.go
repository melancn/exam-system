package models

import (
	"crypto/sha256"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username  string `gorm:"not null"`
	Password  string `gorm:"not null"`
	Role      byte   `gorm:"not null"`  // 1: student, 2: teacher
	IsAdmin   byte   `gorm:"default:0"` // 是否为管理员（仅对教师有效）
	Name      string
	Phone     string
	StudentID string // 学号，仅学生使用
	ClassId   int    `gorm:"default:0"` // 班级，仅学生使用
	Major     string // 专业，仅学生使用
	Status    byte   `gorm:"default:1"` // 状态，仅教师使用：active, inactive
}

// UniqueIndex 为 User 表添加唯一索引，确保用户名+角色的组合唯一
func (User) UniqueIndex() []string {
	return []string{"username", "role"}
}

// HashPassword 使用SHA256加密密码
func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	return HashPassword(password) == hash
}

type Exam struct {
	gorm.Model
	Title       string `gorm:"not null"`
	Description string
	Duration    int    // in minutes
	TotalScore  int    `gorm:"not null"`
	Status      string // draft, published, archived
	Questions   []Question
}

type Question struct {
	gorm.Model
	ExamID      uint    `gorm:"not null"`
	Type        string  `gorm:"not null"` // single, fill
	Content     string  `gorm:"not null"`
	Score       int     `gorm:"not null;default:0"`
	Options     Options `gorm:"type:text"` // JSON string for single choice questions
	Placeholder string  `gorm:"not null"`
	Answer      string  `gorm:"not null"`
	Answers     Answers `gorm:"type:text"` // JSON string for fill-in answers
}

type ExamResult struct {
	gorm.Model
	ExamAssignmentID uint   `gorm:"not null;default:0"`
	StudentID        uint   `gorm:"not null;default:0"`
	Score            int    `gorm:"not null;default:0"`
	Answers          string `gorm:"type:text"` // JSON string
	TimeUsed         int    // in minutes

	ExamAssignment ExamAssignment `gorm:"foreignKey:ExamAssignmentID"`
}

type Options []struct {
	Key  string `json:"key"`
	Text string `json:"text"`
}

type Answers []struct {
	Options []string `json:"options"`
	Type    string   `json:"type"`
}

func (j *Options) Scan(value interface{}) error {
	var bytes []byte
	switch v := value.(type) {
	case string:
		if len(v) == 0 {
			return nil
		}
		bytes = []byte(v)
	case []byte:
		if len(v) == 0 {
			return nil
		}
		bytes = v
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	return json.Unmarshal(bytes, &j)
}

func (j Options) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "[]", nil
	}
	return json.Marshal(j)
}

func (j *Answers) Scan(value interface{}) error {
	var bytes []byte
	switch v := value.(type) {
	case string:
		if len(v) == 0 {
			return nil
		}
		bytes = []byte(v)
	case []byte:
		if len(v) == 0 {
			return nil
		}
		bytes = v
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	return json.Unmarshal(bytes, &j)
}

func (j Answers) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "[]", nil
	}
	return json.Marshal(j)
}
