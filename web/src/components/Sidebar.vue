<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import type { DiagramMeta } from '@/api/types'
import { iconSvg } from '@/utils/icons'

const props = defineProps<{
  diagrams: DiagramMeta[]
  activeId: string
  collapsed: boolean
  isMobile?: boolean
  mobileOpen?: boolean
}>()

const emit = defineEmits<{
  select: [id: string]
  create: []
  rename: [id: string, newTitle: string]
  delete: [id: string]
  duplicate: [id: string]
  toggleCollapse: []
  closeMobile: []
}>()

const searchQuery = ref('')
const editingId = ref<string | null>(null)
const editingTitle = ref('')
const editInputRef = ref<HTMLInputElement | null>(null)

const filteredDiagrams = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return props.diagrams
  return props.diagrams.filter((d) => d.title.toLowerCase().includes(q))
})

function startRename(d: DiagramMeta, e?: Event) {
  if (e) e.stopPropagation()
  editingId.value = d.id
  editingTitle.value = d.title
  nextTick(() => {
    editInputRef.value?.focus()
    editInputRef.value?.select()
  })
}

function saveRename() {
  if (editingId.value) {
    const trimmed = editingTitle.value.trim()
    if (trimmed) {
      emit('rename', editingId.value, trimmed)
    }
  }
  editingId.value = null
  editingTitle.value = ''
}

function cancelRename() {
  editingId.value = null
  editingTitle.value = ''
}

function handleDelete(id: string, title: string, e: Event) {
  e.stopPropagation()
  if (confirm(`确定要删除图表 "${title}" 吗？此操作不可撤销。`)) {
    emit('delete', id)
  }
}

function handleDuplicate(id: string, e: Event) {
  e.stopPropagation()
  emit('duplicate', id)
}

function handleItemClick(id: string) {
  emit('select', id)
  if (props.isMobile) {
    emit('closeMobile')
  }
}

function formatDate(iso: string): string {
  try {
    const d = new Date(iso)
    const now = new Date()
    const isToday = d.toDateString() === now.toDateString()
    const pad = (n: number) => String(n).padStart(2, '0')
    if (isToday) {
      return `${pad(d.getHours())}:${pad(d.getMinutes())}`
    }
    return `${d.getMonth() + 1}/${d.getDate()} ${pad(d.getHours())}:${pad(d.getMinutes())}`
  } catch {
    return ''
  }
}
</script>

