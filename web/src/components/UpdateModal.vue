<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import { applySystemUpdate, checkSystemUpdate } from '@/api/client'
import type { UpdateCheckResponse } from '@/api/types'
import { iconSvg } from '@/utils/icons'

const props = defineProps<{
  open: boolean
  currentVersion?: string
}>()

const emit = defineEmits<{
  close: []
  toast: [msg: string, isError?: boolean]
}>()

type Status = 'idle' | 'checking' | 'uptodate' | 'available' | 'updating' | 'restarting' | 'error'

const status = ref<Status>('idle')
const updateInfo = ref<UpdateCheckResponse | null>(null)
const errorMsg = ref('')
const copySuccess = ref(false)
const restartCountdown = ref(30)
let pollTimer: number | null = null

async function doCheck(refresh = false) {
  status.value = 'checking'
  errorMsg.value = ''
  try {
    const res = await checkSystemUpdate(refresh)
    updateInfo.value = res
    if (res.has_update) {
      status.value = 'available'
    } else {
      status.value = 'uptodate'
    }
  } catch (err: unknown) {
    status.value = 'error'
    errorMsg.value = err instanceof Error ? err.message : '检查更新失败，请稍后重试'
  }
}

async function handleApplyUpdate() {
  if (!updateInfo.value) return
  status.value = 'updating'
  errorMsg.value = ''

  try {
    await applySystemUpdate(updateInfo.value.asset_url)
    // Server scheduled restart
    status.value = 'restarting'
    startPollingServer()
  } catch (err: unknown) {
    status.value = 'error'
    errorMsg.value = err instanceof Error ? err.message : '应用更新失败，请查看日志或手动更新'
  }
}

function startPollingServer() {
  const pollIntervalSec = 1.5
  restartCountdown.value = 30
  let attempts = 0
  const maxAttempts = 20

  pollTimer = window.setInterval(async () => {
    attempts++
    // Count down in real seconds: each tick fires every pollIntervalSec.
    restartCountdown.value = Math.max(0, 30 - attempts * pollIntervalSec)

    try {
      const resp = await fetch('/api/auth/session', { credentials: 'same-origin' })
      if (resp.ok) {
        if (pollTimer) clearInterval(pollTimer)
        emit('toast', '服务更新成功，正在刷新页面…')
        setTimeout(() => {
          window.location.reload()
        }, 1200)
        return
      }
    } catch {
      // Expected while restarting
    }

    if (attempts >= maxAttempts) {
      if (pollTimer) clearInterval(pollTimer)
      errorMsg.value = '服务重启超时，请检查控制台或手动刷新页面。'
      status.value = 'error'
    }
  }, 1500)
}

async function copyCommand(cmd: string) {
  try {
    await navigator.clipboard.writeText(cmd)
    copySuccess.value = true
    emit('toast', '命令已复制到剪切板')
    setTimeout(() => {
      copySuccess.value = false
    }, 2000)
  } catch {
    emit('toast', '复制失败，请手动选取复制', true)
  }
}

function handleClose() {
  if (status.value === 'updating' || status.value === 'restarting') {
    return
  }
  emit('close')
}

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      doCheck(false)
    } else {
      if (pollTimer) clearInterval(pollTimer)
      status.value = 'idle'
      errorMsg.value = ''
    }
  }
)

