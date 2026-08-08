package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/kareelio/backend/internal/encryption"
	"github.com/user/kareelio/backend/internal/model"
)

type JobApplicationRepository struct {
	db                    *pgxpool.Pool
	enc                   *encryption.Manager
	requireEncryptedReads bool
	legacyReadCount       int64
}

func NewJobApplicationRepository(db *pgxpool.Pool, enc *encryption.Manager, requireEncryptedReads bool) *JobApplicationRepository {
	return &JobApplicationRepository{db: db, enc: enc, requireEncryptedReads: requireEncryptedReads}
}

type jobApplicationRecord struct {
	model.JobApplication
	CompanyEnc          sql.NullString
	TitleEnc            sql.NullString
	SalaryMinEnc        sql.NullString
	SalaryMaxEnc        sql.NullString
	SalaryCurrencyEnc   sql.NullString
	LocationEnc         sql.NullString
	BenefitsEnc         sql.NullString
	AnnouncementURLEnc  sql.NullString
	TestNotesEnc        sql.NullString
	OfferAmountEnc      sql.NullString
	RecruiterContactEnc sql.NullString
	NotesEnc            sql.NullString
}

type jobApplicationWriter interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

const jobApplicationSelectColumns = `id, owner_user_id, company, title, status, salary_min, salary_max, salary_currency,
	contract_type, location, remote, benefits, announcement_url, applied_at,
	response_received, response_date, first_contact_date, first_contact_type,
	has_test, test_date, test_notes, offer_received, offer_date, offer_amount,
	priority, source, recruiter_contact, notes, created_at, updated_at,
	company_enc, title_enc, salary_min_enc, salary_max_enc, salary_currency_enc,
	location_enc, benefits_enc, announcement_url_enc, test_notes_enc, offer_amount_enc,
	recruiter_contact_enc, notes_enc`

const jobApplicationInsertColumns = `id, owner_user_id, company, title, status, salary_min, salary_max, salary_currency,
	contract_type, location, remote, benefits, announcement_url, applied_at,
	response_received, response_date, first_contact_date, first_contact_type,
	has_test, test_date, test_notes, offer_received, offer_date, offer_amount,
	priority, source, recruiter_contact, notes,
	company_enc, title_enc, salary_min_enc, salary_max_enc, salary_currency_enc,
	location_enc, benefits_enc, announcement_url_enc, test_notes_enc, offer_amount_enc,
	recruiter_contact_enc, notes_enc`

const jobApplicationReturningColumns = `id, owner_user_id, company, title, status, salary_min, salary_max, salary_currency,
	contract_type, location, remote, benefits, announcement_url, applied_at,
	response_received, response_date, first_contact_date, first_contact_type,
	has_test, test_date, test_notes, offer_received, offer_date, offer_amount,
	priority, source, recruiter_contact, notes, created_at, updated_at,
	company_enc, title_enc, salary_min_enc, salary_max_enc, salary_currency_enc,
	location_enc, benefits_enc, announcement_url_enc, test_notes_enc, offer_amount_enc,
	recruiter_contact_enc, notes_enc`

