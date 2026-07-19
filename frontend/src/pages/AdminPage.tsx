import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { adminApi } from '../services/api';
import { KpiCard } from '../components/KpiCard';
import { MapCard } from '../components/MapCard';
import type { AdminDashboard } from '../types';

export function AdminPage() {
  const [dash, setDash] = useState<AdminDashboard | null>(null);
  const [loading, setLoading] = useState(true);
  const { t } = useTranslation();

  useEffect(() => {
    adminApi.dashboard().then(setDash).finally(() => setLoading(false));
  }, []);

  if (loading) {
    return <div className="flex justify-center items-center h-64"><p className="text-gray-500">{t('common.loading')}</p></div>;
  }

  if (!dash) return null;

  const funnel = [
    { label: t('admin.responseRate'), value: dash.funnels.response_rate, color: 'bg-cyan-500' },
    { label: t('admin.interviewRate'), value: dash.funnels.interview_rate, color: 'bg-yellow-500' },
    { label: t('admin.testRate'),     value: dash.funnels.test_rate,     color: 'bg-orange-500' },
    { label: t('admin.offerRate'),    value: dash.funnels.offer_rate,    color: 'bg-green-500' },
  ];

  return (
    <div className="max-w-7xl mx-auto px-4 py-8 space-y-8">
      <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('admin.dashboard')}</h1>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <KpiCard label={t('admin.totalUsers')}  value={dash.users.total}  accent="text-blue-600 dark:text-blue-400" />
        <KpiCard label={t('admin.activeUsers')} value={dash.users.active} accent="text-green-600 dark:text-green-400" />
        <KpiCard label={t('admin.disabledUsers')} value={dash.users.disabled} accent="text-red-600 dark:text-red-400" />
        <KpiCard label={t('admin.avgPerUser')} value={dash.applications.average_per_active_user.toFixed(1)} accent="text-purple-600 dark:text-purple-400" />
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <KpiCard label={t('admin.totalApps')}   value={dash.applications.total}                accent="text-gray-900 dark:text-white" />
        <KpiCard label={t('admin.last7days')}  value={dash.applications.created_last_7_days}  accent="text-emerald-600 dark:text-emerald-400" />
        <KpiCard label={t('admin.last30days')} value={dash.applications.created_last_30_days} accent="text-amber-600 dark:text-amber-400" />
      </div>

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-sm font-semibold text-gray-500 dark:text-gray-400 uppercase mb-4">{t('admin.funnel')}</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {funnel.map((f) => (
            <div key={f.label}>
              <div className="flex justify-between text-xs text-gray-500 dark:text-gray-400 mb-1">
                <span>{f.label}</span>
                <span className="font-medium text-gray-900 dark:text-white">{f.value.toFixed(1)}%</span>
              </div>
              <div className="w-full h-2 rounded bg-gray-200 dark:bg-gray-700 overflow-hidden">
                <div className={`h-full rounded ${f.color}`} style={{ width: `${f.value}%` }} />
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <MapCard title={t('admin.byStatus')}   items={dash.by_status}   mapKey="status"   />
        <MapCard title={t('admin.bySource')}   items={dash.by_source}   mapKey="source_type" />
        <MapCard title={t('admin.byRemote')}   items={dash.by_remote}   mapKey="remote"   />
        <MapCard title={t('admin.byPriority')} items={dash.by_priority} mapKey="priority" />
      </div>
    </div>
  );
}

