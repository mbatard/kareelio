# Marketing Screenshots

## Prerequisites

- Docker Compose running (`docker compose up -d`)
- Node.js 20+
- Playwright installed

## Setup

```bash
# Install Playwright
npm install playwright
npx playwright install chromium

# Seed demo data
docker exec -i kareelio-postgres-1 psql -U kareelio -d kareelio < marketing/seed-screenshots.sql
```

## Take screenshots

```bash
node marketing/take-screenshots.js
```

Screenshots are saved to `marketing/screenshots/`.

## Files produced

| File | Description |
|------|-------------|
| `dashboard-applications-light.png` | Main dashboard, light mode, 1440x900 |
| `dashboard-applications-dark.png` | Main dashboard, dark mode, 1440x900 |
| `application-detail-light.png` | Application form/detail, light mode, 1440x900 |
| `profile-preferences-light.png` | Profile page, light mode |
| `login-light.png` | Login page, light mode |
| `admin-dashboard-dark.png` | Admin KPIs, dark mode |
| `admin-users-light.png` | Admin user management, light mode |

## Cleanup

To remove demo data:

```bash
docker exec -i kareelio-postgres-1 psql -U kareelio -d kareelio -c "
DELETE FROM job_applications WHERE owner_user_id = (SELECT id FROM users WHERE email = 'jean.dupont@example.com');
DELETE FROM users WHERE email = 'jean.dupont@example.com';
"
```
