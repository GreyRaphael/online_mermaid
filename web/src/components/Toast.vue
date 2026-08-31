<script setup lang="ts">
defineProps<{
  message: string
  isError?: boolean
  styleObj?: Record<string, string>
}>()
</script>

<template>
  <Transition name="toast-pop">
    <div
      v-if="message"
      class="mermaid-toast"
      :class="{ error: isError }"
      :style="styleObj"
      role="status"
    >
      <span class="toast-icon" aria-hidden="true">{{ isError ? '⚠️' : '✓' }}</span>
      <span class="toast-text">{{ message }}</span>
    </div>
  </Transition>
</template>

<style scoped>
.mermaid-toast {
  position: fixed;
  z-index: 100000;
  transform: translateX(-50%);
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  background: #1e1e2e;
  color: #ffffff;
  box-shadow: 0 6px 24px rgba(0, 0, 0, 0.35);
  font-family: var(--font-sans);
  font-size: 13px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
  white-space: nowrap;
  pointer-events: none;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.mermaid-toast.error {
  background: #dc2626;
  border-color: rgba(255, 255, 255, 0.2);
}

.toast-icon {
  font-size: 14px;
  line-height: 1;
}

.toast-text {
  line-height: 1.4;
}

.toast-pop-enter-active,
.toast-pop-leave-active {
  transition: opacity 180ms ease, transform 180ms cubic-bezier(0.16, 1, 0.3, 1);
}

.toast-pop-enter-from,
.toast-pop-leave-to {
  opacity: 0;
  transform: translate(-50%, -12px);
}
</style>
