<template>
  <div class="game-container">
    <!-- Starry Background -->
    <div class="stars"></div>
    <div class="twinkling"></div>

    <!-- 游戏选择界面 -->
    <div v-if="!currentGame" class="game-selection">
      <div class="selection-header">
        <h1>选择你的冒险</h1>
        <p>在不同的世界中展开独一无二的旅程</p>
      </div>

      <div class="game-list">
        <div 
          v-for="mod in availableMods" 
          :key="mod.game_id"
          class="game-card"
          @click="selectGame(mod.game_id)"
        >
          <div class="game-card-icon">
            <img v-if="mod.game_id === 'FateBound'" :src="gameLogo" class="card-logo-img" />
            <span v-else>🎮</span>
          </div>
          <h3>{{ mod.name }}</h3>
          <p>{{ mod.description }}</p>
          <!--  <div class="game-card-footer">
            <span class="version">v{{ mod.version }}</span>
            <span class="author">{{ mod.author }}</span>
          </div>
          -->
        </div>
      </div>
    </div>

    <!-- 游戏进行界面 -->
    <div v-else class="game-play">
      <!-- 游戏头部 -->
      <header class="game-header">
        <div class="header-content">
          <div class="game-title">
            <img v-if="currentGame === 'FateBound'" :src="gameLogo" class="header-logo-img" />
            <h2>{{ currentModInfo?.name }}</h2>
            <span class="mod-id">{{ currentGame }}</span>
          </div>
          
          <div class="header-actions">
            <!-- PC Actions -->
            <div class="pc-actions">
              <button @click="showRestartConfirm" class="btn-icon" title="重入轮回">
                <img :src="iconRestart" class="btn-icon-img" alt="重新开始" />
                <span class="btn-label">重新开始</span>
              </button>
              <button @click="logout" class="btn-icon danger" title="退出">
                <img :src="iconExit" class="btn-icon-img" alt="退出" />
                <span class="btn-label">退出</span>
              </button>
            </div>
            
            <!-- Mobile Toggle -->
            <button class="menu-toggle" @click="toggleMobileMenu">
              ☰
            </button>
          </div>
        </div>
      </header>

      <!-- Mobile Menu Overlay -->
      <div class="mobile-menu-overlay" :class="{ active: showMobileMenu }" @click="closeMobileMenu">
        <div class="mobile-menu" @click.stop>
          <div class="mobile-menu-header">
            <h3>菜单</h3>
            <button class="close-btn" @click="closeMobileMenu">×</button>
          </div>
          <div class="mobile-menu-items">
             <button @click="toggleStatusPanel_Mobile">
              <img :src="iconCharacter" class="menu-icon-img" alt="角色状态" />
              角色状态
            </button>
            <button @click="handleMobileRestart">
              <img :src="iconRestart" class="menu-icon-img" alt="重新开始" />
              重新开始
            </button>
            <button @click="handleMobileLogout" class="danger">
              <img :src="iconExit" class="menu-icon-img" alt="退出" />
              退出
            </button>
          </div>
        </div>
      </div>

      <div class="game-content-wrapper">
        <!-- Status Panel (Sidebar on PC) -->
        <aside class="status-panel" :class="{ 'mobile-show': showStatusPanel }">
          <div class="panel-header">
            <div class="panel-title">
              <img :src="iconCharacter" class="panel-icon-img" alt="角色状态" />
              <h3>角色状态</h3>
            </div>
            <button class="close-panel-btn" @click="closeStatusPanel">×</button>
          </div>
          <div class="panel-content custom-scrollbar">
            <div v-if="filteredCurrentLife" class="character-status">
              <div v-for="(value, key) in filteredCurrentLife" :key="key" class="status-block">
                <div class="status-block-header">{{ formatKey(key) }}</div>
                <div class="status-block-content" v-html="formatValue(value)"></div>
              </div>
            </div>
            <div v-else class="no-character">
              <p>尚未开始冒险</p>
            </div>

            <!-- Soul Burn Penalties -->
            <div v-if="soulBurnPenalties.length > 0" class="soul-burn-section">
              <h4 class="soul-burn-title">🔥 燃魂代价</h4>
              <div v-for="(penalty, index) in soulBurnPenalties" :key="index" class="penalty-item">
                <span class="penalty-icon">💀</span>
                <span class="penalty-text">{{ penalty }}</span>
              </div>
            </div>
          </div>
        </aside>

        <!-- Main Game Area -->
        <main class="game-main">
          <!-- Narrative Window -->
          <div class="narrative-window custom-scrollbar" ref="narrativeWindow">
            <div 
              v-for="(text, index) in displayHistory" 
              :key="index"
              :class="['narrative-block', getBlockClass(text)]"
              v-html="renderMarkdown(text)"
            ></div>
            
             <!-- Loading indicator inside narrative -->
            <div v-if="isProcessing || isRolling" class="input-loading">
               <span class="dot">.</span><span class="dot">.</span><span class="dot">.</span>
            </div>
          </div>
          
          <!-- Input Area -->
          <div class="input-container">
            <!-- Start Button -->
            <div v-if="!sessionState?.is_in_trial && !isGameReallyEnded" class="start-game-wrapper">
               <button
                @click="startTrial"
                :disabled="!wsReady || isProcessing || isRolling"
                class="btn-adventure large"
              >
                {{ getStartButtonText() }}
                <div class="btn-shine"></div>
              </button>
            </div>

            <!-- Action Input -->
            <div v-else-if="sessionState?.is_in_trial" class="action-input-wrapper">
              <div class="input-group">
                <input
                  v-model="userInput"
                  type="text"
                  :placeholder="soulBurnMode ? '燃魂爆运中... (代价不可逆)' : '汝欲何为...'"
                  @keydown.enter="sendAction"
                  :disabled="isProcessing || isRolling"
                  :class="['adventure-input', { 'soul-burn-active': soulBurnMode }]"
                  autocomplete="off"
                />
                <button
                  @click="sendAction"
                  :disabled="isProcessing || isRolling || !userInput.trim()"
                  class="btn-adventure"
                >
                  {{ isProcessing ? '...' : '行动' }}
                </button>
              </div>
              
              <button
                @click="toggleSoulBurnMode"
                :class="['btn-soul-burn', { active: soulBurnMode }]"
                :title="soulBurnMode ? '关闭燃魂' : '开启燃魂'"
              >
                {{ soulBurnMode ? '🔥' : '💀' }}
              </button>
            </div>

            <!-- Game Ended -->
            <div v-else-if="isGameReallyEnded" class="game-ended-message">
              <p>🎉 今日功德圆满！明日再来。</p>
            </div>

             <!-- Error Recovery -->
            <div v-else-if="sessionState?.daily_success_achieved && !isGameReallyEnded" class="warning-box">
              <p>⚠️ 状态异常</p>
              <button @click="forceContinueGame" class="btn-text">强制继续</button>
            </div>
          </div>
        </main>
      </div>
    </div>

    <!-- 判定动画 Modal -->
    <div v-if="showRollAnimation" class="modal-overlay">
      <div class="roll-modal">
        <div class="dice-icon">🎲</div>
        <div class="roll-info">
          <h3>{{ rollEvent?.description || '天道裁决' }}</h3>
          <p v-if="rollEvent?.target" class="roll-target">目标: ≤ {{ rollEvent.target }}</p>
        </div>
        <div v-if="rollEvent?.result !== undefined" class="roll-result-box">
          <div class="roll-outcome" :class="rollEvent.success ? 'success' : 'fail'">
            {{ rollEvent.success ? '成功' : '失败' }}
          </div>
          <div class="roll-value">{{ rollEvent.result }}</div>
        </div>
      </div>
    </div>

    <!-- Character Creation Modal -->
    <div v-if="showCharacterCreation" class="modal-overlay">
      <div class="character-modal">
        <div class="modal-header">
          <h2>创建角色</h2>
          <div class="modal-decoration"></div>
          <button class="modal-close-btn" @click="showCharacterCreation = false">×</button>
        </div>
        
        <div class="modal-body custom-scrollbar">
          <!-- 姓名 -->
          <div class="form-group">
            <label>姓名</label>
            <div class="input-effects">
              <input 
                v-model="characterForm.name" 
                type="text" 
                maxlength="4" 
                placeholder="2-4个中文字符" 
                class="adventure-input"
              >
              <div class="focus-line"></div>
            </div>
          </div>
          
          <!-- 性别 -->
          <div class="form-group">
            <label>性别</label>
            <div class="gender-selector">
              <div 
                class="gender-option" 
                :class="{ active: characterForm.gender === '男' }"
                @click="characterForm.gender = '男'"
              >
                <span class="gender-icon">‍♂️</span>
                <span>男</span>
              </div>
              <div 
                class="gender-option" 
                :class="{ active: characterForm.gender === '女' }"
                @click="characterForm.gender = '女'"
              >
                <span class="gender-icon">‍♀️</span>
                <span>女</span>
              </div>
            </div>
          </div>

          <!-- 资质 (Custom Select) -->
          <div class="form-group">
            <label>资质</label>
            <div class="custom-select" :class="{ open: dropdowns.qualification }" @click.stop="toggleDropdown('qualification')">
              <div class="select-trigger">
                <span>{{ characterForm.qualification || '请选择资质' }}</span>
                <span class="arrow">▼</span>
              </div>
              <div v-show="dropdowns.qualification" class="select-options custom-scrollbar">
                <div 
                  v-for="opt in qualificationOptions" 
                  :key="opt.value" 
                  class="select-option"
                  :class="{ selected: characterForm.qualification === opt.value }"
                  @click.stop="selectOption('qualification', opt.value)"
                >
                  <div class="opt-main">{{ opt.value }}</div>
                  <div class="opt-desc">{{ opt.desc }}</div>
                </div>
              </div>
            </div>
          </div>

          <!-- 初始修为 (Custom Select) -->
          <div class="form-group">
            <label>初始修为</label>
            <div class="custom-select" :class="{ open: dropdowns.cultivation }" @click.stop="toggleDropdown('cultivation')">
              <div class="select-trigger">
                <span>{{ characterForm.cultivation || '请选择修为' }}</span>
                <span class="arrow">▼</span>
              </div>
              <div v-show="dropdowns.cultivation" class="select-options custom-scrollbar">
                <div 
                  v-for="opt in ['一转初阶', '一转中阶', '一转高阶']" 
                  :key="opt" 
                  class="select-option"
                  :class="{ selected: characterForm.cultivation === opt }"
                  @click.stop="selectOption('cultivation', opt)"
                >
                  {{ opt }}
                </div>
              </div>
            </div>
          </div>

          <!-- 初始元石 -->
          <div class="form-group">
            <label class="range-label">
              <span>初始元石</span>
              <span class="range-value">{{ characterForm.spiritStones }}</span>
            </label>
            <div class="range-wrapper">
              <input 
                type="range" 
                v-model.number="characterForm.spiritStones" 
                min="10" 
                max="100" 
                step="10" 
                class="adventure-range"
                :style="{ backgroundSize: ((characterForm.spiritStones - 10) * 100 / 90) + '% 100%' }"
              >
            </div>
          </div>

          <!-- 出身背景 -->
          <div class="form-group">
            <label>出身背景</label>
            <div class="textarea-wrapper">
              <textarea 
                v-model="characterForm.background" 
                rows="3" 
                placeholder="（可选）描述角色的出身背景，留空则随机..." 
                class="adventure-textarea"
              ></textarea>
            </div>
          </div>
        </div>
        
        <div class="modal-footer">
          <button @click="useRandomCharacter" class="btn-secondary">🎲 随机生成</button>
          <button @click="confirmCharacterCreation" class="btn-adventure confirm-btn">
            <span>开启征程</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Custom Confirm Modal -->
    <div v-if="confirmState.show" class="modal-overlay">
      <div class="confirm-modal">
        <div class="confirm-header">
          <h3>{{ confirmState.title }}</h3>
        </div>
        <div class="confirm-body">
          <p>{{ confirmState.message }}</p>
        </div>
        <div class="confirm-footer">
          <button @click="confirmState.show = false" class="btn-cancel">取消</button>
          <button @click="confirmState.onConfirm" class="btn-confirm" :class="confirmState.type">确定</button>
        </div>
      </div>
    </div>

    <!-- Toast Notification -->
    <div v-if="toast.show" :class="['toast', toast.type]">
      <div class="toast-icon">{{ toast.type === 'success' ? '✨' : (toast.type === 'error' ? '💀' : '📜') }}</div>
      <div class="toast-message">{{ toast.message }}</div>
    </div>
    
    <!-- Loading Overlay -->
     <div v-if="isLoading" class="loading-overlay">
      <div class="spinner"></div>
      <div class="loading-text">{{ loadingText }}</div>
    </div>

  </div>
