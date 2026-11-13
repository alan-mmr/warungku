package repository

import (
	"context"
	"errors"
	"fmt"

	"warungku/app/model"
	"warungku/helper"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DebtRepository interface {
	FindAll(ctx context.Context, businessID string, status string) ([]model.Debt, error)
	FindByID(ctx context.Context, businessID, id string) (*model.Debt, error)
	PayDebt(ctx context.Context, businessID, debtID string, payment *model.DebtPayment) error
}

type debtRepository struct {
	pool *pgxpool.Pool
}

func NewDebtRepository(pool *pgxpool.Pool) DebtRepository {
	return &debtRepository{pool: pool}
}

func (r *debtRepository) FindAll(ctx context.Context, businessID string, status string) ([]model.Debt, error) {
	query := `
		SELECT id, business_id, customer_id, transaction_id, amount, remaining, TO_CHAR(due_date, 'YYYY-MM-DD'), status, created_at
		FROM debts
		WHERE business_id = $1
	`
	args := []any{businessID}
	if status != "" {
		query += " AND status = $2"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var debts []model.Debt
	for rows.Next() {
		var d model.Debt
		if err := rows.Scan(
			&d.ID, &d.BusinessID, &d.CustomerID, &d.TransactionID,
			&d.Amount, &d.Remaining, &d.DueDate, &d.Status, &d.CreatedAt,
		); err != nil {
			return nil, err
		}
		debts = append(debts, d)
	}

	return debts, nil
}

func (r *debtRepository) FindByID(ctx context.Context, businessID, id string) (*model.Debt, error) {
	query := `
		SELECT id, business_id, customer_id, transaction_id, amount, remaining, TO_CHAR(due_date, 'YYYY-MM-DD'), status, created_at
		FROM debts
		WHERE business_id = $1 AND id = $2
	`
	var d model.Debt
	err := r.pool.QueryRow(ctx, query, businessID, id).Scan(
		&d.ID, &d.BusinessID, &d.CustomerID, &d.TransactionID,
		&d.Amount, &d.Remaining, &d.DueDate, &d.Status, &d.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, helper.ErrNotFound
		}
		return nil, err
	}

	// Fetch payments history
	payQuery := `
		SELECT id, debt_id, amount, method, note, received_by, paid_at
		FROM debt_payments
		WHERE debt_id = $1
		ORDER BY paid_at DESC
	`
	rows, err := r.pool.Query(ctx, payQuery, d.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p model.DebtPayment
		if err := rows.Scan(&p.ID, &p.DebtID, &p.Amount, &p.Method, &p.Note, &p.ReceivedBy, &p.PaidAt); err != nil {
			return nil, err
		}
		d.Payments = append(d.Payments, p)
	}

	return &d, nil
}

func (r *debtRepository) PayDebt(ctx context.Context, businessID, debtID string, payment *model.DebtPayment) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var remaining float64
	var status string
	err = tx.QueryRow(ctx, "SELECT remaining, status FROM debts WHERE id = $1 AND business_id = $2 FOR UPDATE", debtID, businessID).
		Scan(&remaining, &status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return helper.ErrNotFound
		}
		return err
	}

	if status == "paid" {
		return errors.New("utang ini sudah lunas sepenuhnya")
	}
	if status == "cancelled" {
		return errors.New("utang ini sudah dibatalkan (transaksi void)")
	}

	if payment.Amount <= 0 {
		return errors.New("nominal pembayaran harus lebih dari 0")
	}
	if payment.Amount > remaining {
		return fmt.Errorf("nominal pembayaran (%v) melebihi sisa utang (%v)", payment.Amount, remaining)
	}

	// Insert payment record
	payQuery := `
		INSERT INTO debt_payments (debt_id, amount, method, note, received_by, paid_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, paid_at
	`
	err = tx.QueryRow(ctx, payQuery, debtID, payment.Amount, payment.Method, payment.Note, payment.ReceivedBy).
		Scan(&payment.ID, &payment.PaidAt)
	if err != nil {
		return err
	}

	// Update debt remaining & status
	newRemaining := remaining - payment.Amount
	newStatus := "open"
	if newRemaining <= 0 {
		newRemaining = 0
		newStatus = "paid"
	}

	updateDebtQuery := `UPDATE debts SET remaining = $1, status = $2 WHERE id = $3`
	_, err = tx.Exec(ctx, updateDebtQuery, newRemaining, newStatus, debtID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
