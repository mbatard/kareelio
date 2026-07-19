CREATE TABLE IF NOT EXISTS job_applications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'applied', 'responded', 'interview', 'test', 'offer', 'rejected', 'withdrawn')),
    salary_min DECIMAL(12,2),
    salary_max DECIMAL(12,2),
    salary_currency VARCHAR(10) DEFAULT 'EUR',
    contract_type VARCHAR(20) DEFAULT 'other' CHECK (contract_type IN ('cdi', 'cdd', 'freelance', 'internship', 'apprentice', 'other')),
    location VARCHAR(255) DEFAULT '',
    remote VARCHAR(20) DEFAULT 'on_site' CHECK (remote IN ('on_site', 'hybrid', 'full_remote')),
    benefits TEXT DEFAULT '',
    announcement_url TEXT DEFAULT '',
    applied_at TIMESTAMPTZ,
    response_received BOOLEAN DEFAULT false,
    response_date TIMESTAMPTZ,
    first_contact_date TIMESTAMPTZ,
    first_contact_type VARCHAR(20) CHECK (first_contact_type IN ('video', 'phone', 'in_person')),
    has_test BOOLEAN DEFAULT false,
    test_date TIMESTAMPTZ,
    test_notes TEXT DEFAULT '',
    offer_received BOOLEAN DEFAULT false,
    offer_date TIMESTAMPTZ,
    offer_amount DECIMAL(12,2),
    priority VARCHAR(10) DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high')),
    source VARCHAR(20) DEFAULT 'other' CHECK (source IN ('linkedin', 'indeed', 'referral', 'agency', 'website', 'wttj', 'other')),
    recruiter_contact TEXT DEFAULT '',
    notes TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_job_applications_owner ON job_applications(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_job_applications_status ON job_applications(status);
CREATE INDEX IF NOT EXISTS idx_job_applications_company ON job_applications(company);
