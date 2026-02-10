<template>
  <div class="results-analysis-page">
    <div v-if="loading" class="loading-overlay">
      <el-loading :visible="true" />
    </div>
    
    <div class="page-header">
      <h2>考试结果分析</h2>
      <div class="header-actions">
        <el-button type="primary" @click="exportReport">导出报告</el-button>
        <el-button @click="refreshData">刷新数据</el-button>
      </div>
    </div>
    
    <div class="page-content">
      <el-card class="card-container">
        <template #header>
          <div class="card-header">
            <h3>统计分析</h3>
            <div class="real-time-indicator">
              <el-tag :type="isWebSocketConnected ? 'success' : 'danger'" size="small">
                {{ isWebSocketConnected ? '实时数据已连接' : '实时数据已断开' }}
              </el-tag>
            </div>
          </div>
        </template>
        
        <el-row :gutter="20">
          <el-col :span="6">
            <div class="stat-card">
              <div class="stat-number">{{ stats.totalExams }}</div>
              <div class="stat-label">考试场次</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card">
              <div class="stat-number">{{ stats.totalStudents }}</div>
              <div class="stat-label">参考学生</div>
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
        </el-row>
        
        <!-- 筛选条件 -->
        <div class="filter-section" style="margin-top: 20px; padding: 15px; background: #f8f9fa; border-radius: 8px; border: 1px solid #e4e7ed;">
          <div class="filter-title" style="margin-bottom: 15px; font-weight: bold; color: #303133;">筛选条件</div>
          <el-row :gutter="15">
            <el-col :span="6">
              <el-select v-model="filterExam" placeholder="选择考试" style="width: 100%;" @change="handleFilterChange">
                <el-option label="所有考试" value="" />
                <el-option 
                  v-for="exam in exams" 
                  :key="exam.id" 
                  :label="exam.title" 
                  :value="exam.id"
                />
              </el-select>
            </el-col>
            <el-col :span="6">
              <el-select v-model="filterClass" placeholder="选择班级" style="width: 100%;" @change="handleFilterChange">
                <el-option label="所有班级" value="" />
                <el-option 
                  v-for="classItem in classes" 
                  :key="classItem.id" 
                  :label="classItem.name" 
                  :value="classItem.id"
                />
              </el-select>
            </el-col>
            <el-col :span="6">
              <el-select v-model="filterScore" placeholder="成绩筛选" style="width: 100%;" @change="handleFilterChange">
                <el-option label="全部" value="" />
                <el-option label="优秀(90+)" value="excellent" />
                <el-option label="良好(80-89)" value="good" />
                <el-option label="及格(60-79)" value="pass" />
                <el-option label="不及格(<60)" value="fail" />
              </el-select>
            </el-col>
            <el-col :span="6">
              <el-input 
                v-model="searchKeyword" 
                placeholder="搜索学生姓名或学号" 
                style="width: 100%;"
                prefix-icon="Search"
                @input="handleSearch"
              />
            </el-col>
          </el-row>
        </div>
        
        <el-row :gutter="20" style="margin-top: 30px;">
          <el-col :span="12">
            <div class="chart-container">
              <div class="chart-header">
                <h4>成绩分布</h4>
                <el-select v-model="scoreRange" placeholder="选择分数段" size="small" style="width: 150px;">
                  <el-option label="5分段" value="5" />
                  <el-option label="10分段" value="10" />
                  <el-option label="20分段" value="20" />
                </el-select>
              </div>
              <div ref="scoreDistributionChart" style="height: 300px;"></div>
            </div>
          </el-col>
          <el-col :span="12">
            <div class="chart-container">
              <div class="chart-header">
                <h4>班级对比</h4>
                <el-select v-model="comparisonMetric" placeholder="对比指标" size="small" style="width: 120px;">
                  <el-option label="平均分" value="avg" />
                  <el-option label="及格率" value="pass" />
                  <el-option label="优秀率" value="excellent" />
                </el-select>
              </div>
              <div ref="classComparisonChart" style="height: 300px;"></div>
            </div>
          </el-col>
        </el-row>
      </el-card>

      <!-- 实时考试信息列表 -->
      <el-card class="card-container" style="margin-top: 20px;">
        <template #header>
          <div class="card-header">
            <h3>实时考试信息</h3>
            <div class="real-time-controls">
              <el-button type="primary" size="small" @click="refreshRealTimeData">刷新</el-button>
              <el-tag :type="isRealTimeConnected ? 'success' : 'danger'" size="small">
                {{ isRealTimeConnected ? '实时数据已连接' : '实时数据已断开' }}
              </el-tag>
            </div>
          </div>
        </template>
        
        <el-table :data="realTimeExamData" style="width: 100%">
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

      <el-card class="card-container" style="margin-top: 20px;">
        <template #header>
          <div class="card-header">
            <h3>考试记录</h3>
            <div class="filter-container">
              <el-select v-model="filterExam" placeholder="选择考试" style="width: 200px; margin-right: 10px;">
                <el-option label="所有考试" value="" />
                <el-option 
                  v-for="exam in exams" 
                  :key="exam.id" 
                  :label="exam.title" 
                  :value="exam.id"
                />
              </el-select>
              <el-select v-model="filterClass" placeholder="选择班级" style="width: 150px; margin-right: 10px;">
                <el-option label="所有班级" value="" />
                <el-option 
                  v-for="classItem in classes" 
                  :key="classItem.id" 
                  :label="classItem.name" 
                  :value="classItem.id"
                />
              </el-select>
              <el-select v-model="filterScore" placeholder="成绩筛选" style="width: 120px; margin-right: 10px;">
                <el-option label="全部" value="" />
                <el-option label="优秀(90+)" value="excellent" />
                <el-option label="良好(80-89)" value="good" />
                <el-option label="及格(60-79)" value="pass" />
                <el-option label="不及格(<60)" value="fail" />
              </el-select>
              <el-input 
                v-model="searchKeyword" 
                placeholder="搜索学生姓名或学号" 
                style="width: 200px;"
                prefix-icon="Search"
              />
            </div>
          </div>
        </template>
        
        <el-table :data="filteredResults" style="width: 100%">
          <el-table-column prop="examTitle" label="考试名称" width="200" />
          <el-table-column prop="studentName" label="学生姓名" width="120" />
          <el-table-column prop="studentId" label="学号" width="120" />
          <el-table-column prop="className" label="班级" width="150" />
          <el-table-column prop="score" label="得分" width="100">
            <template #default="scope">
              <span :class="getScoreClass(scope.row.score)">{{ scope.row.score }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="totalScore" label="总分" width="100" />
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
          <el-table-column prop="submitTime" label="提交时间" width="180" />
          <el-table-column label="操作" width="120">
            <template #default="scope">
              <el-button type="text" @click="viewExamDetail(scope.row)">查看详情</el-button>
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

      <!-- 考试详情对话框 -->
      <el-dialog v-model="showDetailDialog" :title="`${selectedResult?.studentName} - ${selectedResult?.examTitle}`" width="800px">
        <div v-if="selectedResult" class="exam-detail">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="学号">{{ selectedResult.studentId }}</el-descriptions-item>
            <el-descriptions-item label="姓名">{{ selectedResult.studentName }}</el-descriptions-item>
            <el-descriptions-item label="班级">{{ selectedResult.className }}</el-descriptions-item>
            <el-descriptions-item label="得分">{{ selectedResult.score }}/{{ selectedResult.totalScore }}</el-descriptions-item>
            <el-descriptions-item label="用时">{{ selectedResult.timeUsed }}分钟</el-descriptions-item>
            <el-descriptions-item label="提交时间">{{ selectedResult.submitTime }}</el-descriptions-item>
          </el-descriptions>
          
          <div class="question-analysis" style="margin-top: 20px;">
            <h4>题目分析</h4>
            <el-table :data="selectedResult.questionDetails" style="width: 100%">
              <el-table-column prop="question" label="题目" width="300" />
              <el-table-column prop="studentAnswer" label="学生答案" width="150" />
              <el-table-column prop="correctAnswer" label="正确答案" width="150" />
              <el-table-column prop="score" label="得分" width="100" />
              <el-table-column label="状态" width="100">
                <template #default="scope">
                  <el-tag :type="scope.row.correct ? 'success' : 'danger'" size="small">
                    {{ scope.row.correct ? '正确' : '错误' }}
                  </el-tag>
                </template>
              </el-table-column>
            </el-table>
          </div>
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
                <el-button type="text" @click="sendBroadcastMessage(scope.row)">发送消息</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-dialog>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import * as XLSX from 'xlsx'
import { teacherAPI } from '../../services/api'
import { mockExamDetail, websocketConfig, chartConfig } from '../../utils/mockData'

const scoreDistributionChart = ref()
const classComparisonChart = ref()

const filterExam = ref('')
const filterClass = ref('')
const filterScore = ref('')
const searchKeyword = ref('')
const scoreRange = ref('10')
const comparisonMetric = ref('avg')
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const showDetailDialog = ref(false)
const selectedResult = ref(null)
const showStudentsDialog = ref(false)
const selectedExam = ref(null)

const loading = ref(false)
const stats = ref({
  totalExams: 0,
  totalStudents: 0,
  avgScore: 0,
  passRate: 0
})

const exams = ref([])
const classes = ref([])
const examResults = ref([])

// 防抖搜索
const debouncedSearch = ref(null)

const filteredResults = computed(() => {
  let result = examResults.value
  
  // 考试筛选
  if (filterExam.value) {
    result = result.filter(item => item.examId === filterExam.value)
  }
  
  // 班级筛选
  if (filterClass.value) {
    result = result.filter(item => item.classId === filterClass.value)
  }
  
  // 成绩筛选
  if (filterScore.value) {
    switch (filterScore.value) {
      case 'excellent':
        result = result.filter(item => item.score >= 90)
        break
      case 'good':
        result = result.filter(item => item.score >= 80 && item.score < 90)
        break
      case 'pass':
        result = result.filter(item => item.score >= 60 && item.score < 80)
        break
      case 'fail':
        result = result.filter(item => item.score < 60)
        break
    }
  }
  
  // 搜索筛选（使用防抖）
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(item => 
      item.studentName.toLowerCase().includes(keyword) ||
      item.studentId.toLowerCase().includes(keyword)
    )
  }
  
  return result
})

