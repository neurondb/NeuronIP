import { test, expect } from './fixtures'

test.describe('Navigation', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/agents')
    await page.waitForLoadState('domcontentloaded')
  })

  test('should navigate to Why NeuronIP', async ({ page }) => {
    await page.locator('a[href="/why-neuronip"]').first().click()
    await expect(page).toHaveURL(/\/why-neuronip/)
  })

  test('should navigate to Semantic Search via feature grid', async ({ page }) => {
    await page.goto('/features')
    await page.waitForLoadState('domcontentloaded')
    await page.locator('a[href="/semantic"]').first().click()
    await expect(page).toHaveURL(/\/semantic/)
  })

  test('should navigate to Warehouse via feature grid', async ({ page }) => {
    await page.goto('/features')
    await page.waitForLoadState('domcontentloaded')
    await page.locator('a[href="/warehouse"]').first().click()
    await expect(page).toHaveURL(/\/warehouse/)
  })

  test('should toggle sidebar collapse', async ({ page }) => {
    const toggle = page.getByRole('button', { name: /Expand sidebar|Collapse sidebar/i })
    await toggle.click()
    await page.waitForTimeout(400)
    const sidebar = page.locator('aside').first()
    const width = await sidebar.evaluate((el) => window.getComputedStyle(el).width)
    expect(parseInt(width, 10)).toBeLessThan(256)
  })
})
