<template>
  <div class="student-layout">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <el-button 
            class="mobile-menu-btn" 
            @click="toggleSidebar"
            v-if="isMobile"
          >
            <el-icon><Menu /></el-icon>
          </el-button>
          <h2>考试系统 - 学生端</h2>
          <div class="user-info">
            <span>欢迎，{{ userStore.userInfo?.name }}</span>
            <el-button type="primary" text @click="handleLogout">退出登录</el-button>
          </div>
        </div>
      </el-header>
      
      <el-container>
        <el-aside 
          width="200px" 
          class="sidebar"
          :class="{ 'show': isSidebarOpen }"
        >
          <el-menu
            :default-active="activeMenu"
            router
            class="sidebar-menu"
            @select="handleMenuSelect"
          >
            <el-menu-item index="/student/profile">
              <el-icon><User /></el-icon>
              <span>个人信息</span>
            </el-menu-item>
            <el-menu-item index="/student/exam-list">
              <el-icon><Document /></el-icon>
              <span>试卷列表</span>
            </el-menu-item>
            <el-menu-item index="/student/scores">
              <el-icon><DataAnalysis /></el-icon>
              <span>考试成绩</span>
            </el-menu-item>
          </el-menu>
        </el-aside>
        
        <el-main class="main-content">
          <router-view />
        </el-main>
      </el-container>
      
      <!-- 移动端遮罩层 -->
      <div 
        class="sidebar-overlay" 
        v-if="isMobile && isSidebarOpen"
        @click="toggleSidebar"
      ></div>
    </el-container>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { User, Document, DataAnalysis, Menu } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const activeMenu = computed(() => route.path)
const isSidebarOpen = ref(false)
const isMobile = ref(false)

// 检测屏幕尺寸
const checkScreenSize = () => {
  isMobile.value = window.innerWidth <= 768
  // 移动端默认关闭侧边栏
  if (isMobile.value) {
    isSidebarOpen.value = false
  }
}

// 切换侧边栏
const toggleSidebar = () => {
  isSidebarOpen.value = !isSidebarOpen.value
}

// 菜单选择处理
const handleMenuSelect = () => {
  // 移动端点击菜单项后关闭侧边栏
  if (isMobile.value) {
    isSidebarOpen.value = false
  }
}

// 退出登录
const handleLogout = () => {
  userStore.logout()
  router.push('/login')
}

// 监听窗口大小变化
onMounted(() => {
  checkScreenSize()
  window.addEventListener('resize', checkScreenSize)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkScreenSize)
})
</script>

<style scoped>
.student-layout {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.header {
  background: linear-gradient(135deg, #28a745 0%, #20c997 100%);
  color: white;
  padding: 0 20px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.1);
  z-index: 1000;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
}

.header h2 {
  font-size: 1.25rem;
  margin: 0;
  white-space: nowrap;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.user-info span {
  font-size: 0.9rem;
  opacity: 0.9;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .header-content {
    padding: 0 16px;
  }
  
  .header h2 {
    font-size: 1.1rem;
  }
  
  .user-info {
    gap: 12px;
  }
  
  .user-info span {
    display: none; /* 移动端隐藏用户名，节省空间 */
  }
}

.el-container {
  flex: 1;
  overflow: hidden;
}

.el-aside {
  background: linear-gradient(180deg, #f8f9fa 0%, #e9ecef 100%);
  border-right: 1px solid #dee2e6;
  transition: transform 0.3s ease;
  z-index: 900;
}

.el-main {
  padding: 24px;
  background: linear-gradient(180deg, #ffffff 0%, #f8f9fa 100%);
  overflow-y: auto;
  min-height: calc(100vh - 60px);
}

.sidebar-menu {
  border: none;
  background: transparent;
  height: 100%;
}

.sidebar-menu .el-menu-item {
  color: #495057;
  margin: 4px 8px;
  border-radius: 8px;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 12px;
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

/* 移动端侧边栏样式 */
@media (max-width: 768px) {
  .el-aside {
    position: fixed;
    top: 60px;
    left: 0;
    height: calc(100vh - 60px);
    width: 250px !important;
    transform: translateX(-100%);
    box-shadow: 2px 0 10px rgba(0,0,0,0.1);
  }
  
  .el-aside.show {
    transform: translateX(0);
  }
  
  .sidebar-menu {
    padding-top: 20px;
  }
  
  .sidebar-menu .el-menu-item {
    margin: 6px 12px;
    height: 44px;
    font-size: 0.95rem;
  }
  
  .sidebar-menu .el-menu-item .el-icon {
    font-size: 1.1rem;
  }
  
  .el-main {
    padding: 16px;
  }
  
  /* 添加遮罩层 */
  .sidebar-overlay {
    position: fixed;
    top: 60px;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.3);
    z-index: 800;
    display: none;
  }
  
  .sidebar-overlay.active {
    display: block;
  }
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

/* 移动端按钮样式 */
.mobile-menu-btn {
  display: none;
  background: transparent;
  border: none;
  color: white;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 8px;
  border-radius: 4px;
  transition: background 0.2s;
}

.mobile-menu-btn:hover {
  background: rgba(255, 255, 255, 0.1);
}

@media (max-width: 768px) {
  .mobile-menu-btn {
    display: block;
  }
  
  .header h2 {
    flex: 1;
    text-align: center;
  }
  
  .user-info {
    margin-left: auto;
  }
}
</style>
