package model

import "time"

type JobStatus string

const (
	StatusDraft     JobStatus = "draft"
	StatusApplied   JobStatus = "applied"
	StatusResponded JobStatus = "responded"
	StatusInterview JobStatus = "interview"
	StatusTest      JobStatus = "test"
	StatusOffer     JobStatus = "offer"
	StatusRejected  JobStatus = "rejected"
	StatusWithdrawn JobStatus = "withdrawn"
)

type RemoteType string

const (
	RemoteOnSite    RemoteType = "on_site"
	RemoteHybrid    RemoteType = "hybrid"
	RemoteFull      RemoteType = "full_remote"
)

type ContractType string

const (
	ContractCDI        ContractType = "cdi"
	ContractCDD        ContractType = "cdd"
	ContractFreelance  ContractType = "freelance"
	ContractInternship ContractType = "internship"
	ContractApprentice ContractType = "apprentice"
	ContractOther      ContractType = "other"
)

type ContactType string

const (
	ContactVideo   ContactType = "video"
	ContactPhone   ContactType = "phone"
	ContactInPerson ContactType = "in_person"
)

type Source string

const (
	SourceLinkedIn  Source = "linkedin"
	SourceIndeed    Source = "indeed"
	SourceReferral  Source = "referral"
	SourceAgency    Source = "agency"
	SourceWebsite   Source = "website"
	SourceWTTJ      Source = "wttj"
	SourceOther     Source = "other"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type JobApplication struct {
	ID                  string       `json:"id"`
	OwnerUserID         string       `json:"-"`
	Company             string       `json:"company"`
	Title               string       `json:"title"`
	Status              JobStatus    `json:"status"`
	SalaryMin           *float64     `json:"salary_min"`
	SalaryMax           *float64     `json:"salary_max"`
	SalaryCurrency      string       `json:"salary_currency"`
	ContractType        ContractType `json:"contract_type"`
	Location            string       `json:"location"`
	Remote              RemoteType   `json:"remote"`
	Benefits            string       `json:"benefits"`
	AnnouncementURL     string       `json:"announcement_url"`
	AppliedAt           *time.Time   `json:"applied_at"`
	ResponseReceived    bool         `json:"response_received"`
	ResponseDate        *time.Time   `json:"response_date"`
	FirstContactDate    *time.Time   `json:"first_contact_date"`
	FirstContactType    *ContactType `json:"first_contact_type"`
	HasTest             bool         `json:"has_test"`
	TestDate            *time.Time   `json:"test_date"`
	TestNotes           string       `json:"test_notes"`
	OfferReceived       bool         `json:"offer_received"`
	OfferDate           *time.Time   `json:"offer_date"`
	OfferAmount         *float64     `json:"offer_amount"`
	Priority            Priority     `json:"priority"`
	Source              Source       `json:"source"`
	RecruiterContact    string       `json:"recruiter_contact"`
	Notes               string       `json:"notes"`
	CreatedAt           time.Time    `json:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at"`
}

type CreateJobApplicationRequest struct {
	Company             string       `json:"company"`
	Title               string       `json:"title"`
	Status              JobStatus    `json:"status"`
	SalaryMin           *float64     `json:"salary_min"`
	SalaryMax           *float64     `json:"salary_max"`
	SalaryCurrency      string       `json:"salary_currency"`
	ContractType        ContractType `json:"contract_type"`
	Location            string       `json:"location"`
	Remote              RemoteType   `json:"remote"`
	Benefits            string       `json:"benefits"`
	AnnouncementURL     string       `json:"announcement_url"`
	AppliedAt           *time.Time   `json:"applied_at"`
	ResponseReceived    bool         `json:"response_received"`
	ResponseDate        *time.Time   `json:"response_date"`
	FirstContactDate    *time.Time   `json:"first_contact_date"`
	FirstContactType    *ContactType `json:"first_contact_type"`
	HasTest             bool         `json:"has_test"`
	TestDate            *time.Time   `json:"test_date"`
	TestNotes           string       `json:"test_notes"`
	OfferReceived       bool         `json:"offer_received"`
	OfferDate           *time.Time   `json:"offer_date"`
	OfferAmount         *float64     `json:"offer_amount"`
	Priority            Priority     `json:"priority"`
	Source              Source       `json:"source"`
	RecruiterContact    string       `json:"recruiter_contact"`
	Notes               string       `json:"notes"`
}

type UpdateJobApplicationRequest struct {
	Company             *string       `json:"company,omitempty"`
	Title               *string       `json:"title,omitempty"`
	Status              *JobStatus    `json:"status,omitempty"`
	SalaryMin           *float64      `json:"salary_min,omitempty"`
	SalaryMax           *float64      `json:"salary_max,omitempty"`
	SalaryCurrency      *string       `json:"salary_currency,omitempty"`
	ContractType        *ContractType `json:"contract_type,omitempty"`
	Location            *string       `json:"location,omitempty"`
	Remote              *RemoteType   `json:"remote,omitempty"`
	Benefits            *string       `json:"benefits,omitempty"`
	AnnouncementURL     *string       `json:"announcement_url,omitempty"`
	AppliedAt           *time.Time    `json:"applied_at,omitempty"`
	ResponseReceived    *bool         `json:"response_received,omitempty"`
	ResponseDate        *time.Time    `json:"response_date,omitempty"`
	FirstContactDate    *time.Time    `json:"first_contact_date,omitempty"`
	FirstContactType    *ContactType  `json:"first_contact_type,omitempty"`
	HasTest             *bool         `json:"has_test,omitempty"`
	TestDate            *time.Time    `json:"test_date,omitempty"`
	TestNotes           *string       `json:"test_notes,omitempty"`
	OfferReceived       *bool         `json:"offer_received,omitempty"`
	OfferDate           *time.Time    `json:"offer_date,omitempty"`
	OfferAmount         *float64      `json:"offer_amount,omitempty"`
	Priority            *Priority     `json:"priority,omitempty"`
	Source              *Source       `json:"source,omitempty"`
	RecruiterContact    *string       `json:"recruiter_contact,omitempty"`
	Notes               *string       `json:"notes,omitempty"`
}
