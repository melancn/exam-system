<template>
  <div class="exam-edit-page">
    <div class="page-header">
      <h2>{{ isEditMode ? '编辑试卷' : '创建新试卷' }}</h2>
    </div>
    
    <div class="page-content">
      <el-card class="card-container">
        <template #header>
          <h3>试卷信息</h3>
        </template>
        
        <el-form :model="examInfo" :rules="examRules" ref="examInfoFormRef" label-width="100px">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="试卷标题" prop="title">
                <el-input v-model="examInfo.title" placeholder="请输入试卷标题" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="总分" prop="totalScore">
                <el-input-number 
                  v-model="examInfo.totalScore" 
                  :min="10" 
                  :max="200" 
                  placeholder="请输入总分"
                  style="width: 100%"
                />
              </el-form-item>
            </el-col>
          </el-row>
          
          <el-form-item label="试卷描述" prop="description">
            <el-input 
              v-model="examInfo.description" 
              type="textarea" 
              :rows="3" 
              placeholder="请输入试卷描述"
            />
          </el-form-item>
        </el-form>
      </el-card>

      <el-card class="card-container" style="margin-top: 20px;">
        <template #header>
          <div class="questions-header">
            <h3>题目管理</h3>
            <div>
              <el-button type="primary" @click="addQuestion('single')">
                添加单选题
              </el-button>
              <el-button type="success" @click="addQuestion('fill')">
                添加填空题
              </el-button>
            </div>
          </div>
        </template>
        
        <div class="questions-summary">
          <p>题目总数：{{ examInfo.questions.length }}题</p>
          <p>当前总分：{{ currentTotalScore }}分</p>
          <p>目标总分：{{ examInfo.totalScore }}分</p>
          <p :class="{'score-match': currentTotalScore === examInfo.totalScore, 'score-mismatch': currentTotalScore !== examInfo.totalScore}">
            总分匹配：{{ currentTotalScore === examInfo.totalScore ? '✓' : '✗' }}
          </p>
        </div>
        
        <div class="questions-list">
          <div v-for="(question, index) in examInfo.questions" :key="question.id" class="question-item">
            <div class="question-header">
              <span class="question-number">第{{ index + 1 }}题</span>
              <span class="question-type">{{ question.type === 'single' ? '单选题' : '填空题' }}</span>
              <span class="question-score">{{ question.score }}分</span>
              <div class="question-actions">
                <el-button size="small" @click="moveQuestion(index, 'up')" :disabled="index === 0">
                  ↑
                </el-button>
                <el-button size="small" @click="moveQuestion(index, 'down')" :disabled="index === examInfo.questions.length - 1">
                  ↓
                </el-button>
                <el-button size="small" type="danger" @click="removeQuestion(index)">
                  删除
                </el-button>
              </div>
            </div>
            
            <div class="question-content">
              <el-form :model="question" :rules="questionRules">
                <el-form-item label="题目内容" prop="content">
                  <el-input 
                    v-model="question.content" 
                    type="textarea" 
                    :rows="3" 
                    placeholder="请输入题目内容"
                  />
                </el-form-item>
                
                <el-form-item label="分数" prop="score">
                  <el-input-number
                    v-model="question.score"
                    :min="1"
                    :max="examInfo.totalScore"
                    placeholder="请输入分数"
                  />
                </el-form-item>
                
                <!-- 单选题选项 -->
                <div v-if="question.type === 'single'">
                  <div class="options-section">
                    <div class="options-header">
                      <span>选项设置</span>
                      <el-button size="small" @click="addOption(question)">添加选项</el-button>
                    </div>
                    
                    <div v-for="(option, optIndex) in question.options" :key="optIndex" class="option-item">
                      <el-input
                        v-model="option.key"
                        placeholder="选项字母"
                        style="width: 80px; margin-right: 10px;"
                        maxlength="1"
                      />
                      <el-input
                        v-model="option.text"
                        placeholder="选项内容"
                        style="flex: 1; margin-right: 10px;"
                      />
                      <el-button 
                        size="small" 
                        type="danger" 
                        @click="removeOption(question, optIndex)"
                        :disabled="question.options.length <= 2"
                      >
                        删除
                      </el-button>
                    </div>
                    
                    <el-form-item label="正确答案" prop="correctAnswer">
                      <el-select v-model="question.correctAnswer" placeholder="请选择正确答案">
                        <el-option 
                          v-for="option in question.options" 
                          :key="option.key" 
                          :label="option.key.toUpperCase()" 
                          :value="option.key"
                        />
                      </el-select>
                    </el-form-item>
                  </div>
                </div>
                
                <!-- 填空题答案提示 -->
                <div v-else>
                  <el-form-item label="参考答案" prop="correctAnswer">
                    <el-input 
                      v-model="question.correctAnswer" 
                      type="textarea" 
                      :rows="2" 
                      placeholder="请输入参考答案（教师参考使用）"
                    />
                  </el-form-item>
                  
                  <!-- 填空题输入框配置 -->
                  <el-form-item label="输入框配置">
                    <div class="fill-input-config">
                      <div class="config-row">
                        <span class="config-label">输入框数量：</span>
                        <el-input-number 
                          v-model="question.inputCount" 
                          :min="1" 
                          :max="10" 
                          :step="1" 
                          size="small"
                          @change="(value) => updateInputCount(question, value)"
                        />
                      </div>
                      
                      <div class="config-row">
                        <span class="config-label">占位符文本：</span>
                        <el-input 
                          v-model="question.placeholder" 
                          size="small" 
                          placeholder="请输入占位符"
                          style="width: 200px;"
                        />
                      </div>
                    </div>
                  </el-form-item>
                  
                  <!-- 填空题预设答案 -->
                  <el-form-item label="预设答案">
                    <div class="fill-answers-config">
                      <div class="answer-item" v-for="(item, index) in question.answers" :key="index">
                        <span class="answer-label">第{{ index+1 }}空：</span>
                        <div class="answer-config">
                          <!-- 输入框类型选择和添加按钮并列显示 -->
                          <div class="type-add-row">
                            <div class="type-select-wrapper">
                              <span class="type-label">输入框类型：</span>
                              <el-select v-model="question.answers[index].type" size="small" style="width: 120px;">
                                <el-option label="文本" value="text" />
                                <el-option label="数字" value="number" />
                              </el-select>
                            </div>
                            <el-button 
                              size="small" 
                              type="primary" 
                              @click="addAnswerOption(question, index)"
                            >
                              添加答案
                            </el-button>
                          </div>
                          
                          <!-- 答案选项列表 - 删除按钮与文本框并列 -->
                          <div class="answer-options">
                            <div v-for="(answer, answerIndex) in item.options" :key="answerIndex" class="answer-option-item">
                              <span class="option-index">{{ answerIndex + 1 }}.</span>
                              <el-input 
                                v-model="question.answers[index].options[answerIndex]" 
                                size="small" 
                                placeholder="请输入标准答案"
                                class="answer-input"
                              />
                              <el-button 
                                size="small" 
                                type="danger" 
                                @click="removeAnswerOption(question, index, answerIndex)"
                                :disabled="item.options.length <= 1"
                                class="remove-btn"
                              >
                                删除
                              </el-button>
                            </div>
                          </div>
                        </div>
                      </div>
                      
                      <div class="answer-hint">
                        <span class="hint-text">提示：可以为每个填空设置多个可接受的答案选项，并为每个填空单独配置输入框类型</span>
                      </div>
                    </div>
                  </el-form-item>
                  
                  <!-- 预览填空题效果 -->
                  <el-form-item label="预览效果">
                    <div class="fill-preview">
                      <div class="preview-content">
                        <span>{{ question.content }}</span>
                        <template v-for="(item,index) in question.answers" :key="index">
                          <span class="fill-input-wrapper">
                            <span class="fill-input-number">({{ index+1 }})</span>
                            <el-input 
                              size="small"
                              :type="item.type || 'text'"
                              :placeholder="question.placeholder || '请输入答案'"
                              style="margin: 0 5px; width: 120px;"
                              disabled
                            />
                          </span>
                        </template>
                      </div>
                    </div>
                  </el-form-item>
                </div>
              </el-form>
            </div>
          </div>
        </div>
      </el-card>

      <div class="actions-footer">
        <el-button type="primary" size="large" @click="saveExam" :loading="saving">
          {{ isEditMode ? '保存修改' : '创建试卷' }}
        </el-button>
        <el-button size="large" @click="saveAsDraft">保存为草稿</el-button>
        <el-button size="large" @click="$router.back()">取消</el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { teacherAPI } from '@/services/api'

