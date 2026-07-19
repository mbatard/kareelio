import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { jobApplicationApi } from '../services/api';
import { STATUS_COLORS } from '../constants/statusColors';
import { KpiCard } from '../components/KpiCard';
import { MapCard } from '../components/MapCard';
import type { JobApplication } from '../types';

const INACTIVE_STATUSES = new Set(['rejected', 'withdrawn']);
const FUNNEL_STATUSES = ['applied', 'responded', 'interview', 'test', 'offer'];

export function DashboardPage() {
  const [applications, setApplications] = useState<JobApplication[]>([]);
  const [loading, setLoading] = useState(true);
  const { t } = useTranslation();

  useEffect(() => {
    jobApplicationApi.list().then(setApplications).finally(() => setLoading(false));
  }, []);

  if (loading) {
    return <div className="flex justify-center items-center h-64"><p className="text-gray-500">{t('common.loading')}</p></div>;
  }

  const total = applications.length;
  const active = applications.filter(a => !INACTIVE_STATUSES.has(a.status)).length;
  const drafts = applications.filter(a => a.status === 'draft').length;
  const responded = applications.filter(a => a.response_received).length;
  const interviewed = applications.filter(a => a.first_contact_date != null).length;
  const tested = applications.filter(a => a.has_test).length;
  const offered = applications.filter(a => a.offer_received).length;
  const rejected = applications.filter(a => a.status === 'rejected').length;

  const now = Date.now();
  const ms7d = 7 * 24 * 60 * 60 * 1000;
  const ms30d = 30 * 24 * 60 * 60 * 1000;
  const createdLast7 = applications.filter(a => now - new Date(a.created_at).getTime() < ms7d).length;
  const createdLast30 = applications.filter(a => now - new Date(a.created_at).getTime() < ms30d).length;

  const nonDraft = applications.filter(a => a.status !== 'draft').length;
  const responseRate = nonDraft > 0 ? (responded / nonDraft) * 100 : 0;
  const interviewRate = nonDraft > 0 ? (interviewed / nonDraft) * 100 : 0;
  const testRate = nonDraft > 0 ? (tested / nonDraft) * 100 : 0;
  const offerRate = nonDraft > 0 ? (offered / nonDraft) * 100 : 0;

  const statusCounts: Record<string, number> = {};
  applications.forEach(a => { statusCounts[a.status] = (statusCounts[a.status] || 0) + 1; });

  const sourceCounts: Record<string, number> = {};
  applications.forEach(a => { sourceCounts[a.source] = (sourceCounts[a.source] || 0) + 1; });

  const remoteCounts: Record<string, number> = {};
  applications.forEach(a => { remoteCounts[a.remote] = (remoteCounts[a.remote] || 0) + 1; });

  const priorityCounts: Record<string, number> = {};
  applications.forEach(a => { priorityCounts[a.priority] = (priorityCounts[a.priority] || 0) + 1; });

  const followUp = applications.filter(a =>
    !INACTIVE_STATUSES.has(a.status) &&
    a.status !== 'draft' &&
    now - new Date(a.updated_at).getTime() > 14 * 24 * 60 * 60 * 1000
  ).slice(0, 5);

  const recent = [...applications].sort((a, b) =>
    new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
  ).slice(0, 5);

  const funnel = FUNNEL_STATUSES.map(s => ({
    status: s,
    label: t(`status.${s}`),
    count: statusCounts[s] || 0,
    rate: s === 'applied'
      ? nonDraft > 0 ? ((statusCounts['applied'] || 0) / nonDraft) * 100 : 0
      : s === 'responded' ? responseRate
      : s === 'interview' ? interviewRate
      : s === 'test' ? testRate
      : offerRate,
  }));

  return (
    <div className="max-w-7xl mx-auto px-4 py-8 space-y-8">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('nav.dashboard')}</h1>
        <Link to="/applications/new"
          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-md text-sm font-medium transition-colors">
          {t('jobs.add')}
        </Link>
      </div>

      {total === 0 ? (
        <p className="text-gray-500 dark:text-gray-400 text-center py-8">{t('jobs.noApplications')}</p>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <KpiCard label={t('dashboard.totalApplications')} value={total} accent="text-gray-900 dark:text-white" />
            <KpiCard label={t('dashboard.activeApplications')} value={active} accent="text-blue-600 dark:text-blue-400" />
            <KpiCard label={t('dashboard.createdLast7Days')} value={createdLast7} accent="text-emerald-600 dark:text-emerald-400" />
            <KpiCard label={t('dashboard.createdLast30Days')} value={createdLast30} accent="text-amber-600 dark:text-amber-400" />
          </div>

          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
            <h2 className="text-sm font-semibold text-gray-500 dark:text-gray-400 uppercase mb-4">{t('dashboard.funnel')}</h2>
            <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
              {funnel.map((f) => (
                <div key={f.status}>
                  <div className="flex justify-between text-xs text-gray-500 dark:text-gray-400 mb-1">
                    <span>{f.label}</span>
                    <span className="font-medium text-gray-900 dark:text-white">{f.rate.toFixed(1)}%</span>
                  </div>
                  <div className="w-full h-2 rounded bg-gray-200 dark:bg-gray-700 overflow-hidden">
                    <div className="h-full rounded bg-blue-500 dark:bg-blue-400" style={{ width: `${f.rate}%` }} />
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <KpiCard label={t('dashboard.drafts')} value={drafts} accent="text-gray-600 dark:text-gray-400" />
            <KpiCard label={t('dashboard.responded')} value={responded} accent="text-cyan-600 dark:text-cyan-400" />
            <KpiCard label={t('dashboard.offers')} value={offered} accent="text-green-600 dark:text-green-400" />
            <KpiCard label={t('dashboard.rejected')} value={rejected} accent="text-red-600 dark:text-red-400" />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <MapCard title={t('dashboard.byStatus')} items={statusCounts} mapKey="status" />
            <MapCard title={t('dashboard.byPriority')} items={priorityCounts} mapKey="priority" />
            <MapCard title={t('dashboard.bySource')} items={sourceCounts} mapKey="source_type" />
            <MapCard title={t('dashboard.byRemote')} items={remoteCounts} mapKey="remote" />
          </div>

          {followUp.length > 0 && (
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
              <h2 className="text-sm font-semibold text-gray-500 dark:text-gray-400 uppercase px-6 py-4 border-b border-gray-200 dark:border-gray-700">
                {t('dashboard.followUpNeeded')}
              </h2>
              <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                <thead className="bg-gray-50 dark:bg-gray-700">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">{t('jobs.company')}</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">{t('jobs.jobTitle')}</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">{t('jobs.status')}</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">{t('dashboard.lastUpdate')}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                  {followUp.map((app) => (
                    <tr key={app.id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                      <td className="px-6 py-4 text-sm">
                        <Link to={`/applications/${app.id}/edit`} className="text-blue-600 hover:underline font-medium">{app.company}</Link>
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-900 dark:text-gray-300">{app.title}</td>
                      <td className="px-6 py-4">
                        <span className={`px-2 py-1 rounded text-xs font-medium ${STATUS_COLORS[app.status] || ''}`}>
                          {t(`status.${app.status}`)}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                        {new Date(app.updated_at).toLocaleDateString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          {recent.length > 0 && (
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
              <h2 className="text-sm font-semibold text-gray-500 dark:text-gray-400 uppercase px-6 py-4 border-b border-gray-200 dark:border-gray-700">
                {t('dashboard.recentApplications')}
              </h2>
              <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                <thead className="bg-gray-50 dark:bg-gray-700">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">{t('jobs.company')}</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">{t('jobs.jobTitle')}</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">{t('jobs.status')}</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">{t('jobs.priority')}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                  {recent.map((app) => (
                    <tr key={app.id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                      <td className="px-6 py-4 text-sm">
                        <Link to={`/applications/${app.id}/edit`} className="text-blue-600 hover:underline font-medium">{app.company}</Link>
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-900 dark:text-gray-300">{app.title}</td>
                      <td className="px-6 py-4">
                        <span className={`px-2 py-1 rounded text-xs font-medium ${STATUS_COLORS[app.status] || ''}`}>
                          {t(`status.${app.status}`)}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-600 dark:text-gray-400">{t(`priority.${app.priority}`)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </>
      )}
    </div>
  );
}

