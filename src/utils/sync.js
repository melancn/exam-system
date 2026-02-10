// 数据同步工具
export class DataSync {
  constructor() {
    this.syncQueue = new Map()
    this.isOnline = navigator.onLine
    this.syncInterval = null
    
    // 监听网络状态变化
    window.addEventListener('online', () => {
      this.isOnline = true
      this.processSyncQueue()
    })
    
    window.addEventListener('offline', () => {
      this.isOnline = false
    })
  }

  // 保存数据到本地存储
  saveToLocal(key, data) {
    try {
      localStorage.setItem(key, JSON.stringify({
        data,
        timestamp: Date.now(),
        synced: false
      }))
      return true
    } catch (error) {
      console.error('保存到本地存储失败:', error)
      return false
    }
  }

  // 从本地存储读取数据
  loadFromLocal(key) {
    try {
      const item = localStorage.getItem(key)
      if (!item) return null
      
      const parsed = JSON.parse(item)
      return parsed.data
    } catch (error) {
      console.error('从本地存储读取数据失败:', error)
      return null
    }
  }

  // 添加到同步队列
  addToSyncQueue(key, data, apiCall) {
    this.syncQueue.set(key, {
      data,
      apiCall,
      timestamp: Date.now(),
      retries: 0,
      maxRetries: 3
    })
    
    // 如果在线，立即尝试同步
    if (this.isOnline) {
      this.processSyncQueue()
    }
  }

  // 处理同步队列
  async processSyncQueue() {
    if (!this.isOnline || this.syncQueue.size === 0) return

    const entries = Array.from(this.syncQueue.entries())
    
    for (const [key, item] of entries) {
      try {
        await item.apiCall(item.data)
        
        // 同步成功，从队列中移除
        this.syncQueue.delete(key)
        
        // 更新本地存储状态
        this.markAsSynced(key)
        
        console.log(`数据同步成功: ${key}`)
      } catch (error) {
        console.error(`数据同步失败: ${key}`, error)
        
        // 增加重试次数
        item.retries++
        
        // 如果达到最大重试次数，从队列中移除
        if (item.retries >= item.maxRetries) {
          this.syncQueue.delete(key)
          console.warn(`数据同步达到最大重试次数，已放弃: ${key}`)
        }
      }
    }
  }

  // 标记数据为已同步
  markAsSynced(key) {
    try {
      const item = localStorage.getItem(key)
      if (item) {
        const parsed = JSON.parse(item)
        parsed.synced = true
        localStorage.setItem(key, JSON.stringify(parsed))
      }
    } catch (error) {
      console.error('标记同步状态失败:', error)
    }
  }

  // 检查数据是否已同步
  isSynced(key) {
    try {
      const item = localStorage.getItem(key)
      if (!item) return false
      
      const parsed = JSON.parse(item)
      return parsed.synced || false
    } catch (error) {
      return false
    }
  }

  // 获取未同步的数据
  getUnsyncedData() {
    const unsynced = []
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i)
      if (key && key.startsWith('exam_') && !this.isSynced(key)) {
        const data = this.loadFromLocal(key)
        if (data) {
          unsynced.push({ key, data })
        }
      }
    }
    return unsynced
  }

  // 手动触发同步
  async syncNow() {
    if (!this.isOnline) {
      throw new Error('网络不可用')
    }
    
    await this.processSyncQueue()
  }

  // 清理过期的本地数据
  cleanupExpiredData(maxAge = 7 * 24 * 60 * 60 * 1000) { // 7天
    const now = Date.now()
    const keysToRemove = []
    
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i)
      if (key && key.startsWith('exam_')) {
        try {
          const item = localStorage.getItem(key)
          const parsed = JSON.parse(item)
          
          if (now - parsed.timestamp > maxAge) {
            keysToRemove.push(key)
          }
        } catch (error) {
          keysToRemove.push(key)
        }
      }
    }
    
    keysToRemove.forEach(key => {
      localStorage.removeItem(key)
      this.syncQueue.delete(key)
    })
    
    return keysToRemove.length
  }
}

// 创建全局同步实例
export const dataSync = new DataSync()

// 考试进度同步工具
export const examProgressSync = {
  // 保存考试进度
  saveProgress: (examId, answers) => {
    const key = `exam_${examId}_progress`
    const data = {
      examId,
      answers,
      timestamp: Date.now()
    }
    
    // 保存到本地
    dataSync.saveToLocal(key, data)
    
    // 添加到同步队列
    dataSync.addToSyncQueue(key, data, async (progressData) => {
      // 这里调用保存进度的API
      // 由于当前项目中没有保存进度的API，这里只是示例
      console.log('同步考试进度:', progressData)
      // await api.saveExamProgress(progressData)
    })
  },

  // 加载考试进度
  loadProgress: (examId) => {
    const key = `exam_${examId}_progress`
    return dataSync.loadFromLocal(key)
  },

  // 清除考试进度
  clearProgress: (examId) => {
    const key = `exam_${examId}_progress`
    localStorage.removeItem(key)
    dataSync.syncQueue.delete(key)
  },

  // 获取所有未同步的考试进度
  getUnsyncedProgress: () => {
    return dataSync.getUnsyncedData().filter(item => item.key.includes('progress'))
  }
}

// 用户数据同步工具
export const userDataSync = {
  // 保存用户设置
  saveSettings: (settings) => {
    const key = 'user_settings'
    dataSync.saveToLocal(key, settings)
    
    dataSync.addToSyncQueue(key, settings, async (settingsData) => {
      // 调用保存用户设置的API
      console.log('同步用户设置:', settingsData)
      // await api.saveUserSettings(settingsData)
    })
  },

  // 加载用户设置
  loadSettings: () => {
    const key = 'user_settings'
    return dataSync.loadFromLocal(key)
  }
}

// 导出同步状态钩子
export const useSyncStatus = () => {
  const isOnline = ref(dataSync.isOnline)
  const unsyncedCount = ref(dataSync.getUnsyncedData().length)

  // 监听网络状态变化
  const handleOnline = () => {
    isOnline.value = true
    dataSync.syncNow().then(() => {
      unsyncedCount.value = dataSync.getUnsyncedData().length
    })
  }

  const handleOffline = () => {
    isOnline.value = false
  }

  onMounted(() => {
    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)
  })

  onUnmounted(() => {
    window.removeEventListener('online', handleOnline)
    window.removeEventListener('offline', handleOffline)
  })

  const syncNow = async () => {
    try {
      await dataSync.syncNow()
      unsyncedCount.value = dataSync.getUnsyncedData().length
      return true
    } catch (error) {
      console.error('同步失败:', error)
      return false
    }
  }

  const cleanupExpired = () => {
    const removedCount = dataSync.cleanupExpiredData()
    unsyncedCount.value = dataSync.getUnsyncedData().length
    return removedCount
  }

  return {
    isOnline,
    unsyncedCount,
    syncNow,
    cleanupExpired
  }
}
