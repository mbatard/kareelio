import { useState, useEffect, useRef, useCallback } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { useTranslation } from 'react-i18next';
import { ThemeToggle } from './Toggles';
import { LanguageToggle } from './Toggles';
import { adminApi } from '../services/api';

export function Navbar() {
  const { user, logout, isAdmin } = useAuth();
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const [menuOpen, setMenuOpen] = useState(false);
  const [adminNotificationCount, setAdminNotificationCount] = useState<number | null>(null);
  const menuRef = useRef<HTMLDivElement>(null);
  const adminNotificationRequestRef = useRef<Promise<void> | null>(null);

  const loadAdminNotifications = useCallback(async () => {
    if (!user || !isAdmin) return;
    try {
      const summary = await adminApi.notificationSummary();
      setAdminNotificationCount(summary.new_user_registrations);
    } catch {
      // fail quietly
    }
  }, [isAdmin, user]);

  const acknowledgeAdminNotifications = useCallback(() => {
    if (!user || !isAdmin) return Promise.resolve();
    if (adminNotificationRequestRef.current) return adminNotificationRequestRef.current;

    adminNotificationRequestRef.current = (async () => {
      try {
        await adminApi.acknowledgeUserRegistrations();
        setAdminNotificationCount(0);
        await loadAdminNotifications();
      } catch {
        // fail quietly
      } finally {
        adminNotificationRequestRef.current = null;
      }
    })();

    return adminNotificationRequestRef.current;
  }, [isAdmin, loadAdminNotifications, user]);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false);
      }
    };
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setMenuOpen(false);
    };
    document.addEventListener('mousedown', handleClickOutside);
    document.addEventListener('keydown', handleEscape);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleEscape);
    };
  }, []);

  useEffect(() => {
    if (!user || !isAdmin) {
      setAdminNotificationCount(null);
      return;
    }

    if (location.pathname !== '/admin/users') {
      void loadAdminNotifications();
    }
    const interval = window.setInterval(() => {
      void loadAdminNotifications();
    }, 60000);

    return () => window.clearInterval(interval);
  }, [isAdmin, loadAdminNotifications, location.pathname, user]);

  useEffect(() => {
    if (!user || !isAdmin) return;
    if (location.pathname === '/admin/users') {
      void acknowledgeAdminNotifications();
    }
  }, [acknowledgeAdminNotifications, isAdmin, location.pathname, user]);

  const handleLogout = async () => {
    await logout();
    setMenuOpen(false);
    navigate('/login');
  };

  if (!user) return null;

  return (
    <nav className="bg-white dark:bg-gray-800 shadow-sm border-b border-gray-200 dark:border-gray-700">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16">
          <div className="flex items-center space-x-8">
            <Link to={isAdmin ? '/admin' : '/'} className="text-xl font-bold text-blue-600 dark:text-blue-400">
              Kareelio
            </Link>
            {!isAdmin && (
              <>
                <Link to="/" className="text-gray-700 dark:text-gray-300 hover:text-blue-600 dark:hover:text-blue-400">
                  {t('nav.dashboard')}
                </Link>
                <Link to="/applications" className="text-gray-700 dark:text-gray-300 hover:text-blue-600 dark:hover:text-blue-400">
                  {t('nav.applications')}
                </Link>
              </>
            )}
            {isAdmin && (
              <>
                <Link to="/admin" className="text-gray-700 dark:text-gray-300 hover:text-blue-600 dark:hover:text-blue-400">
                  {t('nav.adminDashboard')}
                </Link>
                <Link
                  to="/admin/users"
                  aria-label={adminNotificationCount && adminNotificationCount > 0 ? t('nav.adminUsersBadge', { count: adminNotificationCount }) : undefined}
                  className="text-gray-700 dark:text-gray-300 hover:text-blue-600 dark:hover:text-blue-400 inline-flex items-center gap-2"
                >
                  <span>{t('nav.adminUsers')}</span>
                  {adminNotificationCount && adminNotificationCount > 0 ? (
                    <span className="inline-flex min-w-5 items-center justify-center rounded-full bg-red-600 px-1.5 py-0.5 text-[10px] font-semibold leading-none text-white" aria-hidden="true">
                      {adminNotificationCount}
                    </span>
                  ) : null}
                </Link>
                <Link to="/admin/audit" className="text-gray-700 dark:text-gray-300 hover:text-blue-600 dark:hover:text-blue-400">
                  {t('nav.audit')}
                </Link>
              </>
            )}
          </div>

          <div className="flex items-center space-x-4">
            <LanguageToggle />
            <ThemeToggle />
            <div className="relative" ref={menuRef}>
              <button
                onClick={() => setMenuOpen(!menuOpen)}
                className="flex items-center space-x-2 text-gray-700 dark:text-gray-300 hover:text-blue-600 px-2 py-1 rounded-md"
                aria-expanded={menuOpen}
                aria-haspopup="true"
              >
                <span className="text-sm">{user.display_name || user.email}</span>
                <svg className={`w-4 h-4 transition-transform ${menuOpen ? 'rotate-180' : ''}`} fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                </svg>
              </button>
              {menuOpen && (
                <div className="absolute right-0 mt-2 w-48 bg-white dark:bg-gray-800 rounded-md shadow-lg py-1 z-50 border border-gray-200 dark:border-gray-700">
                  <Link
                    to="/profile"
                    onClick={() => setMenuOpen(false)}
                    className="block px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700"
                  >
                    {t('nav.profile')}
                  </Link>
                  <Link
                    to="/about"
                    onClick={() => setMenuOpen(false)}
                    className="block px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700"
                  >
                    {t('nav.about')}
                  </Link>
                  <hr className="border-gray-200 dark:border-gray-700" />
                  <button
                    onClick={handleLogout}
                    className="w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-gray-100 dark:hover:bg-gray-700"
                  >
                    {t('nav.logout')}
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </nav>
  );
}
