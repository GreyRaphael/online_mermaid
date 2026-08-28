<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { AUTH_EXPIRED_EVENT, getSession } from '@/api/client'
import type { SessionResponse } from '@/api/types'
import ErrorBoundary from '@/components/ErrorBoundary.vue'
import LoginView from '@/views/LoginView.vue'
import MermaidView from '@/views/MermaidView.vue'
import { useTheme } from '@/composables/useTheme'

useTheme()

const checkingSession = ref(true)
const session = ref<SessionResponse>({ authenticated: false })
const startupError = ref('')

function handleAuthenticated(value: SessionResponse): void {
  session.value = value
  startupError.value = ''
}

function handleSignedOut(): void {
  session.value = { authenticated: false }
}

function handleExpired(): void {
  if (session.value.authenticated) {
    startupError.value = '会话已过期，请重新登录'
    handleSignedOut()
  }
}

onMounted(async () => {
  window.addEventListener(AUTH_EXPIRED_EVENT, handleExpired)
  try {
    session.value = await getSession()
  } catch (error) {
    startupError.value = error instanceof Error ? error.message : '无法连接服务器'
  } finally {
    checkingSession.value = false
  }
})

onBeforeUnmount(() => {
  window.removeEventListener(AUTH_EXPIRED_EVENT, handleExpired)
})
</script>

<template>
  <div v-if="checkingSession" class="app-loading" role="status" aria-live="polite">
    <div class="brand-mark">M</div>
    <span class="loading-ring" aria-hidden="true"></span>
    <p>正在初始化工作区…</p>
  </div>
  <ErrorBoundary v-else :reset-key="session.username">
    <MermaidView
      v-if="session.authenticated"
      :username="session.username || '用户'"
      :version="session.version"
      @signed-out="handleSignedOut"
    />
    <LoginView v-else :initial-error="startupError" @authenticated="handleAuthenticated" />
  </ErrorBoundary>
</template>

<style scoped>
.app-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100vw;
  height: 100vh;
  gap: 16px;
  background: var(--bg);
  color: var(--text-muted);
  font-size: 13px;
}

.brand-mark {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: var(--radius-md);
  background: var(--accent);
  color: #ffffff;
  font-size: 20px;
  font-weight: 700;
  box-shadow: 0 4px 12px color-mix(in srgb, var(--accent) 30%, transparent);
}

.loading-ring {
  width: 24px;
  height: 24px;
  border: 2px solid var(--border);
  border-top-color: var(--accent);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
