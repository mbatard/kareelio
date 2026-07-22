import axios from 'axios';
import type { User, LoginRequest, RegisterRequest, UpdateProfileRequest, CreateUserRequest, UpdateUserRequest, JobApplication, CreateJobApplicationRequest, UpdateJobApplicationRequest, AboutInfo, AdminDashboard, AuditListResponse } from '../types';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '',
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    return Promise.reject(error);
  }
);

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
