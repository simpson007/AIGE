<template>
  <div class="user-management">
    <div class="page-header">
      <h3>用户管理1</h3>
      <div class="header-actions">
        <el-button type="primary" @click="openCreateDialog">
          <el-icon><Plus /></el-icon>
          新增用户
        </el-button>
        <el-button @click="refreshUsers">
          <el-icon><Refresh /></el-icon>
          刷新1
        </el-button>
      </div>
    </div>

    <el-card>
      <el-table :data="users" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" />
        <el-table-column prop="email" label="邮箱" />
        <el-table-column prop="is_admin" label="管理员" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_admin ? 'success' : 'info'">
              {{ row.is_admin ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280">
          <template #default="{ row }">
            <el-button 
              type="primary" 
              size="small" 
              @click="openEditDialog(row)"
            >
              编辑
            </el-button>
            <el-button 
              type="warning" 
              size="small" 
              @click="openPasswordDialog(row)"
            >
              密码
            </el-button>
            <el-button 
              v-if="row.id !== authStore.user?.id"
              :type="row.is_admin ? 'info' : 'success'"
              size="small" 
              @click="toggleAdmin(row)"
            >
              {{ row.is_admin ? '取消管理' : '设为管理' }}
            </el-button>
            <el-button 
              v-if="row.id !== authStore.user?.id"
              type="danger" 
              size="small" 
              @click="deleteUser(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新增/编辑用户对话框 -->
    <el-dialog
      v-model="userDialogVisible"
      :title="isEditing ? '编辑用户' : '新增用户'"
      width="450px"
    >
      <el-form 
        ref="userFormRef"
        :model="userForm"
        :rules="userRules"
        label-width="100px"
      >
        <el-form-item label="用户名" prop="username">
          <el-input
            v-model="userForm.username"
            placeholder="请输入用户名"
          />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input
            v-model="userForm.email"
            placeholder="请输入邮箱"
          />
        </el-form-item>
        <el-form-item v-if="!isEditing" label="密码" prop="password">
          <el-input
            v-model="userForm.password"
            type="password"
            show-password
            placeholder="请输入密码"
          />
        </el-form-item>
        <el-form-item label="管理员">
          <el-switch v-model="userForm.is_admin" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="userDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitUser">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 修改密码对话框 -->
    <el-dialog
      v-model="passwordDialogVisible"
      title="修改用户密码"
      width="400px"
    >
      <el-form 
        ref="passwordFormRef"
        :model="passwordForm"
        :rules="passwordRules"
        label-width="100px"
      >
        <el-form-item label="用户名">
          <el-input :value="currentUser?.username" readonly />
        </el-form-item>
        <el-form-item label="新密码" prop="newPassword">
          <el-input
            v-model="passwordForm.newPassword"
            type="password"
            show-password
            placeholder="请输入新密码"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="passwordDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="updating" @click="updatePassword">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Refresh, Plus } from '@element-plus/icons-vue'
import api from '@/utils/api'
import type { User } from '@/types'

const authStore = useAuthStore()

const users = ref<User[]>([])
const loading = ref(false)
const submitting = ref(false)
const updating = ref(false)
const userDialogVisible = ref(false)
const passwordDialogVisible = ref(false)
const currentUser = ref<User | null>(null)
const isEditing = ref(false)
const userFormRef = ref<FormInstance>()
const passwordFormRef = ref<FormInstance>()

const userForm = ref({
  username: '',
  email: '',
  password: '',
  is_admin: false
})

const passwordForm = ref({
  newPassword: ''
})

const userRules = computed<FormRules>(() => ({
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 2, max: 20, message: '用户名长度在 2 到 20 个字符', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  password: isEditing.value ? [] : [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度在 6 到 20 个字符', trigger: 'blur' }
  ]
}))

const passwordRules: FormRules = {
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度在 6 到 20 个字符', trigger: 'blur' }
  ]
}

const loadUsers = async () => {
  loading.value = true
  try {
    const response = await api.get('/admin/users')
    users.value = response.users || []
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '加载用户列表失败')
  } finally {
    loading.value = false
  }
}

const refreshUsers = () => {
  loadUsers()
}

const openCreateDialog = () => {
  isEditing.value = false
  currentUser.value = null
  userForm.value = {
    username: '',
    email: '',
    password: '',
    is_admin: false
  }
  userDialogVisible.value = true
}

const openEditDialog = (user: User) => {
  isEditing.value = true
  currentUser.value = user
  userForm.value = {
    username: user.username,
    email: user.email,
    password: '',
    is_admin: user.is_admin
  }
  userDialogVisible.value = true
}

const submitUser = async () => {
  if (!userFormRef.value) return
  
  const valid = await userFormRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (isEditing.value && currentUser.value) {
      // 编辑用户
      await api.put(`/admin/users/${currentUser.value.id}`, {
        username: userForm.value.username,
        email: userForm.value.email,
        is_admin: userForm.value.is_admin
      })
      ElMessage.success('用户信息更新成功')
    } else {
      // 新增用户
      await api.post('/admin/users', {
        username: userForm.value.username,
        email: userForm.value.email,
        password: userForm.value.password,
        is_admin: userForm.value.is_admin
      })
      ElMessage.success('用户创建成功')
    }
    userDialogVisible.value = false
    loadUsers()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || (isEditing.value ? '更新用户失败' : '创建用户失败'))
  } finally {
    submitting.value = false
  }
}

const openPasswordDialog = (user: User) => {
  currentUser.value = user
  passwordForm.value.newPassword = ''
  passwordDialogVisible.value = true
}

const updatePassword = async () => {
  if (!passwordFormRef.value || !currentUser.value) return
  
  const valid = await passwordFormRef.value.validate().catch(() => false)
  if (!valid) return

  updating.value = true
  try {
    await api.put(`/admin/users/${currentUser.value.id}/password`, {
      new_password: passwordForm.value.newPassword
    })
    ElMessage.success('密码修改成功')
    passwordDialogVisible.value = false
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '密码修改失败')
  } finally {
    updating.value = false
  }
}

const toggleAdmin = async (user: User) => {
  try {
    const action = user.is_admin ? '取消管理员权限' : '设为管理员'
    await ElMessageBox.confirm(
      `确定要${action}用户 "${user.username}" 吗？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    await api.put(`/admin/users/${user.id}/toggle-admin`)
    ElMessage.success(`${action}成功`)
    loadUsers()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '操作失败')
    }
  }
}

const deleteUser = async (user: User) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户 "${user.username}" 吗？此操作不可恢复。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    await api.delete(`/admin/users/${user.id}`)
    ElMessage.success('用户删除成功')
    loadUsers()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '删除用户失败')
    }
  }
}

onMounted(() => {
  loadUsers()
})
</script>

<style scoped>
.user-management {
  background: white;
  border-radius: 8px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h3 {
  margin: 0;
  color: #333;
  font-size: 18px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 10px;
}
</style>
