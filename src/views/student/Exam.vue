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
import { examProgressSync } from '@/utils/sync'
import { 
  sendWebSocketMessage, 
  registerMessageHandler, 
  removeMessageHandler, 
  getWebSocketStatus 
} from '@/composables/useWebSocket'

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
const isWebSocketConnected = getWebSocketStatus()
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
      if (remainingTime.value % 10 === 0) {
        sendTimeUpdate()
      }
    } else {
      submitExam(true) // 自动提交
    }
  }, 1000)
}

// 注册WebSocket消息处理器
const examMessageHandler = (data) => {
  console.log('考试页面收到WebSocket消息:', data)
  
  switch (data.type) {
    case 'auth_success':
      console.log('考试页面WebSocket认证成功')
      // 发送开始计时消息
      sendTimerStart()
      break
    case 'pause':
      ElMessage.warning('考试已被教师暂停')
      break
    case 'resume':
      ElMessage.success('考试已被教师恢复')
      break
  }
}

// 发送开始计时消息
const sendTimerStart = () => {
  sendWebSocketMessage({
    type: 'start',
    examId: examInfo.value.id
  })
}

// 发送时间更新消息
const sendTimeUpdate = () => {
  sendWebSocketMessage({
    type: 'update',
    examId: examInfo.value.id,
    timeUsed: Math.floor((totalTime.value - remainingTime.value) / 60)
  })
}

// 发送结束计时消息
const sendTimerEnd = () => {
  sendWebSocketMessage({
    type: 'end',
    examId: examInfo.value.id,
    timeUsed: Math.floor((totalTime.value - remainingTime.value) / 60)
  })
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
    // 注册WebSocket消息处理器
    registerMessageHandler('exam', examMessageHandler)
    
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
  // 移除WebSocket消息处理器
  removeMessageHandler('exam')
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
