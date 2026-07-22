import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { userApi } from '../services/api';
import { isValidEmail } from '../utils/email';
import type { User } from '../types';

export function AdminUsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [formData, setFormData] = useState({ email: '', password: '', display_name: '', description: '' });
  const [emailError, setEmailError] = useState('');
  const { t } = useTranslation();

  const loadUsers = () => {
    userApi.list().then(setUsers).finally(() => setLoading(false));
  };

  useEffect(() => { loadUsers(); }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setEmailError('');
    const email = formData.email.trim();
    if (!isValidEmail(email)) {
      setEmailError(t('common.invalidEmail'));
      return;
    }
    await userApi.create({ ...formData, email });
    setFormData({ email: '', password: '', display_name: '', description: '' });
    setShowForm(false);
    loadUsers();
  };

  const handleToggleActive = async (user: User) => {
    await userApi.update(user.id, { is_active: !user.is_active });
    loadUsers();
  };

  const handleDelete = async (user: User) => {
    if (!confirm(t('admin.confirmDelete'))) return;
    await userApi.delete(user.id);
    loadUsers();
  };

  if (loading) {
    return <div className="flex justify-center items-center h-64"><p className="text-gray-500">{t('common.loading')}</p></div>;
  }

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('admin.users')}</h1>
        <button onClick={() => setShowForm(!showForm)}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-md text-sm font-medium transition-colors">
          {t('admin.addUser')}
        </button>
      </div>

      {showForm && (
        <form onSubmit={handleCreate} className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 mb-6 space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <input placeholder={t('admin.email')} type="email" required value={formData.email}
              onChange={(e) => { setFormData({ ...formData, email: e.target.value }); setEmailError(''); }}
              className={`px-3 py-2 border rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none ${
                emailError ? 'border-red-500 dark:border-red-500' : 'border-gray-300 dark:border-gray-600'
              }`} />
            {emailError && <p className="text-red-500 text-xs mt-1">{emailError}</p>}
            <input placeholder={t('admin.displayName')} required value={formData.display_name}
              onChange={(e) => setFormData({ ...formData, display_name: e.target.value })}
              className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" />
            <input placeholder={t('auth.password')} type="password" required value={formData.password}
              onChange={(e) => setFormData({ ...formData, password: e.target.value })}
              className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" />
          </div>
          <textarea placeholder={t('profile.description')} value={formData.description}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })} rows={2}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" />
          <div className="flex space-x-2">
            <button type="submit" className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-md text-sm font-medium transition-colors">
              {t('common.create')}
            </button>
            <button type="button" onClick={() => setShowForm(false)}
              className="px-4 py-2 bg-gray-300 dark:bg-gray-600 hover:bg-gray-400 dark:hover:bg-gray-500 text-gray-700 dark:text-gray-200 rounded-md text-sm font-medium transition-colors">
              {t('common.cancel')}
            </button>
          </div>
        </form>
      )}

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead className="bg-gray-50 dark:bg-gray-700">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">{t('admin.email')}</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">{t('admin.displayName')}</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">{t('admin.role')}</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">{t('admin.emailVerified')}</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">{t('admin.active')}</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">{t('admin.actions')}</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
            {users.map((u) => (
              <tr key={u.id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                <td className="px-6 py-4 text-sm">
                  {u.role !== 'admin' ? (
                    <Link to={`/admin/users/${u.id}/edit`} className="text-blue-600 hover:underline">{u.email}</Link>
                  ) : (
                    <span className="text-gray-900 dark:text-gray-300">{u.email}</span>
                  )}
                </td>
                <td className="px-6 py-4 text-sm text-gray-900 dark:text-gray-300">{u.display_name}</td>
                <td className="px-6 py-4">
                  <span className={`px-2 py-1 rounded text-xs font-medium ${u.role === 'admin' ? 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300' : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'}`}>
                    {u.role}
                  </span>
                </td>
                <td className="px-6 py-4">
                  <span className={`px-2 py-1 rounded text-xs font-medium ${
                    u.email_verified_at
                      ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
                      : 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300'
                  }`}>
                    {u.email_verified_at ? t('admin.verified') : t('admin.pending')}
                  </span>
                </td>
                <td className="px-6 py-4">
                  {u.role !== 'admin' ? (
                    <button
                      onClick={() => handleToggleActive(u)}
                      className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors cursor-pointer ${
                        u.is_active ? 'bg-blue-600' : 'bg-gray-300 dark:bg-gray-600'
                      }`}
                    >
                      <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                        u.is_active ? 'translate-x-6' : 'translate-x-1'
                      }`} />
                    </button>
                  ) : (
                    <span className={`w-2 h-2 inline-block rounded-full ${u.is_active ? 'bg-green-500' : 'bg-red-500'}`} />
                  )}
                </td>
                <td className="px-6 py-4 space-x-2">
                  {u.role !== 'admin' ? (
                    <button onClick={() => handleDelete(u)}
                      className="text-sm text-red-600 hover:underline cursor-pointer">
                      {t('common.delete')}
                    </button>
                  ) : (
                    <span className="text-xs text-gray-400">Protected</span>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
