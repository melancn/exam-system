<template>
  <div class="student-management-page">
    <div class="page-header">
      <h2>学生账号录入</h2>
    </div>
    
    <div class="page-content">
        <!-- 学生信息对话框（新建/编辑） -->
        <el-dialog v-model="showStudentDialog" :title="studentFormData.id === 0 ? '添加学生' : '编辑学生'" width="600px">
          <el-form :model="studentFormData" :rules="studentRules" ref="studentFormRef" label-width="100px">
            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item label="学生账号" prop="username">
                  <el-input v-model="studentFormData.username" placeholder="请输入学生账号" :disabled="studentFormData.id !== 0" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="学号" prop="studentId">
                  <el-input v-model="studentFormData.studentId" placeholder="请输入学号（可选）" />
                </el-form-item>
              </el-col>
            </el-row>
            
            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item label="姓名" prop="name">
                  <el-input v-model="studentFormData.name" placeholder="请输入姓名" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="联系电话" prop="phone">
                  <el-input v-model="studentFormData.phone" placeholder="请输入联系电话" />
                </el-form-item>
              </el-col>
            </el-row>
            
            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item label="班级" prop="classId">
                  <el-select v-model="studentFormData.classId" placeholder="请选择班级" style="width: 100%">
                    <el-option v-for="classItem in classes" :key="classItem.id" 
                      :label="classItem.name" :value="classItem.id" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="专业" prop="major">
                  <el-input v-model="studentFormData.major" placeholder="请输入专业（可选）" />
                </el-form-item>
              </el-col>
            </el-row>
            
            <el-form-item label="密码" prop="password">
              <el-input v-model="studentFormData.password" placeholder="请输入密码（编辑时留空则不修改）" />
            </el-form-item>
          </el-form>
          
          <template #footer>
            <el-button @click="showStudentDialog = false">取消</el-button>
            <el-button type="primary" @click="handleStudentSubmit">
              {{ studentFormData.id === 0 ? '添加' : '更新' }}
            </el-button>
          </template>
        </el-dialog>

      <el-card class="card-container" style="margin-top: 20px;">
        <template #header>
          <div class="card-header">
            <h3>学生列表</h3>
            <div class="action-buttons">
              <el-button type="primary" @click="openStudentDialog()">
                添加学生
              </el-button>
              <el-button type="primary" @click="showBatchDialog = true">
                批量导入
              </el-button>
              <el-button type="primary" @click="exportStudents">导出列表</el-button>
            </div>
          </div>
        </template>
        
        <div class="table-toolbar">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索学生姓名或学生账号"
            style="width: 380px"
            clearable
            @keyup.enter="handleSearch"
          >
            <template #append>
              <el-button :icon="Search" @click="handleSearch" />
            </template>
          </el-input>
        </div>
        
        <el-table :data="students" style="width: 100%" v-loading="loading">
          <el-table-column prop="username" label="学生账号" width="120" />
          <el-table-column prop="studentId" label="学号" width="120" />
          <el-table-column prop="name" label="姓名" width="120" />
          <el-table-column prop="className" label="班级" width="150" />
          <el-table-column prop="major" label="专业" width="150" />
          <el-table-column prop="createTime" label="创建时间" width="180" />
          <el-table-column label="操作" width="200">
            <template #default="scope">
              <el-button size="small" @click="openStudentDialog(scope.row)">编辑</el-button>
              <el-button size="small" type="danger" @click="deleteStudent(scope.row)">
                删除
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
    </div>


        <!-- 批量导入对话框 -->
        <el-dialog v-model="showBatchDialog" title="批量导入学生" width="600px">
      <el-upload
        class="upload-demo"
        drag
        action="#"
        :auto-upload="false"
        :on-change="handleFileChange"
        accept=".xlsx,.xls,.csv"
      >
        <el-icon class="el-icon--upload"><upload-filled /></el-icon>
        <div class="el-upload__text">将文件拖到此处，或<em>点击上传</em></div>
        <template #tip>
          <div class="el-upload__tip">支持 xlsx/xls/csv 格式，文件大小不超过 5MB</div>
        </template>
      </el-upload>
      
      <template #footer>
        <el-button @click="downloadTemplate">下载模板</el-button>
        <el-button @click="showBatchDialog = false">取消</el-button>
        <el-button type="primary" @click="batchImport">导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { UploadFilled, Search } from '@element-plus/icons-vue'
