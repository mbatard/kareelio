package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/kareelio/backend/internal/model"
)

type JobApplicationRepository struct {
	db *pgxpool.Pool
}

func NewJobApplicationRepository(db *pgxpool.Pool) *JobApplicationRepository {
	return &JobApplicationRepository{db: db}
}

func (r *JobApplicationRepository) Create(ctx context.Context, userID string, req model.CreateJobApplicationRequest) (*model.JobApplication, error) {
	id := uuid.New().String()

	var ja model.JobApplication
	err := r.db.QueryRow(ctx,
		`INSERT INTO job_applications (
			id, owner_user_id, company, title, status, salary_min, salary_max, salary_currency,
			contract_type, location, remote, benefits, announcement_url, applied_at,
			response_received, response_date, first_contact_date, first_contact_type,
			has_test, test_date, test_notes, offer_received, offer_date, offer_amount,
			priority, source, recruiter_contact, notes
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
			$15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28
		) RETURNING id, owner_user_id, company, title, status, salary_min, salary_max, salary_currency,
			contract_type, location, remote, benefits, announcement_url, applied_at,
			response_received, response_date, first_contact_date, first_contact_type,
			has_test, test_date, test_notes, offer_received, offer_date, offer_amount,
			priority, source, recruiter_contact, notes, created_at, updated_at`,
		id, userID, req.Company, req.Title, req.Status, req.SalaryMin, req.SalaryMax, req.SalaryCurrency,
		req.ContractType, req.Location, req.Remote, req.Benefits, req.AnnouncementURL, req.AppliedAt,
		req.ResponseReceived, req.ResponseDate, req.FirstContactDate, req.FirstContactType,
		req.HasTest, req.TestDate, req.TestNotes, req.OfferReceived, req.OfferDate, req.OfferAmount,
		req.Priority, req.Source, req.RecruiterContact, req.Notes,
	).Scan(&ja.ID, &ja.OwnerUserID, &ja.Company, &ja.Title, &ja.Status, &ja.SalaryMin, &ja.SalaryMax, &ja.SalaryCurrency,
		&ja.ContractType, &ja.Location, &ja.Remote, &ja.Benefits, &ja.AnnouncementURL, &ja.AppliedAt,
		&ja.ResponseReceived, &ja.ResponseDate, &ja.FirstContactDate, &ja.FirstContactType,
		&ja.HasTest, &ja.TestDate, &ja.TestNotes, &ja.OfferReceived, &ja.OfferDate, &ja.OfferAmount,
		&ja.Priority, &ja.Source, &ja.RecruiterContact, &ja.Notes, &ja.CreatedAt, &ja.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to create job application: %w", err)
	}

	return &ja, nil
}

func (r *JobApplicationRepository) GetByID(ctx context.Context, userID string, id string) (*model.JobApplication, error) {
	var ja model.JobApplication
	err := r.db.QueryRow(ctx,
		`SELECT id, owner_user_id, company, title, status, salary_min, salary_max, salary_currency,
			contract_type, location, remote, benefits, announcement_url, applied_at,
			response_received, response_date, first_contact_date, first_contact_type,
			has_test, test_date, test_notes, offer_received, offer_date, offer_amount,
			priority, source, recruiter_contact, notes, created_at, updated_at
		 FROM job_applications WHERE id = $1 AND owner_user_id = $2`,
		id, userID,
	).Scan(&ja.ID, &ja.OwnerUserID, &ja.Company, &ja.Title, &ja.Status, &ja.SalaryMin, &ja.SalaryMax, &ja.SalaryCurrency,
		&ja.ContractType, &ja.Location, &ja.Remote, &ja.Benefits, &ja.AnnouncementURL, &ja.AppliedAt,
		&ja.ResponseReceived, &ja.ResponseDate, &ja.FirstContactDate, &ja.FirstContactType,
		&ja.HasTest, &ja.TestDate, &ja.TestNotes, &ja.OfferReceived, &ja.OfferDate, &ja.OfferAmount,
		&ja.Priority, &ja.Source, &ja.RecruiterContact, &ja.Notes, &ja.CreatedAt, &ja.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("job application not found: %w", err)
	}

	return &ja, nil
}

