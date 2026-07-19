package middleware

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/user/kareelio/backend/internal/model"
	"github.com/user/kareelio/backend/internal/repository"
)

func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		ip := strings.TrimSpace(parts[0])
		if parsed := net.ParseIP(ip); parsed != nil {
			return ip
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		ip := strings.TrimSpace(xri)
		if parsed := net.ParseIP(ip); parsed != nil {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

type auditDataKey string
type clientIPKey string

const auditDataCtxKey auditDataKey = "audit_data"
const clientIPCtxKey clientIPKey = "client_ip"

type AuditData struct {
	TargetType string
	TargetID   string
	Metadata   map[string]any
}

func AuditCapture(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ad := &AuditData{Metadata: make(map[string]any)}
		ctx := context.WithValue(r.Context(), auditDataCtxKey, ad)
		ctx = context.WithValue(ctx, clientIPCtxKey, ClientIP(r))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetAuditData(ctx context.Context) *AuditData {
	ad, _ := ctx.Value(auditDataCtxKey).(*AuditData)
	return ad
}

func LogAudit(ctx context.Context, auditRepo *repository.AuditRepository, action string) {
	user := GetUserFromContext(ctx)
	ad := GetAuditData(ctx)
	ip, _ := ctx.Value(clientIPCtxKey).(string)
	if ad == nil {
		return
	}

	var actorID *string
	actorEmail := "system"
	actorRole := "system"

	if user != nil {
		actorID = &user.ID
		actorEmail = user.Email
		actorRole = string(user.Role)
	}

	event := &model.AuditEvent{
		ActorUserID: actorID,
		ActorEmail:  actorEmail,
		ActorRole:   actorRole,
		ActorIP:     ip,
		Action:      action,
		TargetType:  ad.TargetType,
		TargetID:    ad.TargetID,
	}

	if metaBytes, err := json.Marshal(ad.Metadata); err == nil && len(metaBytes) > 2 {
		event.Metadata = metaBytes
	}

	_ = auditRepo.Log(ctx, event)
}