<template>
  <aside
    class="sidebar-container"
    :class="{
      collapsed: !isMobile && collapsed,
      'mobile-drawer': isMobile,
      'mobile-open': isMobile && mobileOpen,
    }"
    :role="isMobile ? 'dialog' : 'complementary'"
    :aria-modal="isMobile ? 'true' : undefined"
    :aria-label="isMobile ? '图表列表' : undefined"
  >
    <!-- Header -->
    <div class="sidebar-header">
      <div v-if="isMobile || !collapsed" class="header-title">
        <span class="logo-emoji" aria-hidden="true">📊</span>
        <span class="title-text">图表库</span>
        <span class="count-badge" :title="`共 ${diagrams.length} 个图表`">{{ diagrams.length }}</span>
      </div>

      <!-- Desktop Collapse Button -->
      <button
        v-if="!isMobile"
        type="button"
        class="icon-btn collapse-toggle-btn"
        :title="collapsed ? '展开侧边栏' : '折叠侧边栏'"
        :aria-label="collapsed ? '展开侧边栏' : '折叠侧边栏'"
        @click="emit('toggleCollapse')"
      >
        <span v-html="iconSvg(collapsed ? 'chevron-right' : 'sidebar', 16)"></span>
      </button>

      <!-- Mobile Close Button -->
      <button
        v-else
        type="button"
        class="icon-btn close-drawer-btn"
        title="关闭抽屉"
        aria-label="关闭抽屉"
        @click="emit('closeMobile')"
      >
        <span v-html="iconSvg('x', 18)"></span>
      </button>
    </div>

    <!-- Content -->
    <div v-if="isMobile || !collapsed" class="sidebar-content">
      <!-- Create New Diagram Action -->
      <div class="action-row">
        <button
          type="button"
          class="create-btn"
          title="新建一个递增命名的图表"
          @click="
            () => {
              emit('create')
              if (isMobile) emit('closeMobile')
            }
          "
        >
          <span v-html="iconSvg('plus', 16)"></span>
          <span>新建图表</span>
        </button>
      </div>

      <!-- Search Box -->
      <div class="search-box">
        <span class="search-icon" v-html="iconSvg('search', 14)"></span>
        <input
          v-model="searchQuery"
          type="text"
          class="search-input"
          placeholder="搜索图表名称..."
          spellcheck="false"
          aria-label="搜索图表"
        />
        <button
          v-if="searchQuery"
          type="button"
          class="clear-search-btn"
          title="清空搜索"
          aria-label="清空搜索"
          @click="searchQuery = ''"
        >
          ✕
        </button>
      </div>

      <!-- Diagram List -->
      <div class="diagram-list" role="list" aria-label="图表列表">
        <div
          v-for="d in filteredDiagrams"
          :key="d.id"
          class="diagram-item"
          :class="{ active: d.id === activeId }"
          role="listitem"
          :aria-current="d.id === activeId ? 'true' : undefined"
          @click="handleItemClick(d.id)"
          @dblclick="startRename(d)"
        >
          <div class="item-main">
            <span class="item-icon" v-html="iconSvg('file-code', 16)"></span>

            <div v-if="editingId === d.id" class="edit-wrapper" @click.stop>
              <input
                ref="editInputRef"
                v-model="editingTitle"
                type="text"
                class="rename-input"
                aria-label="重命名图表"
                @blur="saveRename"
                @keydown.enter="saveRename"
                @keydown.esc="cancelRename"
              />
            </div>
            <div v-else class="item-info">
              <span class="item-title" :title="d.title">{{ d.title }}</span>
              <span class="item-date">{{ formatDate(d.updatedAt) }}</span>
            </div>
          </div>

          <div v-if="editingId !== d.id" class="item-actions" @click.stop>
            <button
              type="button"
              class="item-btn"
              title="重命名图表"
              aria-label="重命名"
              @click="startRename(d, $event)"
            >
              <span v-html="iconSvg('edit', 14)"></span>
            </button>
            <button
              type="button"
              class="item-btn"
              title="创建副本"
              aria-label="复制"
              @click="handleDuplicate(d.id, $event)"
            >
              <span v-html="iconSvg('duplicate', 14)"></span>
            </button>
            <button
              type="button"
              class="item-btn danger"
              title="删除图表"
              aria-label="删除"
              :disabled="diagrams.length <= 1"
              @click="handleDelete(d.id, d.title, $event)"
            >
              <span v-html="iconSvg('trash', 14)"></span>
            </button>
          </div>
        </div>

        <div v-if="filteredDiagrams.length === 0" class="empty-hint">
          <p v-if="searchQuery">未找到匹配图表</p>
          <p v-else>暂无图表，点击上方新建</p>
        </div>
      </div>
    </div>
  </aside>
</template>

<style scoped>
.sidebar-container {
  display: flex;
  flex-direction: column;
  width: var(--sidebar-width);
  height: 100%;
  background: var(--surface-raised);
  border-right: 1px solid var(--border);
  transition: width 180ms cubic-bezier(0.16, 1, 0.3, 1);
  flex-shrink: 0;
  overflow: hidden;
}

.sidebar-container.collapsed {
  width: 48px;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: var(--header-height);
  padding: 0 14px;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  font-size: 13px;
  color: var(--text);
}

.logo-emoji {
  font-size: 18px;
  line-height: 1;
}

.title-text {
  font-weight: 700;
  letter-spacing: 0.3px;
}

.count-badge {
  padding: 1px 7px;
  border-radius: 10px;
  background: var(--surface-muted);
  color: var(--text-muted);
  font-size: 11px;
  font-weight: 600;
}

.icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--radius-xs);
  color: var(--text-muted);
  transition: all 120ms ease;
}

.icon-btn:hover {
  color: var(--text);
  background: var(--surface-hover);
}

.sidebar-content {
  display: flex;
  flex-direction: column;
  flex: 1;
  padding: 12px;
  gap: 12px;
  overflow: hidden;
}

.action-row {
  width: 100%;
}

