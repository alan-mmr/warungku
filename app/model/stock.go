package model

import "time"

type StockMovement struct {
	ID          string    `json:"id"`
	BusinessID  string    `json:"business_id"`
	ProductID   string    `json:"product_id"`
	ProductName string    `json:"product_name,omitempty"`
	Type        string    `json:"type"` // 'in', 'out', 'adjust', 'sale', 'void_return'
	Qty         float64   `json:"qty"`
	Note        string    `json:"note"`
	RefType     *string   `json:"ref_type,omitempty"`
	RefID       *string   `json:"ref_id,omitempty"`
	CreatedBy   *string   `json:"created_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type StockInRequest struct {
	ProductID string  `json:"product_id"`
	Qty       float64 `json:"qty"`
	Note      string  `json:"note"`
}

type StockAdjustRequest struct {
	ProductID   string  `json:"product_id"`
	ActualStock float64 `json:"actual_stock"`
	Note        string  `json:"note"`
}
