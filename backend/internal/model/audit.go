package model

import (
	"encoding/json"
	"time"
)

type AuditEvent struct {
	ID          string          `json:"id"`
	ActorUserID *string         `json:"actor_user_id"`
	ActorEmail  string          `json:"actor_email"`
	ActorRole   string          `json:"actor_role"`
	ActorIP     string          `json:"actor_ip"`
	Action      string          `json:"action"`
	TargetType  string          `json:"target_type"`
	TargetID    string          `json:"target_id"`
	Metadata    json.RawMessage `json:"metadata"`
	CreatedAt   time.Time       `json:"created_at"`
}

const (
	AuditActionUserCreated        = "user_created"
	AuditActionUserUpdated        = "user_updated"
	AuditActionUserActivated      = "user_activated"
	AuditActionUserDeactivated    = "user_deactivated"
	AuditActionUserDeleted        = "user_deleted"
	AuditActionJobAppCreated      = "job_application_created"
	AuditActionJobAppUpdated      = "job_application_updated"
	AuditActionJobAppDeleted      = "job_application_deleted"
	AuditActionLoginSuccess       = "login_success"
	AuditActionLoginFailed        = "login_failed"
	AuditActionLogout             = "logout"
	AuditActionProfileUpdated     = "profile_updated"
	AuditActionPasswordChanged    = "password_changed"
	AuditActionAdminPasswordReset = "user_password_changed"
	AuditActionJobAppsExported    = "job_applications_exported"
	AuditActionJobAppsImported    = "job_applications_imported"
	AuditActionUserRegistered     = "user_registered"
	AuditActionEmailVerified      = "email_verified"
)
