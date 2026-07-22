-- Seed script for marketing screenshots
-- Run: docker exec -i kareelio-postgres-1 psql -U kareelio -d kareelio < marketing/seed-screenshots.sql

-- Clean existing demo data
DELETE FROM job_applications WHERE owner_user_id = (SELECT id FROM users WHERE email = 'jean.dupont@example.com');
DELETE FROM users WHERE email = 'jean.dupont@example.com';

-- Create demo user (password: demo1234)
INSERT INTO users (email, display_name, description, password_hash, role, is_active, language, theme)
VALUES ('jean.dupont@example.com', 'Jean Dupont', 'Développeur Full Stack passionné', '$2b$10$dXvw1sN7pL8ycDt5RJDkZOGYABgDxwGtCuDmLDLin0X0Ku5FVyz4C', 'user', true, 'fr', 'light');

-- Get user ID and insert applications
DO $$
DECLARE
  uid UUID;
BEGIN
  SELECT id INTO uid FROM users WHERE email = 'jean.dupont@example.com';

  INSERT INTO job_applications (owner_user_id, company, title, status, salary_min, salary_max, salary_currency, contract_type, location, remote, applied_at, response_received, response_date, has_test, test_date, test_notes, offer_received, priority, source, recruiter_contact, notes)
  VALUES
    (uid, 'TechCorp Paris', 'Développeur Full Stack Senior', 'interview', 55000, 68000, 'EUR', 'cdi', 'Paris 10e', 'hybrid', '2026-06-15', true, '2026-06-20', true, '2026-07-05', 'Test technique React + Node.js', false, 'high', 'linkedin', 'Marie Lemaire', 'Startup en forte croissance'),
    (uid, 'InnovateTech', 'Lead Developer', 'applied', 60000, 75000, 'EUR', 'cdi', 'Lyon', 'full_remote', '2026-07-01', false, NULL, false, NULL, '', false, 'medium', 'wttj', 'Thomas Bernard', 'Stack technique moderne'),
    (uid, 'DataFlow SAS', 'Ingénieur Backend', 'responded', 48000, 58000, 'EUR', 'cdi', 'Marseille', 'on_site', '2026-06-28', true, '2026-07-10', false, NULL, '', false, 'medium', 'indeed', 'Sophie Martin', 'Spécialiste data processing'),
    (uid, 'CloudNine', 'DevOps Engineer', 'offer', 52000, 65000, 'EUR', 'cdi', 'Bordeaux', 'hybrid', '2026-05-20', true, '2026-05-28', true, '2026-06-10', 'Entretien technique + culture fit', true, 'high', 'referral', 'Paul Dubois', 'Offre reçue, en réflexion'),
    (uid, 'GreenTech Solutions', 'Frontend Developer', 'test', 42000, 52000, 'EUR', 'cdi', 'Nantes', 'full_remote', '2026-07-10', true, '2026-07-15', true, '2026-07-22', 'Challenge technique React Native', false, 'medium', 'wttj', 'Claire Petit', 'Projet mobile ambitieux'),
    (uid, 'FinPlus', 'Développeur Full Stack', 'rejected', 50000, 62000, 'EUR', 'cdi', 'Paris 8e', 'on_site', '2026-06-01', true, '2026-06-10', false, NULL, '', false, 'low', 'website', 'Jean Moreau', 'Pas retenu'),
    (uid, 'StartupLab', 'CTO / Co-fondateur', 'draft', NULL, NULL, 'EUR', 'freelance', 'Remote', 'full_remote', NULL, false, NULL, false, NULL, '', false, 'low', 'other', '', 'Projet personnel'),
    (uid, 'EduTech France', 'Développeur Backend Python', 'interview', 45000, 55000, 'EUR', 'cdi', 'Toulouse', 'hybrid', '2026-07-05', true, '2026-07-12', true, '2026-07-18', 'Entretien technique Python/Django', false, 'high', 'linkedin', 'Antoine Leroy', 'Secteur éducation'),
    (uid, 'SecureNet', 'Ingénieur Sécurité', 'applied', 55000, 70000, 'EUR', 'cdi', 'Paris 2e', 'on_site', '2026-07-18', false, NULL, false, NULL, '', false, 'medium', 'agency', 'Nathalie Roux', 'Cabinet de conseil'),
    (uid, 'MobiApp', 'React Native Developer', 'withdrawn', 40000, 48000, 'EUR', 'cdd', 'Lyon', 'full_remote', '2026-06-20', true, '2026-06-25', true, '2026-07-02', 'Mission trop courte', false, 'low', 'linkedin', 'Marc Fernandez', 'Retiré, CDD 6 mois');

  RAISE NOTICE 'Seeded % applications for user %', 10, uid;
END $$;
