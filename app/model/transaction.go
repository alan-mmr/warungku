package model

import "time"

type Transaction struct {
	ID            string            `json:"id"`
	BusinessID    string            `json:"business_id"`
	InvoiceNo     string            `json:"invoice_no"`
	CashierID     *string           `json:"cashier_id,omitempty"`
	CustomerID    *string           `json:"customer_id,omitempty"`
	Subtotal      float64           `json:"subtotal"`
	Discount      float64           `json:"discount"`
	Total         float64           `json:"total"`
	PaidAmount    float64           `json:"paid_amount"`
	PaymentMethod string            `json:"payment_method"` // 'cash', 'qris', 'transfer', 'debt'
	Status        string            `json:"status"`         // 'paid', 'partial', 'debt', 'void'
	Note          string            `json:"note"`
	Items         []TransactionItem `json:"items,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
}

type TransactionItem struct {
	ID            string   `json:"id"`
	TransactionID string   `json:"transaction_id"`
	ProductID     *string  `json:"product_id,omitempty"`
	VariantID     *string  `json:"variant_id,omitempty"`
	ProductName   string   `json:"product_name"`
	Qty           float64  `json:"qty"`
	Price         float64  `json:"price"`
	CostPrice     float64  `json:"cost_price"`
	Subtotal      float64  `json:"subtotal"`
}

type CheckoutItemRequest struct {
	ProductID *string `json:"product_id"`
	VariantID *string `json:"variant_id"`
	Qty       float64 `json:"qty"`
}

type CreateTransactionRequest struct {
	CustomerID    *string               `json:"customer_id"`
	Items         []CheckoutItemRequest `json:"items"`
	Discount      float64               `json:"discount"`
	PaidAmount    float64               `json:"paid_amount"`
	PaymentMethod string                `json:"payment_method"`
	Note          string                `json:"note"`
	DueDate       *string               `json:"due_date"` // format YYYY-MM-DD jika ada utang
}
