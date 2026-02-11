<template>
  <div class="class-management-page">
    <div class="page-header">
      <h2>班级管理</h2>
    </div>
    
    <div class="page-content">
      <el-card class="card-container">
        <template #header>
          <div class="card-header">
            <h3>班级信息</h3>
            <div class="header-actions">
              <el-button type="primary" @click="showImportDialog = true">
                批量导入
              </el-button>
              <el-button type="primary" @click="showClassDialog = true">
                新建班级
              </el-button>
            </div>
          </div>
        </template>
        
        <el-row :gutter="20">
          <el-col :span="6" v-for="classItem in classes" :key="classItem.id">
            <el-card class="class-card" shadow="hover">
              <template #header>
                <div class="class-header">
                  <h4>{{ classItem.name }}</h4>
                  <el-dropdown @command="handleClassCommand($event, classItem)">
                    <el-icon><MoreFilled /></el-icon>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item command="edit">编辑</el-dropdown-item>
                        <el-dropdown-item command="students">查看学生</el-dropdown-item>
                        <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
                </div>
              </template>
              
              <div class="class-info">
                <div class="info-item">
                  <span class="label">专业：</span>
                  <span>{{ classItem.major }}</span>
                </div>
                <div class="info-item">
                  <span class="label">学生人数：</span>
                  <el-tag type="primary">{{ classItem.studentCount }}人</el-tag>
                </div>
                <div class="info-item">
                  <span class="label">创建时间：</span>
                  <span>{{ classItem.createTime }}</span>
                </div>
                <div class="info-item">
                  <span class="label">班主任：</span>
                  <span>{{ classItem.teacher }}</span>
                </div>
                <div class="info-item" v-if="classItem.statistics">
                  <span class="label">平均分：</span>
                  <span>{{ classItem.statistics.avgScore }}</span>
                </div>
                <div class="info-item" v-if="classItem.statistics">
                  <span class="label">及格率：</span>
                  <span>{{ classItem.statistics.passRate }}%</span>
                </div>
                <div class="info-item" v-if="classItem.statistics">
                  <span class="label">优秀率：</span>
                  <span>{{ classItem.statistics.excellentRate }}%</span>
                </div>
              </div>
            </el-card>
          </el-col>
        </el-row>
        
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

    <!-- 班级对话框（新建/编辑） -->
    <el-dialog 
      v-model="showClassDialog" 
      :title="classFormData.id === 0 ? '新建班级' : '编辑班级'" 
      width="500px"
      @open="initClassFormData"
    >
      <el-form :model="classFormData" :rules="classRules" ref="classFormRef" label-width="80px">
        <el-form-item label="班级名称" prop="name">
          <el-input v-model="classFormData.name" placeholder="请输入班级名称" />
        </el-form-item>
        <el-form-item label="专业" prop="major">
          <el-input v-model="classFormData.major" placeholder="请输入专业名称（可选）" />
        </el-form-item>
        <el-form-item label="班主任" prop="teacher">
          <el-select v-model="classFormData.teacherId" placeholder="请选择班主任（可选）" style="width: 100%">
            <el-option label="暂无班主任" :value="0" />
            <el-option 
              v-for="teacher in teachers" 
              :key="teacher.id" 
              :label="teacher.name" 
              :value="teacher.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="入学年份" prop="enrollmentYear">
          <el-select v-model="classFormData.enrollmentYear" placeholder="请选择入学年份（可选）" style="width: 100%">
            <el-option 
              v-for="year in enrollmentYears" 
              :key="year" 
              :label="year + '级'" 
              :value="year"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="classFormData.description" type="textarea" :rows="3" placeholder="请输入班级描述" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showClassDialog = false">取消</el-button>
        <el-button type="primary" @click="saveClass">确定</el-button>
      </template>
    </el-dialog>

    <!-- 查看班级学生对话框 -->
    <el-dialog v-model="showStudentsDialog" :title="`${currentClass?.name} - 学生列表`" width="800px">
      <el-table :data="classStudents" style="width: 100%">
        <el-table-column prop="studentId" label="学号" width="120" />
        <el-table-column prop="name" label="姓名" width="120" />
        <el-table-column prop="major" label="专业" width="150" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'active' ? 'success' : 'info'">
              {{ scope.row.status === 'active' ? '激活' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="180" />
      </el-table>
    </el-dialog>

    <!-- 批量导入对话框 -->
    <el-dialog v-model="showImportDialog" title="批量导入班级" width="600px">
      <div class="import-dialog-content">
        <p class="import-tip">支持上传Excel(.xlsx/.xls)或CSV文件，文件大小不超过10MB</p>
        
        <el-upload
          class="upload-demo"
          drag
          :show-file-list="false"
          :before-upload="handleImport"
          accept=".xlsx,.xls,.csv"
          :auto-upload="false"
        >
          <el-icon class="el-icon--upload"><upload-filled /></el-icon>
          <div class="el-upload__text">
            将文件拖到此处，或<em>点击上传</em>
          </div>
          <template #tip>
            <div class="el-upload__tip">
              <el-link type="primary" @click="downloadTemplate">下载模板文件</el-link>
            </div>
          </template>
        </el-upload>
      </div>
      
      <template #footer>
        <el-button @click="showImportDialog = false">取消</el-button>
        <el-button type="primary" :loading="importLoading" @click="handleImportManually">
          {{ importLoading ? '导入中...' : '开始导入' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { MoreFilled, UploadFilled } from '@element-plus/icons-vue'
import { teacherAPI } from '@/services/api'
import { formatDateTime } from '@/utils/dateUtils'
import * as XLSX from 'xlsx'


const showClassDialog = ref(false)
const showStudentsDialog = ref(false)
const classFormRef = ref()
const currentClass = ref(null)

const classFormData = reactive({
  id: 0,
  name: '',
  major: '',
  teacherId: 0,
  enrollmentYear: 0,
  description: ''
})

const classRules = {
  name: [{ required: true, message: '请输入班级名称', trigger: 'blur' }],
  major: [{ required: false, message: '请输入专业名称', trigger: 'blur' }],
  teacherId: [{ required: false, message: '请选择班主任', trigger: 'change' }],
  enrollmentYear: [{ required: true, message: '请选择入学年份', trigger: 'change' }]
}

// 生成入学年份选项（当前年份到前5年）
const currentYear = new Date().getFullYear()
const enrollmentYears = ref([])
for (let i = -1; i <= 5; i++) {
  enrollmentYears.value.push(currentYear - i)
}

const classes = ref([])
const classStatistics = ref([])
const classStudents = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const handleClassCommand = (command, classItem) => {
  switch (command) {
    case 'edit':
      editClass(classItem)
      break
    case 'students':
      viewStudents(classItem)
      break
    case 'delete':
      deleteClass(classItem)
      break
  }
}

// 保存班级（新建或编辑）
const saveClass = async () => {
  if (!classFormRef.value) return
  
  try {
    await classFormRef.value.validate()
    
    if (classFormData.id === 0) {
      // 新建班级
      try {
        const classData = {
          name: classFormData.name,
          description: classFormData.description,
          teacherId: classFormData.teacherId || 0,
          major: classFormData.major || '',
          enrollmentYear: classFormData.enrollmentYear || 0
        }
        
        const response = await teacherAPI.createClass(classData)
        ElMessage.success('班级创建成功')
        showClassDialog.value = false
        
        // 重新加载班级列表
        await loadClasses()
      } catch (error) {
        ElMessage.error('创建失败：' + (error.response?.data?.error || error.message))
      }
    } else {
      // 编辑班级
      try {
        const updateData = {
          name: classFormData.name,
          major: classFormData.major,
          teacherId: classFormData.teacherId,
          description: classFormData.description
        }
        
        await teacherAPI.updateClass(classFormData.id, updateData)
        ElMessage.success('班级更新成功')
        showClassDialog.value = false
        
        // 重新加载班级列表
        await loadClasses()
      } catch (error) {
        ElMessage.error('更新失败：' + (error.response?.data?.error || error.message))
      }
    }
  } catch (error) {
    ElMessage.error('请完善表单信息')
  }
}

const teachers = ref([])
const importLoading = ref(false)
const showImportDialog = ref(false)

// 下载模板文件
const downloadTemplate = () => {
  try {
    // 创建模板数据
    const templateData = [
      ['班级名称', '专业', '班主任ID', '入学年份', '描述'],
      ['计算机科学与技术1班', '计算机科学与技术', '', '2024', '计算机科学与技术专业一年级班级'],
    ]
    
    // 创建工作簿
    const workbook = XLSX.utils.book_new()
    const worksheet = XLSX.utils.aoa_to_sheet(templateData)
    
    // 设置列宽
    worksheet['!cols'] = [
      { width: 25 }, // 班级名称
      { width: 20 }, // 专业
      { width: 15 }, // 班主任ID
      { width: 12 }, // 入学年份
      { width: 40 }  // 描述
    ]
    
    // 添加工作表到工作簿
    XLSX.utils.book_append_sheet(workbook, worksheet, '班级模板')
    
    // 生成Excel文件
    const excelBuffer = XLSX.write(workbook, { bookType: 'xlsx', type: 'array' })
    
    // 创建下载链接
    const blob = new Blob([excelBuffer], { type: 'application/octet-stream' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = '班级导入模板.xlsx'
    
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    
    // 清理URL对象
    URL.revokeObjectURL(url)
    
    ElMessage.success('模板文件下载成功')
  } catch (error) {
    ElMessage.error('模板文件生成失败：' + error.message)
  }
}

// 批量导入班级
const handleImport = async (file) => {
  try {
    importLoading.value = true
    
    // 检查文件类型
    const isExcel = file.type === 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' || 
                   file.type === 'application/vnd.ms-excel' ||
                   file.name.endsWith('.xlsx') || 
                   file.name.endsWith('.xls')
    const isCSV = file.type === 'text/csv' || file.name.endsWith('.csv')
    
    if (!isExcel && !isCSV) {
      ElMessage.error('请上传Excel(.xlsx/.xls)或CSV文件')
      return false
    }
    
    // 读取文件
    const data = await readFile(file)
    const classesData = parseFileData(data, isCSV)
    
    if (classesData.length === 0) {
      ElMessage.warning('文件中没有找到有效的班级数据')
      return false
    }
    
    // 验证数据
    const validationErrors = validateImportData(classesData)
    if (validationErrors.length > 0) {
      ElMessage.error(`数据验证失败：${validationErrors.join('；')}`)
      return false
    }
    
    // 调用批量导入API
    const response = await teacherAPI.importClasses(classesData)
    ElMessage.success(`成功导入 ${response.successCount} 个班级，失败 ${response.failCount} 个`)
    
    // 重新加载班级列表
    await loadClasses()
    
    return false // 阻止自动上传
  } catch (error) {
    ElMessage.error('导入失败：' + (error.response?.data?.error || error.message))
    return false
  } finally {
    importLoading.value = false
  }
}

// 手动导入功能
const handleImportManually = async () => {
  // 触发文件选择
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.xlsx,.xls,.csv'
  
  input.onchange = async (e) => {
    const file = e.target.files[0]
    if (file) {
      await handleImport(file)
    }
  }
  
  input.click()
}

// 读取文件内容
const readFile = (file) => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = (e) => {
      resolve(e.target.result)
    }
    reader.onerror = reject
    reader.readAsArrayBuffer(file)
  })
}

// 解析文件数据
const parseFileData = (data, isCSV) => {
  try {
    let workbook
    if (isCSV) {
      // 处理CSV文件
      const text = new TextDecoder().decode(data)
      const lines = text.split('\n').filter(line => line.trim())
      if (lines.length < 2) return []
      
      const headers = lines[0].split(',').map(h => h.trim())
      const rows = lines.slice(1).map(line => {
        const values = line.split(',').map(v => v.trim())
        const row = {}
        headers.forEach((header, index) => {
          row[header.toLowerCase()] = values[index] || ''
        })
        return row
      })
      workbook = { Sheets: { Sheet1: rows } }
    } else {
      // 处理Excel文件
      workbook = XLSX.read(data, { type: 'array' })
    }
    
    const sheetName = workbook.SheetNames[0]
    const worksheet = workbook.Sheets[sheetName]
    
    // 将工作表转换为JSON
    const jsonData = XLSX.utils.sheet_to_json(worksheet, { header: 1 })
    
    if (jsonData.length < 2) return []
    
    // 获取表头
    const headers = jsonData[0].map(h => String(h || '').toLowerCase().trim())
    
    // 转换数据为对象数组
    const result = []
    for (let i = 1; i < jsonData.length; i++) {
      const row = {}
      headers.forEach((header, index) => {
        row[header] = String(jsonData[i][index] || '').trim()
      })
      result.push(row)
    }
    
    return result
  } catch (error) {
    throw new Error('文件解析失败：' + error.message)
  }
}

// 验证导入数据
const validateImportData = (data) => {
  const errors = []
  const requiredFields = ['name']
  
  data.forEach((item, index) => {
    // 检查必填字段
    requiredFields.forEach(field => {
      if (!item[field] || item[field].trim() === '') {
        errors.push(`第 ${index + 1} 行：${field} 字段不能为空`)
      }
    })
    
    // 检查字段长度
    if (item.name && item.name.length > 100) {
      errors.push(`第 ${index + 1} 行：班级名称过长`)
    }
    
    if (item.major && item.major.length > 100) {
      errors.push(`第 ${index + 1} 行：专业名称过长`)
    }
    
    if (item.description && item.description.length > 500) {
      errors.push(`第 ${index + 1} 行：描述过长`)
    }
  })
  
  return errors
}

// 编辑班级
const editClass = async (classItem) => {
  try {
    // 加载教师列表
    if (teachers.value.length === 0) {
      const response = await teacherAPI.getTeachers()
      teachers.value = response
    }
    
    // 填充编辑表单数据
    classFormData.id = classItem.id
    classFormData.name = classItem.name
    classFormData.major = classItem.major
    classFormData.teacherId = classItem.teacherId
    classFormData.enrollmentYear = classItem.enrollmentYear || currentYear
    classFormData.description = classItem.description
    
    showClassDialog.value = true
  } catch (error) {
    ElMessage.error('加载编辑数据失败：' + (error.response?.data?.error || error.message))
  }
}

const viewStudents = async (classItem) => {
  try {
    currentClass.value = classItem
    const response = await teacherAPI.getClassStudents(classItem.id)
    classStudents.value = response.students || []
    showStudentsDialog.value = true
  } catch (error) {
    ElMessage.error('加载学生列表失败：' + (error.response?.data?.error || error.message))
  }
}

const deleteClass = async (classItem) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除班级 "${classItem.name}" 吗？此操作不可恢复。`,
      '删除班级',
      {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      }
    )
    
    try {
      await teacherAPI.deleteClass(classItem.id)
      ElMessage.success('班级删除成功')
      // 重新加载班级列表
      await loadClasses()
    } catch (error) {
      ElMessage.error('删除失败：' + (error.response?.data?.error || error.message))
    }
    
  } catch {
    // 用户取消操作
  }
}

const viewClassDetail = (classItem) => {
  ElMessage.info(`查看班级 ${classItem.className} 的详细分析`)
}

const resetClassForm = () => {
  Object.keys(newClass).forEach(key => {
    newClass[key] = ''
  })
}

// 加载班级列表
const loadClasses = async () => {
  try {
    const response = await teacherAPI.getClasses({
      page: currentPage.value,
      pageSize: pageSize.value
    })
    
    // 获取教师列表
    const teachers = await teacherAPI.getTeachers()
    
    // 获取班级统计信息
    const statistics = await teacherAPI.getClassStatistics()
    
    classes.value = response.classes.map(cls => {
      // 查找对应教师
      const teacher = teachers.data.find(t => t.id === cls.teacherId) || {}
      
      // 查找对应统计信息
      const stats = statistics.find(stat => stat.className === cls.name)
      
      return {
        id: cls.id,
        name: cls.name,
        description: cls.description,
        major: cls.major || cls.description, // 兼容旧数据
        teacher: teacher.name || '未知教师',
        teacherId: cls.teacherId,
        studentCount: cls.studentCount,
        createTime: formatDateTime(cls.createdAt),
        statistics: stats ? {
          avgScore: stats.avgScore,
          passRate: stats.passRate,
          excellentRate: stats.excellentRate
        } : null
      }
    })
    
    // 设置总数
    total.value = response.total || 0
  } catch (error) {
    ElMessage.error('加载班级列表失败：' + (error.response?.data?.error || error.message))
  }
}

// 初始化班级表单数据
const initClassFormData = async () => {
  try {
    // 获取教师列表用于班主任选择
    const teachersResponse = await teacherAPI.getTeachers()
    teachers.value = teachersResponse.data || []
    
    // 如果是新建班级，设置默认值
    if (classFormData.id === 0) {
      classFormData.name = ''
      classFormData.major = ''
      classFormData.teacherId = 0
      classFormData.enrollmentYear = currentYear
      classFormData.description = ''
    }
  } catch (error) {
    ElMessage.error('初始化班级数据失败：' + (error.response?.data?.error || error.message))
  }
}


// 加载班级统计
const loadClassStatistics = async () => {
  try {
    const response = await teacherAPI.getClassStatistics()
    classStatistics.value = response.map(stat => ({
      className: stat.className,
      studentCount: stat.studentCount,
      avgScore: stat.avgScore,
      passRate: stat.passRate,
      excellentRate: stat.excellentRate
    }))
  } catch (error) {
    ElMessage.error('加载班级统计失败：' + (error.response?.data?.error || error.message))
  }
}

// 监听分页参数变化
watch([currentPage, pageSize], () => {
  loadClasses()
})

// 组件挂载时加载数据
onMounted(async () => {
  await Promise.all([
    loadClasses(),
    loadClassStatistics()
  ])
})
</script>

<style scoped>
.class-management-page {
  min-height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.class-card {
  margin-bottom: 20px;
}

.class-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.class-info {
  padding: 10px 0;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-size: 14px;
}

.info-item .label {
  color: #666;
  min-width: 70px;
}

.import-dialog-content {
  padding: 20px 0;
}

.import-tip {
  color: #606266;
  font-size: 14px;
  margin-bottom: 20px;
}

.el-icon--upload {
  font-size: 67px;
  color: #c0c4cc;
  margin-bottom: 16px;
}

.el-upload__text {
  color: #606266;
  font-size: 14px;
  text-align: center;
}

.el-upload__text em {
  color: #409eff;
  font-style: normal;
  font-weight: 500;
}

.el-upload__tip {
  text-align: center;
  margin-top: 10px;
}
</style>
