package repository

import (
	"context"
	"errors"

	"warungku/app/model"
	"warungku/helper"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository interface {
	Create(ctx context.Context, c *model.Category) error
	FindAll(ctx context.Context, businessID string) ([]model.Category, error)
	FindByID(ctx context.Context, businessID, id string) (*model.Category, error)
	Update(ctx context.Context, c *model.Category) error
	DeleteSoft(ctx context.Context, businessID, id string) error
}

type categoryRepository struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) CategoryRepository {
	return &categoryRepository{pool: pool}
}

func (r *categoryRepository) Create(ctx context.Context, c *model.Category) error {
	query := `
		INSERT INTO categories (business_id, name, is_active, created_at)
		VALUES ($1, $2, true, NOW())
		RETURNING id, is_active, created_at
	`
	return r.pool.QueryRow(ctx, query, c.BusinessID, c.Name).
		Scan(&c.ID, &c.IsActive, &c.CreatedAt)
}

func (r *categoryRepository) FindAll(ctx context.Context, businessID string) ([]model.Category, error) {
	query := `
		SELECT id, business_id, name, is_active, created_at
		FROM categories
		WHERE business_id = $1 AND is_active = true
		ORDER BY name ASC
	`
	rows, err := r.pool.Query(ctx, query, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.BusinessID, &c.Name, &c.IsActive, &c.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}

	return categories, nil
}

func (r *categoryRepository) FindByID(ctx context.Context, businessID, id string) (*model.Category, error) {
	query := `
		SELECT id, business_id, name, is_active, created_at
		FROM categories
		WHERE business_id = $1 AND id = $2 AND is_active = true
	`
	var c model.Category
	err := r.pool.QueryRow(ctx, query, businessID, id).
		Scan(&c.ID, &c.BusinessID, &c.Name, &c.IsActive, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, helper.ErrNotFound
		}
		return nil, err
	}

	return &c, nil
}

func (r *categoryRepository) Update(ctx context.Context, c *model.Category) error {
	query := `
		UPDATE categories
		SET name = $1
		WHERE business_id = $2 AND id = $3 AND is_active = true
	`
	tag, err := r.pool.Exec(ctx, query, c.Name, c.BusinessID, c.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return helper.ErrNotFound
	}
	return nil
}

func (r *categoryRepository) DeleteSoft(ctx context.Context, businessID, id string) error {
	query := `
		UPDATE categories
		SET is_active = false
		WHERE business_id = $1 AND id = $2 AND is_active = true
	`
	tag, err := r.pool.Exec(ctx, query, businessID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return helper.ErrNotFound
	}
	return nil
}