const getScoreClass = (score) => {
  if (score >= 90) return 'score-excellent'
  if (score >= 80) return 'score-good'
  if (score >= 60) return 'score-pass'
  return 'score-fail'
}

const viewExamDetail = (result) => {
  // 模拟获取详细的题目分析数据
  fetchExamDetail(result.id).then(detail => {
    selectedResult.value = {
      ...result,
      questionDetails: detail.questions
    }
    showDetailDialog.value = true
  }).catch(error => {
    console.error('获取考试详情失败:', error)
    ElMessage.error('获取考试详情失败')
  })
}

const fetchExamDetail = async (examResultId) => {
  try {
    const response = await teacherAPI.getExamDetail(examResultId)
    return response
  } catch (error) {
    console.error('获取考试详情失败:', error)
    // 使用模拟数据作为降级方案
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve(mockExamDetail)
      }, 500)
    })
  }
}

// 导出报告功能
const exportReport = async () => {
  try {
    loading.value = true
    
    // 获取考试分析数据
    const reportData = await teacherAPI.exportExamReport()
    
    // 创建工作簿
    const workbook = XLSX.utils.book_new()
    
    // 1. 创建统计概览工作表
    const summaryData = [
      ['考试分析报告', new Date().toLocaleDateString()],
      [''],
      ['统计指标', '数值'],
      ['考试场次', reportData.stats.totalExams],
      ['参考学生', reportData.stats.totalStudents],
      ['平均分', reportData.stats.avgScore],
      ['及格率', reportData.stats.passRate + '%'],
      [''],
      ['分数段分布'],
      ['分数段', '人数'],
      ['90分以上', reportData.scoreDistribution.excellent],
      ['80-89分', reportData.scoreDistribution.good],
      ['60-79分', reportData.scoreDistribution.pass],
      ['不及格(<60)', reportData.scoreDistribution.fail],
    ]
    const summarySheet = XLSX.utils.aoa_to_sheet(summaryData)
    XLSX.utils.book_append_sheet(workbook, summarySheet, '统计概览')
    
    // 2. 创建题目分析工作表
    const questionData = [
      ['题目ID', '题目内容', '题目类型', '分值', '正确答案', '正确率']
    ]
    reportData.questions.forEach(question => {
      questionData.push([
        question.id,
        question.content,
        question.type === 'single' ? '单选题' : '填空题',
        question.score,
        question.answer,
        question.correctRate + '%'
      ])
    })
    const questionSheet = XLSX.utils.aoa_to_sheet(questionData)
    XLSX.utils.book_append_sheet(workbook, questionSheet, '题目分析')
    
    // 3. 创建学生成绩工作表
    const studentData = [
      ['学号', '姓名', '班级', '考试名称', '得分', '总分', '用时(分钟)', '状态']
    ]
    reportData.studentResults.forEach(result => {
      studentData.push([
        result.studentId,
        result.studentName,
        result.className,
        result.examTitle,
        result.score,
        result.totalScore,
        result.timeUsed,
        result.passed ? '及格' : '不及格'
      ])
    })
    const studentSheet = XLSX.utils.aoa_to_sheet(studentData)
    XLSX.utils.book_append_sheet(workbook, studentSheet, '学生成绩')
    
    // 4. 创建班级统计工作表
    const classData = [
      ['班级名称', '学生人数', '平均分', '及格率', '优秀率']
    ]
    reportData.classStatistics.forEach(classStat => {
      classData.push([
        classStat.className,
        classStat.studentCount,
        classStat.avgScore,
        classStat.passRate + '%',
        classStat.excellentRate + '%'
      ])
    })
    const classSheet = XLSX.utils.aoa_to_sheet(classData)
    XLSX.utils.book_append_sheet(workbook, classSheet, '班级统计')
    
    // 生成Excel文件并下载
    const fileName = `考试分析报告_${new Date().toISOString().slice(0, 10)}.xlsx`
    XLSX.writeFile(workbook, fileName)
    
    ElMessage.success('报告导出成功')
  } catch (error) {
    console.error('导出失败:', error)
    ElMessage.error('导出失败，请重试')
  } finally {
    loading.value = false
  }
}

