package service_test

import (
	"testing"
	"warungku/app/model"
)

func TestLowStockThresholdDetection(t *testing.T) {
	product := model.Product{
		Name:       "Minyak Goreng 2L",
		TrackStock: true,
		Stock:      3,
		MinStock:   5,
	}

	isLowStock := product.TrackStock && (product.Stock <= product.MinStock)
	if !isLowStock {
		t.Errorf("expected product to trigger low stock alert")
	}

	restockQty := 10.0
	newStock := product.Stock + restockQty
	if newStock != 13.0 {
		t.Errorf("expected new stock 13.0, got %.2f", newStock)
	}

	isStillLow := product.TrackStock && (newStock <= product.MinStock)
	if isStillLow {
		t.Errorf("product should no longer be low stock after restock")
	}
}
