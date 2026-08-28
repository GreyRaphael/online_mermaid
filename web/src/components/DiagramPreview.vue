<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'
import DOMPurify from 'dompurify'
import panzoom, { type PanZoom } from 'panzoom'
import type { ResolvedTheme } from '@/composables/useTheme'
import { iconSvg } from '@/utils/icons'
import { exportPngImage, exportSvgImage, svgToPngBlob } from '@/utils/export'

const props = defineProps<{
  code: string
  theme: ResolvedTheme
}>()

const emit = defineEmits<{
  toast: [message: string, isError?: boolean]
  openFullscreen: [svgHtml: string, sourceCode: string]
}>()

const previewContainerRef = ref<HTMLElement | null>(null)
const outputRef = ref<HTMLElement | null>(null)
const renderedSvg = ref<string>('')
const renderError = ref<string>('')
const errorLine = ref<string>('')

let pzInstance: PanZoom | null = null
let renderSeq = 0
let debounceTimer: number | null = null

function getMermaidFontFamily(): string {
  if (typeof window === 'undefined') return 'ui-sans-serif, system-ui, sans-serif'
  return (
    getComputedStyle(document.documentElement).getPropertyValue('--font-sans').trim() ||
    'ui-sans-serif, system-ui, sans-serif'
  )
}

function destroyPanzoom() {
  if (pzInstance) {
    pzInstance.dispose()
    pzInstance = null
  }
}

function preserveMermaidSize(output: HTMLElement): void {
  const svg = output.querySelector<SVGSVGElement>('svg')
  const viewBox = svg?.viewBox?.baseVal
  if (!svg || !viewBox || viewBox.width <= 0 || viewBox.height <= 0) return

  svg.style.width = '100%'
  svg.style.height = '100%'
  svg.style.maxWidth = '100%'
  svg.style.maxHeight = '100%'
}

function initPanzoom() {
  destroyPanzoom()
  const output = outputRef.value
  if (!output) return
  const svg = output.querySelector<SVGSVGElement>('svg')
  if (!svg) return

  pzInstance = panzoom(svg, {
    maxZoom: 10,
    minZoom: 0.1,
    bounds: true,
    boundsPadding: 0.1,
  })
}

function resetZoom() {
  if (pzInstance) {
    pzInstance.moveTo(0, 0)
    pzInstance.zoomAbs(0, 0, 1)
  }
}

function handleZoom(type: 'in' | 'out') {
  if (!pzInstance || !outputRef.value) return
  const rect = outputRef.value.getBoundingClientRect()
  const cx = rect.width / 2
  const cy = rect.height / 2
  if (type === 'in') pzInstance.smoothZoom(cx, cy, 1.25)
  else pzInstance.smoothZoom(cx, cy, 1 / 1.25)
}

async function renderDiagram(sourceCode: string) {
  const currentSeq = ++renderSeq
  if (!sourceCode || !sourceCode.trim()) {
    renderedSvg.value = ''
    renderError.value = ''
    errorLine.value = ''
    destroyPanzoom()
    if (outputRef.value) outputRef.value.innerHTML = ''
    return
  }

  try {
    const mermaidModule = await import('mermaid')
    if (currentSeq !== renderSeq) return
    const mermaid = mermaidModule.default

    mermaid.initialize({
      startOnLoad: false,
      securityLevel: 'strict',
      theme: props.theme === 'night' ? 'dark' : 'default',
      htmlLabels: false,
      flowchart: { htmlLabels: false, curve: 'rounded' },
      fontFamily: getMermaidFontFamily(),
      suppressErrorRendering: true,
    })

    const id = `online-mermaid-${currentSeq}`
    const result = await mermaid.render(id, sourceCode)
    if (currentSeq !== renderSeq) return

    const sanitized = DOMPurify.sanitize(result.svg, {
      USE_PROFILES: { svg: true, svgFilters: true },
      ADD_TAGS: ['style'],
      ADD_ATTR: ['dominant-baseline', 'alignment-baseline'],
      FORBID_TAGS: ['script', 'foreignObject', 'iframe', 'object', 'embed'],
    })

    renderedSvg.value = sanitized
    renderError.value = ''
    errorLine.value = ''

    await nextTick()
    const output = outputRef.value
    if (output) {
      output.innerHTML = sanitized
      preserveMermaidSize(output)
      initPanzoom()
    }
  } catch (err: any) {
    if (currentSeq !== renderSeq) return
    const msg = err?.message || String(err)
    renderError.value = msg

    const lineMatch = msg.match(/on line (\d+)/i) || msg.match(/line (\d+)/i)
    errorLine.value = lineMatch ? `第 ${lineMatch[1]} 行解析错误` : 'Mermaid 语法解析错误'
    destroyPanzoom()
    if (outputRef.value) outputRef.value.innerHTML = ''
  }
}

