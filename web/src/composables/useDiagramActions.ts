import {
  copyPngToClipboard,
  copyTextToClipboard,
  exportPngImage,
  exportSvgImage,
} from '@/utils/export'

type ToastFn = (message: string, isError?: boolean) => void

/**
 * Shared copy/download actions for a rendered Mermaid SVG.
 * Used by both the inline preview toolbar and the fullscreen modal, which
 * previously carried near-identical copies of these functions.
 */
export function useDiagramActions(toast: ToastFn) {
  async function copySource(code: string) {
    if (!code) return
    const ok = await copyTextToClipboard(code)
    if (ok) {
      toast('已复制 Mermaid 源码')
    } else {
      toast('复制源码失败', true)
    }
  }

  async function copyPng(svg: SVGSVGElement | null | undefined) {
    if (!svg) return
    const res = await copyPngToClipboard(svg, { transparent: false })
    toast(res.message, !res.success)
  }

  async function downloadPng(svg: SVGSVGElement | null | undefined) {
    if (!svg) return
    try {
      await exportPngImage(svg, 'mermaid-diagram.png', { transparent: true })
      toast('已下载透明 PNG 图片')
    } catch {
      toast('导出 PNG 失败', true)
    }
  }

  function downloadSvg(svg: SVGSVGElement | null | undefined) {
    if (!svg) return
    try {
      exportSvgImage(svg, 'mermaid-diagram.svg')
      toast('已下载 SVG 矢量图')
    } catch {
      toast('导出 SVG 失败', true)
    }
  }

  return { copySource, copyPng, downloadPng, downloadSvg }
}
