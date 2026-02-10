import { ref, onUnmounted } from 'vue'
import { ElNotification } from 'element-plus'
import config from '@/config'

// 全局WebSocket状态
const isWebSocketConnected = ref(false)
let ws = null
let reconnectTimer = null
let reconnectAttempts = 0
let messageHandlers = new Map() // 存储消息处理器

// 初始化WebSocket连接
export const initWebSocket = () => {
  const token = localStorage.getItem('token')
  if (!token) {
    console.warn('未找到token，无法建立WebSocket连接')
    return
  }

  const wsUrl = `${config.websocket.baseURL}${config.websocket.endpoint}`
  console.log('连接WebSocket:', wsUrl)

  try {
    ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      console.log('WebSocket连接已建立')
      isWebSocketConnected.value = true
      reconnectAttempts = 0 // 重置重连次数
      
      // 发送认证消息
      ws.send(JSON.stringify({
        type: 'auth',
        token: token
      }))
    }

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        handleWebSocketMessage(data)
      } catch (error) {
        console.error('解析WebSocket消息失败:', error)
      }
    }

    ws.onerror = (error) => {
      console.error('WebSocket错误:', error)
      isWebSocketConnected.value = false
    }

    ws.onclose = () => {
      console.log('WebSocket连接已关闭')
      isWebSocketConnected.value = false
      // 尝试重连
      attemptReconnect()
    }
  } catch (error) {
    console.error('创建WebSocket连接失败:', error)
    attemptReconnect()
  }
}

// 处理WebSocket消息
const handleWebSocketMessage = (data) => {
  console.log('收到WebSocket消息:', data)
  
  // 调用所有注册的消息处理器
  messageHandlers.forEach(handler => {
    try {
      handler(data)
    } catch (error) {
      console.error('消息处理器执行失败:', error)
    }
  })
  
  // 默认处理通知消息
  switch (data.type) {
    case 'auth_success':
      console.log('WebSocket认证成功')
      break
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
      console.log('未知消息类型:', data.type)
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

// 尝试重连WebSocket
const attemptReconnect = () => {
  if (reconnectAttempts >= config.websocket.maxReconnectAttempts) {
    console.warn('达到最大重连次数，停止重连')
    return
  }

  reconnectAttempts++
  console.log(`尝试重连WebSocket (${reconnectAttempts}/${config.websocket.maxReconnectAttempts})`)

  reconnectTimer = setTimeout(() => {
    initWebSocket()
  }, config.websocket.reconnectInterval)
}

// 关闭WebSocket连接
export const closeWebSocket = () => {
  if (ws) {
    ws.close()
    ws = null
  }
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
  messageHandlers.clear()
}

// 发送WebSocket消息
export const sendWebSocketMessage = (message) => {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(message))
    return true
  }
  console.warn('WebSocket未连接，无法发送消息')
  return false
}

// 注册消息处理器
export const registerMessageHandler = (key, handler) => {
  messageHandlers.set(key, handler)
}

// 移除消息处理器
export const removeMessageHandler = (key) => {
  messageHandlers.delete(key)
}

// 获取WebSocket连接状态
export const getWebSocketStatus = () => isWebSocketConnected

// 导出WebSocket实例（用于特定场景）
export const getWebSocketInstance = () => ws

// 在组件卸载时自动清理
export const useWebSocketCleanup = () => {
  onUnmounted(() => {
    closeWebSocket()
  })
}