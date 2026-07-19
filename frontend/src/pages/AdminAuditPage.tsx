import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { adminApi } from '../services/api';
import type { AuditEvent } from '../types';

const ACTION_COLORS: Record<string, string> = {
  login_success: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300',
  login_failed: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
  logout: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300',
  user_created: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300',
  user_updated: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300',
  user_activated: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300',
  user_deactivated: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300',
  user_deleted: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
  job_application_created: 'bg-cyan-100 text-cyan-800 dark:bg-cyan-900 dark:text-cyan-300',
  job_application_updated: 'bg-cyan-100 text-cyan-800 dark:bg-cyan-900 dark:text-cyan-300',
  job_application_deleted: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
  profile_updated: 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300',
  password_changed: 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-300',
  user_password_changed: 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-300',
};

export function AdminAuditPage() {
  const [events, setEvents] = useState<AuditEvent[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [offset, setOffset] = useState(0);
  const limit = 100;
  const { t } = useTranslation();

  const loadEvents = (newOffset: number) => {
    setLoading(true);
    adminApi.audit(limit, newOffset).then((res) => {
      setEvents(res.events);
      setTotal(res.total);
      setOffset(newOffset);
    }).finally(() => setLoading(false));
  };

  useEffect(() => { loadEvents(0); }, []);

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('audit.title')}</h1>
        <span className="text-sm text-gray-500 dark:text-gray-400">
          {t('audit.totalEvents', { count: total })}
        </span>
      </div>

      {loading ? (
        <p className="text-gray-500">{t('common.loading')}</p>
      ) : (
        <>
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                <thead className="bg-gray-50 dark:bg-gray-900">
                  <tr>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">{t('audit.dateTime')}</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">{t('audit.actor')}</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">IP</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">{t('audit.action')}</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">{t('audit.target')}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                  {events.map((e) => (
                    <tr key={e.id} className="hover:bg-gray-50 dark:hover:bg-gray-700/50">
                      <td className="px-4 py-3 text-sm text-gray-900 dark:text-white whitespace-nowrap">
                        {new Date(e.created_at).toLocaleString()}
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-700 dark:text-gray-300">
                        <div>{e.actor_email}</div>
                        <div className="text-xs text-gray-400">{e.actor_role}</div>
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-500 dark:text-gray-400 font-mono">
                        {e.actor_ip || '-'}
                      </td>
                      <td className="px-4 py-3">
                        <span className={`px-2 py-1 rounded text-xs font-medium ${ACTION_COLORS[e.action] ?? 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300'}`}>
                          {t(`audit.actions.${e.action}`, e.action)}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-700 dark:text-gray-300">
                        <div>{t(`audit.targetTypes.${e.target_type}`, e.target_type)}</div>
                        {typeof e.metadata?.target_email === 'string' && (
                          <div className="text-xs text-gray-400">{e.metadata.target_email}</div>
                        )}
                      </td>
                    </tr>
                  ))}
                  {events.length === 0 && (
                    <tr>
                      <td colSpan={5} className="px-4 py-8 text-center text-gray-400">
                        {t('common.noResults')}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>

          <div className="flex items-center justify-between mt-4">
            <p className="text-sm text-gray-500 dark:text-gray-400">
              {t('audit.showing', { from: offset + 1, to: Math.min(offset + limit, total), total })}
            </p>
            <div className="flex space-x-2">
              <button onClick={() => loadEvents(Math.max(0, offset - limit))} disabled={offset === 0}
                className="px-3 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed">
                {t('audit.previous')}
              </button>
              <button onClick={() => loadEvents(offset + limit)} disabled={offset + limit >= total}
                className="px-3 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed">
                {t('audit.next')}
              </button>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
