package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"warungku/app/model"
	"warungku/helper"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExpenseRepository interface {
	Create(ctx context.Context, e *model.Expense) error
	FindAll(ctx context.Context, businessID, category, date string) ([]model.Expense, error)
	FindByID(ctx context.Context, businessID, id string) (*model.Expense, error)
	Update(ctx context.Context, e *model.Expense) error
	Delete(ctx context.Context, businessID, id string) error
}

type expenseRepository struct {
	pool *pgxpool.Pool
}

func NewExpenseRepository(pool *pgxpool.Pool) ExpenseRepository {
	return &expenseRepository{pool: pool}
}

func (r *expenseRepository) Create(ctx context.Context, e *model.Expense) error {
	query := `
		INSERT INTO expenses (business_id, category, amount, note, spent_at, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id, created_at
	`
	if e.SpentAt.IsZero() {
		e.SpentAt = time.Now()
	}
	return r.pool.QueryRow(ctx, query, e.BusinessID, e.Category, e.Amount, e.Note, e.SpentAt, e.CreatedBy).
		Scan(&e.ID, &e.CreatedAt)
}

func (r *expenseRepository) FindAll(ctx context.Context, businessID, category, date string) ([]model.Expense, error) {
	query := `
		SELECT id, business_id, category, amount, note, spent_at, created_by, created_at
		FROM expenses
		WHERE business_id = $1
	`
	args := []any{businessID}
	argIdx := 2

	if category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, category)
		argIdx++
	}
	if date != "" {
		query += fmt.Sprintf(" AND DATE(spent_at) = $%d", argIdx)
		args = append(args, date)
		argIdx++
	}
	query += " ORDER BY spent_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []model.Expense
	for rows.Next() {
		var e model.Expense
		if err := rows.Scan(&e.ID, &e.BusinessID, &e.Category, &e.Amount, &e.Note, &e.SpentAt, &e.CreatedBy, &e.CreatedAt); err != nil {
			return nil, err
		}
		expenses = append(expenses, e)
	}

	return expenses, nil
}

func (r *expenseRepository) FindByID(ctx context.Context, businessID, id string) (*model.Expense, error) {
	query := `
		SELECT id, business_id, category, amount, note, spent_at, created_by, created_at
		FROM expenses
		WHERE business_id = $1 AND id = $2
	`
	var e model.Expense
	err := r.pool.QueryRow(ctx, query, businessID, id).
		Scan(&e.ID, &e.BusinessID, &e.Category, &e.Amount, &e.Note, &e.SpentAt, &e.CreatedBy, &e.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, helper.ErrNotFound
		}
		return nil, err
	}

	return &e, nil
}

func (r *expenseRepository) Update(ctx context.Context, e *model.Expense) error {
	query := `
		UPDATE expenses
		SET category = $1, amount = $2, note = $3, spent_at = $4
		WHERE business_id = $5 AND id = $6
	`
	tag, err := r.pool.Exec(ctx, query, e.Category, e.Amount, e.Note, e.SpentAt, e.BusinessID, e.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return helper.ErrNotFound
	}
	return nil
}

func (r *expenseRepository) Delete(ctx context.Context, businessID, id string) error {
	query := `DELETE FROM expenses WHERE business_id = $1 AND id = $2`
	tag, err := r.pool.Exec(ctx, query, businessID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return helper.ErrNotFound
	}
	return nil
}
