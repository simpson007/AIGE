<template>
  <div class="provider-management">
    <!-- 游戏AI配置区域 -->
    <div class="game-ai-config" style="margin-bottom: 24px;">
      <div class="config-header">
        <h3>游戏AI配置</h3>
        <el-button type="success" size="small" @click="saveGameConfig" :loading="savingGameConfig">
          保存配置
        </el-button>
      </div>
      
      <div class="config-content">
        <el-form label-width="120px">
          <el-form-item label="默认模型">
            <div class="default-model-selectors">
              <el-select 
                v-model="defaultSelectedProviderId" 
                placeholder="选择提供商" 
                @change="onDefaultProviderChange"
                size="small"
                style="width: 150px"
              >
                <el-option
                  v-for="provider in enabledProviders"
                  :key="provider.id"
                  :label="provider.name"
                  :value="provider.id"
                />
              </el-select>
              
              <el-select 
                v-model="defaultSelectedModelId" 
                placeholder="选择模型"
                :disabled="!defaultSelectedProviderId"
                @change="onDefaultModelChange"
                size="small"
                style="width: 200px"
              >
                <el-option
                  v-for="model in defaultAvailableModels"
                  :key="model.id"
                  :label="model.name"
                  :value="model.model_id"
                />
              </el-select>
              
            </div>
          </el-form-item>
          
          <el-form-item label="游戏专用模型">
            <div class="game-model-list">
              <div v-for="mod in availableMods" :key="mod.game_id" class="game-model-item">
                <div class="game-info">
                  <span class="game-name">{{ mod.name }}</span>
                  <el-tag size="small" type="info">{{ mod.game_id }}</el-tag>
                </div>
                
                <div class="model-selectors">
                  <el-select 
                    v-model="gameProviderSelections[mod.game_id]" 
                    placeholder="选择提供商" 
                    @change="() => onGameProviderChange(mod.game_id)"
                    size="small"
                    style="width: 150px"
                  >
                    <el-option
                      v-for="provider in enabledProviders"
                      :key="provider.id"
                      :label="provider.name"
                      :value="provider.id"
                    />
                  </el-select>
                  
                  <el-select 
                    v-model="gameModelSelections[mod.game_id]"
                    placeholder="选择模型"
                    :disabled="!gameProviderSelections[mod.game_id]"
                    @change="() => onGameModelChange(mod.game_id)"
                    size="small"
                    style="width: 200px"
                  >
                    <el-option
                      v-for="model in getGameAvailableModels(mod.game_id)"
                      :key="model.id"
                      :label="model.name"
                      :value="model.model_id"
                    />
                  </el-select>
                </div>
              </div>
              
              <div v-if="!availableMods || availableMods.length === 0" class="no-games">
                暂无可用游戏
              </div>
            </div>
          </el-form-item>
        </el-form>
      </div>
    </div>

    <el-divider />

    <div style="margin-bottom: 24px;">
      <div class="page-header">
        <h3>AI服务提供商</h3>
        <el-button type="primary" @click="showProviderTypeDialog = true">
          <el-icon><Plus /></el-icon>
          <span>添加提供商</span>
        </el-button>
      </div>

      <div class="info-box">
        <h4 class="info-title">API配置说明</h4>
        <div class="info-content">
          <div><strong>OpenAI及兼容服务：</strong>API URL填写完整路径，如 <code class="bg-blue-100 px-1 rounded">https://api.openai.com/v1/chat/completions</code></div>
          <div><strong>Anthropic Claude：</strong>API URL填写 <code class="bg-blue-100 px-1 rounded">https://api.anthropic.com/v1/messages</code></div>
          <div><strong>Google Gemini：</strong>API URL填写 <code class="bg-blue-100 px-1 rounded">https://generativelanguage.googleapis.com/v1beta</code></div>
          <div><strong>自定义提供商：</strong>大多数第三方服务使用OpenAI兼容格式</div>
        </div>
      </div>

      <div v-if="providers.length === 0" class="empty-state">
        <el-icon style="font-size: 60px; color: #dcdfe6; margin-bottom: 16px;"><Setting /></el-icon>
        <p>还没有配置任何AI提供商</p>
        <p style="font-size: 14px; margin-top: 8px; color: #909399;">点击上方按钮添加您的第一个AI服务</p>
      </div>

      <div v-else class="providers-list">
        <div
          v-for="provider in providers"
          :key="provider.id"
          class="provider-card"
        >
          <div class="provider-header">
            <div class="provider-title">
              <el-checkbox
                v-model="provider.enabled"
                @change="handleToggleProvider(provider)"
              />
              <h4 style="margin: 0 8px; font-size: 16px; font-weight: 500;">{{ provider.name }}</h4>
              <el-tag size="small" :type="getProviderTypeColor(provider.type)">
                {{ getProviderTypeName(provider.type) }}
              </el-tag>
              <el-icon v-if="provider.enabled && provider.api_key" style="color: #67c23a;" title="已配置">
                <SuccessFilled />
              </el-icon>
            </div>
            <div class="provider-actions">
              <el-button 
                size="small"
                @click="handleEditProvider(provider)"
                title="编辑提供商"
              >
                <el-icon><Edit /></el-icon>
              </el-button>
              <el-button
                size="small"
                type="danger"
                @click="handleDeleteProvider(provider)"
                title="删除提供商"
              >
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
          </div>

          <div class="provider-fields">
            <div class="field-item">
              <label class="field-label">API密钥</label>
              <el-input
                :model-value="provider.api_key"
                type="password"
                show-password
                placeholder="输入API密钥"
                @update:model-value="(val: string) => updateProviderField(provider, 'api_key', val)"
              />
            </div>
            <div class="field-item">
              <label class="field-label">
                API URL
                <span v-if="provider.type !== 'custom'" style="font-size: 12px; color: #909399;">(可选)</span>
              </label>
              <el-input
                :model-value="provider.base_url"
                type="url"
                :placeholder="getDefaultBaseUrl(provider.type)"
                @update:model-value="(val: string) => updateProviderField(provider, 'base_url', val)"
              />
            </div>
          </div>

          <div>
            <div class="models-header">
              <label class="field-label">可用模型 ({{ provider.models?.length || 0 }})</label>
              <el-button
                size="small"
                type="primary"
                link
                @click="showAddModelDialog(provider)"
              >
                添加模型
              </el-button>
            </div>
            <div class="models-list">
              <div v-if="!provider.models || provider.models.length === 0" class="models-empty">
                暂无模型
              </div>
              <div
                v-for="model in provider.models"
                :key="model.id"
                class="model-item"
              >
                <div class="model-info">
                  <el-checkbox
                    v-model="model.enabled"
                    @change="handleToggleModel(model)"
                  />
                  <span class="model-name">{{ model.name }}</span>
                  <el-tag v-if="model.api_type" size="small" :type="getApiTypeColor(model.api_type)">
                    {{ getApiTypeLabel(model.api_type) }}
                  </el-tag>
                </div>
                <div class="model-actions">
                  <el-button
                    size="small"
                    @click="testModelConnection(provider, model)"
                    :loading="model.id === testingModelId"
                  >
                    测试
                  </el-button>
                  <el-button
                    size="small"
                    @click="handleEditModel(provider, model)"
                  >
                    <el-icon><Edit /></el-icon>
                  </el-button>
                  <el-button
                    size="small"
                    type="danger"
                    @click="handleDeleteModel(provider, model)"
                  >
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 选择提供商类型对话框 -->
    <el-dialog
      v-model="showProviderTypeDialog"
      title="选择提供商类型"
      width="500px"
    >
      <div class="provider-types">
        <div
          v-for="type in providerTypes"
          :key="type.value"
          class="provider-type-card"
          @click="selectProviderType(type.value)"
        >
          <h4>{{ type.label }}</h4>
          <p>{{ type.description }}</p>
        </div>
      </div>
    </el-dialog>

    <!-- 添加/编辑提供商对话框 -->
    <el-dialog
      v-model="showProviderDialog"
      :title="editingProvider ? '编辑提供商' : '添加提供商'"
      width="500px"
    >
      <el-form :model="providerForm" label-width="100px">
        <el-form-item label="类型">
          <el-tag>{{ getProviderTypeName(providerForm.type) }}</el-tag>
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="providerForm.name" placeholder="输入提供商名称" />
        </el-form-item>
        <el-form-item label="API密钥" required>
          <el-input
            v-model="providerForm.api_key"
            type="password"
            show-password
            placeholder="输入API密钥"
          />
        </el-form-item>
        <el-form-item label="API URL">
          <el-input
            v-model="providerForm.base_url"
            type="url"
            :placeholder="getDefaultBaseUrl(providerForm.type)"
          />
          <div style="font-size: 12px; color: #909399; margin-top: 4px;">留空使用默认地址</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showProviderDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveProvider">
          {{ editingProvider ? '更新' : '添加' }}
        </el-button>
      </template>
    </el-dialog>


    <!-- 添加/编辑模型对话框 -->
    <el-dialog
      v-model="showModelDialog"
      :title="editingModel ? '编辑模型' : '添加模型'"
      width="600px"
    >
      <el-form :model="modelForm" label-width="100px">
        <el-form-item label="API类型">
          <el-select v-model="modelForm.api_type" placeholder="选择API类型" @change="handleApiTypeChange">
            <el-option label="OpenAI" value="openai" />
            <el-option label="Anthropic" value="anthropic" />
            <el-option label="Google" value="google" />
          </el-select>
        </el-form-item>

        <el-form-item label="模型ID" required>
          <el-input v-model="modelForm.model_id" placeholder="如：gpt-4o" />
        </el-form-item>

        <el-form-item label="显示名称" required>
          <el-input v-model="modelForm.name" placeholder="如：GPT-4o" />
        </el-form-item>

        <el-form-item label="">
          <el-button @click="fetchModels" :loading="fetchingModels" type="primary">
            <el-icon><Refresh /></el-icon>
            获取模型列表
          </el-button>
          <span v-if="availableModels.length > 0" style="margin-left: 12px; color: #67c23a;">
            已获取 {{ availableModels.length }} 个模型
          </span>
        </el-form-item>

        <el-form-item v-if="availableModels.length > 0" label="搜索过滤">
          <el-input
            v-model="modelSearchKeyword"
            placeholder="搜索模型（支持多关键词）"
            clearable
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item v-if="availableModels.length > 0" label="可用模型">
          <div class="model-select-list" >
            <div
              v-for="modelId in filteredModels"
              :key="modelId"
              class="model-select-item"
              :class="{ 'selected': modelForm.model_id === modelId }"
              @click="selectModel(modelId)"
            >
              {{ modelId }}
            </div>
            <div v-if="filteredModels.length === 0" class="model-select-empty">
              没有匹配的模型
            </div>
          </div>
        </el-form-item>

        <el-alert v-if="modelFetchError" type="error" :title="modelFetchError" :closable="false" style="margin-bottom: 16px;" />
      </el-form>
      <template #footer>
        <el-button @click="closeModelDialog">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveModel">
          {{ editingModel ? '更新' : '添加' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick, reactive, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Edit, Delete, Setting, SuccessFilled, Search, Refresh } from '@element-plus/icons-vue'
import { useAdminStore } from '@/stores/admin'
import type { Provider, Model } from '@/types'

const adminStore = useAdminStore()

const providers = ref<Provider[]>([])
const loading = ref(false)
const showProviderTypeDialog = ref(false)
const showProviderDialog = ref(false)
const showModelDialog = ref(false)
const saving = ref(false)
const editingProvider = ref<Provider | null>(null)
const editingModel = ref<Model | null>(null)
const currentProvider = ref<Provider | null>(null)

const fetchingModels = ref(false)
const availableModels = ref<string[]>([])
const modelSearchKeyword = ref('')
const modelFetchError = ref('')
const testingModelId = ref<number | null>(null)

// 游戏配置相关
const availableMods = ref<any[]>([])
const gameConfig = reactive<{
  defaultModelId: string
  gameModels: Record<string, string>
}>({
  defaultModelId: '',
  gameModels: {}
})
const savingGameConfig = ref(false)

// 游戏模型选择相关 - 完全参考操练场的简洁实现
const defaultSelectedProviderId = ref<number | null>(null)
const defaultSelectedModelId = ref<string>('')

// 每个游戏的选择状态 - 参考操练场
const gameProviderSelections = reactive<Record<string, number | null>>({})
const gameModelSelections = reactive<Record<string, string>>({})

// 启用的提供商列表 - 参考操练场
const enabledProviders = computed(() => {
  return providers.value.filter(p => p.enabled)
})

// 所有启用的模型（扁平化列表）
const allEnabledModels = computed(() => {
  const models: any[] = []
  for (const provider of providers.value) {
    if (provider.models) {
      for (const model of provider.models) {
        if (model.enabled && provider.enabled) {
          models.push({
            ...model,
            provider: provider
          })
        }
      }
    }
  }
  return models
})

// 默认模型的可用模型列表 - 参考操练场
const defaultAvailableModels = computed(() => {
  if (!defaultSelectedProviderId.value) return []
  return allEnabledModels.value.filter(m => 
    m.provider_id === defaultSelectedProviderId.value
  )
})

// 游戏的可用模型列表 - 参考操练场
const getGameAvailableModels = (gameId: string) => {
  const providerId = gameProviderSelections[gameId]
  if (!providerId) return []
  return allEnabledModels.value.filter(m => 
    m.provider_id === providerId
  )
}

const providerForm = ref({
  type: 'openai' as 'openai' | 'anthropic' | 'google' | 'custom',
  name: '',
  api_key: '',
  base_url: ''
})

const modelForm = ref({
  model_id: '',
  name: '',
  api_type: 'openai' as string,
  provider_id: 0
})

interface ProviderType {
  value: 'openai' | 'anthropic' | 'google' | 'custom'
  label: string
  description: string
}

const providerTypes: ProviderType[] = [
  { value: 'openai', label: 'OpenAI', description: 'GPT-4, GPT-3.5等模型' },
  { value: 'anthropic', label: 'Anthropic', description: 'Claude系列模型' },
  { value: 'google', label: 'Google', description: 'Gemini系列模型' },
  { value: 'custom', label: '自定义', description: '兼容OpenAI格式的API' }
]

const filteredModels = computed(() => {
  if (!modelSearchKeyword.value.trim()) {
    return availableModels.value
  }
  const keywords = modelSearchKeyword.value.toLowerCase().trim().split(/\s+/)
  return availableModels.value.filter(modelId => {
    const modelIdLower = modelId.toLowerCase()
    return keywords.every(keyword => modelIdLower.includes(keyword))
  })
})

const getProviderTypeName = (type: string) => {
  const map: Record<string, string> = {
    openai: 'OpenAI',
    anthropic: 'Anthropic',
    google: 'Google',
    custom: '自定义'
  }
  return map[type] || type
}

const getProviderTypeColor = (type: string) => {
  const map: Record<string, any> = {
    openai: 'success',
    anthropic: '',
    google: 'warning',
    custom: 'info'
  }
  return map[type] || 'info'
}

const getDefaultBaseUrl = (type: string) => {
  const map: Record<string, string> = {
    openai: 'https://api.openai.com/v1/chat/completions',
    anthropic: 'https://api.anthropic.com/v1/messages',
    google: 'https://generativelanguage.googleapis.com/v1beta',
    custom: 'https://api.example.com/v1'
  }
  return map[type] || ''
}

const getApiTypeLabel = (type: string) => {
  const map: Record<string, string> = {
    openai: 'OpenAI',
    anthropic: 'Claude',
    google: 'Gemini'
  }
  return map[type] || type
}

const getApiTypeColor = (type: string) => {
  const map: Record<string, any> = {
    openai: 'success',
    anthropic: '',
    google: 'warning'
  }
  return map[type] || 'info'
}

const loadProviders = async () => {
  loading.value = true
  try {
    providers.value = await adminStore.getProviders()
  } catch (error: any) {
    ElMessage.error(error.message || '加载提供商失败')
  } finally {
    loading.value = false
  }
}

const selectProviderType = (type: 'openai' | 'anthropic' | 'google' | 'custom') => {
  showProviderTypeDialog.value = false
  providerForm.value = {
    type,
    name: getProviderTypeName(type),
    api_key: '',
    base_url: ''
  }
  showProviderDialog.value = true
}

const handleEditProvider = (provider: Provider) => {
  editingProvider.value = provider
  providerForm.value = {
    type: provider.type,
    name: provider.name,
    api_key: provider.api_key,
    base_url: provider.base_url || ''
  }
  showProviderDialog.value = true
}

const handleToggleProvider = async (provider: Provider) => {
  try {
    await adminStore.toggleProvider(provider.id)
  } catch (error: any) {
    provider.enabled = !provider.enabled
    ElMessage.error(error.message || '操作失败')
  }
}

const updateProviderField = async (provider: Provider, field: 'api_key' | 'base_url', value: string) => {
  const updatedData = {
    api_key: field === 'api_key' ? value : provider.api_key,
    base_url: field === 'base_url' ? value : provider.base_url
  }
  
  try {
    await adminStore.updateProvider(provider.id, updatedData)
    if (field === 'api_key') {
      provider.api_key = value
    } else {
      provider.base_url = value
    }
  } catch (error: any) {
    ElMessage.error(error.message || '更新失败')
  }
}

const handleDeleteProvider = async (provider: Provider) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除提供商 "${provider.name}" 吗？这将同时删除其所有模型配置。`,
      '确认删除',
      { type: 'warning' }
    )
    await adminStore.deleteProvider(provider.id)
    ElMessage.success('删除成功')
    await loadProviders()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

const saveProvider = async () => {
  if (!providerForm.value.name || !providerForm.value.api_key) {
    ElMessage.warning('请填写必填项')
    return
  }

  saving.value = true
  try {
    if (editingProvider.value) {
      await adminStore.updateProvider(editingProvider.value.id, providerForm.value)
      ElMessage.success('更新成功')
    } else {
      await adminStore.createProvider(providerForm.value)
      ElMessage.success('添加成功')
    }
    showProviderDialog.value = false
    editingProvider.value = null
    await loadProviders()
  } catch (error: any) {
    ElMessage.error(error.message || '保存失败')
  } finally {
    saving.value = false
  }
}

const showAddModelDialog = (provider: Provider) => {
  currentProvider.value = provider
  editingModel.value = null
  modelForm.value = {
    model_id: '',
    name: '',
    api_type: provider.type === 'custom' ? 'openai' : provider.type,
    provider_id: provider.id
  }
  availableModels.value = []
  modelSearchKeyword.value = ''
  modelFetchError.value = ''
  showModelDialog.value = true
}

const handleEditModel = (provider: Provider, model: Model) => {
  currentProvider.value = provider
  editingModel.value = model
  modelForm.value = {
    model_id: model.model_id,
    name: model.name,
    api_type: model.api_type || provider.type,
    provider_id: provider.id
  }
  availableModels.value = []
  modelSearchKeyword.value = ''
  modelFetchError.value = ''
  showModelDialog.value = true
}

const handleToggleModel = async (model: Model) => {
  try {
    await adminStore.toggleModel(model.id)
  } catch (error: any) {
    model.enabled = !model.enabled
    ElMessage.error(error.message || '操作失败')
  }
}

const handleDeleteModel = async (provider: Provider, model: Model) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除模型 "${model.name}" 吗？`,
      '确认删除',
      { type: 'warning' }
    )
    await adminStore.deleteModel(model.id)
    ElMessage.success('删除成功')
    await loadProviders()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

