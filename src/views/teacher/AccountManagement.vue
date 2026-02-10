<template>
  <div class="account-management-page">
    <div class="page-header">
      <h2>教师账号管理</h2>
    </div>
    
    <div class="page-content">
      <el-card class="card-container">
        <template #header>
          <div class="card-header">
            <h3>教师账号列表</h3>
            <el-button type="primary" @click="showAddDialog = true">
              添加教师账号
            </el-button>
          </div>
        </template>
        
        <el-table :data="teachers" style="width: 100%" v-loading="loading">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="username" label="账号" width="120" />
          <el-table-column prop="name" label="姓名" width="120" />
          <el-table-column prop="role" label="角色" width="120">
            <template #default="scope">
              <el-tag :type="scope.row.role === 'admin' ? 'danger' : 'primary'">
                {{ scope.row.role === 'admin' ? '管理员' : '普通教师' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="createTime" label="创建时间" width="180" />
          <el-table-column prop="lastLogin" label="最后登录" width="180" />
          <el-table-column label="状态" width="100">
            <template #default="scope">
              <el-switch
                v-model="scope.row.status"
                :active-value="1"
                :inactive-value="0"
                @click="toggleStatus(scope.row)"
              />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="scope">
              <el-button size="small" @click="editTeacher(scope.row)">编辑</el-button>
              <el-button 
                size="small" 
                type="danger" 
                @click="deleteTeacher(scope.row)"
                :disabled="scope.row.role === 'admin'"
              >
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

    <!-- 添加/编辑教师对话框 -->
    <el-dialog v-model="showAddDialog" :title="isEditMode ? '编辑教师账号' : '添加教师账号'" width="500px">
      <el-form :model="teacherForm" :rules="rules" ref="teacherFormRef" label-width="100px">
        <el-form-item label="教师姓名" prop="name">
          <el-input v-model="teacherForm.name" placeholder="请输入教师姓名" />
        </el-form-item>
        <el-form-item label="登录账号" prop="username">
          <el-input v-model="teacherForm.username" placeholder="请输入登录账号" />
        </el-form-item>
        <el-form-item label="登录密码" prop="password">
          <el-input 
            v-model="teacherForm.password" 
            type="password" 
            placeholder="请输入登录密码"
            show-password
          />
        </el-form-item>
        <el-form-item label="账号角色" prop="role">
          <el-radio-group v-model="teacherForm.role">
            <el-radio label="teacher">普通教师</el-radio>
            <el-radio label="admin">管理员</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="submitForm">
          {{ isEditMode ? '保存修改' : '添加账号' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { teacherAPI } from '@/services/api'
import { formatDateTime } from '@/utils/dateUtils'

const showAddDialog = ref(false)
const isEditMode = ref(false)
const teacherFormRef = ref()

const teacherForm = reactive({
  id: '',
  name: '',
  username: '',
  password: '',
  role: 'teacher'
})

const rules = {
  name: [{ required: true, message: '请输入教师姓名', trigger: 'blur' }],
  username: [{ required: true, message: '请输入登录账号', trigger: 'blur' }],
  password: [{ required: true, message: '请输入登录密码', trigger: 'blur' }],
  role: [{ required: true, message: '请选择账号角色', trigger: 'change' }]
}

const teachers = ref([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const fetchTeachers = async () => {
  try {
    loading.value = true
    const response = await teacherAPI.getTeachers({
      page: currentPage.value,
      pageSize: pageSize.value
    })
    
    teachers.value = response.data.map(teacher => ({
      ...teacher,
      id: teacher.id,
      username: teacher.username,
      name: teacher.name,
      role: teacher.role,
      status: teacher.status || 'active',
      createTime: formatDateTime(teacher.createTime || teacher.createdAt),
      lastLogin: teacher.lastLogin || '从未登录'
    }))
    
    total.value = response.pagination.total
  } catch (error) {
    ElMessage.error('获取教师列表失败')
  } finally {
    loading.value = false
  }
}

// 监听分页参数变化
watch([currentPage, pageSize], () => {
  fetchTeachers()
})

onMounted(async () => {
  fetchTeachers()
})

const toggleStatus = async (teacher) => {
  try {
    // 根据后端接口，status字段使用数字：1表示激活，0表示禁用
    const newStatus = teacher.status === 1 ? 0 : 1
    
    await teacherAPI.updateTeacher(teacher.id, {
      name: teacher.name,
      isAdmin: teacher.isAdmin || 0,
      status: newStatus
    })
    
    // 更新本地状态
    teacher.status = newStatus
    ElMessage.success(`账号 ${teacher.username} 状态已${newStatus === 1 ? '激活' : '禁用'}`)
  } catch (error) {
    // 恢复状态
    teacher.status = teacher.status === 1 ? 0 : 1
    ElMessage.error('状态更新失败')
  }
}

const editTeacher = (teacher) => {
  Object.assign(teacherForm, teacher)
  isEditMode.value = true
  showAddDialog.value = true
}

const deleteTeacher = async (teacher) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除教师账号 "${teacher.name}" 吗？此操作不可恢复。`,
      '删除教师账号',
      { type: 'warning' }
    )
    
    await teacherAPI.deleteTeacher(teacher.id)
    teachers.value = teachers.value.filter(t => t.id !== teacher.id)
    ElMessage.success('教师账号删除成功')
  } catch {
    // 用户取消操作
  }
}

const submitForm = async () => {
  if (!teacherFormRef.value) return
  
    try {
      await teacherFormRef.value.validate()
      
      if (isEditMode.value) {
        // 更新现有教师
        const updateData = {
          name: teacherForm.name,
          isAdmin: teacherForm.role === 'admin' ? 1 : 0,
          status: teacherForm.status
        }
        await teacherAPI.updateTeacher(teacherForm.id, updateData)
        const index = teachers.value.findIndex(t => t.id === teacherForm.id)
        if (index !== -1) {
          teachers.value[index] = { ...teacherForm }
        }
        ElMessage.success('教师账号更新成功')
      } else {
        // 添加新教师
        const createData = {
          username: teacherForm.username,
          password: teacherForm.password,
          name: teacherForm.name,
          role: teacherForm.role,
          isAdmin: teacherForm.role === 'admin' ? 1 : 0
        }
        const response = await teacherAPI.createTeacher(createData)
        const newTeacher = {
          ...response,
          id: response.id,
          status: 'active',
          createTime: new Date().toLocaleString(),
          lastLogin: '从未登录'
        }
        teachers.value.push(newTeacher)
        ElMessage.success('教师账号添加成功')
      }
    
    showAddDialog.value = false
    resetForm()
  } catch (error) {
    ElMessage.error('请完善表单信息')
  }
}

const resetForm = () => {
  teacherFormRef.value?.resetFields()
  Object.keys(teacherForm).forEach(key => {
    if (key !== 'role') teacherForm[key] = ''
  })
  teacherForm.role = 'teacher'
  isEditMode.value = false
}
</script>

<style scoped>
.account-management-page {
  min-height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