func scanJobApplicationRecord(scanner interface{ Scan(...any) error }) (*jobApplicationRecord, error) {
	var rec jobApplicationRecord
	if err := scanner.Scan(
		&rec.ID, &rec.OwnerUserID, &rec.Company, &rec.Title, &rec.Status, &rec.SalaryMin, &rec.SalaryMax, &rec.SalaryCurrency,
		&rec.ContractType, &rec.Location, &rec.Remote, &rec.Benefits, &rec.AnnouncementURL, &rec.AppliedAt,
		&rec.ResponseReceived, &rec.ResponseDate, &rec.FirstContactDate, &rec.FirstContactType,
		&rec.HasTest, &rec.TestDate, &rec.TestNotes, &rec.OfferReceived, &rec.OfferDate, &rec.OfferAmount,
		&rec.Priority, &rec.Source, &rec.RecruiterContact, &rec.Notes, &rec.CreatedAt, &rec.UpdatedAt,
		&rec.CompanyEnc, &rec.TitleEnc, &rec.SalaryMinEnc, &rec.SalaryMaxEnc, &rec.SalaryCurrencyEnc,
		&rec.LocationEnc, &rec.BenefitsEnc, &rec.AnnouncementURLEnc, &rec.TestNotesEnc, &rec.OfferAmountEnc,
		&rec.RecruiterContactEnc, &rec.NotesEnc,
	); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *JobApplicationRepository) materialize(rec *jobApplicationRecord) (*model.JobApplication, error) {
	if r.isLegacyRecord(rec) {
		if r.requireEncryptedReads {
			return nil, fmt.Errorf("legacy plaintext job application row %s requires backfill", rec.ID)
		}
		r.recordLegacyRead(rec.ID)
	}

	ja := rec.JobApplication

	company, err := r.resolveString(rec.Company, rec.CompanyEnc)
	if err != nil {
		return nil, fmt.Errorf("company: %w", err)
	}
	ja.Company = company

	title, err := r.resolveString(rec.Title, rec.TitleEnc)
	if err != nil {
		return nil, fmt.Errorf("title: %w", err)
	}
	ja.Title = title

	salaryCurrency, err := r.resolveString(rec.SalaryCurrency, rec.SalaryCurrencyEnc)
	if err != nil {
		return nil, fmt.Errorf("salary_currency: %w", err)
	}
	ja.SalaryCurrency = salaryCurrency

	location, err := r.resolveString(rec.Location, rec.LocationEnc)
	if err != nil {
		return nil, fmt.Errorf("location: %w", err)
	}
	ja.Location = location

	benefits, err := r.resolveString(rec.Benefits, rec.BenefitsEnc)
	if err != nil {
		return nil, fmt.Errorf("benefits: %w", err)
	}
	ja.Benefits = benefits

	announcementURL, err := r.resolveString(rec.AnnouncementURL, rec.AnnouncementURLEnc)
	if err != nil {
		return nil, fmt.Errorf("announcement_url: %w", err)
	}
	ja.AnnouncementURL = announcementURL

	testNotes, err := r.resolveString(rec.TestNotes, rec.TestNotesEnc)
	if err != nil {
		return nil, fmt.Errorf("test_notes: %w", err)
	}
	ja.TestNotes = testNotes

	offerAmount, err := r.resolveFloat(rec.OfferAmount, rec.OfferAmountEnc)
	if err != nil {
		return nil, fmt.Errorf("offer_amount: %w", err)
	}
	ja.OfferAmount = offerAmount

	recruiterContact, err := r.resolveString(rec.RecruiterContact, rec.RecruiterContactEnc)
	if err != nil {
		return nil, fmt.Errorf("recruiter_contact: %w", err)
	}
	ja.RecruiterContact = recruiterContact

	notes, err := r.resolveString(rec.Notes, rec.NotesEnc)
	if err != nil {
		return nil, fmt.Errorf("notes: %w", err)
	}
	ja.Notes = notes

	if salaryMin, err := r.resolveFloat(rec.SalaryMin, rec.SalaryMinEnc); err != nil {
		return nil, fmt.Errorf("salary_min: %w", err)
	} else {
		ja.SalaryMin = salaryMin
	}

	if salaryMax, err := r.resolveFloat(rec.SalaryMax, rec.SalaryMaxEnc); err != nil {
		return nil, fmt.Errorf("salary_max: %w", err)
	} else {
		ja.SalaryMax = salaryMax
	}

	return &ja, nil
}

func (r *JobApplicationRepository) isLegacyRecord(rec *jobApplicationRecord) bool {
	return !rec.CompanyEnc.Valid || !rec.TitleEnc.Valid || !rec.SalaryMinEnc.Valid || !rec.SalaryMaxEnc.Valid || !rec.SalaryCurrencyEnc.Valid || !rec.LocationEnc.Valid || !rec.BenefitsEnc.Valid || !rec.AnnouncementURLEnc.Valid || !rec.TestNotesEnc.Valid || !rec.OfferAmountEnc.Valid || !rec.RecruiterContactEnc.Valid || !rec.NotesEnc.Valid
}

func (r *JobApplicationRepository) recordLegacyRead(id string) {
	atomic.AddInt64(&r.legacyReadCount, 1)
	log.Printf("legacy plaintext job application read id=%s", id)
}

func (r *JobApplicationRepository) LegacyReadCount() int64 {
	return atomic.LoadInt64(&r.legacyReadCount)
}

