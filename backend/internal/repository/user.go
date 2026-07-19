package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/kareelio/backend/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, req model.CreateUserRequest, role model.UserRole) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("unable to hash password: %w", err)
	}

	var user model.User
	err = r.db.QueryRow(ctx,
		`INSERT INTO users (email, display_name, description, password_hash, role, is_active, language, theme)
		 VALUES ($1, $2, $3, $4, $5, true, 'system', 'system')
		 RETURNING id, email, display_name, description, role, is_active, language, theme, created_at, updated_at`,
		req.Email, req.DisplayName, req.Description, string(hash), role,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.Description, &user.Role, &user.IsActive, &user.Language, &user.Theme, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to create user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := r.db.QueryRow(ctx,
		`SELECT id, email, display_name, description, role, is_active, language, theme, password_hash, created_at, updated_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.Description, &user.Role, &user.IsActive, &user.Language, &user.Theme, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.QueryRow(ctx,
		`SELECT id, email, display_name, description, role, is_active, language, theme, password_hash, created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.Description, &user.Role, &user.IsActive, &user.Language, &user.Theme, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, id string, req model.UpdateUserRequest) (*model.User, error) {
	sets := []string{}
	args := []any{}
	argIdx := 1

	if req.Email != nil {
		sets = append(sets, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, *req.Email)
		argIdx++
	}
	if req.DisplayName != nil {
		sets = append(sets, fmt.Sprintf("display_name = $%d", argIdx))
		args = append(args, *req.DisplayName)
		argIdx++
	}
	if req.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *req.Description)
		argIdx++
	}
	if req.IsActive != nil {
		sets = append(sets, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *req.IsActive)
		argIdx++
	}

	if len(sets) == 0 {
		return r.GetByID(ctx, id)
	}

	sets = append(sets, "updated_at = NOW()")

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d RETURNING id, email, display_name, description, role, is_active, language, theme, created_at, updated_at",
		joinStrings(sets, ", "), argIdx)
	args = append(args, id)

	var user model.User
	err := r.db.QueryRow(ctx, query, args...).Scan(&user.ID, &user.Email, &user.DisplayName, &user.Description, &user.Role, &user.IsActive, &user.Language, &user.Theme, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to update user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id string, req model.UpdateProfileRequest) (*model.User, error) {
	sets := []string{}
	args := []any{}
	argIdx := 1

	if req.Email != nil {
		sets = append(sets, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, *req.Email)
		argIdx++
	}
	if req.DisplayName != nil {
		sets = append(sets, fmt.Sprintf("display_name = $%d", argIdx))
		args = append(args, *req.DisplayName)
		argIdx++
	}
	if req.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *req.Description)
		argIdx++
	}
	if req.Password != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("unable to hash password: %w", err)
		}
		sets = append(sets, fmt.Sprintf("password_hash = $%d", argIdx))
		args = append(args, string(hash))
		argIdx++
	}
	if req.Language != nil {
		sets = append(sets, fmt.Sprintf("language = $%d", argIdx))
		args = append(args, *req.Language)
		argIdx++
	}
	if req.Theme != nil {
		sets = append(sets, fmt.Sprintf("theme = $%d", argIdx))
		args = append(args, *req.Theme)
		argIdx++
	}

	if len(sets) == 0 {
		return r.GetByID(ctx, id)
	}

	sets = append(sets, "updated_at = NOW()")

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d RETURNING id, email, display_name, description, role, is_active, language, theme, created_at, updated_at",
		joinStrings(sets, ", "), argIdx)
	args = append(args, id)

	var user model.User
	err := r.db.QueryRow(ctx, query, args...).Scan(&user.ID, &user.Email, &user.DisplayName, &user.Description, &user.Role, &user.IsActive, &user.Language, &user.Theme, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to update profile: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id string, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("unable to hash password: %w", err)
	}

	_, err = r.db.Exec(ctx, "UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2", string(hash), id)
	return err
}

func (r *UserRepository) List(ctx context.Context) ([]model.User, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, email, display_name, description, role, is_active, language, theme, created_at, updated_at
		 FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("unable to list users: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &u.Description, &u.Role, &u.IsActive, &u.Language, &u.Theme, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("unable to scan user: %w", err)
		}
		users = append(users, u)
	}

	return users, nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, "DELETE FROM users WHERE id = $1 AND role != 'admin'", id)
	if err != nil {
		return fmt.Errorf("unable to delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("cannot delete admin user or user not found")
	}
	return nil
}

func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
