package model

import "time"

type Product struct {
	ID          string           `json:"id"`
	BusinessID  string           `json:"business_id"`
	CategoryID  *string          `json:"category_id,omitempty"`
	Name        string           `json:"name"`
	SKU         *string          `json:"sku,omitempty"`
	Unit        string           `json:"unit"`
	Price       float64          `json:"price"`
	CostPrice   float64          `json:"cost_price"`
	TrackStock  bool             `json:"track_stock"`
	Stock       float64          `json:"stock"`
	MinStock    float64          `json:"min_stock"`
	IsActive    bool             `json:"is_active"`
	Variants    []ProductVariant `json:"variants,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

type ProductVariant struct {
	ID         string  `json:"id"`
	ProductID  string  `json:"product_id"`
	Name       string  `json:"name"`
	ExtraPrice float64 `json:"extra_price"`
	IsActive   bool    `json:"is_active"`
}

type CreateProductRequest struct {
	CategoryID *string `json:"category_id"`
	Name       string  `json:"name"`
	SKU        *string `json:"sku"`
	Unit       string  `json:"unit"`
	Price      float64 `json:"price"`
	CostPrice  float64 `json:"cost_price"`
	TrackStock bool    `json:"track_stock"`
	Stock      float64 `json:"stock"`
	MinStock   float64 `json:"min_stock"`
}
