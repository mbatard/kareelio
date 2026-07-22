package model

import "time"

type EmailVerificationToken struct {
	ID        string     `json:"-"`
	UserID    string     `json:"-"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"-"`
	UsedAt    *time.Time `json:"-"`
	CreatedAt time.Time  `json:"-"`
}
