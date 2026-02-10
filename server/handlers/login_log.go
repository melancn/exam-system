package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"server/models"
)

// GetLoginLogs 获取登录日志
func GetLoginLogs(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 默认查询参数
		query := models.LoginLogQuery{
			Page:      1,
			PageSize:  20,
			EndTime:   time.Now(),
			StartTime: time.Now().AddDate(0, 0, -7), // 默认近一周
		}

		// 获取查询参数
		if username := c.Query("username"); username != "" {
			query.Username = username
		}

		if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 {
			query.Page = page
		}

		if pageSize, err := strconv.Atoi(c.Query("page_size")); err == nil && pageSize > 0 {
			query.PageSize = pageSize
		}

		if startTime := c.Query("start_time"); startTime != "" {
			if t, err := time.Parse("2006-01-02", startTime); err == nil {
				query.StartTime = t
			}
		}

		if endTime := c.Query("end_time"); endTime != "" {
			if t, err := time.Parse("2006-01-02", endTime); err == nil {
				query.EndTime = t
			}
		}

		// 构建查询条件
		var total int64
		var logs []models.LoginLog

		dbQuery := db.Model(&models.LoginLog{})

		// 添加查询条件
		if query.Username != "" {
			dbQuery = dbQuery.Where("username LIKE ?", "%"+query.Username+"%")
		}

		dbQuery = dbQuery.Where("created_at BETWEEN ? AND ?", query.StartTime, query.EndTime)

		// 获取总数
		dbQuery.Count(&total)

		// 获取分页数据
		offset := (query.Page - 1) * query.PageSize
		dbQuery.Order("created_at DESC").
			Limit(query.PageSize).
			Offset(offset).
			Find(&logs)

		response := models.LoginLogResponse{
			Total: int(total),
			Items: logs,
		}

		c.JSON(http.StatusOK, response)
	}
}