onBeforeUnmount(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<template>
  <Transition name="modal-fade">
    <div v-if="open" class="modal-backdrop" @click.self="handleClose">
      <div class="modal-card" role="dialog" aria-modal="true" aria-labelledby="update-modal-title">
        <!-- Header -->
        <div class="modal-header">
          <div class="header-title-wrap">
            <span class="header-icon" v-html="iconSvg('refresh-cw', 18)"></span>
            <h2 id="update-modal-title" class="header-title">系统更新</h2>
          </div>
          <button
            type="button"
            class="close-btn"
            title="关闭"
            aria-label="关闭"
            :disabled="status === 'updating' || status === 'restarting'"
            @click="handleClose"
          >
            <span v-html="iconSvg('x', 16)"></span>
          </button>
        </div>

        <!-- Body -->
        <div class="modal-body">
          <!-- Environment Info Bar -->
          <div class="env-bar">
            <div class="env-item">
              <span class="env-label">当前版本</span>
              <span class="env-value badge">{{ currentVersion || updateInfo?.current_version || 'v1.1.1' }}</span>
            </div>
            <div v-if="updateInfo?.os" class="env-item">
              <span class="env-label">系统环境</span>
              <span class="env-value">{{ updateInfo.os }} / {{ updateInfo.arch }}</span>
            </div>
          </div>

          <!-- 1. Checking State -->
          <div v-if="status === 'checking'" class="state-container checking">
            <div class="spinner-ring" aria-hidden="true"></div>
            <p class="state-text">正在连接 GitHub 检查最新发布版本…</p>
          </div>

          <!-- 2. Up To Date State -->
          <div v-else-if="status === 'uptodate'" class="state-container uptodate">
            <div class="state-icon-wrap success">
              <span v-html="iconSvg('check-circle', 36)"></span>
            </div>
            <h3 class="state-heading">当前已是最新版本</h3>
            <p class="state-subtext">恭喜！您正在运行最新版本的 Online Mermaid。</p>
            <div class="action-footer">
              <button type="button" class="btn secondary" @click="doCheck(true)">
                <span v-html="iconSvg('refresh-cw', 14)"></span>
                <span>重新检查</span>
              </button>
              <button type="button" class="btn primary" @click="handleClose">完成</button>
            </div>
          </div>

          <!-- 3. Update Available State -->
          <div v-else-if="status === 'available' && updateInfo" class="state-container available">
            <div class="release-header-card">
              <div class="release-meta">
                <span class="new-tag">发现新版本</span>
                <span class="version-tag">{{ updateInfo.latest_version }}</span>
                <span v-if="updateInfo.published_at" class="date-tag">
                  {{ new Date(updateInfo.published_at).toLocaleDateString() }}
                </span>
              </div>
              <a
                :href="updateInfo.release_url"
                target="_blank"
                rel="noopener noreferrer"
                class="external-link"
                title="在 GitHub 上查看此版本"
              >
                <span>Release 详情</span>
                <span v-html="iconSvg('external-link', 13)"></span>
              </a>
            </div>

            <!-- Release Notes -->
            <div v-if="updateInfo.release_notes" class="release-notes-box">
              <div class="notes-header">更新日志：</div>
              <pre class="notes-content">{{ updateInfo.release_notes }}</pre>
            </div>

            <!-- Cannot Auto Update Alert -->
            <div v-if="!updateInfo.can_auto_update" class="alert-box warn">
              <span class="alert-icon" v-html="iconSvg('alert-circle', 16)"></span>
              <div class="alert-body">
                <p class="alert-title">自动更新受限</p>
                <p class="alert-desc">{{ updateInfo.update_message || '当前目录无写入权限或未找到兼容安装包。' }}</p>
              </div>
            </div>

            <!-- Manual Command Snippet -->
            <div v-if="updateInfo.manual_command" class="cmd-box">
              <div class="cmd-title-row">
                <span class="cmd-label">手动升级命令 (终端执行)：</span>
                <button
                  type="button"
                  class="copy-btn"
                  :title="copySuccess ? '已复制' : '复制命令'"
                  @click="copyCommand(updateInfo?.manual_command || '')"
                >
                  <span v-html="iconSvg(copySuccess ? 'check' : 'copy', 13)"></span>
                  <span>{{ copySuccess ? '已复制' : '复制' }}</span>
                </button>
              </div>
              <code class="cmd-code">{{ updateInfo.manual_command }}</code>
            </div>

            <div class="action-footer">
              <button type="button" class="btn secondary" @click="handleClose">暂不更新</button>
              <button
                v-if="updateInfo.can_auto_update"
                type="button"
                class="btn primary pulse-btn"
                @click="handleApplyUpdate"
              >
                <span v-html="iconSvg('refresh-cw', 14)"></span>
                <span>一键升级并重启</span>
              </button>
              <a
                v-else
                :href="updateInfo.release_url"
                target="_blank"
                rel="noopener noreferrer"
                class="btn primary"
              >
                前往下载
              </a>
            </div>
          </div>

          <!-- 4. Updating State -->
          <div v-else-if="status === 'updating'" class="state-container updating">
            <div class="spinner-ring large" aria-hidden="true"></div>
            <h3 class="state-heading">正在下载并应用更新</h3>
            <p class="state-subtext">正在从 GitHub 获取目标平台的二进制包并校验 SHA256 完整性…</p>
            <p class="state-notice">⚠️ 请勿刷新或关闭网页</p>
          </div>

          <!-- 5. Restarting State -->
          <div v-else-if="status === 'restarting'" class="state-container restarting">
            <div class="spinner-ring large success" aria-hidden="true"></div>
            <h3 class="state-heading">更新成功，服务正在重启</h3>
            <p class="state-subtext">已完成二进制文件原子替换，正在等待新服务恢复上线…</p>
            <div class="reboot-indicator">
              <span>状态探测中 ({{ Math.ceil(restartCountdown) }}s)</span>
            </div>
          </div>

          <!-- 6. Error State -->
          <div v-else-if="status === 'error'" class="state-container error">
            <div class="state-icon-wrap error">
              <span v-html="iconSvg('alert-circle', 36)"></span>
            </div>
            <h3 class="state-heading">更新失败</h3>
            <p class="state-subtext error-desc">{{ errorMsg }}</p>

            <div v-if="updateInfo?.manual_command" class="cmd-box error-fallback">
              <div class="cmd-title-row">
                <span class="cmd-label">您可尝试通过下方命令手动更新：</span>
                <button
                  type="button"
                  class="copy-btn"
                  @click="copyCommand(updateInfo?.manual_command || '')"
                >
                  <span v-html="iconSvg(copySuccess ? 'check' : 'copy', 13)"></span>
                  <span>{{ copySuccess ? '已复制' : '复制' }}</span>
                </button>
              </div>
              <code class="cmd-code">{{ updateInfo.manual_command }}</code>
            </div>

            <div class="action-footer">
              <button type="button" class="btn secondary" @click="handleClose">关闭</button>
              <button type="button" class="btn primary" @click="doCheck(true)">
                <span v-html="iconSvg('refresh-cw', 14)"></span>
                <span>重试</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.modal-backdrop {
  position: fixed;
  inset: 0;
  z-index: 10000;
  background: rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.modal-card {
  width: 100%;
  max-width: 520px;
  background: var(--surface-raised);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  animation: modal-pop 200ms cubic-bezier(0.16, 1, 0.3, 1);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  border-bottom: 1px solid var(--border);
  background: var(--surface);
}

.header-title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-icon {
  display: flex;
  color: var(--accent);
}

.header-title {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text);
}

.close-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: var(--radius-xs);
  border: none;
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 120ms;
}

