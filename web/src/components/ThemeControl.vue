<script setup lang="ts">
import { useTheme, type ThemeMode } from '@/composables/useTheme'
import { iconSvg } from '@/utils/icons'

const { mode, setMode } = useTheme()

const themes: { id: ThemeMode; label: string; icon: string }[] = [
  { id: 'day', label: '浅色', icon: 'sun' },
  { id: 'night', label: '深色', icon: 'moon' },
  { id: 'system', label: '跟随系统', icon: 'monitor' },
]
</script>

<template>
  <div class="theme-control" role="group" aria-label="主题选择">
    <button
      v-for="item in themes"
      :key="item.id"
      type="button"
      class="theme-btn"
      :class="{ active: mode === item.id }"
      :title="item.label"
      :aria-label="item.label"
      :aria-pressed="mode === item.id ? 'true' : 'false'"
      @click="setMode(item.id)"
    >
      <span class="icon" v-html="iconSvg(item.icon, 14)"></span>
    </button>
  </div>
</template>

<style scoped>
.theme-control {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  padding: 2px;
  border-radius: var(--radius-sm);
  background: var(--surface-muted);
  border: 1px solid var(--border);
}

.theme-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: var(--radius-xs);
  color: var(--text-muted);
  transition: all 120ms ease;
}

.theme-btn:hover {
  color: var(--text);
  background: var(--surface-hover);
}

.theme-btn.active {
  background: var(--surface-raised);
  color: var(--accent-strong);
  box-shadow: var(--shadow-sm);
}

.icon {
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
