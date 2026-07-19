package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/kareelio/backend/internal/model"
)

type SessionRepository struct {
	db            *pgxpool.Pool
	durationHours int
}

func NewSessionRepository(db *pgxpool.Pool, durationHours int) *SessionRepository {
	return &SessionRepository{db: db, durationHours: durationHours}
}

func (r *SessionRepository) Create(ctx context.Context, userID string) (*model.Session, error) {
	id := uuid.New().String()
	expiresAt := time.Now().Add(time.Duration(r.durationHours) * time.Hour)

	var session model.Session
	err := r.db.QueryRow(ctx,
		`INSERT INTO sessions (id, user_id, expires_at)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, expires_at, created_at`,
		id, userID, expiresAt,
	).Scan(&session.ID, &session.UserID, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to create session: %w", err)
	}

	return &session, nil
}

func (r *SessionRepository) GetByID(ctx context.Context, id string) (*model.Session, error) {
	var session model.Session
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, expires_at, created_at
		 FROM sessions WHERE id = $1 AND expires_at > NOW()`,
		id,
	).Scan(&session.ID, &session.UserID, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("session not found or expired: %w", err)
	}

	return &session, nil
}

func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM sessions WHERE id = $1", id)
	return err
}

func (r *SessionRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.db.Exec(ctx, "DELETE FROM sessions WHERE expires_at <= NOW()")
	return err
}

func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM sessions WHERE user_id = $1", userID)
	return err
}
