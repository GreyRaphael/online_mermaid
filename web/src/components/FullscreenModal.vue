<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'
import panzoom, { type PanZoom } from 'panzoom'
import { ICON_PATHS, iconSvg } from '@/utils/icons'
import { exportPngImage, exportSvgImage, svgToPngBlob } from '@/utils/export'

const props = defineProps<{
  svgHtml: string
  sourceCode: string
  open: boolean
}>()

const emit = defineEmits<{
  close: []
  toast: [msg: string, isError?: boolean]
}>()

const dialogRef = ref<HTMLDialogElement | null>(null)
const outputRef = ref<HTMLElement | null>(null)
const rotation = ref(0)
let pzInstance: PanZoom | null = null

function resetView() {
  if (!pzInstance || !outputRef.value) return
  const container = outputRef.value
  const svg = container.querySelector('svg')
  if (!svg) {
    pzInstance.moveTo(0, 0)
    pzInstance.zoomAbs(0, 0, 1)
    return
  }

  const rectWidth = container.clientWidth || 800
  const rectHeight = container.clientHeight || 600
  const cx = rectWidth / 2
  const cy = rectHeight / 2

  let contentWidth = svg.clientWidth || svg.viewBox.baseVal.width || 600
  let contentHeight = svg.clientHeight || svg.viewBox.baseVal.height || 400

  if (rotation.value % 180 !== 0) {
    const temp = contentWidth
    contentWidth = contentHeight
    contentHeight = temp
  }

  if (contentWidth > 0 && contentHeight > 0) {
    const scaleX = rectWidth / contentWidth
    const scaleY = rectHeight / contentHeight
    const scale = Math.min(scaleX, scaleY) * 0.95

    pzInstance.zoomAbs(0, 0, 1)
    pzInstance.moveTo(0, 0)
    pzInstance.zoomAbs(cx, cy, scale)
  } else {
    pzInstance.moveTo(0, 0)
    pzInstance.zoomAbs(0, 0, 1)
  }
}

function handleZoom(type: 'in' | 'out') {
  if (!pzInstance || !outputRef.value) return
  const cx = outputRef.value.clientWidth / 2
  const cy = outputRef.value.clientHeight / 2
  if (type === 'in') pzInstance.smoothZoom(cx, cy, 1.25)
  else pzInstance.smoothZoom(cx, cy, 1 / 1.25)
}

function handleRotate() {
  rotation.value = (rotation.value + 90) % 360
  nextTick(() => {
    resetView()
  })
}

async function copyMermaidSource() {
  if (!props.sourceCode) return
  try {
    await navigator.clipboard.writeText(props.sourceCode)
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

function destroyPanzoom() {
  if (pzInstance) {
    pzInstance.dispose()
    pzInstance = null
  }
}

watch(
  () => props.open,
  async (isOpen) => {
    await nextTick()
    const dialog = dialogRef.value
    if (!dialog) return

    if (isOpen) {
      if (!dialog.open) dialog.showModal()
      rotation.value = 0
      if (outputRef.value) {
        destroyPanzoom()
        pzInstance = panzoom(outputRef.value, {
          maxZoom: 10,
          minZoom: 0.05,
          bounds: false,
        })
        resetView()
      }
    } else {
      destroyPanzoom()
      if (dialog.open) dialog.close()
    }
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  destroyPanzoom()
})
</script>

<template>
  <dialog
    v-if="open"
    ref="dialogRef"
    class="fullscreen-dialog"
    @close="emit('close')"
    @cancel.prevent="emit('close')"
  >
    <div class="modal-backdrop" @click="emit('close')"></div>
    <div class="modal-window">
      <div class="modal-toolbar">
        <div class="toolbar-group">
          <button type="button" class="tool-btn" title="放大" @click="handleZoom('in')">
            <span v-html="iconSvg('zoom-in', 16)"></span>
          </button>
          <button type="button" class="tool-btn" title="缩小" @click="handleZoom('out')">
            <span v-html="iconSvg('zoom-out', 16)"></span>
          </button>
          <button type="button" class="tool-btn" title="重置视角" @click="resetView">
            <span v-html="iconSvg('rotate-ccw', 16)"></span>
          </button>
          <button type="button" class="tool-btn" title="旋转 90°" @click="handleRotate">
            <span v-html="iconSvg('rotate-cw', 16)"></span>
          </button>
        </div>

        <div class="toolbar-divider"></div>

        <div class="toolbar-group">
          <button type="button" class="tool-btn" title="复制 Mermaid 源码" @click="copyMermaidSource">
            <span v-html="iconSvg('file-code', 16)"></span>
            <span class="btn-text">源码</span>
          </button>
          <button type="button" class="tool-btn" title="复制 PNG (白底)" @click="copyPng">
            <span v-html="iconSvg('copy', 16)"></span>
            <span class="btn-text">复制图片</span>
          </button>
          <button type="button" class="tool-btn" title="导出透明 PNG" @click="downloadPng">
            <span v-html="iconSvg('download', 16)"></span>
            <span class="btn-text">导出 PNG</span>
          </button>
          <button type="button" class="tool-btn" title="导出 SVG" @click="downloadSvg">
            <span v-html="iconSvg('svg', 16)"></span>
            <span class="btn-text">SVG</span>
          </button>
        </div>

        <div class="spacer"></div>

        <button type="button" class="tool-btn close-btn" title="关闭全屏" @click="emit('close')">
          <span v-html="iconSvg('minimize', 16)"></span>
          <span class="btn-text">退出全屏</span>
        </button>
      </div>

      <div class="modal-canvas">
        <div ref="outputRef" class="panzoom-viewport">
          <div
            class="svg-rotator"
            :style="{ transform: `rotate(${rotation}deg)` }"
            v-html="svgHtml"
          ></div>
        </div>
      </div>
    </div>
  </dialog>
</template>

<style scoped>
.fullscreen-dialog {
  position: fixed;
  inset: 0;
  width: 100vw;
  height: 100vh;
  max-width: 100vw;
  max-height: 100vh;
  margin: 0;
  padding: 0;
  border: none;
  background: transparent;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.65);
  backdrop-filter: blur(4px);
}

.modal-window {
  position: relative;
  width: 95vw;
  height: 92vh;
  background: var(--surface-raised);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  z-index: 10;
}

.modal-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: var(--surface-muted);
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}

.toolbar-group {
  display: flex;
  align-items: center;
  gap: 4px;
}

.toolbar-divider {
  width: 1px;
  height: 18px;
  background: var(--border-strong);
  margin: 0 4px;
}

.spacer {
  flex: 1;
}

.tool-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 10px;
  border-radius: var(--radius-sm);
  color: var(--text);
  background: var(--surface);
  border: 1px solid var(--border);
  font-size: 12px;
  font-weight: 500;
  transition: all 120ms ease;
}

.tool-btn:hover {
  background: var(--surface-hover);
  border-color: var(--border-strong);
}

.close-btn {
  color: var(--danger);
  border-color: var(--danger-soft);
  background: var(--danger-soft);
}

.close-btn:hover {
  filter: brightness(0.95);
}

.btn-text {
  font-size: 12px;
}

.modal-canvas {
  flex: 1;
  width: 100%;
  height: 100%;
  overflow: hidden;
  background: var(--surface);
  position: relative;
  cursor: grab;
}

.modal-canvas:active {
  cursor: grabbing;
}

.panzoom-viewport {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.svg-rotator {
  display: inline-block;
  transition: transform 180ms ease;
}

.svg-rotator :deep(svg) {
  max-width: none;
  height: auto;
  user-select: none;
}
</style>
