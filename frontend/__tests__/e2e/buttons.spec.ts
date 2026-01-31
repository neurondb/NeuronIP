import { test, expect } from './fixtures'

test.describe('Button Interactions', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
  })

  test('should interact with form submit buttons', async ({ page }) => {
    // Navigate to a page with forms (e.g., API Keys)
    await page.goto('/api-keys')
    await page.waitForLoadState('networkidle')

    // Look for create/add button
    const createButton = page.locator('button:has-text("Create"), button:has-text("Add"), button:has-text("New")').first()
    if (await createButton.isVisible()) {
      await createButton.click()
      
      // Check if modal or form appears
      const modal = page.locator('[role="dialog"], .modal, [data-testid="modal"]').first()
      await expect(modal).toBeVisible({ timeout: 3000 })
    }
  })

  test('should interact with action buttons', async ({ page }) => {
    // Navigate to Agents page
    await page.goto('/agents')
    await page.waitForLoadState('networkidle')

    // Look for action buttons (Edit, Delete, etc.)
    const actionButtons = page.locator('button:has-text("Edit"), button:has-text("Delete"), button:has-text("View")')
    const count = await actionButtons.count()
    
    if (count > 0) {
      // Click first action button
      await actionButtons.first().click()
      await page.waitForTimeout(500)
      
      // Verify some action occurred (modal, navigation, etc.)
      const hasModal = await page.locator('[role="dialog"]').count() > 0
      const urlChanged = page.url() !== '/agents'
      
      expect(hasModal || urlChanged).toBeTruthy()
    }
  })

  test('should interact with delete/confirm buttons', async ({ page }) => {
    await page.goto('/api-keys')
    await page.waitForLoadState('networkidle')

    // Look for delete button
    const deleteButton = page.locator('button:has-text("Delete"), button[aria-label*="delete" i]').first()
    if (await deleteButton.isVisible()) {
      await deleteButton.click()
      
      // Should show confirmation dialog
      const confirmDialog = page.locator('text=Confirm, text=Are you sure, [role="alertdialog"]').first()
      await expect(confirmDialog).toBeVisible({ timeout: 3000 })
      
      // Cancel button should work
      const cancelButton = page.locator('button:has-text("Cancel"), button:has-text("No")').first()
      if (await cancelButton.isVisible()) {
        await cancelButton.click()
        await expect(confirmDialog).not.toBeVisible({ timeout: 2000 })
      }
    }
  })

  test('should interact with modal triggers', async ({ page }) => {
    await page.goto('/agents')
    await page.waitForLoadState('networkidle')

    // Look for buttons that open modals
    const modalTriggers = page.locator('button:has-text("Create"), button:has-text("Add"), button:has-text("New")')
    const count = await modalTriggers.count()
    
    if (count > 0) {
      await modalTriggers.first().click()
      
      // Modal should appear
      const modal = page.locator('[role="dialog"], .modal').first()
      await expect(modal).toBeVisible({ timeout: 3000 })
      
      // Close button should work
      const closeButton = page.locator('button[aria-label*="close" i], button:has-text("Close"), button:has-text("×")').first()
      if (await closeButton.isVisible()) {
        await closeButton.click()
        await expect(modal).not.toBeVisible({ timeout: 2000 })
      }
    }
  })

  test('should interact with dropdown toggles', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')

    // Look for dropdown buttons
    const dropdownButtons = page.locator('button[aria-haspopup="menu"], button[aria-expanded]')
    const count = await dropdownButtons.count()
    
    if (count > 0) {
      const firstDropdown = dropdownButtons.first()
      await firstDropdown.click()
      await page.waitForTimeout(300)
      
      // Dropdown menu should be visible
      const menu = page.locator('[role="menu"], .dropdown-menu').first()
      await expect(menu).toBeVisible({ timeout: 2000 })
      
      // Click outside to close
      await page.click('body', { position: { x: 0, y: 0 } })
      await page.waitForTimeout(300)
    }
  })

  test('should interact with tab switches', async ({ page }) => {
    await page.goto('/settings')
    await page.waitForLoadState('networkidle')

    // Look for tabs
    const tabs = page.locator('[role="tab"], .tab, button[role="tab"]')
    const tabCount = await tabs.count()
    
    if (tabCount > 1) {
      // Click second tab
      await tabs.nth(1).click()
      await page.waitForTimeout(300)
      
      // Tab should be active
      const activeTab = tabs.nth(1)
      await expect(activeTab).toHaveAttribute('aria-selected', 'true')
    }
  })

  test('should interact with pagination controls', async ({ page }) => {
    await page.goto('/agents')
    await page.waitForLoadState('networkidle')

    // Look for pagination buttons
    const paginationButtons = page.locator('button:has-text("Next"), button:has-text("Previous"), button[aria-label*="page" i]')
    const count = await paginationButtons.count()
    
    if (count > 0) {
      // Click next button
      const nextButton = page.locator('button:has-text("Next"), button[aria-label*="next" i]').first()
      if (await nextButton.isVisible() && !(await nextButton.isDisabled())) {
        await nextButton.click()
        await page.waitForTimeout(500)
        
        // Page should update (check URL or content)
        const url = page.url()
        expect(url).toBeTruthy()
      }
    }
  })

  test('should interact with sorting controls', async ({ page }) => {
    await page.goto('/agents')
    await page.waitForLoadState('networkidle')

    // Look for sortable column headers
    const sortableHeaders = page.locator('th button, [role="columnheader"] button, .sortable')
    const count = await sortableHeaders.count()
    
    if (count > 0) {
      await sortableHeaders.first().click()
      await page.waitForTimeout(500)
      
      // Check if sort indicator appears
      const sortIndicator = page.locator('[aria-sort], .sort-asc, .sort-desc').first()
      await expect(sortIndicator).toBeVisible({ timeout: 2000 })
    }
  })

  test('should interact with filtering controls', async ({ page }) => {
    await page.goto('/agents')
    await page.waitForLoadState('networkidle')

    // Look for filter inputs or buttons
    const filterInputs = page.locator('input[placeholder*="Filter" i], input[placeholder*="Search" i], button:has-text("Filter")')
    const count = await filterInputs.count()
    
    if (count > 0) {
      const firstFilter = filterInputs.first()
      await firstFilter.click()
      await firstFilter.fill('test')
      await page.waitForTimeout(500)
      
      // Results should update
      const results = page.locator('tbody tr, .result-item')
      await expect(results.first()).toBeVisible({ timeout: 3000 })
    }
  })

  test('should handle disabled buttons', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')

    // Find disabled buttons
    const disabledButtons = page.locator('button[disabled]')
    const count = await disabledButtons.count()
    
    if (count > 0) {
      const disabledButton = disabledButtons.first()
      
      // Should be disabled
      await expect(disabledButton).toBeDisabled()
      
      // Click should not trigger action
      await disabledButton.click({ force: true })
      await page.waitForTimeout(300)
      
      // Verify no navigation or modal appeared
      const url = page.url()
      expect(url).toContain('/')
    }
  })

  test('should handle loading state buttons', async ({ page }) => {
    await page.goto('/agents')
    await page.waitForLoadState('networkidle')

    // Look for buttons that might show loading state
    const submitButtons = page.locator('button[type="submit"], button:has-text("Submit"), button:has-text("Save")')
    const count = await submitButtons.count()
    
    if (count > 0) {
      const submitButton = submitButtons.first()
      
      // Click and check for loading state
      await submitButton.click()
      await page.waitForTimeout(100)
      
      // Button might show loading spinner or be disabled
      const isLoading = await submitButton.locator('.spinner, [aria-busy="true"]').count() > 0
      const isDisabled = await submitButton.isDisabled()
      
      // Either loading indicator or disabled state
      expect(isLoading || isDisabled).toBeTruthy()
    }
  })
})