const initCharts = () => {
  // 成绩分布图表
  const distributionChart = echarts.init(scoreDistributionChart.value)
  const distributionOption = {
    tooltip: {
      trigger: 'axis'
    },
    xAxis: {
      type: 'category',
      data: ['0-59', '60-69', '70-79', '80-89', '90-100']
    },
    yAxis: {
      type: 'value',
      name: '人数'
    },
    series: [{
      name: '人数',
      type: 'bar',
      data: [15, 25, 35, 30, 20],
      itemStyle: {
        color: '#409eff'
      }
    }]
  }
  distributionChart.setOption(distributionOption)

  // 班级对比图表
  const comparisonChart = echarts.init(classComparisonChart.value)
  const comparisonOption = {
    tooltip: {
      trigger: 'axis'
    },
    legend: {
      data: ['平均分', '及格率']
    },
    xAxis: {
      type: 'category',
      data: ['计算机1班', '计算机2班', '软件1班', '软件2班']
    },
    yAxis: [
      {
        type: 'value',
        name: '分数',
        min: 0,
        max: 100
      },
      {
        type: 'value',
        name: '百分比',
        min: 0,
        max: 100
      }
    ],
    series: [
      {
        name: '平均分',
        type: 'bar',
        data: [85.5, 82.3, 79.8, 81.2],
        itemStyle: {
          color: '#409eff'
        }
      },
      {
        name: '及格率',
        type: 'line',
        yAxisIndex: 1,
        data: [92, 89, 85, 88],
        itemStyle: {
          color: '#67c23a'
        }
      }
    ]
  }
  comparisonChart.setOption(comparisonOption)
}

