package models

import (
	"time"
)

// LoginLog 登录日志模型
type LoginLog struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	IP        string    `json:"ip" db:"ip"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// LoginLogQuery 查询参数
type LoginLogQuery struct {
	Username  string    `json:"username"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Page      int       `json:"page"`
	PageSize  int       `json:"page_size"`
}

// LoginLogResponse 查询响应
type LoginLogResponse struct {
	Total int        `json:"total"`
	Items []LoginLog `json:"items"`
}
