# 考试系统 - Full Stack

一个完整的考试系统，包含前端Vue.js应用和后端Go Gin服务，支持学生端和教师端的完整考试流程。

## 项目结构

```
exam-system/
├── server/           # Go Gin后端服务
│   ├── main.go       # 主入口
│   ├── handlers/     # 路由处理程序
│   ├── models/       # 数据模型
│   ├── middlewares/  # 中间件
│   └── go.mod        # Go模块配置
├── src/              # Vue.js前端应用
│   ├── views/        # 页面组件
│   ├── layouts/      # 布局组件
│   ├── router/       # 路由配置
│   ├── stores/       # 状态管理
│   ├── services/     # API服务
│   └── utils/        # 工具函数
└── package.json      # 前端依赖配置
```

## 功能特性

### 学生端功能
- ✅ 个人信息管理
- ✅ 试卷列表查看
- ✅ 在线考试（单选+填空题）
- ✅ 考试成绩查询
- ✅ 成绩统计图表

### 教师端功能
- ✅ 学生账号管理
- ✅ 班级管理
- ✅ 试卷分配（设置考试时间、班级）
- ✅ 试卷编辑（添加单选、填空题）
- ✅ 考试结果分析
- ✅ 教师账号管理（管理员权限）
- ✅ 菜单权限管理
- ✅ 可折叠菜单布局

### 系统管理
- ✅ 权限控制（学生/教师/管理员）
- ✅ JWT认证
- ✅ 前后端API交互
- ✅ 响应式设计

## 技术栈

### 前端技术
- **Vue 3** - 渐进式JavaScript框架
- **Vite** - 快速构建工具
- **Element Plus** - UI组件库
- **Vue Router** - 路由管理
- **Pinia** - 状态管理
- **Axios** - HTTP客户端
- **ECharts** - 数据可视化

### 后端技术
- **Golang** - 编程语言
- **Gin** - Web框架
- **GORM** - ORM库
- **SQLite** - 数据库
- **JWT** - 身份验证

## 快速开始

### 前端启动
```bash
# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 访问地址：http://localhost:3000
```

### 后端启动
```bash
# 进入后端目录
cd server

# 下载依赖
go mod tidy

# 启动服务
go run main.go

# 服务地址：http://localhost:8080
```

## 演示账号

### 学生账号
- 账号：student001
- 密码：123456

### 教师账号
- 账号：teacher001
- 密码：123456
- 角色：普通教师

### 管理员账号
- 账号：admin
- 密码：admin123
- 角色：管理员

## API文档

后端服务启动后，所有API接口通过 `/api` 前缀访问：

- `POST /api/auth/login` - 用户登录
- `GET /api/teacher/students` - 获取学生列表
- `POST /api/teacher/students` - 添加学生
- `GET /api/teacher/exams` - 获取试卷列表
- `POST /api/teacher/exams` - 创建试卷
- `GET /api/student/exams` - 学生获取试卷
- `POST /api/student/exams/:id/submit` - 提交考试
- `GET /api/student/results` - 获取成绩

## 开发说明

### 前端开发
- 遵循Vue 3 Composition API规范
- 使用Element Plus组件库
- 支持响应式布局
- 模块化组件设计

### 后端开发
- RESTful API设计
- JWT身份验证
- GORM数据库操作
- 中间件权限控制

## 部署说明

### 前端部署
```bash
# 构建生产版本
npm run build

# 构建产物在 dist/ 目录
```

### 后端部署
```bash
# 编译可执行文件
go build -o exam-system main.go

# 运行二进制文件
./exam-system
```

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request来改善这个项目。

## 联系方式

如有问题请联系项目维护者。
```