function scheduleRender(codeText: string, delay = 200) {
  if (debounceTimer !== null) {
    clearTimeout(debounceTimer)
  }
  debounceTimer = window.setTimeout(() => {
    debounceTimer = null
    void renderDiagram(codeText)
  }, delay)
}

watch(
  () => [props.code, props.theme],
  ([newCode]) => {
    scheduleRender(newCode as string)
  },
  { immediate: true },
)

async function copyMermaidSource() {
  if (!props.code) return
  try {
    await navigator.clipboard.writeText(props.code)
    emit('toast', '已复制 Mermaid 源码')
  } catch {
    emit('toast', '复制源码失败', true)
  }
}

async function copyPng() {
  const svg = outputRef.value?.querySelector('svg')
  if (!svg) return
  try {
    const blob = await svgToPngBlob(svg, { transparent: false })
    await navigator.clipboard.write([new ClipboardItem({ 'image/png': blob })])
    emit('toast', '已复制 PNG 图片 (白底)')
  } catch {
    emit('toast', '复制图片受限或失败', true)
  }
}

async function downloadPng() {
  const svg = outputRef.value?.querySelector('svg')
  if (!svg) return
  try {
    await exportPngImage(svg, 'mermaid-diagram.png', { transparent: true })
    emit('toast', '已下载透明 PNG 图片')
  } catch {
    emit('toast', '导出 PNG 失败', true)
  }
}

function downloadSvg() {
  const svg = outputRef.value?.querySelector('svg')
  if (!svg) return
  try {
    exportSvgImage(svg, 'mermaid-diagram.svg')
    emit('toast', '已下载 SVG 矢量图')
  } catch {
    emit('toast', '导出 SVG 失败', true)
  }
}

function triggerFullscreen() {
  const svgHtml = outputRef.value?.innerHTML || renderedSvg.value
  if (!svgHtml) return
  emit('openFullscreen', svgHtml, props.code)
}

onBeforeUnmount(() => {
  if (debounceTimer !== null) {
    clearTimeout(debounceTimer)
  }
  destroyPanzoom()
})
</script>

