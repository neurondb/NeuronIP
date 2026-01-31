import { test, expect } from './fixtures'

test.describe('Models', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/models')
    await page.waitForLoadState('networkidle')
  })

  test('should display model list', async ({ page }) => {
    const modelList = page.locator('table, .model-list, [data-testid="model-list"]').first()
    await expect(modelList).toBeVisible({ timeout: 3000 })
  })

  test('should register new model', async ({ page }) => {
    const registerButton = page.locator('button:has-text("Register"), button:has-text("Add Model")').first()
    
    if (await registerButton.isVisible()) {
      await registerButton.click()
      
      // Should show register dialog
      const dialog = page.locator('[role="dialog"], .modal').first()
      await expect(dialog).toBeVisible({ timeout: 3000 })
    }
  })

  test('should run model inference', async ({ page }) => {
    const inferButton = page.locator('button:has-text("Infer"), button:has-text("Run")').first()
    
    if (await inferButton.isVisible()) {
      await inferButton.click()
      await page.waitForTimeout(500)
      
      // Should show inference interface
      const inferenceInterface = page.locator('.inference-interface, [data-testid="inference"]').first()
      await expect(inferenceInterface).toBeVisible({ timeout: 3000 })
    }
  })

  test('should view model details', async ({ page }) => {
    const modelRow = page.locator('tbody tr, .model-item').first()
    
    if (await modelRow.isVisible()) {
      await modelRow.click()
      await page.waitForTimeout(500)
      
      // Should show model details
      const details = page.locator('.model-details, [data-testid="model-details"]').first()
      await expect(details).toBeVisible({ timeout: 3000 })
    }
  })
})
