<template>
  <div class="login-container">
    <!-- 背景层 -->
    <div class="stars"></div>
    <div class="twinkling"></div>
    
    <div class="adventure-card">
      <div class="card-ornament top-left"></div>
      <div class="card-ornament top-right"></div>
      <div class="card-ornament bottom-left"></div>
      <div class="card-ornament bottom-right"></div>
      
      <div class="card-content">
        <header class="login-header">
          <div class="logo-wrapper">
             <img :src="gameLogo" alt="宿命之下" class="main-logo" />
          </div>
          <h1>宿命之下</h1>
          <div class="divider">
            <span class="rune">⚔</span>
          </div>
          <p class="subtitle">以魂为契，争一线生机</p>
        </header>

        <form class="login-form" @submit.prevent="handleSubmit">
          <div class="form-group">
            <label for="username">尊号</label>
            <div class="input-wrapper">
              <input 
                id="username"
                v-model="form.username" 
                type="text" 
                placeholder="Name"
                autocomplete="off"
                required
              >
              <span class="focus-border"></span>
            </div>
          </div>
          
          <div class="form-group">
            <label for="password">命魂之契</label>
            <div class="input-wrapper">
              <input 
                id="password"
                v-model="form.password" 
                type="password" 
                placeholder="Password"
                autocomplete="current-password"
                required
              >
              <span class="focus-border"></span>
            </div>
          </div>
          
          <button type="submit" class="btn-adventure" :disabled="loading">
            <span class="btn-text">{{ loading ? '咏唱中...' : '踏入蛊界' }}</span>
            <div class="btn-shine"></div>
          </button>
        </form>
      </div>
    </div>
    
    <!-- 自定义 Toast -->
    <div v-if="toast.show" :class="['toast', toast.type]">
      {{ toast.message }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import gameLogo from '@/img/sumingzhixia.png'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const form = reactive({
  username: '',
  password: ''
})

// 简单的 Toast 系统
const toast = reactive({
  show: false,
  message: '',
  type: 'info'
})

const showToast = (msg: string, type: 'success' | 'error' = 'info') => {
  toast.message = msg
  toast.type = type
  toast.show = true
  setTimeout(() => {
    toast.show = false
  }, 3000)
}

const handleSubmit = async () => {
  if (!form.username || !form.password) {
    showToast('请完整填写契约', 'error')
    return
  }
  
  if (form.username.length < 3) {
    showToast('名字太短，法则不予承认', 'error')
    return
  }
  
  if (form.password.length < 6) {
    showToast('契约之力不足（密码需6位以上）', 'error')
    return
  }
  
  try {
    loading.value = true
    await authStore.login({
      username: form.username,
      password: form.password
    })
    
    showToast('欢迎回来，冒险者', 'success')
    
    setTimeout(() => {
      if (authStore.isAdmin()) {
        router.push('/admin')
      } else {
        router.push('/chat')
      }
    }, 800)
  } catch (error: any) {
    console.error('登录失败:', error)
    showToast('契约验证失败，请重试', 'error')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Noto+Serif+SC:wght@400;700&display=swap');

/* 基础重置 */
* {
  box-sizing: border-box;
}

.login-container {
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: #000;
  color: #d4a574;
  font-family: 'Noto Serif SC', serif;
  position: relative;
  overflow: hidden;
}

/* 星空背景动画 */
.stars, .twinkling {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  width: 100%;
  height: 100%;
  display: block;
}

.stars {
  background: #000 url("data:image/svg+xml,%3Csvg width='200' height='200' viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Ccircle cx='20' cy='30' r='1' fill='%23FFF' opacity='0.5'/%3E%3Ccircle cx='50' cy='80' r='1' fill='%23FFF' opacity='0.4'/%3E%3Ccircle cx='90' cy='10' r='1' fill='%23FFF' opacity='0.6'/%3E%3Ccircle cx='150' cy='130' r='1.5' fill='%23FFF' opacity='0.4'/%3E%3Ccircle cx='180' cy='50' r='1' fill='%23FFF' opacity='0.5'/%3E%3C/svg%3E") repeat;
  z-index: 0;
}

.twinkling {
  background: transparent url("data:image/svg+xml,%3Csvg width='200' height='200' viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noise'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.6' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noise)' opacity='0.05'/%3E%3C/svg%3E");
  z-index: 1;
  animation: move-twink-back 200s linear infinite;
}

@keyframes move-twink-back {
  from {background-position:0 0;}
  to {background-position:-10000px 5000px;}
}

/* 主卡片 styled as an ancient tome or tablet */
.adventure-card {
  position: relative;
  width: 380px;
  background: rgba(18, 14, 12, 0.9);
  border: 2px solid #5c4d3c;
  padding: 3px; /* for double border effect */
  z-index: 10;
  box-shadow: 0 0 50px rgba(0, 0, 0, 0.8);
}

.card-content {
  border: 1px solid #3a2e22;
  padding: 40px 30px;
  background: 
    linear-gradient(rgba(18, 14, 12, 0.95), rgba(18, 14, 12, 0.95)),
    url("data:image/svg+xml,%3Csvg width='100' height='100' viewBox='0 0 100 100' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noise'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.8'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noise)' opacity='0.08'/%3E%3C/svg%3E");
}

/* 角落装饰 */
.card-ornament {
  position: absolute;
  width: 20px;
  height: 20px;
  border: 2px solid #d4a574;
  transition: all 0.3s ease;
}

.top-left { top: -2px; left: -2px; border-right: none; border-bottom: none; }
.top-right { top: -2px; right: -2px; border-left: none; border-bottom: none; }
.bottom-left { bottom: -2px; left: -2px; border-right: none; border-top: none; }
.bottom-right { bottom: -2px; right: -2px; border-left: none; border-top: none; }

.adventure-card:hover .card-ornament {
  width: 30px;
  height: 30px;
  border-color: #ffd700;
  box-shadow: 0 0 10px rgba(212, 165, 116, 0.3);
}

/* Header */
.login-header {
  text-align: center;
  margin-bottom: 30px;
}

.logo-wrapper {
  margin-bottom: 20px;
  display: flex; justify-content: center;
}
.main-logo {
  width: 120px; height: auto; max-height: 120px;
  border-radius: 12px;
  border: 1px solid rgba(212, 165, 116, 0.5);
  box-shadow: 0 0 20px rgba(0, 0, 0, 0.5), 0 0 10px rgba(212, 165, 116, 0.2);
  transition: transform 0.3s ease;
}
.main-logo:hover {
  transform: scale(1.05);
  border-color: #ffd700;
  box-shadow: 0 0 25px rgba(212, 165, 116, 0.4);
}

h1 {
  font-size: 2rem;
  margin: 0;
  color: #e8dcd0;
  text-shadow: 0 2px 4px rgba(0,0,0,0.5);
  letter-spacing: 2px;
}

.divider {
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 15px 0;
}

.divider::before, .divider::after {
  content: '';
  height: 1px;
  width: 50px;
  background: linear-gradient(90deg, transparent, #8b6b4e, transparent);
}

.rune {
  margin: 0 10px;
  color: #8b6b4e;
  font-size: 0.8rem;
}

.subtitle {
  margin: 0;
  color: #8b6b4e;
  font-size: 0.9rem;
  font-style: italic;
}

/* Form Styles */
.login-form {
  display: flex;
  flex-direction: column;
  gap: 25px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

label {
  font-size: 0.85rem;
  color: #d4a574;
  margin-left: 4px;
}

.input-wrapper {
  position: relative;
}

input {
  width: 100%;
  background: rgba(0, 0, 0, 0.3);
  border: none;
  border-bottom: 1px solid #5c4d3c;
  padding: 12px 10px;
  color: #e8dcd0;
  font-family: 'Noto Serif SC', serif;
  font-size: 1rem;
  transition: all 0.3s ease;
  outline: none;
}

input:focus {
  background: rgba(0, 0, 0, 0.5);
}

/* 动态边框效果 */
.focus-border {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 0;
  height: 1px;
  background-color: #d4a574;
  transition: 0.4s;
}

input:focus ~ .focus-border {
  width: 100%;
  box-shadow: 0 0 8px rgba(212, 165, 116, 0.6);
}

input::placeholder {
  color: rgba(139, 107, 78, 0.4);
  font-style: italic;
  font-size: 0.9rem;
}

/* 按钮样式 */
.btn-adventure {
  position: relative;
  margin-top: 15px;
  padding: 14px;
  background: linear-gradient(to bottom, #3a2e22, #261e16);
  border: 1px solid #5c4d3c;
  color: #d4a574;
  font-family: inherit;
  font-size: 1.1rem;
  font-weight: bold;
  cursor: pointer;
  overflow: hidden;
  transition: all 0.3s;
  text-transform: uppercase;
  letter-spacing: 2px;
  box-shadow: 0 4px 6px rgba(0,0,0,0.3);
}

.btn-adventure:hover {
  border-color: #d4a574;
  color: #ffd700;
  box-shadow: 0 0 15px rgba(212, 165, 116, 0.2);
  text-shadow: 0 0 5px rgba(255, 215, 0, 0.5);
}

.btn-adventure:active {
  transform: translateY(1px);
}

.btn-adventure:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* 按钮光效 */
.btn-shine {
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(
    90deg, 
    transparent, 
    rgba(212, 165, 116, 0.2), 
    transparent
  );
  transition: 0.5s;
}

.btn-adventure:hover .btn-shine {
  left: 100%;
  transition: 0.5s;
}

/* Toast */
.toast {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  padding: 12px 24px;
  background: rgba(18, 14, 12, 0.95);
  border: 1px solid;
  border-radius: 4px;
  z-index: 100;
  font-size: 0.9rem;
  box-shadow: 0 4px 12px rgba(0,0,0,0.5);
  animation: slideDown 0.3s ease-out;
}

.toast.success {
  border-color: #4caf50;
  color: #4caf50;
}

.toast.error {
  border-color: #d32f2f;
  color: #d32f2f;
}

.toast.info {
  border-color: #d4a574;
  color: #d4a574;
}

@keyframes slideDown {
  from { top: -50px; opacity: 0; }
  to { top: 20px; opacity: 1; }
}

/* 响应式 */
@media (max-width: 480px) {
  .adventure-card {
    width: 90%;
    margin: 20px;
  }
}
</style>