// 实时考试信息数据
const realTimeExamData = ref([])
const isRealTimeConnected = ref(false)
const realTimeWs = ref(null)
const realTimeRetryCount = ref(0)

// 初始化实时考试信息WebSocket连接
const initRealTimeExamData = () => {
  const token = localStorage.getItem('token')
  if (!token) {
    ElMessage.error('请先登录')
    return
  }

  const wsUrl = `${config.websocket.baseURL}${config.websocket.endpoint}`
  
  try {
    realTimeWs.value = new WebSocket(wsUrl)

    realTimeWs.value.onopen = () => {
      console.log('实时考试信息WebSocket连接已建立')
      isRealTimeConnected.value = true
      realTimeRetryCount.value = 0 // 重置重试次数
      
      // 发送认证消息
      const authMessage = {
        type: 'auth',
        token: token
      }
      realTimeWs.value.send(JSON.stringify(authMessage))
    }

    realTimeWs.value.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        console.log('收到实时考试信息数据:', data)
        handleRealTimeMessage(data)
      } catch (error) {
        console.error('解析WebSocket消息失败:', error)
      }
    }

    realTimeWs.value.onerror = (error) => {
      console.error('实时考试信息WebSocket错误:', error)
      isRealTimeConnected.value = false
    }

    realTimeWs.value.onclose = (event) => {
      console.log('实时考试信息WebSocket连接已关闭:', event.code, event.reason)
      isRealTimeConnected.value = false
      
      // 如果是意外断开且未达到最大重试次数，则尝试重连
      if (event.code !== 1000 && realTimeRetryCount.value < config.websocket.maxReconnectAttempts) {
        realTimeRetryCount.value++
        console.log(`WebSocket连接断开，${config.websocket.reconnectInterval / 1000}秒后尝试重连(${realTimeRetryCount.value}/${config.websocket.maxReconnectAttempts})`)
        
        setTimeout(() => {
          initRealTimeExamData()
        }, config.websocket.reconnectInterval)
      }
    }
  } catch (error) {
    console.error('创建实时考试信息WebSocket连接失败:', error)
    ElMessage.error('实时数据连接失败')
  }
}