<template>
  <div ref="previewContainerRef" class="preview-container">
    <div class="preview-toolbar">
      <div class="toolbar-left">
        <span class="preview-tag">👁 实时预览</span>
      </div>

      <div class="toolbar-actions">
        <button type="button" class="tool-btn" title="放大" @click="handleZoom('in')">
          <span v-html="iconSvg('zoom-in', 15)"></span>
        </button>
        <button type="button" class="tool-btn" title="缩小" @click="handleZoom('out')">
          <span v-html="iconSvg('zoom-out', 15)"></span>
        </button>
        <button type="button" class="tool-btn" title="重置视角" @click="resetZoom">
          <span v-html="iconSvg('rotate-ccw', 15)"></span>
        </button>

        <div class="toolbar-sep"></div>

        <button type="button" class="tool-btn" title="复制 Mermaid 源码" @click="copyMermaidSource">
          <span v-html="iconSvg('file-code', 15)"></span>
        </button>
        <button type="button" class="tool-btn" title="复制 PNG (白底)" @click="copyPng">
          <span v-html="iconSvg('copy', 15)"></span>
        </button>
        <button type="button" class="tool-btn" title="导出透明 PNG" aria-label="导出 PNG" @click="downloadPng">
          <span v-html="iconSvg('image', 15)"></span>
        </button>
        <button type="button" class="tool-btn" title="导出 SVG 矢量图" aria-label="导出 SVG" @click="downloadSvg">
          <span v-html="iconSvg('svg', 15)"></span>
        </button>

        <div class="toolbar-sep"></div>

        <button type="button" class="tool-btn highlight" title="全屏查看" aria-label="全屏查看" @click="triggerFullscreen">
          <span v-html="iconSvg('maximize', 15)"></span>
        </button>
      </div>
    </div>

    <div class="preview-canvas-wrapper">
      <div v-show="!renderError" class="canvas-viewport">
        <div v-show="renderedSvg" ref="outputRef" class="mermaid-output" role="img" aria-label="Mermaid 图表"></div>
        <div v-if="!renderedSvg" class="empty-state">
          <p>暂无图表内容，请在编辑器中输入 Mermaid 代码</p>
        </div>
      </div>

      <div v-if="renderError" class="error-container" role="alert">
        <div class="error-header">
          <span class="err-icon">⚠️</span>
          <span class="err-title">{{ errorLine }}</span>
        </div>
        <div class="error-msg-box">
          <pre>{{ renderError }}</pre>
        </div>
        <div class="error-code-box">
          <div class="code-title">当前源码：</div>
          <pre><code>{{ code }}</code></pre>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.preview-container {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  background: var(--surface);
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  overflow: hidden;
  position: relative;
}

.preview-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 10px;
  background: var(--surface-muted);
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
  gap: 8px;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
}

.preview-toolbar::-webkit-scrollbar {
  display: none;
}

.toolbar-left {
  flex-shrink: 0;
}

.preview-tag {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  letter-spacing: 0.5px;
  white-space: nowrap;
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

.toolbar-sep {
  width: 1px;
  height: 14px;
  background: var(--border);
  margin: 0 2px;
}

.tool-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: var(--radius-xs);
  color: var(--text-muted);
  background: var(--surface);
  border: 1px solid var(--border);
  transition: all 120ms ease;
  flex-shrink: 0;
}

.tool-btn:hover {
  color: var(--text);
  background: var(--surface-hover);
  border-color: var(--border-strong);
}

.tool-btn.highlight {
  color: var(--accent-strong);
  border-color: var(--accent-soft);
  background: var(--accent-soft);
}

.tool-btn.highlight:hover {
  filter: brightness(0.95);
}

.preview-canvas-wrapper {
  flex: 1;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  position: relative;
  background: var(--surface);
}

.canvas-viewport {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.mermaid-output {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  overflow: hidden;
  font-family: var(--font-sans);
}

.mermaid-output :deep(svg) {
  display: block;
  width: 100%;
  height: 100%;
  max-width: 100%;
  max-height: 100%;
  margin: auto;
  overflow: visible;
  cursor: grab;
  user-select: none;
  touch-action: none;
}

.mermaid-output :deep(svg:active) {
  cursor: grabbing;
}

.empty-state {
  color: var(--text-faint);
  font-size: 13px;
  text-align: center;
  padding: 16px;
}

.error-container {
  padding: 16px;
  height: 100%;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  background: var(--surface);
}

.error-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}

.err-icon {
  font-size: 18px;
}

.err-title {
  font-weight: 600;
  color: var(--danger);
  font-size: 13px;
}

.error-msg-box {
  background: var(--danger-soft);
  color: var(--danger);
  padding: 10px 14px;
  border-radius: var(--radius-xs);
  font-family: var(--font-mono);
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-word;
  margin-bottom: 14px;
}

.error-code-box {
  background: var(--surface-muted);
  border: 1px solid var(--border);
  border-radius: var(--radius-xs);
  padding: 12px;
}

.code-title {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 6px;
}

.error-code-box pre {
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--text);
  white-space: pre-wrap;
}

@media (max-width: 600px) {
  .tool-btn {
    width: 32px;
    height: 32px;
  }
}
</style>
