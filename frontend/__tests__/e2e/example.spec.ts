import { test, expect } from './fixtures'

test('homepage loads', async ({ page }) => {
  await page.goto('/')
  await page.waitForURL(/\/login|\/$/, { timeout: 10000 })
  await expect(page).toHaveTitle(/NeuronIP/i)
})

test('navigation works', async ({ page }) => {
  await page.goto('/agents')
  await page.waitForLoadState('domcontentloaded')
  await expect(page).toHaveURL(/\/agents/)
})
