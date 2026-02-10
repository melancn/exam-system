import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useUserStore = defineStore('user', () => {
  const userInfo = ref(null)
  const safeUserInfo = computed(() => userInfo.value || {})
  const isLogin = computed(() => !!userInfo.value)
  const isTeacher = computed(() => safeUserInfo.value.role === 2) // 2: teacher
  const isAdmin = computed(() => safeUserInfo.value.role === 2 && safeUserInfo.value.isAdmin) // 2: teacher 且 IsAdmin为true

  const login = (userData) => {
    userInfo.value = userData
  }

  const logout = () => {
    userInfo.value = null
  }

  // 更新用户信息
  const updateUserInfo = (newData) => {
    if (userInfo.value) {
      // 深度合并用户信息，确保嵌套对象也被正确更新
      userInfo.value = {
        ...userInfo.value,
        ...newData,
        // 如果有stats数据，确保正确合并
        stats: newData.stats ? {
          ...userInfo.value.stats,
          ...newData.stats
        } : userInfo.value.stats
      }
    }
  }

  // 获取用户统计信息
  const getStats = () => {
    return safeUserInfo.value.stats || {
      totalExams: 0,
      avgScore: 0,
      highestScore: 0,
      passedExams: 0
    }
  }

  return {
    userInfo,
    safeUserInfo,
    isLogin,
    isTeacher,
    isAdmin,
    login,
    logout,
    updateUserInfo,
    getStats
  }
})
