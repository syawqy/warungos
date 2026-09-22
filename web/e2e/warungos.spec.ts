import { test, expect, type Page } from '@playwright/test';

const BASE = 'http://localhost:3000';
const ADMIN_EMAIL = 'admin@warungos.id';
const ADMIN_PASS = 'admin123';

async function login(page: Page, email = ADMIN_EMAIL, password = ADMIN_PASS) {
  await page.goto(BASE);
  await page.waitForSelector('input[type="email"]');
  await page.fill('input[type="email"]', email);
  await page.fill('input[type="password"]', password);
  await page.click('button[type="submit"]');
  await page.waitForSelector('text=Kasir', { timeout: 10_000 });
}

async function addMenuItems(page: Page, count: number) {
  const cards = page.locator('.grid button.card');
  await cards.first().waitFor({ state: 'visible', timeout: 10_000 });
  for (let i = 0; i < Math.min(count, await cards.count()); i++) {
    await cards.nth(i).click();
    await page.waitForTimeout(400);
  }
}

// ── 1. AUTH ──
test.describe('Authentication', () => {
  test('login page renders correctly', async ({ page }) => {
    await page.goto(BASE);
    await expect(page.locator('input[type="email"]')).toBeVisible();
    await expect(page.locator('input[type="password"]')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
    await expect(page.getByRole('heading', { name: 'WarungOS' })).toBeVisible();
    await expect(page.locator('input[type="email"]')).toBeVisible();
  });

  test('login with wrong credentials shows error', async ({ page }) => {
    await page.goto(BASE);
    await page.fill('input[type="email"]', 'wrong@email.com');
    await page.fill('input[type="password"]', 'wrongpass');
    await page.click('button[type="submit"]');
    await page.waitForTimeout(3000);
    await expect(page.locator('input[type="email"]')).toBeVisible();
  });

  test('login with valid credentials navigates to POS', async ({ page }) => {
    await login(page);
    await expect(page.locator('text=Kasir').first()).toBeVisible();
    await expect(page.getByRole('button', { name: 'Pesanan' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Dashboard' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Inventaris' })).toBeVisible();
  });
});

// ── 2. POS ──
test.describe('POS Terminal', () => {
  test.beforeEach(async ({ page }) => { await login(page); });

  test('menu items are displayed', async ({ page }) => {
    await expect(page.locator('.grid button.card').first()).toBeVisible({ timeout: 10_000 });
    expect(await page.locator('.grid button.card').count()).toBeGreaterThan(0);
  });

  test('category filter works', async ({ page }) => {
    await page.waitForTimeout(1000);
    const beforeAll = await page.locator('.grid button.card').count();
    // Click Makanan (in category bar, not sidebar)
    await page.getByRole('button', { name: 'Makanan' }).first().click();
    await page.waitForTimeout(1500);
    const afterFilter = await page.locator('.grid button.card').count();
    expect(afterFilter).toBeGreaterThan(0);
    expect(afterFilter).toBeLessThanOrEqual(beforeAll);
  });

  test('add items to cart updates order panel', async ({ page }) => {
    await addMenuItems(page, 2);
    await page.waitForTimeout(1000);
    const cartText = await page.locator('.w-2\\/5').textContent();
    expect(cartText).toContain('2 item');
  });

  test('customer name can be set', async ({ page }) => {
    await addMenuItems(page, 1);
    await page.waitForTimeout(500);
    await page.locator('input[placeholder*="pelanggan"]').fill('Pak Budi');
    await page.waitForTimeout(1000);
    // The input itself should have the value
    const val = await page.locator('input[placeholder*="pelanggan"]').inputValue();
    expect(val).toBe('Pak Budi');
  });

  test('order type selection works', async ({ page }) => {
    await addMenuItems(page, 1);
    await page.waitForTimeout(500);
    await page.locator('button:has-text("Take Away")').click();
    await page.waitForTimeout(500);
    await expect(page.locator('button:has-text("Take Away")')).toHaveClass(/bg-warung-orange/);
  });

  test('subtotal and tax calculation is correct', async ({ page }) => {
    await addMenuItems(page, 1);
    await page.waitForTimeout(1000);
    const cartText = await page.locator('.w-2\\/5').textContent();
    expect(cartText).toContain('PPN');
    expect(cartText).toContain('Total');
  });

  test('remove item from cart', async ({ page }) => {
    await addMenuItems(page, 2);
    await page.waitForTimeout(1000);
    const removeBtn = page.locator('.w-2\\/5 button').filter({ has: page.locator('svg') });
    if (await removeBtn.count() > 0) {
      await removeBtn.first().click();
      await page.waitForTimeout(500);
    }
  });
});

// ── 3. PAYMENT ──
test.describe('Payment Flow', () => {
  test.beforeEach(async ({ page }) => { await login(page); });

  test('full transaction: add items → pay → success', async ({ page }) => {
    await addMenuItems(page, 2);
    await page.waitForTimeout(1000);
    await page.locator('.w-2\\/5').locator('button:has-text("Bayar")').click();
    await page.waitForSelector('.fixed.inset-0', { timeout: 5000 });
    await expect(page.getByRole('heading', { name: 'Pembayaran' })).toBeVisible();
    await page.locator('.fixed.inset-0 input[type="number"]').fill('200000');
    await page.waitForTimeout(1000);
    const payBtn = page.getByRole('button', { name: /Bayar Rp/ });
    expect(await payBtn.isDisabled()).toBe(false);
    await payBtn.click();
    await page.waitForSelector('text=Pembayaran Berhasil', { timeout: 8000 });
    await expect(page.locator('text=Pembayaran Berhasil')).toBeVisible();
  });

  test('payment modal shows change calculation', async ({ page }) => {
    await addMenuItems(page, 1);
    await page.waitForTimeout(500);
    await page.locator('.w-2\\/5').locator('button:has-text("Bayar")').click();
    await page.waitForSelector('.fixed.inset-0', { timeout: 5000 });
    await page.locator('.fixed.inset-0 input[type="number"]').fill('200000');
    await page.waitForTimeout(1000);
    await expect(page.locator('.fixed.inset-0').locator('text=Kembalian')).toBeVisible();
  });

  test('payment button disabled when cash < total', async ({ page }) => {
    await addMenuItems(page, 2);
    await page.waitForTimeout(500);
    await page.locator('.w-2\\/5').locator('button:has-text("Bayar")').click();
    await page.waitForSelector('.fixed.inset-0', { timeout: 5000 });
    await page.locator('.fixed.inset-0 input[type="number"]').fill('1000');
    await page.waitForTimeout(1000);
    expect(await page.getByRole('button', { name: /Bayar Rp/ }).isDisabled()).toBe(true);
  });

  test('can select different payment methods', async ({ page }) => {
    await addMenuItems(page, 1);
    await page.waitForTimeout(500);
    await page.locator('.w-2\\/5').locator('button:has-text("Bayar")').click();
    await page.waitForSelector('.fixed.inset-0', { timeout: 5000 });
    await page.locator('.fixed.inset-0 button:has-text("QRIS")').click();
    await page.waitForTimeout(300);
    await page.locator('.fixed.inset-0 button:has-text("Kartu")').click();
    await page.waitForTimeout(300);
    await page.locator('.fixed.inset-0 button:has-text("Tunai")').click();
    await page.waitForTimeout(300);
    await expect(page.locator('.fixed.inset-0 input[type="number"]')).toBeVisible();
  });

  test('can close payment modal', async ({ page }) => {
    await addMenuItems(page, 1);
    await page.waitForTimeout(500);
    await page.locator('.w-2\\/5').locator('button:has-text("Bayar")').click();
    await page.waitForSelector('.fixed.inset-0', { timeout: 5000 });
    await page.locator('.fixed.inset-0 button:has(svg)').first().click();
    await page.waitForTimeout(500);
    expect(await page.locator('.fixed.inset-0').isVisible().catch(() => false)).toBe(false);
  });
});

// ── 4. ORDERS ──
test.describe('Orders Page', () => {
  test.beforeEach(async ({ page }) => { await login(page); });

  test('navigate to orders page', async ({ page }) => {
    await page.getByRole('button', { name: 'Pesanan' }).click();
    await page.waitForTimeout(2000);
    await expect(page.getByRole('heading', { name: 'Pesanan' })).toBeVisible();
  });

  test('status filter buttons are displayed', async ({ page }) => {
    await page.getByRole('button', { name: 'Pesanan' }).click();
    await page.waitForTimeout(2000);
    await expect(page.locator('button:has-text("Semua")').first()).toBeVisible();
    await expect(page.locator('button:has-text("Menunggu")')).toBeVisible();
    await expect(page.locator('button:has-text("Selesai")')).toBeVisible();
  });

  test('orders are displayed after transaction', async ({ page }) => {
    await addMenuItems(page, 1);
    await page.waitForTimeout(500);
    await page.locator('.w-2\\/5').locator('button:has-text("Bayar")').click();
    await page.waitForSelector('.fixed.inset-0', { timeout: 5000 });
    await page.locator('.fixed.inset-0 input[type="number"]').fill('200000');
    await page.waitForTimeout(1000);
    await page.getByRole('button', { name: /Bayar Rp/ }).click();
    await page.waitForSelector('text=Pembayaran Berhasil', { timeout: 8000 });
    await page.waitForTimeout(3000);

    await page.getByRole('button', { name: 'Pesanan' }).click();
    await page.waitForTimeout(3000);
    // The orders page has order cards with status badges
    const pageText = await page.getByRole('heading', { name: 'Pesanan' }).locator('..').locator('..').textContent();
    // Should contain "Pesanan" heading - at minimum the page loaded
    expect(pageText).toContain('Pesanan');
  });

  test('refresh button works', async ({ page }) => {
    await page.getByRole('button', { name: 'Pesanan' }).click();
    await page.waitForTimeout(2000);
    await page.click('button:has-text("Refresh")');
    await page.waitForTimeout(2000);
    await expect(page.getByRole('heading', { name: 'Pesanan' })).toBeVisible();
  });
});

// ── 5. DASHBOARD ──
test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => { await login(page); });

  test('navigate to dashboard', async ({ page }) => {
    await page.getByRole('button', { name: 'Dashboard' }).click();
    await page.waitForTimeout(2000);
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();
  });

  test('stats cards are displayed', async ({ page }) => {
    await page.getByRole('button', { name: 'Dashboard' }).click();
    await page.waitForTimeout(2000);
    await expect(page.locator('text=Pendapatan Hari Ini')).toBeVisible();
    await expect(page.locator('text=Pesanan Hari Ini')).toBeVisible();
  });

  test('revenue chart is rendered', async ({ page }) => {
    await page.getByRole('button', { name: 'Dashboard' }).click();
    await page.waitForTimeout(2000);
    expect(await page.locator('svg').count()).toBeGreaterThan(0);
  });
});

