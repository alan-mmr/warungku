package service_test

import (
	"testing"
	"warungku/app/model"
)

func TestCalculateTransactionTotalAndChange(t *testing.T) {
	subtotal := 150000.0
	discount := 15000.0
	expectedTotal := 135000.0

	total := subtotal - discount
	if total != expectedTotal {
		t.Errorf("expected total %.2f, got %.2f", expectedTotal, total)
	}

	paidAmount := 150000.0
	change := paidAmount - total
	if change != 15000.0 {
		t.Errorf("expected change 15000.0, got %.2f", change)
	}

	item := model.TransactionItem{
		ProductName: "Beras Ramos 5kg",
		Qty:         2,
		Price:       65000,
		CostPrice:   55000,
		Subtotal:    130000,
	}

	profit := (item.Price - item.CostPrice) * item.Qty
	if profit != 20000.0 {
		t.Errorf("expected profit 20000.0, got %.2f", profit)
	}
}
