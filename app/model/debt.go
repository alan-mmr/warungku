package model

import "time"

type Debt struct {
	ID            string        `json:"id"`
	BusinessID    string        `json:"business_id"`
	CustomerID    string        `json:"customer_id"`
	TransactionID string        `json:"transaction_id"`
	Amount        float64       `json:"amount"`
	Remaining     float64       `json:"remaining"`
	DueDate       *string       `json:"due_date,omitempty"`
	Status        string        `json:"status"` // 'open', 'paid', 'cancelled'
	Payments      []DebtPayment `json:"payments,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
}

type DebtPayment struct {
	ID         string    `json:"id"`
	DebtID     string    `json:"debt_id"`
	Amount     float64   `json:"amount"`
	Method     string    `json:"method"`
	Note       string    `json:"note"`
	ReceivedBy *string   `json:"received_by,omitempty"`
	PaidAt     time.Time `json:"paid_at"`
}

type PayDebtRequest struct {
	Amount float64 `json:"amount"`
	Method string  `json:"method"`
	Note   string  `json:"note"`
}
