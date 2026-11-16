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

type ProductRepository interface {
	Create(ctx context.Context, p *model.Product) error
	FindAll(ctx context.Context, businessID string, categoryID string, search string) ([]model.Product, error)
	FindByID(ctx context.Context, businessID, id string) (*model.Product, error)
	Update(ctx context.Context, p *model.Product) error
	DeleteSoft(ctx context.Context, businessID, id string) error
	CreateVariant(ctx context.Context, businessID string, v *model.ProductVariant) error
}

type productRepository struct {
	pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) ProductRepository {
	return &productRepository{pool: pool}
}

func (r *productRepository) Create(ctx context.Context, p *model.Product) error {
	query := `
		INSERT INTO products (business_id, category_id, name, sku, unit, price, cost_price, track_stock, stock, min_stock, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, true, NOW(), NOW())
		RETURNING id, is_active, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		p.BusinessID, p.CategoryID, p.Name, p.SKU, p.Unit, p.Price, p.CostPrice, p.TrackStock, p.Stock, p.MinStock,
	).Scan(&p.ID, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
}

func (r *productRepository) FindAll(ctx context.Context, businessID string, categoryID string, search string) ([]model.Product, error) {
	query := `
		SELECT p.id, p.business_id, p.category_id, c.name AS category_name, p.name, p.sku, p.unit, 
		       p.price, p.cost_price, p.track_stock, p.stock, p.min_stock, p.is_active, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.business_id = $1 AND p.is_active = true
	`
	args := []any{businessID}
	argIdx := 2

	if categoryID != "" {
		query += fmt.Sprintf(" AND p.category_id = $%d", argIdx)
		args = append(args, categoryID)
		argIdx++
	}

	if search != "" {
		query += fmt.Sprintf(" AND (p.name ILIKE $%d OR p.sku ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+search+"%")
		argIdx++
	}

	query += " ORDER BY p.name ASC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(
			&p.ID, &p.BusinessID, &p.CategoryID, &p.CategoryName, &p.Name, &p.SKU, &p.Unit,
			&p.Price, &p.CostPrice, &p.TrackStock, &p.Stock, &p.MinStock, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

func (r *productRepository) FindByID(ctx context.Context, businessID, id string) (*model.Product, error) {
	query := `
		SELECT p.id, p.business_id, p.category_id, c.name AS category_name, p.name, p.sku, p.unit, 
		       p.price, p.cost_price, p.track_stock, p.stock, p.min_stock, p.is_active, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.business_id = $1 AND p.id = $2 AND p.is_active = true
	`
	var p model.Product
	err := r.pool.QueryRow(ctx, query, businessID, id).Scan(
		&p.ID, &p.BusinessID, &p.CategoryID, &p.CategoryName, &p.Name, &p.SKU, &p.Unit,
		&p.Price, &p.CostPrice, &p.TrackStock, &p.Stock, &p.MinStock, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, helper.ErrNotFound
		}
		return nil, err
	}

	// Fetch variants
	vQuery := `
		SELECT id, product_id, name, extra_price, is_active
		FROM product_variants
		WHERE product_id = $1 AND is_active = true
		ORDER BY extra_price ASC
	`
	vRows, err := r.pool.Query(ctx, vQuery, p.ID)
	if err != nil {
		return nil, err
	}
	defer vRows.Close()

	for vRows.Next() {
		var v model.ProductVariant
		if err := vRows.Scan(&v.ID, &v.ProductID, &v.Name, &v.ExtraPrice, &v.IsActive); err != nil {
			return nil, err
		}
		p.Variants = append(p.Variants, v)
	}

	return &p, nil
}

func (r *productRepository) Update(ctx context.Context, p *model.Product) error {
	query := `
		UPDATE products
		SET category_id = $1, name = $2, sku = $3, unit = $4, price = $5, cost_price = $6, track_stock = $7, min_stock = $8, updated_at = NOW()
		WHERE business_id = $9 AND id = $10 AND is_active = true
	`
	tag, err := r.pool.Exec(ctx, query,
		p.CategoryID, p.Name, p.SKU, p.Unit, p.Price, p.CostPrice, p.TrackStock, p.MinStock, p.BusinessID, p.ID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return helper.ErrNotFound
	}
	return nil
}

func (r *productRepository) DeleteSoft(ctx context.Context, businessID, id string) error {
	query := `
		UPDATE products
		SET is_active = false, updated_at = NOW()
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

func (r *productRepository) CreateVariant(ctx context.Context, businessID string, v *model.ProductVariant) error {
	// Verify product exists and belongs to business
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1 AND business_id = $2 AND is_active = true)`
	_ = r.pool.QueryRow(ctx, checkQuery, v.ProductID, businessID).Scan(&exists)
	if !exists {
		return helper.ErrNotFound
	}

	query := `
		INSERT INTO product_variants (product_id, name, extra_price, is_active)
		VALUES ($1, $2, $3, true)
		RETURNING id, is_active
	`
	return r.pool.QueryRow(ctx, query, v.ProductID, v.Name, v.ExtraPrice).
		Scan(&v.ID, &v.IsActive)
}
