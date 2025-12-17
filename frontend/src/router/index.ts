import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/chat'
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/auth/callback/linux-do',
      name: 'oauth-callback',
      component: () => import('@/views/OAuthCallbackView.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/chat',
      name: 'chat',
      component: () => import('@/views/GameView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/admin',
      name: 'admin',
      component: () => import('@/views/AdminView.vue'),
      meta: { requiresAuth: true, requiresAdmin: true }
    }
  ]
})

router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()
  
  // 如果路由需要认证
  if (to.meta.requiresAuth) {
    if (!authStore.isAuthenticated()) {
      next('/login')
      return
    }
    
    // 如果token存在但用户信息不存在，尝试获取用户信息
    if (!authStore.user) {
      try {
        await authStore.getProfile()
      } catch (error) {
        // 获取用户信息失败，清除token并跳转到登录页
        authStore.logout()
        next('/login')
        return
      }
    }
    
    // 如果路由需要管理员权限
    if (to.meta.requiresAdmin && !authStore.isAdmin()) {
      next('/chat')
      return
    }
  } else {
    // 如果已登录用户访问登录页，重定向到相应页面
    if (to.path === '/login' && authStore.isAuthenticated()) {
      if (authStore.isAdmin()) {
        next('/admin')
      } else {
        next('/chat')
      }
      return
    }
  }
  
  next()
})

export default router