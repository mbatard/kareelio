import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { Navbar } from './Navbar';

const adminApiMock = vi.hoisted(() => ({
  notificationSummary: vi.fn(),
  acknowledgeUserRegistrations: vi.fn(),
}));

vi.mock('../contexts/AuthContext', () => ({
  useAuth: () => ({
    user: { id: 'admin-1', display_name: 'Admin', email: 'admin@example.com', role: 'admin' },
    logout: vi.fn(),
    isAdmin: true,
  }),
}));

vi.mock('../services/api', () => ({
  adminApi: adminApiMock,
}));

vi.mock('./Toggles', () => ({
  ThemeToggle: () => <button type="button">theme</button>,
  LanguageToggle: () => <button type="button">language</button>,
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: { count?: number }) => {
      if (key === 'nav.adminUsersBadge') {
        return `${opts?.count ?? 0} new registrations`;
      }
      const labels: Record<string, string> = {
        'nav.adminDashboard': 'Dashboard',
        'nav.adminUsers': 'Users',
        'nav.audit': 'Audit',
        'nav.profile': 'Profile',
        'nav.about': 'About',
        'nav.logout': 'Logout',
      };
      return labels[key] ?? key;
    },
  }),
}));

describe('Navbar admin notification badge', () => {
  beforeEach(() => {
    adminApiMock.notificationSummary.mockReset();
    adminApiMock.acknowledgeUserRegistrations.mockReset();
  });

  it('shows the unread count for admin users', async () => {
    adminApiMock.notificationSummary.mockResolvedValueOnce({ new_user_registrations: 3 });

    render(
      <MemoryRouter initialEntries={['/admin']}>
        <Navbar />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByLabelText('3 new registrations')).toBeInTheDocument();
    });
    expect(adminApiMock.notificationSummary).toHaveBeenCalled();
  });

  it('acknowledges when the admin users page is open', async () => {
    adminApiMock.notificationSummary
      .mockResolvedValueOnce({ new_user_registrations: 2 })
      .mockResolvedValueOnce({ new_user_registrations: 0 });
    adminApiMock.acknowledgeUserRegistrations.mockResolvedValueOnce(undefined);

    render(
      <MemoryRouter initialEntries={['/admin/users']}>
        <Navbar />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(adminApiMock.acknowledgeUserRegistrations).toHaveBeenCalledTimes(1);
    });
    await waitFor(() => {
      expect(adminApiMock.notificationSummary).toHaveBeenCalled();
    });
  });
});
