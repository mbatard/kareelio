package handler

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/user/kareelio/backend/internal/middleware"
	"github.com/user/kareelio/backend/internal/model"
	"github.com/user/kareelio/backend/internal/repository"
)

const maxImportRows = 1000

type CSVHandler struct {
	jaRepo    *repository.JobApplicationRepository
	auditRepo *repository.AuditRepository
}

func NewCSVHandler(jaRepo *repository.JobApplicationRepository, auditRepo *repository.AuditRepository) *CSVHandler {
	return &CSVHandler{jaRepo: jaRepo, auditRepo: auditRepo}
}

var csvHeaders = []string{
	"company", "title", "status", "salary_min", "salary_max", "salary_currency",
	"contract_type", "location", "remote", "benefits", "announcement_url",
	"applied_at", "response_received", "response_date", "first_contact_date",
	"first_contact_type", "has_test", "test_date", "test_notes", "offer_received",
	"offer_date", "offer_amount", "priority", "source", "recruiter_contact", "notes",
}

var validStatuses = map[string]bool{
	"draft": true, "applied": true, "responded": true, "interview": true,
	"test": true, "offer": true, "rejected": true, "withdrawn": true,
}
var validRemote = map[string]bool{"on_site": true, "hybrid": true, "full_remote": true}
var validContract = map[string]bool{
	"cdi": true, "cdd": true, "freelance": true, "internship": true, "apprentice": true, "other": true,
}
var validPriority = map[string]bool{"low": true, "medium": true, "high": true}
var validSource = map[string]bool{
	"linkedin": true, "indeed": true, "referral": true, "agency": true, "website": true, "wttj": true, "other": true,
}
var validContact = map[string]bool{"video": true, "phone": true, "in_person": true}

func (h *CSVHandler) Export(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	apps, err := h.jaRepo.List(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to list job applications"})
		return
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	if err := writer.Write(csvHeaders); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to write CSV"})
		return
	}

	for _, app := range apps {
		row := []string{
			sanitizeCSVField(app.Company),
			sanitizeCSVField(app.Title),
			sanitizeCSVField(string(app.Status)),
			sanitizeCSVField(floatPtrToString(app.SalaryMin)),
			sanitizeCSVField(floatPtrToString(app.SalaryMax)),
			sanitizeCSVField(app.SalaryCurrency),
			sanitizeCSVField(string(app.ContractType)),
			sanitizeCSVField(app.Location),
			sanitizeCSVField(string(app.Remote)),
			sanitizeCSVField(app.Benefits),
			sanitizeCSVField(app.AnnouncementURL),
			sanitizeCSVField(timePtrToString(app.AppliedAt)),
			sanitizeCSVField(strconv.FormatBool(app.ResponseReceived)),
			sanitizeCSVField(timePtrToString(app.ResponseDate)),
			sanitizeCSVField(timePtrToString(app.FirstContactDate)),
			sanitizeCSVField(contactPtrToString(app.FirstContactType)),
			sanitizeCSVField(strconv.FormatBool(app.HasTest)),
			sanitizeCSVField(timePtrToString(app.TestDate)),
			sanitizeCSVField(app.TestNotes),
			sanitizeCSVField(strconv.FormatBool(app.OfferReceived)),
			sanitizeCSVField(timePtrToString(app.OfferDate)),
			sanitizeCSVField(floatPtrToString(app.OfferAmount)),
			sanitizeCSVField(string(app.Priority)),
			sanitizeCSVField(string(app.Source)),
			sanitizeCSVField(app.RecruiterContact),
			sanitizeCSVField(app.Notes),
		}
		if err := writer.Write(row); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to write CSV row"})
			return
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to flush CSV"})
		return
	}

	ad := middleware.GetAuditData(r.Context())
	if ad != nil {
		ad.TargetType = "job_application"
		ad.Metadata["exported_count"] = len(apps)
	}
	middleware.LogAudit(r.Context(), h.auditRepo, model.AuditActionJobAppsExported)

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=kareelio_export_%s.csv", time.Now().Format("2006-01-02")))
	w.Write(buf.Bytes())
}

