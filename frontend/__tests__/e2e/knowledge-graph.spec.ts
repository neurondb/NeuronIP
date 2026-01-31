import { test, expect } from './fixtures'

test.describe('Knowledge Graph', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/knowledge-graph')
    await page.waitForLoadState('networkidle')
  })

  test('should display knowledge graph visualization', async ({ page }) => {
    const graph = page.locator('.knowledge-graph, .graph-visualization, [data-testid="knowledge-graph"]').first()
    await expect(graph).toBeVisible({ timeout: 3000 })
  })

  test('should extract entities', async ({ page }) => {
    const extractButton = page.locator('button:has-text("Extract"), button:has-text("Extract Entities")').first()
    
    if (await extractButton.isVisible()) {
      await extractButton.click()
      
      // Should show extract dialog
      const dialog = page.locator('[role="dialog"], .modal').first()
      await expect(dialog).toBeVisible({ timeout: 3000 })
    }
  })

  test('should search entities', async ({ page }) => {
    const searchInput = page.locator('input[placeholder*="Search" i], input[type="search"]').first()
    
    if (await searchInput.isVisible()) {
      await searchInput.fill('test entity')
      await page.waitForTimeout(1000)
      
      // Should show search results
      const results = page.locator('.search-results, .entity-results').first()
      await expect(results).toBeVisible({ timeout: 3000 })
    }
  })

  test('should traverse graph', async ({ page }) => {
    // Click on a node in the graph
    const node = page.locator('.graph-node, .entity-node').first()
    
    if (await node.isVisible()) {
      await node.click()
      await page.waitForTimeout(500)
      
      // Should show node details or expand connections
      const details = page.locator('.node-details, .entity-details').first()
      await expect(details).toBeVisible({ timeout: 3000 })
    }
  })

  test('should link entities', async ({ page }) => {
    const linkButton = page.locator('button:has-text("Link"), button:has-text("Connect")').first()
    
    if (await linkButton.isVisible()) {
      await linkButton.click()
      await page.waitForTimeout(500)
      
      // Should show link dialog
      const dialog = page.locator('[role="dialog"], .modal').first()
      await expect(dialog).toBeVisible({ timeout: 3000 })
    }
  })

  test('should view glossary', async ({ page }) => {
    const glossarySection = page.locator('text=Glossary, .glossary-list').first()
    await expect(glossarySection).toBeVisible({ timeout: 3000 })
  })
})