.create-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  width: 100%;
  min-height: 38px;
  padding: 8px 14px;
  background: var(--accent);
  color: #ffffff;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 600;
  box-shadow: 0 2px 6px color-mix(in srgb, var(--accent) 25%, transparent);
  transition: all 120ms ease;
}

.create-btn:hover {
  opacity: 0.92;
  transform: translateY(-1px);
}

.create-btn:active {
  transform: translateY(0);
}

.search-box {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  min-height: 36px;
  background: var(--surface-muted);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  transition: border-color 120ms;
}

.search-box:focus-within {
  border-color: var(--accent);
}

.search-icon {
  display: flex;
  color: var(--text-faint);
}

.search-input {
  flex: 1;
  border: none;
  background: transparent;
  outline: none;
  font-size: 13px;
  color: var(--text);
}

.clear-search-btn {
  color: var(--text-faint);
  font-size: 12px;
  line-height: 1;
  padding: 3px 5px;
  border-radius: 50%;
  background: var(--surface);
}

.clear-search-btn:hover {
  color: var(--text);
}

.diagram-list {
  flex: 1;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding-right: 2px;
}

.diagram-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 10px;
  min-height: 46px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: background 120ms, border-color 120ms;
  border: 1px solid transparent;
}

.diagram-item:hover {
  background: var(--surface-hover);
}

.diagram-item.active {
  background: var(--accent-soft);
  border-color: color-mix(in srgb, var(--accent) 35%, transparent);
}

.diagram-item.active .item-title {
  color: var(--accent-strong);
  font-weight: 600;
}

.diagram-item.active .item-icon {
  color: var(--accent-strong);
}

.item-main {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  flex: 1;
}

.item-icon {
  display: flex;
  color: var(--text-muted);
  flex-shrink: 0;
}

.item-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
  flex: 1;
  gap: 2px;
}

.item-title {
  font-size: 13px;
  color: var(--text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-date {
  font-size: 11px;
  color: var(--text-faint);
}

.edit-wrapper {
  flex: 1;
  min-width: 0;
}

.rename-input {
  width: 100%;
  padding: 4px 6px;
  font-size: 13px;
  border: 1.5px solid var(--accent);
  border-radius: var(--radius-xs);
  background: var(--surface-raised);
  outline: none;
  color: var(--text);
}

.item-actions {
  display: none;
  align-items: center;
  gap: 3px;
  margin-left: 6px;
}

.diagram-item:hover .item-actions,
.diagram-item.active .item-actions {
  display: flex;
}

.item-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: var(--radius-xs);
  color: var(--text-muted);
  transition: all 100ms;
}

.item-btn:hover {
  color: var(--text);
  background: var(--surface-muted);
}

.item-btn.danger:hover {
  color: var(--danger);
  background: var(--danger-soft);
}

.item-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.empty-hint {
  padding: 36px 8px;
  text-align: center;
  color: var(--text-faint);
  font-size: 13px;
}

/* Mobile Drawer Overlay Styles */
.sidebar-container.mobile-drawer {
  position: fixed;
  top: 0;
  bottom: 0;
  left: 0;
  z-index: 100;
  width: min(84vw, 320px);
  height: 100vh;
  height: 100dvh;
  max-height: 100dvh;
  background: var(--surface-raised);
  box-shadow: 0 0 40px rgba(0, 0, 0, 0.3);
  transform: translate3d(-105%, 0, 0);
  transition:
    transform 240ms cubic-bezier(0.16, 1, 0.3, 1),
    visibility 240ms cubic-bezier(0.16, 1, 0.3, 1);
  visibility: hidden;
  padding-top: max(8px, env(safe-area-inset-top));
  padding-bottom: max(12px, env(safe-area-inset-bottom));
  padding-left: max(8px, env(safe-area-inset-left));
}

.sidebar-container.mobile-drawer.mobile-open {
  visibility: visible;
  transform: translate3d(0, 0, 0);
}

@media (max-width: 840px), (pointer: coarse) {
  .diagram-item {
    min-height: 50px;
    padding: 10px 12px;
  }

  .item-actions {
    display: flex;
  }

  .item-btn {
    width: 32px;
    height: 32px;
  }
}
</style>
