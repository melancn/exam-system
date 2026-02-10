<template>
  <div class="teacher-layout">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <h2>考试系统 - 教师端</h2>
          <div class="user-info">
            <div class="websocket-status">
              <el-tag :type="websocketStore.isWebSocketConnected ? 'success' : 'danger'" size="small">
                实时数据: {{ websocketStore.isWebSocketConnected ? '已连接' : '已断开' }}
              </el-tag>
              <el-tag :type="websocketStore.isRealTimeConnected ? 'success' : 'danger'" size="small">
                考试监控: {{ websocketStore.isRealTimeConnected ? '已连接' : '已断开' }}
              </el-tag>
            </div>
            <span>欢迎，{{ userStore.userInfo?.name }} 老师</span>
            <el-button type="primary" text @click="handleLogout">退出登录</el-button>
          </div>
        </div>
      </el-header>
      
      <el-container>
        <el-aside :width="isCollapse ? '64px' : '220px'" class="sidebar">
          <div class="collapse-button" @click="toggleCollapse">
            <el-icon>
              <component :is="isCollapse ? 'Expand' : 'Fold'" />
            </el-icon>
          </div>
          <el-menu
            :default-active="activeMenu"
            router
            class="sidebar-menu"
            :collapse="isCollapse"
          >
            <el-sub-menu index="1">
              <template #title>
                <el-icon><User /></el-icon>
                <span>学生管理</span>
              </template>
              <el-menu-item index="/teacher/student-management">
                <el-icon><User /></el-icon>
                <span>学生账号录入</span>
              </el-menu-item>
            </el-sub-menu>
            
            <el-sub-menu index="2">
              <template #title>
                <el-icon><Tickets /></el-icon>
                <span>考试管理</span>
              </template>
              <el-menu-item index="/teacher/exam-assignment">
                <el-icon><Tickets /></el-icon>
                <span>试卷分配</span>
              </el-menu-item>
              <el-menu-item index="/teacher/exam-edit">
                <el-icon><Edit /></el-icon>
                <span>试卷编辑</span>
              </el-menu-item>
              <el-menu-item index="/teacher/exam-list">
                <el-icon><Document /></el-icon>
                <span>试卷列表</span>
              </el-menu-item>
              <el-menu-item index="/teacher/results-analysis">
                <el-icon><DataAnalysis /></el-icon>
                <span>考试结果分析</span>
              </el-menu-item>
              <el-menu-item index="/teacher/broadcast-message">
                <el-icon><Bell /></el-icon>
                <span>广播消息</span>
              </el-menu-item>
              <el-menu-item index="/teacher/login-logs">
                <el-icon><Document /></el-icon>
                <span>登录日志</span>
              </el-menu-item>
            </el-sub-menu>
            
            <el-sub-menu index="3">
              <template #title>
                <el-icon><OfficeBuilding /></el-icon>
                <span>班级管理</span>
              </template>
              <el-menu-item index="/teacher/class-management">
                <el-icon><OfficeBuilding /></el-icon>
                <span>班级管理</span>
              </el-menu-item>
            </el-sub-menu>
            
            <el-sub-menu index="4" v-if="userStore.isAdmin">
              <template #title>
                <el-icon><Setting /></el-icon>
                <span>系统管理</span>
              </template>
              <el-menu-item index="/teacher/account-management">
                <el-icon><User /></el-icon>
                <span>教师账号管理</span>
              </el-menu-item>
              <el-menu-item index="/teacher/menu-management">
                <el-icon><Menu /></el-icon>
                <span>菜单权限管理</span>
              </el-menu-item>
            </el-sub-menu>
          </el-menu>
        </el-aside>
        
        <el-main class="main-content">
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useWebSocketStore } from '@/stores/websocket'
import { User, OfficeBuilding, Tickets, Document, Edit, DataAnalysis, Setting, Menu, Fold, Expand, Bell } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const websocketStore = useWebSocketStore()

const isCollapse = ref(false)
const activeMenu = computed(() => route.path)

const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value
}

const handleLogout = () => {
  userStore.logout()
  router.push('/login')
}

// 组件挂载时初始化WebSocket连接
onMounted(() => {
  // 初始化实时时间显示
  websocketStore.initRealTimeDisplay()
  // 初始化实时考试信息
  websocketStore.initRealTimeExamData()
})

// 组件卸载时关闭WebSocket连接
onUnmounted(() => {
  websocketStore.closeAllConnections()
})
</script>

<style scoped>
.teacher-layout {
  height: 100vh;
}

.header {
  background: linear-gradient(135deg, #2d5a27 0%, #4a7c59 100%);
  color: white;
  padding: 0 20px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.1);
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
}

.sidebar {
  background: linear-gradient(180deg, #f8f9fa 0%, #e9ecef 100%);
  position: relative;
  border-right: 1px solid #dee2e6;
}

.collapse-button {
  height: 48px;
  line-height: 48px;
  text-align: center;
  color: #495057;
  cursor: pointer;
  background: linear-gradient(180deg, #ffffff 0%, #f8f9fa 100%);
  border-bottom: 1px solid #dee2e6;
  transition: all 0.3s ease;
  font-weight: 500;
}

.collapse-button:hover {
  background: linear-gradient(180deg, #e9ecef 0%, #dee2e6 100%);
  color: #2d5a27;
}

.collapse-button .el-icon {
  font-size: 20px;
  vertical-align: middle;
  color: #2d5a27;
}

.sidebar-menu {
  border: none;
  background: transparent;
}

.sidebar-menu .el-menu-item {
  color: #495057;
  margin: 4px 8px;
  border-radius: 8px;
  transition: all 0.3s ease;
}

.sidebar-menu .el-menu-item:hover {
  background: linear-gradient(180deg, #d4edda 0%, #c3e6cb 100%);
  color: #155724;
  border-left: 3px solid #28a745;
}

.sidebar-menu .el-menu-item.is-active {
  background: linear-gradient(180deg, #d1ecf1 0%, #bee5eb 100%);
  color: #0c5460;
  border-left: 3px solid #17a2b8;
  font-weight: 600;
}

.sidebar-menu .el-sub-menu__title {
  color: #495057;
  font-weight: 500;
  margin: 4px 8px;
  border-radius: 8px;
  transition: all 0.3s ease;
}

.sidebar-menu .el-sub-menu__title:hover {
  background: linear-gradient(180deg, #fff3cd 0%, #ffeaa7 100%);
  color: #856404;
  border-left: 3px solid #ffc107;
}

.main-content {
  padding: 24px;
  background: linear-gradient(180deg, #ffffff 0%, #f8f9fa 100%);
  min-height: calc(100vh - 60px);
}

.websocket-status {
  display: flex;
  gap: 8px;
  align-items: center;
}

.websocket-status .el-tag {
  font-size: 12px;
}

/* 添加自然元素装饰 */
.sidebar::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, #28a745, #20c997, #17a2b8);
  opacity: 0.8;
}
</style>
