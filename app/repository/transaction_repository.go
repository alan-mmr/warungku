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

type TransactionRepository interface {
	CreateTransactionAtomic(ctx context.Context, t *model.Transaction, dueDate *string) error
	FindAll(ctx context.Context, businessID string, status string, date string) ([]model.Transaction, error)
	FindByID(ctx context.Context, businessID, id string) (*model.Transaction, error)
	VoidTransaction(ctx context.Context, businessID, id string, voidedBy string) error
}

type transactionRepository struct {
	pool *pgxpool.Pool
}

func NewTransactionRepository(pool *pgxpool.Pool) TransactionRepository {
	return &transactionRepository{pool: pool}
}

func (r *transactionRepository) CreateTransactionAtomic(ctx context.Context, t *model.Transaction, dueDate *string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Generate Invoice No (format: INV-YYYYMMDD-HHMMSS)
	t.InvoiceNo = fmt.Sprintf("INV-%s-%04d", time.Now().Format("20060102-150405"), time.Now().Nanosecond()%10000)

	// 2. Validate Products, Prices, Stock, and Calculate Subtotal
	var subtotal float64
	for i := range t.Items {
		item := &t.Items[i]
		if item.ProductID == nil {
			return errors.New("product_id wajib diisi")
		}

		var pName string
		var pPrice, pCostPrice, pStock float64
		var trackStock bool
		prodQuery := `
			SELECT name, price, cost_price, stock, track_stock 
			FROM products 
			WHERE id = $1 AND business_id = $2 AND is_active = true
		`
		err := tx.QueryRow(ctx, prodQuery, *item.ProductID, t.BusinessID).
			Scan(&pName, &pPrice, &pCostPrice, &pStock, &trackStock)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("produk ID %s tidak ditemukan atau tidak aktif", *item.ProductID)
			}
			return err
		}

		// Check Variant if provided
		if item.VariantID != nil {
			var vName string
			var vExtraPrice float64
			vQuery := `SELECT name, extra_price FROM product_variants WHERE id = $1 AND product_id = $2 AND is_active = true`
			err := tx.QueryRow(ctx, vQuery, *item.VariantID, *item.ProductID).Scan(&vName, &vExtraPrice)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
			pPrice += vExtraPrice
			pName = fmt.Sprintf("%s (%s)", pName, vName)
		}

		// Validate stock if track_stock is enabled
		if trackStock && pStock < item.Qty {
			return fmt.Errorf("%w: produk '%s' sisa %v, diminta %v", helper.ErrOutOfStock, pName, pStock, item.Qty)
		}

		// Snapshot product details into item
		item.ProductName = pName
		item.Price = pPrice
		item.CostPrice = pCostPrice
		item.Subtotal = pPrice * item.Qty
		subtotal += item.Subtotal

		// Decrement stock if track_stock is enabled
		if trackStock {
			stockDecrQuery := `UPDATE products SET stock = stock - $1, updated_at = NOW() WHERE id = $2 AND business_id = $3`
			_, err = tx.Exec(ctx, stockDecrQuery, item.Qty, *item.ProductID, t.BusinessID)
			if err != nil {
				return err
			}

			// Record Stock Movement
			mvQuery := `
				INSERT INTO stock_movements (business_id, product_id, type, qty, note, ref_type, created_by, created_at)
				VALUES ($1, $2, 'sale', -$3, $4, 'transaction', $5, NOW())
			`
			note := fmt.Sprintf("Penjualan nota %s", t.InvoiceNo)
			_, err = tx.Exec(ctx, mvQuery, t.BusinessID, *item.ProductID, item.Qty, note, t.CashierID)
			if err != nil {
				return err
			}
		}
	}

	t.Subtotal = subtotal
	t.Total = subtotal - t.Discount
	if t.Total < 0 {
		t.Total = 0
	}

	// 3. Determine status
	if t.PaidAmount >= t.Total {
		t.Status = "paid"
	} else if t.PaidAmount > 0 {
		t.Status = "partial"
	} else {
		t.Status = "debt"
	}

	// 4. Insert Transaction
	txInsertQuery := `
		INSERT INTO transactions (business_id, invoice_no, cashier_id, customer_id, subtotal, discount, total, paid_amount, payment_method, status, note, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
		RETURNING id, created_at
	`
	err = tx.QueryRow(ctx, txInsertQuery,
		t.BusinessID, t.InvoiceNo, t.CashierID, t.CustomerID, t.Subtotal, t.Discount, t.Total, t.PaidAmount, t.PaymentMethod, t.Status, t.Note,
	).Scan(&t.ID, &t.CreatedAt)
	if err != nil {
		return err
	}

	// 5. Insert Transaction Items
	itemInsertQuery := `
		INSERT INTO transaction_items (transaction_id, product_id, variant_id, product_name, qty, price, cost_price, subtotal)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	for i := range t.Items {
		item := &t.Items[i]
		item.TransactionID = t.ID
		err = tx.QueryRow(ctx, itemInsertQuery,
			t.ID, item.ProductID, item.VariantID, item.ProductName, item.Qty, item.Price, item.CostPrice, item.Subtotal,
		).Scan(&item.ID)
		if err != nil {
			return err
		}
	}

	// 6. Handle Debt if partial or unpaid
	if t.PaidAmount < t.Total {
		if t.CustomerID == nil {
			return errors.New("pelanggan wajib dipilih jika pembayaran berstatus utang/sebagian")
		}
		remaining := t.Total - t.PaidAmount
		debtQuery := `
			INSERT INTO debts (business_id, customer_id, transaction_id, amount, remaining, due_date, status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, 'open', NOW())
		`
		var dDate any = nil
		if dueDate != nil && *dueDate != "" {
			dDate = *dueDate
		}
		_, err = tx.Exec(ctx, debtQuery, t.BusinessID, *t.CustomerID, t.ID, t.Total, remaining, dDate)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *transactionRepository) FindAll(ctx context.Context, businessID string, status string, date string) ([]model.Transaction, error) {
	query := `
		SELECT id, business_id, invoice_no, cashier_id, customer_id, subtotal, discount, total, paid_amount, payment_method, status, note, created_at
		FROM transactions
		WHERE business_id = $1
	`
	args := []any{businessID}
	argIdx := 2

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	if date != "" {
		query += fmt.Sprintf(" AND DATE(created_at) = $%d", argIdx)
		args = append(args, date)
		argIdx++
	}

	query += " ORDER BY created_at DESC LIMIT 100"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []model.Transaction
	for rows.Next() {
		var t model.Transaction
		if err := rows.Scan(
			&t.ID, &t.BusinessID, &t.InvoiceNo, &t.CashierID, &t.CustomerID,
			&t.Subtotal, &t.Discount, &t.Total, &t.PaidAmount, &t.PaymentMethod,
			&t.Status, &t.Note, &t.CreatedAt,
		); err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}

	return transactions, nil
}

func (r *transactionRepository) FindByID(ctx context.Context, businessID, id string) (*model.Transaction, error) {
	query := `
		SELECT id, business_id, invoice_no, cashier_id, customer_id, subtotal, discount, total, paid_amount, payment_method, status, note, created_at
		FROM transactions
		WHERE business_id = $1 AND id = $2
	`
	var t model.Transaction
	err := r.pool.QueryRow(ctx, query, businessID, id).Scan(
		&t.ID, &t.BusinessID, &t.InvoiceNo, &t.CashierID, &t.CustomerID,
		&t.Subtotal, &t.Discount, &t.Total, &t.PaidAmount, &t.PaymentMethod,
		&t.Status, &t.Note, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, helper.ErrNotFound
		}
		return nil, err
	}

	// Fetch items
	itemQuery := `
		SELECT id, transaction_id, product_id, variant_id, product_name, qty, price, cost_price, subtotal
		FROM transaction_items
		WHERE transaction_id = $1
	`
	rows, err := r.pool.Query(ctx, itemQuery, t.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item model.TransactionItem
		if err := rows.Scan(
			&item.ID, &item.TransactionID, &item.ProductID, &item.VariantID,
			&item.ProductName, &item.Qty, &item.Price, &item.CostPrice, &item.Subtotal,
		); err != nil {
			return nil, err
		}
		t.Items = append(t.Items, item)
	}

	return &t, nil
}

func (r *transactionRepository) VoidTransaction(ctx context.Context, businessID, id string, voidedBy string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Get transaction and verify it's not already void
	var status, invoiceNo string
	err = tx.QueryRow(ctx, "SELECT status, invoice_no FROM transactions WHERE id = $1 AND business_id = $2", id, businessID).
		Scan(&status, &invoiceNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return helper.ErrNotFound
		}
		return err
	}

	if status == "void" {
		return errors.New("transaksi sudah dibatalkan sebelumnya")
	}

	// 2. Return stocks for items
	itemQuery := `SELECT product_id, qty FROM transaction_items WHERE transaction_id = $1`
	rows, err := tx.Query(ctx, itemQuery, id)
	if err != nil {
		return err
	}
	defer rows.Close()

	type returnItem struct {
		prodID *string
		qty    float64
	}
	var toReturn []returnItem
	for rows.Next() {
		var ri returnItem
		if err := rows.Scan(&ri.prodID, &ri.qty); err != nil {
			return err
		}
		toReturn = append(toReturn, ri)
	}
	rows.Close()

	for _, item := range toReturn {
		if item.prodID != nil {
			// Restore stock
			_, err = tx.Exec(ctx, "UPDATE products SET stock = stock + $1, updated_at = NOW() WHERE id = $2 AND business_id = $3", item.qty, *item.prodID, businessID)
			if err != nil {
				return err
			}

			// Record Stock Movement void_return
			mvQuery := `
				INSERT INTO stock_movements (business_id, product_id, type, qty, note, ref_type, created_by, created_at)
				VALUES ($1, $2, 'void_return', $3, $4, 'transaction', $5, NOW())
			`
			note := fmt.Sprintf("Pengembalian stok pembatalan nota %s", invoiceNo)
			_, err = tx.Exec(ctx, mvQuery, businessID, *item.prodID, item.qty, note, voidedBy)
			if err != nil {
				return err
			}
		}
	}

	// 3. Mark transaction as void
	_, err = tx.Exec(ctx, "UPDATE transactions SET status = 'void' WHERE id = $1 AND business_id = $2", id, businessID)
	if err != nil {
		return err
	}

	// 4. Cancel any related debt
	_, err = tx.Exec(ctx, "UPDATE debts SET status = 'cancelled' WHERE transaction_id = $1 AND business_id = $2", id, businessID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