.close-btn:hover:not(:disabled) {
  background: var(--surface-hover);
  color: var(--text);
}

.close-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.modal-body {
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  max-height: calc(85vh - 70px);
  overflow-y: auto;
}

/* Env Info Bar */
.env-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: var(--surface-muted);
  border-radius: var(--radius-sm);
  font-size: 12px;
}

.env-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.env-label {
  color: var(--text-muted);
}

.env-value {
  color: var(--text);
  font-weight: 500;
}

.env-value.badge {
  background: var(--surface);
  border: 1px solid var(--border);
  padding: 1px 6px;
  border-radius: var(--radius-xs);
  font-family: var(--font-mono);
  font-size: 11px;
}

/* State Containers */
.state-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 16px 0 8px;
}

.state-container.available {
  align-items: stretch;
  text-align: left;
  padding: 0;
  gap: 12px;
}

.state-heading {
  margin: 12px 0 4px;
  font-size: 16px;
  font-weight: 600;
  color: var(--text);
}

.state-subtext {
  margin: 0;
  font-size: 13px;
  color: var(--text-muted);
  line-height: 1.5;
}

.state-subtext.error-desc {
  color: var(--danger);
  word-break: break-word;
}

.state-notice {
  margin-top: 10px;
  font-size: 12px;
  font-weight: 500;
  color: var(--accent);
}

.state-icon-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 54px;
  height: 54px;
  border-radius: 50%;
}

.state-icon-wrap.success {
  background: var(--accent-soft);
  color: var(--accent);
}

