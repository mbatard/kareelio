package repository

import (
	"database/sql"
	"encoding/base64"
	"testing"
	"time"

	"github.com/user/kareelio/backend/internal/encryption"
	"github.com/user/kareelio/backend/internal/model"
)

func TestMaterializePrefersEncryptedValues(t *testing.T) {
	mgr := mustTestManager(t)
	repo := &JobApplicationRepository{enc: mgr, requireEncryptedReads: true}

	encCompany := mustEncrypt(t, mgr, "Encrypted Co")
	encTitle := mustEncrypt(t, mgr, "Encrypted Title")
	encSalaryMin := mustEncrypt(t, mgr, "1234.5")
	encSalaryMax := mustEncrypt(t, mgr, "9876.5")
	encSalaryCurrency := mustEncrypt(t, mgr, "USD")
	encLocation := mustEncrypt(t, mgr, "Paris")
	encBenefits := mustEncrypt(t, mgr, "health")
	encAnnouncementURL := mustEncrypt(t, mgr, "https://example.com")
	encTestNotes := mustEncrypt(t, mgr, "test")
	encOfferAmount := mustEncrypt(t, mgr, "12345.67")
	encRecruiterContact := mustEncrypt(t, mgr, "Recruiter")
	encNotes := mustEncrypt(t, mgr, "private notes")

	rec := &jobApplicationRecord{
		JobApplication: model.JobApplication{
			Company:          "Plain Co",
			Title:            "Plain Title",
			SalaryMin:        floatPtr(100),
			SalaryMax:        floatPtr(200),
			SalaryCurrency:   "EUR",
			Location:         "Lyon",
			Benefits:         "plain benefits",
			AnnouncementURL:  "https://plain.example.com",
			TestNotes:        "plain test",
			OfferAmount:      floatPtr(300),
			RecruiterContact: "Plain Recruiter",
			Notes:            "plain notes",
		},
		CompanyEnc:          nullString(encCompany),
		TitleEnc:            nullString(encTitle),
		SalaryMinEnc:        nullString(encSalaryMin),
		SalaryMaxEnc:        nullString(encSalaryMax),
		SalaryCurrencyEnc:   nullString(encSalaryCurrency),
		LocationEnc:         nullString(encLocation),
		BenefitsEnc:         nullString(encBenefits),
		AnnouncementURLEnc:  nullString(encAnnouncementURL),
		TestNotesEnc:        nullString(encTestNotes),
		OfferAmountEnc:      nullString(encOfferAmount),
		RecruiterContactEnc: nullString(encRecruiterContact),
		NotesEnc:            nullString(encNotes),
	}

	got, err := repo.materialize(rec)
	if err != nil {
		t.Fatalf("materialize() error = %v", err)
	}

	assertEqual(t, got.Company, "Encrypted Co")
	assertEqual(t, got.Title, "Encrypted Title")
	assertFloat(t, got.SalaryMin, 1234.5)
	assertFloat(t, got.SalaryMax, 9876.5)
	assertEqual(t, got.SalaryCurrency, "USD")
	assertEqual(t, got.Location, "Paris")
	assertEqual(t, got.Benefits, "health")
	assertEqual(t, got.AnnouncementURL, "https://example.com")
	assertEqual(t, got.TestNotes, "test")
	assertFloat(t, got.OfferAmount, 12345.67)
	assertEqual(t, got.RecruiterContact, "Recruiter")
	assertEqual(t, got.Notes, "private notes")
}

func TestMaterializeFallsBackToPlaintext(t *testing.T) {
	repo := &JobApplicationRepository{requireEncryptedReads: false}
	rec := &jobApplicationRecord{
		JobApplication: model.JobApplication{
			Company:          "Plain Co",
			Title:            "Plain Title",
			SalaryMin:        floatPtr(100),
			SalaryMax:        floatPtr(200),
			SalaryCurrency:   "EUR",
			Location:         "Lyon",
			Benefits:         "plain benefits",
			AnnouncementURL:  "https://plain.example.com",
			TestNotes:        "plain test",
			OfferAmount:      floatPtr(300),
			RecruiterContact: "Plain Recruiter",
			Notes:            "plain notes",
		},
	}

	got, err := repo.materialize(rec)
	if err != nil {
		t.Fatalf("materialize() error = %v", err)
	}

	assertEqual(t, got.Company, "Plain Co")
	assertEqual(t, got.Title, "Plain Title")
	assertFloat(t, got.SalaryMin, 100)
	assertFloat(t, got.SalaryMax, 200)
	assertEqual(t, got.SalaryCurrency, "EUR")
	assertEqual(t, got.Location, "Lyon")
	assertEqual(t, got.Benefits, "plain benefits")
	assertEqual(t, got.AnnouncementURL, "https://plain.example.com")
	assertEqual(t, got.TestNotes, "plain test")
	assertFloat(t, got.OfferAmount, 300)
	assertEqual(t, got.RecruiterContact, "Plain Recruiter")
	assertEqual(t, got.Notes, "plain notes")
}