func (r *JobApplicationRepository) List(ctx context.Context, userID string) ([]model.JobApplication, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, owner_user_id, company, title, status, salary_min, salary_max, salary_currency,
			contract_type, location, remote, benefits, announcement_url, applied_at,
			response_received, response_date, first_contact_date, first_contact_type,
			has_test, test_date, test_notes, offer_received, offer_date, offer_amount,
			priority, source, recruiter_contact, notes, created_at, updated_at
		 FROM job_applications WHERE owner_user_id = $1 ORDER BY updated_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to list job applications: %w", err)
	}
	defer rows.Close()

	var applications []model.JobApplication
	for rows.Next() {
		var ja model.JobApplication
		if err := rows.Scan(&ja.ID, &ja.OwnerUserID, &ja.Company, &ja.Title, &ja.Status, &ja.SalaryMin, &ja.SalaryMax, &ja.SalaryCurrency,
			&ja.ContractType, &ja.Location, &ja.Remote, &ja.Benefits, &ja.AnnouncementURL, &ja.AppliedAt,
			&ja.ResponseReceived, &ja.ResponseDate, &ja.FirstContactDate, &ja.FirstContactType,
			&ja.HasTest, &ja.TestDate, &ja.TestNotes, &ja.OfferReceived, &ja.OfferDate, &ja.OfferAmount,
			&ja.Priority, &ja.Source, &ja.RecruiterContact, &ja.Notes, &ja.CreatedAt, &ja.UpdatedAt); err != nil {
			return nil, fmt.Errorf("unable to scan job application: %w", err)
		}
		applications = append(applications, ja)
	}

	return applications, nil
}