func (r *JobApplicationRepository) resolveString(plain string, encrypted sql.NullString) (string, error) {
	if encrypted.Valid && encrypted.String != "" && r.enc != nil {
		return r.enc.Decrypt(encrypted.String)
	}
	return plain, nil
}

func (r *JobApplicationRepository) resolveFloat(plain *float64, encrypted sql.NullString) (*float64, error) {
	if encrypted.Valid && encrypted.String != "" && r.enc != nil {
		decoded, err := r.enc.Decrypt(encrypted.String)
		if err != nil {
			return nil, err
		}
		if decoded == "" {
			return nil, nil
		}
		value, err := strconv.ParseFloat(decoded, 64)
		if err != nil {
			return nil, fmt.Errorf("parse float: %w", err)
		}
		return &value, nil
	}
	return plain, nil
}

func (r *JobApplicationRepository) encryptString(value string) string {
	if r.enc == nil || value == "" {
		return ""
	}
	encrypted, err := r.enc.Encrypt(value)
	if err != nil {
		return ""
	}
	return encrypted
}

func (r *JobApplicationRepository) encryptFloat(value *float64) string {
	if value == nil {
		return ""
	}
	return r.encryptString(strconv.FormatFloat(*value, 'f', -1, 64))
}

func (r *JobApplicationRepository) createEncryptedValues(req model.CreateJobApplicationRequest) []any {
	return []any{
		r.encryptString(req.Company),
		r.encryptString(req.Title),
		r.encryptFloat(req.SalaryMin),
		r.encryptFloat(req.SalaryMax),
		r.encryptString(req.SalaryCurrency),
		r.encryptString(req.Location),
		r.encryptString(req.Benefits),
		r.encryptString(req.AnnouncementURL),
		r.encryptString(req.TestNotes),
		r.encryptFloat(req.OfferAmount),
		r.encryptString(req.RecruiterContact),
		r.encryptString(req.Notes),
	}
}

func (r *JobApplicationRepository) mergeUpdateRequest(existing *model.JobApplication, req model.UpdateJobApplicationRequest) model.CreateJobApplicationRequest {
	company := existing.Company
	title := existing.Title
	status := existing.Status
	salaryMin := existing.SalaryMin
	salaryMax := existing.SalaryMax
	salaryCurrency := existing.SalaryCurrency
	contractType := existing.ContractType
	location := existing.Location
	remote := existing.Remote
	benefits := existing.Benefits
	announcementURL := existing.AnnouncementURL
	appliedAt := existing.AppliedAt
	responseReceived := existing.ResponseReceived
	responseDate := existing.ResponseDate
	firstContactDate := existing.FirstContactDate
	firstContactType := existing.FirstContactType
	hasTest := existing.HasTest
	testDate := existing.TestDate
	testNotes := existing.TestNotes
	offerReceived := existing.OfferReceived
	offerDate := existing.OfferDate
	offerAmount := existing.OfferAmount
	priority := existing.Priority
	source := existing.Source
	recruiterContact := existing.RecruiterContact
	notes := existing.Notes

	if req.Company != nil {
		company = *req.Company
	}
	if req.Title != nil {
		title = *req.Title
	}
	if req.Status != nil {
		status = *req.Status
	}
	if req.SalaryMin != nil {
		salaryMin = req.SalaryMin
	}
	if req.SalaryMax != nil {
		salaryMax = req.SalaryMax
	}
	if req.SalaryCurrency != nil {
		salaryCurrency = *req.SalaryCurrency
	}
	if req.ContractType != nil {
		contractType = *req.ContractType
	}
	if req.Location != nil {
		location = *req.Location
	}
	if req.Remote != nil {
		remote = *req.Remote
	}
	if req.Benefits != nil {
		benefits = *req.Benefits
	}
	if req.AnnouncementURL != nil {
		announcementURL = *req.AnnouncementURL
	}
	if req.AppliedAt != nil {
		appliedAt = req.AppliedAt
	}
	if req.ResponseReceived != nil {
		responseReceived = *req.ResponseReceived
	}
	if req.ResponseDate != nil {
		responseDate = req.ResponseDate
	}
	if req.FirstContactDate != nil {
		firstContactDate = req.FirstContactDate
	}
	if req.FirstContactType != nil {
		firstContactType = req.FirstContactType
	}
	if req.HasTest != nil {
		hasTest = *req.HasTest
	}
	if req.TestDate != nil {
		testDate = req.TestDate
	}
	if req.TestNotes != nil {
		testNotes = *req.TestNotes
	}
	if req.OfferReceived != nil {
		offerReceived = *req.OfferReceived
	}
	if req.OfferDate != nil {
		offerDate = req.OfferDate
	}
	if req.OfferAmount != nil {
		offerAmount = req.OfferAmount
	}
	if req.Priority != nil {
		priority = *req.Priority
	}
	if req.Source != nil {
		source = *req.Source
	}
	if req.RecruiterContact != nil {
		recruiterContact = *req.RecruiterContact
	}
	if req.Notes != nil {
		notes = *req.Notes
	}

	return model.CreateJobApplicationRequest{
		Company:          company,
		Title:            title,
		Status:           status,
		SalaryMin:        salaryMin,
		SalaryMax:        salaryMax,
		SalaryCurrency:   salaryCurrency,
		ContractType:     contractType,
		Location:         location,
		Remote:           remote,
		Benefits:         benefits,
		AnnouncementURL:  announcementURL,
		AppliedAt:        appliedAt,
		ResponseReceived: responseReceived,
		ResponseDate:     responseDate,
		FirstContactDate: firstContactDate,
		FirstContactType: firstContactType,
		HasTest:          hasTest,
		TestDate:         testDate,
		TestNotes:        testNotes,
		OfferReceived:    offerReceived,
		OfferDate:        offerDate,
		OfferAmount:      offerAmount,
		Priority:         priority,
		Source:           source,
		RecruiterContact: recruiterContact,
		Notes:            notes,
	}
}