.state-icon-wrap.error {
  background: var(--danger-soft);
  color: var(--danger);
}

/* Spinner */
.spinner-ring {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border);
  border-top-color: var(--accent);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.spinner-ring.large {
  width: 44px;
  height: 44px;
}

.spinner-ring.success {
  border-top-color: var(--accent);
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* Release Header Card */
.release-header-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  background: color-mix(in srgb, var(--accent) 8%, var(--surface-raised));
  border: 1px solid color-mix(in srgb, var(--accent) 25%, transparent);
  border-radius: var(--radius-sm);
}

.release-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.new-tag {
  font-size: 11px;
  font-weight: 600;
  color: #ffffff;
  background: var(--accent);
  padding: 2px 7px;
  border-radius: var(--radius-xs);
}

.version-tag {
  font-size: 14px;
  font-weight: 700;
  color: var(--text);
  font-family: var(--font-mono);
}

.date-tag {
  font-size: 11px;
  color: var(--text-muted);
}

.external-link {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--accent);
  text-decoration: none;
  font-weight: 500;
}

.external-link:hover {
  text-decoration: underline;
}

/* Release Notes Box */
.release-notes-box {
  border: 1px solid var(--border);
  background: var(--surface);
  border-radius: var(--radius-sm);
  padding: 10px 12px;
}

.notes-header {
  font-size: 12px;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 6px;
}

.notes-content {
  margin: 0;
  font-family: var(--font-sans);
  font-size: 12px;
  line-height: 1.5;
  color: var(--text-muted);
  max-height: 140px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-word;
}

/* Alert Box */
.alert-box {
  display: flex;
  gap: 10px;
  padding: 10px 12px;
  border-radius: var(--radius-sm);
  background: var(--surface-muted);
  border: 1px solid var(--border);
}

.alert-box.warn {
  border-color: color-mix(in srgb, var(--danger) 30%, transparent);
  background: var(--danger-soft);
  color: var(--danger);
}

.alert-body {
  font-size: 12px;
  line-height: 1.4;
}

.alert-title {
  margin: 0 0 2px;
  font-weight: 600;
}

.alert-desc {
  margin: 0;
  opacity: 0.9;
}

/* Command Snippet */
.cmd-box {
  background: var(--surface-muted);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 8px 12px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.cmd-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.cmd-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
}

.copy-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 2px 7px;
  border-radius: var(--radius-xs);
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text);
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  transition: all 120ms;
}

.copy-btn:hover {
  background: var(--surface-hover);
}

.cmd-code {
  font-family: var(--font-mono);
  font-size: 11px;
  padding: 6px 8px;
  background: var(--surface-raised);
  border: 1px solid var(--border);
  border-radius: var(--radius-xs);
  color: var(--text);
  word-break: break-all;
  user-select: all;
}

/* Footer & Buttons */
.action-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 10px;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 7px 16px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 120ms;
  text-decoration: none;
}

.btn.primary {
  background: var(--accent);
  color: #ffffff;
}

.btn.primary:hover {
  background: var(--accent-strong);
}

.btn.secondary {
  background: var(--surface);
  border-color: var(--border);
  color: var(--text);
}

.btn.secondary:hover {
  background: var(--surface-hover);
}

.pulse-btn {
  box-shadow: 0 0 0 0 color-mix(in srgb, var(--accent) 60%, transparent);
  animation: pulse-glow 2s infinite;
}

@keyframes pulse-glow {
  0% {
    box-shadow: 0 0 0 0 color-mix(in srgb, var(--accent) 60%, transparent);
  }
  70% {
    box-shadow: 0 0 0 8px transparent;
  }
  100% {
    box-shadow: 0 0 0 0 transparent;
  }
}

.reboot-indicator {
  margin-top: 12px;
  padding: 6px 14px;
  background: var(--surface-muted);
  border-radius: var(--radius-xs);
  font-size: 12px;
  font-family: var(--font-mono);
  color: var(--text-muted);
}

/* Transitions */
.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 160ms ease;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}

@keyframes modal-pop {
  from {
    transform: scale(0.96) translateY(4px);
    opacity: 0;
  }
  to {
    transform: scale(1) translateY(0);
    opacity: 1;
  }
}
</style>
