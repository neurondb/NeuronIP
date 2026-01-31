import { test, expect } from './fixtures'

test.describe('Agents', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/agents')
    await page.waitForLoadState('domcontentloaded')
  })

  test('should display agent hub page', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /Agent Hub/i })).toBeVisible({ timeout: 8000 })
    const hasTable = await page.locator('table').first().isVisible().catch(() => false)
    const hasEmpty = await page.getByText(/No agents found|Create one to get started/i).first().isVisible().catch(() => false)
    expect(hasTable || hasEmpty).toBeTruthy()
  })

  test('should open create dialog via Quick Create', async ({ page }) => {
    const btn = page.getByRole('button', { name: /Quick Create/i })
    await expect(btn).toBeVisible({ timeout: 5000 })
    await btn.click()
    await expect(page.locator('[role="dialog"]').first()).toBeVisible({ timeout: 3000 })
  })

  test('should open create wizard via Create with Wizard', async ({ page }) => {
    const btn = page.getByRole('button', { name: /Create with Wizard/i })
    await expect(btn).toBeVisible({ timeout: 5000 })
    await btn.click()
    await expect(page.locator('[role="dialog"]').first()).toBeVisible({ timeout: 3000 })
  })

  test('should show agent detail when clicking a row', async ({ page }) => {
    const row = page.locator('tbody tr').first()
    if (!(await row.isVisible().catch(() => false))) {
      test.skip()
      return
    }
    await row.click()
    await expect(
      page.getByRole('button', { name: /Back to List/i }).or(page.getByText(/Status/i))
    ).toBeVisible({ timeout: 5000 })
  })
})
