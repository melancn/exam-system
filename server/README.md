# 考试系统后端服务

基于Golang Gin框架和SQLite数据库的后端服务

## 功能特性

- 用户认证（JWT）
- 学生管理
- 教师管理（普通教师/管理员）
- 考试管理
- 成绩管理
- 权限控制

## 开发环境要求

- Go 1.21+
- SQLite3

## 安装与运行

1. 克隆仓库
2. 进入server目录：
   ```bash
   cd server
   ```
3. 下载依赖：
   ```bash
   go mod tidy
   ```
4. 运行服务：
   ```bash
   go run main.go
   ```

## API文档

服务启动后，访问 `http://localhost:8080/swagger/index.html` 查看API文档（待添加Swagger支持）

## 数据库

数据库文件将自动创建在项目根目录下：`exam.db`

## 环境变量

可以配置以下环境变量：

- `PORT` - 服务端口，默认8080
- `DB_PATH` - 数据库路径，默认`exam.db`
- `JWT_SECRET` - JWT密钥，默认随机生成

## 测试账号

- 管理员：admin / admin123
- 教师：teacher001 / teacher123
- 学生：student001 / student123
