// 配置文件
const config = {
  // API配置
  api: {
    baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
    timeout: 10000
  },
  
  // WebSocket配置
  websocket: {
    // 支持环境变量配置，便于不同环境使用
    baseURL: import.meta.env.VITE_WS_BASE_URL || 'ws://localhost:8080',
    endpoint: '/api/exam-timer',
    reconnectInterval: 5000, // 重连间隔
    maxReconnectAttempts: 5  // 最大重连次数
  },
  
  // 应用配置
  app: {
    title: import.meta.env.VITE_APP_TITLE || 'xxx学校考试系统',
    version: import.meta.env.VITE_APP_VERSION || '1.0.0',
    theme: {
      primaryColor: '#409eff',
      successColor: '#67c23a',
      warningColor: '#e6a23c',
      dangerColor: '#f56c6c'
    }
  }
}

export default config