// 处理实时消息
const handleRealTimeMessage = (data) => {
  if (!data || !data.type) return
  
  switch (data.type) {
    case 'auth_success':
      console.log('实时考试信息WebSocket认证成功')
      // 请求当前考试状态
      requestCurrentExamStatus()
      break
    case 'exam_status':
      updateRealTimeExamData(data)
      break
    case 'student_start':
      updateStudentStart(data)
      break
    case 'student_end':
      updateStudentEnd(data)
      break
    case 'update':
      updateStudentTime(data)
      break
    default:
      console.log('未知的实时消息类型:', data.type)
  }
}

// 请求当前考试状态
const requestCurrentExamStatus = () => {
  if (realTimeWs.value && realTimeWs.value.readyState === WebSocket.OPEN) {
    const message = {
      type: 'get_exam_status',
      examId: 0 // 0表示获取所有考试状态
    }
    realTimeWs.value.send(JSON.stringify(message))
  }
}

// 更新实时考试数据
const updateRealTimeExamData = (data) => {
  // 根据WebSocket推送的心跳数据更新考试信息
  const examData = {
    examId: data.examId,
    examTitle: `考试场次 ${data.examId}`,
    paperTitle: `试卷标题 ${data.examId}`,
    students: data.timers || []
  }
  
  // 计算当前在线人数（活跃状态的学生数量）
  examData.onlineCount = examData.students.filter(student => student.isActive).length
  
  // 更新或添加考试数据
  const index = realTimeExamData.value.findIndex(item => item.examId === examData.examId)
  if (index >= 0) {
    realTimeExamData.value[index] = examData
  } else {
    realTimeExamData.value.push(examData)
  }
}

