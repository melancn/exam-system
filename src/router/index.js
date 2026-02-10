import { createRouter, createWebHistory } from 'vue-router'

// 检查Token是否有效
const checkToken = () => {
  const token = localStorage.getItem('token')
  if (!token) {
    return false
  }
  
  // 简单的Token格式检查（JWT通常有3段）
  const tokenParts = token.split('.')
  if (tokenParts.length !== 3) {
    return false
  }
  
  // 检查Token是否过期（JWT的payload部分）
  try {
    const payload = JSON.parse(atob(tokenParts[1]))
    const currentTime = Math.floor(Date.now() / 1000)
    
    // 如果有exp字段且已过期
    if (payload.exp && payload.exp < currentTime) {
      return false
    }
    
    return true
  } catch (error) {
    return false
  }
}

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue')
  },
  {
    path: '/student',
    name: 'Student',
    component: () => import('@/layouts/StudentLayout.vue'),
    redirect: '/student/profile',
    children: [
      {
        path: 'profile',
        name: 'StudentProfile',
        component: () => import('@/views/student/Profile.vue')
      },
      {
        path: 'exam-list',
        name: 'StudentExamList',
        component: () => import('@/views/student/ExamList.vue')
      },
      {
        path: 'exam/:id',
        name: 'StudentExam',
        component: () => import('@/views/student/Exam.vue')
      },
      {
        path: 'exam-result/:resultId',
        name: 'StudentExamResult',
        component: () => import('@/views/student/ExamResult.vue')
      },
      {
        path: 'scores',
        name: 'StudentScores',
        component: () => import('@/views/student/Scores.vue')
      }
    ]
  },
  {
    path: '/teacher',
    name: 'Teacher',
    component: () => import('@/layouts/TeacherLayout.vue'),
    redirect: '/teacher/student-management',
    children: [
      {
        path: 'student-management',
        name: 'StudentManagement',
        component: () => import('@/views/teacher/StudentManagement.vue')
      },
      {
        path: 'class-management',
        name: 'ClassManagement',
        component: () => import('@/views/teacher/ClassManagement.vue')
      },
      {
        path: 'exam-assignment',
        name: 'ExamAssignment',
        component: () => import('@/views/teacher/ExamAssignment.vue')
      },
      {
        path: 'exam-list',
        name: 'TeacherExamList',
        component: () => import('@/views/teacher/ExamList.vue')
      },
      {
        path: 'exam-edit/:id?',
        name: 'ExamEdit',
        component: () => import('@/views/teacher/ExamEdit.vue')
      },
      {
        path: 'results-analysis',
        name: 'ResultsAnalysis',
        component: () => import('@/views/teacher/ResultsAnalysis.vue')
      },
      {
        path: 'broadcast-message',
        name: 'BroadcastMessage',
        component: () => import('@/views/teacher/BroadcastMessage.vue')
      },
      {
        path: 'login-logs',
        name: 'LoginLogs',
        component: () => import('@/views/teacher/LoginLogs.vue')
      },
      {
        path: 'account-management',
        name: 'AccountManagement',
        component: () => import('@/views/teacher/AccountManagement.vue'),
        meta: { requiresAdmin: true }
      },
      {
        path: 'menu-management',
        name: 'MenuManagement',
        component: () => import('@/views/teacher/MenuManagement.vue'),
        meta: { requiresAdmin: true }
      }
    ]
  },
  {
    path: '/',
    redirect: '/login'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 全局前置守卫
router.beforeEach((to, from, next) => {
  // 如果访问的是登录页面，直接放行
  if (to.path === '/login') {
    next()
    return
  }

  // 检查Token
  const hasValidToken = checkToken()

  if (!hasValidToken) {
    // Token不存在或已失效，重定向到登录页面
    next('/login')
    return
  }

  // Token有效，继续访问
  next()
})

export default router