func (r *JobApplicationRepository) Update(ctx context.Context, userID string, id string, req model.UpdateJobApplicationRequest) (*model.JobApplication, error) {
	existing, err := r.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

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

	var ja model.JobApplication
	err = r.db.QueryRow(ctx,
		`UPDATE job_applications SET
			company = $3, title = $4, status = $5, salary_min = $6, salary_max = $7, salary_currency = $8,
			contract_type = $9, location = $10, remote = $11, benefits = $12, announcement_url = $13,
			applied_at = $14, response_received = $15, response_date = $16, first_contact_date = $17,
			first_contact_type = $18, has_test = $19, test_date = $20, test_notes = $21,
			offer_received = $22, offer_date = $23, offer_amount = $24, priority = $25,
			source = $26, recruiter_contact = $27, notes = $28, updated_at = NOW()
		 WHERE id = $1 AND owner_user_id = $2
		 RETURNING id, owner_user_id, company, title, status, salary_min, salary_max, salary_currency,
			contract_type, location, remote, benefits, announcement_url, applied_at,
			response_received, response_date, first_contact_date, first_contact_type,
			has_test, test_date, test_notes, offer_received, offer_date, offer_amount,
			priority, source, recruiter_contact, notes, created_at, updated_at`,
		id, userID, company, title, status, salaryMin, salaryMax, salaryCurrency,
		contractType, location, remote, benefits, announcementURL, appliedAt,
		responseReceived, responseDate, firstContactDate, firstContactType,
		hasTest, testDate, testNotes, offerReceived, offerDate, offerAmount,
		priority, source, recruiterContact, notes,
	).Scan(&ja.ID, &ja.OwnerUserID, &ja.Company, &ja.Title, &ja.Status, &ja.SalaryMin, &ja.SalaryMax, &ja.SalaryCurrency,
		&ja.ContractType, &ja.Location, &ja.Remote, &ja.Benefits, &ja.AnnouncementURL, &ja.AppliedAt,
		&ja.ResponseReceived, &ja.ResponseDate, &ja.FirstContactDate, &ja.FirstContactType,
		&ja.HasTest, &ja.TestDate, &ja.TestNotes, &ja.OfferReceived, &ja.OfferDate, &ja.OfferAmount,
		&ja.Priority, &ja.Source, &ja.RecruiterContact, &ja.Notes, &ja.CreatedAt, &ja.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("job application not found or access denied")
		}
		return nil, fmt.Errorf("unable to update job application: %w", err)
	}

	return &ja, nil
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
		id := uuid.New().String()
		var ja model.JobApplication
		err := tx.QueryRow(ctx,
			`INSERT INTO job_applications (
				id, owner_user_id, company, title, status, salary_min, salary_max, salary_currency,
				contract_type, location, remote, benefits, announcement_url, applied_at,
				response_received, response_date, first_contact_date, first_contact_type,
				has_test, test_date, test_notes, offer_received, offer_date, offer_amount,
				priority, source, recruiter_contact, notes
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
				$15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28
			) RETURNING id, owner_user_id, company, title, status, salary_min, salary_max, salary_currency,
				contract_type, location, remote, benefits, announcement_url, applied_at,
				response_received, response_date, first_contact_date, first_contact_type,
				has_test, test_date, test_notes, offer_received, offer_date, offer_amount,
				priority, source, recruiter_contact, notes, created_at, updated_at`,
			id, userID, req.Company, req.Title, req.Status, req.SalaryMin, req.SalaryMax, req.SalaryCurrency,
			req.ContractType, req.Location, req.Remote, req.Benefits, req.AnnouncementURL, req.AppliedAt,
			req.ResponseReceived, req.ResponseDate, req.FirstContactDate, req.FirstContactType,
			req.HasTest, req.TestDate, req.TestNotes, req.OfferReceived, req.OfferDate, req.OfferAmount,
			req.Priority, req.Source, req.RecruiterContact, req.Notes,
		).Scan(&ja.ID, &ja.OwnerUserID, &ja.Company, &ja.Title, &ja.Status, &ja.SalaryMin, &ja.SalaryMax, &ja.SalaryCurrency,
			&ja.ContractType, &ja.Location, &ja.Remote, &ja.Benefits, &ja.AnnouncementURL, &ja.AppliedAt,
			&ja.ResponseReceived, &ja.ResponseDate, &ja.FirstContactDate, &ja.FirstContactType,
			&ja.HasTest, &ja.TestDate, &ja.TestNotes, &ja.OfferReceived, &ja.OfferDate, &ja.OfferAmount,
			&ja.Priority, &ja.Source, &ja.RecruiterContact, &ja.Notes, &ja.CreatedAt, &ja.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("unable to create job application: %w", err)
		}
		created = append(created, ja)
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
		id := uuid.New().String()
		var ja model.JobApplication
		err := tx.QueryRow(ctx,
			`INSERT INTO job_applications (
				id, owner_user_id, company, title, status, salary_min, salary_max, salary_currency,
				contract_type, location, remote, benefits, announcement_url, applied_at,
				response_received, response_date, first_contact_date, first_contact_type,
				has_test, test_date, test_notes, offer_received, offer_date, offer_amount,
				priority, source, recruiter_contact, notes
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
				$15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28
			) RETURNING id, owner_user_id, company, title, status, salary_min, salary_max, salary_currency,
				contract_type, location, remote, benefits, announcement_url, applied_at,
				response_received, response_date, first_contact_date, first_contact_type,
				has_test, test_date, test_notes, offer_received, offer_date, offer_amount,
				priority, source, recruiter_contact, notes, created_at, updated_at`,
			id, userID, req.Company, req.Title, req.Status, req.SalaryMin, req.SalaryMax, req.SalaryCurrency,
			req.ContractType, req.Location, req.Remote, req.Benefits, req.AnnouncementURL, req.AppliedAt,
			req.ResponseReceived, req.ResponseDate, req.FirstContactDate, req.FirstContactType,
			req.HasTest, req.TestDate, req.TestNotes, req.OfferReceived, req.OfferDate, req.OfferAmount,
			req.Priority, req.Source, req.RecruiterContact, req.Notes,
		).Scan(&ja.ID, &ja.OwnerUserID, &ja.Company, &ja.Title, &ja.Status, &ja.SalaryMin, &ja.SalaryMax, &ja.SalaryCurrency,
			&ja.ContractType, &ja.Location, &ja.Remote, &ja.Benefits, &ja.AnnouncementURL, &ja.AppliedAt,
			&ja.ResponseReceived, &ja.ResponseDate, &ja.FirstContactDate, &ja.FirstContactType,
			&ja.HasTest, &ja.TestDate, &ja.TestNotes, &ja.OfferReceived, &ja.OfferDate, &ja.OfferAmount,
			&ja.Priority, &ja.Source, &ja.RecruiterContact, &ja.Notes, &ja.CreatedAt, &ja.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("unable to create job application: %w", err)
		}
		created = append(created, ja)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("unable to commit transaction: %w", err)
	}

	return created, nil
}