func (r *JobApplicationRepository) insertApplication(ctx context.Context, q jobApplicationWriter, id, userID string, req model.CreateJobApplicationRequest) (*model.JobApplication, error) {
	enc := r.createEncryptedValues(req)
	var rec jobApplicationRecord
	err := q.QueryRow(ctx,
		`INSERT INTO job_applications (`+jobApplicationInsertColumns+`) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
			$15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28,
			$29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40
		) RETURNING `+jobApplicationReturningColumns,
		append([]any{
			id, userID, req.Company, req.Title, req.Status, req.SalaryMin, req.SalaryMax, req.SalaryCurrency,
			req.ContractType, req.Location, req.Remote, req.Benefits, req.AnnouncementURL, req.AppliedAt,
			req.ResponseReceived, req.ResponseDate, req.FirstContactDate, req.FirstContactType,
			req.HasTest, req.TestDate, req.TestNotes, req.OfferReceived, req.OfferDate, req.OfferAmount,
			req.Priority, req.Source, req.RecruiterContact, req.Notes,
		}, enc...)...,
	).Scan(
		&rec.ID, &rec.OwnerUserID, &rec.Company, &rec.Title, &rec.Status, &rec.SalaryMin, &rec.SalaryMax, &rec.SalaryCurrency,
		&rec.ContractType, &rec.Location, &rec.Remote, &rec.Benefits, &rec.AnnouncementURL, &rec.AppliedAt,
		&rec.ResponseReceived, &rec.ResponseDate, &rec.FirstContactDate, &rec.FirstContactType,
		&rec.HasTest, &rec.TestDate, &rec.TestNotes, &rec.OfferReceived, &rec.OfferDate, &rec.OfferAmount,
		&rec.Priority, &rec.Source, &rec.RecruiterContact, &rec.Notes, &rec.CreatedAt, &rec.UpdatedAt,
		&rec.CompanyEnc, &rec.TitleEnc, &rec.SalaryMinEnc, &rec.SalaryMaxEnc, &rec.SalaryCurrencyEnc,
		&rec.LocationEnc, &rec.BenefitsEnc, &rec.AnnouncementURLEnc, &rec.TestNotesEnc, &rec.OfferAmountEnc,
		&rec.RecruiterContactEnc, &rec.NotesEnc,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create job application: %w", err)
	}
	return r.materialize(&rec)
}

