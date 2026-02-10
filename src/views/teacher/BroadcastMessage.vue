<template>
  <div class="broadcast-message-page">
    <div v-if="loading" class="loading-overlay">
      <el-loading :visible="true" />
    </div>
    
    <div class="page-header">
      <h2>广播消息</h2>
      <div class="header-actions">
        <el-button type="primary" @click="refreshExamList">刷新列表</el-button>
      </div>
    </div>
    
    <div class="page-content">
      <el-card class="card-container">
        <template #header>
          <div class="card-header">
            <h3>发送广播消息</h3>
            <div class="real-time-indicator">
              <el-tag :type="websocketStore.isRealTimeConnected ? 'success' : 'danger'" size="small">
                {{ websocketStore.isRealTimeConnected ? 'WebSocket已连接' : 'WebSocket已断开' }}
              </el-tag>
            </div>
          </div>
        </template>
        
        <el-form :model="messageForm" :rules="messageRules" ref="messageFormRef" label-width="120px">
          <el-form-item label="消息类型" prop="messageType">
            <el-select v-model="messageForm.messageType" placeholder="选择消息类型" style="width: 100%;">
              <el-option label="考试提醒" value="exam_reminder" />
              <el-option label="系统通知" value="system_notice" />
              <el-option label="紧急通知" value="urgent_notice" />
              <el-option label="一般消息" value="general" />
            </el-select>
          </el-form-item>
          
          <el-form-item label="目标考试" prop="targetExam">
            <el-select v-model="messageForm.targetExam" placeholder="选择目标考试" style="width: 100%;">
              <el-option label="所有考试" value="" />
              <el-option 
                v-for="exam in examList" 
                :key="exam.id" 
                :label="exam.title" 
                :value="exam.id"
              />
            </el-select>
          </el-form-item>
          
          <el-form-item label="目标班级" prop="targetClass">
            <el-select v-model="messageForm.targetClass" placeholder="选择目标班级" style="width: 100%;">
              <el-option label="所有班级" value="" />
              <el-option 
                v-for="classItem in classList" 
                :key="classItem.id" 
                :label="classItem.name" 
                :value="classItem.id"
              />
            </el-select>
          </el-form-item>
          
          <el-form-item label="消息标题" prop="title">
            <el-input 
              v-model="messageForm.title" 
              placeholder="请输入消息标题"
              maxlength="100"
              show-word-limit
            />
          </el-form-item>
          
          <el-form-item label="消息内容" prop="content">
            <el-input 
              v-model="messageForm.content" 
              type="textarea" 
              :rows="4"
              placeholder="请输入消息内容"
              maxlength="500"
              show-word-limit
            />
          </el-form-item>
          
          <el-form-item label="发送方式" prop="sendMethod">
            <el-radio-group v-model="messageForm.sendMethod">
              <el-radio value="immediate">立即发送</el-radio>
              <el-radio value="scheduled">定时发送</el-radio>
            </el-radio-group>
          </el-form-item>
          
          <el-form-item v-if="messageForm.sendMethod === 'scheduled'" label="发送时间" prop="sendTime">
            <el-date-picker
              v-model="messageForm.sendTime"
              type="datetime"
              placeholder="选择发送时间"
              format="YYYY-MM-DD HH:mm:ss"
              value-format="YYYY-MM-DD HH:mm:ss"
              style="width: 100%;"
            />
          </el-form-item>
          
          <el-form-item>
            <el-button type="primary" @click="sendMessage" :loading="sending">
              {{ messageForm.sendMethod === 'immediate' ? '立即发送' : '安排发送' }}
            </el-button>
            <el-button @click="resetForm">重置</el-button>
            <el-button type="info" @click="previewMessage">预览消息</el-button>
          </el-form-item>
        </el-form>
      </el-card>
      
      <!-- 消息发送历史 -->
      <el-card class="card-container" style="margin-top: 20px;">
        <template #header>
          <div class="card-header">
            <h3>消息发送历史</h3>
            <div class="filter-container">
              <el-select v-model="historyFilter.type" placeholder="消息类型" style="width: 150px; margin-right: 10px;">
                <el-option label="全部" value="" />
                <el-option label="考试提醒" value="exam_reminder" />
                <el-option label="系统通知" value="system_notice" />
                <el-option label="紧急通知" value="urgent_notice" />
                <el-option label="一般消息" value="general" />
              </el-select>
              <el-input 
                v-model="historyFilter.keyword" 
                placeholder="搜索标题或内容" 
                style="width: 200px;"
                prefix-icon="Search"
                @input="handleHistorySearch"
              />
            </div>
          </div>
        </template>
        
        <el-table :data="filteredHistory" style="width: 100%">
          <el-table-column prop="messageType" label="消息类型" width="120">
            <template #default="scope">
              <el-tag :type="getMessageTypeTag(scope.row.messageType)">
                {{ getMessageTypeText(scope.row.messageType) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="title" label="标题" width="200" />
          <el-table-column prop="content" label="内容" width="300" show-overflow-tooltip />
          <el-table-column prop="targetExam" label="目标考试" width="150" />
          <el-table-column prop="targetClass" label="目标班级" width="150" />
          <el-table-column prop="sendTime" label="发送时间" width="180">
            <template #default="scope">
              {{ scope.row.sendTime || '未发送' }}
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.status === 'sent' ? 'success' : scope.row.status === 'pending' ? 'warning' : 'danger'">
                {{ scope.row.status === 'sent' ? '已发送' : scope.row.status === 'pending' ? '待发送' : '发送失败' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150">
            <template #default="scope">
              <el-button type="text" @click="viewMessageDetail(scope.row)">查看详情</el-button>
              <el-button 
                v-if="scope.row.status === 'pending'" 
                type="text" 
                @click="cancelMessage(scope.row)"
              >
                取消
              </el-button>
            </template>
          </el-table-column>
        </el-table>
        
        <div class="pagination">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="total"
            layout="total, sizes, prev, pager, next"
          />
        </div>
      </el-card>
      
      <!-- 实时连接学生列表 -->
      <el-card class="card-container" style="margin-top: 20px;">
        <template #header>
          <div class="card-header">
            <h3>实时连接学生列表</h3>
            <div class="real-time-controls">
              <el-button type="primary" size="small" @click="refreshRealTimeData">刷新</el-button>
              <el-tag :type="websocketStore.isRealTimeConnected ? 'success' : 'danger'" size="small">
                {{ websocketStore.isRealTimeConnected ? '已连接' : '已断开' }}
              </el-tag>
            </div>
          </div>
        </template>
        
        <el-table :data="websocketStore.realTimeExamData" style="width: 100%">
          <el-table-column prop="examTitle" label="考试场次" width="200" />
          <el-table-column prop="paperTitle" label="试卷标题" width="200" />
          <el-table-column prop="onlineCount" label="当前在线人数" width="150">
            <template #default="scope">
              <el-tag type="primary">{{ scope.row.onlineCount }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="详情" width="150">
            <template #default="scope">
              <el-button type="text" @click="viewExamStudents(scope.row)">查看明细</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
    
    <!-- 消息预览对话框 -->
    <el-dialog v-model="showPreviewDialog" title="消息预览" width="600px">
      <div v-if="messageForm.title || messageForm.content" class="message-preview">
        <div class="preview-header">
          <h4>{{ messageForm.title || '无标题' }}</h4>
          <el-tag :type="getMessageTypeTag(messageForm.messageType)">
            {{ getMessageTypeText(messageForm.messageType) }}
          </el-tag>
        </div>
        <div class="preview-content">
          <p>{{ messageForm.content || '无内容' }}</p>
        </div>
        <div class="preview-meta">
          <p><strong>目标考试:</strong> {{ getExamName(messageForm.targetExam) }}</p>
          <p><strong>目标班级:</strong> {{ getClassName(messageForm.targetClass) }}</p>
          <p><strong>发送方式:</strong> {{ messageForm.sendMethod === 'immediate' ? '立即发送' : '定时发送' }}</p>
          <p v-if="messageForm.sendMethod === 'scheduled'"><strong>发送时间:</strong> {{ messageForm.sendTime }}</p>
        </div>
      </div>
      <div v-else class="empty-preview">
        <el-empty description="请填写消息内容后预览" />
      </div>
    </el-dialog>
    
    <!-- 考试学生明细对话框 -->
    <el-dialog v-model="showStudentsDialog" :title="`考试场次：${selectedExam?.examTitle} - 学生明细`" width="900px">
      <div v-if="selectedExam" class="exam-students-detail">
        <div class="exam-info" style="margin-bottom: 20px; padding: 15px; background: #f8f9fa; border-radius: 8px;">
          <el-descriptions :column="3" border>
            <el-descriptions-item label="考试场次">{{ selectedExam.examTitle }}</el-descriptions-item>
            <el-descriptions-item label="试卷标题">{{ selectedExam.paperTitle }}</el-descriptions-item>
            <el-descriptions-item label="当前在线人数">
              <el-tag type="primary">{{ selectedExam.onlineCount }}</el-tag>
            </el-descriptions-item>
          </el-descriptions>
        </div>
        
        <el-table :data="selectedExam.students" style="width: 100%">
          <el-table-column prop="studentId" label="学号" width="120" />
          <el-table-column prop="studentName" label="学生姓名" width="150" />
          <el-table-column prop="className" label="班级" width="150" />
          <el-table-column prop="timeUsed" label="已用时间" width="120">
            <template #default="scope">
              {{ scope.row.timeUsed }}分钟
            </template>
          </el-table-column>
          <el-table-column prop="startTime" label="开始时间" width="180">
            <template #default="scope">
              {{ new Date(scope.row.startTime * 1000).toLocaleString() }}
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.isActive ? 'primary' : 'success'">
                {{ scope.row.isActive ? '考试中' : '已结束' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120">
            <template #default="scope">
              <el-button type="text" @click="sendDirectMessage(scope.row)">发送消息</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-dialog>
    
    <!-- 消息详情对话框 -->
    <el-dialog v-model="showDetailDialog" title="消息详情" width="800px">
      <div v-if="selectedMessage" class="message-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="消息类型">
            <el-tag :type="getMessageTypeTag(selectedMessage.messageType)">
              {{ getMessageTypeText(selectedMessage.messageType) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="selectedMessage.status === 'sent' ? 'success' : selectedMessage.status === 'pending' ? 'warning' : 'danger'">
              {{ selectedMessage.status === 'sent' ? '已发送' : selectedMessage.status === 'pending' ? '待发送' : '发送失败' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="标题" span="2">{{ selectedMessage.title }}</el-descriptions-item>
          <el-descriptions-item label="内容" span="2">{{ selectedMessage.content }}</el-descriptions-item>
          <el-descriptions-item label="目标考试">{{ selectedMessage.targetExam }}</el-descriptions-item>
          <el-descriptions-item label="目标班级">{{ selectedMessage.targetClass }}</el-descriptions-item>
          <el-descriptions-item label="发送时间">{{ selectedMessage.sendTime || '未发送' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ selectedMessage.createTime }}</el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useWebSocketStore } from '@/stores/websocket'
import { teacherAPI } from '../../services/api'

const websocketStore = useWebSocketStore()
const messageFormRef = ref()
const loading = ref(false)
const sending = ref(false)
const showPreviewDialog = ref(false)
const showStudentsDialog = ref(false)
const showDetailDialog = ref(false)
const selectedExam = ref(null)
const selectedMessage = ref(null)

// 表单数据
const messageForm = reactive({
  messageType: '',
  targetExam: '',
  targetClass: '',
  title: '',
  content: '',
  sendMethod: 'immediate',
  sendTime: ''
})

// 表单验证规则
const messageRules = {
  messageType: [
    { required: true, message: '请选择消息类型', trigger: 'change' }
  ],
  title: [
    { required: true, message: '请输入消息标题', trigger: 'blur' },
    { min: 1, max: 100, message: '标题长度在 1 到 100 个字符', trigger: 'blur' }
  ],
  content: [
    { required: true, message: '请输入消息内容', trigger: 'blur' },
    { min: 1, max: 500, message: '内容长度在 1 到 500 个字符', trigger: 'blur' }
  ],
  sendMethod: [
    { required: true, message: '请选择发送方式', trigger: 'change' }
  ],
  sendTime: [
    { 
      validator: (rule, value, callback) => {
        if (messageForm.sendMethod === 'scheduled' && !value) {
          callback(new Error('请选择发送时间'))
        } else {
          callback()
        }
      }, 
      trigger: 'change' 
    }
  ]
}

// 历史记录相关
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const messageHistory = ref([])
const historyFilter = reactive({
  type: '',
  keyword: ''
})

// 考试和班级列表
const examList = ref([])
const classList = ref([])

// 计算属性
const filteredHistory = computed(() => {
  let result = messageHistory.value
  
  // 消息类型筛选
  if (historyFilter.type) {
    result = result.filter(item => item.messageType === historyFilter.type)
  }
  
  // 关键词搜索
  if (historyFilter.keyword) {
    const keyword = historyFilter.keyword.toLowerCase()
    result = result.filter(item => 
      item.title.toLowerCase().includes(keyword) ||
      item.content.toLowerCase().includes(keyword)
    )
  }
  
  return result
})

// 获取消息类型标签样式
const getMessageTypeTag = (type) => {
  const typeMap = {
    'exam_reminder': 'primary',
    'system_notice': 'info',
    'urgent_notice': 'danger',
    'general': 'success'
  }
  return typeMap[type] || 'info'
}

// 获取消息类型文本
const getMessageTypeText = (type) => {
  const typeMap = {
    'exam_reminder': '考试提醒',
    'system_notice': '系统通知',
    'urgent_notice': '紧急通知',
    'general': '一般消息'
  }
  return typeMap[type] || type
}

// 获取考试名称
const getExamName = (examId) => {
  if (!examId) return '所有考试'
  const exam = examList.value.find(item => item.id === examId)
  return exam ? exam.title : '未知考试'
}

// 获取班级名称
const getClassName = (classId) => {
  if (!classId) return '所有班级'
  const classItem = classList.value.find(item => item.id === classId)
  return classItem ? classItem.name : '未知班级'
}

// 发送消息
const sendMessage = async () => {
  if (!messageFormRef.value) return
  
  try {
    await messageFormRef.value.validate()
    
    if (!websocketStore.isRealTimeConnected) {
      ElMessage.error('WebSocket连接未建立，无法发送消息')
      return
    }
    
    sending.value = true
    
    const messageData = {
      messageType: messageForm.messageType,
      targetExam: messageForm.targetExam,
      targetClass: messageForm.targetClass,
      title: messageForm.title,
      content: messageForm.content,
      sendMethod: messageForm.sendMethod,
      sendTime: messageForm.sendTime
    }
    
    // 使用WebSocket发送广播消息
    const success = websocketStore.broadcastMessage(messageData)
    
    if (success) {
      // 添加到历史记录
      const historyItem = {
        id: Date.now(),
        ...messageData,
        status: messageForm.sendMethod === 'immediate' ? 'sent' : 'pending',
        createTime: new Date().toISOString(),
        sendTime: messageForm.sendMethod === 'immediate' ? new Date().toISOString() : messageForm.sendTime
      }
      messageHistory.value.unshift(historyItem)
      total.value++
      
      // 重置表单
      resetForm()
      ElMessage.success('消息发送成功')
    }
  } catch (error) {
    console.error('发送消息失败:', error)
    if (error.message) {
      ElMessage.error(error.message)
    }
  } finally {
    sending.value = false
  }
}

// 重置表单
const resetForm = () => {
  if (messageFormRef.value) {
    messageFormRef.value.resetFields()
  }
}

// 预览消息
const previewMessage = () => {
  if (!messageForm.title && !messageForm.content) {
    ElMessage.warning('请填写消息标题或内容')
    return
  }
  showPreviewDialog.value = true
}

// 查看考试学生明细
const viewExamStudents = (exam) => {
  selectedExam.value = exam
  showStudentsDialog.value = true
}

// 发送直接消息
const sendDirectMessage = (student) => {
  ElMessageBox.prompt('请输入要发送给该学生的消息', '发送消息', {
    confirmButtonText: '发送',
    cancelButtonText: '取消',
    inputType: 'textarea',
    inputValidator: (value) => {
      if (!value) {
        return '消息内容不能为空'
      }
      if (value.length > 200) {
        return '消息内容不能超过200个字符'
      }
      return true
    }
  }).then(({ value }) => {
    const messageData = {
      messageType: 'direct',
      targetStudent: student.studentId,
      title: '教师消息',
      content: value,
      sendMethod: 'immediate'
    }
    
    const success = websocketStore.broadcastMessage(messageData)
    if (success) {
      ElMessage.success('消息发送成功')
    }
  }).catch(() => {
    // 用户取消操作
  })
}

// 查看消息详情
const viewMessageDetail = (message) => {
  selectedMessage.value = message
  showDetailDialog.value = true
}

// 取消消息
const cancelMessage = (message) => {
  ElMessageBox.confirm('确定要取消这条待发送的消息吗？', '确认取消', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    const index = messageHistory.value.findIndex(item => item.id === message.id)
    if (index >= 0) {
      messageHistory.value.splice(index, 1)
      total.value--
      ElMessage.success('消息已取消')
    }
  }).catch(() => {
    // 用户取消操作
  })
}

// 处理历史记录搜索
const handleHistorySearch = () => {
  currentPage.value = 1
}

// 刷新考试列表
const refreshExamList = async () => {
  await fetchExamList()
  ElMessage.success('考试列表已刷新')
}

// 刷新实时数据
const refreshRealTimeData = () => {
  websocketStore.refreshRealTimeData()
}

// 获取考试列表
const fetchExamList = async () => {
  try {
    const response = await teacherAPI.getExams()
    examList.value = response.data || []
  } catch (error) {
    console.error('获取考试列表失败:', error)
    ElMessage.error('获取考试列表失败')
  }
}

// 获取班级列表
const fetchClassList = async () => {
  try {
    const response = await teacherAPI.getClasses()
    classList.value = response.classes || []
  } catch (error) {
    console.error('获取班级列表失败:', error)
    ElMessage.error('获取班级列表失败')
  }
}

// 获取消息历史记录
const fetchMessageHistory = async () => {
  try {
    // 这里应该调用API获取消息历史记录
    // 目前使用模拟数据
    messageHistory.value = [
      {
        id: 1,
        messageType: 'exam_reminder',
        title: '考试即将开始',
        content: '请各位同学做好准备，考试将在10分钟后开始',
        targetExam: 'JavaScript基础考试',
        targetClass: '计算机1班',
        status: 'sent',
        sendTime: '2023-12-01 09:50:00',
        createTime: '2023-12-01 09:50:00'
      },
      {
        id: 2,
        messageType: 'system_notice',
        title: '系统维护通知',
        content: '系统将于今晚22:00进行维护，预计持续1小时',
        targetExam: '所有考试',
        targetClass: '所有班级',
        status: 'pending',
        sendTime: '2023-12-01 22:00:00',
        createTime: '2023-12-01 14:30:00'
      }
    ]
    total.value = messageHistory.value.length
  } catch (error) {
    console.error('获取消息历史记录失败:', error)
  }
}

// 初始化
onMounted(() => {
  fetchExamList()
  fetchClassList()
  fetchMessageHistory()
})
</script>

<style scoped>
.broadcast-message-page {
  min-height: 100%;
  padding: 20px;
  background-color: #f5f7fa;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  color: #303133;
  font-size: 24px;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.card-container {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.real-time-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
}

.real-time-indicator .el-tag {
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(64, 158, 255, 0.7);
  }
  70% {
    box-shadow: 0 0 0 10px rgba(64, 158, 255, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(64, 158, 255, 0);
  }
}

.filter-container {
  display: flex;
  gap: 10px;
  align-items: center;
}

.pagination {
  margin-top: 20px;
  text-align: right;
}

.message-preview {
  padding: 20px;
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.preview-header h4 {
  margin: 0;
  color: #303133;
}

.preview-content {
  margin-bottom: 15px;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 8px;
  border-left: 4px solid #409eff;
}

.preview-meta {
  color: #666;
  font-size: 14px;
}

.preview-meta p {
  margin: 5px 0;
}

.empty-preview {
  padding: 40px;
}

.exam-students-detail {
  max-height: 60vh;
  overflow-y: auto;
}

.exam-info {
  margin-bottom: 20px;
}

.real-time-controls {
  display: flex;
  gap: 10px;
  align-items: center;
}

.message-detail {
  max-height: 60vh;
  overflow-y: auto;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .card-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
  
  .filter-container {
    flex-direction: column;
    align-items: stretch;
  }
  
  .header-actions {
    width: 100%;
    justify-content: center;
  }
}
</style>