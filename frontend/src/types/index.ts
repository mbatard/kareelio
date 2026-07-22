export interface User {
  id: string;
  email: string;
  display_name: string;
  description: string;
  role: 'admin' | 'user';
  is_active: boolean;
  email_verified_at: string | null;
  language: 'fr' | 'en' | 'system';
  theme: 'light' | 'dark' | 'system';
  created_at: string;
  updated_at: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface UpdateProfileRequest {
  email?: string;
  display_name?: string;
  description?: string;
  password?: string;
  language?: 'fr' | 'en' | 'system';
  theme?: 'light' | 'dark' | 'system';
}

export interface CreateUserRequest {
  email: string;
  password: string;
  display_name: string;
  description?: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  display_name?: string;
}

export interface UpdateUserRequest {
  email?: string;
  display_name?: string;
  description?: string;
  is_active?: boolean;
}

export type JobStatus = 'draft' | 'applied' | 'responded' | 'interview' | 'test' | 'offer' | 'rejected' | 'withdrawn';
export type RemoteType = 'on_site' | 'hybrid' | 'full_remote';
export type ContractType = 'cdi' | 'cdd' | 'freelance' | 'internship' | 'apprentice' | 'other';
export type ContactType = 'video' | 'phone' | 'in_person';
export type Source = 'linkedin' | 'indeed' | 'referral' | 'agency' | 'website' | 'wttj' | 'other';
export type Priority = 'low' | 'medium' | 'high';

export interface JobApplication {
  id: string;
  company: string;
  title: string;
  status: JobStatus;
  salary_min: number | null;
  salary_max: number | null;
  salary_currency: string;
  contract_type: ContractType;
  location: string;
  remote: RemoteType;
  benefits: string;
  announcement_url: string;
  applied_at: string | null;
  response_received: boolean;
  response_date: string | null;
  first_contact_date: string | null;
  first_contact_type: ContactType | null;
  has_test: boolean;
  test_date: string | null;
  test_notes: string;
  offer_received: boolean;
  offer_date: string | null;
  offer_amount: number | null;
  priority: Priority;
  source: Source;
  recruiter_contact: string;
  notes: string;
  created_at: string;
  updated_at: string;
}

export type CreateJobApplicationRequest = Omit<JobApplication, 'id' | 'created_at' | 'updated_at'>;
export type UpdateJobApplicationRequest = Partial<CreateJobApplicationRequest>;

export interface AboutInfo {
  version: string;
  name: string;
  description: string;
  go_version: string;
}

export interface AdminDashboard {
  users: { total: number; active: number; unverified: number; disabled: number };
  applications: { total: number; created_last_7_days: number; created_last_30_days: number; average_per_active_user: number };
  funnels: { response_rate: number; interview_rate: number; test_rate: number; offer_rate: number };
  by_status: Record<string, number>;
  by_source: Record<string, number>;
  by_remote: Record<string, number>;
  by_priority: Record<string, number>;
}

export interface AuditEvent {
  id: string;
  actor_user_id: string | null;
  actor_email: string;
  actor_role: string;
  actor_ip: string;
  action: string;
  target_type: string;
  target_id: string;
  metadata: Record<string, unknown> | null;
  created_at: string;
}

export interface AuditListResponse {
  events: AuditEvent[];
  total: number;
  limit: number;
  offset: number;
}
