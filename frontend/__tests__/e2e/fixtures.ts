import { test as base, expect } from '@playwright/test'

/**
 * E2E fixtures: mock auth + agents API so dashboard pages load without a real backend.
 * Use: import { test, expect } from './fixtures'
 */
export const test = base.extend({
  page: async ({ page }, use) => {
    await page.route('**/api/v1/auth/me', (route) =>
      route.fulfill({ status: 200, contentType: 'application/json', body: '{}' })
    )
    await page.route('**/api/v1/agents**', (route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ agents: [] }),
      })
    )
    await use(page)
  },
})

export { expect }
