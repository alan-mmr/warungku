package model

import "time"

type Business struct {
	ID           string    `json:"id"`
	OwnerID      string    `json:"owner_id"`
	Name         string    `json:"name"`
	BusinessType string    `json:"business_type"`
	Address      string    `json:"address"`
	Phone        string    `json:"phone"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UpdateBusinessRequest struct {
	Name         string `json:"name"`
	BusinessType string `json:"business_type"`
	Address      string `json:"address"`
	Phone        string `json:"phone"`
}