const route = useRoute()
const router = useRouter()

const examInfoFormRef = ref()
const saving = ref(false)
const isEditMode = computed(() => !!route.params.id)

const examInfo = reactive({
  title: '',
  description: '',
  totalScore: 100,
  questions: [],
  status: 'draft'
})

const examRules = {
  title: [{ required: true, message: '请输入试卷标题', trigger: 'blur' }],
  totalScore: [{ required: true, message: '请输入总分', trigger: 'blur' }]
}

const questionRules = {
  content: [{ required: true, message: '请输入题目内容', trigger: 'blur' }],
  score: [{ required: true, message: '请输入分数', trigger: 'blur' }]
}

const currentTotalScore = computed(() => {
  return examInfo.questions.reduce((sum, question) => sum + (question.score || 0), 0)
})

const addQuestion = (type) => {
  const question = {
    id: Date.now() + Math.random(),
    type: type,
    content: '',
    score: 10
  }
  
  question.correctAnswer = ''
  if (type === 'single') {
    question.options = [
      { key: 'a', text: '' },
      { key: 'b', text: '' },
      { key: 'c', text: '' },
      { key: 'd', text: '' }
    ]
  } else {
    question.inputCount = 1
    question.placeholder = '请输入答案'
    // 初始化每个填空项的配置
    question.answers = []
    for (let i = 0; i < question.inputCount; i++) {
      question.answers.push({
        type: 'text',
        options: [''] // 每个填空项至少有一个答案选项
      })
    }
  }
  
  examInfo.questions.push(question)
}

