package repository

import (
	"context"
	"errors"

	"warungku/app/model"
	"warungku/helper"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CustomerRepository interface {
	Create(ctx context.Context, c *model.Customer) error
	FindAll(ctx context.Context, businessID string) ([]model.Customer, error)
	FindByID(ctx context.Context, businessID, id string) (*model.Customer, error)
	Update(ctx context.Context, c *model.Customer) error
}

type customerRepository struct {
	pool *pgxpool.Pool
}

func NewCustomerRepository(pool *pgxpool.Pool) CustomerRepository {
	return &customerRepository{pool: pool}
}

func (r *customerRepository) Create(ctx context.Context, c *model.Customer) error {
	query := `
		INSERT INTO customers (business_id, name, phone, note)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query, c.BusinessID, c.Name, c.Phone, c.Note).
		Scan(&c.ID, &c.CreatedAt)
}

func (r *customerRepository) FindAll(ctx context.Context, businessID string) ([]model.Customer, error) {
	query := `
		SELECT id, business_id, name, phone, note, created_at
		FROM customers
		WHERE business_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []model.Customer
	for rows.Next() {
		var c model.Customer
		if err := rows.Scan(&c.ID, &c.BusinessID, &c.Name, &c.Phone, &c.Note, &c.CreatedAt); err != nil {
			return nil, err
		}
		customers = append(customers, c)
	}

	return customers, nil
}

func (r *customerRepository) FindByID(ctx context.Context, businessID, id string) (*model.Customer, error) {
	query := `
		SELECT id, business_id, name, phone, note, created_at
		FROM customers
		WHERE business_id = $1 AND id = $2
	`
	var c model.Customer
	err := r.pool.QueryRow(ctx, query, businessID, id).
		Scan(&c.ID, &c.BusinessID, &c.Name, &c.Phone, &c.Note, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, helper.ErrNotFound
		}
		return nil, err
	}

	return &c, nil
}

func (r *customerRepository) Update(ctx context.Context, c *model.Customer) error {
	query := `
		UPDATE customers
		SET name = $1, phone = $2, note = $3
		WHERE business_id = $4 AND id = $5
	`
	tag, err := r.pool.Exec(ctx, query, c.Name, c.Phone, c.Note, c.BusinessID, c.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return helper.ErrNotFound
	}
	return nil
}