func (r *JobApplicationRepository) updateApplication(ctx context.Context, q jobApplicationWriter, id, userID string, req model.CreateJobApplicationRequest) (*model.JobApplication, error) {
	enc := r.createEncryptedValues(req)
	var rec jobApplicationRecord
	err := q.QueryRow(ctx,
		`UPDATE job_applications SET
			company = $3, title = $4, status = $5, salary_min = $6, salary_max = $7, salary_currency = $8,
			contract_type = $9, location = $10, remote = $11, benefits = $12, announcement_url = $13,
			applied_at = $14, response_received = $15, response_date = $16, first_contact_date = $17,
			first_contact_type = $18, has_test = $19, test_date = $20, test_notes = $21,
			offer_received = $22, offer_date = $23, offer_amount = $24, priority = $25,
			source = $26, recruiter_contact = $27, notes = $28,
			company_enc = $29, title_enc = $30, salary_min_enc = $31, salary_max_enc = $32,
			salary_currency_enc = $33, location_enc = $34, benefits_enc = $35,
			announcement_url_enc = $36, test_notes_enc = $37, offer_amount_enc = $38,
			recruiter_contact_enc = $39, notes_enc = $40, updated_at = NOW()
		 WHERE id = $1 AND owner_user_id = $2
		 RETURNING `+jobApplicationReturningColumns,
		append([]any{
			id, userID, req.Company, req.Title, req.Status, req.SalaryMin, req.SalaryMax, req.SalaryCurrency,
			req.ContractType, req.Location, req.Remote, req.Benefits, req.AnnouncementURL, req.AppliedAt,
			req.ResponseReceived, req.ResponseDate, req.FirstContactDate, req.FirstContactType,
			req.HasTest, req.TestDate, req.TestNotes, req.OfferReceived, req.OfferDate, req.OfferAmount,
			req.Priority, req.Source, req.RecruiterContact, req.Notes,
		}, enc...)...,
	).Scan(
		&rec.ID, &rec.OwnerUserID, &rec.Company, &rec.Title, &rec.Status, &rec.SalaryMin, &rec.SalaryMax, &rec.SalaryCurrency,
		&rec.ContractType, &rec.Location, &rec.Remote, &rec.Benefits, &rec.AnnouncementURL, &rec.AppliedAt,
		&rec.ResponseReceived, &rec.ResponseDate, &rec.FirstContactDate, &rec.FirstContactType,
		&rec.HasTest, &rec.TestDate, &rec.TestNotes, &rec.OfferReceived, &rec.OfferDate, &rec.OfferAmount,
		&rec.Priority, &rec.Source, &rec.RecruiterContact, &rec.Notes, &rec.CreatedAt, &rec.UpdatedAt,
		&rec.CompanyEnc, &rec.TitleEnc, &rec.SalaryMinEnc, &rec.SalaryMaxEnc, &rec.SalaryCurrencyEnc,
		&rec.LocationEnc, &rec.BenefitsEnc, &rec.AnnouncementURLEnc, &rec.TestNotesEnc, &rec.OfferAmountEnc,
		&rec.RecruiterContactEnc, &rec.NotesEnc,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("job application not found or access denied")
		}
		return nil, fmt.Errorf("unable to update job application: %w", err)
	}
	return r.materialize(&rec)
}

func (r *JobApplicationRepository) Create(ctx context.Context, userID string, req model.CreateJobApplicationRequest) (*model.JobApplication, error) {
	return r.insertApplication(ctx, r.db, uuid.New().String(), userID, req)
}