const removeQuestion = (index) => {
  examInfo.questions.splice(index, 1)
}

const moveQuestion = (index, direction) => {
  if (direction === 'up' && index > 0) {
    const temp = examInfo.questions[index - 1]
    examInfo.questions[index - 1] = examInfo.questions[index]
    examInfo.questions[index] = temp
  } else if (direction === 'down' && index < examInfo.questions.length - 1) {
    const temp = examInfo.questions[index + 1]
    examInfo.questions[index + 1] = examInfo.questions[index]
    examInfo.questions[index] = temp
  }
}

const addOption = (question) => {
  if (question.options.length < 6) {
    const lastKey = question.options[question.options.length - 1].key
    const nextKey = String.fromCharCode(lastKey.charCodeAt(0) + 1)
    question.options.push({ key: nextKey, text: '' })
  } else {
    ElMessage.warning('最多只能添加6个选项')
  }
}

const removeOption = (question, index) => {
  if (question.options.length > 2) {
    question.options.splice(index, 1)
    if (question.correctAnswer === question.options[index]?.key) {
      question.correctAnswer = ''
    }
  }
}

const saveExam = async () => {
  if (!examInfoFormRef.value) return
  
  try {
    await examInfoFormRef.value.validate()
    
    // 验证题目完整性
    for (const question of examInfo.questions) {
      if (!question.content.trim()) {
        ElMessage.error('请完善所有题目的内容')
        return
      }
      if (question.type === 'single' && !question.correctAnswer) {
        ElMessage.error('请为单选题设置正确答案')
        return
      }
      // 验证填空题配置
      if (question.type === 'fill') {
        if (!question.answers || question.answers.length === 0) {
          ElMessage.error('请为填空题配置答案选项')
          return
        }
        for (let i = 0; i < question.answers.length; i++) {
          const config = question.answers[i]
          if (!config.options || config.options.length === 0 || !config.options[0].trim()) {
            ElMessage.error(`请为第${i + 1}个填空配置有效的答案选项`)
            return
          }

          if (question.score % question.answers.length != 0) {
            ElMessage.error(`请确保每个填空项的答案数量与总分相匹配，例如总分100，每个填空项答案数量应为25个`)
            return
          }
        }
      }
    }
    
    // 验证总分匹配
    if (currentTotalScore.value !== examInfo.totalScore) {
      ElMessage.error('题目总分与试卷总分不匹配')
      return
    }
    
    saving.value = true
    
    try {
      // 准备请求数据
      const examData = {
        exam: {
          title: examInfo.title,
          description: examInfo.description,
          totalScore: examInfo.totalScore,
          status: examInfo.status,
          duration: 60 // 设置默认考试时长
        },
        questions: examInfo.questions.map(q => {
          const questionData = {
            type: q.type,
            content: q.content,
            score: q.score,
            options: q.options,
            answer: q.correctAnswer,
            inputCount: q.inputCount || 1,
            placeholder: q.placeholder || '请输入答案'
          }
          
          // 填空题需要保存完整的answers配置
          if (q.type === 'fill' && q.answers) {
            questionData.answers = q.answers
          }
          
          return questionData
        })
      }
      
      console.log('发送的试卷数据:', JSON.stringify(examData, null, 2))
      
      if (isEditMode.value) {
        // 更新试卷
        await teacherAPI.updateExam(route.params.id, examData)
        ElMessage.success('试卷修改成功')
      } else {
        // 创建新试卷
        await teacherAPI.createExam(examData)
        ElMessage.success('试卷创建成功')
      }
      
      router.push('/teacher/exam-list')
    } catch (error) {
      console.error('保存试卷失败:', error)
      ElMessage.error('保存失败：' + (error.response?.data?.error || error.message))
    }
    
  } catch (error) {
    console.error('表单验证失败:', error)
    ElMessage.error('请完善表单信息')
  } finally {
    saving.value = false
  }
}

