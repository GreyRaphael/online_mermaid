<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { MERMAID_TEMPLATES, type MermaidTemplate } from '@/utils/templates'
import { iconSvg } from '@/utils/icons'

const props = defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  save: []
  scroll: [event: Event]
}>()

const textareaRef = ref<HTMLTextAreaElement | null>(null)
const lineNumbersRef = ref<HTMLElement | null>(null)
const showTemplateMenu = ref(false)

const linesCount = computed(() => {
  return Math.max(1, props.modelValue.split('\n').length)
})

function handleInput(e: Event) {
  const target = e.target as HTMLTextAreaElement
  emit('update:modelValue', target.value)
}

function handleKeydown(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key === 's') {
    e.preventDefault()
    emit('save')
    return
  }

  // Handle Tab key
  if (e.key === 'Tab') {
    e.preventDefault()
    const textarea = textareaRef.value
    if (!textarea) return

    const start = textarea.selectionStart
    const end = textarea.selectionEnd
    const value = textarea.value

    const insertText = '    ' // 4 spaces
    const updated = value.substring(0, start) + insertText + value.substring(end)
    emit('update:modelValue', updated)

    nextTick(() => {
      textarea.selectionStart = textarea.selectionEnd = start + insertText.length
    })
  }
}

function handleScroll(e: Event) {
  if (lineNumbersRef.value && textareaRef.value) {
    lineNumbersRef.value.scrollTop = textareaRef.value.scrollTop
  }
  emit('scroll', e)
}

function insertTemplate(tpl: MermaidTemplate) {
  emit('update:modelValue', tpl.code)
  showTemplateMenu.value = false
  nextTick(() => {
    textareaRef.value?.focus()
  })
}

function closeTemplateMenu() {
  showTemplateMenu.value = false
}

defineExpose({
  getTextarea: () => textareaRef.value,
  focus: () => textareaRef.value?.focus(),
})
</script>

<template>
  <div class="editor-container" @click="closeTemplateMenu">
    <div class="editor-sub-header">
      <div class="editor-info">
        <span class="badge">Mermaid</span>
        <span class="line-count">{{ linesCount }} 行</span>
      </div>

      <div class="editor-actions" @click.stop>
        <div class="template-dropdown">
          <button
            type="button"
            class="action-btn template-btn"
            title="选择模版快速插入"
            @click="showTemplateMenu = !showTemplateMenu"
          >
            <span v-html="iconSvg('sparkles', 14)"></span>
            <span>模版示例 ▾</span>
          </button>

          <div v-if="showTemplateMenu" class="template-menu" role="menu">
            <div class="menu-header">常用图表模板</div>
            <button
              v-for="tpl in MERMAID_TEMPLATES"
              :key="tpl.id"
              type="button"
              class="menu-item"
              @click="insertTemplate(tpl)"
            >
              <span class="tpl-name">{{ tpl.name }}</span>
              <span class="tpl-cat">{{ tpl.category }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="editor-body">
      <div ref="lineNumbersRef" class="line-numbers" aria-hidden="true">
        <div v-for="n in linesCount" :key="n" class="line-num">{{ n }}</div>
      </div>
      <textarea
        ref="textareaRef"
        class="code-textarea"
        :value="modelValue"
        spellcheck="false"
        autocomplete="off"
        autocapitalize="off"
        placeholder="在此输入 Mermaid 源码..."
        @input="handleInput"
        @keydown="handleKeydown"
        @scroll="handleScroll"
      ></textarea>
    </div>
  </div>
</template>

<style scoped>
.editor-container {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  background: var(--surface);
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  overflow: hidden;
}

.editor-sub-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 12px;
  background: var(--surface-muted);
  border-bottom: 1px solid var(--border);
  font-size: 12px;
  flex-shrink: 0;
}

.editor-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.badge {
  padding: 2px 6px;
  background: var(--accent-soft);
  color: var(--accent-strong);
  font-weight: 600;
  border-radius: var(--radius-xs);
  font-size: 11px;
}

.line-count {
  color: var(--text-muted);
  font-size: 11px;
}

.editor-actions {
  position: relative;
}

.template-dropdown {
  position: relative;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 8px;
  border-radius: var(--radius-xs);
  background: var(--surface);
  border: 1px solid var(--border);
  color: var(--text);
  font-size: 12px;
  font-weight: 500;
  transition: all 120ms ease;
}

.action-btn:hover {
  background: var(--surface-hover);
  border-color: var(--border-strong);
}

.template-menu {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 4px;
  width: 220px;
  background: var(--surface-raised);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  box-shadow: var(--shadow);
  z-index: 100;
  padding: 4px 0;
  max-height: 320px;
  overflow-y: auto;
}

.menu-header {
  padding: 6px 12px 4px;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-faint);
  border-bottom: 1px solid var(--border);
}

.menu-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 6px 12px;
  text-align: left;
  font-size: 12px;
  color: var(--text);
  transition: background 100ms;
}

.menu-item:hover {
  background: var(--surface-hover);
  color: var(--accent-strong);
}

.tpl-name {
  font-weight: 500;
}

.tpl-cat {
  font-size: 10px;
  padding: 1px 4px;
  background: var(--surface-muted);
  border-radius: 3px;
  color: var(--text-muted);
}

.editor-body {
  display: flex;
  flex: 1;
  width: 100%;
  height: 100%;
  min-height: 0;
  position: relative;
}

.line-numbers {
  width: 44px;
  padding: 12px 6px;
  background: var(--surface-muted);
  border-right: 1px solid var(--border);
  color: var(--text-faint);
  font-family: var(--font-mono);
  font-size: 13px;
  line-height: 1.6;
  text-align: right;
  user-select: none;
  overflow: hidden;
}

.line-num {
  height: 20.8px;
}

.code-textarea {
  flex: 1;
  width: 100%;
  height: 100%;
  padding: 12px 14px;
  border: none;
  background: transparent;
  color: var(--text);
  font-family: var(--font-mono);
  font-size: 13px;
  line-height: 1.6;
  resize: none;
  outline: none;
  white-space: pre;
  overflow: auto;
  tab-size: 4;
}

.code-textarea::placeholder {
  color: var(--text-faint);
}
</style>