// 更新学生开始考试
const updateStudentStart = (data) => {
  const index = realTimeExamData.value.findIndex(item => item.examId === data.examId)
  if (index >= 0) {
    // 添加新学生到学生列表
    const newStudent = {
      studentId: data.studentId,
      studentName: `学生${data.studentId}`,
      className: '未知班级',
      timeUsed: 0,
      startTime: data.startTime,
      isActive: true
    }
    realTimeExamData.value[index].students.push(newStudent)
    // 重新计算在线人数
    realTimeExamData.value[index].onlineCount = realTimeExamData.value[index].students.filter(s => s.isActive).length
  } else {
    // 如果没有该考试，创建一个新的
    realTimeExamData.value.push({
      examId: data.examId,
      examTitle: `考试场次 ${data.examId}`,
      paperTitle: `试卷标题 ${data.examId}`,
      students: [{
        studentId: data.studentId,
        studentName: `学生${data.studentId}`,
        className: '未知班级',
        timeUsed: 0,
        startTime: data.startTime,
        isActive: true
      }],
      onlineCount: 1
    })
  }
}

// 更新学生结束考试
const updateStudentEnd = (data) => {
  const index = realTimeExamData.value.findIndex(item => item.examId === data.examId)
  if (index >= 0) {
    // 查找对应的学生并更新状态
    const studentIndex = realTimeExamData.value[index].students.findIndex(s => s.studentId === data.studentId)
    if (studentIndex >= 0) {
      realTimeExamData.value[index].students[studentIndex].isActive = false
      realTimeExamData.value[index].students[studentIndex].timeUsed = data.timeUsed
    }
    // 重新计算在线人数
    realTimeExamData.value[index].onlineCount = realTimeExamData.value[index].students.filter(s => s.isActive).length
  }
}

// 更新学生时间信息
const updateStudentTime = (data) => {
  const index = realTimeExamData.value.findIndex(item => item.examId === data.examId)
  if (index >= 0) {
    const studentIndex = realTimeExamData.value[index].students.findIndex(s => s.studentId === data.studentId)
    if (studentIndex >= 0) {
      realTimeExamData.value[index].students[studentIndex].timeUsed = data.timeUsed
    }
  }
}

// 刷新实时数据
const refreshRealTimeData = () => {
  requestCurrentExamStatus()
  ElMessage.success('实时数据刷新请求已发送')
}

// 实时时间显示功能（保留原有功能）
const realTimeData = ref([])
const isWebSocketConnected = ref(false)
const ws = ref(null)
const retryCount = ref(0)

const initRealTimeDisplay = () => {
  const wsUrl = `ws://localhost:8080/api/exam-timer`
  
  try {
    ws.value = new WebSocket(wsUrl)

    ws.value.onopen = () => {
      console.log('实时时间WebSocket连接已建立')
      isWebSocketConnected.value = true
      retryCount.value = 0 // 重置重试次数
    }

    ws.value.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        console.log('收到实时时间数据:', data)
        updateRealTimeData(data)
      } catch (error) {
        console.error('解析WebSocket消息失败:', error)
      }
    }

    ws.value.onerror = (error) => {
      console.error('实时时间WebSocket错误:', error)
      isWebSocketConnected.value = false
    }

    ws.value.onclose = (event) => {
      console.log('实时时间WebSocket连接已关闭:', event.code, event.reason)
      isWebSocketConnected.value = false
      
      // 如果是意外断开且未达到最大重试次数，则尝试重连
      if (event.code !== 1000 && retryCount.value < websocketConfig.maxRetries) {
        retryCount.value++
        console.log(`WebSocket连接断开，${websocketConfig.retryInterval / 1000}秒后尝试重连(${retryCount.value}/${websocketConfig.maxRetries})`)
        
        setTimeout(() => {
          initRealTimeDisplay()
        }, websocketConfig.retryInterval)
      }
    }
  } catch (error) {
    console.error('创建WebSocket连接失败:', error)
    ElMessage.error('实时数据连接失败')
  }
}