const saveAsDraft = () => {
  examInfo.status = 'draft'
  saveExam()
}

// 添加填空题答案选项
const addAnswerOption = (question, fillIndex) => {
  // 确保answers数组存在且有足够长度
  if (!question.answers) {
    question.answers = []
  }
  
  // 确保指定索引的配置存在
  if (!question.answers[fillIndex]) {
    question.answers[fillIndex] = {
      type: 'text',
      options: ['']
    }
  }
  
  // 添加新的答案选项
  question.answers[fillIndex].options.push('')
}

// 删除填空题答案选项
const removeAnswerOption = (question, fillIndex, optionIndex) => {
  if (question.answers && question.answers[fillIndex] && question.answers[fillIndex].options.length > 1) {
    question.answers[fillIndex].options.splice(optionIndex, 1)
  }
}

// 更新填空题输入框数量
const updateInputCount = (question, newCount) => {
  // 确保answers数组存在
  if (!question.answers) {
    question.answers = []
  }
  
  const currentCount = question.answers.length
  const countDiff = newCount - currentCount
  
  if (countDiff > 0) {
    // 增加填空项
    for (let i = 0; i < countDiff; i++) {
      question.answers.push({
        type: 'text',
        options: ['']
      })
    }
  } else if (countDiff < 0) {
    // 减少填空项，删除多余的配置
    question.answers.splice(newCount, Math.abs(countDiff))
  }
}

// 如果是编辑模式，加载现有试卷数据
onMounted(async () => {
  if (isEditMode.value) {
    try {
      const response = await teacherAPI.getExamDetails(route.params.id)
      
      if (response && response.title) {
        // 处理试卷基本信息
        examInfo.title = response.title
        examInfo.description = response.description || ''
        examInfo.totalScore = response.totalScore || 100
        examInfo.status = response.status || 'draft'
        
        // 处理题目数据
        if (response.questions && Array.isArray(response.questions)) {
          examInfo.questions = response.questions.map(question => {
            const processedQuestion = {
              ...question,
              id: question.id,
              type: question.type,
              content: question.content,
              score: question.score || 10
            }
            
            processedQuestion.correctAnswer = question.correctAnswer || ''
            if (question.type === 'single') {
              // 处理单选题选项
              processedQuestion.options = question.options || []
            } else if (question.type === 'fill') {
              // 处理填空题配置
              processedQuestion.inputCount = question.answers.length || 1
              processedQuestion.placeholder = question.placeholder || '请输入答案'
              
              // 处理填空题的详细配置
              if (!question.answers) {
                // 如果没有配置，创建默认配置
                processedQuestion.answers = []
                for (let i = 0; i < processedQuestion.inputCount; i++) {
                  processedQuestion.answers.push({
                    type: 'text',
                    options: ['']
                  })
                }
              }
            }
            
            return processedQuestion
          })
          console.log('加载的试卷数据:', response.questions)
        }
        
        ElMessage.success('试卷数据加载成功')
      } else {
        ElMessage.error('获取试卷数据失败')
        router.push('/teacher/exam-list')
      }
    } catch (error) {
      console.error('加载试卷数据失败:', error)
      ElMessage.error('加载试卷数据失败')
      router.push('/teacher/exam-list')
    }
  }
})
</script>

