<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  createDiagram,
  deleteDiagram,
  duplicateDiagram,
  fetchDiagram,
  fetchDiagrams,
  fetchNextName,
  logout,
  updateDiagram,
} from '@/api/client'
import type { Diagram, DiagramMeta } from '@/api/types'
import { useTheme } from '@/composables/useTheme'
import { iconSvg } from '@/utils/icons'
import Sidebar from '@/components/Sidebar.vue'
import DiagramEditor from '@/components/DiagramEditor.vue'
import DiagramPreview from '@/components/DiagramPreview.vue'
import FullscreenModal from '@/components/FullscreenModal.vue'
import ThemeControl from '@/components/ThemeControl.vue'
import Toast from '@/components/Toast.vue'

type ViewMode = 'preview' | 'edit' | 'split'

const props = defineProps<{
  username: string
  version?: string
}>()
const emit = defineEmits<{ signedOut: [] }>()

const { resolved } = useTheme()

const VIEW_MODE_KEY = 'online-mermaid-view-mode'
function getInitialViewMode(): ViewMode {
  try {
    const saved = window.localStorage.getItem(VIEW_MODE_KEY) as ViewMode
    if (saved && ['preview', 'edit', 'split'].includes(saved)) return saved
  } catch {
    // fallback
  }
  return 'split'
}

const viewMode = ref<ViewMode>(getInitialViewMode())
watch(viewMode, (mode) => {
  try {
    window.localStorage.setItem(VIEW_MODE_KEY, mode)
  } catch {
    // ignore
  }
})

// Mobile and Sidebar state
const mobileViewport = ref(false)
const mobileSidebarOpen = ref(false)
const sidebarCollapsed = ref(false)
const userMenuOpen = ref(false)
const signingOut = ref(false)

function handleViewportChange() {
  if (typeof window === 'undefined') return
  mobileViewport.value = window.matchMedia('(max-width: 840px)').matches
  if (!mobileViewport.value) {
    mobileSidebarOpen.value = false
  }
}

function toggleSidebar() {
  if (mobileViewport.value) {
    mobileSidebarOpen.value = !mobileSidebarOpen.value
  } else {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }
}

function closeSidebar() {
  mobileSidebarOpen.value = false
}

function handleGlobalClick(e: MouseEvent) {
  if (userMenuOpen.value && !(e.target as Element).closest('.user-menu-container')) {
    userMenuOpen.value = false
  }
}

const diagrams = ref<DiagramMeta[]>([])
const currentDiagram = ref<Diagram | null>(null)
const editableCode = ref('')
const isSaving = ref(false)
const saveSuccess = ref(false)
const saveError = ref('')

// Toast notifications
const toastMessage = ref('')
const toastIsError = ref(false)
let toastTimer: number | null = null

function showToast(msg: string, isError = false) {
  if (toastTimer) clearTimeout(toastTimer)
  toastMessage.value = msg
  toastIsError.value = isError
  toastTimer = window.setTimeout(() => {
    toastMessage.value = ''
    toastTimer = null
  }, 2200)
}

// Fullscreen state
const fullscreenOpen = ref(false)
const fullscreenSvg = ref('')
const fullscreenSource = ref('')

function handleOpenFullscreen(svgHtml: string, source: string) {
  fullscreenSvg.value = svgHtml
  fullscreenSource.value = source
  fullscreenOpen.value = true
}

function handleCloseFullscreen() {
  fullscreenOpen.value = false
  fullscreenSvg.value = ''
  fullscreenSource.value = ''
}

// Draft recovery mechanism
const DRAFT_PREFIX = 'online-mermaid-draft:'
const draftAvailable = ref(false)
const draftSavedAt = ref('')
const pendingDraftCode = ref('')