func (r *JobApplicationRepository) GetByID(ctx context.Context, userID string, id string) (*model.JobApplication, error) {
	var rec jobApplicationRecord
	err := r.db.QueryRow(ctx,
		`SELECT `+jobApplicationSelectColumns+` FROM job_applications WHERE id = $1 AND owner_user_id = $2`,
		id, userID,
	).Scan(
		&rec.ID, &rec.OwnerUserID, &rec.Company, &rec.Title, &rec.Status, &rec.SalaryMin, &rec.SalaryMax, &rec.SalaryCurrency,
		&rec.ContractType, &rec.Location, &rec.Remote, &rec.Benefits, &rec.AnnouncementURL, &rec.AppliedAt,
		&rec.ResponseReceived, &rec.ResponseDate, &rec.FirstContactDate, &rec.FirstContactType,
		&rec.HasTest, &rec.TestDate, &rec.TestNotes, &rec.OfferReceived, &rec.OfferDate, &rec.OfferAmount,
		&rec.Priority, &rec.Source, &rec.RecruiterContact, &rec.Notes, &rec.CreatedAt, &rec.UpdatedAt,
		&rec.CompanyEnc, &rec.TitleEnc, &rec.SalaryMinEnc, &rec.SalaryMaxEnc, &rec.SalaryCurrencyEnc,
		&rec.LocationEnc, &rec.BenefitsEnc, &rec.AnnouncementURLEnc, &rec.TestNotesEnc, &rec.OfferAmountEnc,
		&rec.RecruiterContactEnc, &rec.NotesEnc,
	)
	if err != nil {
		return nil, fmt.Errorf("job application not found: %w", err)
	}
	return r.materialize(&rec)
}

func (r *JobApplicationRepository) List(ctx context.Context, userID string) ([]model.JobApplication, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+jobApplicationSelectColumns+` FROM job_applications WHERE owner_user_id = $1 ORDER BY updated_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to list job applications: %w", err)
	}
	defer rows.Close()

	var applications []model.JobApplication
	for rows.Next() {
		rec, err := scanJobApplicationRecord(rows)
		if err != nil {
			return nil, fmt.Errorf("unable to scan job application: %w", err)
		}
		ja, err := r.materialize(rec)
		if err != nil {
			return nil, err
		}
		applications = append(applications, *ja)
	}

	return applications, nil
}

func (r *JobApplicationRepository) Update(ctx context.Context, userID string, id string, req model.UpdateJobApplicationRequest) (*model.JobApplication, error) {
	existing, err := r.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return r.updateApplication(ctx, r.db, id, userID, r.mergeUpdateRequest(existing, req))
}

func (r *JobApplicationRepository) Delete(ctx context.Context, userID string, id string) error {
	tag, err := r.db.Exec(ctx,
		"DELETE FROM job_applications WHERE id = $1 AND owner_user_id = $2", id, userID)
	if err != nil {
		return fmt.Errorf("unable to delete job application: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("job application not found or access denied")
	}
	return nil
}

func (r *JobApplicationRepository) DeleteAll(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM job_applications WHERE owner_user_id = $1", userID)
	return err
}

func (r *JobApplicationRepository) ReplaceAll(ctx context.Context, userID string, apps []model.CreateJobApplicationRequest) ([]model.JobApplication, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "DELETE FROM job_applications WHERE owner_user_id = $1", userID); err != nil {
		return nil, fmt.Errorf("unable to delete existing applications: %w", err)
	}

	var created []model.JobApplication
	for _, req := range apps {
		ja, err := r.insertApplication(ctx, tx, uuid.New().String(), userID, req)
		if err != nil {
			return nil, err
		}
		created = append(created, *ja)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("unable to commit transaction: %w", err)
	}

	return created, nil
}

func (r *JobApplicationRepository) BulkCreate(ctx context.Context, userID string, apps []model.CreateJobApplicationRequest) ([]model.JobApplication, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var created []model.JobApplication
	for _, req := range apps {
		ja, err := r.insertApplication(ctx, tx, uuid.New().String(), userID, req)
		if err != nil {
			return nil, err
		}
		created = append(created, *ja)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("unable to commit transaction: %w", err)
	}

	return created, nil
}

type JobApplicationBackfillStats struct {
	Candidates int
	Updated    int
}

const jobApplicationBackfillWhereClause = `company_enc IS NULL OR title_enc IS NULL OR salary_min_enc IS NULL OR salary_max_enc IS NULL OR salary_currency_enc IS NULL OR location_enc IS NULL OR benefits_enc IS NULL OR announcement_url_enc IS NULL OR test_notes_enc IS NULL OR offer_amount_enc IS NULL OR recruiter_contact_enc IS NULL OR notes_enc IS NULL`

func (r *JobApplicationRepository) CountBackfillCandidates(ctx context.Context) (int, error) {
	var count int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM job_applications WHERE `+jobApplicationBackfillWhereClause,
	).Scan(&count); err != nil {
		return 0, fmt.Errorf("unable to count backfill candidates: %w", err)
	}
	return count, nil
}

