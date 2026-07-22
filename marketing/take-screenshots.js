const { chromium } = require('playwright');
const fs = require('fs');
const path = require('path');

const BASE = process.env.BASE_URL || 'http://localhost:5173';
const OUT = path.join(__dirname, 'screenshots');

async function login(page, email, password) {
  await page.goto(BASE + '/login', { waitUntil: 'networkidle' });
  await page.waitForSelector('input[type="email"]', { timeout: 10000 });
  await page.fill('input[type="email"]', email);
  await page.fill('input[type="password"]', password);

  // Intercept login response to capture session cookie
  const [response] = await Promise.all([
    page.waitForResponse(resp => resp.url().includes('/api/auth/login') && resp.status() === 200, { timeout: 10000 }),
    page.click('button[type="submit"]'),
  ]);

  // Extract session_id from Set-Cookie
  const headers = response.headers();
  const setCookie = headers['set-cookie'] || '';
  const match = setCookie.match(/session_id=([^;]+)/);
  if (match) {
    await page.context().addCookies([{
      name: 'session_id',
      value: match[1],
      domain: 'localhost',
      path: '/',
      httpOnly: true,
      secure: false,
      sameSite: 'Lax',
    }]);
  }

  await page.waitForTimeout(500);
  await page.waitForLoadState('networkidle');
}

async function setTheme(page, theme) {
  await page.evaluate((t) => {
    localStorage.setItem('kareelio_theme', t);
    document.documentElement.classList.toggle('dark', t === 'dark');
  }, theme);
}

(async () => {
  fs.mkdirSync(OUT, { recursive: true });

  const browser = await chromium.launch({ args: ['--no-sandbox'] });

  // ===== 1. Dashboard applications LIGHT 1440x900 =====
  console.log('1. Desktop Light - Applications');
  const ctx1 = await browser.newContext({
    viewport: { width: 1440, height: 900 },
    deviceScaleFactor: 2,
  });
  const p1 = await ctx1.newPage();
  await login(p1, 'jean.dupont@example.com', 'demo1234');
  await setTheme(p1, 'light');
  await p1.goto(BASE + '/applications', { waitUntil: 'networkidle' });
  await p1.waitForTimeout(1500);
  await p1.screenshot({ path: `${OUT}/dashboard-applications-light.png` });
  await ctx1.close();

  // ===== 2. Dashboard applications DARK 1440x900 =====
  console.log('2. Desktop Dark - Applications');
  const ctx2 = await browser.newContext({
    viewport: { width: 1440, height: 900 },
    deviceScaleFactor: 2,
  });
  const p2 = await ctx2.newPage();
  await login(p2, 'jean.dupont@example.com', 'demo1234');
  await setTheme(p2, 'dark');
  await p2.goto(BASE + '/applications', { waitUntil: 'networkidle' });
  await p2.waitForTimeout(1500);
  await p2.screenshot({ path: `${OUT}/dashboard-applications-dark.png` });
  await ctx2.close();

  // ===== 3. Application detail LIGHT 1440x900 =====
  console.log('3. Desktop Light - Application Detail');
  const ctx3 = await browser.newContext({
    viewport: { width: 1440, height: 900 },
    deviceScaleFactor: 2,
  });
  const p3 = await ctx3.newPage();
  await login(p3, 'jean.dupont@example.com', 'demo1234');
  await setTheme(p3, 'light');
  // Get first app ID via API
  const appsResp = await p3.evaluate(async () => {
    const r = await fetch('/api/job-applications', { credentials: 'include' });
    return r.json();
  });
  if (appsResp && appsResp.length > 0) {
    await p3.goto(BASE + `/applications/${appsResp[0].id}/edit`, { waitUntil: 'networkidle' });
    await p3.waitForTimeout(1500);
  }
  await p3.screenshot({ path: `${OUT}/application-detail-light.png` });
  await ctx3.close();

  // ===== 4. Profile LIGHT =====
  console.log('4. Desktop Light - Profile');
  const ctx5 = await browser.newContext({
    viewport: { width: 1440, height: 900 },
    deviceScaleFactor: 2,
  });
  const p5 = await ctx5.newPage();
  await login(p5, 'jean.dupont@example.com', 'demo1234');
  await setTheme(p5, 'light');
  await p5.goto(BASE + '/profile', { waitUntil: 'networkidle' });
  await p5.waitForTimeout(1500);
  await p5.screenshot({ path: `${OUT}/profile-preferences-light.png` });
  await ctx5.close();

  // ===== 5. Login page LIGHT =====
  console.log('5. Desktop Light - Login');
  const ctx6 = await browser.newContext({
    viewport: { width: 1440, height: 900 },
    deviceScaleFactor: 2,
  });
  const p6 = await ctx6.newPage();
  await p6.goto(BASE + '/login', { waitUntil: 'networkidle' });
  await p6.waitForTimeout(1000);
  await p6.screenshot({ path: `${OUT}/login-light.png` });
  await ctx6.close();

  // ===== 6. Admin Dashboard DARK =====
  console.log('6. Desktop Dark - Admin Dashboard');
  const ctx7 = await browser.newContext({
    viewport: { width: 1440, height: 900 },
    deviceScaleFactor: 2,
  });
  const p7 = await ctx7.newPage();
  await login(p7, 'admin@kareelio.local', 'admin');
  await setTheme(p7, 'dark');
  await p7.goto(BASE + '/admin', { waitUntil: 'networkidle' });
  await p7.waitForTimeout(1500);
  await p7.screenshot({ path: `${OUT}/admin-dashboard-dark.png` });
  await ctx7.close();

  // ===== 7. Admin Users LIGHT =====
  console.log('7. Desktop Light - Admin Users');
  const ctx8 = await browser.newContext({
    viewport: { width: 1440, height: 900 },
    deviceScaleFactor: 2,
  });
  const p8 = await ctx8.newPage();
  await login(p8, 'admin@kareelio.local', 'admin');
  await setTheme(p8, 'light');
  await p8.goto(BASE + '/admin/users', { waitUntil: 'networkidle' });
  await p8.waitForTimeout(1500);
  await p8.screenshot({ path: `${OUT}/admin-users-light.png` });
  await ctx8.close();

  await browser.close();
  console.log('\nDone! Screenshots in: ' + OUT);
})();