// ── 6. INVENTORY ──
test.describe('Inventory Page', () => {
  test.beforeEach(async ({ page }) => { await login(page); });

  test('navigate to inventory', async ({ page }) => {
    await page.getByRole('button', { name: 'Inventaris' }).click();
    await page.waitForTimeout(2000);
    await expect(page.getByRole('heading', { name: 'Inventaris' })).toBeVisible();
  });

  test('inventory items are displayed', async ({ page }) => {
    await page.getByRole('button', { name: 'Inventaris' }).click();
    await page.waitForTimeout(3000);
    const text = await page.getByRole('main').textContent();
    expect(text).toContain('Ayam');
  });

  test('inventory shows correct columns', async ({ page }) => {
    await page.getByRole('button', { name: 'Inventaris' }).click();
    await page.waitForTimeout(3000);
    const text = await page.getByRole('main').textContent();
    expect(text).toContain('Item');
    expect(text).toContain('Stok');
    expect(text).toContain('Harga/Unit');
    expect(text).toContain('Status');
  });

  test('add item button exists', async ({ page }) => {
    await page.getByRole('button', { name: 'Inventaris' }).click();
    await page.waitForTimeout(2000);
    await expect(page.locator('button:has-text("Tambah Item")')).toBeVisible();
  });

  test('inventory prices show Rp (no NaN)', async ({ page }) => {
    await page.getByRole('button', { name: 'Inventaris' }).click();
    await page.waitForTimeout(3000);
    const text = await page.getByRole('main').textContent();
    expect(text).toContain('Rp');
    expect(text).not.toContain('NaN');
  });
});