const updateRealTimeData = (data) => {
  if (!data || !data.type) return
  
  if (data.type === 'update') {
    // 检查数据完整性
    if (!data.examId || !data.studentId || typeof data.timeUsed !== 'number') {
      console.warn('实时数据不完整:', data)
      return
    }
    
    // 更新表格中的实时数据
    const resultIndex = examResults.value.findIndex(item => 
      item.examId === data.examId && item.studentId === data.studentId
    )
    
    if (resultIndex >= 0) {
      examResults.value[resultIndex].timeUsed = data.timeUsed
    }
    
    // 安全地更新实时数据数组
    const index = realTimeData.value.findIndex(item => 
      item.examId === data.examId && item.studentId === data.studentId
    )
    
    if (index >= 0) {
      realTimeData.value[index].timeUsed = data.timeUsed
    } else {
      realTimeData.value.push({
        examId: data.examId,
        studentId: data.studentId,
        timeUsed: data.timeUsed,
        startTime: data.startTime
      })
    }
  }
}

// 获取数据
const fetchData = async () => {
  loading.value = true
  try {
    // 验证分页参数
    validateInputs()
    
    // 获取统计信息
    const statsResponse = await teacherAPI.getAnalysis()
    if (!statsResponse || typeof statsResponse !== 'object') {
      throw new Error('统计信息响应格式错误')
    }
    stats.value = statsResponse

    // 获取考试列表
    const examsResponse = await teacherAPI.getExams()
    exams.value = examsResponse.data || []

    // 获取班级列表
    const classesResponse = await teacherAPI.getClasses()
    classes.value = classesResponse.classes || []

    // 获取考试结果分析
    const resultsResponse = await teacherAPI.getExamResultsAnalysis({
      page: currentPage.value,
      pageSize: pageSize.value,
      examId: filterExam.value,
      classId: filterClass.value,
      scoreFilter: filterScore.value,
      keyword: searchKeyword.value
    })
    
    if (!resultsResponse || typeof resultsResponse !== 'object') {
      throw new Error('考试结果响应格式错误')
    }
    
    examResults.value = resultsResponse.data || []
    total.value = resultsResponse.pagination?.total || 0

    // 更新图表
    updateCharts()
  } catch (error) {
    console.error('获取数据失败:', error)
    ElMessage.error('获取数据失败，请检查网络连接')
  } finally {
    loading.value = false
  }
}

// 验证输入参数
const validateInputs = () => {
  if (currentPage.value <= 0) {
    currentPage.value = 1
  }
  if (pageSize.value <= 0) {
    pageSize.value = 10
  }
}

// 更新图表数据
const updateCharts = () => {
  // 更新成绩分布图
  updateScoreDistributionChart()
  
  // 更新班级对比图
  updateClassComparisonChart()
}

// 更新成绩分布图
const updateScoreDistributionChart = async () => {
  try {
    const response = await teacherAPI.getScoreDistribution('all', scoreRange.value)
    const distributionChart = echarts.init(scoreDistributionChart.value)
    
    const option = {
      tooltip: {
        trigger: 'axis'
      },
      xAxis: {
        type: 'category',
        data: response.distribution.map(item => item.range)
      },
      yAxis: {
        type: 'value',
        name: '人数'
      },
      series: [{
        name: '人数',
        type: 'bar',
        data: response.distribution.map(item => item.count),
        itemStyle: {
          color: '#409eff'
        }
      }]
    }
    
    distributionChart.setOption(option)
  } catch (error) {
    console.error('获取成绩分布数据失败:', error)
  }
}