func (r *JobApplicationRepository) BackfillEncryptedColumns(ctx context.Context, dryRun bool) (JobApplicationBackfillStats, error) {
	count, err := r.CountBackfillCandidates(ctx)
	if err != nil {
		return JobApplicationBackfillStats{}, err
	}

	stats := JobApplicationBackfillStats{Candidates: count}
	if dryRun || count == 0 {
		return stats, nil
	}

	rows, err := r.db.Query(ctx,
		`SELECT `+jobApplicationSelectColumns+` FROM job_applications WHERE `+jobApplicationBackfillWhereClause+` ORDER BY updated_at ASC`,
	)
	if err != nil {
		return stats, fmt.Errorf("unable to query backfill candidates: %w", err)
	}

	var candidates []*jobApplicationRecord
	for rows.Next() {
		rec, err := scanJobApplicationRecord(rows)
		if err != nil {
			rows.Close()
			return stats, fmt.Errorf("unable to scan backfill candidate: %w", err)
		}
		candidates = append(candidates, rec)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return stats, fmt.Errorf("unable to iterate backfill candidates: %w", err)
	}
	rows.Close()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return stats, fmt.Errorf("unable to begin backfill transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, rec := range candidates {

		app, err := r.materialize(rec)
		if err != nil {
			return stats, fmt.Errorf("unable to materialize backfill candidate: %w", err)
		}

		createReq := jobApplicationToCreateRequest(app)
		encValues := r.createEncryptedValues(createReq)

		result, err := tx.Exec(ctx,
			`UPDATE job_applications SET
				company_enc = COALESCE(company_enc, $3),
				title_enc = COALESCE(title_enc, $4),
				salary_min_enc = COALESCE(salary_min_enc, $5),
				salary_max_enc = COALESCE(salary_max_enc, $6),
				salary_currency_enc = COALESCE(salary_currency_enc, $7),
				location_enc = COALESCE(location_enc, $8),
				benefits_enc = COALESCE(benefits_enc, $9),
				announcement_url_enc = COALESCE(announcement_url_enc, $10),
				test_notes_enc = COALESCE(test_notes_enc, $11),
				offer_amount_enc = COALESCE(offer_amount_enc, $12),
				recruiter_contact_enc = COALESCE(recruiter_contact_enc, $13),
				notes_enc = COALESCE(notes_enc, $14),
				updated_at = NOW()
			 WHERE id = $1 AND owner_user_id = $2`,
			rec.ID, rec.OwnerUserID,
			encValues[0], encValues[1], encValues[2], encValues[3], encValues[4], encValues[5],
			encValues[6], encValues[7], encValues[8], encValues[9], encValues[10], encValues[11],
		)
		if err != nil {
			return stats, fmt.Errorf("unable to backfill job application: %w", err)
		}
		if result.RowsAffected() > 0 {
			stats.Updated++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return stats, fmt.Errorf("unable to commit backfill transaction: %w", err)
	}

	return stats, nil
}

func jobApplicationToCreateRequest(app *model.JobApplication) model.CreateJobApplicationRequest {
	return model.CreateJobApplicationRequest{
		Company:          app.Company,
		Title:            app.Title,
		Status:           app.Status,
		SalaryMin:        app.SalaryMin,
		SalaryMax:        app.SalaryMax,
		SalaryCurrency:   app.SalaryCurrency,
		ContractType:     app.ContractType,
		Location:         app.Location,
		Remote:           app.Remote,
		Benefits:         app.Benefits,
		AnnouncementURL:  app.AnnouncementURL,
		AppliedAt:        app.AppliedAt,
		ResponseReceived: app.ResponseReceived,
		ResponseDate:     app.ResponseDate,
		FirstContactDate: app.FirstContactDate,
		FirstContactType: app.FirstContactType,
		HasTest:          app.HasTest,
		TestDate:         app.TestDate,
		TestNotes:        app.TestNotes,
		OfferReceived:    app.OfferReceived,
		OfferDate:        app.OfferDate,
		OfferAmount:      app.OfferAmount,
		Priority:         app.Priority,
		Source:           app.Source,
		RecruiterContact: app.RecruiterContact,
		Notes:            app.Notes,
	}
}
