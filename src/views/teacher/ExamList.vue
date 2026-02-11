<template>
  <div class="exam-list-page">
    <div class="page-header">
      <h2>试卷列表</h2>
    </div>
    
    <div class="page-content">
      <el-card class="card-container">
        <template #header>
          <div class="card-header">
            <h3>试卷管理</h3>
            <div>
              <el-button type="primary" @click="$router.push('/teacher/exam-edit')">
                创建新试卷
              </el-button>
              <el-button @click="exportExams">导出试卷</el-button>
            </div>
          </div>
        </template>
        
        <div class="table-toolbar">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索试卷名称"
            style="width: 300px"
            clearable
          />
          <el-select v-model="filterStatus" placeholder="试卷状态" style="width: 150px; margin-left: 10px;">
            <el-option label="全部状态" value="" />
            <el-option label="已发布" value="published" />
            <el-option label="草稿" value="draft" />
            <el-option label="已归档" value="archived" />
          </el-select>
        </div>
        
        <el-table :data="exams" style="width: 100%">
          <el-table-column prop="title" label="试卷名称" width="200" />
          <el-table-column prop="questionCount" label="题目数量" width="100" />
          <el-table-column prop="totalScore" label="总分" width="100" />
          <el-table-column label="题型分布" width="200">
            <template #default="scope">
              <div class="question-types">
                <el-tag v-if="scope.row.singleChoiceCount > 0" size="small">
                  单选 {{ scope.row.singleChoiceCount }}
                </el-tag>
                <el-tag v-if="scope.row.fillBlankCount > 0" type="success" size="small">
                  填空 {{ scope.row.fillBlankCount }}
                </el-tag>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="createTime" label="创建时间" width="180" />
          <el-table-column prop="updateTime" label="更新时间" width="180" />
          <el-table-column label="状态" width="100">
            <template #default="scope">
              <el-tag :type="getStatusType(scope.row.status)">
                {{ getStatusText(scope.row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="350">
            <template #default="scope">
              <el-button size="small" @click="previewExam(scope.row)">预览</el-button>
              <el-button size="small" type="primary" @click="editExam(scope.row)">编辑</el-button>
              <el-button 
                size="small" 
                :type="scope.row.status === 'published' ? 'warning' : 'success'"
                @click="toggleExamStatus(scope.row)"
              >
                {{ scope.row.status === 'published' ? '下架' : '发布' }}
              </el-button>
              <el-button size="small" type="danger" @click="deleteExam(scope.row)">删除</el-button>
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
    </div>

    <!-- 试卷预览对话框 -->
    <el-dialog v-model="showPreviewDialog" title="试卷预览" width="800px" top="5vh">
      <div v-if="previewExamData" class="preview-container">
        <h3>{{ previewExamData.title }}</h3>
        <p class="exam-info">总分：{{ previewExamData.totalScore }}分 | 题目数量：{{ previewExamData.questionCount }}</p>
        
        <div v-for="(question, index) in previewExamData.questions" :key="question.id" class="preview-question">
          <h4>第{{ index + 1 }}题 ({{ question.type === 'single' ? '单选题' : '填空题' }}) - {{ question.score }}分</h4>
          <p class="question-text">{{ question.content }}</p>
          
          <div v-if="question.type === 'single'" class="preview-options">
            <div v-for="option in question.options" :key="option.key" class="option-item">
              <span class="option-key">{{ option.key.toUpperCase() }}.</span>
              <span class="option-text">{{ option.text }}</span>
            </div>
          </div>
          
          <div v-else class="preview-fill-blank">
            <div class="fill-preview-content">
              <span>{{ question.content }}</span>
              <template v-for="(item, index) in question.answers" :key="index">
                <span class="fill-input-wrapper">
                  <span class="fill-input-number">({{ index + 1 }})</span>
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
            <div v-if="question.answers && question.answers.length > 0" class="fill-hint">
              <span class="hint-text">共 {{ question.answers.length }} 个填空</span>
            </div>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { teacherAPI } from '@/services/api'

const router = useRouter()

const searchKeyword = ref('')
const filterStatus = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const showPreviewDialog = ref(false)
const previewExamData = ref(null)

const exams = ref([])
const loading = ref(false)

const fetchExams = async () => {
  try {
    loading.value = true
    const response = await teacherAPI.getExams({
      page: currentPage.value,
      pageSize: pageSize.value,
      keyword: searchKeyword.value,
      status: filterStatus.value
    })
    
    // 检查响应数据是否存在
    if (!response || !response.data) {
      throw new Error('无效的响应数据')
    }
    
    exams.value = Array.isArray(response.data) 
      ? response.data.map(exam => ({
          ...exam,
          id: exam.id,
          title: exam.title,
          questionCount: exam.questionCount,
          singleChoiceCount: exam.singleChoiceCount,
          fillBlankCount: exam.fillBlankCount,
          totalScore: exam.totalScore,
          status: exam.status,
          createTime: new Date(exam.createTime).toLocaleString(),
          updateTime: new Date(exam.updateTime).toLocaleString(),
        }))
      : []
    
    // 检查分页数据
    if (response.pagination && typeof response.pagination.total === 'number') {
      total.value = response.pagination.total
    }
  } catch (error) {
    ElMessage.error('获取试卷列表失败')
  } finally {
    loading.value = false
  }
}

const getStatusType = (status) => {
  const types = {
    'published': 'success',
    'draft': 'info',
    'archived': 'warning'
  }
  return types[status] || 'info'
}

const getStatusText = (status) => {
  const texts = {
    'published': '已发布',
    'draft': '草稿',
    'archived': '已归档'
  }
  return texts[status] || status
}

const previewExam = async (exam) => {
  try {
    loading.value = true
    const response = await teacherAPI.getExamDetails(exam.id)
    
    if (response && response.questions) {
      previewExamData.value = {
        ...exam,
        questions: response.questions.map(question => ({
          id: question.id,
          type: question.type,
          content: question.content,
          score: question.score,
          options: question.options || [],
          // 填空题配置
          answers: question.answers || [],
          inputCount: question.inputCount || 1,
          placeholder: question.placeholder || '请输入答案'
        }))
      }
      showPreviewDialog.value = true
    } else {
      ElMessage.error('获取试卷题目失败')
    }
  } catch (error) {
    console.error('预览试卷错误:', error)
    ElMessage.error('获取试卷题目失败')
  } finally {
    loading.value = false
  }
}

const editExam = (exam) => {
  router.push(`/teacher/exam-edit/${exam.id}`)
}

const toggleExamStatus = async (exam) => {
  const action = exam.status === 'published' ? '下架' : '发布'
  try {
    await ElMessageBox.confirm(
      `确定要${action}试卷 "${exam.title}" 吗？`,
      `${action}试卷`,
      {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      }
    )
    
    // 调用API更新试卷状态
    const response = await teacherAPI.updateExamStatus(exam.id, {
      status: exam.status === 'published' ? 'draft' : 'published'
    })
    
    if (response && response.success) {
      // 更新本地数据
      exam.status = exam.status === 'published' ? 'draft' : 'published'
      exam.updateTime = new Date().toLocaleString()
      ElMessage.success(`${action}成功`)
      
      // 重新获取试卷列表以确保数据同步
      fetchExams()
    } else {
      ElMessage.error(`${action}失败：${response?.message || '未知错误'}`)
    }
  } catch (error) {
    console.error('切换试卷状态错误:', error)
    ElMessage.error(`${action}失败：${error.message || '网络错误'}`)
  }
}

const deleteExam = async (exam) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除试卷 "${exam.title}" 吗？此操作不可恢复。`,
      '删除试卷',
      {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      }
    )
    
    exams.value = exams.value.filter(e => e.id !== exam.id)
    ElMessage.success('试卷删除成功')
  } catch {
    // 用户取消操作
  }
}

const exportExams = () => {
  ElMessage.success('导出功能待实现')
}

// 监听分页参数变化
watch([currentPage, pageSize], () => {
  fetchExams()
})

// 监听搜索关键词和状态变化
watch([searchKeyword, filterStatus], () => {
  currentPage.value = 1 // 重置到第一页
  fetchExams()
})

// 从API获取数据
onMounted(async () => {
  fetchExams()
})
</script>

<style scoped>
.exam-list-page {
  min-height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.table-toolbar {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
}

.question-types {
  display: flex;
  gap: 5px;
}

.pagination {
  margin-top: 20px;
  text-align: right;
}

.preview-container {
  max-height: 60vh;
  overflow-y: auto;
}

.exam-info {
  color: #666;
  margin-bottom: 20px;
  padding-bottom: 10px;
  border-bottom: 1px solid #eee;
}

.preview-question {
  margin-bottom: 30px;
  padding: 15px;
  border: 1px solid #eee;
  border-radius: 4px;
}

.question-text {
  font-size: 14px;
  line-height: 1.6;
  margin: 10px 0;
}

.preview-options {
  margin: 15px 0;
}

.option-item {
  display: flex;
  align-items: flex-start;
  margin: 8px 0;
}

.option-key {
  font-weight: bold;
  margin-right: 8px;
  min-width: 20px;
}

.preview-fill-blank {
  margin: 15px 0;
  padding: 15px;
  background-color: #f8f9fa;
  border-radius: 4px;
}

.fill-preview-content {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 5px;
  line-height: 1.8;
}

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

.fill-hint {
  margin-top: 10px;
  padding: 8px;
  background: #e6f7ff;
  border-radius: 4px;
  border-left: 3px solid #409eff;
}

.hint-text {
  font-size: 12px;
  color: #409eff;
}
</style>
