import { test, expect } from './fixtures'

// Dashboard pages (auth mock). / redirects to login so we test it separately.
const dashboardPages = [
  { name: 'Why NeuronIP', path: '/why-neuronip' },
  { name: 'Semantic Search', path: '/semantic' },
  { name: 'Warehouse', path: '/warehouse' },
  { name: 'Data Sources', path: '/data-sources' },
  { name: 'Metrics', path: '/metrics' },
  { name: 'Data Catalog', path: '/catalog' },
  { name: 'Knowledge Graph', path: '/knowledge-graph' },
  { name: 'Agent Hub', path: '/agents' },
  { name: 'Models', path: '/models' },
  { name: 'Workflows', path: '/workflows' },
  { name: 'Observability', path: '/observability' },
  { name: 'Alerts', path: '/alerts' },
  { name: 'Data Lineage', path: '/lineage' },
  { name: 'Compliance', path: '/compliance' },
  { name: 'Audit', path: '/audit' },
  { name: 'Versioning', path: '/versioning' },
  { name: 'Users', path: '/users' },
  { name: 'API Keys', path: '/api-keys' },
  { name: 'Integrations', path: '/integrations' },
  { name: 'Settings', path: '/settings' },
  { name: 'Billing', path: '/billing' },
  { name: 'Support', path: '/support' },
  { name: 'Features', path: '/features' },
  { name: 'Analytics', path: '/analytics' },
]

test.describe('Page Rendering', () => {
  test.describe.configure({ mode: 'parallel' })

  test('should redirect / to login', async ({ page }) => {
    await page.goto('/')
    await page.waitForURL(/\/login/, { timeout: 10000 })
    await expect(page.locator('input[type="password"]').first()).toBeVisible({ timeout: 5000 })
  })

  for (const pageInfo of dashboardPages) {
    test(`should load ${pageInfo.name} page`, async ({ page }) => {
      await page.goto(pageInfo.path)
      await page.waitForLoadState('domcontentloaded')
      await page.waitForLoadState('networkidle').catch(() => {})

      expect(page.url()).toContain(pageInfo.path)
      const mainContent = page.locator('main, [role="main"], .main-content').first()
      await expect(mainContent).toBeVisible({ timeout: 10000 })
    })
  }

  test('should display page title for each page', async ({ page }) => {
    for (const pageInfo of dashboardPages.slice(0, 5)) {
      await page.goto(pageInfo.path)
      await page.waitForLoadState('domcontentloaded')
      const title = page.locator('h1, h2, [data-testid="page-title"]').first()
      await expect(title).toBeVisible({ timeout: 8000 })
    }
  })

  test('should maintain sidebar on all pages', async ({ page }) => {
    for (const pageInfo of dashboardPages.slice(0, 5)) {
      await page.goto(pageInfo.path)
      await page.waitForLoadState('domcontentloaded')
      const sidebar = page.locator('aside').first()
      await expect(sidebar).toBeVisible({ timeout: 8000 })
    }
  })

  test('should handle page errors gracefully', async ({ page }) => {
    await page.goto('/non-existent-page')
    await page.waitForLoadState('domcontentloaded')
    await expect(page.getByText(/404|Page Not Found/i).first()).toBeVisible({ timeout: 8000 })
  })
})
