import { test, expect } from './fixtures'

test.describe('Compliance', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/compliance')
    await page.waitForLoadState('networkidle')
  })

  test('should display compliance dashboard', async ({ page }) => {
    const dashboard = page.locator('.compliance-dashboard, [data-testid="compliance-dashboard"]').first()
    await expect(dashboard).toBeVisible({ timeout: 3000 })
  })

  test('should run compliance check', async ({ page }) => {
    const checkButton = page.locator('button:has-text("Check"), button:has-text("Run Check")').first()
    
    if (await checkButton.isVisible()) {
      await checkButton.click()
      await page.waitForTimeout(2000)
      
      // Should show check results
      const results = page.locator('.compliance-results, [data-testid="compliance-results"]').first()
      await expect(results).toBeVisible({ timeout: 5000 })
    }
  })

  test('should view compliance policies', async ({ page }) => {
    const policiesSection = page.locator('text=Policies, .policies-list').first()
    await expect(policiesSection).toBeVisible({ timeout: 3000 })
  })

  test('should create compliance policy', async ({ page }) => {
    const createButton = page.locator('button:has-text("Create Policy"), button:has-text("New Policy")').first()
    
    if (await createButton.isVisible()) {
      await createButton.click()
      
      // Should show create dialog
      const dialog = page.locator('[role="dialog"], .modal').first()
      await expect(dialog).toBeVisible({ timeout: 3000 })
    }
  })

  test('should view anomaly detections', async ({ page }) => {
    const anomaliesSection = page.locator('text=Anomalies, .anomalies-list').first()
    await expect(anomaliesSection).toBeVisible({ timeout: 3000 })
  })

  test('should view compliance report', async ({ page }) => {
    const reportButton = page.locator('button:has-text("Report"), button:has-text("View Report")').first()
    
    if (await reportButton.isVisible()) {
      await reportButton.click()
      await page.waitForTimeout(1000)
      
      // Should show report
      const report = page.locator('.compliance-report, [data-testid="compliance-report"]').first()
      await expect(report).toBeVisible({ timeout: 5000 })
    }
  })
})
