export async function svgToPngBlob(
  svgElement: SVGSVGElement,
  options: { transparent?: boolean; scale?: number } = {},
): Promise<Blob> {
  return new Promise((resolve, reject) => {
    const clone = svgElement.cloneNode(true) as SVGSVGElement
    const viewBox = svgElement.viewBox?.baseVal
    const width = viewBox && viewBox.width > 0 ? viewBox.width : svgElement.clientWidth || 800
    const height = viewBox && viewBox.height > 0 ? viewBox.height : svgElement.clientHeight || 600

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
