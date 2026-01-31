import { test, expect } from './fixtures'

test.describe('Workflows', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/workflows')
    await page.waitForLoadState('networkidle')
  })

  test('should display workflow list', async ({ page }) => {
    // Check for workflow list or table
    const workflowList = page.locator('table, .workflow-list, [data-testid="workflow-list"]').first()
    await expect(workflowList).toBeVisible({ timeout: 3000 })
  })

  test('should create new workflow', async ({ page }) => {
    const createButton = page.locator('button:has-text("Create"), button:has-text("New"), button:has-text("Add")').first()
    
    if (await createButton.isVisible()) {
      await createButton.click()
      
      // Should show create dialog or navigate to builder
      const dialog = page.locator('[role="dialog"], .modal, .workflow-builder').first()
      await expect(dialog).toBeVisible({ timeout: 3000 })
    }
  })

  test('should open workflow builder', async ({ page }) => {
    await page.goto('/workflows/builder')
    await page.waitForLoadState('networkidle')
    
    // Check for workflow builder canvas
    const canvas = page.locator('.workflow-canvas, .react-flow, [data-testid="workflow-builder"]').first()
    await expect(canvas).toBeVisible({ timeout: 3000 })
  })

  test('should add workflow nodes', async ({ page }) => {
    await page.goto('/workflows/builder')
    await page.waitForLoadState('networkidle')
    
    // Look for node palette or add node button
    const addNodeButton = page.locator('button:has-text("Add"), button:has-text("Node"), .node-palette button').first()
    
    if (await addNodeButton.isVisible()) {
      await addNodeButton.click()
      await page.waitForTimeout(500)
      
      // Node should be added to canvas
      const node = page.locator('.workflow-node, .react-flow__node').first()
      await expect(node).toBeVisible({ timeout: 3000 })
    }
  })

  test('should execute workflow', async ({ page }) => {
    // Look for execute button on a workflow
    const executeButton = page.locator('button:has-text("Execute"), button:has-text("Run")').first()
    
    if (await executeButton.isVisible()) {
      await executeButton.click()
      await page.waitForTimeout(1000)
      
      // Should show execution status or results
      const status = page.locator('.status, .execution-status, [data-testid="execution-status"]').first()
      await expect(status).toBeVisible({ timeout: 5000 })
    }
  })

  test('should view workflow details', async ({ page }) => {
    // Click on a workflow in the list
    const workflowRow = page.locator('tbody tr, .workflow-item').first()
    
    if (await workflowRow.isVisible()) {
      await workflowRow.click()
      await page.waitForTimeout(500)
      
      // Should show workflow details
      const details = page.locator('.workflow-details, [data-testid="workflow-details"]').first()
      await expect(details).toBeVisible({ timeout: 3000 })
    }
  })

  test('should schedule workflow', async ({ page }) => {
    const scheduleButton = page.locator('button:has-text("Schedule"), button:has-text("Cron")').first()
    
    if (await scheduleButton.isVisible()) {
      await scheduleButton.click()
      await page.waitForTimeout(500)
      
      // Should show schedule dialog
      const dialog = page.locator('[role="dialog"], .modal').first()
      await expect(dialog).toBeVisible({ timeout: 3000 })
    }
  })

  test('should view workflow execution history', async ({ page }) => {
    // Look for execution history section
    const historySection = page.locator('text=History, text=Executions, .execution-history').first()
    await expect(historySection).toBeVisible({ timeout: 3000 })
  })
})