// 更新班级对比图
const updateClassComparisonChart = async () => {
  try {
    const response = await teacherAPI.getClassComparison('all', comparisonMetric.value)
    const comparisonChart = echarts.init(classComparisonChart.value)
    
    const option = {
      tooltip: {
        trigger: 'axis'
      },
      legend: {
        data: [comparisonMetric.value === 'avg' ? '平均分' : comparisonMetric.value === 'pass' ? '及格率' : '优秀率']
      },
      xAxis: {
        type: 'category',
        data: response.comparisons.map(item => item.className)
      },
      yAxis: {
        type: 'value',
        name: comparisonMetric.value === 'avg' ? '分数' : '百分比',
        min: 0,
        max: comparisonMetric.value === 'avg' ? 100 : 100
      },
      series: [{
        name: comparisonMetric.value === 'avg' ? '平均分' : comparisonMetric.value === 'pass' ? '及格率' : '优秀率',
        type: comparisonMetric.value === 'avg' ? 'bar' : 'line',
        data: response.comparisons.map(item => item.value),
        itemStyle: {
          color: comparisonMetric.value === 'avg' ? '#409eff' : '#67c23a'
        }
      }]
    }
    
    comparisonChart.setOption(option)
  } catch (error) {
    console.error('获取班级对比数据失败:', error)
  }
}

// 监听筛选条件变化
const watchFilters = () => {
  // 监听分页变化
  watch(currentPage, fetchData)
  watch(pageSize, fetchData)
  
  // 监听筛选条件变化
  watch([filterExam, filterClass, filterScore, searchKeyword], fetchData)
  
  // 监听图表参数变化
  watch(scoreRange, updateScoreDistributionChart)
  watch(comparisonMetric, updateClassComparisonChart)
}

// 添加监听器
import { watch } from 'vue'
watchFilters()

// 刷新数据功能
const refreshData = async () => {
  await fetchData()
  ElMessage.success('数据刷新成功')
}

// 处理筛选条件变化
const handleFilterChange = () => {
  currentPage.value = 1 // 重置到第一页
  fetchData()
}

// 处理搜索输入
const handleSearch = () => {
  // 使用防抖处理搜索
  if (debouncedSearch.value) {
    clearTimeout(debouncedSearch.value)
  }
  debouncedSearch.value = setTimeout(() => {
    currentPage.value = 1 // 重置到第一页
    fetchData()
  }, 500)
}

// 初始化
onMounted(() => {
  fetchData()
  // 初始化实时时间显示
  initRealTimeDisplay()
})

onUnmounted(() => {
  if (ws.value) {
    ws.value.close()
  }
})
</script>

<style scoped>
.results-analysis-page {
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

.stat-card {
  text-align: center;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
  transition: transform 0.3s ease;
  border: 1px solid #e4e7ed;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.stat-number {
  font-size: 32px;
  font-weight: bold;
  color: #409eff;
  margin-bottom: 8px;
  transition: color 0.3s ease;
}

.stat-card:hover .stat-number {
  color: #337ecc;
}

.stat-label {
  color: #666;
  font-size: 14px;
}

.chart-container {
  padding: 15px;
  background: white;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  transition: box-shadow 0.3s ease;
}

.chart-container:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.chart-container h4 {
  margin: 0 0 15px 0;
  text-align: center;
  color: #303133;
  font-size: 16px;
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
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

.exam-detail {
  max-height: 60vh;
  overflow-y: auto;
}

.question-analysis h4 {
  margin: 20px 0 15px 0;
  color: #333;
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

/* 响应式设计 */
@media (max-width: 768px) {
  .stat-card {
    margin-bottom: 15px;
  }
  
  .chart-header {
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
