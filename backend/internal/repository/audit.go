package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/kareelio/backend/internal/model"
)

type AuditRepository struct {
	db *pgxpool.Pool
}

func NewAuditRepository(db *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Log(ctx context.Context, event *model.AuditEvent) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO audit_events (actor_user_id, actor_email, actor_role, actor_ip, action, target_type, target_id, metadata)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		event.ActorUserID, event.ActorEmail, event.ActorRole, event.ActorIP,
		event.Action, event.TargetType, event.TargetID, event.Metadata,
	)
	return err
}

func (r *AuditRepository) List(ctx context.Context, limit, offset int) ([]model.AuditEvent, int, error) {
	var total int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM audit_events").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to count audit events: %w", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, actor_user_id, actor_email, actor_role, actor_ip, action, target_type, target_id, metadata, created_at
		 FROM audit_events
		 ORDER BY created_at DESC
		 LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to list audit events: %w", err)
	}
	defer rows.Close()

	var events []model.AuditEvent
	for rows.Next() {
		var e model.AuditEvent
		if err := rows.Scan(&e.ID, &e.ActorUserID, &e.ActorEmail, &e.ActorRole, &e.ActorIP,
			&e.Action, &e.TargetType, &e.TargetID, &e.Metadata, &e.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("unable to scan audit event: %w", err)
		}
		events = append(events, e)
	}

	return events, total, nil
}

func MetadataJSON(data map[string]any) json.RawMessage {
	b, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	return b
}
