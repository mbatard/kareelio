import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { aboutApi } from '../services/api';
import type { AboutInfo } from '../types';

export function AboutPage() {
  const [info, setInfo] = useState<AboutInfo | null>(null);
  const { t } = useTranslation();

  useEffect(() => {
    aboutApi.get().then(setInfo);
  }, []);

  return (
    <div className="max-w-2xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold mb-6 text-gray-900 dark:text-white">{t('about.title')}</h1>
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 space-y-4">
        {info && (
          <>
            <div>
              <h2 className="text-lg font-semibold text-gray-900 dark:text-white">{info.name}</h2>
              <p className="text-gray-600 dark:text-gray-400">{info.description}</p>
            </div>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="font-medium text-gray-700 dark:text-gray-300">{t('about.version')}:</span>
                <span className="ml-2 text-gray-600 dark:text-gray-400">{info.version}</span>
              </div>
              <div>
                <span className="font-medium text-gray-700 dark:text-gray-300">Go:</span>
                <span className="ml-2 text-gray-600 dark:text-gray-400">{info.go_version}</span>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
