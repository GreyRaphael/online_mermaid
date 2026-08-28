<script setup lang="ts">
defineProps<{
  message: string
  isError?: boolean
  styleObj?: Record<string, string>
}>()
</script>

<template>
  <div
    v-if="message"
    class="mermaid-toast"
    :class="{ error: isError }"
    :style="styleObj"
    role="status"
  >
    <span>{{ message }}</span>
    <div class="toast-arrow"></div>
  </div>
</template>

<style scoped>
.mermaid-toast {
  position: fixed;
  z-index: 100000;
  transform: translate(-50%, -100%);
  padding: 6px 12px;
  border-radius: var(--radius-sm);
  background: #1e1e2e;
  color: #ffffff;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.28);
  font-family: var(--font-sans);
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
  pointer-events: none;
  transition: all 150ms ease;
}

.mermaid-toast.error {
  background: var(--danger);
}

.toast-arrow {
  position: absolute;
  bottom: -4px;
  left: 50%;
  transform: translateX(-50%) rotate(45deg);
  width: 8px;
  height: 8px;
  background: inherit;
}
</style>
