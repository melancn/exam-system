<template>
  <div class="scores-page">
    <div class="page-header">
      <h2>考试成绩</h2>
    </div>
    
    <div class="page-content">
      <el-card class="card-container">
        <template #header>
          <h3>成绩统计</h3>
        </template>
        
        <el-row :gutter="20">
          <el-col :span="6">
            <div class="stat-card">
              <div class="stat-number">{{ stats.totalExams }}</div>
              <div class="stat-label">总考试次数</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card">
              <div class="stat-number">{{ stats.avgScore }}</div>
              <div class="stat-label">平均分</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card">
              <div class="stat-number">{{ stats.passRate }}%</div>
              <div class="stat-label">及格率</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card">
              <div class="stat-number">{{ stats.bestScore }}</div>
              <div class="stat-label">最高分</div>
            </div>
          </el-col>
        </el-row>
        
        <div class="chart-container" style="margin-top: 30px;">
          <h4>成绩趋势</h4>
          <div ref="scoreChart" style="height: 300px;"></div>
        </div>
      </el-card>

      <el-card class="card-container" style="margin-top: 20px;">
        <template #header>
          <h3>考试记录</h3>
        </template>
        
        <el-table :data="examRecords" style="width: 100%">
          <el-table-column prop="examTitle" label="试卷名称" width="200" />
          <el-table-column prop="score" label="得分" width="100">
            <template #default="scope">
              <span :class="getScoreClass(scope.row.score)">{{ scope.row.score }}</span>
            </template>
          </el-table-column>
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
              <el-button type="text" @click="viewExamDetail(scope.row)">
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
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import { studentAPI } from '@/services/api'

const scoreChart = ref()

const stats = ref({
  totalExams: 0,
  avgScore: 0,
  passRate: 0,
  bestScore: 0
})

const examRecords = ref([])

const getScoreClass = (score) => {
  if (score >= 90) return 'score-excellent'
  if (score >= 80) return 'score-good'
  if (score >= 60) return 'score-pass'
  return 'score-fail'
}

const loading = ref(false)

// 获取真实考试成绩数据
const fetchScores = async () => {
  loading.value = true
  try {
    const results = await studentAPI.getResults()
    
    // 设置考试记录
    examRecords.value = results.map(result => ({
      id: result.id,
      examTitle: result.examTitle,
      score: result.score,
      totalScore: result.totalScore,
      submitTime: result.submitTime,
      timeUsed: result.timeUsed,
      passed: result.passed
    }))
    
    // 计算统计数据
    const totalExams = examRecords.value.length
    const avgScore = totalExams > 0 
      ? examRecords.value.reduce((sum, record) => sum + record.score, 0) / totalExams 
      : 0
    const bestScore = totalExams > 0 
      ? Math.max(...examRecords.value.map(record => record.score)) 
      : 0
    const passedCount = examRecords.value.filter(record => record.passed).length
    const passRate = totalExams > 0 
      ? Math.round((passedCount / totalExams) * 100) 
      : 0
    
    stats.value = {
      totalExams,
      avgScore: Math.round(avgScore * 100) / 100,
      passRate,
      bestScore
    }
    
    // 初始化图表
    initChart()
  } catch (error) {
    console.error('获取成绩失败:', error)
    ElMessage.error('获取成绩失败，请稍后重试')
  } finally {
    loading.value = false
  }
}

const viewExamDetail = (record) => {
  ElMessage.info(`查看考试 ${record.examTitle} 的详细结果`)
}

const initChart = () => {
  const chart = echarts.init(scoreChart.value)
  
  // 从考试记录中提取数据
  const dates = examRecords.value.map(record => {
    const date = new Date(record.submitTime)
    return `${date.getMonth() + 1}月${date.getDate()}日`
  }).reverse()
  
  const scores = examRecords.value.map(record => record.score).reverse()
  
  const option = {
    tooltip: {
      trigger: 'axis'
    },
    xAxis: {
      type: 'category',
      data: dates.length > 0 ? dates : ['暂无数据']
    },
    yAxis: {
      type: 'value',
      max: 100,
      min: 0
    },
    series: [
      {
        name: '考试成绩',
        type: 'line',
        data: scores.length > 0 ? scores : [0],
        smooth: true,
        lineStyle: {
          color: '#409eff'
        },
        itemStyle: {
          color: '#409eff'
        }
      },
      {
        name: '及格线',
        type: 'line',
        data: scores.length > 0 ? new Array(scores.length).fill(60) : [60],
        lineStyle: {
          color: '#67c23a',
          type: 'dashed'
        },
        itemStyle: {
          color: '#67c23a'
        }
      }
    ]
  }
  chart.setOption(option)
}

onMounted(() => {
  fetchScores()
})
</script>

<style scoped>
.scores-page {
  min-height: 100%;
}

.stat-card {
  text-align: center;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
}

.stat-number {
  font-size: 32px;
  font-weight: bold;
  color: #409eff;
  margin-bottom: 8px;
}

.stat-label {
  color: #666;
  font-size: 14px;
}

.chart-container {
  margin-top: 30px;
}

.score-excellent {
  color: #67c23a;
  font-weight: bold;
}

.score-good {
  color: #409eff;
  font-weight: bold;
}

.score-pass {
  color: #e6a23c;
  font-weight: bold;
}

.score-fail {
  color: #f56c6c;
  font-weight: bold;
}
</style>
