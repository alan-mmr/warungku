package model

import "time"

type Expense struct {
	ID         string    `json:"id"`
	BusinessID string    `json:"business_id"`
	Category   string    `json:"category"` // 'bahan', 'listrik', 'gaji', 'sewa', 'lainnya'
	Amount     float64   `json:"amount"`
	Note       string    `json:"note"`
	SpentAt    time.Time `json:"spent_at"`
	CreatedBy  *string   `json:"created_by,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreateExpenseRequest struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Note     string  `json:"note"`
	SpentAt  *string `json:"spent_at"` // ISO8601 atau YYYY-MM-DD
}
