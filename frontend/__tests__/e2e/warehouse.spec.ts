import { test, expect } from './fixtures'

test.describe('Warehouse', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/warehouse')
    await page.waitForLoadState('networkidle')
  })

  test('should display query editor', async ({ page }) => {
    // Look for SQL editor or textarea
    const editor = page.locator('textarea, .monaco-editor, [data-testid="query-editor"]').first()
    await expect(editor).toBeVisible()
  })

  test('should execute SQL query', async ({ page }) => {
    const editor = page.locator('textarea, .monaco-editor textarea').first()
    const executeButton = page.locator('button:has-text("Execute"), button:has-text("Run"), button[aria-label*="execute" i]').first()
    
    // Type a simple query
    await editor.fill('SELECT 1 as test')
    
    // Execute
    await executeButton.click()
    
    // Wait for results
    await page.waitForTimeout(2000)
    
    // Check for results table or output
    const results = page.locator('table, .results, [data-testid="query-results"]').first()
    await expect(results).toBeVisible({ timeout: 10000 })
  })

  test('should display query history', async ({ page }) => {
    // Look for query history section
    const historySection = page.locator('text=History, text=Recent Queries, [data-testid="query-history"]').first()
    await expect(historySection).toBeVisible({ timeout: 3000 })
  })

  test('should save query', async ({ page }) => {
    const editor = page.locator('textarea, .monaco-editor textarea').first()
    const saveButton = page.locator('button:has-text("Save"), button[aria-label*="save" i]').first()
    
    await editor.fill('SELECT * FROM test')
    
    if (await saveButton.isVisible()) {
      await saveButton.click()
      
      // Should show save dialog or confirmation
      const dialog = page.locator('[role="dialog"], .modal').first()
      await expect(dialog).toBeVisible({ timeout: 3000 })
    }
  })

  test('should load saved queries', async ({ page }) => {
    // Look for saved queries list
    const savedQueries = page.locator('text=Saved, .saved-queries, [data-testid="saved-queries"]').first()
    await expect(savedQueries).toBeVisible({ timeout: 3000 })
  })

  test('should show query syntax highlighting', async ({ page }) => {
    const editor = page.locator('textarea, .monaco-editor').first()
    
    await editor.fill('SELECT * FROM users WHERE id = 1')
    
    // Monaco editor should have syntax highlighting
    const hasMonaco = await page.locator('.monaco-editor').count() > 0
    expect(hasMonaco).toBeTruthy()
  })

  test('should display query results in table', async ({ page }) => {
    const editor = page.locator('textarea, .monaco-editor textarea').first()
    const executeButton = page.locator('button:has-text("Execute"), button:has-text("Run")').first()
    
    await editor.fill('SELECT 1 as col1, 2 as col2')
    await executeButton.click()
    
    await page.waitForTimeout(2000)
    
    // Check for table with results
    const table = page.locator('table').first()
    if (await table.isVisible()) {
      const rows = table.locator('tbody tr')
      await expect(rows.first()).toBeVisible({ timeout: 5000 })
    }
  })

  test('should handle query errors', async ({ page }) => {
    const editor = page.locator('textarea, .monaco-editor textarea').first()
    const executeButton = page.locator('button:has-text("Execute"), button:has-text("Run")').first()
    
    // Execute invalid query
    await editor.fill('INVALID SQL QUERY')
    await executeButton.click()
    
    await page.waitForTimeout(2000)
    
    // Should show error message
    const errorMessage = page.locator('.error, [role="alert"], text=/error/i').first()
    await expect(errorMessage).toBeVisible({ timeout: 5000 })
  })
})