func TestCreateEncryptedValuesRoundTrip(t *testing.T) {
	mgr := mustTestManager(t)
	repo := &JobApplicationRepository{enc: mgr, requireEncryptedReads: true}
	req := model.CreateJobApplicationRequest{
		Company:          "Acme",
		Title:            "Engineer",
		SalaryMin:        floatPtr(1000),
		SalaryMax:        floatPtr(2000),
		SalaryCurrency:   "EUR",
		Location:         "Remote",
		Benefits:         "health",
		AnnouncementURL:  "https://example.com",
		TestNotes:        "notes",
		OfferAmount:      floatPtr(3000),
		RecruiterContact: "Recruiter",
		Notes:            "private",
	}

	enc := repo.createEncryptedValues(req)
	assertEncryptedEquals(t, mgr, enc[0].(string), "Acme")
	assertEncryptedEquals(t, mgr, enc[1].(string), "Engineer")
	assertEncryptedEquals(t, mgr, enc[2].(string), "1000")
	assertEncryptedEquals(t, mgr, enc[3].(string), "2000")
	assertEncryptedEquals(t, mgr, enc[4].(string), "EUR")
	assertEncryptedEquals(t, mgr, enc[5].(string), "Remote")
	assertEncryptedEquals(t, mgr, enc[6].(string), "health")
	assertEncryptedEquals(t, mgr, enc[7].(string), "https://example.com")
	assertEncryptedEquals(t, mgr, enc[8].(string), "notes")
	assertEncryptedEquals(t, mgr, enc[9].(string), "3000")
	assertEncryptedEquals(t, mgr, enc[10].(string), "Recruiter")
	assertEncryptedEquals(t, mgr, enc[11].(string), "private")
}

func TestMergeUpdateRequestAppliesOverrides(t *testing.T) {
	repo := &JobApplicationRepository{}
	existing := &model.JobApplication{
		Company:          "Old Co",
		Title:            "Old Title",
		Status:           model.StatusDraft,
		SalaryMin:        floatPtr(100),
		SalaryMax:        floatPtr(200),
		SalaryCurrency:   "EUR",
		ContractType:     model.ContractOther,
		Location:         "Paris",
		Remote:           model.RemoteOnSite,
		Benefits:         "old benefits",
		AnnouncementURL:  "https://old.example.com",
		AppliedAt:        timePtr(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
		ResponseReceived: false,
		TestNotes:        "old test",
		OfferAmount:      floatPtr(300),
		Priority:         model.PriorityMedium,
		Source:           model.SourceOther,
		RecruiterContact: "Old Recruiter",
		Notes:            "old notes",
	}
	newCompany := "New Co"
	newSalary := 500.0
	newNotes := "new notes"
	merged := repo.mergeUpdateRequest(existing, model.UpdateJobApplicationRequest{
		Company:   &newCompany,
		SalaryMin: &newSalary,
		Notes:     &newNotes,
	})

	assertEqual(t, merged.Company, "New Co")
	assertFloat(t, merged.SalaryMin, 500)
	assertEqual(t, merged.Notes, "new notes")
	assertEqual(t, merged.Title, "Old Title")
	assertEqual(t, merged.SalaryCurrency, "EUR")
	assertEqual(t, merged.RecruiterContact, "Old Recruiter")
}

func TestJobApplicationToCreateRequestCopiesSensitiveFields(t *testing.T) {
	appliedAt := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	responseDate := time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)
	firstContactDate := time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC)
	testDate := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	offerDate := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	firstContactType := model.ContactPhone
	app := &model.JobApplication{
		Company:          "Acme",
		Title:            "Engineer",
		Status:           model.StatusApplied,
		SalaryMin:        floatPtr(1000),
		SalaryMax:        floatPtr(2000),
		SalaryCurrency:   "EUR",
		ContractType:     model.ContractCDI,
		Location:         "Paris",
		Remote:           model.RemoteHybrid,
		Benefits:         "benefits",
		AnnouncementURL:  "https://example.com",
		AppliedAt:        &appliedAt,
		ResponseReceived: true,
		ResponseDate:     &responseDate,
		FirstContactDate: &firstContactDate,
		FirstContactType: &firstContactType,
		HasTest:          true,
		TestDate:         &testDate,
		TestNotes:        "tests",
		OfferReceived:    true,
		OfferDate:        &offerDate,
		OfferAmount:      floatPtr(3000),
		Priority:         model.PriorityHigh,
		Source:           model.SourceReferral,
		RecruiterContact: "Recruiter",
		Notes:            "notes",
	}

	req := jobApplicationToCreateRequest(app)
	assertEqual(t, req.Company, "Acme")
	assertEqual(t, req.Title, "Engineer")
	assertFloat(t, req.SalaryMin, 1000)
	assertFloat(t, req.SalaryMax, 2000)
	assertEqual(t, req.SalaryCurrency, "EUR")
	assertEqual(t, req.Location, "Paris")
	assertEqual(t, req.Benefits, "benefits")
	assertEqual(t, req.AnnouncementURL, "https://example.com")
	if req.AppliedAt == nil || !req.AppliedAt.Equal(appliedAt) {
		t.Fatalf("AppliedAt mismatch: got %v want %v", req.AppliedAt, appliedAt)
	}
	if req.ResponseDate == nil || !req.ResponseDate.Equal(responseDate) {
		t.Fatalf("ResponseDate mismatch: got %v want %v", req.ResponseDate, responseDate)
	}
	if req.FirstContactDate == nil || !req.FirstContactDate.Equal(firstContactDate) {
		t.Fatalf("FirstContactDate mismatch: got %v want %v", req.FirstContactDate, firstContactDate)
	}
	if req.FirstContactType == nil || *req.FirstContactType != firstContactType {
		t.Fatalf("FirstContactType mismatch: got %v want %v", req.FirstContactType, firstContactType)
	}
	if req.TestDate == nil || !req.TestDate.Equal(testDate) {
		t.Fatalf("TestDate mismatch: got %v want %v", req.TestDate, testDate)
	}
	assertEqual(t, req.TestNotes, "tests")
	if req.OfferDate == nil || !req.OfferDate.Equal(offerDate) {
		t.Fatalf("OfferDate mismatch: got %v want %v", req.OfferDate, offerDate)
	}
	assertFloat(t, req.OfferAmount, 3000)
	assertEqual(t, string(req.Priority), string(model.PriorityHigh))
	assertEqual(t, string(req.Source), string(model.SourceReferral))
	assertEqual(t, req.RecruiterContact, "Recruiter")
	assertEqual(t, req.Notes, "notes")
}

