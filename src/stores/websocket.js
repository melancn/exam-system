import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage, ElNotification } from 'element-plus'
import config from '@/config'

export const useWebSocketStore = defineStore('websocket', () => {
  // WebSocket连接状态
  const isConnected = ref(false)
  const ws = ref(null)
  const retryCount = ref(0)
  let reconnectTimer = null
  
  // 实时数据
  const realTimeData = ref([])
  const realTimeExamData = ref([])
  
  // 消息处理器（用于组件注册自定义处理逻辑）
  const messageHandlers = new Map()
  
  // 计算属性
  const connectionStatus = computed(() => isConnected.value)
  
  // 初始化WebSocket连接（统一入口）
  const initWebSocket = () => {
    const token = localStorage.getItem('token')
    if (!token) {
      console.warn('未找到token，无法建立WebSocket连接')
      return
    }

    const wsUrl = `${config.websocket.baseURL}${config.websocket.endpoint}`
    console.log('连接WebSocket:', wsUrl)
    
    try {
      ws.value = new WebSocket(wsUrl)

      ws.value.onopen = () => {
        console.log('WebSocket连接已建立')
        isConnected.value = true
        retryCount.value = 0 // 重置重试次数
        
        // 发送认证消息
        const authMessage = {
          type: 'auth',
          token: token
        }
        ws.value.send(JSON.stringify(authMessage))
      }

      ws.value.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          console.log('收到WebSocket消息:', data)
          handleMessage(data)
        } catch (error) {
          console.error('解析WebSocket消息失败:', error)
        }
      }

      ws.value.onerror = (error) => {
        console.error('WebSocket错误:', error)
        isConnected.value = false
      }

      ws.value.onclose = (event) => {
        console.log('WebSocket连接已关闭:', event.code, event.reason)
        isConnected.value = false
        
        // 如果是意外断开且未达到最大重试次数，则尝试重连
        if (event.code !== 1000 && retryCount.value < config.websocket.maxReconnectAttempts) {
          retryCount.value++
          console.log(`WebSocket连接断开，${config.websocket.reconnectInterval / 1000}秒后尝试重连(${retryCount.value}/${config.websocket.maxReconnectAttempts})`)
          
          reconnectTimer = setTimeout(() => {
            initWebSocket()
          }, config.websocket.reconnectInterval)
        }
      }
    } catch (error) {
      console.error('创建WebSocket连接失败:', error)
      ElMessage.error('WebSocket连接失败')
    }
  }
  
  // 统一消息处理器 - 通过type区分消息类型
  const handleMessage = (data) => {
    if (!data || !data.type) return
    
    // 先调用所有注册的消息处理器
    messageHandlers.forEach(handler => {
      try {
        handler(data)
      } catch (error) {
        console.error('消息处理器执行失败:', error)
      }
    })
    
    // 然后处理默认消息类型
    switch (data.type) {
      // 认证相关
      case 'auth_success':
        console.log('WebSocket认证成功')
        // 认证成功后请求当前考试状态
        requestCurrentExamStatus()
        break
        
      // 实时时间相关
      case 'timer_update':
        handleTimerUpdate(data)
        break
        
      // 考试状态相关
      case 'exam_status':
        updateRealTimeExamData(data)
        break
      case 'student_start':
        updateStudentStart(data)
        break
      case 'student_end':
        updateStudentEnd(data)
        break
      case 'update':
        // 兼容旧的消息类型，处理学生时间更新
        updateStudentTime(data)
        break
        
      // 广播消息相关
      case 'broadcast':
        showNotification('广播通知', data.message, 'info')
        break
      case 'pause':
        showNotification('考试通知', '考试已被教师暂停', 'warning')
        break
      case 'resume':
        showNotification('考试通知', '考试已被教师恢复', 'success')
        break
        
      default:
        console.log('未知的消息类型:', data.type)
    }
  }
  
  // 处理定时器更新（实时时间显示）
  const handleTimerUpdate = (data) => {
    if (data.examId && data.studentId && typeof data.timeUsed !== 'number') {
      console.warn('定时器数据不完整:', data)
      return
    }
    
    // 更新实时数据数组
    const index = realTimeData.value.findIndex(item => 
      item.examId === data.examId && item.studentId === data.studentId
    )
    
    if (index >= 0) {
      realTimeData.value[index].timeUsed = data.timeUsed
    } else {
      realTimeData.value.push({
        examId: data.examId,
        studentId: data.studentId,
        timeUsed: data.timeUsed,
        startTime: data.startTime
      })
    }
  }
  
  // 请求当前考试状态
  const requestCurrentExamStatus = () => {
    if (ws.value && ws.value.readyState === WebSocket.OPEN) {
      const message = {
        type: 'get_exam_status',
        examId: 0 // 0表示获取所有考试状态
      }
      ws.value.send(JSON.stringify(message))
    }
  }
  
  // 更新实时考试数据
  const updateRealTimeExamData = (data) => {
    const examData = {
      examId: data.examId,
      examTitle: data.examTitle || `考试场次 ${data.examId}`,
      paperTitle: data.paperTitle || `试卷标题 ${data.examId}`,
      students: data.timers || []
    }
    
    // 计算当前在线人数（活跃状态的学生数量）
    examData.onlineCount = examData.students.filter(student => student.isActive).length
    
    // 更新或添加考试数据
    const index = realTimeExamData.value.findIndex(item => item.examId === examData.examId)
    if (index >= 0) {
      realTimeExamData.value[index] = examData
    } else {
      realTimeExamData.value.push(examData)
    }
  }
  
  // 更新学生开始考试
  const updateStudentStart = (data) => {
    const index = realTimeExamData.value.findIndex(item => item.examId === data.examId)
    if (index >= 0) {
      const newStudent = {
        studentId: data.studentId,
        studentName: data.studentName || `学生${data.studentId}`,
        className: data.className || '未知班级',
        timeUsed: 0,
        startTime: data.startTime,
        isActive: true
      }
      realTimeExamData.value[index].students.push(newStudent)
      realTimeExamData.value[index].onlineCount = realTimeExamData.value[index].students.filter(s => s.isActive).length
    } else {
      realTimeExamData.value.push({
        examId: data.examId,
        examTitle: data.examTitle || `考试场次 ${data.examId}`,
        paperTitle: data.paperTitle || `试卷标题 ${data.examId}`,
        students: [{
          studentId: data.studentId,
          studentName: data.studentName || `学生${data.studentId}`,
          className: data.className || '未知班级',
          timeUsed: 0,
          startTime: data.startTime,
          isActive: true
        }],
        onlineCount: 1
      })
    }
  }
  
  // 更新学生结束考试
  const updateStudentEnd = (data) => {
    const index = realTimeExamData.value.findIndex(item => item.examId === data.examId)
    if (index >= 0) {
      const studentIndex = realTimeExamData.value[index].students.findIndex(s => s.studentId === data.studentId)
      if (studentIndex >= 0) {
        realTimeExamData.value[index].students[studentIndex].isActive = false
        realTimeExamData.value[index].students[studentIndex].timeUsed = data.timeUsed
      }
      realTimeExamData.value[index].onlineCount = realTimeExamData.value[index].students.filter(s => s.isActive).length
    }
  }
  
  // 更新学生时间信息
  const updateStudentTime = (data) => {
    const index = realTimeExamData.value.findIndex(item => item.examId === data.examId)
    if (index >= 0) {
      const studentIndex = realTimeExamData.value[index].students.findIndex(s => s.studentId === data.studentId)
      if (studentIndex >= 0) {
        realTimeExamData.value[index].students[studentIndex].timeUsed = data.timeUsed
      }
    }
  }
  
  // 刷新实时数据
  const refreshRealTimeData = () => {
    requestCurrentExamStatus()
    ElMessage.success('实时数据刷新请求已发送')
  }
  
  // 广播消息功能
  const broadcastMessage = (messageData) => {
    if (!ws.value || ws.value.readyState !== WebSocket.OPEN) {
      ElMessage.error('WebSocket连接未建立，无法发送消息')
      return false
    }
    
    try {
      const message = {
        type: 'broadcast',
        ...messageData
      }
      ws.value.send(JSON.stringify(message))
      ElMessage.success('消息广播成功')
      return true
    } catch (error) {
      console.error('广播消息失败:', error)
      ElMessage.error('消息广播失败')
      return false
    }
  }
  
  // 发送暂停考试消息
  const pauseExam = (examId) => {
    if (!ws.value || ws.value.readyState !== WebSocket.OPEN) {
      ElMessage.error('WebSocket连接未建立')
      return false
    }
    
    try {
      const message = {
        type: 'pause',
        examId: examId
      }
      ws.value.send(JSON.stringify(message))
      ElMessage.success('暂停考试消息已发送')
      return true
    } catch (error) {
      console.error('发送暂停考试消息失败:', error)
      ElMessage.error('发送暂停考试消息失败')
      return false
    }
  }
  
  // 发送恢复考试消息
  const resumeExam = (examId) => {
    if (!ws.value || ws.value.readyState !== WebSocket.OPEN) {
      ElMessage.error('WebSocket连接未建立')
      return false
    }
    
    try {
      const message = {
        type: 'resume',
        examId: examId
      }
      ws.value.send(JSON.stringify(message))
      ElMessage.success('恢复考试消息已发送')
      return true
    } catch (error) {
      console.error('发送恢复考试消息失败:', error)
      ElMessage.error('发送恢复考试消息失败')
      return false
    }
  }
  
  // 显示通知
  const showNotification = (title, message, type = 'info') => {
    ElNotification({
      title,
      message,
      type,
      duration: 5000,
      position: 'top-right'
    })
  }
  
  // 关闭WebSocket连接
  const closeConnection = () => {
    if (ws.value) {
      ws.value.close(1000, '用户主动关闭')
      ws.value = null
    }
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    isConnected.value = false
    retryCount.value = 0
    messageHandlers.clear()
  }
  
  // 发送WebSocket消息
  const sendMessage = (message) => {
    if (ws.value && ws.value.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify(message))
      return true
    }
    console.warn('WebSocket未连接，无法发送消息')
    return false
  }
  
  // 注册消息处理器
  const registerMessageHandler = (key, handler) => {
    messageHandlers.set(key, handler)
  }
  
  // 移除消息处理器
  const removeMessageHandler = (key) => {
    messageHandlers.delete(key)
  }
  
  // 兼容性方法 - 保持向后兼容
  const initRealTimeDisplay = () => {
    initWebSocket()
  }
  
  const initRealTimeExamData = () => {
    initWebSocket()
  }
  
  const closeAllConnections = () => {
    closeConnection()
  }
  
  // 兼容性计算属性
  const isWebSocketConnected = computed(() => isConnected.value)
  const isRealTimeConnected = computed(() => isConnected.value)
  
  return {
    // 状态
    isConnected,
    isWebSocketConnected,
    isRealTimeConnected,
    realTimeData,
    realTimeExamData,
    connectionStatus,
    
    // 方法
    initWebSocket,
    initRealTimeDisplay,
    initRealTimeExamData,
    refreshRealTimeData,
    broadcastMessage,
    pauseExam,
    resumeExam,
    closeConnection,
    closeAllConnections,
    sendMessage,
    registerMessageHandler,
    removeMessageHandler
  }
})