<style scoped>
.exam-edit-page {
  min-height: 100%;
}

.questions-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.questions-list {
  max-height: 600px;
  overflow-y: auto;
}

.question-item {
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  margin-bottom: 20px;
  padding: 15px;
}

.question-header {
  display: flex;
  align-items: center;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid #f0f0f0;
}

.question-number {
  font-weight: bold;
  margin-right: 10px;
}

.question-type {
  background: #409eff;
  color: white;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  margin-right: 10px;
}

.question-score {
  color: #e6a23c;
  font-weight: bold;
  margin-right: auto;
}

.question-actions {
  display: flex;
  gap: 5px;
}

.options-section {
  margin-top: 15px;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 4px;
}

.options-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.option-item {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
}

.questions-summary {
  margin-top: 20px;
  padding: 15px;
  background: #f0f9ff;
  border-radius: 4px;
  text-align: center;
}

.questions-summary p {
  margin: 5px 0;
  font-size: 14px;
}

.score-match {
  color: #67c23a;
  font-weight: bold;
}

.score-mismatch {
  color: #e6a23c;
  font-weight: bold;
}

.actions-footer {
  text-align: center;
  margin-top: 30px;
  padding: 20px;
}

/* 填空题配置样式 */
.fill-input-config {
  background: #f8f9fa;
  border-radius: 4px;
  padding: 15px;
  margin-bottom: 15px;
}

.config-row {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
}

.config-label {
  width: 120px;
  font-size: 14px;
  color: #606266;
  margin-right: 15px;
}

.fill-preview {
  background: #f0f9ff;
  border-radius: 4px;
  padding: 15px;
  border: 1px dashed #409eff;
}

.preview-content {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 5px;
}

/* 填空题编号样式 */
.fill-input-wrapper {
  display: inline-flex;
  align-items: center;
  gap: 5px;
}

.fill-input-number {
  font-size: 12px;
  color: #409eff;
  font-weight: bold;
  min-width: 20px;
  text-align: center;
}

/* 填空题答案配置样式 */
.fill-answers-config {
  background: #f8f9fa;
  border-radius: 4px;
  padding: 15px;
  margin-bottom: 15px;
}

.answer-item {
  margin-bottom: 15px;
  padding: 10px;
  background: white;
  border-radius: 4px;
  border: 1px solid #e4e7ed;
}

.answer-label {
  display: inline-block;
  font-weight: bold;
  color: #409eff;
  margin-bottom: 10px;
  font-size: 14px;
}

.answer-config {
  margin-left: 20px;
}

.type-add-row {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 10px;
}

.type-select-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
}

.type-label {
  font-size: 12px;
  color: #606266;
  white-space: nowrap;
}

.answer-options {
  margin-top: 10px;
}

.answer-option-item {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
  padding: 8px;
  background: #fafafa;
  border-radius: 4px;
}

.option-index {
  font-weight: bold;
  color: #409eff;
  width: 20px;
  text-align: center;
  font-size: 12px;
}

.answer-input {
  flex: 1;
  margin-right: 10px;
}

.remove-btn {
  white-space: nowrap;
}

.answer-hint {
  margin-top: 15px;
  padding: 10px;
  background: #e6f7ff;
  border-radius: 4px;
  border-left: 4px solid #409eff;
}

.hint-text {
  font-size: 12px;
  color: #409eff;
  line-height: 1.4;
}
</style>
