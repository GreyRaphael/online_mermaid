<script setup lang="ts">
import { ref } from 'vue'
import { login } from '@/api/client'
import type { SessionResponse } from '@/api/types'
import ThemeControl from '@/components/ThemeControl.vue'

const props = defineProps<{ initialError?: string }>()
const emit = defineEmits<{ authenticated: [session: SessionResponse] }>()

const username = ref('admin')
const password = ref('')
const errorMessage = ref(props.initialError ?? '')
const submitting = ref(false)

async function submit(): Promise<void> {
  if (!username.value.trim() || !password.value) {
    errorMessage.value = '请输入用户名和密码'
    return
  }
  submitting.value = true
  errorMessage.value = ''
  try {
    const session = await login(username.value.trim(), password.value)
    if (!session.authenticated) throw new Error('登录未建立有效会话')
    password.value = ''
    emit('authenticated', session)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '登录失败，请稍后重试'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <main class="login-view">
    <div class="login-theme"><ThemeControl /></div>
    <section class="login-card" aria-labelledby="login-title">
      <div class="brand-mark" aria-hidden="true">M</div>
      <p class="eyebrow">ONLINE MERMAID WORKSPACE</p>
      <h1 id="login-title">Online Mermaid</h1>
      <p class="login-intro">登录后创建、编辑和渲染 Mermaid 流程图与图表。</p>

      <form class="login-form" @submit.prevent="submit">
        <div class="form-group">
          <label for="username">用户名</label>
          <input
            id="username"
            v-model="username"
            name="username"
            type="text"
            autocomplete="username"
            autocapitalize="none"
            spellcheck="false"
            :disabled="submitting"
            autofocus
          />
        </div>

        <div class="form-group">
          <label for="password">密码</label>
          <input
            id="password"
            v-model="password"
            name="password"
            type="password"
            autocomplete="current-password"
            :disabled="submitting"
          />
        </div>

        <p v-if="errorMessage" class="form-error" role="alert">{{ errorMessage }}</p>
        <button class="primary-button login-button" type="submit" :disabled="submitting">
          <span v-if="submitting" class="button-spinner" aria-hidden="true"></span>
          {{ submitting ? '正在登录…' : '登录' }}
        </button>
      </form>
      <p class="login-footnote">数据存储在本地 SQLite3 · 会话安全加密</p>
    </section>
  </main>
</template>

<style scoped>
.login-view {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100vw;
  height: 100vh;
  background: var(--bg);
}

.login-theme {
  position: absolute;
  top: 20px;
  right: 20px;
}

.login-card {
  width: 100%;
  max-width: 380px;
  padding: 36px 32px;
  background: var(--surface-raised);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow);
  display: flex;
  flex-direction: column;
  align-items: center;
}

.brand-mark {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  background: var(--accent);
  color: #ffffff;
  font-size: 22px;
  font-weight: 700;
  margin-bottom: 16px;
  box-shadow: 0 4px 12px color-mix(in srgb, var(--accent) 30%, transparent);
}

.eyebrow {
  font-size: 10px;
  font-weight: 700;
  color: var(--accent-strong);
  letter-spacing: 1.5px;
  margin-bottom: 4px;
}

h1 {
  font-size: 22px;
  font-weight: 700;
  color: var(--text);
  margin-bottom: 6px;
}

.login-intro {
  font-size: 12px;
  color: var(--text-muted);
  text-align: center;
  margin-bottom: 24px;
}

.login-form {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text);
}

input {
  width: 100%;
  padding: 8px 12px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  color: var(--text);
  font-size: 13px;
  transition: border-color 120ms;
}

input:focus {
  outline: none;
  border-color: var(--accent);
}

.form-error {
  font-size: 12px;
  color: var(--danger);
  background: var(--danger-soft);
  padding: 8px 12px;
  border-radius: var(--radius-xs);
  text-align: center;
}

.login-button {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  padding: 9px;
  background: var(--accent);
  color: #ffffff;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 600;
  margin-top: 6px;
  transition: opacity 120ms;
}

.login-button:hover {
  opacity: 0.9;
}

.login-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.login-footnote {
  font-size: 11px;
  color: var(--text-faint);
  margin-top: 24px;
}

.button-spinner {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: #ffffff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