const testModelConnection = async (provider: Provider, model: Model) => {
  testingModelId.value = model.id
  try {
    const { AIService } = await import('@/services/aiService')
    const result = await AIService.testConnection({
      provider_id: provider.id,
      model_id: model.model_id,
      api_type: model.api_type || provider.type
    })
    
    if (result.success) {
      ElMessage.success(`模型 "${model.name}" 连接测试成功`)
    } else {
      ElMessage.error(`测试失败: ${result.message}`)
    }
  } catch (error: any) {
    ElMessage.error(`测试失败: ${error.message}`)
  } finally {
    testingModelId.value = null
  }
}

const handleApiTypeChange = () => {
  availableModels.value = []
  modelFetchError.value = ''
}

const fetchModels = async () => {
  if (!currentProvider.value) return

  fetchingModels.value = true
  modelFetchError.value = ''
  try {
    const models = await adminStore.getAvailableModels(
      currentProvider.value.id,
      modelForm.value.api_type
    )
    availableModels.value = models
    if (models.length === 0) {
      modelFetchError.value = '未找到可用模型'
    } else {
      ElMessage.success(`获取到 ${models.length} 个模型`)
    }
  } catch (error: any) {
    modelFetchError.value = error.message || '获取模型列表失败，请检查API密钥和URL配置'
  } finally {
    fetchingModels.value = false
  }
}

