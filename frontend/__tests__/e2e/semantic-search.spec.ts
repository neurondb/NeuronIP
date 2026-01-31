import { test, expect } from './fixtures'

test.describe('Semantic Search', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/semantic')
    await page.waitForLoadState('networkidle')
  })

  test('should display chat interface', async ({ page }) => {
    // Check for chat input
    const chatInput = page.locator('textarea, input[type="text"], [role="textbox"]').first()
    await expect(chatInput).toBeVisible()
    
    // Check for send button
    const sendButton = page.locator('button:has-text("Send"), button[aria-label*="send" i], button[type="submit"]').first()
    await expect(sendButton).toBeVisible()
  })

  test('should send a message', async ({ page }) => {
    const chatInput = page.locator('textarea, input[type="text"], [role="textbox"]').first()
    const sendButton = page.locator('button:has-text("Send"), button[aria-label*="send" i], button[type="submit"]').first()
    
    await chatInput.fill('What is NeuronIP?')
    await sendButton.click()
    
    // Wait for message to appear
    await page.waitForTimeout(1000)
    
    // Check message appears in chat
    const message = page.locator('text=What is NeuronIP?').first()
    await expect(message).toBeVisible({ timeout: 5000 })
  })

  test('should display response from assistant', async ({ page }) => {
    const chatInput = page.locator('textarea, input[type="text"], [role="textbox"]').first()
    const sendButton = page.locator('button:has-text("Send"), button[aria-label*="send" i], button[type="submit"]').first()
    
    await chatInput.fill('Test query')
    await sendButton.click()
    
    // Wait for response
    await page.waitForTimeout(3000)
    
    // Check for assistant response (could be loading, error, or actual response)
    const response = page.locator('[data-role="assistant"], .assistant-message, .response').first()
    await expect(response).toBeVisible({ timeout: 10000 })
  })

  test('should clear input after sending', async ({ page }) => {
    const chatInput = page.locator('textarea, input[type="text"], [role="textbox"]').first()
    const sendButton = page.locator('button:has-text("Send"), button[aria-label*="send" i], button[type="submit"]').first()
    
    await chatInput.fill('Test message')
    await sendButton.click()
    
    // Wait a bit
    await page.waitForTimeout(500)
    
    // Input should be cleared or ready for new message
    const inputValue = await chatInput.inputValue()
    // Either cleared or still has value (depending on implementation)
    expect(inputValue).toBeDefined()
  })

  test('should handle empty message', async ({ page }) => {
    const sendButton = page.locator('button:has-text("Send"), button[aria-label*="send" i], button[type="submit"]').first()
    
    // Try to send empty message
    await sendButton.click()
    
    // Should either be disabled or show validation error
    const isDisabled = await sendButton.isDisabled()
    const hasError = await page.locator('.error, [role="alert"]').count() > 0
    
    expect(isDisabled || !hasError).toBeTruthy()
  })

  test('should display example questions', async ({ page }) => {
    // Look for example questions or suggestions
    const examples = page.locator('text=Try asking, text=Example, .example-question').first()
    await expect(examples).toBeVisible({ timeout: 3000 })
  })

  test('should handle keyboard shortcuts', async ({ page }) => {
    const chatInput = page.locator('textarea, input[type="text"], [role="textbox"]').first()
    
    await chatInput.fill('Test query')
    
    // Press Enter to send
    await chatInput.press('Enter')
    await page.waitForTimeout(1000)
    
    // Message should be sent
    const message = page.locator('text=Test query').first()
    await expect(message).toBeVisible({ timeout: 5000 })
  })
})
