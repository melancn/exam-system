<template>
  <div class="exam-page">
    <div class="exam-header">
      <h2>{{ examInfo.title }}</h2>
      <div class="exam-timer">
        <span>剩余时间：</span>
        <el-statistic :value="remainingTime" format="HH:mm:ss" />
      </div>
    </div>
    
    <div class="exam-content">
      <el-card class="question-card" v-for="(question, index) in questions" :key="question.id">
        <template #header>
          <h3>第{{ index + 1 }}题 ({{ question.type === 'single' ? '单选题' : '填空题' }}) - {{ question.score }}分</h3>
        </template>
        
        <div class="question-content">
          <p class="question-text">{{ question.content }}</p>
          
          <!-- 单选题 -->
          <div v-if="question.type === 'single'" class="options">
            <el-radio-group v-model="answers[index].answer">
              <div v-for="option in question.options" :key="option.key" class="option-item">
                <el-radio :label="option.key">{{ option.key.toUpperCase() }}. {{ option.text }}</el-radio>
              </div>
            </el-radio-group>
          </div>
          
          <!-- 填空题 -->
          <div v-else class="fill-blanks">
            <div v-for="(item, key) in question.answerConfigs" :key="item.questionId" class="fill-blank-item">
              <span class="fill-blank-number">({{ key+1 }})</span>
              <el-input
                v-model="answers[index].answers[key]"
                :placeholder="question.placeholder || '请输入答案'"
                :type="item.type || 'text'"
                size="medium"
                :rows="question.inputType === 'textarea' ? 3 : undefined"
              />
            </div>
          </div>
        </div>
      </el-card>
      
      <div class="exam-actions">
        <el-button type="primary" size="large" @click="handleSubmitExam" :loading="submitting">
          提交试卷
        </el-button>
        <el-button size="large" @click="saveProgress">
          保存进度
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { studentAPI } from '@/services/api'
import config from '@/config'
import { examProgressSync } from '@/utils/sync'

// WebSocket连接
let ws = null
let timerInterval = null
let reconnectAttempts = 0
let reconnectTimer = null

const route = useRoute()
const router = useRouter()

const examInfo = ref({
  id: parseInt(route.params.id),
  title: '',
  duration: 60,
  totalScore: 100
})

const questions = ref([])
const answers = ref([])
const remainingTime = ref(0)
const timer = ref(null)
const submitting = ref(false)
const isWebSocketConnected = ref(false)
const isSyncing = ref(false)

const totalTime = computed(() => examInfo.value.duration * 60) // 转换为秒

// 处理从教师端获取的试卷数据
const processExamData = (examData) => {
  if (!examData || !examData.questions) return []
  
  return examData.questions.map(question => {
    const processedQuestion = {
      id: question.id,
      type: question.type,
      content: question.content,
      score: question.score || 10
    }
    
    let answerInpt = []
    if (question.type === 'single') {
      // 处理单选题选项
      processedQuestion.options = question.options || []
    } else if (question.type === 'fill') {
      // 处理填空题配置
      processedQuestion.inputCount = question.inputCount || 1
      processedQuestion.inputWidth = question.inputWidth || 'medium'
      processedQuestion.inputType = question.inputType || 'text'
      processedQuestion.placeholder = question.placeholder || '请输入答案'
      
      // 处理填空题的详细配置
      processedQuestion.answerConfigs = question.answerConfigs || []
      
      processedQuestion.answerConfigs.forEach(v=>{
        answerInpt.push('')
      })
    }
    answers.value.push({
      questionId: question.id,
      type: question.type,
      answer: '',
      answers: answerInpt
    })
    
    return processedQuestion
  })
}

const startTimer = () => {
  remainingTime.value = totalTime.value
  timer.value = setInterval(() => {
    if (remainingTime.value > 0) {
      remainingTime.value--
      // 每10秒更新一次WebSocket时间
      if (remainingTime.value % 10 === 0 && ws) {
        sendTimeUpdate()
      }
    } else {
      submitExam(true) // 自动提交
    }
  }, 1000)
}

