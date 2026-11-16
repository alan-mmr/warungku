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

type StockRepository interface {
	StockIn(ctx context.Context, businessID string, req model.StockInRequest, createdBy string) error
	StockAdjust(ctx context.Context, businessID string, req model.StockAdjustRequest, createdBy string) error
	GetMovements(ctx context.Context, businessID, productID string) ([]model.StockMovement, error)
	GetLowStock(ctx context.Context, businessID string) ([]model.Product, error)
}

type stockRepository struct {
	pool *pgxpool.Pool
}

func NewStockRepository(pool *pgxpool.Pool) StockRepository {
	return &stockRepository{pool: pool}
}

func (r *stockRepository) StockIn(ctx context.Context, businessID string, req model.StockInRequest, createdBy string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Check product
	var exists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM products WHERE id = $1 AND business_id = $2 AND is_active = true)", req.ProductID, businessID).Scan(&exists)
	if err != nil || !exists {
		return helper.ErrNotFound
	}

	// Update product stock
	updateQuery := `UPDATE products SET stock = stock + $1, updated_at = NOW() WHERE id = $2 AND business_id = $3`
	_, err = tx.Exec(ctx, updateQuery, req.Qty, req.ProductID, businessID)
	if err != nil {
		return err
	}

	// Record movement
	note := req.Note
	if note == "" {
		note = "Stok masuk (restock belanja)"
	}
	mvQuery := `
		INSERT INTO stock_movements (business_id, product_id, type, qty, note, ref_type, created_by, created_at)
		VALUES ($1, $2, 'in', $3, $4, 'manual_in', $5, NOW())
	`
	_, err = tx.Exec(ctx, mvQuery, businessID, req.ProductID, req.Qty, note, createdBy)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *stockRepository) StockAdjust(ctx context.Context, businessID string, req model.StockAdjustRequest, createdBy string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Get current stock
	var currentStock float64
	err = tx.QueryRow(ctx, "SELECT stock FROM products WHERE id = $1 AND business_id = $2 AND is_active = true", req.ProductID, businessID).Scan(&currentStock)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return helper.ErrNotFound
		}
		return err
	}

	diff := req.ActualStock - currentStock

	// Update stock to actual
	updateQuery := `UPDATE products SET stock = $1, updated_at = NOW() WHERE id = $2 AND business_id = $3`
	_, err = tx.Exec(ctx, updateQuery, req.ActualStock, req.ProductID, businessID)
	if err != nil {
		return err
	}

	// Record movement
	note := req.Note
	if note == "" {
		note = fmt.Sprintf("Penyesuaian stok fisik (opname): %v -> %v", currentStock, req.ActualStock)
	}
	mvQuery := `
		INSERT INTO stock_movements (business_id, product_id, type, qty, note, ref_type, created_by, created_at)
		VALUES ($1, $2, 'adjust', $3, $4, 'stock_opname', $5, NOW())
	`
	_, err = tx.Exec(ctx, mvQuery, businessID, req.ProductID, diff, note, createdBy)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *stockRepository) GetMovements(ctx context.Context, businessID, productID string) ([]model.StockMovement, error) {
	query := `
		SELECT sm.id, sm.business_id, sm.product_id, p.name AS product_name, sm.type, sm.qty, sm.note, sm.ref_type, sm.ref_id, sm.created_by, sm.created_at
		FROM stock_movements sm
		JOIN products p ON sm.product_id = p.id
		WHERE sm.business_id = $1
	`
	args := []any{businessID}
	if productID != "" {
		query += " AND sm.product_id = $2"
		args = append(args, productID)
	}
	query += " ORDER BY sm.created_at DESC LIMIT 100"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movements []model.StockMovement
	for rows.Next() {
		var sm model.StockMovement
		if err := rows.Scan(
			&sm.ID, &sm.BusinessID, &sm.ProductID, &sm.ProductName, &sm.Type, &sm.Qty,
			&sm.Note, &sm.RefType, &sm.RefID, &sm.CreatedBy, &sm.CreatedAt,
		); err != nil {
			return nil, err
		}
		movements = append(movements, sm)
	}

	return movements, nil
}

func (r *stockRepository) GetLowStock(ctx context.Context, businessID string) ([]model.Product, error) {
	query := `
		SELECT p.id, p.business_id, p.category_id, c.name AS category_name, p.name, p.sku, p.unit, 
		       p.price, p.cost_price, p.track_stock, p.stock, p.min_stock, p.is_active, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.business_id = $1 AND p.track_stock = true AND p.stock <= p.min_stock AND p.is_active = true
		ORDER BY p.stock ASC
	`
	rows, err := r.pool.Query(ctx, query, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(
			&p.ID, &p.BusinessID, &p.CategoryID, &p.CategoryName, &p.Name, &p.SKU, &p.Unit,
			&p.Price, &p.CostPrice, &p.TrackStock, &p.Stock, &p.MinStock, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, p)
	}

	return list, nil
}