</template>



<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { marked } from 'marked'
import gameLogo from '@/img/sumingzhixia.png'
import iconCharacter from '@/img/character.png'
import iconRestart from '@/img/restart.png'
import iconExit from '@/img/exit.png'

const router = useRouter()
const authStore = useAuthStore()

// --- Toast Logic ---
const toast = reactive({
  show: false,
  message: '',
  type: 'info'
})

const showToast = (msg: string, type: 'success' | 'error' | 'warning' | 'info' = 'info') => {
  toast.message = msg
  toast.type = type
  toast.show = true
  setTimeout(() => {
    toast.show = false
  }, 3000)
}

// --- State Management ---
const availableMods = ref<any[]>([])
const currentGame = ref<string>('')
const currentModInfo = ref<any>(null)
const gameState = ref<any>(null)
const displayHistory = computed(() => gameState.value?.display_history || [])

const sessionState = computed(() => gameState.value?.state || gameState.value || {})

// Game Ended Logic
const isGameReallyEnded = computed(() => {
  const dailySuccess = sessionState.value?.daily_success_achieved === true
  const hasExplicitEnding = sessionState.value?.ending_type !== undefined
  return dailySuccess && hasExplicitEnding
})

// Filtered Life Stats
const filteredCurrentLife = computed(() => {
  const currentLife = sessionState.value?.current_life
  if (!currentLife || typeof currentLife !== 'object') {
    return null
  }
  
  const filtered: Record<string, any> = {}
  for (const [key, value] of Object.entries(currentLife)) {
    if (value !== null && value !== undefined && value !== '' && value !== 0) {
      if (Array.isArray(value) && value.length === 0) continue
      if (typeof value === 'object' && !Array.isArray(value) && Object.keys(value).length === 0) continue
      filtered[key] = value
    }
  }
  return Object.keys(filtered).length > 0 ? filtered : null
})

const userInput = ref('')

// Character Creation State
const showCharacterCreation = ref(false)
const characterForm = ref({
  name: '',
  gender: '男',
  qualification: '丙等资质',
  cultivation: '一转初阶',
  spiritStones: 50,
  background: ''
})

const qualificationOptions = [
  { value: '丁等资质', desc: '最高二转' },
  { value: '丙等资质', desc: '通常二转' },
  { value: '乙等资质', desc: '三至四转' },
  { value: '甲等资质', desc: '可达五转' },
  { value: '十绝体', desc: '极度危险' }
]

const dropdowns = reactive({
  qualification: false,
  cultivation: false
})

function toggleDropdown(key: 'qualification' | 'cultivation') {
  // Close others
  if (key === 'qualification') dropdowns.cultivation = false
  else dropdowns.qualification = false
  
  dropdowns[key] = !dropdowns[key]
}

function selectOption(key: 'qualification' | 'cultivation', value: string) {
  characterForm.value[key] = value
  dropdowns[key] = false
}

function closeAllDropdowns() {
  dropdowns.qualification = false
  dropdowns.cultivation = false
}

// Custom Confirm Modal State
const confirmState = reactive({
  show: false,
  title: '',
  message: '',
  type: 'warning', // warning | danger
  onConfirm: () => {}
})

function showConfirm(title: string, message: string, onOk: () => void, type: 'warning' | 'danger' = 'warning') {
  confirmState.title = title
  confirmState.message = message
  confirmState.type = type
  confirmState.onConfirm = () => {
    onOk()
    confirmState.show = false
  }
  confirmState.show = true
}

// Soul Burn
const soulBurnMode = ref(false)
const soulBurnPenalties = computed(() => sessionState.value?.soul_burn_penalties || [])

// UI State
const showStatusPanel = ref(false)
const showMobileMenu = ref(false)
const isProcessing = computed(() => sessionState.value?.is_processing || false)
const isLoading = ref(false)
const loadingText = ref('加载中...')
const showRollAnimation = ref(false)
const rollEvent = ref<any>(null)
const isRolling = ref(false) 
const isSaving = ref(false)

// WebSocket
let ws: WebSocket | null = null
const wsReady = ref(false)
const shouldReconnect = ref(true)
const narrativeWindow = ref<HTMLElement>()

// --- Helpers ---
function buildDisplayHistoryFromRecent(recentHistory: any[]): string[] {
  const displayHistory: string[] = []
  recentHistory.forEach((msg: any) => {
    const { role, content } = msg
    if (role === 'user') {
      const userText = content === 'start_trial' ? '> 开始试炼' : `> ${content}`
      displayHistory.push(userText)
    } else if (role === 'assistant') {
      let narrative = ''
      try {
        let jsonContent = content
        if (jsonContent.startsWith('\n\n{')) jsonContent = jsonContent.slice(2)
        const parsed = JSON.parse(jsonContent)
        narrative = parsed.narrative || ''
        narrative = narrative.replace(/\\n/g, '\n').trim()
        if (narrative) {
            displayHistory.push(narrative)
            return
        }
      } catch (e) {}

       const dollarMatch = content.match(/\$\s*(.*?)\s*\$/s)
      if (dollarMatch && dollarMatch[1]) {
        narrative = dollarMatch[1]
      } else {
        const parts = content.split('@\n{', 1)
        narrative = parts.length > 0 ? parts[0].trim() : content
      }
      narrative = narrative.replace(/\\n/g, '\n').trim()
      if (narrative) displayHistory.push(narrative)
    }
  })
  return displayHistory
}

// --- API Calls ---
async function loadAvailableMods() {
  try {
    isLoading.value = true
    const response = await fetch('/api/game/mods', {
      headers: { 'Authorization': `Bearer ${authStore.token}` }
    })
    if (response.ok) {
      availableMods.value = await response.json()
    } else {
      showToast('加载游戏列表失败', 'error')
    }
  } catch (error) {
    showToast('网络错误', 'error')
  } finally {
    isLoading.value = false
  }
}

