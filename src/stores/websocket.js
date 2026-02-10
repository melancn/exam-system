import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'

export const useWebSocketStore = defineStore('websocket', () => {
  // WebSocket连接状态
  const isWebSocketConnected = ref(false)
  const isRealTimeConnected = ref(false)
  const ws = ref(null)
  const realTimeWs = ref(null)
  const retryCount = ref(0)
  const realTimeRetryCount = ref(0)
  
  // 实时数据
  const realTimeData = ref([])
  const realTimeExamData = ref([])
  
  // WebSocket配置
  const websocketConfig = {
    maxRetries: 5,
    retryInterval: 3000,
    timeout: 5000
  }
  
  const config = {
    websocket: {
      baseURL: 'ws://localhost:8080',
      endpoint: '/ws/exam-monitor',
      maxReconnectAttempts: 5,
      reconnectInterval: 3000
    }
  }
  
  // 计算属性
  const connectionStatus = computed(() => {
    return {
      realTime: isRealTimeConnected.value,
      data: isWebSocketConnected.value
    }
  })
  
  // 初始化实时时间显示WebSocket连接
  const initRealTimeDisplay = () => {
    const wsUrl = `ws://localhost:8080/api/exam-timer`
    
    try {
      ws.value = new WebSocket(wsUrl)

      ws.value.onopen = () => {
        console.log('实时时间WebSocket连接已建立')
        isWebSocketConnected.value = true
        retryCount.value = 0 // 重置重试次数
      }

      ws.value.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          console.log('收到实时时间数据:', data)
          updateRealTimeData(data)
        } catch (error) {
          console.error('解析WebSocket消息失败:', error)
        }
      }

      ws.value.onerror = (error) => {
        console.error('实时时间WebSocket错误:', error)
        isWebSocketConnected.value = false
      }

      ws.value.onclose = (event) => {
        console.log('实时时间WebSocket连接已关闭:', event.code, event.reason)
        isWebSocketConnected.value = false
        
        // 如果是意外断开且未达到最大重试次数，则尝试重连
        if (event.code !== 1000 && retryCount.value < websocketConfig.maxRetries) {
          retryCount.value++
          console.log(`WebSocket连接断开，${websocketConfig.retryInterval / 1000}秒后尝试重连(${retryCount.value}/${websocketConfig.maxRetries})`)
          
          setTimeout(() => {
            initRealTimeDisplay()
          }, websocketConfig.retryInterval)
        }
      }
    } catch (error) {
      console.error('创建WebSocket连接失败:', error)
      ElMessage.error('实时数据连接失败')
    }
  }
  
  // 初始化实时考试信息WebSocket连接
  const initRealTimeExamData = () => {
    const token = localStorage.getItem('token')
    if (!token) {
      ElMessage.error('请先登录')
      return
    }

    const wsUrl = `${config.websocket.baseURL}${config.websocket.endpoint}`
    
    try {
      realTimeWs.value = new WebSocket(wsUrl)

      realTimeWs.value.onopen = () => {
        console.log('实时考试信息WebSocket连接已建立')
        isRealTimeConnected.value = true
        realTimeRetryCount.value = 0 // 重置重试次数
        
        // 发送认证消息
        const authMessage = {
          type: 'auth',
          token: token
        }
        realTimeWs.value.send(JSON.stringify(authMessage))
      }

      realTimeWs.value.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          console.log('收到实时考试信息数据:', data)
          handleRealTimeMessage(data)
        } catch (error) {
          console.error('解析WebSocket消息失败:', error)
        }
      }

      realTimeWs.value.onerror = (error) => {
        console.error('实时考试信息WebSocket错误:', error)
        isRealTimeConnected.value = false
      }

      realTimeWs.value.onclose = (event) => {
        console.log('实时考试信息WebSocket连接已关闭:', event.code, event.reason)
        isRealTimeConnected.value = false
        
        // 如果是意外断开且未达到最大重试次数，则尝试重连
        if (event.code !== 1000 && realTimeRetryCount.value < config.websocket.maxReconnectAttempts) {
          realTimeRetryCount.value++
          console.log(`WebSocket连接断开，${config.websocket.reconnectInterval / 1000}秒后尝试重连(${realTimeRetryCount.value}/${config.websocket.maxReconnectAttempts})`)
          
          setTimeout(() => {
            initRealTimeExamData()
          }, config.websocket.reconnectInterval)
        }
      }
    } catch (error) {
      console.error('创建实时考试信息WebSocket连接失败:', error)
      ElMessage.error('实时数据连接失败')
    }
  }
  
  // 处理实时消息
  const handleRealTimeMessage = (data) => {
    if (!data || !data.type) return
    
    switch (data.type) {
      case 'auth_success':
        console.log('实时考试信息WebSocket认证成功')
        // 请求当前考试状态
        requestCurrentExamStatus()
        break
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
        updateStudentTime(data)
        break
      default:
        console.log('未知的实时消息类型:', data.type)
    }
  }
  
  // 请求当前考试状态
  const requestCurrentExamStatus = () => {
    if (realTimeWs.value && realTimeWs.value.readyState === WebSocket.OPEN) {
      const message = {
        type: 'get_exam_status',
        examId: 0 // 0表示获取所有考试状态
      }
      realTimeWs.value.send(JSON.stringify(message))
    }
  }
  
  // 更新实时考试数据
  const updateRealTimeExamData = (data) => {
    // 根据WebSocket推送的心跳数据更新考试信息
    const examData = {
      examId: data.examId,
      examTitle: `考试场次 ${data.examId}`,
      paperTitle: `试卷标题 ${data.examId}`,
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
      // 添加新学生到学生列表
      const newStudent = {
        studentId: data.studentId,
        studentName: `学生${data.studentId}`,
        className: '未知班级',
        timeUsed: 0,
        startTime: data.startTime,
        isActive: true
      }
      realTimeExamData.value[index].students.push(newStudent)
      // 重新计算在线人数
      realTimeExamData.value[index].onlineCount = realTimeExamData.value[index].students.filter(s => s.isActive).length
    } else {
      // 如果没有该考试，创建一个新的
      realTimeExamData.value.push({
        examId: data.examId,
        examTitle: `考试场次 ${data.examId}`,
        paperTitle: `试卷标题 ${data.examId}`,
        students: [{
          studentId: data.studentId,
          studentName: `学生${data.studentId}`,
          className: '未知班级',
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
      // 查找对应的学生并更新状态
      const studentIndex = realTimeExamData.value[index].students.findIndex(s => s.studentId === data.studentId)
      if (studentIndex >= 0) {
        realTimeExamData.value[index].students[studentIndex].isActive = false
        realTimeExamData.value[index].students[studentIndex].timeUsed = data.timeUsed
      }
      // 重新计算在线人数
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
  
  // 更新实时数据
  const updateRealTimeData = (data) => {
    if (!data || !data.type) return
    
    if (data.type === 'update') {
      // 检查数据完整性
      if (!data.examId || !data.studentId || typeof data.timeUsed !== 'number') {
        console.warn('实时数据不完整:', data)
        return
      }
      
      // 安全地更新实时数据数组
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
  }
  
  // 刷新实时数据
  const refreshRealTimeData = () => {
    requestCurrentExamStatus()
    ElMessage.success('实时数据刷新请求已发送')
  }
  
  // 广播消息功能
  const broadcastMessage = (messageData) => {
    if (!realTimeWs.value || realTimeWs.value.readyState !== WebSocket.OPEN) {
      ElMessage.error('WebSocket连接未建立，无法发送消息')
      return false
    }
    
    try {
      const message = {
        type: 'broadcast',
        ...messageData
      }
      realTimeWs.value.send(JSON.stringify(message))
      ElMessage.success('消息广播成功')
      return true
    } catch (error) {
      console.error('广播消息失败:', error)
      ElMessage.error('消息广播失败')
      return false
    }
  }
  
  // 关闭所有WebSocket连接
  const closeAllConnections = () => {
    if (ws.value) {
      ws.value.close()
    }
    if (realTimeWs.value) {
      realTimeWs.value.close()
    }
    isWebSocketConnected.value = false
    isRealTimeConnected.value = false
  }
  
  return {
    // 状态
    isWebSocketConnected,
    isRealTimeConnected,
    realTimeData,
    realTimeExamData,
    connectionStatus,
    
    // 方法
    initRealTimeDisplay,
    initRealTimeExamData,
    refreshRealTimeData,
    broadcastMessage,
    closeAllConnections
  }
})