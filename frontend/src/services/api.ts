import axios, { type InternalAxiosRequestConfig } from 'axios';
import type { User, LoginRequest, RegisterRequest, UpdateProfileRequest, CreateUserRequest, UpdateUserRequest, JobApplication, CreateJobApplicationRequest, UpdateJobApplicationRequest, AboutInfo, AdminDashboard, AuditListResponse } from '../types';

type LoggedRequestConfig = InternalAxiosRequestConfig & {
  metadata?: {
    action?: string;
    startedAt?: number;
  };
};

const frontendLogsEnabled = import.meta.env.DEV || import.meta.env.VITE_DEBUG_LOGS === 'true';

function logFrontendEvent(event: string, fields: string[] = []) {
  if (!frontendLogsEnabled) return;
  const parts = [`event=${event}`, ...fields];
  console.info(`[frontend] ${parts.join(' ')}`);
}

function getRequestId(headers: Record<string, unknown> | undefined) {
  const value = headers?.['x-request-id'] ?? headers?.['X-Request-ID'];
  return typeof value === 'string' && value.length > 0 ? value : undefined;
}

function resolveAction(method: string | undefined, path: string) {
  const upper = method?.toUpperCase() ?? 'GET';
  if (path === '/api/auth/login' && upper === 'POST') return 'auth.login';
  if (path === '/api/auth/register' && upper === 'POST') return 'auth.register';
  if (path === '/api/auth/verify-email' && upper === 'POST') return 'auth.verify_email';
  if (path === '/api/auth/resend-verification' && upper === 'POST') return 'auth.resend_verification';
  if (path === '/api/auth/logout' && upper === 'POST') return 'auth.logout';
  if (path === '/api/auth/me' && upper === 'GET') return 'auth.me';
  if (path === '/api/profile' && upper === 'GET') return 'profile.get';
  if (path === '/api/profile' && upper === 'PUT') return 'profile.update';
  if (path === '/api/profile/password' && upper === 'PUT') return 'profile.change_password';
  if (path === '/api/users' && upper === 'GET') return 'users.list';
  if (path === '/api/users' && upper === 'POST') return 'users.create';
  if (/^\/api\/users\/[^/]+$/.test(path) && upper === 'GET') return 'users.get';
  if (/^\/api\/users\/[^/]+$/.test(path) && upper === 'PUT') return 'users.update';
  if (/^\/api\/users\/[^/]+$/.test(path) && upper === 'DELETE') return 'users.delete';
  if (/^\/api\/users\/[^/]+\/password$/.test(path) && upper === 'PUT') return 'users.change_password';
  if (path === '/api/job-applications' && upper === 'GET') return 'job_applications.list';
  if (path === '/api/job-applications' && upper === 'POST') return 'job_applications.create';
  if (/^\/api\/job-applications\/[^/]+$/.test(path) && upper === 'GET') return 'job_applications.get';
  if (/^\/api\/job-applications\/[^/]+$/.test(path) && upper === 'PUT') return 'job_applications.update';
  if (/^\/api\/job-applications\/[^/]+$/.test(path) && upper === 'DELETE') return 'job_applications.delete';
  if (path === '/api/job-applications/export' && upper === 'GET') return 'job_applications.export';
  if (path === '/api/job-applications/import' && upper === 'POST') return 'job_applications.import';
  if (path === '/api/about' && upper === 'GET') return 'about.get';
  if (path === '/api/admin/dashboard' && upper === 'GET') return 'admin.dashboard';
  if (path === '/api/admin/audit' && upper === 'GET') return 'admin.audit';
  return 'api.unknown';
}

function resolvePath(config?: LoggedRequestConfig) {
  if (!config) return '';
  const baseURL = config.baseURL || window.location.origin;
  try {
    return new URL(config.url ?? '', baseURL).pathname;
  } catch {
    return config.url ?? '';
  }
}

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '',
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.response.use(
  (response) => {
    const config = response.config as LoggedRequestConfig;
    const action = config.metadata?.action ?? resolveAction(config.method, resolvePath(config));
    const requestId = getRequestId(response.headers as Record<string, unknown> | undefined);
    logFrontendEvent('api.request_success', [
      `action=${action}`,
      `status=${response.status}`,
      ...(requestId ? [`request_id=${requestId}`] : []),
    ]);
    return response;
  },
  (error) => {
    const config = error?.config as LoggedRequestConfig | undefined;
    const action = config?.metadata?.action ?? resolveAction(config?.method, resolvePath(config));
    const status = error?.response?.status ?? 'network_error';
    const requestId = getRequestId(error?.response?.headers as Record<string, unknown> | undefined);
    logFrontendEvent('api.request_error', [
      `action=${action}`,
      `status=${status}`,
      ...(requestId ? [`request_id=${requestId}`] : []),
    ]);
    return Promise.reject(error);
  }
);

