package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailVerificationRepository struct {
	db *pgxpool.Pool
}

func NewEmailVerificationRepository(db *pgxpool.Pool) *EmailVerificationRepository {
	return &EmailVerificationRepository{db: db}
}

func (r *EmailVerificationRepository) Create(ctx context.Context, userID string, tokenHash string, expiresAt string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	return err
}

func (r *EmailVerificationRepository) GetValid(ctx context.Context, tokenHash string) (string, error) {
	var userID string
	err := r.db.QueryRow(ctx,
		`SELECT user_id FROM email_verification_tokens
		 WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW()`,
		tokenHash,
	).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("token not found or expired")
	}
	return userID, nil
}

func (r *EmailVerificationRepository) MarkUsed(ctx context.Context, tokenHash string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE email_verification_tokens SET used_at = NOW() WHERE token_hash = $1`,
		tokenHash,
	)
	return err
}

func (r *EmailVerificationRepository) DeleteForUser(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM email_verification_tokens WHERE user_id = $1`,
		userID,
	)
	return err
}

func (r *EmailVerificationRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM email_verification_tokens WHERE expires_at < NOW() OR used_at IS NOT NULL`,
	)
	return err
}
