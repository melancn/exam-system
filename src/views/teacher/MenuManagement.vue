<template>
  <div class="menu-management-page">
    <div class="page-header">
      <h2>菜单权限管理</h2>
    </div>
    
    <div class="page-content">
      <el-card class="card-container">
        <template #header>
          <div class="card-header">
            <h3>菜单列表</h3>
            <div>
              <el-button type="primary" @click="addTopMenu">
                添加顶级菜单
              </el-button>
              <el-button @click="expandAll">展开全部</el-button>
              <el-button @click="collapseAll">折叠全部</el-button>
            </div>
          </div>
        </template>
        
        <el-table
          :data="menus"
          row-key="id"
          :tree-props="{children: 'children', hasChildren: 'hasChildren'}"
          style="width: 100%"
        >
          <el-table-column prop="name" label="菜单名称" width="200" />
          <el-table-column prop="path" label="路由路径" width="200" />
          <el-table-column prop="icon" label="图标" width="120">
            <template #default="scope">
              <el-icon v-if="scope.row.icon">
                <component :is="scope.row.icon" />
              </el-icon>
            </template>
          </el-table-column>
          <el-table-column label="权限控制" width="120">
            <template #default="scope">
              <el-tag :type="scope.row.requiresAdmin ? 'danger' : 'success'">
                {{ scope.row.requiresAdmin ? '管理员' : '所有教师' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="scope">
              <el-switch
                v-model="scope.row.enabled"
                active-text="启用"
                inactive-text="禁用"
              />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="scope">
              <el-button size="small" @click="addSubMenu(scope.row)">添加子菜单</el-button>
              <el-button size="small" @click="editMenu(scope.row)">编辑</el-button>
              <el-button 
                size="small" 
                type="danger" 
                @click="deleteMenu(scope.row)"
                :disabled="scope.row.children && scope.row.children.length > 0"
              >
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <!-- 添加/编辑菜单对话框 -->
    <el-dialog v-model="showMenuDialog" :title="isEditMode ? '编辑菜单' : '添加菜单'" width="500px">
      <el-form :model="menuForm" :rules="rules" ref="menuFormRef" label-width="100px">
        <el-form-item label="菜单名称" prop="name">
          <el-input v-model="menuForm.name" placeholder="请输入菜单名称" />
        </el-form-item>
        <el-form-item label="路由路径" prop="path">
          <el-input v-model="menuForm.path" placeholder="请输入路由路径" />
        </el-form-item>
        <el-form-item label="图标" prop="icon">
          <el-select v-model="menuForm.icon" placeholder="请选择图标" clearable>
            <el-option
              v-for="icon in icons"
              :key="icon"
              :label="icon"
              :value="icon"
            >
              <el-icon><component :is="icon" /></el-icon>
              <span style="margin-left: 10px;">{{ icon }}</span>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="权限控制" prop="requiresAdmin">
          <el-switch
            v-model="menuForm.requiresAdmin"
            active-text="仅管理员"
            inactive-text="所有教师"
          />
        </el-form-item>
        <el-form-item label="父级菜单" prop="parentId" v-if="!isTopMenu">
          <el-select v-model="menuForm.parentId" placeholder="请选择父级菜单" style="width: 100%">
            <el-option
              v-for="menu in parentMenus"
              :key="menu.id"
              :label="menu.name"
              :value="menu.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showMenuDialog = false">取消</el-button>
        <el-button type="primary" @click="submitForm">
          {{ isEditMode ? '保存修改' : '添加菜单' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

const icons = Object.keys(ElementPlusIconsVue)

const showMenuDialog = ref(false)
const isEditMode = ref(false)
const isTopMenu = ref(false)
const menuFormRef = ref()

const menuForm = reactive({
  id: '',
  name: '',
  path: '',
  icon: '',
  requiresAdmin: false,
  parentId: null,
  enabled: true
})

const rules = {
  name: [{ required: true, message: '请输入菜单名称', trigger: 'blur' }],
  path: [{ required: true, message: '请输入路由路径', trigger: 'blur' }]
}

const menus = ref([
  {
    id: 1,
    name: '学生管理',
    path: '/teacher/student-management',
    icon: 'User',
    requiresAdmin: false,
    enabled: true,
    children: [
      {
        id: 4,
        name: '学生账号录入',
        path: '/teacher/student-management',
        icon: 'User',
        requiresAdmin: false,
        enabled: true
      }
    ]
  },
  {
    id: 2,
    name: '考试管理',
    path: '',
    icon: 'Tickets',
    requiresAdmin: false,
    enabled: true,
    children: [
      {
        id: 5,
        name: '试卷分配',
        path: '/teacher/exam-assignment',
        icon: 'Tickets',
        requiresAdmin: false,
        enabled: true
      },
      {
        id: 6,
        name: '试卷编辑',
        path: '/teacher/exam-edit',
        icon: 'Edit',
        requiresAdmin: false,
        enabled: true
      }
    ]
  },
  {
    id: 3,
    name: '系统管理',
    path: '',
    icon: 'Setting',
    requiresAdmin: true,
    enabled: true,
    children: [
      {
        id: 7,
        name: '教师账号管理',
        path: '/teacher/account-management',
        icon: 'User',
        requiresAdmin: true,
        enabled: true
      },
      {
        id: 8,
        name: '菜单权限管理',
        path: '/teacher/menu-management',
        icon: 'Menu',
        requiresAdmin: true,
        enabled: true
      }
    ]
  }
])

const parentMenus = computed(() => {
  return menus.value.filter(menu => !menu.path)
})

const expandAll = () => {
  menus.value.forEach(menu => {
    if (menu.children) {
      menu.expanded = true
    }
  })
}

const collapseAll = () => {
  menus.value.forEach(menu => {
    if (menu.children) {
      menu.expanded = false
    }
  })
}

const addTopMenu = () => {
  isTopMenu.value = true
  resetForm()
  showMenuDialog.value = true
}

const addSubMenu = (parentMenu) => {
  isTopMenu.value = false
  resetForm()
  menuForm.parentId = parentMenu.id
  showMenuDialog.value = true
}

const editMenu = (menu) => {
  isTopMenu.value = !menu.parentId
  Object.assign(menuForm, menu)
  isEditMode.value = true
  showMenuDialog.value = true
}

const deleteMenu = async (menu) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除菜单 "${menu.name}" 吗？此操作不可恢复。`,
      '删除菜单',
      { type: 'warning' }
    )
    
    if (menu.parentId) {
      // 删除子菜单
      const parent = menus.value.find(m => m.id === menu.parentId)
      if (parent) {
        parent.children = parent.children.filter(m => m.id !== menu.id)
      }
    } else {
      // 删除顶级菜单
      menus.value = menus.value.filter(m => m.id !== menu.id)
    }
    
    ElMessage.success('菜单删除成功')
  } catch {
    // 用户取消操作
  }
}

const submitForm = async () => {
  if (!menuFormRef.value) return
  
  try {
    await menuFormRef.value.validate()
    
    if (isEditMode.value) {
      // 更新现有菜单
      const updateMenuInTree = (tree, menu) => {
        for (let i = 0; i < tree.length; i++) {
          if (tree[i].id === menu.id) {
            tree[i] = { ...menu }
            return true
          }
          if (tree[i].children) {
            if (updateMenuInTree(tree[i].children, menu)) {
              return true
            }
          }
        }
        return false
      }
      
      updateMenuInTree(menus.value, menuForm)
      ElMessage.success('菜单更新成功')
    } else {
      // 添加新菜单
      const newMenu = {
        ...menuForm,
        id: Date.now(),
        enabled: true
      }
      
      if (isTopMenu.value) {
        newMenu.children = []
        menus.value.push(newMenu)
      } else {
        const parent = menus.value.find(m => m.id === menuForm.parentId)
        if (parent) {
          if (!parent.children) {
            parent.children = []
          }
          parent.children.push(newMenu)
        }
      }
      
      ElMessage.success('菜单添加成功')
    }
    
    showMenuDialog.value = false
    resetForm()
  } catch (error) {
    ElMessage.error('请完善表单信息')
  }
}

const resetForm = () => {
  menuFormRef.value?.resetFields()
  Object.keys(menuForm).forEach(key => {
    menuForm[key] = key === 'requiresAdmin' ? false : null
  })
  menuForm.enabled = true
  isEditMode.value = false
}
</script>

<style scoped>
.menu-management-page {
  min-height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>