api.interceptors.request.use((config) => {
  const loggedConfig = config as LoggedRequestConfig;
  const action = loggedConfig.metadata?.action ?? resolveAction(loggedConfig.method, resolvePath(loggedConfig));
  loggedConfig.metadata = {
    ...(loggedConfig.metadata ?? {}),
    action,
    startedAt: Date.now(),
  };
  logFrontendEvent('api.request_start', [`action=${action}`]);
  return loggedConfig;
});

export { logFrontendEvent, getRequestId };

export const authApi = {
  login: async (data: LoginRequest): Promise<{ user: User }> => {
    const res = await api.post('/api/auth/login', data);
    return res.data;
  },
  register: async (data: RegisterRequest): Promise<{ message: string }> => {
    const res = await api.post('/api/auth/register', data);
    return res.data;
  },
  verifyEmail: async (token: string): Promise<{ message: string }> => {
    const res = await api.post('/api/auth/verify-email', { token });
    return res.data;
  },
  resendVerification: async (email: string): Promise<{ message: string }> => {
    const res = await api.post('/api/auth/resend-verification', { email });
    return res.data;
  },
  logout: async (): Promise<void> => {
    await api.post('/api/auth/logout');
  },
  me: async (): Promise<User> => {
    const res = await api.get('/api/auth/me');
    return res.data;
  },
};

export const profileApi = {
  get: async (): Promise<User> => {
    const res = await api.get('/api/profile');
    return res.data;
  },
  update: async (data: UpdateProfileRequest): Promise<User> => {
    const res = await api.put('/api/profile', data);
    return res.data;
  },
  changePassword: async (currentPassword: string, newPassword: string): Promise<void> => {
    await api.put('/api/profile/password', {
      current_password: currentPassword,
      new_password: newPassword,
    });
  },
};

export const userApi = {
  list: async (): Promise<User[]> => {
    const res = await api.get('/api/users');
    return res.data;
  },
  get: async (id: string): Promise<User> => {
    const res = await api.get(`/api/users/${id}`);
    return res.data;
  },
  create: async (data: CreateUserRequest): Promise<User> => {
    const res = await api.post('/api/users', data);
    return res.data;
  },
  update: async (id: string, data: UpdateUserRequest): Promise<User> => {
    const res = await api.put(`/api/users/${id}`, data);
    return res.data;
  },
  delete: async (id: string): Promise<void> => {
    await api.delete(`/api/users/${id}`);
  },
  changePassword: async (id: string, newPassword: string): Promise<void> => {
    await api.put(`/api/users/${id}/password`, { new_password: newPassword });
  },
  resendVerification: async (id: string): Promise<{ message: string }> => {
    const res = await api.post(`/api/users/${id}/resend-verification`);
    return res.data;
  },
};

export const jobApplicationApi = {
  list: async (): Promise<JobApplication[]> => {
    const res = await api.get('/api/job-applications');
    return res.data;
  },
  get: async (id: string): Promise<JobApplication> => {
    const res = await api.get(`/api/job-applications/${id}`);
    return res.data;
  },
  create: async (data: CreateJobApplicationRequest): Promise<JobApplication> => {
    const res = await api.post('/api/job-applications', data);
    return res.data;
  },
  update: async (id: string, data: UpdateJobApplicationRequest): Promise<JobApplication> => {
    const res = await api.put(`/api/job-applications/${id}`, data);
    return res.data;
  },
  delete: async (id: string): Promise<void> => {
    await api.delete(`/api/job-applications/${id}`);
  },
  exportCsv: async (): Promise<void> => {
    const res = await api.get('/api/job-applications/export', { responseType: 'blob' });
    const url = window.URL.createObjectURL(new Blob([res.data]));
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', `kareelio_export_${new Date().toISOString().slice(0, 10)}.csv`);
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);
  },
  importCsv: async (file: File, mode: 'append' | 'replace'): Promise<{ imported: number }> => {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('mode', mode);
    const res = await api.post('/api/job-applications/import', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
    return res.data;
  },
};

export const aboutApi = {
  get: async (): Promise<AboutInfo> => {
    const res = await api.get('/api/about');
    return res.data;
  },
};

export const adminApi = {
  dashboard: async (): Promise<AdminDashboard> => {
    const res = await api.get('/api/admin/dashboard');
    return res.data;
  },
  audit: async (limit = 100, offset = 0): Promise<AuditListResponse> => {
    const res = await api.get('/api/admin/audit', { params: { limit, offset } });
    return res.data;
  },
};

export default api;
