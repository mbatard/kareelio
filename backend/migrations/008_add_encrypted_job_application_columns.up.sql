ALTER TABLE job_applications
    ADD COLUMN IF NOT EXISTS company_enc TEXT,
    ADD COLUMN IF NOT EXISTS title_enc TEXT,
    ADD COLUMN IF NOT EXISTS salary_min_enc TEXT,
    ADD COLUMN IF NOT EXISTS salary_max_enc TEXT,
    ADD COLUMN IF NOT EXISTS salary_currency_enc TEXT,
    ADD COLUMN IF NOT EXISTS location_enc TEXT,
    ADD COLUMN IF NOT EXISTS benefits_enc TEXT,
    ADD COLUMN IF NOT EXISTS announcement_url_enc TEXT,
    ADD COLUMN IF NOT EXISTS test_notes_enc TEXT,
    ADD COLUMN IF NOT EXISTS offer_amount_enc TEXT,
    ADD COLUMN IF NOT EXISTS recruiter_contact_enc TEXT,
    ADD COLUMN IF NOT EXISTS notes_enc TEXT;
