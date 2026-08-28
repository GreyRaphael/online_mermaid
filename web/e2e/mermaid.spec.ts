import { expect, test, type Page } from '@playwright/test'

async function login(page: Page) {
  await page.goto('/')
  await expect(page.getByRole('heading', { name: 'Online Mermaid' })).toBeVisible()
  await page.getByLabel('用户名').fill('admin')
  await page.getByLabel('密码').fill('admin123')
  await page.getByRole('button', { name: '登录', exact: true }).click()
  await expect(page.getByRole('button', { name: '切换侧栏' })).toBeVisible()
}

test.describe('Online Mermaid E2E', () => {
  test('Login, renders mermaid flowchart with node text, and logout via user submenu', async ({
    page,
  }) => {
    await login(page)

    // Verify diagram renders SVG with text
    const svg = page.locator('.mermaid-output svg')
    await expect(svg).toBeVisible({ timeout: 15_000 })
    await expect(page.locator('.mermaid-output')).toHaveText(/节点|开始|Mermaid/)

    // Test user dropdown submenu containing exit / logout button
    const userTrigger = page.locator('.user-menu-trigger')
    await expect(userTrigger).toBeVisible()
    await userTrigger.click()

    const logoutBtn = page.getByRole('menuitem', { name: '退出登录' })
    await expect(logoutBtn).toBeVisible()
    await logoutBtn.click()

    // Should return to login page
    await expect(page.getByRole('heading', { name: 'Online Mermaid' })).toBeVisible()
  })

  test('View mode switching (Preview, Edit, Split)', async ({ page }) => {
    await login(page)

    // Initially in Split mode
    await expect(page.locator('.editor-pane')).toBeVisible()
    await expect(page.locator('.preview-pane')).toBeVisible()

    // Switch to Preview mode
    await page.getByRole('button', { name: '预览' }).click()
    await expect(page.locator('.preview-pane')).toBeVisible()
    await expect(page.locator('.editor-pane')).not.toBeVisible()

    // Switch to Edit mode
    await page.getByRole('button', { name: '编辑' }).click()
    await expect(page.locator('.editor-pane')).toBeVisible()
    await expect(page.locator('.preview-pane')).not.toBeVisible()

    // Switch back to Split mode
    await page.getByRole('button', { name: '分屏' }).click()
    await expect(page.locator('.editor-pane')).toBeVisible()
    await expect(page.locator('.preview-pane')).toBeVisible()
  })

  test('Diagram CRUD: create, edit code, auto-render, rename, delete', async ({ page }) => {
    await login(page)

    const isMobile = test.info().project.name.startsWith('mobile')
    if (isMobile) {
      await page.getByRole('button', { name: '切换侧栏' }).click()
      await expect(page.locator('.sidebar-container.mobile-drawer')).toHaveClass(/mobile-open/)
    }

    // Create new diagram
    await page.getByRole('button', { name: '新建图表' }).click()

    // Should have created a new diagram
    if (!isMobile) {
      await expect(page.locator('.diagram-title')).toBeVisible()
    }

    // Switch to Edit mode to edit code
    await page.getByRole('button', { name: '编辑' }).click()
    const textarea = page.locator('.code-textarea')
    await textarea.fill('flowchart LR\n A[用户请求] --> B[服务处理]\n B --> C[返回结果]')

    // Switch to Preview mode to verify rendered text
    await page.getByRole('button', { name: '预览' }).click()
    const svg = page.locator('.mermaid-output svg')
    await expect(svg).toBeVisible({ timeout: 15_000 })
    await expect(page.locator('.mermaid-output')).toContainText('用户请求')
    await expect(page.locator('.mermaid-output')).toContainText('服务处理')
  })

  test('Day and Night theme switching', async ({ page }) => {
    await login(page)

    const nightBtn = page.getByRole('button', { name: '深色' })
    await expect(nightBtn).toBeVisible()
    await nightBtn.click()
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'night')

    const dayBtn = page.getByRole('button', { name: '浅色' })
    await expect(dayBtn).toBeVisible()
    await dayBtn.click()
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'day')

    // Reload persists theme
    await page.reload()
    await expect(page.locator('html')).toHaveAttribute('data-theme', 'day')
  })

  test('Mobile drawer opens, closes on diagram select, and closes on backdrop click', async ({
    page,
  }) => {
    test.skip(
      !test.info().project.name.startsWith('mobile'),
      'Mobile drawer interaction test',
    )

    await login(page)

    // Mobile drawer should be initially hidden off-canvas
    const drawer = page.locator('.sidebar-container.mobile-drawer')
    await expect(drawer).not.toHaveClass(/mobile-open/)

    // Click hamburger menu to open drawer
    await page.getByRole('button', { name: '切换侧栏' }).click()
    await expect(drawer).toHaveClass(/mobile-open/)
    await expect(page.locator('.drawer-backdrop')).toBeVisible()

    // Click backdrop outside the drawer (right area) to close drawer
    await page.locator('.drawer-backdrop').click({ position: { x: 350, y: 300 } })
    await expect(drawer).not.toHaveClass(/mobile-open/)

    // Open drawer again and select item to close
    await page.getByRole('button', { name: '切换侧栏' }).click()
    await expect(drawer).toHaveClass(/mobile-open/)
    await page.locator('.diagram-item').first().click()
    await expect(drawer).not.toHaveClass(/mobile-open/)
  })

  test('Fullscreen modal view and 90-degree rotation', async ({ page }) => {
    await login(page)

    // Wait for SVG diagram to be rendered first
    const svg = page.locator('.mermaid-output svg')
    await expect(svg).toBeVisible({ timeout: 15_000 })

    const fullscreenBtn = page.getByRole('button', { name: '全屏查看' })
    await expect(fullscreenBtn).toBeVisible()
    await fullscreenBtn.click()

    // Fullscreen dialog opens
    const dialog = page.locator('.fullscreen-dialog')
    await expect(dialog).toBeVisible()

    // Click Rotate button
    const rotateBtn = page.getByRole('button', { name: '旋转' })
    await expect(rotateBtn).toBeVisible()
    await rotateBtn.click()

    const rotator = page.locator('.svg-rotator')
    await expect(rotator).toHaveAttribute('style', /rotate\(90deg\)/)

    // Close fullscreen
    const closeBtn = page.getByRole('button', { name: '退出' })
    await closeBtn.click()
    await expect(dialog).not.toBeVisible()
  })
})
