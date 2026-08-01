ALTER TABLE job_applications
    DROP COLUMN IF EXISTS notes_enc,
    DROP COLUMN IF EXISTS recruiter_contact_enc,
    DROP COLUMN IF EXISTS offer_amount_enc,
    DROP COLUMN IF EXISTS test_notes_enc,
    DROP COLUMN IF EXISTS announcement_url_enc,
    DROP COLUMN IF EXISTS benefits_enc,
    DROP COLUMN IF EXISTS location_enc,
    DROP COLUMN IF EXISTS salary_currency_enc,
    DROP COLUMN IF EXISTS salary_max_enc,
    DROP COLUMN IF EXISTS salary_min_enc,
    DROP COLUMN IF EXISTS title_enc,
    DROP COLUMN IF EXISTS company_enc;
