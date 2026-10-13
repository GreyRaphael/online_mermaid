import { expect, type Page } from '@playwright/test'

/** Log in with the default e2e credentials and wait for the workspace. */
export async function login(page: Page) {
  await page.goto('/')
  await expect(page.getByRole('heading', { name: 'Online Mermaid' })).toBeVisible()
  await page.getByLabel('用户名').fill('admin')
  await page.getByLabel('密码').fill('admin123')
  await page.getByRole('button', { name: '登录', exact: true }).click()
  await expect(page.getByRole('button', { name: '切换侧栏' })).toBeVisible()
}
