<script setup lang="ts">
import { onErrorCaptured, ref, watch } from 'vue'

const props = defineProps<{ resetKey?: any }>()
const error = ref<Error | null>(null)

onErrorCaptured((err) => {
  error.value = err instanceof Error ? err : new Error(String(err))
  return false
})

watch(
  () => props.resetKey,
  () => {
    error.value = null
  },
)

function reset(): void {
  error.value = null
}
</script>

<template>
  <div v-if="error" class="error-boundary-view" role="alert">
    <div class="error-boundary-card">
      <div class="error-icon">⚠️</div>
      <h2>页面遇到意外错误</h2>
      <p class="error-msg">{{ error.message }}</p>
      <button type="button" class="retry-btn" @click="reset">重试刷新</button>
    </div>
  </div>
  <slot v-else></slot>
</template>

<style scoped>
.error-boundary-view {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  padding: 20px;
  background: var(--bg);
}

.error-boundary-card {
  max-width: 460px;
  padding: 32px;
  background: var(--surface-raised);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow);
  text-align: center;
}

.error-icon {
  font-size: 36px;
  margin-bottom: 12px;
}

h2 {
  font-size: 18px;
  margin-bottom: 8px;
  color: var(--text);
}

.error-msg {
  color: var(--danger);
  font-family: var(--font-mono);
  font-size: 12px;
  margin-bottom: 20px;
  word-break: break-word;
  background: var(--surface-muted);
  padding: 10px;
  border-radius: var(--radius-xs);
}

.retry-btn {
  padding: 6px 16px;
  background: var(--accent);
  color: #ffffff;
  border-radius: var(--radius-sm);
  font-weight: 500;
  transition: opacity 120ms;
}

.retry-btn:hover {
  opacity: 0.9;
}
</style>
