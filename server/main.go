package main

import (
	"log"
	"os"
	"server/handlers"
	"server/middlewares"
	"server/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	init := false
	if _, err := os.Stat("exam.db"); err != nil {
		init = true
	}
	// 初始化数据库
	db, err := gorm.Open(sqlite.Open("exam.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	// 自动迁移模型
	err = db.Debug().AutoMigrate(
		&models.User{},
		&models.Exam{},
		&models.Question{},
		&models.ExamResult{},
		&models.Class{},
		&models.ExamAssignment{},
		&models.LoginLog{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	if init {
		// 创建唯一索引
		err = db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_user_username_role ON users (username, role)").Error
		if err != nil {
			log.Fatal("Failed to create unique index:", err)
		}

		// 检查并创建默认管理员账号
		createDefaultAdmin(db)
	}

	// 初始化WebSocket管理器
	handlers.InitWebSocketManager(db)

	// 创建Gin路由
	r := gin.Default()

	// 设置路由
	setupRoutes(r, db)

	r.NoRoute(
		func(c *gin.Context) {
			path := "dist" + c.Request.URL.Path
			if i, err := os.Stat(path); err == nil && !i.IsDir() {
				log.Println(1)
				c.File(path)
			} else {
				log.Println(2, err, path)
				c.File("dist/index.html")
			}
		})

	// 启动服务
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// createDefaultAdmin 创建默认管理员账号
func createDefaultAdmin(db *gorm.DB) {
	var adminUser models.User
	result := db.Where("username = ?", "admin").First(&adminUser)

	// 如果不存在管理员账号，则创建
	if result.Error != nil {
		adminUser = models.User{
			Username: "admin",
			Password: models.HashPassword("admin"), // 使用SHA256加密密码
			Role:     2,                            // 2: teacher
			IsAdmin:  1,                            // 1: 是管理员
			Name:     "系统管理员",
		}

		if err := db.Create(&adminUser).Error; err != nil {
			log.Printf("Failed to create admin user: %v", err)
		} else {
			log.Println("Default admin account created: admin/admin")
		}
	}
}

func setupRoutes(r *gin.Engine, db *gorm.DB) {
	// 添加路由组
	api := r.Group("/api")
	{
		// WebSocket路由
		api.GET("/exam-timer", handlers.ExamTimerWebSocket(db))
		api.GET("/exam-status/:id", handlers.GetExamRealTimeStatus(db))
		api.GET("/student-history/:id", handlers.GetStudentExamHistory(db))

		// 认证路由
		auth := api.Group("/auth")
		{
			auth.POST("/login", handlers.Login(db))
		}

		// 学生路由
		student := api.Group("/student")
		student.Use(middlewares.AuthMiddleware()).Use(middlewares.StudentAuthMiddleware())
		{
			student.GET("/exams", handlers.GetStudentExams(db))
			student.GET("/exams/:id", handlers.GetExamDetailsForStudent(db))
			student.POST("/exams/:id/submit", handlers.SubmitExam(db))
			student.GET("/results", handlers.GetExamResults(db))
			student.GET("/profile", handlers.GetStudentProfile(db))
		}

		// 教师路由
		teacher := api.Group("/teacher")
		teacher.Use(middlewares.AuthMiddleware()).Use(middlewares.TeacherAuthMiddleware())
		{
			// 考试管理
			teacher.GET("/exams", handlers.GetExams(db))
			teacher.POST("/exams", handlers.CreateExam(db))
			teacher.GET("/exams/:id", handlers.GetExamDetails(db))
			teacher.PUT("/exams/:id", handlers.UpdateExam(db))
			teacher.PUT("/exams/:id/status", handlers.UpdateExamStatus(db))
			teacher.DELETE("/exams/:id", handlers.DeleteExam(db))

			// 试卷分配
			teacher.GET("/exam-assignments", handlers.GetAssignments(db))
			teacher.POST("/exam-assignments", handlers.CreateAssignment(db))
			teacher.PUT("/exam-assignments/:id", handlers.UpdateAssignment(db))
			teacher.DELETE("/exam-assignments/:id", handlers.DeleteAssignment(db))

			// 班级管理
			teacher.GET("/classes", handlers.GetClasses(db))
			teacher.POST("/classes", handlers.CreateClass(db))
			teacher.GET("/classes/:id", handlers.GetClassDetails(db))
			teacher.PUT("/classes/:id", handlers.UpdateClass(db))
			teacher.DELETE("/classes/:id", handlers.DeleteClass(db))
			teacher.GET("/classes/:id/students", handlers.GetClassStudents(db))
			teacher.GET("/classes/statistics", handlers.GetClassStatistics(db))

			// 学生管理
			teacher.GET("/students", handlers.GetStudents(db))
			teacher.POST("/students", handlers.CreateStudent(db))
			teacher.PUT("/students/:id", handlers.UpdateStudent(db))
			teacher.DELETE("/students/:id", handlers.DeleteStudent(db))
			teacher.POST("/students/import", handlers.ImportStudents(db))
			teacher.GET("/students/export", handlers.ExportStudents(db))

			// 结果分析
			teacher.GET("/results-analysis", handlers.GetResultsAnalysis(db))
			teacher.GET("/results-analysis/details", handlers.GetExamResultsAnalysis(db))
			teacher.GET("/results-analysis/score-distribution/:id", handlers.GetScoreDistribution(db))
			teacher.GET("/results-analysis/class-comparison/:id", handlers.GetClassComparison(db))
			teacher.GET("/results-analysis/exam-detail/:id", handlers.GetExamDetail(db))
			teacher.GET("/results-analysis/export", handlers.ExportExamReport(db))

			// 登录日志
			teacher.GET("/login-logs", handlers.GetLoginLogs(db))
		}

		// 管理员路由
		admin := api.Group("/teacher/admin")
		admin.Use(middlewares.AuthMiddleware())
		admin.Use(middlewares.TeacherAuthMiddleware())
		admin.Use(middlewares.AdminMiddleware())
		{
			// 教师账号管理
			admin.GET("/accounts", handlers.GetTeachers(db))
			admin.POST("/accounts", handlers.CreateTeacher(db))
			admin.PUT("/accounts/:id", handlers.UpdateTeacher(db))
			admin.DELETE("/accounts/:id", handlers.DeleteTeacher(db))
		}
	}
}