func TestMaterializeRejectsLegacyRowsWhenEnforced(t *testing.T) {
	repo := &JobApplicationRepository{requireEncryptedReads: true}
	rec := &jobApplicationRecord{
		JobApplication: model.JobApplication{Company: "Legacy Co", Title: "Legacy Title"},
	}

	if _, err := repo.materialize(rec); err == nil {
		t.Fatal("materialize() unexpectedly accepted a legacy row")
	}
}

func TestMaterializeCountsLegacyReadsWhenCompatibilityEnabled(t *testing.T) {
	repo := &JobApplicationRepository{requireEncryptedReads: false}
	rec := &jobApplicationRecord{
		JobApplication: model.JobApplication{Company: "Legacy Co", Title: "Legacy Title"},
	}

	if _, err := repo.materialize(rec); err != nil {
		t.Fatalf("materialize() error = %v", err)
	}
	if got := repo.LegacyReadCount(); got != 1 {
		t.Fatalf("LegacyReadCount() = %d, want 1", got)
	}
}

func mustTestManager(t *testing.T) *encryption.Manager {
	t.Helper()
	mgr, err := encryption.New("primary", base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")))
	if err != nil {
		t.Fatalf("encryption.New() error = %v", err)
	}
	return mgr
}

func mustEncrypt(t *testing.T, mgr *encryption.Manager, plaintext string) string {
	t.Helper()
	ciphertext, err := mgr.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	return ciphertext
}

func assertEncryptedEquals(t *testing.T, mgr *encryption.Manager, ciphertext, want string) {
	t.Helper()
	got, err := mgr.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if got != want {
		t.Fatalf("Decrypt() = %q, want %q", got, want)
	}
}

func assertEqual(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func assertFloat(t *testing.T, got *float64, want float64) {
	t.Helper()
	if got == nil {
		t.Fatalf("got nil, want %v", want)
	}
	if *got != want {
		t.Fatalf("got %v, want %v", *got, want)
	}
}

func floatPtr(v float64) *float64 { return &v }

func timePtr(v time.Time) *time.Time { return &v }

func nullString(v string) sql.NullString { return sql.NullString{String: v, Valid: true} }
