import { test, expect } from '@playwright/test';

const routes = {
  login: '/',
  register: '/auth/register'
};

test.describe('Authentication shell', () => {
  test('shows login form by default', async ({ page }) => {
    await page.goto(routes.login);
    await expect(page.getByRole('heading', { name: /sign in/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /sign in/i })).toBeVisible();
  });

  test('navigates to registration screen', async ({ page }) => {
    await page.goto(routes.login);
    await page.getByRole('button', { name: /create one/i }).click();
    await expect(page).toHaveURL(/auth\/register/);
    await expect(page.getByRole('heading', { name: /join the smart insurance workspace/i })).toBeVisible();
  });

  test('redirects unauthenticated dashboard access back to login with returnUrl', async ({ page }) => {
    await page.goto('/dashboard');
    await expect(page).toHaveURL(/auth\/login\?returnUrl=%2Fdashboard/);
    await expect(page.getByRole('heading', { name: /sign in/i })).toBeVisible();
  });
});
