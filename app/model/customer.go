package model

import "time"

type Customer struct {
	ID         string    `json:"id"`
	BusinessID string    `json:"business_id"`
	Name       string    `json:"name"`
	Phone      *string   `json:"phone,omitempty"`
	Note       *string   `json:"note,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreateCustomerRequest struct {
	Name  string  `json:"name"`
	Phone *string `json:"phone"`
	Note  *string `json:"note"`
}

type UpdateCustomerRequest struct {
	Name  string  `json:"name"`
	Phone *string `json:"phone"`
	Note  *string `json:"note"`
}
