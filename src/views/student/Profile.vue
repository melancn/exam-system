<template>
  <div class="profile-page">
    <div class="page-header">
      <h2>个人信息</h2>
    </div>
    
    <div class="page-content">
      <el-card class="card-container">
        <template #header>
          <h3>基本信息</h3>
        </template>
        
        <el-descriptions :column="2" border>
          <el-descriptions-item label="学号">
            {{ studentInfo.studentId }}
          </el-descriptions-item>
          <el-descriptions-item label="姓名">
            {{ studentInfo.name }}
          </el-descriptions-item>
          <el-descriptions-item label="班级">
            {{ studentInfo.class }}
          </el-descriptions-item>
          <el-descriptions-item label="专业">
            {{ studentInfo.major }}
          </el-descriptions-item>
          <el-descriptions-item label="入学时间">
            {{ studentInfo.enrollmentDate }}
          </el-descriptions-item>
          <el-descriptions-item label="联系电话">
            {{ studentInfo.phone }}
          </el-descriptions-item>
        </el-descriptions>
      </el-card>
      
      <el-card class="card-container" style="margin-top: 20px;">
        <template #header>
          <h3>学习统计</h3>
        </template>
        
        <el-row :gutter="20">
          <el-col :span="6">
            <div class="stat-card">
              <div class="stat-number">{{ stats.totalExams }}</div>
              <div class="stat-label">已参加考试</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card">
              <div class="stat-number">{{ stats.avgScore }}</div>
              <div class="stat-label">平均分</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card">
              <div class="stat-number">{{ stats.highestScore }}</div>
              <div class="stat-label">最高分</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card">
              <div class="stat-number">{{ stats.passedExams }}</div>
              <div class="stat-label">及格考试</div>
            </div>
          </el-col>
        </el-row>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { studentAPI } from '@/services/api'

const userStore = useUserStore()

const studentInfo = ref({
  studentId: '',
  name: '',
  class: '',
  major: '',
  enrollmentDate: '',
  phone: ''
})

const stats = ref({
  totalExams: 0,
  avgScore: 0,
  highestScore: 0,
  passedExams: 0
})

const loading = ref(false)

// 获取真实用户信息
const fetchProfile = async () => {
  loading.value = true
  try {
    const data = await studentAPI.getProfile()
    studentInfo.value = {
      studentId: data.studentId || '',
      name: data.name || '',
      class: data.class || '',
      major: data.major || '',
      enrollmentDate: data.enrollmentDate || '',
      phone: data.phone || ''
    }
    stats.value = {
      totalExams: data.stats?.totalExams || 0,
      avgScore: data.stats?.avgScore || 0,
      highestScore: data.stats?.highestScore || 0,
      passedExams: data.stats?.passedExams || 0
    }
    
    // 更新用户存储
    if (!userStore.userInfo) {
      userStore.login({
        id: data.studentId,
        name: data.name,
        role: 1, // 1: student
        ...data
      })
    } else {
      // 更新现有用户信息
      userStore.updateUserInfo({
        studentId: data.studentId,
        name: data.name,
        ...data
      })
    }
  } catch (error) {
    console.error('获取用户信息失败:', error)
    // 使用默认数据
    studentInfo.value = {
      studentId: userStore.safeUserInfo.studentId || '',
      name: userStore.safeUserInfo.name || '',
      class: userStore.safeUserInfo.class || '',
      major: userStore.safeUserInfo.major || '',
      enrollmentDate: userStore.safeUserInfo.enrollmentDate || '',
      phone: userStore.safeUserInfo.phone || ''
    }
    stats.value = userStore.getStats()
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchProfile()
})
</script>

<style scoped>
.profile-page {
  min-height: 100%;
}

.stat-card {
  text-align: center;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
}

.stat-number {
  font-size: 32px;
  font-weight: bold;
  color: #409eff;
  margin-bottom: 8px;
}

.stat-label {
  color: #666;
  font-size: 14px;
}
</style>