async function selectGame(modId: string) {
  shouldReconnect.value = true
  currentGame.value = modId
  currentModInfo.value = availableMods.value.find(m => m.game_id === modId)
  await initializeGame()
}

async function initializeGame() {
  try {
    isLoading.value = true
    loadingText.value = '正在唤醒世界...'
    const response = await fetch('/api/game/init', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${authStore.token}`
      },
      body: JSON.stringify({ mod_id: currentGame.value })
    })

    if (response.ok) {
      const data = await response.json()
      gameState.value = data.state || data
      const recentHistory = data.recent_history || gameState.value?.recent_history || []
      if (recentHistory.length > 0) {
         if (!gameState.value.display_history || gameState.value.display_history.length === 0) {
          gameState.value.display_history = buildDisplayHistoryFromRecent(recentHistory)
        } else {
          gameState.value.display_history = [
            ...gameState.value.display_history,
             ...buildDisplayHistoryFromRecent(recentHistory)
          ]
        }
      }
      loadingText.value = '连接位面通道...'
      connectWebSocket()
    }
  } catch (error) {
     showToast('初始化失败', 'error')
  } finally {
    isLoading.value = false
  }
}

// WebSocket
function connectWebSocket() {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/api/game/ws?mod_id=${currentGame.value}&token=${authStore.token}`
  
  ws = new WebSocket(wsUrl)
  
  ws.onopen = () => {
    wsReady.value = true
    isLoading.value = false
    showToast('连接已建立', 'success')
  }
  
  ws.onmessage = (event) => {
    try {
      const message = JSON.parse(event.data)
      handleWebSocketMessage(message)
    } catch (e) {}
  }
  
  ws.onerror = () => {
    wsReady.value = false
    showToast('连接中断', 'error')
  }
  
  ws.onclose = () => {
    wsReady.value = false
    if (shouldReconnect.value && currentGame.value) {
      setTimeout(connectWebSocket, 3000)
    }
  }
}

// Stream Handling
const streamingNarrative = ref('')
const isStreaming = ref(false)
const pendingRollResult = ref<string | null>(null)
const secondStageNarrative = ref('')
const isSecondStageStreaming = ref(false)

function handleWebSocketMessage(message: any) {
  isLoading.value = false
  switch (message.type) {
    case 'narrative_chunk':
        if (!isStreaming.value) {
            isStreaming.value = true
            streamingNarrative.value = ''
            if (gameState.value?.display_history) {
                 gameState.value.display_history = [...gameState.value.display_history, streamingNarrative.value]
            }
        }
        streamingNarrative.value += message.data.content
        
        let content = streamingNarrative.value
        // Clean JSON
        const jsonMark = content.indexOf('```json')
        if (jsonMark >= 0) content = content.substring(0, jsonMark).trim()
        const codeMark = content.indexOf('```')
        if (codeMark >= 0) content = content.substring(0, codeMark).trim()

        const rollMatch = content.match(/【判定结果：(成功|失败)】/)
        if (rollMatch && !pendingRollResult.value) {
            pendingRollResult.value = content
            showDiceRollAnimation(rollMatch[1] === '成功')
            return
        }

        if (gameState.value?.display_history) {
            const lastIdx = gameState.value.display_history.length - 1
            gameState.value.display_history = [
                ...gameState.value.display_history.slice(0, lastIdx),
                content
            ]
        }
        nextTick(scrollToBottom)
        break
    case 'full_state':
        isStreaming.value = false
        isSecondStageStreaming.value = false
        isRolling.value = false
        const history = gameState.value?.display_history || []
        gameState.value = { ...message.data, display_history: history }
        streamingNarrative.value = ''
        secondStageNarrative.value = ''
        pendingRollResult.value = null
        nextTick(scrollToBottom)
        break
    case 'roll_event':
        isRolling.value = true
        showDiceRollAnimation(message.data.success, message.data)
        break
    case 'roll_result':
        if (gameState.value?.display_history) {
            gameState.value.display_history = [...gameState.value.display_history, message.data.content]
            nextTick(scrollToBottom)
        }
        break
    case 'error':
        showToast(message.detail || '错误', 'error')
        break
  }
}

function showDiceRollAnimation(success: boolean, rollData?: any) {
  showRollAnimation.value = true
  
  if (rollData) {
    rollEvent.value = {
      description: rollData.description || '判定中...',
      target: rollData.target,
      result: rollData.result,
      success: success
    }
  } else {
    rollEvent.value = { description: '判定中...', success }
  }

  setTimeout(() => {
     if (rollEvent.value) rollEvent.value.description = success ? '判定成功！' : '判定失败！'
  }, 1500)

  setTimeout(() => {
    showRollAnimation.value = false
    rollEvent.value = null
    
    if (pendingRollResult.value && gameState.value) {
        const lastIdx = gameState.value.display_history.length - 1
         gameState.value.display_history = [
            ...gameState.value.display_history.slice(0, lastIdx),
            pendingRollResult.value
        ]
        nextTick(scrollToBottom)
    }
    pendingRollResult.value = null
  }, 3000)
}

function forceContinueGame() {
  if (gameState.value?.state) {
    gameState.value.state.daily_success_achieved = false
    if (!gameState.value.state.is_in_trial) gameState.value.state.is_in_trial = false
  }
  showToast('已强制继续', 'success')
}

async function startTrial() {
  if (!ws || !wsReady.value) {
    showToast('连接未就绪', 'error')
    return
  }
  showCharacterCreation.value = true
}

function useRandomCharacter() {
  characterForm.value = {
    name: '', gender: '', qualification: '', cultivation: '', spiritStones: 0, background: ''
  }
  confirmCharacterCreation()
}

async function confirmCharacterCreation() {
  if (characterForm.value.name) {
    if (characterForm.value.name.length < 2 || characterForm.value.name.length > 4) {
      showToast('姓名需2-4字', 'error')
      return
    }
  }
  
  showCharacterCreation.value = false
  if (gameState.value?.state) gameState.value.state.is_processing = true
  
  if (gameState.value?.display_history) {
    gameState.value.display_history = [...gameState.value.display_history, '> 开始试炼']
    nextTick(scrollToBottom)
  }

  isLoading.value = true
  loadingText.value = '命运齿轮开始转动...'
  
  const message: any = { action: 'start_trial' }
  // Build custom attributes...
  const attrs: any = {}
  if (characterForm.value.name) attrs.姓名 = characterForm.value.name
  if (characterForm.value.gender) attrs.性别 = characterForm.value.gender
  if (characterForm.value.qualification) attrs.资质 = characterForm.value.qualification
  if (characterForm.value.cultivation) attrs.修为 = characterForm.value.cultivation
  if (characterForm.value.spiritStones > 0) attrs.元石 = characterForm.value.spiritStones
  if (characterForm.value.background) attrs.出身 = characterForm.value.background

  if (Object.keys(attrs).length > 0) message.custom_attributes = attrs
  
  if (ws && wsReady.value) ws.send(JSON.stringify(message))
  
  // Reset
   characterForm.value = {
    name: '', gender: '男', qualification: '丙等资质', cultivation: '一转初阶', spiritStones: 50, background: ''
  }
}

function sendAction() {
  if (!userInput.value.trim()) return
  if (ws && wsReady.value) {
    let action = userInput.value.trim()
    let customAttrs = {}
    if (soulBurnMode.value) {
      customAttrs = { soul_burn_mode: true, action_content: action }
      action = `[SOUL_BURN] ${action}`
    }

    if (gameState.value?.state) gameState.value.state.is_processing = true
    if (gameState.value?.display_history) {
        gameState.value.display_history = [...gameState.value.display_history, `> ${userInput.value.trim()}`]
        nextTick(scrollToBottom)
    }

    ws.send(JSON.stringify({ action, custom_attributes: customAttrs }))
    userInput.value = ''
  }
}

function toggleSoulBurnMode() {
  if (!soulBurnMode.value) {
    showConfirm(
      '开启燃魂', 
      '🔥 燃魂爆运将透支未来以换取当下的极致气运。此操作不可逆，确定要开启吗？',
      () => {
        soulBurnMode.value = true
        showToast('燃魂模式已开启', 'warning')
      },
      'danger'
    )
  } else {
    soulBurnMode.value = false
    showToast('燃魂模式已关闭')
  }
}

async function saveGame() {
    // Save logic implementation (same as before but simplified)
}

function showRestartConfirm() {
  showConfirm(
    '重入轮回', 
    '🔄 确定要结束当前的命运，重新开始吗？所有的进度和机缘都将化为泡影。', 
    restartOpportunities,
    'warning'
  )
}

async function restartOpportunities() {
    // Restart logic
    isLoading.value = true
    try {
        const res = await fetch('/api/game/restart-opportunities', {
             method: 'POST',
             headers: { 'Authorization': `Bearer ${authStore.token}`, 'Content-Type': 'application/json' },
             body: JSON.stringify({ mod_id: currentGame.value })
        })
        if (res.ok) {
            shouldReconnect.value = false
            if(ws) ws.close()
            currentGame.value = ''
            gameState.value = null
            await loadAvailableMods()
            showToast('轮回已重置', 'success')
        }
    } catch(e) {
        showToast('重置失败', 'error')
    } finally {
        isLoading.value = false
    }
}

async function logout() {
    authStore.logout()
    router.push('/login')
}

// UI Helpers
function toggleMobileMenu() { showMobileMenu.value = !showMobileMenu.value }
function closeMobileMenu() { showMobileMenu.value = false }
function toggleStatusPanel_Mobile() { 
    showStatusPanel.value = !showStatusPanel.value 
    showMobileMenu.value = false
}
function closeStatusPanel() { showStatusPanel.value = false }
function handleMobileRestart() { showRestartConfirm(); closeMobileMenu() }
function handleMobileLogout() { logout() }

function getStartButtonText() {
    if(!wsReady.value) return '连接中...'
    return sessionState.value?.current_life ? '开启下一次试炼' : '踏入轮回' // Renamed nicely
}

function scrollToBottom() {
    if (narrativeWindow.value) narrativeWindow.value.scrollTop = narrativeWindow.value.scrollHeight
}

function getBlockClass(text: string) {
    if (text.startsWith('> ')) return 'user-message'
    return 'narrative-message'
}

function renderMarkdown(text: string) {
    try {
        return marked.parse(text.replace(/\\n/g, '\n'), { breaks: true, gfm: true })
    } catch (e) { return text }
}

// --- Formatting Helpers (Kept generic logic but clean HTML) ---
function formatKey(key: string) { return key }

// 递归格式化值
function formatValue(value: any): string {
    if (value === null || value === undefined) return '<span class="mute">无</span>'
    
    // 数组处理
    if (Array.isArray(value)) {
        if (value.length === 0) return '<span class="mute">无</span>'
        const items = value.map(item => {
             if (typeof item === 'object') {
                 // 如果是对象，简化显示或递归
                 return Object.keys(item).length > 0 ? formatValue(item) : String(item)
             }
             return String(item)
        }).join('<span class="separator">, </span>')
        // 如果内容太长，包裹在div中
        return `<div class="array-content">${items}</div>`
    }
    
    // 对象处理
    if (typeof value === 'object') {
        const keys = Object.keys(value)
        if (keys.length === 0) return '<span class="mute">无</span>'
        
        let html = '<div class="nested-obj">'
        for (const k of keys) {
            const v = value[k]
            if (v !== null && v !== undefined && v !== '') {
                 html += `<div class="nested-row">
                    <span class="key">${k}:</span> 
                    <span class="val">${formatValue(v)}</span>
                 </div>`
            }
        }
        html += '</div>'
        return html
    }
    
    return String(value)
}

onMounted(() => {
  loadAvailableMods()
  window.addEventListener('click', closeAllDropdowns)
})

onUnmounted(() => {
  shouldReconnect.value = false
  if(ws) ws.close()
  window.removeEventListener('click', closeAllDropdowns)
})
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Noto+Serif+SC:wght@400;500;700&display=swap');

/* ========== GLOBAL RESET & SCROLLBAR ========== */
* {
  box-sizing: border-box;
}

::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}
::-webkit-scrollbar-track { 
  background: rgba(20, 15, 10, 0.6); 
  border-radius: 4px;
}
::-webkit-scrollbar-thumb { 
  background: linear-gradient(180deg, #6b5344 0%, #4a3828 100%); 
  border-radius: 4px;
  border: 1px solid rgba(212, 165, 116, 0.2);
}
::-webkit-scrollbar-thumb:hover { 
  background: linear-gradient(180deg, #8b6b4e 0%, #5c4d3c 100%); 
}

.game-container {
  height: 100vh;
  width: 100vw;
  background: 
    radial-gradient(ellipse at top left, rgba(139, 107, 78, 0.08) 0%, transparent 50%),
    radial-gradient(ellipse at bottom right, rgba(212, 165, 116, 0.05) 0%, transparent 50%),
    linear-gradient(180deg, #0a0806 0%, #12100e 50%, #0d0b09 100%);
  color: #c8b8a8;
  font-family: 'Noto Serif SC', serif;
  overflow: hidden;
  position: relative;
  display: flex;
  flex-direction: column;
}

/* ========== BACKGROUND EFFECTS ========== */
.stars, .twinkling {
  position: absolute;
  top: 0; left: 0; width: 100%; height: 100%;
  pointer-events: none;
}
.stars {
  background: transparent;
  background-image: 
    radial-gradient(1px 1px at 20px 30px, rgba(255,255,255,0.3), transparent),
    radial-gradient(1px 1px at 40px 70px, rgba(255,255,255,0.2), transparent),
    radial-gradient(1px 1px at 50px 160px, rgba(255,255,255,0.3), transparent),
    radial-gradient(1px 1px at 90px 40px, rgba(255,255,255,0.2), transparent),
    radial-gradient(1px 1px at 130px 80px, rgba(255,255,255,0.3), transparent),
    radial-gradient(1px 1px at 160px 120px, rgba(255,255,255,0.2), transparent);
  background-size: 200px 200px;
  z-index: 0;
  animation: twinkle 8s ease-in-out infinite;
}
.twinkling {
  background: transparent;
  z-index: 1;
  opacity: 0.15;
}

@keyframes twinkle {
  0%, 100% { opacity: 0.6; }
  50% { opacity: 1; }
}

/* ========== GAME SELECTION ========== */
.game-selection {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  z-index: 10;
  padding: 40px 20px;
  overflow-y: auto;
}

.selection-header { 
  text-align: center; 
  margin-bottom: 50px; 
}
.selection-header h1 { 
  font-size: 2.8rem; 
  color: #e8dcd0; 
  text-shadow: 0 2px 20px rgba(0,0,0,0.8), 0 0 40px rgba(212, 165, 116, 0.15);
  margin: 0 0 15px 0;
  font-weight: 500;
  letter-spacing: 4px;
}
.selection-header p {
  color: #8b7b6b;
  font-size: 1.1rem;
  margin: 0;
  letter-spacing: 2px;
}

.game-list { 
  display: flex; 
  gap: 30px; 
  flex-wrap: wrap; 
  justify-content: center; 
}

.game-card {
  background: linear-gradient(145deg, rgba(30, 25, 20, 0.9) 0%, rgba(20, 16, 12, 0.95) 100%);
  border: 1px solid rgba(92, 77, 60, 0.6);
  padding: 30px 25px;
  width: 320px;
  cursor: pointer;
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  border-radius: 12px;
  position: relative;
  overflow: hidden;
}
.game-card::before {
  content: '';
  position: absolute;
  top: 0; left: 0; right: 0;
  height: 3px;
  background: linear-gradient(90deg, transparent, #d4a574, transparent);
  opacity: 0;
  transition: opacity 0.3s;
}
.game-card:hover::before { opacity: 1; }
.game-card:hover { 
  transform: translateY(-8px); 
  border-color: rgba(212, 165, 116, 0.6);
  box-shadow: 
    0 20px 40px rgba(0,0,0,0.4),
    0 0 30px rgba(212, 165, 116, 0.1),
    inset 0 1px 0 rgba(255,255,255,0.05);
}
.game-card-icon {
  margin-bottom: 15px;
}
.game-card h3 { 
  color: #d4a574; 
  margin: 15px 0 12px; 
  font-size: 1.4rem;
  font-weight: 500;
}
.game-card p {
  color: #9a8a7a;
  font-size: 0.95rem;
  line-height: 1.6;
  margin: 0;
}

/* ========== GAME PLAY LAYOUT ========== */
.game-play {
  flex: 1;
  display: flex;
  flex-direction: column;
  z-index: 10;
  height: 100%;
  overflow: hidden;
}

/* ========== HEADER ========== */
.game-header {
  height: 64px;
  flex-shrink: 0;
  background: linear-gradient(180deg, rgba(18, 14, 10, 0.98) 0%, rgba(12, 10, 8, 0.98) 100%);
  border-bottom: 1px solid rgba(92, 77, 60, 0.4);
  display: flex;
  align-items: center;
  padding: 0 24px;
  z-index: 20;
  box-shadow: 0 2px 20px rgba(0,0,0,0.3);
}
.header-content { 
  width: 100%; 
  display: flex; 
  justify-content: space-between; 
  align-items: center; 
}
.game-title { 
  display: flex; 
  align-items: center; 
}
.game-title h2 { 
  margin: 0; 
  font-size: 1.25rem; 
  color: #e8dcd0; 
  font-weight: 500;
  letter-spacing: 1px;
}
.mod-id { 
  font-size: 0.75rem; 
  color: #6b5b4b; 
  margin-left: 12px;
  padding: 3px 10px;
  background: rgba(0,0,0,0.3);
  border-radius: 12px;
  letter-spacing: 0.5px;
}

.header-actions { display: flex; align-items: center; gap: 8px; }
.pc-actions { display: flex; gap: 12px; align-items: center; }

.btn-icon {
  background: rgba(40, 32, 24, 0.6); 
  border: 1px solid rgba(92, 77, 60, 0.5); 
  color: #9a8a7a; 
  cursor: pointer; 
  padding: 8px 16px;
  border-radius: 6px;
  display: flex; 
  align-items: center; 
  gap: 8px; 
  font-family: inherit; 
  transition: all 0.25s ease;
  font-size: 0.9rem;
}
.btn-icon:hover { 
  color: #d4a574; 
  border-color: rgba(212, 165, 116, 0.5); 
  background: rgba(58, 46, 34, 0.8);
  transform: translateY(-1px);
}
.btn-icon.danger:hover { 
  color: #ef5350; 
  border-color: rgba(239, 83, 80, 0.5); 
  background: rgba(239, 83, 80, 0.1); 
}
.btn-label { font-size: 0.85rem; }
.btn-icon-img {
  width: 18px;
  height: 18px;
  object-fit: contain;
  opacity: 0.85;
  transition: opacity 0.2s;
}
.btn-icon:hover .btn-icon-img {
  opacity: 1;
}
.menu-toggle { 
  display: none; 
  background: none; 
  border: none; 
  color: #d4a574; 
  font-size: 1.6rem; 
  cursor: pointer;
  padding: 8px;
}

/* ========== MOBILE MENU ========== */
.mobile-menu-overlay {
  display: none;
  position: fixed; top: 0; left: 0; width: 100%; height: 100%;
  background: rgba(0,0,0,0.85); 
  backdrop-filter: blur(8px);
  z-index: 100;
  align-items: center;
  justify-content: center;
}
.mobile-menu-overlay.active { display: flex; }

.mobile-menu {
  background: linear-gradient(145deg, rgba(30, 25, 20, 0.98) 0%, rgba(18, 14, 10, 0.98) 100%);
  border: 1px solid rgba(92, 77, 60, 0.5);
  padding: 0;
  border-radius: 16px;
  width: 85%;
  max-width: 320px;
  box-shadow: 0 25px 50px rgba(0,0,0,0.6);
  overflow: hidden;
}
.mobile-menu-header { 
  display: flex; 
  justify-content: space-between; 
  align-items: center;
  padding: 20px 24px;
  background: rgba(0,0,0,0.2);
  border-bottom: 1px solid rgba(92, 77, 60, 0.3);
}
.mobile-menu-header h3 { margin: 0; color: #d4a574; font-size: 1.1rem; letter-spacing: 1px; }
.close-btn { 
  background: none; 
  border: none; 
  color: #6b5b4b; 
  font-size: 1.8rem; 
  cursor: pointer;
  transition: color 0.2s;
}
.close-btn:hover { color: #d4a574; }

.mobile-menu-items { 
  display: flex; 
  flex-direction: column; 
  padding: 16px;
  gap: 8px;
}
.mobile-menu-items button {
  padding: 16px 20px;
  background: rgba(40, 32, 24, 0.4);
  border: 1px solid rgba(92, 77, 60, 0.3);
  color: #c8b8a8;
  cursor: pointer;
  text-align: left;
  border-radius: 10px;
  transition: all 0.25s;
  font-family: inherit;
  font-size: 1rem;
  display: flex;
  align-items: center;
  gap: 12px;
}
.mobile-menu-items button:hover { 
  background: rgba(58, 46, 34, 0.6); 
  border-color: rgba(212, 165, 116, 0.4);
  transform: translateX(4px);
}
.mobile-menu-items button.danger { 
  color: #ef5350; 
  border-color: rgba(239, 83, 80, 0.3); 
}
.mobile-menu-items button.danger:hover {
  background: rgba(239, 83, 80, 0.1);
  border-color: rgba(239, 83, 80, 0.5);
}
.menu-icon-img {
  width: 20px;
  height: 20px;
  object-fit: contain;
  opacity: 0.8;
}

/* ========== CONTENT WRAPPER ========== */
.game-content-wrapper {
  flex: 1;
  display: grid;
  grid-template-columns: 280px 1fr;
  overflow: hidden;
  height: calc(100vh - 64px);
}

/* ========== SIDEBAR ========== */
.status-panel {
  background: linear-gradient(180deg, rgba(16, 12, 10, 0.95) 0%, rgba(12, 10, 8, 0.98) 100%);
  border-right: 1px solid rgba(92, 77, 60, 0.25);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  height: 100%;
}
.panel-header {
  padding: 16px 20px;
  background: rgba(0,0,0,0.25);
  border-bottom: 1px solid rgba(92, 77, 60, 0.2);
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
}
.panel-title {
  display: flex;
  align-items: center;
  gap: 10px;
}
.panel-icon-img {
  width: 22px;
  height: 22px;
  object-fit: contain;
  opacity: 0.85;
}
.panel-header h3 {
  margin: 0;
  font-size: 1rem;
  color: #a68b6c;
  font-weight: 500;
  letter-spacing: 2px;
}
.close-panel-btn { 
  display: none; 
  background: none; 
  border: none; 
  color: #6b5b4b; 
  font-size: 1.6rem; 
  cursor: pointer;
  transition: color 0.2s;
}
.close-panel-btn:hover { color: #d4a574; }

.panel-content { 
  flex: 1; 
  overflow-y: auto; 
  padding: 16px;
}

.character-status {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.status-block { 
  background: linear-gradient(135deg, rgba(30, 25, 20, 0.6) 0%, rgba(20, 16, 12, 0.8) 100%);
  border: 1px solid rgba(92, 77, 60, 0.2);
  border-radius: 8px;
  padding: 14px 16px;
  transition: all 0.25s ease;
}
.status-block:hover {
  border-color: rgba(92, 77, 60, 0.4);
  background: linear-gradient(135deg, rgba(35, 28, 22, 0.7) 0%, rgba(25, 20, 15, 0.85) 100%);
}
.status-block-header { 
  font-weight: 600;
  color: #b89b7a;
  font-size: 0.8rem;
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 1.5px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.status-block-header::before {
  content: '';
  width: 3px;
  height: 12px;
  background: linear-gradient(180deg, #d4a574, #8b6b4e);
  border-radius: 2px;
}
.status-block-content { 
  font-size: 0.9rem; 
  color: #c8b8a8; 
  line-height: 1.6;
}

.nested-obj { 
  padding-left: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.nested-row { 
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  padding: 4px 0;
  border-bottom: 1px solid rgba(92, 77, 60, 0.1);
}
.nested-row:last-child { border-bottom: none; }
.nested-row .key { 
  color: #8b7b6b;
  font-weight: 500;
  font-size: 0.85rem;
}
.nested-row .val { 
  color: #c8b8a8;
  font-size: 0.85rem;
}
.array-content {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.separator { color: #5c4d3c; }
.mute { color: #5d4d3d; font-style: italic; }

.no-character {
  text-align: center;
  padding: 40px 20px;
  color: #6b5b4b;
}
.no-character p {
  margin: 0;
  font-style: italic;
}

.soul-burn-section { 
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid rgba(239, 83, 80, 0.2);
}
.soul-burn-title {
  color: #ef5350;
  font-size: 0.9rem;
  margin: 0 0 12px 0;
}
.penalty-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: rgba(239, 83, 80, 0.1);
  border-radius: 6px;
  margin-bottom: 8px;
}
.penalty-icon { font-size: 0.9rem; }
.penalty-text { color: #ffcdd2; font-size: 0.85rem; }

/* ========== MAIN AREA ========== */
.game-main {
  display: flex;
  flex-direction: column;
  background: linear-gradient(135deg, rgba(8, 6, 4, 0.4) 0%, rgba(12, 10, 8, 0.6) 100%);
  position: relative;
  overflow: hidden;
  height: 100%;
}

/* ========== NARRATIVE ========== */
.narrative-window {
  flex: 1;
  overflow-y: auto;
  padding: 24px 40px;
  scroll-behavior: smooth;
}

.narrative-block { 
  margin-bottom: 16px;
  line-height: 1.75;
  color: #d8ccc0;
  max-width: 860px;
  margin-left: auto;
  margin-right: auto;
  font-size: 1rem;
  letter-spacing: 0.2px;
}

/* ========== MARKDOWN 样式 ========== */

/* 大标题 */
.narrative-block :deep(h1) {
  color: #e8c89a;
  font-size: 1.5rem;
  font-weight: 600;
  text-align: center;
  margin: 1em 0 0.8em;
  padding-bottom: 10px;
  letter-spacing: 3px;
  position: relative;
}
.narrative-block :deep(h1)::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 80px;
  height: 1px;
  background: linear-gradient(90deg, transparent, #d4a574, transparent);
}

/* 二级标题 */
.narrative-block :deep(h2) {
  color: #d4a574;
  font-size: 1.2rem;
  font-weight: 600;
  margin: 1em 0 0.5em;
  padding-left: 12px;
  border-left: 3px solid #d4a574;
  letter-spacing: 1px;
}

/* 三级标题 */
.narrative-block :deep(h3) {
  color: #c8a882;
  font-size: 1.1rem;
  font-weight: 600;
  margin: 0.8em 0 0.4em;
}

/* 四级标题 */
.narrative-block :deep(h4) {
  color: #b89b7a;
  font-size: 1rem;
  font-weight: 600;
  margin: 0.6em 0 0.3em;
}

/* 段落 */
.narrative-block :deep(p) {
  margin: 0 0 0.6em;
  text-align: justify;
  text-indent: 0;
  line-height: 1.75;
}

/* 无序列表 */
.narrative-block :deep(ul) {
  list-style: none;
  padding-left: 0;
  margin: 0.3em 0;
}

.narrative-block :deep(ul > li) {
  position: relative;
  padding-left: 20px;
  margin-bottom: 0.4em;
  line-height: 1.7;
  display: list-item;
  list-style: none;
}

.narrative-block :deep(ul > li)::before {
  content: '◆';
  position: absolute;
  left: 2px;
  top: 4px;
  color: #a68b6c;
  font-size: 9px;
}

/* 有序列表 */
.narrative-block :deep(ol) {
  padding-left: 1.2em;
  margin: 0.3em 0;
}

.narrative-block :deep(ol > li) {
  margin-bottom: 0.3em;
  line-height: 1.65;
  padding-left: 4px;
}

.narrative-block :deep(ol > li)::marker {
  color: #d4a574;
  font-weight: 600;
}

/* 嵌套列表 */
.narrative-block :deep(li ul),
.narrative-block :deep(li ol) {
  margin: 2px 0;
}

.narrative-block :deep(li li) {
  margin-bottom: 2px;
}

/* 加粗 */
.narrative-block :deep(strong) {
  color: #f0e0d0;
  font-weight: 600;
}

/* 斜体 */
.narrative-block :deep(em) {
  color: #c8b090;
  font-style: italic;
}

/* 引用块 - 诗词样式 */
.narrative-block :deep(blockquote) {
  border: none;
  margin: 0.2em 0 0.5em 1em;
  padding: 8px 0 8px 14px;
  background: transparent;
  border-left: 2px solid rgba(139, 107, 78, 0.4);
  font-style: normal;
  color: #b8a890;
  line-height: 1.7;
}

.narrative-block :deep(blockquote p) {
  margin: 0;
  text-indent: 0;
  line-height: 1.7;
}

.narrative-block :deep(blockquote p + p) {
  margin-top: 0.2em;
}

/* 分隔线 */
.narrative-block :deep(hr) {
  border: none;
  height: 1px;
  background: linear-gradient(90deg, transparent, #5c4d3c, transparent);
  margin: 1.2em 0;
}

/* 行内代码 */
.narrative-block :deep(code) {
  background: rgba(212, 165, 116, 0.1);
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Consolas', monospace;
  font-size: 0.9em;
  color: #e8c89a;
}

/* 代码块 */
.narrative-block :deep(pre) {
  background: rgba(10, 8, 6, 0.8);
  padding: 12px 16px;
  border-radius: 6px;
  overflow-x: auto;
  margin: 0.8em 0;
  border: 1px solid rgba(92, 77, 60, 0.2);
}

.narrative-block :deep(pre code) {
  background: none;
  padding: 0;
  font-size: 0.85rem;
  line-height: 1.5;
}

/* 链接 */
.narrative-block :deep(a) {
  color: #d4a574;
  text-decoration: none;
  border-bottom: 1px dashed rgba(212, 165, 116, 0.4);
  transition: all 0.2s;
}

.narrative-block :deep(a:hover) {
  color: #e8c89a;
  border-bottom-color: #e8c89a;
}

/* 表格 */
.narrative-block :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 1.5em 0;
}

.narrative-block :deep(th),
.narrative-block :deep(td) {
  padding: 12px 16px;
  border: 1px solid rgba(92, 77, 60, 0.3);
  text-align: left;
}

.narrative-block :deep(th) {
  background: rgba(30, 25, 20, 0.6);
  color: #d4a574;
  font-weight: 600;
}

.narrative-block :deep(tr:nth-child(even)) {
  background: rgba(20, 16, 12, 0.4);
}

/* 图片 */
.narrative-block :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: 8px;
  margin: 1em 0;
}

.user-message { 
  color: #9a8a7a;
  text-align: right;
  font-style: italic;
  padding: 8px 16px;
  margin: 16px auto 12px;
  max-width: 860px;
  background: linear-gradient(90deg, transparent 0%, rgba(30, 25, 20, 0.3) 100%);
  border-right: 2px solid rgba(212, 165, 116, 0.3);
  border-radius: 6px 0 0 6px;
  font-size: 0.9rem;
}
.user-message :deep(p) {
  text-indent: 0;
  text-align: right;
  margin: 0;
}

.narrative-message { 
  text-align: justify;
  padding: 16px 20px;
  background: rgba(15, 12, 10, 0.3);
  border-radius: 8px;
  border: 1px solid rgba(92, 77, 60, 0.1);
}

.input-loading {
  text-align: center;
  padding: 16px;
  color: #8b6b4e;
  font-size: 1.3rem;
}
.input-loading .dot {
  animation: blink 1.4s infinite both;
}
.input-loading .dot:nth-child(2) { animation-delay: 0.2s; }
.input-loading .dot:nth-child(3) { animation-delay: 0.4s; }
@keyframes blink {
  0%, 80%, 100% { opacity: 0; }
  40% { opacity: 1; }
}

/* ========== INPUT AREA ========== */
.input-container {
  flex-shrink: 0;
  padding: 20px 40px;
  background: linear-gradient(180deg, rgba(14, 11, 9, 0.95) 0%, rgba(10, 8, 6, 0.98) 100%);
  border-top: 1px solid rgba(92, 77, 60, 0.3);
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100px;
  width: 100%;
  z-index: 20;
  box-sizing: border-box;
}

.start-game-wrapper, .action-input-wrapper { 
  width: 100%; 
  max-width: 860px; 
}
.action-input-wrapper { 
  display: flex; 
  gap: 16px; 
  align-items: center; 
}
.input-group { 
  flex: 1; 
  display: flex; 
  position: relative;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0,0,0,0.3);
}

.adventure-input {
  width: 100%;
  height: 54px;
  background: rgba(20, 16, 12, 0.9);
  border: 1px solid rgba(92, 77, 60, 0.4);
  border-right: none;
  color: #e8dcd0;
  padding: 0 24px;
  font-family: inherit;
  font-size: 1.05rem;
  line-height: 54px;
  outline: none;
  transition: all 0.3s ease;
  border-radius: 8px 0 0 8px;
  box-sizing: border-box;
}
.adventure-input::placeholder {
  color: #6b5b4b;
  font-style: italic;
}
.adventure-input:focus { 
  border-color: rgba(212, 165, 116, 0.5);
  background: rgba(25, 20, 15, 0.95);
  box-shadow: inset 0 0 20px rgba(0,0,0,0.2);
}
.adventure-input.soul-burn-active { 
  border-color: rgba(211, 47, 47, 0.5);
  color: #ffcdd2;
  background: rgba(40, 15, 15, 0.9);
}

.btn-adventure {
  height: 54px;
  background: linear-gradient(135deg, #3a2e22 0%, #2a2018 100%);
  border: 1px solid rgba(92, 77, 60, 0.5);
  color: #d4a574;
  padding: 0 32px;
  cursor: pointer;
  font-family: inherit;
  font-weight: 600;
  font-size: 1.05rem;
  white-space: nowrap;
  transition: all 0.25s ease;
  border-radius: 0 8px 8px 0;
  letter-spacing: 1px;
  box-sizing: border-box;
}
.btn-adventure.large { 
  border-radius: 8px;
  width: 100%;
  height: 60px;
  padding: 0 20px;
  font-size: 1.2rem;
  margin-top: 0;
  background: linear-gradient(135deg, #3a2e22 0%, #26201a 100%);
  box-shadow: 0 4px 20px rgba(0,0,0,0.4);
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
}
.btn-adventure.large::before {
  content: '';
  position: absolute;
  top: 0; left: -100%;
  width: 100%; height: 100%;
  background: linear-gradient(90deg, transparent, rgba(212, 165, 116, 0.1), transparent);
  transition: left 0.5s ease;
}
.btn-adventure.large:hover::before {
  left: 100%;
}
.btn-adventure:hover:not(:disabled) { 
  border-color: rgba(212, 165, 116, 0.6);
  color: #e8c89a;
  background: linear-gradient(135deg, #4a3b2a 0%, #3a2e22 100%);
  transform: translateY(-1px);
}
.btn-adventure:disabled { 
  opacity: 0.4;
  cursor: not-allowed;
}

.btn-soul-burn {
  width: 54px;
  height: 54px;
  background: rgba(20, 16, 12, 0.8);
  border: 1px solid rgba(92, 77, 60, 0.3);
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  transition: all 0.3s ease;
  flex-shrink: 0;
}
.btn-soul-burn:hover { 
  border-color: rgba(211, 47, 47, 0.5);
  background: rgba(211, 47, 47, 0.1);
  transform: scale(1.05);
}
.btn-soul-burn.active { 
  border-color: #d32f2f;
  background: rgba(211, 47, 47, 0.2);
  box-shadow: 0 0 20px rgba(211, 47, 47, 0.3);
  animation: pulse-burn 2s infinite;
}

@keyframes pulse-burn {
  0%, 100% { box-shadow: 0 0 20px rgba(211, 47, 47, 0.3); }
  50% { box-shadow: 0 0 30px rgba(211, 47, 47, 0.5); }
}

.game-ended-message {
  text-align: center;
  padding: 20px;
  color: #a68b6c;
}
.game-ended-message p {
  margin: 0;
  font-size: 1.1rem;
}

.warning-box {
  text-align: center;
  padding: 16px;
  background: rgba(255, 193, 7, 0.1);
  border: 1px solid rgba(255, 193, 7, 0.3);
  border-radius: 8px;
}
.warning-box p { margin: 0 0 10px; color: #ffc107; }
.btn-text {
  background: none;
  border: none;
  color: #d4a574;
  cursor: pointer;
  text-decoration: underline;
  font-family: inherit;
}

/* ========== MODALS ========== */
.modal-overlay {
  position: fixed; 
  top: 0; left: 0; 
  width: 100%; height: 100%;
  background: rgba(0,0,0,0.9);
  backdrop-filter: blur(12px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 200;
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.character-modal, .roll-modal {
  background: linear-gradient(145deg, #1a1614 0%, #0f0d0b 100%);
  border: 1px solid rgba(92, 77, 60, 0.4);
  padding: 0;
  border-radius: 16px;
  box-shadow: 
    0 25px 80px rgba(0,0,0,0.8),
    0 0 0 1px rgba(212, 165, 116, 0.08) inset,
    0 0 60px rgba(212, 165, 116, 0.05);
  max-width: 480px;
  width: 92%;
  display: flex;
  flex-direction: column;
  position: relative;
  overflow: hidden;
  animation: modalSlideUp 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes modalSlideUp {
  from { 
    opacity: 0;
    transform: translateY(30px) scale(0.95);
  }
  to { 
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.modal-header {
  padding: 24px 28px;
  background: linear-gradient(180deg, rgba(30, 25, 20, 0.8) 0%, transparent 100%);
  border-bottom: 1px solid rgba(92, 77, 60, 0.25);
  text-align: center;
  position: relative;
}

.modal-header h2 {
  margin: 0;
  color: #d4a574;
  font-size: 1.4rem;
  letter-spacing: 3px;
  font-weight: 500;
  text-shadow: 0 2px 8px rgba(0,0,0,0.5);
}

.modal-decoration {
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(212, 165, 116, 0.4), transparent);
  margin-top: 16px;
}

.modal-body {
  padding: 28px 32px;
  max-height: 65vh;
  overflow-y: auto;
}

/* ========== FORM STYLES ========== */
.form-group { margin-bottom: 24px; }
.form-group label {
  display: block;
  margin-bottom: 10px;
  color: #9a8a7a;
  font-size: 0.85rem;
  font-weight: 500;
  letter-spacing: 1px;
  text-transform: uppercase;
}

/* Input Effects */
.input-effects { position: relative; }
.input-effects .adventure-input {
  width: 100%;
  background: rgba(20, 16, 12, 0.6);
  border: none;
  border-bottom: 2px solid rgba(92, 77, 60, 0.4);
  color: #e8dcd0;
  padding: 12px 4px;
  font-family: inherit;
  font-size: 1.1rem;
  outline: none;
  transition: all 0.3s;
  border-radius: 0;
}
.focus-line {
  position: absolute;
  bottom: 0; left: 0;
  width: 0; height: 2px;
  background: linear-gradient(90deg, #d4a574, #b89060);
  transition: width 0.3s ease;
  box-shadow: 0 0 12px rgba(212, 165, 116, 0.5);
}
.input-effects .adventure-input:focus + .focus-line { width: 100%; }

/* Gender Selector */
.gender-selector { 
  display: flex;
  gap: 16px;
}
.gender-option {
  flex: 1;
  background: rgba(25, 20, 16, 0.6);
  border: 1px solid rgba(92, 77, 60, 0.3);
  padding: 18px 15px;
  border-radius: 10px;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  transition: all 0.3s ease;
  color: #8b7b6b;
}
.gender-option:hover { 
  background: rgba(40, 32, 24, 0.6);
  border-color: rgba(92, 77, 60, 0.5);
  transform: translateY(-2px);
}
.gender-option.active {
  background: rgba(212, 165, 116, 0.12);
  border-color: rgba(212, 165, 116, 0.5);
  color: #d4a574;
  box-shadow: 
    0 0 20px rgba(212, 165, 116, 0.1) inset,
    0 4px 15px rgba(0,0,0,0.2);
}
.gender-icon { font-size: 2rem; }

/* Custom Select */
.custom-select { 
  position: relative;
  user-select: none;
}
.select-trigger {
  background: rgba(25, 20, 16, 0.6);
  border: 1px solid rgba(92, 77, 60, 0.3);
  padding: 14px 18px;
  border-radius: 8px;
  color: #e8dcd0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
  transition: all 0.3s;
}
.custom-select:hover .select-trigger { 
  border-color: rgba(92, 77, 60, 0.5);
  background: rgba(30, 24, 18, 0.7);
}
.custom-select.open .select-trigger { 
  border-color: rgba(212, 165, 116, 0.5);
  border-bottom-left-radius: 0;
  border-bottom-right-radius: 0;
}
.arrow { 
  color: #8b7b6b;
  font-size: 0.75rem;
  transition: transform 0.3s;
}
.custom-select.open .arrow { transform: rotate(180deg); }

.select-options {
  position: absolute;
  top: 100%; left: 0;
  width: 100%;
  background: linear-gradient(180deg, #1f1a16 0%, #181412 100%);
  border: 1px solid rgba(212, 165, 116, 0.4);
  border-top: none;
  border-radius: 0 0 8px 8px;
  z-index: 100;
  max-height: 220px;
  overflow-y: auto;
  box-shadow: 0 15px 30px rgba(0,0,0,0.5);
}
.select-option {
  padding: 14px 18px;
  cursor: pointer;
  border-bottom: 1px solid rgba(92, 77, 60, 0.15);
  transition: all 0.2s;
}
.select-option:last-child { border-bottom: none; }
.select-option:hover { 
  background: rgba(212, 165, 116, 0.1);
}
.select-option.selected { 
  background: rgba(212, 165, 116, 0.15);
  color: #d4a574;
}
.opt-main { font-size: 1rem; }
.opt-desc { 
  font-size: 0.8rem;
  color: #7b6b5b;
  margin-top: 4px;
}

/* Custom Range */
.range-label { 
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.range-value { 
  color: #d4a574;
  font-weight: 600;
  font-size: 1.1rem;
}
.range-wrapper { 
  margin-top: 14px;
  height: 8px;
  background: rgba(30, 25, 20, 0.8);
  border-radius: 4px;
  position: relative;
}
.adventure-range {
  -webkit-appearance: none;
  width: 100%;
  height: 8px;
  background: transparent;
  outline: none;
  padding: 0;
  margin: 0;
  position: relative;
  z-index: 5;
  background-color: rgba(30, 25, 20, 0.8);
  border-radius: 4px;
  background-image: linear-gradient(90deg, #d4a574, #b89060);
  background-repeat: no-repeat;
}
.adventure-range::-webkit-slider-thumb {
  -webkit-appearance: none;
  height: 22px;
  width: 22px;
  border-radius: 50%;
  background: linear-gradient(135deg, #1a1614 0%, #0f0d0b 100%);
  border: 2px solid #d4a574;
  cursor: pointer;
  box-shadow: 0 2px 10px rgba(0,0,0,0.4);
  transition: all 0.2s;
}
.adventure-range::-webkit-slider-thumb:hover { 
  transform: scale(1.15);
  box-shadow: 0 0 15px rgba(212, 165, 116, 0.4);
}

/* Textarea */
.textarea-wrapper {
  background: rgba(20, 16, 12, 0.5);
  padding: 4px;
  border: 1px solid rgba(92, 77, 60, 0.3);
  border-radius: 8px;
  transition: all 0.3s;
}
.textarea-wrapper:focus-within { 
  border-color: rgba(212, 165, 116, 0.5);
  background: rgba(25, 20, 15, 0.6);
}
.adventure-textarea {
  width: 100%;
  background: transparent;
  border: none;
  color: #e8dcd0;
  padding: 12px 14px;
  font-family: inherit;
  font-size: 1rem;
  resize: none;
  outline: none;
  line-height: 1.6;
}
.adventure-textarea::placeholder { 
  color: #5c4d3c;
  font-style: italic;
}

/* Modal Footer */
.modal-footer {
  padding: 22px 32px;
  background: rgba(0,0,0,0.25);
  border-top: 1px solid rgba(92, 77, 60, 0.2);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.btn-secondary { 
  background: rgba(25, 20, 16, 0.6);
  border: 1px solid rgba(92, 77, 60, 0.4);
  color: #9a8a7a; 
  padding: 12px 22px;
  border-radius: 8px;
  cursor: pointer; 
  font-family: inherit;
  font-size: 0.95rem; 
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 10px;
}
.btn-secondary:hover { 
  border-color: rgba(212, 165, 116, 0.5);
  color: #d4a574;
  background: rgba(58, 46, 34, 0.4);
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(0,0,0,0.25);
}

.confirm-btn {
  background: linear-gradient(135deg, #3a2e22 0%, #26201a 100%);
  border: 1px solid rgba(139, 107, 78, 0.6);
  padding: 14px 32px;
  border-radius: 8px;
  color: #d4a574;
  font-weight: 600;
  letter-spacing: 1.5px;
  box-shadow: 0 4px 15px rgba(0,0,0,0.3);
  transition: all 0.3s ease;
  font-family: inherit;
  font-size: 1rem;
}
.confirm-btn:hover {
  border-color: rgba(212, 165, 116, 0.7);
  color: #e8c89a;
  background: linear-gradient(135deg, #4a3b2a 0%, #3a2e22 100%);
  box-shadow: 0 6px 20px rgba(212, 165, 116, 0.2);
  transform: translateY(-2px);
}

.modal-close-btn {
  position: absolute;
  top: 18px; right: 22px;
  background: none;
  border: none;
  color: #5c4d3c;
  font-size: 1.8rem;
  cursor: pointer;
  transition: all 0.3s ease;
  line-height: 1;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
}
.modal-close-btn:hover { 
  color: #d4a574;
  background: rgba(212, 165, 116, 0.1);
  transform: rotate(90deg);
}

/* ========== ROLL MODAL ========== */
.roll-modal { 
  text-align: center;
  width: 340px;
  padding: 40px 30px;
}
.dice-icon { 
  font-size: 4.5rem;
  margin-bottom: 24px;
  animation: roll 0.8s ease-in-out infinite;
  filter: drop-shadow(0 0 20px rgba(212, 165, 116, 0.3));
}
@keyframes roll {
  0%, 100% { transform: rotate(0deg) scale(1); }
  25% { transform: rotate(-15deg) scale(1.1); }
  75% { transform: rotate(15deg) scale(1.1); }
}
.roll-info h3 {
  color: #d4a574;
  font-size: 1.2rem;
  margin: 0 0 8px;
  font-weight: 500;
}
.roll-target {
  color: #8b7b6b;
  font-size: 0.95rem;
  margin: 0;
}
.roll-result-box {
  margin-top: 24px;
  padding: 20px;
  background: rgba(0,0,0,0.3);
  border-radius: 12px;
}
.roll-outcome { 
  font-size: 1.8rem;
  font-weight: 600;
  margin: 0 0 12px;
  letter-spacing: 2px;
}
.roll-outcome.success { 
  color: #66bb6a;
  text-shadow: 0 0 20px rgba(102, 187, 106, 0.4);
}
.roll-outcome.fail { 
  color: #ef5350;
  text-shadow: 0 0 20px rgba(239, 83, 80, 0.4);
}
.roll-value { 
  font-size: 3rem;
  color: #e8dcd0;
  font-weight: 700;
}

/* ========== TOAST ========== */
.toast {
  position: fixed;
  top: 32px;
  left: 50%;
  transform: translateX(-50%);
  padding: 14px 28px;
  background: rgba(18, 14, 12, 0.95);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(212, 165, 116, 0.25);
  border-radius: 12px;
  color: #e8dcd0;
  z-index: 1000;
  box-shadow: 0 12px 40px rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  gap: 14px;
  animation: toastSlide 0.5s cubic-bezier(0.16, 1, 0.3, 1);
  font-size: 0.95rem;
  pointer-events: none;
}
.toast-icon { font-size: 1.3rem; }
.toast-message { letter-spacing: 0.5px; }
.toast.info { border-color: rgba(212, 165, 116, 0.4); }
.toast.error { 
  border-color: rgba(239, 83, 80, 0.5);
  background: rgba(30, 12, 12, 0.95);
  box-shadow: 0 12px 40px rgba(239, 83, 80, 0.15);
}
.toast.success { 
  border-color: rgba(102, 187, 106, 0.5);
  background: rgba(12, 25, 15, 0.95);
  box-shadow: 0 12px 40px rgba(102, 187, 106, 0.15);
}
.toast.warning { 
  border-color: rgba(255, 193, 7, 0.5);
  background: rgba(30, 25, 12, 0.95);
  box-shadow: 0 12px 40px rgba(255, 193, 7, 0.15);
}

@keyframes toastSlide {
  from { 
    top: -20px;
    opacity: 0;
    transform: translateX(-50%) scale(0.9);
  }
  to { 
    top: 32px;
    opacity: 1;
    transform: translateX(-50%) scale(1);
  }
}

/* ========== CONFIRM MODAL ========== */
.confirm-modal {
  background: linear-gradient(145deg, #1a1614 0%, #0f0d0b 100%);
  border: 1px solid rgba(92, 77, 60, 0.4);
  padding: 0;
  border-radius: 16px;
  width: 92%;
  max-width: 420px;
  box-shadow: 0 25px 80px rgba(0,0,0,0.8);
  overflow: hidden;
  animation: modalSlideUp 0.4s cubic-bezier(0.16, 1, 0.3, 1);
  display: flex;
  flex-direction: column;
}
.confirm-header {
  padding: 20px 24px;
  background: linear-gradient(180deg, rgba(30, 25, 20, 0.6) 0%, transparent 100%);
  border-bottom: 1px solid rgba(92, 77, 60, 0.25);
}
.confirm-header h3 { 
  margin: 0;
  color: #d4a574;
  font-size: 1.15rem;
  letter-spacing: 1.5px;
  font-weight: 500;
}
.confirm-body { 
  padding: 28px 24px;
  color: #b8a898;
  line-height: 1.7;
  font-size: 1rem;
}
.confirm-body p { margin: 0; }
.confirm-footer {
  padding: 18px 24px;
  display: flex;
  justify-content: flex-end;
  gap: 14px;
  background: rgba(0,0,0,0.2);
  border-top: 1px solid rgba(92, 77, 60, 0.2);
}
.btn-cancel {
  background: transparent;
  border: 1px solid rgba(92, 77, 60, 0.4);
  color: #8b7b6b;
  padding: 10px 22px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.25s;
  font-family: inherit;
  font-size: 0.95rem;
}
.btn-cancel:hover { 
  border-color: rgba(139, 107, 78, 0.6);
  color: #a89b8b;
  background: rgba(255,255,255,0.03);
}

.btn-confirm {
  background: linear-gradient(135deg, #3a2e22 0%, #2a2018 100%);
  border: 1px solid rgba(212, 165, 116, 0.5);
  color: #d4a574;
  padding: 10px 26px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.25s;
  font-weight: 600;
  font-family: inherit;
  font-size: 0.95rem;
}
.btn-confirm:hover { 
  background: linear-gradient(135deg, #d4a574 0%, #b89060 100%);
  color: #1a1614;
  box-shadow: 0 4px 20px rgba(212, 165, 116, 0.3);
}
.btn-confirm.danger { 
  border-color: rgba(239, 83, 80, 0.5);
  color: #ef5350;
  background: rgba(239, 83, 80, 0.1);
}
.btn-confirm.danger:hover { 
  background: linear-gradient(135deg, #ef5350 0%, #d32f2f 100%);
  color: white;
  box-shadow: 0 4px 20px rgba(239, 83, 80, 0.3);
}

/* ========== LOGO ========== */
.card-logo-img { 
  width: 100%;
  height: 100%;
  object-fit: contain;
  border-radius: 10px;
}
.header-logo-img { 
  width: 42px;
  height: 42px;
  border-radius: 8px;
  margin-right: 16px;
  border: 1px solid rgba(92, 77, 60, 0.4);
  object-fit: cover;
  box-shadow: 0 2px 10px rgba(0,0,0,0.3);
}

/* ========== LOADING ========== */
.loading-overlay {
  position: fixed;
  top: 0; left: 0;
  width: 100%; height: 100%;
  background: rgba(10, 8, 6, 0.95);
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  z-index: 300;
}
.spinner {
  width: 50px;
  height: 50px;
  border: 3px solid rgba(92, 77, 60, 0.3);
  border-top-color: #d4a574;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}
@keyframes spin {
  to { transform: rotate(360deg); }
}
.loading-text {
  margin-top: 20px;
  color: #8b7b6b;
  font-size: 1rem;
  letter-spacing: 1px;
}

/* ========== RESPONSIVE ========== */
@media (max-width: 1024px) {
  .game-content-wrapper { 
    grid-template-columns: 260px 1fr;
  }
  .narrative-window {
    padding: 30px 35px;
  }
  .input-container {
    padding: 20px 35px;
  }
}

@media (max-width: 768px) {
  .game-content-wrapper { 
    grid-template-columns: 1fr;
  }
  .status-panel {
    position: fixed;
    top: 0; right: 0;
    width: 88%;
    max-width: 360px;
    height: 100%;
    transform: translateX(100%);
    transition: transform 0.35s cubic-bezier(0.4, 0, 0.2, 1);
    z-index: 150;
    box-shadow: -15px 0 40px rgba(0,0,0,0.6);
    border-left: 1px solid rgba(92, 77, 60, 0.3);
  }
  .status-panel.mobile-show { 
    transform: translateX(0);
  }
  .close-panel-btn { display: block; }
  .pc-actions { display: none; }
  .menu-toggle { display: block; }
  
  .narrative-window {
    padding: 24px 20px;
  }
  .narrative-block {
    font-size: 1rem;
    line-height: 1.9;
  }
  
  .input-container { 
    padding: 16px 20px;
    min-height: 85px;
  }
  .adventure-input { 
    height: 48px;
    font-size: 1rem;
    padding: 0 18px;
    line-height: 48px;
  }
  .btn-adventure { 
    height: 48px;
    padding: 0 20px;
    font-size: 1rem;
  }
  .btn-soul-burn { 
    width: 48px;
    height: 48px;
  }
  
  .selection-header h1 {
    font-size: 2rem;
    letter-spacing: 2px;
  }
  .game-card {
    width: 100%;
    max-width: 340px;
  }
}

@media (max-width: 480px) {
  .game-header {
    padding: 0 16px;
    height: 56px;
  }
  .game-title h2 {
    font-size: 1.1rem;
  }
  .mod-id {
    display: none;
  }
  
  .narrative-window {
    padding: 20px 16px;
  }
  .narrative-block {
    font-size: 0.95rem;
  }
  
  .input-container {
    padding: 14px 16px;
  }
  .action-input-wrapper {
    gap: 10px;
  }
  .btn-adventure {
    padding: 0 16px;
    font-size: 0.95rem;
  }
  .btn-soul-burn {
    width: 44px;
    height: 44px;
    font-size: 1.3rem;
  }
}
</style>