import { teacherAPI } from '@/services/api'
import * as XLSX from 'xlsx'
import { validators } from '@/utils/validators'

const studentFormRef = ref()
const loading = ref(false)
const showBatchDialog = ref(false)
const showStudentDialog = ref(false)
const searchKeyword = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const studentFormData = reactive({
  id: 0,
  username: '',
  studentId: '',
  name: '',
  classId: 0,
  major: '',
  phone: '',
  password: '123456a'
})

const studentRules = {
  username: [
    { required: true, message: '请输入学生账号', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        const result = validators.validateStudentId(value)
        if (!result.valid) {
          callback(new Error(result.message))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  studentId: [
    {
      validator: (rule, value, callback) => {
        if (value && value.trim().length > 0) {
          const result = validators.validateStudentId(value)
          if (!result.valid) {
            callback(new Error(result.message))
          } else {
            callback()
          }
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  name: [
    { required: true, message: '请输入姓名', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        const result = validators.validateName(value)
        if (!result.valid) {
          callback(new Error(result.message))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  classId: [{ required: true, message: '请选择班级', trigger: 'change' }],
  major: [
    {
      validator: (rule, value, callback) => {
        if (value && value.trim().length > 0) {
          const result = validators.validateMajor(value)
          if (!result.valid) {
            callback(new Error(result.message))
          } else {
            callback()
          }
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  phone: [
    {
      validator: (rule, value, callback) => {
        if (value && value.trim().length > 0) {
          const result = validators.validatePhone(value)
          if (!result.valid) {
            callback(new Error(result.message))
          } else {
            callback()
          }
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  password: [
    {
      validator: (rule, value, callback) => {
        // 编辑模式下密码可以为空（不修改密码）
        if (studentFormData.id !== 0 && (!value || value.trim().length === 0)) {
          callback()
          return
        }
        // 新建模式下密码必填，且需要验证格式
        if (studentFormData.id === 0 && (!value || value.trim().length === 0)) {
          callback(new Error('请输入密码'))
          return
        }
        // 验证密码格式
        const result = validators.validatePassword(value)
        if (!result.valid) {
          callback(new Error(result.message))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

const classes = ref([])

const students = ref([])

const fetchStudents = async () => {
  try {
    loading.value = true
    const response = await teacherAPI.getStudents({
      page: currentPage.value,
      pageSize: pageSize.value,
      keyword: searchKeyword.value
    })
    
    // 检查响应数据是否存在
    if (!response) {
      throw new Error('无效的响应数据')
    }
    
    // 后端返回分页格式: {data: [...], pagination: {...}}
    const studentData = response.data || []
    const totalCount = response.pagination?.total || 0
    
    students.value = studentData.map(student => ({
      ...student,
      id: student.id,
      studentId: student.studentId,
      name: student.name,
      className: student.className || student.class,
      major: student.major,
      createTime: new Date(student.createTime || student.createdAt).toLocaleString(),
      status: 'active'
    }))
    
    total.value = totalCount
  } catch (error) {
    console.error('获取学生列表失败:', error)
    ElMessage.error('获取学生列表失败: ' + (error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

const handleStudentSubmit = async () => {
  if (!studentFormRef.value) return
  
  try {
    await studentFormRef.value.validate()
    
    // 根据选择的classId找到班级名称
    const selectedClass = classes.value.find(cls => cls.id === studentFormData.classId)
    if (!selectedClass) {
      ElMessage.error('请选择有效的班级')
      return
    }
    
    if (studentFormData.id === 0) {
      // 新建学生
      const response = await teacherAPI.createStudent({
        username: studentFormData.username,
        studentId: studentFormData.studentId,
        name: studentFormData.name,
        classId: studentFormData.classId,
        major: studentFormData.major,
        phone: studentFormData.phone,
        password: studentFormData.password
      })
      
      if (response && response.id) {
        ElMessage.success('学生添加成功')
        showStudentDialog.value = false
        resetForm()
        fetchStudents() // 刷新列表
      } else {
        ElMessage.error('添加学生失败')
      }
    } else {
      // 编辑学生
      const updateData = {
        name: studentFormData.name,
        classId: studentFormData.classId,
        major: studentFormData.major
      }
      
      // 如果密码不为空，则更新密码
      if (studentFormData.password && studentFormData.password.trim().length > 0) {
        updateData.password = studentFormData.password
      }
      
      await teacherAPI.updateStudent(studentFormData.id, updateData)
      
      ElMessage.success('学生信息更新成功')
      showStudentDialog.value = false
      fetchStudents() // 刷新列表
    }
  } catch (error) {
    console.error('操作学生失败:', error)
    ElMessage.error('操作学生失败: ' + (error.response?.data?.error || error.message || '未知错误'))
  }
}

const resetForm = () => {
  if (studentFormRef.value) {
    studentFormRef.value.resetFields()
  }
}

const openStudentDialog = (student = null) => {
    if (student) {
      // 编辑模式
      studentFormData.id = student.id
      studentFormData.username = student.username || ''
      studentFormData.studentId = student.studentId || ''
      studentFormData.name = student.name
      studentFormData.classId = student.classId || 0
      studentFormData.major = student.major || ''
      studentFormData.phone = student.phone || ''
      studentFormData.password = '' // 编辑时不显示密码
    } else {
      // 新建模式
      studentFormData.id = 0
      studentFormData.username = ''
      studentFormData.studentId = ''
      studentFormData.name = ''
      studentFormData.classId = classes.value[0]?.id || 0
      studentFormData.major = ''
      studentFormData.phone = ''
      studentFormData.password = '123456a'
    }
  
  showStudentDialog.value = true
}

const deleteStudent = async (student) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除学生 ${student.name} 吗？此操作不可恢复！`,
      '删除学生',
      {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      }
    )
    
    // 调用后端API删除
    await teacherAPI.deleteStudent(student.id)
    ElMessage.success('学生删除成功')
    fetchStudents() // 刷新列表
  } catch {
    // 用户取消操作
  }
}

const selectedFile = ref(null)
const importLoading = ref(false)

const handleFileChange = (file) => {
  selectedFile.value = file.raw
  ElMessage.success(`已选择文件：${file.name}`)
}

const parseExcelFile = (file) => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    
    reader.onload = (e) => {
      try {
        const data = new Uint8Array(e.target.result)
        const workbook = XLSX.read(data, { type: 'array' })
        
        // 获取第一个工作表
        const firstSheetName = workbook.SheetNames[0]
        const worksheet = workbook.Sheets[firstSheetName]
        
        // 转换为JSON格式
        const jsonData = XLSX.utils.sheet_to_json(worksheet, { header: 1 })
        
        // 验证表头
        const headers = jsonData[0]
        const expectedHeaders = ['学生账号', '学号', '姓名', '班级', '专业', '联系电话']
        
        const headerIndex = {}
        expectedHeaders.forEach((header, index) => {
          const foundIndex = headers.findIndex(h => h === header)
          if (foundIndex === -1) {
            throw new Error(`缺少必要列：${header}`)
          }
          headerIndex[header] = foundIndex
        })
        
// 解析数据
const students = []
for (let i = 1; i < jsonData.length; i++) {
  const row = jsonData[i]
  if (!row || row.length === 0) continue
  
  const className = row[headerIndex['班级']]
  const classObj = classes.value.find(cls => cls.name === className)
  
  if (!classObj) {
    throw new Error(`第${i + 1}行班级"${className}"不存在`)
  }
  
  const student = {
    username: row[headerIndex['学生账号']],
    studentId: row[headerIndex['学号']],
    name: row[headerIndex['姓名']],
    className: className,
    major: row[headerIndex['专业']] || '',
    phone: row[headerIndex['联系电话']] || ''
  }
  
  // 验证必要字段
  if (!student.studentId || !student.name || !student.className) {
    throw new Error(`第${i + 1}行数据不完整`)
  }
  
  students.push(student)
}
        
        resolve(students)
      } catch (error) {
        reject(error)
      }
    }
    
    reader.onerror = () => {
      reject(new Error('文件读取失败'))
    }
    
    reader.readAsArrayBuffer(file)
  })
}

const batchImport = async () => {
  if (!selectedFile.value) {
    ElMessage.error('请选择要导入的文件')
    return
  }
  
  try {
    importLoading.value = true
    
    // 解析Excel文件
    const students = await parseExcelFile(selectedFile.value)
    
    if (students.length === 0) {
      ElMessage.warning('文件中没有找到有效数据')
      return
    }
    
    // 转换数据格式以匹配后端期望的User模型格式
    const studentsData = students.map(s => ({
      username: s.username,
      studentId: s.studentId,
      name: s.name,
      classId: s.classId,
      major: s.major,
      phone: s.phone
    }))
    
    // 调用后端API导入
    const response = await teacherAPI.importStudents(studentsData)
    
    ElMessage.success(`成功导入 ${response.count} 名学生`)
    showBatchDialog.value = false
    selectedFile.value = null
    fetchStudents() // 刷新列表
  } catch (error) {
    console.error('导入失败:', error)
    ElMessage.error(error.message || '导入失败，请检查文件格式')
  } finally {
    importLoading.value = false
  }
}

const exportStudents = async () => {
  try {
    const response = await teacherAPI.exportStudents()
    
    if (!response || response.length === 0) {
      ElMessage.warning('没有数据可导出')
      return
    }
    
    // 创建工作表
    const worksheet = XLSX.utils.json_to_sheet(response)
    
    // 创建工作簿
    const workbook = XLSX.utils.book_new()
    XLSX.utils.book_append_sheet(workbook, worksheet, '学生列表')
    
    // 生成文件名
    const timestamp = new Date().toISOString().slice(0, 19).replace(/[:.]/g, '-')
    const filename = `学生列表_${timestamp}.xlsx`
    
    // 导出文件
    XLSX.writeFile(workbook, filename)
    
    ElMessage.success('导出成功')
  } catch (error) {
    console.error('导出失败:', error)
    ElMessage.error('导出失败')
  }
}

// 下载模板文件
const downloadTemplate = () => {
  try {
    // 创建模板数据
    const templateData = [
      ['学生账号', '学号', '姓名', '班级', '专业', '联系电话'],
      ['20240001', '20240001', '张三', '计算机1班', '计算机科学与技术', '13812345678'],
      ['20240002', '20240002', '李四', '计算机1班', '软件工程', '13987654321'],
      ['20240003', '20240003', '王五', '计算机2班', '网络工程', '01012345678']
    ]
    
    // 创建工作表
    const worksheet = XLSX.utils.aoa_to_sheet(templateData)
    
    // 创建工作簿
    const workbook = XLSX.utils.book_new()
    XLSX.utils.book_append_sheet(workbook, worksheet, '学生导入模板')
    
    // 导出文件
    XLSX.writeFile(workbook, '学生导入模板.xlsx')
    
    ElMessage.success('模板下载成功')
  } catch (error) {
    console.error('下载模板失败:', error)
    ElMessage.error('下载模板失败')
  }
}

// 监听分页参数变化
watch([currentPage, pageSize], () => {
  fetchStudents()
})

// 搜索按钮点击事件
const handleSearch = () => {
  currentPage.value = 1 // 重置到第一页
  fetchStudents()
}

// 监听搜索关键词变化（可选：如果需要实时搜索）
// watch(searchKeyword, (newVal) => {
//   currentPage.value = 1 // 重置到第一页
//   fetchStudents()
// })

// 加载班级列表
const loadClasses = async () => {
  try {
    const response = await teacherAPI.getClasses()
    classes.value = response.classes.map(cls => ({
      id: cls.id,
      name: cls.name
    }))
  } catch (error) {
    ElMessage.error('加载班级列表失败：' + (error.response?.data?.error || error.message))
  }
}

// 从API获取数据
onMounted(async () => {
  await Promise.all([
    fetchStudents(),
    loadClasses()
  ])
})
</script>

<style scoped>
.student-management-page {
  min-height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.action-buttons {
  display: flex;
  gap: 10px;
}

.table-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  text-align: right;
}
</style>
