import axios from 'axios'
import { ElMessage } from 'element-plus'

// 创建axios实例
const api = axios.create({
  baseURL: '/api',
  timeout: 10000
})

// 请求拦截器 - 添加token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器 - 处理错误
api.interceptors.response.use(
  (response) => {
    // 检查响应数据结构
    if (!response || typeof response !== 'object') {
      throw new Error('无效的响应格式')
    }
    return response.data
  },
  (error) => {
    // 统一错误处理
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      // 当前不是登录页面时才跳转
      if (window.location.pathname !== '/login') {
        window.location.href = '/login'
      }
    }
    ElMessage.error(error.response?.data?.error || '请求失败')
    return Promise.reject(error)
  }
)

// 认证相关API
export const authAPI = {
  login: (data) => api.post('/auth/login', data)
}

// 学生相关API
export const studentAPI = {
  getExams: (params = {}) => api.get('/student/exams', { params }),
  submitExam: (examId, data) => api.post(`/student/exams/${examId}/submit`, data),
  getResults: (params = {}) => api.get('/student/results', { params }),
  getProfile: () => api.get('/student/profile'),
  getExamResult: (resultId) => api.get(`/student/results/${resultId}`),
  getExamDetails: (examId) => api.get(`/student/exams/${examId}`)
}

// 教师相关API
export const teacherAPI = {
  // 学生管理
  getStudents: (params = {}) => api.get('/teacher/students', { params }),
  createStudent: (data) => api.post('/teacher/students', data),
  updateStudent: (id, data) => api.put(`/teacher/students/${id}`, data),
  deleteStudent: (id) => api.delete(`/teacher/students/${id}`),
  importStudents: (data) => api.post('/teacher/students/import', data),
  exportStudents: () => api.get('/teacher/students/export'),
  
  // 班级管理
  getClasses: (params = {}) => api.get('/teacher/classes', { params }),
  createClass: (data) => api.post('/teacher/classes', data),
  getClassDetails: (id) => api.get(`/teacher/classes/${id}`),
  updateClass: (id, data) => api.put(`/teacher/classes/${id}`, data),
  deleteClass: (id) => api.delete(`/teacher/classes/${id}`),
  getClassStudents: (id) => api.get(`/teacher/classes/${id}/students`),
  getClassStatistics: () => api.get('/teacher/classes/statistics'),
  importClasses: (data) => api.post('/teacher/classes/import', data),
  
  // 考试管理
  getExams: (params = {}) => api.get('/teacher/exams', { params }),
  createExam: (data) => api.post('/teacher/exams', data),
  getExamDetails: (id) => api.get(`/teacher/exams/${id}`),
  updateExam: (id, data) => api.put(`/teacher/exams/${id}`, data),
  updateExamStatus: (id, data) => api.put(`/teacher/exams/${id}/status`, data),
  deleteExam: (id) => api.delete(`/teacher/exams/${id}`),
  
  // 试卷分配
  getAssignments: () => api.get('/teacher/exam-assignments'),
  createAssignment: (data) => api.post('/teacher/exam-assignments', data),
  updateAssignment: (id, data) => api.put(`/teacher/exam-assignments/${id}`, data),
  deleteAssignment: (id) => api.delete(`/teacher/exam-assignments/${id}`),
  
  // 登录日志
  getLoginLogs: (params = {}) => api.get('/teacher/login-logs', { params }),
  
  // 结果分析
  getAnalysis: () => api.get('/teacher/results-analysis'),
  getExamResultsAnalysis: (params = {}) => api.get('/teacher/results-analysis/details', { params }),
  getScoreDistribution: (id, range = '10') => api.get(`/teacher/results-analysis/score-distribution/${id}`, { params: { range } }),
  getClassComparison: (id, metric = 'avg') => api.get(`/teacher/results-analysis/class-comparison/${id}`, { params: { metric } }),
  getExamDetail: (id) => api.get(`/teacher/results-analysis/exam-detail/${id}`),
  exportExamReport: () => api.get('/teacher/results-analysis/export'),
  
  // 管理员权限API
  getTeachers: () => api.get('/teacher/admin/accounts'),
  createTeacher: (data) => api.post('/teacher/admin/accounts', data),
  updateTeacher: (id, data) => api.put(`/teacher/admin/accounts/${id}`, data),
  deleteTeacher: (id) => api.delete(`/teacher/admin/accounts/${id}`)
}

export default api
