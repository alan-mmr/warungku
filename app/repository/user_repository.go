package repository

import (
	"context"
	"errors"
	"time"

	"warungku/app/model"
	"warungku/helper"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	CreateUserWithBusiness(ctx context.Context, u *model.User, b *model.Business) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
	UpdatePassword(ctx context.Context, id, passwordHash string) error
}

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{pool: pool}
}

func (r *userRepository) CreateUserWithBusiness(ctx context.Context, u *model.User, b *model.Business) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	now := time.Now()

	// 1. Insert user first with business_id null
	userQuery := `
		INSERT INTO users (name, email, password_hash, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRow(ctx, userQuery, u.Name, u.Email, u.PasswordHash, u.Role, true, now, now).
		Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return err
	}

	// 2. Insert business with owner_id = user.id
	b.OwnerID = u.ID
	bizQuery := `
		INSERT INTO businesses (owner_id, name, business_type, address, phone, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRow(ctx, bizQuery, b.OwnerID, b.Name, b.BusinessType, b.Address, b.Phone, true, now, now).
		Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return err
	}

	// 3. Update user.business_id = business.id
	u.BusinessID = &b.ID
	updateUserQuery := `UPDATE users SET business_id = $1 WHERE id = $2`
	_, err = tx.Exec(ctx, updateUserQuery, b.ID, u.ID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, business_id, name, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	row := r.pool.QueryRow(ctx, query, email)

	var u model.User
	err := row.Scan(&u.ID, &u.BusinessID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, helper.ErrNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	query := `
		SELECT id, business_id, name, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)

	var u model.User
	err := row.Scan(&u.ID, &u.BusinessID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, helper.ErrNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`
	tag, err := r.pool.Exec(ctx, query, passwordHash, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return helper.ErrNotFound
	}
	return nil
}
