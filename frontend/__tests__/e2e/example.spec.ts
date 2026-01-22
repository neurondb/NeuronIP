import { test, expect } from '@playwright/test'

test('homepage loads', async ({ page }) => {
  await page.goto('/')
  await expect(page).toHaveTitle(/NeuronIP/i)
})

test('navigation works', async ({ page }) => {
  await page.goto('/')
  // Add navigation tests here
})