// 初始化WebSocket连接
const initWebSocket = () => {
  const token = localStorage.getItem('token')
  if (!token) return

  const wsUrl = `${config.websocket.baseURL}${config.websocket.endpoint}`
  console.log('连接WebSocket:', wsUrl)
  
  ws = new WebSocket(wsUrl)

  ws.onopen = () => {
    console.log('WebSocket连接已建立')
    isWebSocketConnected.value = true
    reconnectAttempts = 0 // 重置重连次数
    
    // 发送认证消息
    const authMessage = {
      type: 'auth',
      token: token
    }
    ws.send(JSON.stringify(authMessage))
  }

  ws.onmessage = (event) => {
    const data = JSON.parse(event.data)
    console.log('收到服务器消息:', data)
    
    // 处理认证成功消息
    if (data.type === 'auth_success') {
      console.log('WebSocket认证成功')
      // 发送开始计时消息
      sendTimerStart()
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

// 发送开始计时消息
const sendTimerStart = () => {
  if (ws && ws.readyState === WebSocket.OPEN) {
    const message = {
      type: 'start',
      examId: examInfo.value.id
    }
    ws.send(JSON.stringify(message))
  }
}

// 发送时间更新消息
const sendTimeUpdate = () => {
  if (ws && ws.readyState === WebSocket.OPEN) {
    const message = {
      type: 'update',
      examId: examInfo.value.id,
      timeUsed: Math.floor((totalTime.value - remainingTime.value) / 60)
    }
    ws.send(JSON.stringify(message))
  }
}

// 发送结束计时消息
const sendTimerEnd = () => {
  if (ws && ws.readyState === WebSocket.OPEN) {
    const message = {
      type: 'end',
      examId: examInfo.value.id,
      timeUsed: Math.floor((totalTime.value - remainingTime.value) / 60)
    }
    ws.send(JSON.stringify(message))
  }
}

const saveProgress = () => {
  // 使用新的同步机制保存进度
  examProgressSync.saveProgress(examInfo.value.id, answers.value)
  ElMessage.success('进度已保存')
}

const submitExam = async (autoSubmit = false) => {
  if (!autoSubmit) {
    try {
      await ElMessageBox.confirm(
        '确定要提交试卷吗？提交后无法修改',
        '确认提交',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      )
    } catch {
      return
    }
  }

  submitting.value = true
  
  try {
    // 提交到API，后端会计算分数并返回结果
    const result = await studentAPI.submitExam(examInfo.value.id, {
      answers: answers.value,
      timeUsed: Math.floor((totalTime.value - remainingTime.value) / 60)
    })
    
    clearInterval(timer.value)
    localStorage.removeItem(`exam_${examInfo.value.id}_answers`)
    
    // 显示后端返回的分数
    ElMessage.success(`提交成功！您的得分：${result.score}/${result.totalScore}`)
    router.push('/student/scores')
  } catch (error) {
    console.error('提交失败:', error)
    ElMessage.error('提交失败，请稍后重试')
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  try {
    // 尝试从API获取试卷数据
    const response = await studentAPI.getExamDetails(examInfo.value.id)
    
    if (response && response.title) {
      // 处理从教师端获取的试卷数据
      examInfo.value = {
        id: response.id,
        title: response.title,
        duration: response.duration || 60,
        totalScore: response.totalScore || 100
      }
      
      // 处理题目数据
      questions.value = processExamData(response)
      
      // 加载保存的答案
      const savedAnswers = localStorage.getItem(`exam_${examInfo.value.id}_answers`)
      if (savedAnswers) {
        answers.value = JSON.parse(savedAnswers)
      }
      
      startTimer()
      // 初始化WebSocket连接
      initWebSocket()
    } else {
      ElMessage.error('试卷不存在')
      router.push('/student/exam-list')
    }
  } catch (error) {
    console.error('获取试卷数据失败:', error)
    ElMessage.error('获取试卷数据失败')
    router.push('/student/exam-list')
  }
})

// 在提交考试时发送结束消息
const handleSubmitExam = async (autoSubmit = false) => {
  // 发送结束计时消息
  sendTimerEnd()
  
  // 调用原始提交逻辑
  return await submitExam(autoSubmit)
}

onUnmounted(() => {
  if (timer.value) {
    clearInterval(timer.value)
  }
})
</script>

<style scoped>
.exam-page {
  padding: 20px;
}

.exam-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.exam-timer {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
}

.question-card {
  margin-bottom: 20px;
}

.question-text {
  font-size: 16px;
  line-height: 1.6;
  margin-bottom: 20px;
}

.options {
  margin: 15px 0;
}

.option-item {
  margin: 0 10px;
}

.fill-blanks {
  margin: 15px 0;
}

.fill-blank-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 15px;
}

.fill-blank-number {
  font-size: 14px;
  color: #409eff;
  font-weight: bold;
  min-width: 20px;
  text-align: center;
  margin-top: 5px;
}

.exam-actions {
  text-align: center;
  margin-top: 30px;
  padding: 20px;
}
</style>
