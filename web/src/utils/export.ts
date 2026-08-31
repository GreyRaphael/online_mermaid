/**
 * Utility functions for exporting and copying Mermaid diagrams (PNG, SVG, Source code).
 * Includes cross-browser compatibility and fallbacks for non-HTTPS (HTTP server) environments.
 */

export async function copyTextToClipboard(text: string): Promise<boolean> {
  if (!text) return false

  // 1. Try modern navigator.clipboard if available and in secure context
  if (
    typeof navigator !== 'undefined' &&
    navigator.clipboard &&
    typeof navigator.clipboard.writeText === 'function'
  ) {
    try {
      await navigator.clipboard.writeText(text)
      return true
    } catch {
      // Fall through to execCommand fallback
    }
  }

  // 2. Fallback to document.execCommand('copy') with hidden textarea for HTTP environments
  try {
    const textArea = document.createElement('textarea')
    textArea.value = text
    textArea.style.position = 'fixed'
    textArea.style.top = '0'
    textArea.style.left = '-9999px'
    textArea.style.width = '2em'
    textArea.style.height = '2em'
    textArea.style.padding = '0'
    textArea.style.border = 'none'
    textArea.style.outline = 'none'
    textArea.style.boxShadow = 'none'
    textArea.style.background = 'transparent'
    textArea.setAttribute('readonly', '')
    document.body.appendChild(textArea)
    textArea.focus()
    textArea.select()
    textArea.setSelectionRange(0, text.length)
    const success = document.execCommand('copy')
    document.body.removeChild(textArea)
    return success
  } catch {
    return false
  }
}

export async function svgToPngBlob(
  svgElement: SVGSVGElement,
  options: { transparent?: boolean; scale?: number } = {},
): Promise<Blob> {
  return new Promise((resolve, reject) => {
    const clone = svgElement.cloneNode(true) as SVGSVGElement
    if (!clone.hasAttribute('xmlns')) {
      clone.setAttribute('xmlns', 'http://www.w3.org/2000/svg')
    }
    if (!clone.hasAttribute('xmlns:xlink')) {
      clone.setAttribute('xmlns:xlink', 'http://www.w3.org/1999/xlink')
    }

    const viewBox = svgElement.viewBox?.baseVal
    const bcr = svgElement.getBoundingClientRect ? svgElement.getBoundingClientRect() : null
    const width = viewBox && viewBox.width > 0 ? viewBox.width : (bcr?.width || svgElement.clientWidth || 800)
    const height = viewBox && viewBox.height > 0 ? viewBox.height : (bcr?.height || svgElement.clientHeight || 600)

    clone.setAttribute('width', `${width}`)
    clone.setAttribute('height', `${height}`)

    const svgString = new XMLSerializer().serializeToString(clone)
    const svgBlob = new Blob([svgString], { type: 'image/svg+xml;charset=utf-8' })
    const url = URL.createObjectURL(svgBlob)

    const img = new Image()
    img.crossOrigin = 'anonymous'
    img.onload = () => {
      const canvas = document.createElement('canvas')
      const scale = options.scale ?? 2
      canvas.width = width * scale
      canvas.height = height * scale
      const ctx = canvas.getContext('2d')
      if (!ctx) {
        URL.revokeObjectURL(url)
        reject(new Error('Canvas 上下文不可用'))
        return
      }

      if (!options.transparent) {
        ctx.fillStyle = '#ffffff'
        ctx.fillRect(0, 0, canvas.width, canvas.height)
      } else {
        ctx.clearRect(0, 0, canvas.width, canvas.height)
      }

      ctx.scale(scale, scale)
      ctx.drawImage(img, 0, 0, width, height)
      URL.revokeObjectURL(url)

      canvas.toBlob((blob) => {
        if (blob) resolve(blob)
        else reject(new Error('转换为图片 Blob 失败'))
      }, 'image/png')
    }
    img.onerror = (err) => {
      URL.revokeObjectURL(url)
      reject(err)
    }
    img.src = url
  })
}

export interface CopyImageResult {
  success: boolean
  fallbackDownloaded: boolean
  message: string
}

export async function copyPngToClipboard(
  svgElement: SVGSVGElement,
  options: { transparent?: boolean; fallbackFilename?: string } = {},
): Promise<CopyImageResult> {
  let blob: Blob
  try {
    blob = await svgToPngBlob(svgElement, { transparent: options.transparent ?? false })
  } catch (err: any) {
    return {
      success: false,
      fallbackDownloaded: false,
      message: '生成图片失败: ' + (err?.message || String(err)),
    }
  }

  // 1. If in Secure Context and ClipboardItem API is supported (HTTPS or localhost)
  if (
    typeof window !== 'undefined' &&
    window.isSecureContext &&
    typeof navigator !== 'undefined' &&
    navigator.clipboard &&
    typeof navigator.clipboard.write === 'function' &&
    typeof window.ClipboardItem !== 'undefined'
  ) {
    try {
      const item = new ClipboardItem({ 'image/png': blob })
      await navigator.clipboard.write([item])
      return {
        success: true,
        fallbackDownloaded: false,
        message: options.transparent ? '已复制 PNG 图片 (透明)' : '已复制 PNG 图片 (白底)',
      }
    } catch {
      // Fall through to automatic download fallback
    }
  }

  // 2. Fallback when browser restricts clipboard write (e.g. non-HTTPS HTTP server)
  try {
    const filename = options.fallbackFilename || 'mermaid-diagram.png'
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    return {
      success: true,
      fallbackDownloaded: true,
      message: '已自动下载 PNG（非 HTTPS 环境剪贴板图片受限）',
    }
  } catch (downloadErr: any) {
    return {
      success: false,
      fallbackDownloaded: false,
      message: '复制图片受限或失败: ' + (downloadErr?.message || String(downloadErr)),
    }
  }
}

export async function exportPngImage(
  svgElement: SVGSVGElement,
  filename = 'mermaid-diagram.png',
  options: { transparent?: boolean } = { transparent: true },
): Promise<void> {
  const blob = await svgToPngBlob(svgElement, options)
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

export function exportSvgImage(svgElement: SVGSVGElement, filename = 'mermaid-diagram.svg'): void {
  const clone = svgElement.cloneNode(true) as SVGSVGElement
  if (!clone.hasAttribute('xmlns')) {
    clone.setAttribute('xmlns', 'http://www.w3.org/2000/svg')
  }
  if (!clone.hasAttribute('xmlns:xlink')) {
    clone.setAttribute('xmlns:xlink', 'http://www.w3.org/1999/xlink')
  }
  const svgString = new XMLSerializer().serializeToString(clone)
  const blob = new Blob([svgString], { type: 'image/svg+xml;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}
