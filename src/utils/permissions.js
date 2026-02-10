// 权限管理工具
export const permissions = {
  // 权限定义
  PERMISSIONS: {
    // 学生权限
    STUDENT_VIEW_EXAMS: 'student:view:exams',
    STUDENT_TAKE_EXAM: 'student:take:exam',
    STUDENT_VIEW_SCORES: 'student:view:scores',
    
    // 教师权限
    TEACHER_MANAGE_STUDENTS: 'teacher:manage:students',
    TEACHER_MANAGE_CLASSES: 'teacher:manage:classes',
    TEACHER_MANAGE_EXAMS: 'teacher:manage:exams',
    TEACHER_ASSIGN_EXAMS: 'teacher:assign:exams',
    TEACHER_VIEW_ANALYSIS: 'teacher:view:analysis',
    
    // 管理员权限
    ADMIN_MANAGE_TEACHERS: 'admin:manage:teachers',
    ADMIN_MANAGE_MENU: 'admin:manage:menu',
    ADMIN_SYSTEM_SETTINGS: 'admin:system:settings'
  },

  // 角色权限映射
  ROLE_PERMISSIONS: {
    student: [
      'student:view:exams',
      'student:take:exam',
      'student:view:scores'
    ],
    teacher: [
      'teacher:manage:students',
      'teacher:manage:classes',
      'teacher:manage:exams',
      'teacher:assign:exams',
      'teacher:view:analysis'
    ],
    admin: [
      'student:view:exams',
      'student:take:exam',
      'student:view:scores',
      'teacher:manage:students',
      'teacher:manage:classes',
      'teacher:manage:exams',
      'teacher:assign:exams',
      'teacher:view:analysis',
      'admin:manage:teachers',
      'admin:manage:menu',
      'admin:system:settings'
    ]
  },

  // 检查用户是否有指定权限
  hasPermission: (user, permission) => {
    if (!user || !user.role) {
      return false
    }

    const userPermissions = permissions.ROLE_PERMISSIONS[user.role] || []
    return userPermissions.includes(permission)
  },

  // 检查用户是否有任意一个权限
  hasAnyPermission: (user, permissionList) => {
    if (!user || !user.role) {
      return false
    }

    const userPermissions = permissions.ROLE_PERMISSIONS[user.role] || []
    return permissionList.some(permission => userPermissions.includes(permission))
  },

  // 检查用户是否有所有指定权限
  hasAllPermissions: (user, permissionList) => {
    if (!user || !user.role) {
      return false
    }

    const userPermissions = permissions.ROLE_PERMISSIONS[user.role] || []
    return permissionList.every(permission => userPermissions.includes(permission))
  },

  // 检查是否为管理员
  isAdmin: (user) => {
    return user && user.role === 'admin'
  },

  // 检查是否为教师
  isTeacher: (user) => {
    return user && (user.role === 'teacher' || user.role === 'admin')
  },

  // 检查是否为学生
  isStudent: (user) => {
    return user && user.role === 'student'
  },

  // 获取用户的所有权限
  getUserPermissions: (user) => {
    if (!user || !user.role) {
      return []
    }
    return permissions.ROLE_PERMISSIONS[user.role] || []
  },

  // 检查路由访问权限
  canAccessRoute: (user, routeName) => {
    // 定义路由权限映射
    const routePermissions = {
      'StudentExamList': 'student:view:exams',
      'StudentExam': 'student:take:exam',
      'StudentScores': 'student:view:scores',
      'StudentProfile': 'student:view:profile',
      
      'StudentManagement': 'teacher:manage:students',
      'ClassManagement': 'teacher:manage:classes',
      'ExamAssignment': 'teacher:assign:exams',
      'ExamList': 'teacher:manage:exams',
      'ExamEdit': 'teacher:manage:exams',
      'ResultsAnalysis': 'teacher:view:analysis',
      
      'AccountManagement': 'admin:manage:teachers',
      'MenuManagement': 'admin:manage:menu'
    }

    const requiredPermission = routePermissions[routeName]
    if (!requiredPermission) {
      return true // 无权限要求的路由
    }

    return permissions.hasPermission(user, requiredPermission)
  },

  // 检查按钮操作权限
  canPerformAction: (user, action) => {
    // 定义操作权限映射
    const actionPermissions = {
      'addStudent': 'teacher:manage:students',
      'editStudent': 'teacher:manage:students',
      'deleteStudent': 'teacher:manage:students',
      'importStudents': 'teacher:manage:students',
      'exportStudents': 'teacher:manage:students',
      
      'addClass': 'teacher:manage:classes',
      'editClass': 'teacher:manage:classes',
      'deleteClass': 'teacher:manage:classes',
      
      'createExam': 'teacher:manage:exams',
      'editExam': 'teacher:manage:exams',
      'deleteExam': 'teacher:manage:exams',
      'assignExam': 'teacher:assign:exams',
      
      'viewAnalysis': 'teacher:view:analysis',
      
      'addTeacher': 'admin:manage:teachers',
      'editTeacher': 'admin:manage:teachers',
      'deleteTeacher': 'admin:manage:teachers',
      'manageMenu': 'admin:manage:menu'
    }

    const requiredPermission = actionPermissions[action]
    if (!requiredPermission) {
      return true // 无权限要求的操作
    }

    return permissions.hasPermission(user, requiredPermission)
  }
}

