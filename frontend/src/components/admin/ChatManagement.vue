<template>
  <div class="chat-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>聊天记录管理</span>
          <div class="header-actions">
            <el-button @click="showStats = true" type="info">
              <el-icon><DataAnalysis /></el-icon>
              统计
            </el-button>
            <el-button @click="exportChats" type="success">
              <el-icon><Download /></el-icon>
              导出
            </el-button>
            <el-button type="primary" @click="refreshChats">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
        </div>
      </template>

      <!-- 搜索和过滤 -->
      <div class="search-filters">
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :md="6">
            <el-input
              v-model="filters.search"
              placeholder="搜索聊天内容..."
              clearable
              @change="handleSearch"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
          </el-col>
          <el-col :xs="24" :sm="12" :md="6">
            <el-select
              v-model="filters.userId"
              placeholder="选择用户"
              clearable
              @change="handleFilterChange"
            >
              <el-option
                v-for="user in userList"
                :key="user.id"
                :label="user.username"
                :value="user.id"
              />
            </el-select>
          </el-col>
          <el-col :xs="24" :sm="12" :md="6">
            <el-select
              v-model="filters.modId"
              placeholder="选择MOD"
              clearable
              @change="handleFilterChange"
            >
              <el-option
                v-for="mod in modList"
                :key="mod.id"
                :label="mod.name"
                :value="mod.id"
              />
            </el-select>
          </el-col>
        </el-row>
      </div>

      <!-- PC端表格 -->
      <div class="table-container pc-table">
        <el-table
          :data="chatRecords"
          v-loading="loading"
          style="width: 100%"
          @row-click="handleRowClick"
          class="chat-table"
        >
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="username" label="用户" min-width="120" />
          <el-table-column prop="mod_id" label="MOD" min-width="120" />
          <el-table-column prop="session_date" label="会话日期" min-width="120" />
          <el-table-column label="内容预览" min-width="200">
            <template #default="scope">
              <div class="content-preview">
                {{ getContentPreview(scope.row.display_history || scope.row.recent_history) }}
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="updated_at" label="更新时间" min-width="160">
            <template #default="scope">
              {{ formatTime(scope.row.updated_at) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="scope">
              <el-button-group size="small">
                <el-button @click.stop="viewChat(scope.row)" type="primary">
                  <el-icon><View /></el-icon>
                </el-button>
                <el-button @click.stop="editChat(scope.row)" type="warning">
                  <el-icon><Edit /></el-icon>
                </el-button>
                <el-button @click.stop="deleteChat(scope.row)" type="danger">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </el-button-group>
            </template>
          </el-table-column>
        </el-table>

        <!-- 分页 -->
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
          class="pagination"
        />
      </div>

      <!-- 移动端列表 -->
      <div class="mobile-chat-list">
        <div
          v-for="chat in chatRecords"
          :key="chat.id"
          class="mobile-chat-card"
          @click="viewChat(chat)"
        >
          <div class="chat-info">
            <div class="chat-header">
              <span class="username">{{ chat.username }}</span>
              <el-tag size="small">{{ chat.mod_id }}</el-tag>
            </div>
            <div class="chat-meta">
              <span>ID: {{ chat.id }}</span>
              <span>{{ formatTime(chat.updated_at) }}</span>
            </div>
            <div class="chat-content">
              {{ getContentPreview(chat.display_history || chat.recent_history) }}
            </div>
          </div>
          <div class="chat-actions">
            <el-button
              size="small"
              circle
              @click.stop="editChat(chat)"
              type="warning"
            >
              <el-icon><Edit /></el-icon>
            </el-button>
            <el-button
              size="small"
              circle
              @click.stop="deleteChat(chat)"
              type="danger"
            >
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
        </div>

        <!-- 移动端分页 -->
        <div class="mobile-pagination">
          <el-pagination
            v-model:current-page="pagination.page"
            :total="pagination.total"
            :page-size="pagination.pageSize"
            layout="prev, pager, next"
            @current-change="handlePageChange"
          />
        </div>
      </div>
    </el-card>

    <!-- 查看聊天记录对话框 -->
    <el-dialog
      v-model="viewDialogVisible"
      title="聊天记录详情"
      width="70%"
      class="chat-view-dialog"
    >
      <div v-if="currentChat" class="chat-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="ID">{{ currentChat.id }}</el-descriptions-item>
          <el-descriptions-item label="用户">{{ currentChat.username }}</el-descriptions-item>
          <el-descriptions-item label="MOD">{{ currentChat.mod_id }}</el-descriptions-item>
          <el-descriptions-item label="会话日期">{{ currentChat.session_date }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatTime(currentChat.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ formatTime(currentChat.updated_at) }}</el-descriptions-item>
          <el-descriptions-item label="压缩轮次" :span="2">{{ currentChat.compression_round }}</el-descriptions-item>
        </el-descriptions>

        <el-divider>聊天内容</el-divider>

        <div class="chat-content-view">
          <el-tabs type="border-card">
            <el-tab-pane label="显示历史">
              <div class="content-container">
                <pre>{{ formatChatContent(currentChat.display_history) }}</pre>
              </div>
            </el-tab-pane>
            <el-tab-pane label="最近历史">
              <div class="content-container">
                <pre>{{ formatChatContent(currentChat.recent_history) }}</pre>
              </div>
            </el-tab-pane>
            <el-tab-pane label="压缩摘要">
              <div class="content-container">
                <pre>{{ formatChatContent(currentChat.compressed_summary) }}</pre>
              </div>
            </el-tab-pane>
            <el-tab-pane label="游戏状态">
              <div class="content-container">
                <pre>{{ formatChatContent(currentChat.state) }}</pre>
              </div>
            </el-tab-pane>
          </el-tabs>
        </div>
      </div>
    </el-dialog>

    <!-- 编辑聊天记录对话框 -->
    <el-dialog
      v-model="editDialogVisible"
      title="编辑聊天记录"
      width="70%"
      class="chat-edit-dialog"
    >
      <el-form v-if="editingChat" :model="editingChat" label-width="120px">
        <el-form-item label="显示历史">
          <el-input
            v-model="editingChat.display_history"
            type="textarea"
            :rows="6"
            placeholder="输入显示历史..."
          />
        </el-form-item>
        <el-form-item label="最近历史">
          <el-input
            v-model="editingChat.recent_history"
            type="textarea"
            :rows="6"
            placeholder="输入最近历史..."
          />
        </el-form-item>
        <el-form-item label="压缩摘要">
          <el-input
            v-model="editingChat.compressed_summary"
            type="textarea"
            :rows="4"
            placeholder="输入压缩摘要..."
          />
        </el-form-item>
        <el-form-item label="游戏状态">
          <el-input
            v-model="editingChat.state"
            type="textarea"
            :rows="4"
            placeholder="输入游戏状态..."
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="editDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="saveChat" :loading="saving">
            保存
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 统计对话框 -->
    <el-dialog
      v-model="showStats"
      title="聊天统计"
      width="50%"
    >
      <el-row :gutter="20" v-loading="statsLoading">
        <el-col :span="12">
          <el-statistic title="总聊天数" :value="stats.total_chats" />
        </el-col>
        <el-col :span="12">
          <el-statistic title="活跃用户" :value="stats.active_users" />
        </el-col>
        <el-col :span="12">
          <el-statistic title="今日聊天" :value="stats.today_chats" />
        </el-col>
        <el-col :span="12">
          <el-statistic
            title="平均聊天数/用户"
            :value="stats.average_chats_per_user"
            :precision="2"
          />
        </el-col>
      </el-row>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search,
  Refresh,
  View,
  Edit,
  Delete,
  Download,
  DataAnalysis
} from '@element-plus/icons-vue'
import { useAdminStore } from '@/stores/admin'
import { useAuthStore } from '@/stores/auth'

interface ChatRecord {
  id: number
  user_id: number
  username: string
  mod_id: string
  session_date: string
  state: string
  recent_history: string
  compressed_summary: string
  compression_round: number
  display_history: string
  created_at: string
  updated_at: string
}

const adminStore = useAdminStore()
const authStore = useAuthStore()

const loading = ref(false)
const saving = ref(false)
const statsLoading = ref(false)
const chatRecords = ref<ChatRecord[]>([])
const currentChat = ref<ChatRecord | null>(null)
const editingChat = ref<ChatRecord | null>(null)
const viewDialogVisible = ref(false)
const editDialogVisible = ref(false)
const showStats = ref(false)

const userList = ref<any[]>([])
const modList = ref<any[]>([
  { id: 'xiuxian2', name: '修仙2' },
  { id: 'guzhenren', name: '古真人' }
])

const filters = reactive({
  search: '',
  userId: '',
  modId: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const stats = reactive({
  total_chats: 0,
  active_users: 0,
  today_chats: 0,
  average_chats_per_user: 0
})

onMounted(() => {
  loadChats()
  loadUsers()
  loadStats()
})

const loadUsers = async () => {
  try {
    const users = await adminStore.getUsers()
    userList.value = users
  } catch (error) {
    console.error('加载用户列表失败:', error)
  }
}

const loadStats = async () => {
  try {
    statsLoading.value = true
    const response = await fetch('/api/admin/chats/stats', {
      headers: {
        'Authorization': `Bearer ${authStore.token}`
      }
    })
    const data = await response.json()
    Object.assign(stats, data)
  } catch (error) {
    console.error('加载统计信息失败:', error)
  } finally {
    statsLoading.value = false
  }
}

const loadChats = async () => {
  loading.value = true
  try {
    const params = new URLSearchParams({
      page: pagination.page.toString(),
      page_size: pagination.pageSize.toString(),
      ...(filters.search && { search: filters.search }),
      ...(filters.userId && { user_id: filters.userId }),
      ...(filters.modId && { mod_id: filters.modId })
    })

    const response = await fetch(`/api/admin/chats?${params}`, {
      headers: {
        'Authorization': `Bearer ${authStore.token}`
      }
    })

    if (!response.ok) {
      throw new Error('加载聊天记录失败')
    }

    const data = await response.json()
    chatRecords.value = data.records || []
    pagination.total = data.total || 0
  } catch (error) {
    ElMessage.error('加载聊天记录失败')
    console.error(error)
  } finally {
    loading.value = false
  }
}

const refreshChats = () => {
  loadChats()
  loadStats()
}

const handleSearch = () => {
  pagination.page = 1
  loadChats()
}

const handleFilterChange = () => {
  pagination.page = 1
  loadChats()
}

const handlePageChange = (page: number) => {
  pagination.page = page
  loadChats()
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  loadChats()
}

const handleRowClick = (row: ChatRecord) => {
  viewChat(row)
}

const viewChat = (chat: ChatRecord) => {
  currentChat.value = chat
  viewDialogVisible.value = true
}

const editChat = (chat: ChatRecord) => {
  editingChat.value = { ...chat }
  editDialogVisible.value = true
}

const saveChat = async () => {
  if (!editingChat.value) return

  try {
    saving.value = true
    const response = await fetch(`/api/admin/chats/${editingChat.value.id}`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${authStore.token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        state: editingChat.value.state,
        recent_history: editingChat.value.recent_history,
        compressed_summary: editingChat.value.compressed_summary,
        display_history: editingChat.value.display_history
      })
    })

    if (!response.ok) {
      throw new Error('保存失败')
    }

    ElMessage.success('聊天记录更新成功')
    editDialogVisible.value = false
    loadChats()
  } catch (error) {
    ElMessage.error('保存聊天记录失败')
    console.error(error)
  } finally {
    saving.value = false
  }
}

const deleteChat = async (chat: ChatRecord) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户 "${chat.username}" 的聊天记录吗？此操作不可恢复！`,
      '确认删除',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    const response = await fetch(`/api/admin/chats/${chat.id}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${authStore.token}`
      }
    })

    if (!response.ok) {
      throw new Error('删除失败')
    }

    ElMessage.success('聊天记录删除成功')
    loadChats()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('删除聊天记录失败')
      console.error(error)
    }
  }
}

const exportChats = async () => {
  try {
    const params = new URLSearchParams({
      ...(filters.userId && { user_id: filters.userId }),
      format: 'json'
    })

    const response = await fetch(`/api/admin/chats/export?${params}`, {
      headers: {
        'Authorization': `Bearer ${authStore.token}`
      }
    })

    if (!response.ok) {
      throw new Error('导出失败')
    }

    const blob = await response.blob()
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `chat_export_${new Date().toISOString()}.json`
    a.click()
    window.URL.revokeObjectURL(url)

    ElMessage.success('导出成功')
  } catch (error) {
    ElMessage.error('导出聊天记录失败')
    console.error(error)
  }
}

const getContentPreview = (content: string) => {
  if (!content) return '无内容'
  const preview = content.replace(/[\r\n]+/g, ' ').substring(0, 100)
  return preview + (content.length > 100 ? '...' : '')
}

const formatTime = (time: string) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

const formatChatContent = (content: string) => {
  if (!content) return '无内容'
  try {
    const parsed = JSON.parse(content)
    return JSON.stringify(parsed, null, 2)
  } catch {
    return content
  }
}
</script>

<style scoped>
.chat-management {
  width: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.search-filters {
  margin-bottom: 20px;
}

.table-container {
  width: 100%;
}

.chat-table {
  width: 100%;
}

.content-preview {
  color: #666;
  font-size: 13px;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.chat-detail {
  width: 100%;
}

.chat-content-view {
  margin-top: 20px;
}

.content-container {
  max-height: 400px;
  overflow-y: auto;
  background: #f5f5f5;
  padding: 10px;
  border-radius: 4px;
}

.content-container pre {
  white-space: pre-wrap;
  word-wrap: break-word;
  margin: 0;
  font-family: monospace;
  font-size: 12px;
}

/* 移动端样式 */
.mobile-chat-list {
  display: none;
}

@media (max-width: 1024px) {
  .pc-table {
    display: none;
  }

  .mobile-chat-list {
    display: block;
  }

  .mobile-chat-card {
    background: white;
    border: 1px solid #e0e0e0;
    border-radius: 8px;
    margin-bottom: 12px;
    padding: 12px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    cursor: pointer;
    transition: all 0.3s;
  }

  .mobile-chat-card:hover {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }

  .chat-info {
    flex: 1;
    min-width: 0;
  }

  .chat-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
  }

  .username {
    font-weight: 600;
    color: #333;
  }

  .chat-meta {
    display: flex;
    gap: 10px;
    font-size: 12px;
    color: #999;
    margin-bottom: 8px;
  }

  .chat-content {
    font-size: 14px;
    color: #666;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .chat-actions {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-left: 12px;
  }

  .mobile-pagination {
    display: flex;
    justify-content: center;
    margin-top: 20px;
  }

  .chat-view-dialog :deep(.el-dialog),
  .chat-edit-dialog :deep(.el-dialog) {
    width: 95% !important;
  }

  .header-actions {
    flex-wrap: wrap;
  }

  .header-actions .el-button {
    margin: 2px;
  }

  .search-filters .el-col {
    margin-bottom: 10px;
  }
}

@media (max-width: 480px) {
  .mobile-chat-card {
    flex-direction: column;
    align-items: flex-start;
  }

  .chat-actions {
    flex-direction: row;
    margin-left: 0;
    margin-top: 10px;
    width: 100%;
    justify-content: flex-end;
  }

  .header-actions .el-button span {
    display: none;
  }

  .header-actions .el-button {
    padding: 8px;
  }
}
</style>