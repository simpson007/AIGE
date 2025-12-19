<template>
  <div class="validator-config">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>叙事校验器配置</span>
          <el-button type="primary" @click="saveConfig" :loading="saving">
            <el-icon><Check /></el-icon>
            保存配置
          </el-button>
        </div>
      </template>

      <el-form :model="config" label-width="140px" v-loading="loading">
        <!-- 总开关 -->
        <el-form-item label="启用校验器">
          <el-switch v-model="config.enabled" />
          <span class="form-tip">{{ config.enabled ? '已启用' : '已禁用' }}</span>
        </el-form-item>

        <el-divider content-position="left">校验功能</el-divider>

        <!-- 规则校验 -->
        <el-form-item label="规则校验">
          <el-switch v-model="config.use_rule_validation" :disabled="!config.enabled" />
          <span class="form-tip">使用AI检测禁止词汇（轮回转生、机缘次数等）</span>
        </el-form-item>

        <!-- 一致性校验 -->
        <el-form-item label="一致性校验">
          <el-switch v-model="config.use_consistency_check" :disabled="!config.enabled" />
          <span class="form-tip">使用AI检测叙事与判定结果是否一致</span>
        </el-form-item>

        <!-- 自动修正 -->
        <el-form-item label="自动修正">
          <el-switch v-model="config.use_auto_correction" :disabled="!config.enabled" />
          <span class="form-tip">发现问题时自动使用AI修正叙事内容</span>
        </el-form-item>

        <el-divider content-position="left">校验模型</el-divider>

        <!-- 模型选择 -->
        <el-form-item label="校验用模型">
          <div class="model-selector">
            <el-select 
              v-model="selectedProviderId" 
              placeholder="选择提供商" 
              @change="onProviderChange"
              :disabled="!config.enabled"
              style="width: 180px"
            >
              <el-option
                v-for="provider in enabledProviders"
                :key="provider.id"
                :label="provider.name"
                :value="provider.id"
              />
            </el-select>
            
            <el-select 
              v-model="config.validator_model_id" 
              placeholder="选择模型"
              :disabled="!config.enabled || !selectedProviderId"
              style="width: 220px"
            >
              <el-option
                v-for="model in availableModels"
                :key="model.id"
                :label="model.name"
                :value="String(model.id)"
              />
            </el-select>
          </div>
          <div class="form-tip model-tip">
            推荐使用轻量模型以降低成本：gpt-4o-mini、claude-3-haiku、gemini-1.5-flash
          </div>
        </el-form-item>

        <!-- 说明 -->
        <el-alert 
          title="校验器说明" 
          type="info" 
          :closable="false"
          style="margin-top: 20px;"
        >
          <template #default>
            <div class="alert-content">
              <p><strong>规则校验：</strong>检测叙事中是否包含禁止出现的内容（如轮回转生、机缘次数等已移除的游戏概念）</p>
              <p><strong>一致性校验：</strong>检测AI生成的叙事是否与骰子判定结果一致（如判定失败但叙事描述成功）</p>
              <p><strong>自动修正：</strong>当发现问题时，使用AI自动修正叙事内容，保持原有文风</p>
              <p><strong>注意：</strong>每次校验都会调用AI接口，会产生额外的API费用</p>
            </div>
          </template>
        </el-alert>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Check } from '@element-plus/icons-vue'
import api from '@/utils/api'
import { useAdminStore } from '@/stores/admin'

interface ValidatorConfig {
  enabled: boolean
  use_rule_validation: boolean
  use_consistency_check: boolean
  use_auto_correction: boolean
  validator_model_id: string
}

interface Model {
  id: number
  name: string
  model_id: string
  enabled: boolean
  provider_id: number
}

interface Provider {
  id: number
  name: string
  type: string
  enabled: boolean
  models?: Model[]
}

const adminStore = useAdminStore()
const loading = ref(false)
const saving = ref(false)
const providers = ref<Provider[]>([])
const selectedProviderId = ref<number | null>(null)

const config = reactive<ValidatorConfig>({
  enabled: true,
  use_rule_validation: true,
  use_consistency_check: true,
  use_auto_correction: true,
  validator_model_id: ''
})

// 启用的提供商
const enabledProviders = computed(() => {
  return providers.value.filter(p => p.enabled)
})

// 当前选中提供商的可用模型
const availableModels = computed(() => {
  if (!selectedProviderId.value) return []
  const provider = providers.value.find(p => p.id === selectedProviderId.value)
  return provider?.models?.filter(m => m.enabled) || []
})

onMounted(async () => {
  await loadProviders()
  await loadConfig()
})

// 加载配置
const loadConfig = async () => {
  loading.value = true
  try {
    const response = await api.get('/admin/game/validator-config')
    Object.assign(config, response.data)
    
    // 如果有模型ID，找到对应的提供商
    if (config.validator_model_id) {
      findProviderByModelId(config.validator_model_id)
    }
  } catch (error: any) {
    console.error('加载校验器配置失败:', error)
    // 如果是404，说明还没有配置，使用默认值
    if (error.response?.status !== 404) {
      ElMessage.error('加载配置失败')
    }
  } finally {
    loading.value = false
  }
}

// 加载提供商列表
const loadProviders = async () => {
  try {
    providers.value = await adminStore.getProviders()
    console.log('加载提供商列表:', providers.value)
  } catch (error) {
    console.error('加载提供商列表失败:', error)
  }
}

// 根据模型ID找到对应的提供商
const findProviderByModelId = (modelId: string) => {
  for (const provider of providers.value) {
    const model = provider.models?.find(m => String(m.id) === modelId)
    if (model) {
      selectedProviderId.value = provider.id
      return
    }
  }
}

// 提供商变更
const onProviderChange = () => {
  config.validator_model_id = ''
}

// 保存配置
const saveConfig = async () => {
  saving.value = true
  try {
    await api.post('/admin/game/validator-config', config)
    ElMessage.success('校验器配置已保存')
  } catch (error: any) {
    console.error('保存配置失败:', error)
    ElMessage.error(error.response?.data?.error || '保存配置失败')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.validator-config {
  max-width: 800px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

.form-tip {
  margin-left: 12px;
  color: #909399;
  font-size: 13px;
}

.model-tip {
  display: block;
  margin-left: 0;
  margin-top: 8px;
}

.model-selector {
  display: flex;
  gap: 12px;
}

.alert-content p {
  margin: 8px 0;
  font-size: 13px;
  line-height: 1.6;
}

.alert-content p:first-child {
  margin-top: 0;
}

.alert-content p:last-child {
  margin-bottom: 0;
}

:deep(.el-divider__text) {
  font-size: 14px;
  color: #606266;
  font-weight: 500;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .model-selector {
    flex-direction: column;
    gap: 8px;
  }
  
  .model-selector .el-select {
    width: 100% !important;
  }
  
  .form-tip {
    display: block;
    margin-left: 0;
    margin-top: 8px;
  }
  
  :deep(.el-form-item__label) {
    width: 100px !important;
  }
}
</style>
