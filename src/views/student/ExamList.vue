<template>
  <div class="exam-list-page">
    <div class="page-header">
      <h2>试卷列表</h2>
    </div>
    
    <div class="page-content">
      <el-card class="card-container">
        <template #header>
          <h3>可参加考试</h3>
        </template>
        
        <el-table :data="availableExams" style="width: 100%" v-loading="loading">
          <el-table-column prop="title" label="试卷名称" width="200" />
          <el-table-column prop="duration" label="考试时长" width="120">
            <template #default="scope">
              {{ scope.row.duration }}分钟
            </template>
          </el-table-column>
          <el-table-column prop="totalScore" label="总分" width="100" />
          <el-table-column prop="questionCount" label="题目数量" width="100" />
          <el-table-column prop="startTime" label="开始时间" width="180" />
          <el-table-column prop="endTime" label="结束时间" width="180" />
          <el-table-column label="状态" width="100">
            <template #default="scope">
              <el-tag :type="getStatusType(scope.row.status)">
                {{ scope.row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120">
            <template #default="scope">
              <el-button 
                type="primary" 
                size="small" 
                :disabled="scope.row.status !== '进行中'"
                @click="startExam(scope.row)"
              >
                开始考试
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

      <el-card class="card-container" style="margin-top: 20px;">
        <template #header>
          <h3>已参加考试</h3>
        </template>
        
        <el-table :data="completedExams" style="width: 100%">
          <el-table-column prop="title" label="试卷名称" width="200" />
          <el-table-column prop="score" label="得分" width="100" />
          <el-table-column prop="totalScore" label="总分" width="100" />
          <el-table-column prop="submitTime" label="提交时间" width="180" />
          <el-table-column prop="timeUsed" label="用时" width="120">
            <template #default="scope">
              {{ scope.row.timeUsed }}分钟
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.passed ? 'success' : 'danger'">
                {{ scope.row.passed ? '及格' : '不及格' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120">
            <template #default="scope">
              <el-button type="text" @click="viewExamResult(scope.row)">
                查看详情
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { studentAPI } from '@/services/api'
import { formatDateTime } from '@/utils/dateUtils'

const router = useRouter()

const availableExams = ref([])
const completedExams = ref([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const fetchExams = async () => {
  try {
    loading.value = true
    const response = await studentAPI.getExams({
      page: currentPage.value,
      pageSize: pageSize.value
    })
    
    // 处理分页数据
      availableExams.value = response.exams.map(exam => ({
        ...exam,
        id: exam.id,
        title: exam.title,
        duration: exam.duration,
        totalScore: exam.totalScore,
        questionCount: exam.questionCount,
        startTime: formatDateTime(exam.startTime),
        endTime: formatDateTime(exam.endTime),
        status: exam.status
      }))
      total.value = response.pagination?.total || 0
  } catch (error) {
    ElMessage.error('获取考试列表失败'+ error.message)
  } finally {
    loading.value = false
  }
}

const fetchResults = async () => {
  try {
    loading.value = true
    const response = await studentAPI.getResults({
      page: currentPage.value,
      pageSize: pageSize.value
    })
    
    // 处理分页数据
      completedExams.value = response.results.map(result => ({
        ...result,
        id: result.id,
        title: result.examTitle,
        score: result.score,
        totalScore: result.totalScore,
        submitTime: formatDateTime(result.submitTime),
        timeUsed: result.timeUsed,
        passed: result.passed
      }))
      total.value = response.pagination?.total || 0
  } catch (error) {
    ElMessage.error('获取考试结果失败')
  } finally {
    loading.value = false
  }
}

const getStatusType = (status) => {
  const types = {
    '进行中': 'success',
    '未开始': 'info',
    '已结束': 'danger'
  }
  return types[status] || 'info'
}

const startExam = (exam) => {
  if (exam.status === '进行中') {
    router.push(`/student/exam/${exam.assignmentId}`)
  } else {
    ElMessage.warning('当前无法参加此考试')
  }
}

const viewExamResult = (exam) => {
  // 跳转到考试结果详情页面
  router.push(`/student/exam-result/${exam.id}`)
}

// 监听分页参数变化
watch([currentPage, pageSize], () => {
  fetchExams()
  fetchResults()
})

onMounted(async () => {
  await Promise.all([
    fetchExams(),
    fetchResults()
  ])
})
</script>

<style scoped>
.exam-list-page {
  min-height: 100%;
}
</style>