// Vue 3 组合式函数 - 权限钩子
export const usePermissions = () => {
  const { safeUserInfo } = useUserStore()

  const hasPermission = (permission) => {
    return permissions.hasPermission(safeUserInfo.value, permission)
  }

  const hasAnyPermission = (permissionList) => {
    return permissions.hasAnyPermission(safeUserInfo.value, permissionList)
  }

  const hasAllPermissions = (permissionList) => {
    return permissions.hasAllPermissions(safeUserInfo.value, permissionList)
  }

  const isAdmin = () => {
    return permissions.isAdmin(safeUserInfo.value)
  }

  const isTeacher = () => {
    return permissions.isTeacher(safeUserInfo.value)
  }

  const isStudent = () => {
    return permissions.isStudent(safeUserInfo.value)
  }

  const canAccessRoute = (routeName) => {
    return permissions.canAccessRoute(safeUserInfo.value, routeName)
  }

  const canPerformAction = (action) => {
    return permissions.canPerformAction(safeUserInfo.value, action)
  }

  return {
    hasPermission,
    hasAnyPermission,
    hasAllPermissions,
    isAdmin,
    isTeacher,
    isStudent,
    canAccessRoute,
    canPerformAction
  }
}

// Vue 3 指令 - v-permission
export const permissionDirective = {
  mounted(el, binding) {
    const { safeUserInfo } = useUserStore()
    const permission = binding.value
    
    if (!permissions.hasPermission(safeUserInfo.value, permission)) {
      el.style.display = 'none'
    }
  },
  updated(el, binding) {
    const { safeUserInfo } = useUserStore()
    const permission = binding.value
    
    if (!permissions.hasPermission(safeUserInfo.value, permission)) {
      el.style.display = 'none'
    } else {
      el.style.display = ''
    }
  }
}

// Vue 3 指令 - v-role
export const roleDirective = {
  mounted(el, binding) {
    const { safeUserInfo } = useUserStore()
    const roles = Array.isArray(binding.value) ? binding.value : [binding.value]
    
    if (!roles.includes(safeUserInfo.value.role)) {
      el.style.display = 'none'
    }
  },
  updated(el, binding) {
    const { safeUserInfo } = useUserStore()
    const roles = Array.isArray(binding.value) ? binding.value : [binding.value]
    
    if (!roles.includes(safeUserInfo.value.role)) {
      el.style.display = 'none'
    } else {
      el.style.display = ''
    }
  }
}