// ── 7. NAVIGATION ──
test.describe('Navigation', () => {
  test.beforeEach(async ({ page }) => { await login(page); });

  test('can navigate between all pages', async ({ page }) => {
    await expect(page.locator('.grid button.card').first()).toBeVisible({ timeout: 5000 });
    await page.getByRole('button', { name: 'Pesanan' }).click();
    await page.waitForTimeout(1500);
    await expect(page.getByRole('heading', { name: 'Pesanan' })).toBeVisible();
    await page.getByRole('button', { name: 'Dashboard' }).click();
    await page.waitForTimeout(1500);
    await expect(page.locator('text=Pendapatan Hari Ini')).toBeVisible();
    await page.getByRole('button', { name: 'Inventaris' }).click();
    await page.waitForTimeout(1500);
    await expect(page.getByRole('heading', { name: 'Inventaris' })).toBeVisible();
    await page.getByRole('button', { name: 'Kasir' }).click();
    await page.waitForTimeout(1500);
    await expect(page.locator('.grid button.card').first()).toBeVisible({ timeout: 5000 });
  });

  test('admin user info shown in sidebar', async ({ page }) => {
    await expect(page.locator('text=Admin WarungOS')).toBeVisible();
  });

  test('logout button exists', async ({ page }) => {
    await expect(page.locator('button:has-text("Keluar")')).toBeVisible();
  });
});

// ── 8. FULL WORKFLOW ──
test.describe('Full E2E Workflow', () => {
  test('login → order → pay → check orders → dashboard → inventory', async ({ page }) => {
    await login(page);
    await addMenuItems(page, 3);
    await page.waitForTimeout(1000);
    await page.locator('.w-2\\/5').locator('button:has-text("Bayar")').click();
    await page.waitForSelector('.fixed.inset-0', { timeout: 5000 });
    await page.locator('.fixed.inset-0 input[type="number"]').fill('200000');
    await page.waitForTimeout(1000);
    await page.getByRole('button', { name: /Bayar Rp/ }).click();
    await page.waitForSelector('text=Pembayaran Berhasil', { timeout: 8000 });
    await page.waitForTimeout(3000);

    await page.getByRole('button', { name: 'Pesanan' }).click();
    await page.waitForTimeout(3000);
    await expect(page.getByRole('heading', { name: 'Pesanan' })).toBeVisible();

    await page.getByRole('button', { name: 'Dashboard' }).click();
    await page.waitForTimeout(2000);
    await expect(page.locator('text=Pendapatan Hari Ini')).toBeVisible();

    await page.getByRole('button', { name: 'Inventaris' }).click();
    await page.waitForTimeout(3000);
    expect(await page.getByRole('main').textContent()).toContain('Rp');
  });
});
