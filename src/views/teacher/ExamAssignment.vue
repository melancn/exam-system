<template>
  <div class="exam-assignment-page">
    <div class="page-header">
      <h2>试卷分配</h2>
    </div>
    
    <div class="page-content">
      <el-card class="card-container">
        <template #header>
          <h3>分配试卷</h3>
        </template>
        
        <el-form :model="assignment" :rules="assignmentRules" ref="assignmentFormRef" label-width="100px">
          <el-row :gutter="20">
            <el-col :span="10">
              <el-form-item label="选择试卷" prop="examId">
                <el-select v-model="assignment.examId" placeholder="请选择试卷" style="width: 100%">
                  <el-option 
                    v-for="exam in availableExams" 
                    :key="exam.id" 
                    :label="exam.title" 
                    :value="exam.id"
                  />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="14">
              <el-form-item label="选择班级" prop="classIds">
                <el-select 
                  v-model="assignment.classIds" 
                  multiple 
                  placeholder="请选择分配到的班级" 
                  style="width: 100%"
                >
                  <el-option 
                    v-for="classItem in classes" 
                    :key="classItem.id" 
                    :label="classItem.name" 
                    :value="classItem.id"
                  />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>
          
          <el-row :gutter="20">
            <el-col :span="6">
              <el-form-item label="考试时长" prop="duration">
                <el-input-number 
                  v-model="assignment.duration" 
                  :min="10" 
                  :max="300" 
                  placeholder="请输入考试时长（分钟）"
                  style="width: 80%"
                />
                <span style="margin-left: 10px; color: #666;">分钟</span>
              </el-form-item>
            </el-col>
            <el-col :span="6">
              <el-form-item label="及格分数" prop="passScore">
                <el-input-number 
                  v-model="assignment.passScore" 
                  :min="0" 
                  :max="100" 
                  placeholder="请输入及格分数"
                  style="width: 80%"
                />
                <span style="margin-left: 10px; color: #666;">分</span>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="考试时间" prop="timeRange">
                <el-date-picker
                  v-model="assignment.timeRange"
                  type="datetimerange"
                  range-separator="至"
                  start-placeholder="开始时间"
                  end-placeholder="结束时间"
                  style="width: 100%"
                />
              </el-form-item>
            </el-col>
          </el-row>
          
          <el-form-item label="考试说明" prop="description">
            <el-input 
              v-model="assignment.description" 
              type="textarea" 
              :rows="3" 
              placeholder="请输入考试说明"
            />
          </el-form-item>
          
          <el-form-item>
            <el-button type="primary" @click="assignExam">分配试卷</el-button>
            <el-button @click="resetForm">重置</el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <el-card class="card-container" style="margin-top: 20px;">
        <template #header>
          <h3>已分配试卷</h3>
        </template>
        
        <el-table :data="assignments" style="width: 100%">
          <el-table-column prop="examTitle" label="试卷名称" width="200" />
          <el-table-column prop="className" label="分配班级" width="200">
            <template #default="scope">
              <el-tag 
                :key="scope.row.className"
                style="margin-right: 5px; margin-bottom: 5px;"
              >
                {{ scope.row.className }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="startTime" label="开始时间" width="180">
            <template #default="scope">
              {{ formatDateTime(scope.row.startTime) }}
            </template>
          </el-table-column>
          <el-table-column prop="endTime" label="结束时间" width="180">
            <template #default="scope">
              {{ formatDateTime(scope.row.endTime) }}
            </template>
          </el-table-column>
          <el-table-column prop="duration" label="时长" width="100">
            <template #default="scope">
              {{ scope.row.duration }}分钟
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="scope">
              <el-tag :type="getAssignmentStatusType(scope.row)">
                {{ getAssignmentStatus(scope.row) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="参与人数" width="100">
            <template #default="scope">
              {{ scope.row.participants }}/{{ scope.row.totalStudents }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="scope">
              <el-button size="small" @click="editAssignment(scope.row)">编辑</el-button>
              <el-button 
                size="small" 
                type="danger" 
                @click="deleteAssignment(scope.row)"
              >
                取消分配
              </el-button>
            </template>
          </el-table-column>
        </el-table>
        
        <!-- 分页组件 -->
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[5, 10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadAssignments"
          @current-change="loadAssignments"
          style="margin-top: 20px; justify-content: flex-end;"
        />
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { teacherAPI } from '@/services/api'
import { formatDateTime } from '@/utils/dateUtils'

const assignmentFormRef = ref()

const assignment = reactive({
  examId: '',
  classIds: [],
  timeRange: [],
  startTime: '',
  endTime: '',
  duration: 60,
  passScore: 60,
  description: ''
})

const assignmentRules = {
  examId: [{ required: true, message: '请选择试卷', trigger: 'change' }],
  classIds: [{ required: true, message: '请选择分配班级', trigger: 'change' }],
  timeRange: [{ required: true, message: '请选择考试时间范围', trigger: 'change' }],
  duration: [{ required: true, message: '请输入考试时长', trigger: 'blur' }]
}

const availableExams = ref([])
const classes = ref([])
const assignments = ref([])

const getAssignmentStatus = (assignment) => {
  const now = new Date()
  const startTime = new Date(assignment.startTime)
  const endTime = new Date(assignment.endTime)
  
  if (now < startTime) return '未开始'
  if (now >= startTime && now <= endTime) return '进行中'
  return '已结束'
}

const getAssignmentStatusType = (assignment) => {
  const status = getAssignmentStatus(assignment)
  const types = {
    '未开始': 'info',
    '进行中': 'success',
    '已结束': 'danger'
  }
  return types[status] || 'info'
}

const assignExam = async () => {
  if (!assignmentFormRef.value) return
  
  try {
    await assignmentFormRef.value.validate()
    
    // 从时间范围提取开始和结束时间
    const [startTime, endTime] = assignment.timeRange
    
    try {
      // API请求数据
      const assignmentData = {
        examId: assignment.examId,
        classIds: assignment.classIds, // 使用所有选中的班级ID
        startTime: startTime,
        endTime: endTime,
        duration: assignment.duration,
        passScore: assignment.passScore,
        description: assignment.description
      }
      
      await teacherAPI.createAssignment(assignmentData)
      ElMessage.success('试卷分配成功')
      
      // 重新加载分配列表
      loadAssignments()
      resetForm()
    } catch (error) {
      ElMessage.error('分配失败：' + (error.response?.data?.error || error.message))
    }
    
  } catch (error) {
    ElMessage.error('请完善表单信息')
  }
}

const resetForm = () => {
  if (assignmentFormRef.value) {
    assignmentFormRef.value.resetFields()
    assignment.duration = 60
    assignment.passScore = 60
  }
}

const editAssignment = async (assignment) => {
  try {
    // 加载分配详情并显示编辑对话框（这里简化处理）
    ElMessage.info(`编辑试卷分配：${assignment.examTitle}`)
  } catch (error) {
    ElMessage.error('加载分配详情失败：' + (error.response?.data?.error || error.message))
  }
}

const deleteAssignment = async (assignment) => {
  try {
    await ElMessageBox.confirm(
      `确定要取消试卷 "${assignment.examTitle}" 的分配吗？`,
      '取消分配',
      {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      }
    )
    
    try {
      await teacherAPI.deleteAssignment(assignment.id)
      // 重新加载分配列表
      await loadAssignments()
      ElMessage.success('分配已取消')
    } catch (error) {
      ElMessage.error('取消分配失败：' + (error.response?.data?.error || error.message))
    }
    
  } catch {
    // 用户取消操作
  }
}

// 加载试卷列表
const loadExams = async () => {
  try {
    const response = await teacherAPI.getExams()
    // 检查响应数据是否存在
    if (!response) {
      throw new Error('无效的响应数据')
    }
    
    // 处理分页响应数据
    const examsData = response.data || response || []
    
    availableExams.value = Array.isArray(examsData) 
      ? examsData.map(exam => ({
          id: exam.id || exam.ID,
          title: exam.title || exam.Title,
          duration: exam.duration || exam.Duration,
          totalScore: exam.totalScore || exam.TotalScore,
          questionCount: exam.questionCount || 0
        }))
      : []
  } catch (error) {
    ElMessage.error('加载试卷列表失败：' + (error.response?.data?.error || error.message))
  }
}

// 加载班级列表
const loadClasses = async () => {
  try {
    const response = await teacherAPI.getClasses()
    // 检查响应数据是否存在
    if (!response) {
      throw new Error('无效的响应数据')
    }
    
    // 处理响应数据
    const classesData = response.classes || response || []
    
    classes.value = Array.isArray(classesData) 
      ? classesData.map(cls => ({
          id: cls.id || cls.ID,
          name: cls.name || cls.Name,
          studentCount: cls.studentCount || cls.StudentCount || 0
        }))
      : []
  } catch (error) {
    ElMessage.error('加载班级列表失败：' + (error.response?.data?.error || error.message))
  }
}

// 分页参数
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// 加载分配列表
const loadAssignments = async () => {
  try {
    const response = await teacherAPI.getAssignments({
      page: pagination.page,
      pageSize: pagination.pageSize
    })
    // 检查响应数据是否存在
    if (!response) {
      throw new Error('无效的响应数据')
    }
    
    // 处理响应数据结构
    const assignmentsData = response.assignments || response || []
    const paginationData = response.pagination || {}
    
    assignments.value = Array.isArray(assignmentsData) 
      ? assignmentsData.map(assignment => ({
          id: assignment.id || assignment.ID,
          examId: assignment.examId || assignment.ExamID,
          examTitle: assignment.examTitle || assignment.ExamTitle,
          classId: assignment.classId || assignment.ClassID, // 保持单个班级ID
          className: assignment.className || assignment.ClassName,
          startTime: assignment.startTime || assignment.StartTime,
          endTime: assignment.endTime || assignment.EndTime,
          duration: assignment.duration || assignment.Duration,
          passScore: assignment.passScore || assignment.PassScore,
          totalStudents: assignment.totalStudents || assignment.TotalStudents || 0,
          participants: assignment.participants || assignment.Participants || 0,
          createTime: assignment.createdAt || assignment.CreatedAt
        }))
      : []
    
    // 更新分页信息
    pagination.total = paginationData.total || response.total || assignments.value.length
  } catch (error) {
    ElMessage.error('加载分配列表失败：' + (error.response?.data?.error || error.message))
  }
}

// 组件挂载时加载数据
onMounted(async () => {
  await Promise.all([
    loadExams(),
    loadClasses(),
    loadAssignments()
  ])
})
</script>

<style scoped>
.exam-assignment-page {
  min-height: 100%;
}
</style>