function checkDraft(id: string, originalCode: string) {
  try {
    const raw = window.localStorage.getItem(`${DRAFT_PREFIX}${id}`)
    if (raw) {
      const parsed = JSON.parse(raw)
      if (parsed && typeof parsed.code === 'string' && parsed.code !== originalCode) {
        draftAvailable.value = true
        pendingDraftCode.value = parsed.code
        const d = new Date(parsed.timestamp || Date.now())
        const pad = (n: number) => String(n).padStart(2, '0')
        draftSavedAt.value = `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
        return
      }
    }
  } catch {
    // ignore
  }
  draftAvailable.value = false
  pendingDraftCode.value = ''
}

function restoreDraft() {
  if (pendingDraftCode.value) {
    editableCode.value = pendingDraftCode.value
  }
  draftAvailable.value = false
  showToast('已恢复本地草稿')
}

function discardDraft() {
  if (currentDiagram.value) {
    try {
      window.localStorage.removeItem(`${DRAFT_PREFIX}${currentDiagram.value.id}`)
    } catch {
      // ignore
    }
    editableCode.value = currentDiagram.value.code
  }
  draftAvailable.value = false
  showToast('已放弃本地草稿')
}

watch(editableCode, (newCode) => {
  if (currentDiagram.value) {
    try {
      if (newCode !== currentDiagram.value.code) {
        window.localStorage.setItem(
          `${DRAFT_PREFIX}${currentDiagram.value.id}`,
          JSON.stringify({ code: newCode, timestamp: Date.now() }),
        )
      } else {
        window.localStorage.removeItem(`${DRAFT_PREFIX}${currentDiagram.value.id}`)
        draftAvailable.value = false
      }
    } catch {
      // ignore quota
    }
  }
})

const isDirty = computed(() => {
  if (!currentDiagram.value) return false
  return editableCode.value !== currentDiagram.value.code
})

// Diagram Operations
async function loadDiagramList(selectId?: string) {
  try {
    const list = await fetchDiagrams()
    diagrams.value = list
    if (list.length > 0) {
      const targetId = selectId || currentDiagram.value?.id || list[0].id
      await selectDiagram(targetId)
    }
  } catch (err: any) {
    showToast(err.message || '加载图表列表失败', true)
  }
}

async function selectDiagram(id: string) {
  if (currentDiagram.value?.id === id) return
  try {
    const d = await fetchDiagram(id)
    currentDiagram.value = d
    editableCode.value = d.code
    checkDraft(d.id, d.code)
  } catch (err: any) {
    showToast(err.message || '加载图表失败', true)
  }
}

async function handleCreateNew() {
  try {
    const nextName = await fetchNextName()
    const d = await createDiagram(nextName, '')
    await loadDiagramList(d.id)
    showToast(`已创建新图表: ${d.title}`)
  } catch (err: any) {
    showToast(err.message || '创建图表失败', true)
  }
}

async function handleRename(id: string, newTitle: string) {
  try {
    const isCurrent = currentDiagram.value?.id === id
    const code = isCurrent ? editableCode.value : (await fetchDiagram(id)).code
    await updateDiagram(id, newTitle, code)
    if (isCurrent) currentDiagram.value!.title = newTitle
    await loadDiagramList(id)
    showToast('重命名成功')
  } catch (err: any) {
    showToast(err.message || '重命名失败', true)
  }
}

async function handleDelete(id: string) {
  try {
    await deleteDiagram(id)
    try {
      window.localStorage.removeItem(`${DRAFT_PREFIX}${id}`)
    } catch {
      // ignore
    }
    showToast('已删除图表')
    await loadDiagramList()
  } catch (err: any) {
    showToast(err.message || '删除图表失败', true)
  }
}

async function handleDuplicate(id: string) {
  try {
    const duplicated = await duplicateDiagram(id)
    await loadDiagramList(duplicated.id)
    showToast(`已复制为 "${duplicated.title}"`)
  } catch (err: any) {
    showToast(err.message || '复制图表失败', true)
  }
}

async function handleSave() {
  if (!currentDiagram.value || isSaving.value) return
  isSaving.value = true
  saveError.value = ''
  try {
    const updated = await updateDiagram(
      currentDiagram.value.id,
      currentDiagram.value.title,
      editableCode.value,
    )
    currentDiagram.value = updated
    try {
      window.localStorage.removeItem(`${DRAFT_PREFIX}${updated.id}`)
    } catch {
      // ignore
    }
    draftAvailable.value = false
    saveSuccess.value = true
    showToast('✓ 保存成功')
    setTimeout(() => {
      saveSuccess.value = false
    }, 2000)

    // Update diagram list metadata (updatedAt)
    const idx = diagrams.value.findIndex((d) => d.id === updated.id)
    if (idx !== -1) {
      diagrams.value[idx].updatedAt = updated.updatedAt
    }
  } catch (err: any) {
    saveError.value = err.message || '保存失败'
    showToast(saveError.value, true)
  } finally {
    isSaving.value = false
  }
}

async function handleLogout() {
  if (signingOut.value) return
  signingOut.value = true
  try {
    await logout()
    userMenuOpen.value = false
    emit('signedOut')
  } catch (err: any) {
    showToast(err.message || '退出登录失败', true)
  } finally {
    signingOut.value = false
  }
}

// Global keybindings
function handleGlobalKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    if (userMenuOpen.value) {
      e.preventDefault()
      userMenuOpen.value = false
      return
    }
    if (mobileSidebarOpen.value) {
      e.preventDefault()
      closeSidebar()
      return
    }
  }
  if ((e.ctrlKey || e.metaKey) && e.key === 's') {
    e.preventDefault()
    handleSave()
  }
}

onMounted(async () => {
  handleViewportChange()
  document.addEventListener('click', handleGlobalClick)
  window.addEventListener('resize', handleViewportChange)
  window.addEventListener('keydown', handleGlobalKeydown)
  await loadDiagramList()
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleGlobalClick)
  window.removeEventListener('resize', handleViewportChange)
  window.removeEventListener('keydown', handleGlobalKeydown)
  if (toastTimer) clearTimeout(toastTimer)
})
</script>

<template>
  <div class="workspace-layout">
    <!-- Sidebar (Desktop Column or Mobile Drawer) -->
    <Sidebar
      :diagrams="diagrams"
      :active-id="currentDiagram?.id || ''"
      :collapsed="sidebarCollapsed"
      :is-mobile="mobileViewport"
      :mobile-open="mobileSidebarOpen"
      @select="selectDiagram"
      @create="handleCreateNew"
      @rename="handleRename"
      @delete="handleDelete"
      @duplicate="handleDuplicate"
      @toggle-collapse="sidebarCollapsed = !sidebarCollapsed"
      @close-mobile="closeSidebar"
    />

    <!-- Mobile Drawer Backdrop Overlay -->
    <Transition name="fade">
      <button
        v-if="mobileViewport && mobileSidebarOpen"
        class="drawer-backdrop"
        type="button"
        aria-label="关闭侧边栏"
        @click="closeSidebar"
      ></button>
    </Transition>

    <main class="main-workspace">
      <!-- Top Workspace Header -->
      <header class="app-header">
        <div class="header-section header-start">
          <button
            type="button"
            class="icon-btn menu-btn"
            :aria-expanded="mobileViewport ? mobileSidebarOpen : !sidebarCollapsed"
            :title="mobileViewport ? '打开图表列表' : (sidebarCollapsed ? '展开侧边栏' : '折叠侧边栏')"
            aria-label="切换侧栏"
            @click="toggleSidebar"
          >
            <span v-html="iconSvg('menu', 18)"></span>
          </button>

          <div v-if="currentDiagram" class="diagram-title-block">
            <span class="diagram-tag">图表</span>
            <h1 class="diagram-title" :title="currentDiagram.title">
              {{ currentDiagram.title }}
            </h1>
            <span v-if="isDirty" class="dirty-badge" title="有未保存修改">* 未保存</span>
          </div>
        </div>

        <div class="header-section header-center">
          <div class="view-mode-group" role="group" aria-label="视图模式切换">
            <button
              type="button"
              class="mode-btn"
              :class="{ active: viewMode === 'preview' }"
              title="仅预览图表画布"
              aria-label="预览"
              @click="viewMode = 'preview'"
            >
              👁 <span class="mode-label">预览</span>
            </button>
            <button
              type="button"
              class="mode-btn"
              :class="{ active: viewMode === 'edit' }"
              title="仅代码编辑"
              aria-label="编辑"
              @click="viewMode = 'edit'"
            >
              ✏️ <span class="mode-label">编辑</span>
            </button>
            <button
              type="button"
              class="mode-btn"
              :class="{ active: viewMode === 'split' }"
              title="分屏视图"
              aria-label="分屏"
              @click="viewMode = 'split'"
            >
              📑 <span class="mode-label">分屏</span>
            </button>
          </div>
        </div>

        <div class="header-section header-end">
          <!-- Save Button -->
          <button
            type="button"
            class="save-btn"
            :disabled="!isDirty || isSaving"
            title="保存修改 (Ctrl+S)"
            aria-label="保存"
            @click="handleSave"
          >
            <span v-html="iconSvg('save', 15)"></span>
            <span class="save-label">{{ isSaving ? '保存中…' : '保存' }}</span>
          </button>

          <div class="header-divider"></div>

          <!-- Theme Day/Night Control -->
          <ThemeControl />

          <div class="header-divider"></div>

          <!-- User Menu Dropdown (Replaces standalone exit button) -->
          <div class="user-menu-container">
            <button
              class="user-menu-trigger"
              type="button"
              :aria-expanded="userMenuOpen"
              :title="`用户: ${username}`"
              aria-label="用户菜单"
              @click="userMenuOpen = !userMenuOpen"
            >
              <span class="user-avatar">{{ username.slice(0, 1).toUpperCase() }}</span>
              <span class="user-name">{{ username }}</span>
              <span class="dropdown-caret">▾</span>
            </button>

            <Transition name="dropdown">
              <div v-if="userMenuOpen" class="user-dropdown" role="menu">
                <div class="dropdown-user-header">
                  <div class="user-avatar large">{{ username.slice(0, 1).toUpperCase() }}</div>
                  <div class="user-details">
                    <p class="user-title">{{ username }}</p>
                    <p class="user-status">在线 · SQLite3 已挂载</p>
                    <p v-if="version" class="user-version">{{ version }}</p>
                  </div>
                </div>

                <div class="dropdown-divider"></div>

                <button
                  class="dropdown-item danger"
                  type="button"
                  role="menuitem"
                  :disabled="signingOut"
                  @click="handleLogout"
                >
                  <span v-html="iconSvg('logout', 15)"></span>
                  <span>{{ signingOut ? '退出中…' : '退出登录' }}</span>
                </button>
              </div>
            </Transition>
          </div>
        </div>
      </header>

      <!-- LocalStorage Draft Recovery Banner -->
      <div v-if="draftAvailable" class="draft-banner" role="alert">
        <div class="draft-info">
          <span class="draft-icon">📝</span>
          <span>检测到本地存在未保存草稿（{{ draftSavedAt }}）</span>
        </div>
        <div class="draft-actions">
          <button type="button" class="draft-btn primary" @click="restoreDraft">恢复草稿</button>
          <button type="button" class="draft-btn" @click="discardDraft">放弃</button>
        </div>
      </div>

      <!-- Main Workspace Stage -->
      <div class="workspace-stage" :class="`mode-${viewMode}`">
        <div v-if="viewMode === 'edit' || viewMode === 'split'" class="pane-column editor-pane">
          <DiagramEditor
            v-model="editableCode"
            @save="handleSave"
          />
        </div>

        <div v-if="viewMode === 'preview' || viewMode === 'split'" class="pane-column preview-pane">
          <DiagramPreview
            :code="editableCode"
            :theme="resolved"
            @toast="showToast"
            @open-fullscreen="handleOpenFullscreen"
          />
        </div>
      </div>
    </main>

    <!-- Fullscreen Modal -->
    <FullscreenModal
      :open="fullscreenOpen"
      :svg-html="fullscreenSvg"
      :source-code="fullscreenSource"
      @close="handleCloseFullscreen"
      @toast="showToast"
    />

    <!-- Global Toast Notification -->
    <Teleport to="body">
      <Toast
        :message="toastMessage"
        :is-error="toastIsError"
        :style-obj="{ top: '24px', left: '50%' }"
      />
    </Teleport>
  </div>
</template>

<style scoped>
.workspace-layout {
  position: relative;
  display: flex;
  width: 100vw;
  height: 100vh;
  min-height: 100dvh;
  overflow: hidden;
  background: var(--bg);
}

.drawer-backdrop {
  position: fixed;
  inset: 0;
  z-index: 90;
  width: 100vw;
  height: 100vh;
  border: 0;
  background: rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(4px);
  cursor: pointer;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 180ms ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.main-workspace {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
  height: 100%;
  overflow: hidden;
}

.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: var(--header-height);
  padding: 0 14px;
  background: var(--surface-raised);
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
  gap: 8px;
  padding-top: max(0px, env(safe-area-inset-top));
  padding-left: max(10px, env(safe-area-inset-left));
  padding-right: max(10px, env(safe-area-inset-right));
}

.header-section {
  display: flex;
  align-items: center;
}

.header-start {
  gap: 10px;
  min-width: 0;
  flex: 1;
}

.menu-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--radius-xs);
  color: var(--text-muted);
  flex-shrink: 0;
  transition: all 120ms ease;
}

.menu-btn:hover {
  color: var(--text);
  background: var(--surface-hover);
}

.diagram-title-block {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.diagram-tag {
  font-size: 10px;
  font-weight: 700;
  padding: 2px 6px;
  background: var(--surface-muted);
  border-radius: var(--radius-xs);
  color: var(--text-muted);
  letter-spacing: 0.5px;
  flex-shrink: 0;
}

.diagram-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: clamp(100px, 20vw, 240px);
}

.dirty-badge {
  font-size: 11px;
  font-weight: 600;
  color: var(--accent-strong);
  white-space: nowrap;
  flex-shrink: 0;
}

.header-center {
  flex-shrink: 0;
}

.view-mode-group {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 2px;
  border-radius: var(--radius-sm);
  background: var(--surface-muted);
  border: 1px solid var(--border);
}

.mode-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  min-height: 28px;
  border-radius: var(--radius-xs);
  font-size: 12px;
  font-weight: 500;
  color: var(--text-muted);
  transition: all 120ms ease;
  white-space: nowrap;
}

.mode-btn:hover {
  color: var(--text);
}

.mode-btn.active {
  background: var(--surface-raised);
  color: var(--accent-strong);
  font-weight: 600;
  box-shadow: var(--shadow-sm);
}

.header-end {
  gap: 8px;
  flex: 1;
  justify-content: flex-end;
}

.save-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 5px 12px;
  min-height: 30px;
  border-radius: var(--radius-sm);
  background: var(--accent);
  color: #ffffff;
  font-size: 12px;
  font-weight: 600;
  transition: opacity 120ms, transform 120ms;
  flex-shrink: 0;
}

.save-btn:hover:not(:disabled) {
  opacity: 0.92;
}

.save-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.header-divider {
  width: 1px;
  height: 16px;
  background: var(--border);
  flex-shrink: 0;
}

/* User Menu & Dropdown */
.user-menu-container {
  position: relative;
  display: inline-block;
  flex-shrink: 0;
}

.user-menu-trigger {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 3px 8px 3px 4px;
  background: var(--surface-muted);
  border: 1px solid var(--border);
  border-radius: 16px;
  transition: background-color 120ms ease;
}

.user-menu-trigger:hover {
  background: var(--surface-hover);
  border-color: var(--border-strong);
}

.user-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: var(--accent);
  color: #ffffff;
  font-size: 11px;
  font-weight: 700;
}

.user-avatar.large {
  width: 32px;
  height: 32px;
  font-size: 14px;
}

.user-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--text);
}

.dropdown-caret {
  font-size: 10px;
  color: var(--text-muted);
}

.user-dropdown {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  z-index: 120;
  min-width: 180px;
  padding: 6px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--surface-raised);
  box-shadow: var(--shadow);
}

.dropdown-user-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 8px;
}

.user-details {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.user-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--text);
}

.user-status {
  font-size: 10px;
  color: var(--accent);
  font-weight: 500;
}

.user-version {
  font-size: 10px;
  color: var(--text-faint);
  font-family: var(--font-mono);
  margin-top: 1px;
}

.dropdown-divider {
  height: 1px;
  background: var(--border);
  margin: 4px 0;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 8px 10px;
  border-radius: var(--radius-xs);
  font-size: 13px;
  font-weight: 500;
  color: var(--text);
  text-align: left;
  transition: background 120ms;
}

.dropdown-item:hover {
  background: var(--surface-muted);
}

.dropdown-item.danger {
  color: var(--danger);
}

.dropdown-item.danger:hover {
  background: var(--danger-soft);
}

.dropdown-item:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.dropdown-enter-active,
.dropdown-leave-active {
  transition: opacity 140ms ease, transform 140ms ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

/* Draft banner */
.draft-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 6px 14px;
  background: color-mix(in srgb, var(--accent) 12%, var(--surface-raised));
  border-bottom: 1px solid var(--border);
  font-size: 12px;
  color: var(--text);
  z-index: 5;
}

.draft-info {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.draft-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.draft-btn {
  padding: 3px 10px;
  border-radius: var(--radius-xs);
  font-size: 11px;
  font-weight: 500;
  background: var(--surface);
  border: 1px solid var(--border);
  color: var(--text);
  transition: all 100ms;
}

.draft-btn:hover {
  background: var(--surface-hover);
}

.draft-btn.primary {
  background: var(--accent);
  color: #ffffff;
  border-color: transparent;
}

/* Main Workspace Stage */
.workspace-stage {
  display: flex;
  flex: 1;
  width: 100%;
  min-height: 0;
  padding: 10px;
  gap: 10px;
  background: var(--bg);
  padding-bottom: max(10px, env(safe-area-inset-bottom));
  padding-left: max(10px, env(safe-area-inset-left));
  padding-right: max(10px, env(safe-area-inset-right));
}

.pane-column {
  height: 100%;
  min-width: 0;
  overflow: hidden;
}

.workspace-stage.mode-preview .preview-pane,
.workspace-stage.mode-edit .editor-pane {
  width: 100%;
  flex: 1;
}

.workspace-stage.mode-split .editor-pane,
.workspace-stage.mode-split .preview-pane {
  flex: 1;
  width: 50%;
}

/* Mobile Responsive Adaptations */
@media (max-width: 840px) {
  .workspace-stage {
    padding: 6px;
    gap: 6px;
  }

  .workspace-stage.mode-split {
    flex-direction: column;
  }

  .workspace-stage.mode-split .editor-pane,
  .workspace-stage.mode-split .preview-pane {
    width: 100%;
    height: 50%;
  }

  .diagram-title-block {
    display: none !important;
  }

  .user-name {
    display: none;
  }

  .user-menu-trigger {
    padding: 2px 4px 2px 2px;
  }
}

@media (max-width: 600px) {
  .diagram-tag {
    display: none;
  }

  .save-btn {
    padding: 4px 8px;
  }

  .save-label {
    display: none;
  }

  .mode-label {
    display: none;
  }

  .mode-btn {
    padding: 4px 8px;
    font-size: 14px;
  }
}
</style>
