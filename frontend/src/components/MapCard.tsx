import { useTranslation } from 'react-i18next';

export function MapCard({ title, items, mapKey }: { title: string; items: Record<string, number>; mapKey: string }) {
  const { t } = useTranslation();
  const total = Object.values(items).reduce((a, b) => a + b, 0);
  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
      <h2 className="text-sm font-semibold text-gray-500 dark:text-gray-400 uppercase mb-4">{title}</h2>
      <div className="space-y-2">
        {Object.entries(items).sort((a, b) => b[1] - a[1]).map(([key, count]) => {
          const pct = total > 0 ? (count / total) * 100 : 0;
          return (
            <div key={key}>
              <div className="flex justify-between text-xs text-gray-600 dark:text-gray-400 mb-1">
                <span>{t(`${mapKey}.${key}`, key)}</span>
                <span className="font-medium text-gray-900 dark:text-white">{count}</span>
              </div>
              <div className="w-full h-1.5 rounded bg-gray-200 dark:bg-gray-700 overflow-hidden">
                <div className="h-full rounded bg-blue-500 dark:bg-blue-400" style={{ width: `${pct}%` }} />
              </div>
            </div>
          );
        })}
        {Object.keys(items).length === 0 && (
          <p className="text-xs text-gray-400">{t('common.noResults')}</p>
        )}
      </div>
    </div>
  );
}
