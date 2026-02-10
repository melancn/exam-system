<template>
  <div class="login-logs">
    <div class="page-header">
      <h1>登录日志</h1>
      <p class="page-description">查看系统登录记录</p>
    </div>

    <div class="search-container">
      <div class="search-form">
        <div class="form-group">
          <label>用户名</label>
          <input
            v-model="search.username"
            type="text"
            placeholder="请输入用户名"
            @keyup.enter="handleSearch"
          />
        </div>
        <div class="form-group">
                <el-date-picker
                  v-model="search.timeRange"
                  type="datetimerange"
                  range-separator="至"
                  start-placeholder="开始时间"
                  end-placeholder="结束时间"
              @change="handleSearch"
                />
        </div>
        <div class="form-actions">
          <button @click="handleSearch" class="btn-primary">查询</button>
          <button @click="resetSearch" class="btn-secondary">重置</button>
        </div>
      </div>
    </div>

    <div class="table-container">
      <div class="table-header">
        <div class="table-title">登录记录列表</div>
        <div class="table-actions">
          <span class="total-count">共 {{ total }} 条记录</span>
        </div>
      </div>

      <div class="table-wrapper">
        <table class="data-table">
          <thead>
            <tr>
              <th>序号</th>
              <th>用户名</th>
              <th>登录时间</th>
              <th>IP地址</th>
              <th>用户代理</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(log, index) in logs" :key="log.id">
              <td>{{ (currentPage - 1) * pageSize + index + 1 }}</td>
              <td>{{ log.username }}</td>
              <td>{{ formatTime(log.created_at) }}</td>
              <td>{{ log.ip }}</td>
              <td class="user-agent-cell">
                <span class="user-agent-text">{{ log.user_agent }}</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="pagination-container" v-if="total > 0">
        <div class="pagination-info">
          第 {{ currentPage }} 页，共 {{ totalPages }} 页
        </div>
        <div class="pagination-controls">
          <button
            @click="changePage(1)"
            :disabled="currentPage === 1"
            class="page-btn"
          >
            首页
          </button>
          <button
            @click="changePage(currentPage - 1)"
            :disabled="currentPage === 1"
            class="page-btn"
          >
            上一页
          </button>
          <span class="page-info">{{ currentPage }} / {{ totalPages }}</span>
          <button
            @click="changePage(currentPage + 1)"
            :disabled="currentPage === totalPages"
            class="page-btn"
          >
            下一页
          </button>
          <button
            @click="changePage(totalPages)"
            :disabled="currentPage === totalPages"
            class="page-btn"
          >
            末页
          </button>
        </div>
        <div class="page-size-selector">
          <label>每页显示：</label>
          <select v-model="pageSize" @change="handlePageSizeChange">
            <option value="10">10条</option>
            <option value="20">20条</option>
            <option value="50">50条</option>
            <option value="100">100条</option>
          </select>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { teacherAPI } from '@/services/api'

export default {
  name: 'LoginLogs',
  setup() {
    const logs = ref([])
    const total = ref(0)
    const currentPage = ref(1)
    const pageSize = ref(20)
    const loading = ref(false)

    const search = ref({
      username: '',
      timeRange: []
    })

    // 初始化默认时间范围（近一周）
    const initDateRange = () => {
      const endDate = new Date()
      const startDate = new Date()
      startDate.setDate(endDate.getDate() - 7)
      
      search.value.timeRange = [startDate, endDate]
    }

    const fetchLogs = async () => {
      loading.value = true
      try {
        // 处理时间范围
        let startTime = ''
        let endTime = ''
        
        if (search.value.timeRange && search.value.timeRange.length === 2) {
          const [start, end] = search.value.timeRange
          if (start && end) {
            startTime = start.toISOString().split('T')[0]
            endTime = end.toISOString().split('T')[0]
          }
        }

        const params = {
          page: currentPage.value,
          page_size: pageSize.value,
          username: search.value.username,
          start_time: startTime,
          end_time: endTime
        }

        const response = await teacherAPI.getLoginLogs(params)
        logs.value = response.items || []
        total.value = response.total || 0
      } catch (error) {
        console.error('获取登录日志失败:', error)
        // 这里可以添加错误提示
      } finally {
        loading.value = false
      }
    }

    const handleSearch = () => {
      currentPage.value = 1
      fetchLogs()
    }

    const resetSearch = () => {
      initDateRange()
      search.value.username = ''
      currentPage.value = 1
      fetchLogs()
    }

    const changePage = (page) => {
      if (page >= 1 && page <= Math.ceil(total.value / pageSize.value)) {
        currentPage.value = page
        fetchLogs()
      }
    }

    const handlePageSizeChange = () => {
      currentPage.value = 1
      fetchLogs()
    }

    const formatTime = (timestamp) => {
      if (!timestamp) return ''
      const date = new Date(timestamp)
      return date.toLocaleString('zh-CN')
    }

    onMounted(() => {
      initDateRange()
      fetchLogs()
    })

    return {
      logs,
      total,
      currentPage,
      pageSize,
      loading,
      search,
      handleSearch,
      resetSearch,
      changePage,
      handlePageSizeChange,
      formatTime,
      totalPages: () => Math.ceil(total.value / pageSize.value)
    }
  }
}
</script>

<style scoped>
.login-logs {
  padding: 24px;
  background: #f5f5f5;
  min-height: 100vh;
}

.page-header {
  margin-bottom: 24px;
}

.page-header h1 {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
  color: #333;
}

.page-description {
  margin: 0;
  color: #666;
  font-size: 14px;
}

.search-container {
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  margin-bottom: 24px;
}

.search-form {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  align-items: end;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  font-size: 14px;
  color: #666;
  font-weight: 500;
}

.form-group input {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  transition: border-color 0.2s;
}

.form-group input:focus {
  outline: none;
  border-color: #007bff;
}

.form-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.btn-primary {
  background: #007bff;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: background-color 0.2s;
}

.btn-primary:hover {
  background: #0056b3;
}

.btn-secondary {
  background: #6c757d;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: background-color 0.2s;
}

.btn-secondary:hover {
  background: #545b62;
}

.table-container {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #eee;
  background: #f8f9fa;
}

.table-title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

.table-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.total-count {
  font-size: 14px;
  color: #666;
}

.table-wrapper {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.data-table th,
.data-table td {
  padding: 12px 16px;
  text-align: left;
  border-bottom: 1px solid #eee;
}

.data-table th {
  background: #f8f9fa;
  font-weight: 600;
  color: #333;
  position: sticky;
  top: 0;
  z-index: 1;
}

.data-table tbody tr:hover {
  background: #f8f9fa;
}

.user-agent-cell {
  max-width: 300px;
}

.user-agent-text {
  display: block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-family: monospace;
  font-size: 12px;
  color: #666;
}

.pagination-container {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-top: 1px solid #eee;
  background: #f8f9fa;
}

.pagination-info {
  font-size: 14px;
  color: #666;
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-btn {
  padding: 6px 12px;
  border: 1px solid #ddd;
  background: white;
  color: #333;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
}

.page-btn:hover:not(:disabled) {
  background: #f8f9fa;
  border-color: #ccc;
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  font-size: 14px;
  color: #666;
  min-width: 80px;
  text-align: center;
}

.page-size-selector {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #666;
}

.page-size-selector select {
  padding: 6px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  background: white;
  cursor: pointer;
}


/* 响应式设计 */
@media (max-width: 768px) {
  .search-form {
    grid-template-columns: 1fr;
  }
  
  .pagination-container {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }
  
  .pagination-controls {
    justify-content: center;
  }
}
</style>
