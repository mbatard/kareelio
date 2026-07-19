import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { jobApplicationApi } from '../services/api';
import type { CreateJobApplicationRequest, JobStatus, RemoteType, ContractType, Source, Priority } from '../types';

function toISOStringDate(val: string | null): string | null {
  if (!val) return null;
  if (val.includes('T')) return val;
  return val + 'T00:00:00Z';
}

export function JobApplicationForm() {
  const { id } = useParams();
  const isEdit = Boolean(id);
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(isEdit);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  const [form, setForm] = useState<CreateJobApplicationRequest>({
    company: '',
    title: '',
    status: 'draft' as JobStatus,
    salary_min: null,
    salary_max: null,
    salary_currency: 'EUR',
    contract_type: 'other' as ContractType,
    location: '',
    remote: 'on_site' as RemoteType,
    benefits: '',
    announcement_url: '',
    applied_at: null,
    response_received: false,
    response_date: null,
    first_contact_date: null,
    first_contact_type: null,
    has_test: false,
    test_date: null,
    test_notes: '',
    offer_received: false,
    offer_date: null,
    offer_amount: null,
    priority: 'medium' as Priority,
    source: 'other' as Source,
    recruiter_contact: '',
    notes: '',
  });

  useEffect(() => {
    if (isEdit && id) {
      jobApplicationApi.get(id).then((app) => {
        setForm(app);
        setLoading(false);
      }).catch(() => {
        setError(t('common.error'));
        setLoading(false);
      });
    }
  }, [id, isEdit, t]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSaving(true);
    try {
      const payload = {
        ...form,
        applied_at: toISOStringDate(form.applied_at),
        response_date: toISOStringDate(form.response_date),
        first_contact_date: toISOStringDate(form.first_contact_date),
        test_date: toISOStringDate(form.test_date),
        offer_date: toISOStringDate(form.offer_date),
      };
      if (isEdit && id) {
        await jobApplicationApi.update(id, payload);
      } else {
        await jobApplicationApi.create(payload);
      }
      navigate('/applications');
    } catch (err: any) {
      const msg = err?.response?.data?.error || t('common.error');
      setError(msg);
    } finally {
      setSaving(false);
    }
  };

  const set = (field: keyof CreateJobApplicationRequest, value: any) => {
    setForm((prev) => ({ ...prev, [field]: value }));
  };

  if (loading) return <div className="flex justify-center items-center h-64"><p>{t('common.loading')}</p></div>;

  return (
    <div className="max-w-3xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold mb-6 text-gray-900 dark:text-white">
        {isEdit ? t('jobs.edit') : t('jobs.add')}
      </h1>
      {error && (
        <div className="mb-4 p-3 bg-red-100 dark:bg-red-900/30 border border-red-300 dark:border-red-700 rounded-md text-red-700 dark:text-red-400 text-sm">
          {error}
        </div>
      )}
      <form onSubmit={handleSubmit} className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 space-y-4">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.company')} *</label>
            <input required value={form.company} onChange={(e) => set('company', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.jobTitle')} *</label>
            <input required value={form.title} onChange={(e) => set('title', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" />
          </div>
        </div>

        <div className="grid grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.status')}</label>
            <select value={form.status} onChange={(e) => set('status', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none">
              {['draft','applied','responded','interview','test','offer','rejected','withdrawn'].map(s => (
                <option key={s} value={s}>{t(`status.${s}`)}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.priority')}</label>
            <select value={form.priority} onChange={(e) => set('priority', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none">
              {['low','medium','high'].map(p => (
                <option key={p} value={p}>{t(`priority.${p}`)}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.source')}</label>
            <select value={form.source} onChange={(e) => set('source', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none">
              {['linkedin','indeed','referral','agency','website','wttj','other'].map(s => (
                <option key={s} value={s}>{t(`source_type.${s}`)}</option>
              ))}
            </select>
          </div>
        </div>

        <div className="grid grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.contractType')}</label>
            <select value={form.contract_type} onChange={(e) => set('contract_type', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none">
              {['cdi','cdd','freelance','internship','apprentice','other'].map(c => (
                <option key={c} value={c}>{t(`contract.${c}`)}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.remote')}</label>
            <select value={form.remote} onChange={(e) => set('remote', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none">
              {['on_site','hybrid','full_remote'].map(r => (
                <option key={r} value={r}>{t(`remote.${r}`)}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.location')}</label>
            <input value={form.location} onChange={(e) => set('location', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" />
          </div>
        </div>

        <div className="grid grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.salaryMin')}</label>
            <input type="number" value={form.salary_min ?? ''} onChange={(e) => set('salary_min', e.target.value ? Number(e.target.value) : null)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.salaryMax')}</label>
            <input type="number" value={form.salary_max ?? ''} onChange={(e) => set('salary_max', e.target.value ? Number(e.target.value) : null)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.currency')}</label>
            <select value={form.salary_currency} onChange={(e) => set('salary_currency', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none">
              <option value="EUR">EUR</option>
              <option value="USD">USD</option>
              <option value="GBP">GBP</option>
              <option value="CHF">CHF</option>
            </select>
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.announcementUrl')}</label>
          <input type="url" value={form.announcement_url} onChange={(e) => set('announcement_url', e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.appliedAt')}</label>
          <input type="date" value={form.applied_at ? form.applied_at.split('T')[0] : ''} onChange={(e) => set('applied_at', e.target.value || null)}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" />
        </div>

        <div className="grid grid-cols-3 gap-4">
          <label className="flex items-center space-x-2">
            <input type="checkbox" checked={form.response_received} onChange={(e) => set('response_received', e.target.checked)} className="rounded" />
            <span className="text-sm text-gray-700 dark:text-gray-300">{t('jobs.responseReceived')}</span>
          </label>
          <label className="flex items-center space-x-2">
            <input type="checkbox" checked={form.has_test} onChange={(e) => set('has_test', e.target.checked)} className="rounded" />
            <span className="text-sm text-gray-700 dark:text-gray-300">{t('jobs.hasTest')}</span>
          </label>
          <label className="flex items-center space-x-2">
            <input type="checkbox" checked={form.offer_received} onChange={(e) => set('offer_received', e.target.checked)} className="rounded" />
            <span className="text-sm text-gray-700 dark:text-gray-300">{t('jobs.offerReceived')}</span>
          </label>
        </div>

        {form.response_received && (
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.responseDate')}</label>
            <input type="date" value={form.response_date ? form.response_date.split('T')[0] : ''} onChange={(e) => set('response_date', e.target.value || null)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" />
          </div>
        )}

        {form.has_test && (
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.testDate')}</label>
              <input type="date" value={form.test_date ? form.test_date.split('T')[0] : ''} onChange={(e) => set('test_date', e.target.value || null)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.testNotes')}</label>
              <input value={form.test_notes} onChange={(e) => set('test_notes', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" />
            </div>
          </div>
        )}

        {form.offer_received && (
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.offerDate')}</label>
              <input type="date" value={form.offer_date ? form.offer_date.split('T')[0] : ''} onChange={(e) => set('offer_date', e.target.value || null)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.offerAmount')}</label>
              <input type="number" value={form.offer_amount ?? ''} onChange={(e) => set('offer_amount', e.target.value ? Number(e.target.value) : null)}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" />
            </div>
          </div>
        )}

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.firstContactDate')}</label>
            <input type="date" value={form.first_contact_date ? form.first_contact_date.split('T')[0] : ''} onChange={(e) => set('first_contact_date', e.target.value || null)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.firstContactType')}</label>
            <select value={form.first_contact_type ?? ''} onChange={(e) => set('first_contact_type', e.target.value || null)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none">
              <option value="">-</option>
              {['video','phone','in_person'].map(c => (
                <option key={c} value={c}>{t(`contact.${c}`)}</option>
              ))}
            </select>
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.recruiterContact')}</label>
          <input value={form.recruiter_contact} onChange={(e) => set('recruiter_contact', e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.benefits')}</label>
          <textarea value={form.benefits} onChange={(e) => set('benefits', e.target.value)} rows={2}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('jobs.notes')}</label>
          <textarea value={form.notes} onChange={(e) => set('notes', e.target.value)} rows={3}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none" />
        </div>

        {isEdit && (form as any).created_at && (
          <div className="text-xs text-gray-400 dark:text-gray-500 flex gap-4">
            <span>{t('jobs.createdAt')}: {new Date((form as any).created_at).toLocaleDateString()}</span>
            <span>{t('jobs.updatedAt')}: {new Date((form as any).updated_at).toLocaleDateString()}</span>
          </div>
        )}

        <div className="flex space-x-4">
          <button type="submit" disabled={saving}
            className="px-6 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-md font-medium transition-colors disabled:opacity-50">
            {saving ? t('common.loading') : t('jobs.save')}
          </button>
          <button type="button" onClick={() => navigate('/applications')}
            className="px-6 py-2 bg-gray-300 dark:bg-gray-600 hover:bg-gray-400 dark:hover:bg-gray-500 text-gray-700 dark:text-gray-200 rounded-md font-medium transition-colors">
            {t('jobs.cancel')}
          </button>
        </div>
      </form>
    </div>
  );
}
