<template>
  <div class="exam-result-page">
    <div class="page-header">
      <h2>考试结果详情</h2>
      <div class="result-summary">
        <div class="score-card">
          <div class="score-number">{{ result.score }}</div>
          <div class="score-label">得分</div>
        </div>
        <div class="score-card">
          <div class="score-number">{{ result.totalScore }}</div>
          <div class="score-label">总分</div>
        </div>
        <div class="score-card">
          <div class="score-number">{{ result.percentage }}%</div>
          <div class="score-label">正确率</div>
        </div>
        <div class="score-card">
          <div class="score-number">{{ result.timeUsed }}分钟</div>
          <div class="score-label">用时</div>
        </div>
      </div>
    </div>
    
    <div class="page-content">
      <el-card class="card-container">
        <template #header>
          <h3>考试信息</h3>
        </template>
        
        <div class="exam-info">
          <div class="info-item">
            <span class="label">试卷名称：</span>
            <span class="value">{{ result.examTitle }}</span>
          </div>
          <div class="info-item">
            <span class="label">提交时间：</span>
            <span class="value">{{ result.submitTime }}</span>
          </div>
          <div class="info-item">
            <span class="label">考试状态：</span>
            <el-tag :type="result.passed ? 'success' : 'danger'">
              {{ result.passed ? '及格' : '不及格' }}
            </el-tag>
          </div>
        </div>
      </el-card>

      <el-card class="card-container" style="margin-top: 20px;">
        <template #header>
          <h3>答题详情</h3>
        </template>
        
        <div class="question-details">
          <div 
            v-for="(question, index) in questions" 
            :key="question.id"
            class="question-item"
            :class="{ 'correct': question.isCorrect, 'incorrect': !question.isCorrect }"
          >
            <div class="question-header">
              <span class="question-number">第{{ index + 1 }}题</span>
              <span class="question-type">{{ question.type === 'single' ? '单选题' : '填空题' }}</span>
              <span class="question-score">({{ question.score }}分)</span>
              <el-tag :type="question.isCorrect ? 'success' : 'danger'" size="small">
                {{ question.isCorrect ? '正确' : '错误' }}
              </el-tag>
            </div>
            
            <div class="question-content">
              <p class="question-text">{{ question.content }}</p>
              
              <!-- 单选题详情 -->
              <div v-if="question.type === 'single'" class="question-analysis">
                <div class="answer-section">
                  <div class="your-answer">
                    <span class="answer-label">你的答案：</span>
                    <span class="answer-value">{{ getSelectedOption(question) }}</span>
                  </div>
                  <div class="correct-answer">
                    <span class="answer-label">正确答案：</span>
                    <span class="answer-value">{{ question.correctAnswer }}</span>
                  </div>
                </div>
                
                <div class="options-list">
                  <div 
                    v-for="option in question.options" 
                    :key="option.key"
                    class="option-item"
                    :class="{ 'correct-option': option.key === question.correctAnswer }"
                  >
                    <span class="option-key">{{ option.key.toUpperCase() }}.</span>
                    <span class="option-text">{{ option.text }}</span>
                  </div>
                </div>
              </div>
              
              <!-- 填空题详情 -->
              <div v-else class="question-analysis">
                <div class="answer-section">
                  <div class="your-answer">
                    <span class="answer-label">你的答案：</span>
                    <span class="answer-value">{{ getFillAnswer(question) }}</span>
                  </div>
                  <div class="correct-answer">
                    <span class="answer-label">参考答案：</span>
                    <span class="answer-value">{{ question.referenceAnswer }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </el-card>
      
      <div class="actions">
        <el-button @click="$router.back()">返回</el-button>
        <el-button type="primary" @click="downloadReport">下载报告</el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { studentAPI } from '@/services/api'

const route = useRoute()
const router = useRouter()

const result = ref({
  id: route.params.resultId,
  examTitle: '',
  score: 0,
  totalScore: 100,
  submitTime: '',
  timeUsed: 0,
  passed: false,
  answers: {}
})

const questions = ref([])

// 计算正确率
const resultPercentage = computed(() => {
  return result.value.totalScore > 0 
    ? Math.round((result.value.score / result.value.totalScore) * 100)
    : 0
})

// 获取单选题选中的选项
const getSelectedOption = (question) => {
  const answerKey = question.id
  const selectedKey = result.value.answers[answerKey]
  if (!selectedKey) return '未作答'
  
  const option = question.options.find(opt => opt.key === selectedKey)
  return option ? `${option.key.toUpperCase()}. ${option.text}` : selectedKey
}

// 获取填空题答案
const getFillAnswer = (question) => {
  const answerKey = `${question.id}_1`
  const answer = result.value.answers[answerKey]
  return answer || '未作答'
}

// 判断题目是否正确
const isQuestionCorrect = (question) => {
  if (question.type === 'single') {
    const answerKey = question.id
    const selectedKey = result.value.answers[answerKey]
    return selectedKey === question.correctAnswer
  } else {
    const answerKey = `${question.id}_1`
    const userAnswer = result.value.answers[answerKey]
    return userAnswer && userAnswer.trim().toLowerCase() === question.referenceAnswer.toLowerCase()
  }
}

// 下载报告
const downloadReport = () => {
  const content = `
考试结果报告
============
试卷名称：${result.value.examTitle}
考试时间：${result.value.submitTime}
得分：${result.value.score}/${result.value.totalScore}
正确率：${resultPercentage.value}%
考试状态：${result.value.passed ? '及格' : '不及格'}

答题详情：
${questions.value.map((q, i) => `
第${i + 1}题（${q.type === 'single' ? '单选题' : '填空题'}，${q.score}分）
题目：${q.content}
${q.type === 'single' ? `你的答案：${getSelectedOption(q)}\n正确答案：${q.correctAnswer}` : `你的答案：${getFillAnswer(q)}\n参考答案：${q.referenceAnswer}`}
状态：${isQuestionCorrect(q) ? '正确' : '错误'}
`).join('')}

  `
  
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${result.value.examTitle}_成绩报告.txt`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
  
  ElMessage.success('报告下载成功')
}

onMounted(async () => {
  try {
    // 获取考试结果
    const response = await studentAPI.getExamResult(result.value.id)
    
    result.value = {
      id: response.id,
      examTitle: response.examTitle,
      score: response.score,
      totalScore: response.totalScore,
      submitTime: response.submitTime,
      timeUsed: response.timeUsed,
      passed: response.passed,
      answers: response.answers ? JSON.parse(response.answers) : {}
    }
    
    // 获取试卷详情
    const examResponse = await studentAPI.getExamDetails(response.examId)
    
    if (examResponse && examResponse.questions) {
      questions.value = examResponse.questions.map(question => ({
        id: question.id,
        type: question.type,
        content: question.content,
        score: question.score,
        options: question.type === 'single' ? JSON.parse(question.options || '[]') : [],
        correctAnswer: question.answer,
        referenceAnswer: question.referenceAnswer || question.answer,
        isCorrect: isQuestionCorrect({
          id: question.id,
          type: question.type,
          answer: question.answer,
          referenceAnswer: question.referenceAnswer || question.answer
        })
      }))
    }
    
  } catch (error) {
    console.error('获取考试结果失败:', error)
    ElMessage.error('获取考试结果失败')
    router.back()
  }
})
</script>

<style scoped>
.exam-result-page {
  min-height: 100%;
}

.page-header {
  margin-bottom: 20px;
}

.result-summary {
  display: flex;
  gap: 20px;
  margin-top: 20px;
}

.score-card {
  text-align: center;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
  flex: 1;
}

.score-number {
  font-size: 32px;
  font-weight: bold;
  color: #409eff;
  margin-bottom: 8px;
}

.score-label {
  color: #666;
  font-size: 14px;
}

.exam-info {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 15px;
}

.info-item {
  display: flex;
  align-items: center;
  padding: 10px;
  background: #f8f9fa;
  border-radius: 4px;
}

.label {
  font-weight: bold;
  color: #666;
  min-width: 100px;
}

.value {
  color: #333;
}

.question-details {
  margin-top: 20px;
}

.question-item {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  margin-bottom: 20px;
  padding: 20px;
  transition: all 0.3s;
}

.question-item.correct {
  border-color: #67c23a;
  background-color: #f0f9ff;
}

.question-item.incorrect {
  border-color: #f56c6c;
  background-color: #fff0f0;
}

.question-header {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 15px;
  flex-wrap: wrap;
}

.question-number {
  font-weight: bold;
  color: #409eff;
  font-size: 16px;
}

.question-type {
  color: #909399;
  font-size: 14px;
}

.question-score {
  color: #409eff;
  font-size: 14px;
}

.question-content {
  margin-top: 10px;
}

.question-text {
  font-size: 16px;
  line-height: 1.6;
  margin-bottom: 20px;
  color: #333;
}

.question-analysis {
  background: white;
  padding: 15px;
  border-radius: 6px;
  border: 1px solid #e4e7ed;
}

.answer-section {
  margin-bottom: 15px;
}

.answer-label {
  font-weight: bold;
  color: #666;
  margin-right: 8px;
}

.answer-value {
  color: #333;
  font-family: monospace;
}

.options-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 10px;
}

.option-item {
  padding: 10px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.option-item.correct-option {
  border-color: #67c23a;
  background-color: #f0f9ff;
}

.option-key {
  font-weight: bold;
  color: #409eff;
  min-width: 20px;
}

.option-text {
  color: #333;
}

.actions {
  text-align: center;
  margin-top: 30px;
  padding: 20px;
}
</style>
