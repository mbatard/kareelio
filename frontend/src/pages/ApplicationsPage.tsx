import { useEffect, useState, useMemo, useRef } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { jobApplicationApi } from '../services/api';
import type { JobApplication, JobStatus } from '../types';

const STATUS_ORDER: JobStatus[] = ['draft', 'applied', 'responded', 'interview', 'test', 'offer', 'rejected', 'withdrawn'];
const REMOTE_ORDER = ['on_site', 'hybrid', 'full_remote'];
const PRIORITY_ORDER = ['low', 'medium', 'high'];

type SortKey = 'company' | 'title' | 'status' | 'location' | 'remote' | 'priority' | 'date';
type SortDir = 'asc' | 'desc';

function getLastDate(app: JobApplication): string | null {
  const dates = [app.applied_at, app.response_date, app.first_contact_date, app.test_date, app.offer_date].filter(Boolean);
  if (dates.length === 0) return null;
  return dates.sort((a, b) => new Date(b!).getTime() - new Date(a!).getTime())[0]!;
}

export function ApplicationsPage() {
  const [applications, setApplications] = useState<JobApplication[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<JobStatus | ''>('');
  const [search, setSearch] = useState('');
  const [sortKey, setSortKey] = useState<SortKey>('date');
  const [sortDir, setSortDir] = useState<SortDir>('desc');
  const [showImport, setShowImport] = useState(false);
  const [importFile, setImportFile] = useState<File | null>(null);
  const [importMode, setImportMode] = useState<'append' | 'replace'>('append');
  const [importing, setImporting] = useState(false);
  const [importResult, setImportResult] = useState('');
  const [importError, setImportError] = useState('');
  const fileInputRef = useRef<HTMLInputElement>(null);
  const { t } = useTranslation();

  useEffect(() => {
    jobApplicationApi.list().then(setApplications).finally(() => setLoading(false));
  }, []);

  const handleDelete = async (id: string) => {
    if (!confirm(t('jobs.confirmDelete'))) return;
    await jobApplicationApi.delete(id);
    setApplications(apps => apps.filter(a => a.id !== id));
  };

  const handleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortDir(d => d === 'asc' ? 'desc' : 'asc');
    } else {
      setSortKey(key);
      setSortDir(key === 'date' ? 'desc' : 'asc');
    }
  };

  const handleExport = async () => {
    await jobApplicationApi.exportCsv();
  };

  const handleImport = async () => {
    if (!importFile) return;
    setImporting(true);
    setImportResult('');
    setImportError('');
    try {
      const res = await jobApplicationApi.importCsv(importFile, importMode);
      setImportResult(t('jobs.importSuccess', { count: res.imported }));
      const apps = await jobApplicationApi.list();
      setApplications(apps);
      setTimeout(() => { setShowImport(false); setImportResult(''); }, 2000);
    } catch {
      setImportError(t('jobs.importError'));
    } finally {
      setImporting(false);
    }
  };

  const openImportModal = () => {
    setImportFile(null);
    setImportMode('append');
    setImportResult('');
    setImportError('');
    setShowImport(true);
  };

  const filtered = useMemo(() => {
    const result = applications.filter(app => {
      if (filter && app.status !== filter) return false;
      if (search) {
        const q = search.toLowerCase();
        return app.company.toLowerCase().includes(q) || app.title.toLowerCase().includes(q) || app.location.toLowerCase().includes(q);
      }
      return true;
    });

    result.sort((a, b) => {
      let cmp = 0;
      switch (sortKey) {
        case 'company': cmp = a.company.localeCompare(b.company); break;
        case 'title': cmp = a.title.localeCompare(b.title); break;
        case 'status': cmp = STATUS_ORDER.indexOf(a.status) - STATUS_ORDER.indexOf(b.status); break;
        case 'location': cmp = a.location.localeCompare(b.location); break;
        case 'remote': cmp = REMOTE_ORDER.indexOf(a.remote) - REMOTE_ORDER.indexOf(b.remote); break;
        case 'priority': cmp = PRIORITY_ORDER.indexOf(a.priority) - PRIORITY_ORDER.indexOf(b.priority); break;
        case 'date': {
          const da = getLastDate(a);
          const db = getLastDate(b);
          if (!da && !db) cmp = 0;
          else if (!da) cmp = 1;
          else if (!db) cmp = -1;
          else cmp = new Date(da).getTime() - new Date(db).getTime();
          break;
        }
      }
      return sortDir === 'asc' ? cmp : -cmp;
    });

    return result;
  }, [applications, filter, search, sortKey, sortDir]);

  const SortHeader = ({ label, sortField }: { label: string; sortField: SortKey }) => (
    <th onClick={() => handleSort(sortField)} className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase cursor-pointer select-none hover:text-gray-700 dark:hover:text-gray-100">
      <span className="inline-flex items-center gap-1">
        {label}
        {sortKey === sortField && (
          <span className="text-blue-600 dark:text-blue-400">{sortDir === 'asc' ? '↑' : '↓'}</span>
        )}
      </span>
    </th>
  );

  if (loading) return <div className="flex justify-center items-center h-64"><p>{t('common.loading')}</p></div>;

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('jobs.title')}</h1>
        <div className="flex space-x-2">
          <button onClick={handleExport}
            className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-md text-sm font-medium transition-colors">
            {t('jobs.export')}
          </button>
          <button onClick={openImportModal}
            className="px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded-md text-sm font-medium transition-colors">
            {t('jobs.import')}
          </button>
          <Link to="/applications/new"
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-md text-sm font-medium transition-colors">
            {t('jobs.add')}
          </Link>
        </div>
      </div>

      <div className="flex space-x-4 mb-6">
        <input value={search} onChange={(e) => setSearch(e.target.value)} placeholder={t('jobs.search')}
          className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" />
        <select value={filter} onChange={(e) => setFilter(e.target.value as JobStatus | '')}
          className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none">
          <option value="">-- {t('jobs.filterByStatus')} --</option>
          {['draft','applied','responded','interview','test','offer','rejected','withdrawn'].map(s => (
            <option key={s} value={s}>{t(`status.${s}`)}</option>
          ))}
        </select>
      </div>

      {filtered.length === 0 ? (
        <p className="text-gray-500 dark:text-gray-400 text-center py-8">{t('jobs.noApplications')}</p>
      ) : (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
            <thead className="bg-gray-50 dark:bg-gray-700">
              <tr>
                <SortHeader label={t('jobs.company')} sortField="company" />
                <SortHeader label={t('jobs.jobTitle')} sortField="title" />
                <SortHeader label={t('jobs.status')} sortField="status" />
                <SortHeader label={t('jobs.location')} sortField="location" />
                <SortHeader label={t('jobs.remote')} sortField="remote" />
                <SortHeader label={t('jobs.priority')} sortField="priority" />
                <SortHeader label={t('jobs.lastDate')} sortField="date" />
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">{t('jobs.announcement')}</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">{t('admin.actions')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
              {filtered.map((app) => {
                const lastDate = getLastDate(app);
                return (
                  <tr key={app.id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                    <td className="px-4 py-3">
                      <Link to={`/applications/${app.id}/edit`} className="text-blue-600 hover:underline font-medium">{app.company}</Link>
                    </td>
                    <td className="px-4 py-3 text-gray-900 dark:text-gray-300">{app.title}</td>
                    <td className="px-4 py-3">
                      <span className="text-xs font-medium px-2 py-1 rounded bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300">
                        {t(`status.${app.status}`)}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">{app.location}</td>
                    <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">{t(`remote.${app.remote}`)}</td>
                    <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">{t(`priority.${app.priority}`)}</td>
                    <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">
                      {lastDate ? new Date(lastDate).toLocaleDateString() : <span className="text-gray-400">-</span>}
                    </td>
                    <td className="px-4 py-3 text-center">
                      {app.announcement_url ? (
                        <a href={app.announcement_url} target="_blank" rel="noreferrer"
                          className="text-blue-600 hover:underline text-sm">{t('common.view')}</a>
                      ) : (
                        <span className="text-gray-400">-</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-center">
                      <button onClick={() => handleDelete(app.id)} className="text-sm text-red-600 hover:underline">{t('common.delete')}</button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {showImport && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 w-full max-w-md">
            <h2 className="text-lg font-bold text-gray-900 dark:text-white mb-4">{t('jobs.import')}</h2>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.importFile')}</label>
                <input ref={fileInputRef} type="file" accept=".csv"
                  onChange={(e) => setImportFile(e.target.files?.[0] ?? null)}
                  className="w-full text-sm text-gray-600 dark:text-gray-300 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-medium file:bg-purple-100 file:text-purple-700 dark:file:bg-purple-900 dark:file:text-purple-300 hover:file:bg-purple-200 dark:hover:file:bg-purple-800" />
              </div>

              <fieldset>
                <legend className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">{t('jobs.importMode')}</legend>
                <div className="space-y-2">
                  <label className="flex items-center space-x-2 cursor-pointer">
                    <input type="radio" name="importMode" value="append" checked={importMode === 'append'}
                      onChange={() => setImportMode('append')}
                      className="text-purple-600 focus:ring-purple-500" />
                    <span className="text-sm text-gray-700 dark:text-gray-300">{t('jobs.importAppend')}</span>
                  </label>
                  <label className="flex items-center space-x-2 cursor-pointer">
                    <input type="radio" name="importMode" value="replace" checked={importMode === 'replace'}
                      onChange={() => setImportMode('replace')}
                      className="text-purple-600 focus:ring-purple-500" />
                    <span className="text-sm text-gray-700 dark:text-gray-300">{t('jobs.importReplace')}</span>
                  </label>
                </div>
                {importMode === 'replace' && (
                  <p className="text-red-600 dark:text-red-400 text-xs mt-2">{t('jobs.importReplaceWarning')}</p>
                )}
              </fieldset>

              {importResult && <p className="text-green-600 dark:text-green-400 text-sm">{importResult}</p>}
              {importError && <p className="text-red-600 dark:text-red-400 text-sm">{importError}</p>}

              <div className="flex space-x-2 pt-2">
                <button onClick={handleImport} disabled={!importFile || importing}
                  className="px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded-md text-sm font-medium transition-colors disabled:opacity-50">
                  {importing ? t('common.loading') : t('jobs.importConfirm')}
                </button>
                <button onClick={() => setShowImport(false)}
                  className="px-4 py-2 bg-gray-300 dark:bg-gray-600 hover:bg-gray-400 dark:hover:bg-gray-500 text-gray-700 dark:text-gray-200 rounded-md text-sm font-medium transition-colors">
                  {t('common.cancel')}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
