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

const props = defineProps<{ username: string }>()
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
  return 'split' // default split mode for great mermaid editing experience
}

const viewMode = ref<ViewMode>(getInitialViewMode())
watch(viewMode, (mode) => {
  try {
    window.localStorage.setItem(VIEW_MODE_KEY, mode)
  } catch {
    // ignore
  }
})

const sidebarCollapsed = ref(false)
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
  try {
    await logout()
    emit('signedOut')
  } catch (err: any) {
    showToast(err.message || '退出登录失败', true)
  }
}

// Global keybindings
function handleGlobalKeydown(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key === 's') {
    e.preventDefault()
    handleSave()
  }
}

onMounted(async () => {
  window.addEventListener('keydown', handleGlobalKeydown)
  await loadDiagramList()
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleGlobalKeydown)
  if (toastTimer) clearTimeout(toastTimer)
})
</script>

<template>
  <div class="workspace-layout">
    <Sidebar
      :diagrams="diagrams"
      :active-id="currentDiagram?.id || ''"
      :collapsed="sidebarCollapsed"
      @select="selectDiagram"
      @create="handleCreateNew"
      @rename="handleRename"
      @delete="handleDelete"
      @duplicate="handleDuplicate"
      @toggle-collapse="sidebarCollapsed = !sidebarCollapsed"
    />

    <main class="main-workspace">
      <!-- Top Workspace Header -->
      <header class="top-header">
        <div class="header-left">
          <div v-if="currentDiagram" class="diagram-title-box">
            <span class="diagram-tag">图表</span>
            <h1 class="diagram-heading" :title="currentDiagram.title">
              {{ currentDiagram.title }}
            </h1>
            <span v-if="isDirty" class="dirty-indicator">* 已修改</span>
          </div>
        </div>

        <div class="header-center">
          <div class="view-mode-group" role="group" aria-label="视图模式切换">
            <button
              type="button"
              class="mode-btn"
              :class="{ active: viewMode === 'preview' }"
              title="仅预览图表"
              @click="viewMode = 'preview'"
            >
              👁 预览
            </button>
            <button
              type="button"
              class="mode-btn"
              :class="{ active: viewMode === 'edit' }"
              title="仅代码编辑"
              @click="viewMode = 'edit'"
            >
              ✏️ 编辑
            </button>
            <button
              type="button"
              class="mode-btn"
              :class="{ active: viewMode === 'split' }"
              title="左右分屏实时渲染"
              @click="viewMode = 'split'"
            >
              📑 分屏
            </button>
          </div>
        </div>

        <div class="header-right">
          <button
            type="button"
            class="save-btn"
            :disabled="!isDirty || isSaving"
            title="保存当前修改 (Ctrl+S)"
            @click="handleSave"
          >
            <span v-html="iconSvg('save', 14)"></span>
            <span>{{ isSaving ? '保存中…' : '保存' }}</span>
          </button>

          <div class="divider"></div>

          <ThemeControl />

          <div class="divider"></div>

          <div class="user-badge" :title="`当前登录用户: ${username}`">
            <span class="user-avatar">{{ username.slice(0, 1).toUpperCase() }}</span>
            <span class="username-text">{{ username }}</span>
          </div>

          <button
            type="button"
            class="icon-btn logout-btn"
            title="退出登录"
            @click="handleLogout"
          >
            <span v-html="iconSvg('logout', 14)"></span>
          </button>
        </div>
      </header>

      <!-- LocalStorage Draft Recovery Banner -->
      <div v-if="draftAvailable" class="draft-banner" role="alert">
        <div class="draft-info">
          <span class="draft-icon">📝</span>
          <span>检测到本地存在未保存的草稿（保存于 {{ draftSavedAt }}）</span>
        </div>
        <div class="draft-actions">
          <button type="button" class="draft-btn primary" @click="restoreDraft">恢复草稿</button>
          <button type="button" class="draft-btn" @click="discardDraft">放弃草稿</button>
        </div>
      </div>

      <!-- Main Stage -->
      <div class="stage-container" :class="`mode-${viewMode}`">
        <div v-if="viewMode === 'edit' || viewMode === 'split'" class="pane-column editor-col">
          <DiagramEditor
            v-model="editableCode"
            @save="handleSave"
          />
        </div>

        <div v-if="viewMode === 'preview' || viewMode === 'split'" class="pane-column preview-col">
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

    <!-- Global Toast notification -->
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
  display: flex;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  background: var(--bg);
}

.main-workspace {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
  height: 100%;
  overflow: hidden;
}

.top-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: var(--header-height);
  padding: 0 16px;
  background: var(--surface-raised);
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
  gap: 12px;
}

.header-left {
  display: flex;
  align-items: center;
  min-width: 0;
  flex: 1;
}

.diagram-title-box {
  display: flex;
  align-items: center;
  gap: 8px;
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
}

.diagram-heading {
  font-size: 14px;
  font-weight: 600;
  color: var(--text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.dirty-indicator {
  font-size: 11px;
  font-weight: 600;
  color: var(--accent-strong);
}

.header-center {
  display: flex;
  align-items: center;
}

.view-mode-group {
  display: flex;
  align-items: center;
  gap: 3px;
  padding: 3px;
  border-radius: var(--radius-sm);
  background: var(--surface-muted);
  border: 1px solid var(--border);
}

.mode-btn {
  padding: 3px 12px;
  border-radius: var(--radius-xs);
  font-size: 12px;
  font-weight: 500;
  color: var(--text-muted);
  transition: all 120ms ease;
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

.header-right {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  justify-content: flex-end;
}

.save-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  border-radius: var(--radius-sm);
  background: var(--accent);
  color: #ffffff;
  font-size: 12px;
  font-weight: 600;
  transition: opacity 120ms;
}

.save-btn:hover:not(:disabled) {
  opacity: 0.9;
}

.save-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.divider {
  width: 1px;
  height: 18px;
  background: var(--border);
}

.user-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 2px 8px 2px 2px;
  background: var(--surface-muted);
  border-radius: 14px;
}

.user-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: var(--accent-strong);
  color: #ffffff;
  font-size: 11px;
  font-weight: 700;
}

.username-text {
  font-size: 12px;
  font-weight: 500;
  color: var(--text);
}

.icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border-radius: var(--radius-xs);
  color: var(--text-muted);
  transition: all 120ms;
}

.icon-btn:hover {
  color: var(--text);
  background: var(--surface-hover);
}

.logout-btn:hover {
  color: var(--danger);
}

.draft-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 6px 16px;
  background: color-mix(in srgb, var(--accent) 12%, var(--surface-raised));
  border-bottom: 1px solid var(--border);
  font-size: 12px;
  color: var(--text);
  z-index: 5;
}

.draft-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.draft-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.draft-btn {
  padding: 2px 8px;
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

.stage-container {
  display: flex;
  flex: 1;
  width: 100%;
  min-height: 0;
  padding: 12px;
  gap: 12px;
  background: var(--bg);
}

.pane-column {
  height: 100%;
  min-width: 0;
  overflow: hidden;
}

.stage-container.mode-preview .preview-col,
.stage-container.mode-edit .editor-col {
  width: 100%;
  flex: 1;
}

.stage-container.mode-split .editor-col,
.stage-container.mode-split .preview-col {
  flex: 1;
  width: 50%;
}
</style>
