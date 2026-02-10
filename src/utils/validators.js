// 数据验证工具
export const validators = {
  // 验证学号格式
  validateStudentId: (studentId) => {
    if (!studentId || studentId.trim().length === 0) {
      return { valid: false, message: '学号不能为空' }
    }
    if (studentId.length < 6 || studentId.length > 20) {
      return { valid: false, message: '学号长度应在6-20位之间' }
    }
    // 学号通常为数字或字母数字组合
    const regex = /^[a-zA-Z0-9]+$/
    if (!regex.test(studentId)) {
      return { valid: false, message: '学号只能包含字母和数字' }
    }
    return { valid: true }
  },

  // 验证姓名
  validateName: (name) => {
    if (!name || name.trim().length === 0) {
      return { valid: false, message: '姓名不能为空' }
    }
    if (name.trim().length < 2 || name.trim().length > 20) {
      return { valid: false, message: '姓名长度应在2-20位之间' }
    }
    // 允许中文、英文、空格
    const regex = /^[\u4e00-\u9fa5a-zA-Z\s]+$/
    if (!regex.test(name.trim())) {
      return { valid: false, message: '姓名只能包含中文、英文和空格' }
    }
    return { valid: true }
  },

  // 验证班级名称
  validateClassName: (className) => {
    if (!className || className.trim().length === 0) {
      return { valid: false, message: '班级名称不能为空' }
    }
    if (className.trim().length > 50) {
      return { valid: false, message: '班级名称长度不能超过50位' }
    }
    return { valid: true }
  },

  // 验证专业名称
  validateMajor: (major) => {
    if (!major || major.trim().length === 0) {
      return { valid: false, message: '专业名称不能为空' }
    }
    if (major.trim().length > 50) {
      return { valid: false, message: '专业名称长度不能超过50位' }
    }
    return { valid: true }
  },

  // 验证教师姓名
  validateTeacherName: (name) => {
    if (!name || name.trim().length === 0) {
      return { valid: false, message: '教师姓名不能为空' }
    }
    if (name.trim().length < 2 || name.trim().length > 20) {
      return { valid: false, message: '教师姓名长度应在2-20位之间' }
    }
    const regex = /^[\u4e00-\u9fa5a-zA-Z\s]+$/
    if (!regex.test(name.trim())) {
      return { valid: false, message: '教师姓名只能包含中文、英文和空格' }
    }
    return { valid: true }
  },

  // 验证教师账号
  validateTeacherUsername: (username) => {
    if (!username || username.trim().length === 0) {
      return { valid: false, message: '登录账号不能为空' }
    }
    if (username.length < 3 || username.length > 20) {
      return { valid: false, message: '登录账号长度应在3-20位之间' }
    }
    const regex = /^[a-zA-Z0-9_]+$/
    if (!regex.test(username)) {
      return { valid: false, message: '登录账号只能包含字母、数字和下划线' }
    }
    return { valid: true }
  },

  // 验证密码
  validatePassword: (password) => {
    if (!password || password.length === 0) {
      return { valid: false, message: '密码不能为空' }
    }
    if (password.length < 6 || password.length > 20) {
      return { valid: false, message: '密码长度应在6-20位之间' }
    }
    // 至少包含字母和数字
    const hasLetter = /[a-zA-Z]/.test(password)
    const hasNumber = /[0-9]/.test(password)
    if (!hasLetter || !hasNumber) {
      return { valid: false, message: '密码必须包含字母和数字' }
    }
    return { valid: true }
  },

  // 验证试卷标题
  validateExamTitle: (title) => {
    if (!title || title.trim().length === 0) {
      return { valid: false, message: '试卷标题不能为空' }
    }
    if (title.trim().length > 100) {
      return { valid: false, message: '试卷标题长度不能超过100位' }
    }
    return { valid: true }
  },

  // 验证总分
  validateTotalScore: (score) => {
    const num = parseInt(score)
    if (isNaN(num)) {
      return { valid: false, message: '总分必须是数字' }
    }
    if (num < 10 || num > 200) {
      return { valid: false, message: '总分应在10-200分之间' }
    }
    return { valid: true }
  },

  // 验证题目分数
  validateQuestionScore: (score) => {
    const num = parseInt(score)
    if (isNaN(num)) {
      return { valid: false, message: '分数必须是数字' }
    }
    if (num < 1 || num > 50) {
      return { valid: false, message: '分数应在1-50分之间' }
    }
    return { valid: true }
  },

  // 验证题目内容
  validateQuestionContent: (content) => {
    if (!content || content.trim().length === 0) {
      return { valid: false, message: '题目内容不能为空' }
    }
    if (content.trim().length > 1000) {
      return { valid: false, message: '题目内容长度不能超过1000字' }
    }
    return { valid: true }
  },

  // 验证选项内容
  validateOptionContent: (content) => {
    if (!content || content.trim().length === 0) {
      return { valid: false, message: '选项内容不能为空' }
    }
    if (content.trim().length > 200) {
      return { valid: false, message: '选项内容长度不能超过200字' }
    }
    return { valid: true }
  },

  // 验证时间格式
  validateTime: (time) => {
    if (!time) {
      return { valid: false, message: '时间不能为空' }
    }
    // 检查是否为有效的日期时间格式
    const date = new Date(time)
    if (isNaN(date.getTime())) {
      return { valid: false, message: '时间格式无效' }
    }
    return { valid: true }
  },

  // 验证时间范围
  validateTimeRange: (startTime, endTime) => {
    const start = new Date(startTime)
    const end = new Date(endTime)
    
    if (isNaN(start.getTime()) || isNaN(end.getTime())) {
      return { valid: false, message: '时间格式无效' }
    }
    
    if (start >= end) {
      return { valid: false, message: '结束时间必须大于开始时间' }
    }
    
    // 检查时间是否在合理范围内（不能是过去的时间）
    const now = new Date()
    if (start < now) {
      return { valid: false, message: '开始时间不能是过去的时间' }
    }
    
    return { valid: true }
  },

  // 验证考试时长
  validateDuration: (duration) => {
    const num = parseInt(duration)
    if (isNaN(num)) {
      return { valid: false, message: '时长必须是数字' }
    }
    if (num < 10 || num > 300) {
      return { valid: false, message: '考试时长应在10-300分钟之间' }
    }
    return { valid: true }
  },

  // 验证及格分数
  validatePassScore: (score, totalScore) => {
    const num = parseInt(score)
    const total = parseInt(totalScore)
    
    if (isNaN(num)) {
      return { valid: false, message: '及格分数必须是数字' }
    }
    
    if (isNaN(total)) {
      return { valid: false, message: '总分无效' }
    }
    
    if (num < 0 || num > total) {
      return { valid: false, message: `及格分数应在0-${total}分之间` }
    }
    
    return { valid: true }
  },

  // 验证联系电话
  validatePhone: (phone) => {
    if (!phone || phone.trim().length === 0) {
      return { valid: false, message: '联系电话不能为空' }
    }
    
    // 移除所有空格和特殊字符进行验证
    const cleanPhone = phone.replace(/[\s-()]/g, '')
    
    // 检查是否为纯数字
    if (!/^\d+$/.test(cleanPhone)) {
      return { valid: false, message: '联系电话只能包含数字' }
    }
    
    // 检查长度（中国大陆手机号11位，固定电话7-8位）
    if (cleanPhone.length !== 11 && cleanPhone.length < 7) {
      return { valid: false, message: '联系电话应为11位手机号或7-8位固定电话' }
    }
    
    // 如果是11位，检查是否为有效的手机号格式（1开头）
    if (cleanPhone.length === 11 && !/^1[3-9]\d{9}$/.test(cleanPhone)) {
      return { valid: false, message: '手机号格式不正确' }
    }
    
    return { valid: true }
  }
}

// 批量验证
export const validateBatch = (data, rules) => {
  const errors = []
  
  for (const [field, value] of Object.entries(data)) {
    if (rules[field]) {
      const rule = rules[field]
      let result
      
      if (typeof rule === 'function') {
        result = rule(value)
      } else if (rule.validator) {
        result = rule.validator(value, data)
      }
      
      if (!result.valid) {
        errors.push({
          field,
          message: rule.message || result.message
        })
      }
    }
  }
  
  return {
    valid: errors.length === 0,
    errors
  }
}

// 表单验证规则生成器
export const createFormRules = (rules) => {
  const formRules = {}
  
  for (const [field, rule] of Object.entries(rules)) {
    formRules[field] = [
      {
        validator: (rule, value, callback) => {
          const result = rule.validator(value)
          if (!result.valid) {
            callback(new Error(result.message))
          } else {
            callback()
          }
        },
        trigger: rule.trigger || 'blur'
      }
    ]
  }
  
  return formRules
}
