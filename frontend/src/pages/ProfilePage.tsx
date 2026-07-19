import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { profileApi } from '../services/api';
import { useAuth } from '../contexts/AuthContext';
import { useTheme } from '../contexts/ThemeContext';
import { useLanguage } from '../contexts/LanguageContext';
import { isValidEmail } from '../utils/email';
import type { User } from '../types';

export function ProfilePage() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState('');
  const [passwordMsg, setPasswordMsg] = useState('');
  const { t } = useTranslation();
  const { theme, setTheme } = useTheme();
  const { language, setLanguage } = useLanguage();
  const navigate = useNavigate();
  const { isAdmin } = useAuth();

  const [displayName, setDisplayName] = useState('');
  const [email, setEmail] = useState('');
  const [description, setDescription] = useState('');
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [emailError, setEmailError] = useState('');

  useEffect(() => {
    profileApi.get().then((u) => {
      setUser(u);
      setDisplayName(u.display_name);
      setEmail(u.email);
      setDescription(u.description);
      setLoading(false);
    }).catch(() => navigate('/login'));
  }, [navigate]);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setEmailError('');
    if (!isAdmin) {
      const trimmedEmail = email.trim();
      if (!isValidEmail(trimmedEmail)) {
        setEmailError(t('common.invalidEmail'));
        return;
      }
    }
    setSaving(true);
    setMessage('');
    try {
      const updated = await profileApi.update({
        display_name: displayName,
        ...(isAdmin ? {} : { email: email.trim() }),
        description,
        language,
        theme,
      });
      setUser(updated);
      setMessage(t('profile.saved'));
    } catch {
      setMessage(t('common.error'));
    } finally {
      setSaving(false);
    }
  };

  const handlePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setPasswordMsg('');
    try {
      await profileApi.changePassword(currentPassword, newPassword);
      setPasswordMsg(t('profile.passwordUpdated'));
      setCurrentPassword('');
      setNewPassword('');
    } catch {
      setPasswordMsg(t('common.error'));
    }
  };

  if (loading || !user) {
    return <div className="flex justify-center items-center h-64"><p className="text-gray-500">{t('common.loading')}</p></div>;
  }

  return (
    <div className="max-w-2xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold mb-6 text-gray-900 dark:text-white">{t('profile.title')}</h1>

      <form onSubmit={handleSave} className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 space-y-4 mb-8">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('profile.displayName')}</label>
          <input value={displayName} onChange={(e) => setDisplayName(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('profile.email')}</label>
          <input type="email" value={email} onChange={(e) => { setEmail(e.target.value); setEmailError(''); }} disabled={isAdmin}
            className={`w-full px-3 py-2 border rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none disabled:bg-gray-100 dark:disabled:bg-gray-900 disabled:cursor-not-allowed disabled:text-gray-500 dark:disabled:text-gray-500 ${
              emailError ? 'border-red-500 dark:border-red-500' : 'border-gray-300 dark:border-gray-600'
            }`} />
          {emailError && <p className="text-red-500 text-xs mt-1">{emailError}</p>}
          {isAdmin && !emailError && <p className="text-xs text-gray-400 mt-1">{t('profile.adminEmailLocked')}</p>}
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('profile.description')}</label>
          <textarea value={description} onChange={(e) => setDescription(e.target.value)} rows={3}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" />
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('profile.language')}</label>
            <select value={language} onChange={(e) => setLanguage(e.target.value as any)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none">
              <option value="system">{t('profile.language')} (System)</option>
              <option value="fr">Francais</option>
              <option value="en">English</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('profile.theme')}</label>
            <select value={theme} onChange={(e) => setTheme(e.target.value as any)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none">
              <option value="system">{t('profile.theme')} (System)</option>
              <option value="light">{t('profile.theme')} Light</option>
              <option value="dark">{t('profile.theme')} Dark</option>
            </select>
          </div>
        </div>
        {message && <p className="text-green-600 dark:text-green-400 text-sm">{message}</p>}
        <button type="submit" disabled={saving}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-md font-medium disabled:opacity-50 transition-colors">
          {saving ? t('common.loading') : t('profile.save')}
        </button>
      </form>

      <form onSubmit={handlePassword} className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 space-y-4">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-white">{t('profile.changePassword')}</h2>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('profile.currentPassword')}</label>
          <input type="password" value={currentPassword} onChange={(e) => setCurrentPassword(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" required />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('profile.newPassword')}</label>
          <input type="password" value={newPassword} onChange={(e) => setNewPassword(e.target.value)} minLength={8}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" required />
        </div>
        {passwordMsg && <p className="text-green-600 dark:text-green-400 text-sm">{passwordMsg}</p>}
        <button type="submit"
          className="px-4 py-2 bg-gray-600 hover:bg-gray-700 text-white rounded-md font-medium transition-colors">
          {t('profile.changePassword')}
        </button>
      </form>
    </div>
  );
}
