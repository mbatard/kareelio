ALTER TABLE job_applications DROP CONSTRAINT IF EXISTS job_applications_source_check;
ALTER TABLE job_applications ADD CONSTRAINT job_applications_source_check CHECK (source IN ('linkedin', 'indeed', 'referral', 'agency', 'website', 'other'));
