import { JobStatus } from '../types';

export const STATUS_COLORS: Record<JobStatus, string> = {
  draft:     'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300',
  applied:   'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300',
  responded: 'bg-cyan-100 text-cyan-700 dark:bg-cyan-900 dark:text-cyan-300',
  interview: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300',
  test:      'bg-orange-100 text-orange-700 dark:bg-orange-900 dark:text-orange-300',
  offer:     'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300',
  rejected:  'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300',
  withdrawn: 'bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400',
};