const selectModel = (modelId: string) => {
  modelForm.value.model_id = modelId
  modelForm.value.name = modelId
}

const saveModel = async () => {
  if (!modelForm.value.model_id || !modelForm.value.name) {
    ElMessage.warning('请填写模型ID和名称')
    return
  }

  saving.value = true
  try {
    if (editingModel.value) {
      await adminStore.updateModel(editingModel.value.id, modelForm.value)
      ElMessage.success('更新成功')
    } else {
      await adminStore.createModel(modelForm.value)
      ElMessage.success('添加成功')
    }
    closeModelDialog()
    await loadProviders()
  } catch (error: any) {
    ElMessage.error(error.message || '保存失败')
  } finally {
    saving.value = false
  }
}

const closeModelDialog = () => {
  showModelDialog.value = false
  editingModel.value = null
  currentProvider.value = null
  availableModels.value = []
  modelSearchKeyword.value = ''
  modelFetchError.value = ''
}

// 加载可用的游戏mod
const loadAvailableMods = async () => {
  try {
    const response = await fetch('/api/game/mods', {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    if (response.ok) {
      availableMods.value = await response.json()
      // 确保所有游戏都有对应的配置槽位
      for (const mod of availableMods.value) {
        if (!(mod.game_id in gameConfig.gameModels)) {
          gameConfig.gameModels[mod.game_id] = ''
        }
      }
    }
  } catch (error) {
    console.error('加载游戏列表失败:', error)
  }
}

// 加载游戏配置
const loadGameConfig = async () => {
  try {
    const response = await fetch('/api/admin/game/model-config', {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    if (response.ok) {
      const data = await response.json()
      gameConfig.defaultModelId = String(data.default_model_id || '')
      // 确保响应性更新 - 将所有值转换为字符串以匹配el-option的value类型
      const processedGameModels: Record<string, string> = {}
      if (data.game_models) {
        for (const [gameId, modelId] of Object.entries(data.game_models)) {
          processedGameModels[gameId] = String(modelId || '')
        }
      }
      gameConfig.gameModels = processedGameModels
      console.log('游戏配置加载完成')
      
      // 同步到操练场风格的变量
      initializeSelections()
    }
  } catch (error) {
    console.error('加载游戏配置失败:', error)
  }
}

// 保存游戏配置
const saveGameConfig = async () => {
  savingGameConfig.value = true
  try {
    const response = await fetch('/api/admin/game/model-config', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify({
        default_model_id: gameConfig.defaultModelId,
        game_models: gameConfig.gameModels
      })
    })
    
    if (response.ok) {
      ElMessage.success('游戏AI配置已保存')
    } else {
      const error = await response.json()
      ElMessage.error(error.error || '保存失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '保存失败')
  } finally {
    savingGameConfig.value = false
  }
}

// 事件处理方法 - 完全参考操练场的简洁实现

// 默认提供商变化处理 - 参考操练场的 onProviderChange
const onDefaultProviderChange = () => {
  defaultSelectedModelId.value = ''
  syncDefaultConfigToStorage()
}

// 默认模型变化处理
const onDefaultModelChange = () => {
  syncDefaultConfigToStorage()
}

// 游戏提供商变化处理 - 参考操练场的 onProviderChange
const onGameProviderChange = (gameId: string) => {
  gameModelSelections[gameId] = ''
  syncGameConfigToStorage(gameId)
}

// 游戏模型变化处理
const onGameModelChange = (gameId: string) => {
  syncGameConfigToStorage(gameId)
}


// 同步默认配置到存储
const syncDefaultConfigToStorage = () => {
  // 找到对应的模型ID
  if (defaultSelectedModelId.value) {
    const model = defaultAvailableModels.value.find(m => m.model_id === defaultSelectedModelId.value)
    gameConfig.defaultModelId = model ? String(model.id) : ''
  } else {
    gameConfig.defaultModelId = ''
  }
}

// 同步游戏配置到存储
const syncGameConfigToStorage = (gameId: string) => {
  // 找到对应的模型ID
  if (gameModelSelections[gameId]) {
    const models = getGameAvailableModels(gameId)
    const model = models.find(m => m.model_id === gameModelSelections[gameId])
    gameConfig.gameModels[gameId] = model ? String(model.id) : ''
  } else {
    gameConfig.gameModels[gameId] = ''
  }
}

// 初始化选择状态 - 从 gameConfig 恢复到操练场风格的变量
const initializeSelections = () => {
  // 初始化默认模型选择
  if (gameConfig.defaultModelId) {
    const model = allEnabledModels.value.find(m => String(m.id) === gameConfig.defaultModelId)
    if (model) {
      defaultSelectedProviderId.value = model.provider_id
      defaultSelectedModelId.value = model.model_id
    }
  }
  
  // 初始化游戏模型选择
  for (const [gameId, modelId] of Object.entries(gameConfig.gameModels)) {
    if (modelId) {
      const model = allEnabledModels.value.find(m => String(m.id) === modelId)
      if (model) {
        gameProviderSelections[gameId] = model.provider_id
        gameModelSelections[gameId] = model.model_id
      }
    }
  }
}


onMounted(async () => {
  await loadProviders()  // 确保先加载提供商和模型
  await loadGameConfig()  // 然后加载配置
  await loadAvailableMods() // 最后加载MOD并初始化空槽位
})
</script>

<style scoped>
.provider-management {
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.page-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.info-box {
  margin-bottom: 16px;
  padding: 16px;
  background: #ecf5ff;
  border: 1px solid #b3d8ff;
  border-radius: 8px;
}

.info-title {
  font-size: 14px;
  font-weight: 500;
  color: #409eff;
  margin: 0 0 8px 0;
}

.info-content {
  font-size: 14px;
  color: #409eff;
}

.info-content > div {
  margin-bottom: 8px;
}

code {
  font-family: 'Courier New', monospace;
  font-size: 0.9em;
  background: #e1f3f8;
  padding: 2px 4px;
  border-radius: 3px;
}

.empty-state {
  text-align: center;
  padding: 48px 0;
  color: #909399;
}

.providers-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.provider-card {
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  padding: 16px;
  background: white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.provider-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.provider-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.provider-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.provider-fields {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 12px;
  margin-bottom: 12px;
}

.field-item {
  display: flex;
  flex-direction: column;
}

.field-label {
  font-size: 14px;
  font-weight: 500;
  color: #606266;
  margin-bottom: 4px;
}

.models-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.models-list {
  max-height: 160px;
  overflow-y: auto;
  background: #f5f7fa;
  border-radius: 4px;
  padding: 8px;
}

.models-empty {
  text-align: center;
  color: #c0c4cc;
  padding: 8px;
  font-size: 14px;
}

.model-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px;
  background: white;
  border-radius: 4px;
  border: 1px solid #e4e7ed;
  margin-bottom: 4px;
}

.model-item:last-child {
  margin-bottom: 0;
}

.model-info {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
}

.model-name {
  font-size: 14px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.model-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

.provider-types {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.provider-type-card {
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.3s;
}

.provider-type-card:hover {
  border-color: #409eff;
  background: #ecf5ff;
}

.provider-type-card h4 {
  margin: 0 0 8px 0;
  font-weight: 600;
}

.provider-type-card p {
  margin: 0;
  font-size: 14px;
  color: #909399;
}

.model-select-list {
  max-height: 240px;
  overflow-y: auto;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 8px;
}

.model-select-item {
  padding: 8px;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.2s;
  margin-bottom: 4px;
}

.model-select-item:hover {
  background: #ecf5ff;
}

.model-select-item.selected {
  background: #409eff;
  color: white;
}

.model-select-empty {
  text-align: center;
  color: #909399;
  padding: 16px;
  font-size: 14px;
}

/* 游戏AI配置样式 */
.game-ai-config {
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  padding: 16px;
  background: white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.config-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.config-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.config-content {
  background: #f5f7fa;
  border-radius: 4px;
  padding: 16px;
}

.game-model-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.game-model-item {
  display: flex;
  align-items: center;
  padding: 12px;
  background: white;
  border-radius: 4px;
  border: 1px solid #e4e7ed;
  flex-wrap: wrap;
  gap: 12px;
}

.game-info {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 200px;
}

.game-name {
  font-weight: 500;
  font-size: 14px;
}

.no-games {
  text-align: center;
  color: #909399;
  padding: 16px;
  font-size: 14px;
}


/* 默认模型选择器样式 */
.default-model-selectors {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  background: #fafbfc;
  border-radius: 6px;
  border: 1px solid #e4e7ed;
  flex-wrap: wrap;
}

.default-model-selectors .el-select {
  min-width: 120px;
  flex: 1;
}

/* 游戏模型选择器样式 */
.game-model-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.game-model-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  background: white;
  border-radius: 6px;
  border: 1px solid #e4e7ed;
  gap: 16px;
}

.game-model-item:hover {
  border-color: #409eff;
  background: #fafbfc;
}

.model-selectors {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  flex-wrap: wrap;
  min-width: 0;
}

.model-selectors .el-select {
  min-width: 0;
  flex: 1;
}

/* 移动端适配 */
@media (max-width: 1024px) {
  .provider-management {
    padding: 0;
    margin: 0;
    box-sizing: border-box;
  }
  
  .default-model-selectors {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }
  
  .default-model-selectors {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
    padding: 16px;
    background: white;
    border-radius: 6px;
    border: 1px solid #e4e7ed;
  }
  
  .default-model-selectors .el-select {
    width: 100%;
    flex: none;
    min-width: auto;
  }
  
  .default-model-selectors :deep(.el-select .el-input) {
    width: 100%;
  }
  
  .default-model-selectors :deep(.el-select .el-input__wrapper) {
    width: 100%;
    min-height: 32px;
  }
  
  .default-model-selectors :deep(.el-select .el-input__inner) {
    width: 100%;
    height: auto;
  }
  
  .game-model-item {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }
  
  .game-info {
    min-width: auto;
    text-align: center;
  }
  
  .model-selectors {
    flex-direction: column;
    gap: 8px;
  }
  
  .model-selectors .el-select {
    width: 100%;
    flex: none;
  }
  
  .model-selectors :deep(.el-select .el-input) {
    width: 100%;
  }
  
  .model-selectors :deep(.el-select .el-input__wrapper) {
    width: 100%;
    min-height: 32px;
  }
  
  .provider-fields {
    grid-template-columns: 1fr;
    gap: 16px;
  }
  
  .page-header {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }
  
  .page-header h3 {
    text-align: center;
  }
  
  .provider-header {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }
  
  .provider-title {
    justify-content: center;
  }
  
  .provider-actions {
    justify-content: center;
  }
  
  .model-actions {
    flex-wrap: wrap;
    justify-content: center;
  }
  
  .models-header {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
    text-align: center;
  }
  
  .provider-types {
    grid-template-columns: 1fr;
  }
  
  .config-header {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
    text-align: center;
  }
}
</style>