func (h *CSVHandler) Import(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unable to parse upload"})
		return
	}

	mode := r.FormValue("mode")
	if mode != "append" && mode != "replace" {
		mode = "append"
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no file provided"})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid CSV format"})
		return
	}

	if len(records) < 2 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "CSV file is empty or has no data rows"})
		return
	}

	if len(records)-1 > maxImportRows {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("too many rows: maximum %d allowed", maxImportRows)})
		return
	}

	header := records[0]
	colIndex := make(map[string]int)
	for i, h := range header {
		colIndex[strings.TrimSpace(h)] = i
	}

	for _, required := range []string{"company", "title"} {
		if _, ok := colIndex[required]; !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("missing required column: %s", required)})
			return
		}
	}

	var apps []model.CreateJobApplicationRequest
	for i, row := range records[1:] {
		get := func(key string) string {
			if idx, ok := colIndex[key]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		company := get("company")
		title := get("title")
		if company == "" || title == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("row %d: company and title are required", i+2)})
			return
		}

		if len(company) > 255 || len(title) > 255 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("row %d: company or title exceeds 255 characters", i+2)})
			return
		}

		status := getOrFallback(get("status"), "draft")
		if !validStatuses[status] {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("row %d: invalid status '%s'", i+2, status)})
			return
		}

		remote := getOrFallback(get("remote"), "on_site")
		if !validRemote[remote] {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("row %d: invalid remote '%s'", i+2, remote)})
			return
		}

		contract := getOrFallback(get("contract_type"), "other")
		if !validContract[contract] {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("row %d: invalid contract_type '%s'", i+2, contract)})
			return
		}

		priority := getOrFallback(get("priority"), "medium")
		if !validPriority[priority] {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("row %d: invalid priority '%s'", i+2, priority)})
			return
		}

		source := getOrFallback(get("source"), "other")
		if !validSource[source] {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("row %d: invalid source '%s'", i+2, source)})
			return
		}

		firstContactType := get("first_contact_type")
		if firstContactType != "" && !validContact[firstContactType] {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("row %d: invalid first_contact_type '%s'", i+2, firstContactType)})
			return
		}

		appliedAt, err := parseOptionalTime(get("applied_at"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("row %d: invalid applied_at date (use YYYY-MM-DD)", i+2)})
			return
		}
		responseDate, err := parseOptionalTime(get("response_date"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("row %d: invalid response_date", i+2)})
			return
		}
		firstContactDate, err := parseOptionalTime(get("first_contact_date"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("row %d: invalid first_contact_date", i+2)})
			return
		}
		testDate, err := parseOptionalTime(get("test_date"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("row %d: invalid test_date", i+2)})
			return
		}
		offerDate, err := parseOptionalTime(get("offer_date"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("row %d: invalid offer_date", i+2)})
			return
		}

		salaryMin, err := parseOptionalFloat(get("salary_min"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("row %d: invalid salary_min", i+2)})
			return
		}
		salaryMax, err := parseOptionalFloat(get("salary_max"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("row %d: invalid salary_max", i+2)})
			return
		}
		offerAmount, err := parseOptionalFloat(get("offer_amount"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("row %d: invalid offer_amount", i+2)})
			return
		}

		ct := parseOptionalContactType(firstContactType)

		app := model.CreateJobApplicationRequest{
			Company:          company,
			Title:            title,
			Status:           model.JobStatus(status),
			SalaryMin:        salaryMin,
			SalaryMax:        salaryMax,
			SalaryCurrency:   getOrFallback(get("salary_currency"), "EUR"),
			ContractType:     model.ContractType(contract),
			Location:         get("location"),
			Remote:           model.RemoteType(remote),
			Benefits:         get("benefits"),
			AnnouncementURL:  get("announcement_url"),
			AppliedAt:        appliedAt,
			ResponseReceived: parseOptionalBool(get("response_received")),
			ResponseDate:     responseDate,
			FirstContactDate: firstContactDate,
			FirstContactType: ct,
			HasTest:          parseOptionalBool(get("has_test")),
			TestDate:         testDate,
			TestNotes:        get("test_notes"),
			OfferReceived:    parseOptionalBool(get("offer_received")),
			OfferDate:        offerDate,
			OfferAmount:      offerAmount,
			Priority:         model.Priority(priority),
			Source:           model.Source(source),
			RecruiterContact: get("recruiter_contact"),
			Notes:            get("notes"),
		}
		apps = append(apps, app)
	}

	var created []model.JobApplication
	if mode == "replace" {
		created, err = h.jaRepo.ReplaceAll(r.Context(), userID, apps)
	} else {
		created, err = h.jaRepo.BulkCreate(r.Context(), userID, apps)
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to import applications"})
		return
	}

	ad := middleware.GetAuditData(r.Context())
	if ad != nil {
		ad.TargetType = "job_application"
		ad.Metadata["imported_count"] = len(created)
		ad.Metadata["mode"] = mode
	}
	middleware.LogAudit(r.Context(), h.auditRepo, model.AuditActionJobAppsImported)

	writeJSON(w, http.StatusOK, map[string]any{"imported": len(created)})
}

func sanitizeCSVField(s string) string {
	if s == "" {
		return s
	}
	first := s[0]
	if first == '=' || first == '+' || first == '-' || first == '@' || first == '\t' || first == '\r' || first == '\n' {
		return "'" + s
	}
	return s
}

func floatPtrToString(f *float64) string {
	if f == nil {
		return ""
	}
	return strconv.FormatFloat(*f, 'f', -1, 64)
}

func timePtrToString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

func contactPtrToString(c *model.ContactType) string {
	if c == nil {
		return ""
	}
	return string(*c)
}

func parseOptionalFloat(s string) (*float64, error) {
	if s == "" {
		return nil, nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid float: %s", s)
	}
	return &f, nil
}

func parseOptionalTime(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, fmt.Errorf("invalid date: %s", s)
	}
	return &t, nil
}

func parseOptionalBool(s string) bool {
	s = strings.ToLower(s)
	return s == "true" || s == "1" || s == "yes"
}

func parseOptionalContactType(s string) *model.ContactType {
	if s == "" {
		return nil
	}
	ct := model.ContactType(s)
	return &ct
}

func getOrFallback(val, fallback string) string {
	if val == "" {
		return fallback
	}
	return val
}
