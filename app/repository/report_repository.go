package repository

import (
	"context"
	"time"

	"warungku/app/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ReportRepository interface {
	GetDailyReport(ctx context.Context, businessID, date string) (*model.DailyReportResponse, error)
	GetMonthlyReport(ctx context.Context, businessID, month string) (*model.MonthlyReportResponse, error)
	GetTopProducts(ctx context.Context, businessID string, limit int) ([]model.TopProductItem, error)
}

type reportRepository struct {
	pool *pgxpool.Pool
}

func NewReportRepository(pool *pgxpool.Pool) ReportRepository {
	return &reportRepository{pool: pool}
}

func (r *reportRepository) GetDailyReport(ctx context.Context, businessID, date string) (*model.DailyReportResponse, error) {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	report := model.DailyReportResponse{Date: date}

	// 1. Total Sales & Transactions Count
	salesQuery := `
		SELECT COALESCE(SUM(total), 0), COUNT(id)
		FROM transactions
		WHERE business_id = $1 AND status != 'void' AND DATE(created_at) = $2
	`
	_ = r.pool.QueryRow(ctx, salesQuery, businessID, date).Scan(&report.TotalGrossSales, &report.TotalTransactions)

	// 2. Gross Profit
	profitQuery := `
		SELECT COALESCE(SUM((ti.price - ti.cost_price) * ti.qty), 0)
		FROM transaction_items ti
		JOIN transactions t ON ti.transaction_id = t.id
		WHERE t.business_id = $1 AND t.status != 'void' AND DATE(t.created_at) = $2
	`
	_ = r.pool.QueryRow(ctx, profitQuery, businessID, date).Scan(&report.GrossProfit)

	// 3. Expenses
	expenseQuery := `
		SELECT COALESCE(SUM(amount), 0)
		FROM expenses
		WHERE business_id = $1 AND DATE(spent_at) = $2
	`
	_ = r.pool.QueryRow(ctx, expenseQuery, businessID, date).Scan(&report.TotalExpenses)

	report.NetProfit = report.GrossProfit - report.TotalExpenses
	return &report, nil
}

func (r *reportRepository) GetMonthlyReport(ctx context.Context, businessID, month string) (*model.MonthlyReportResponse, error) {
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	report := model.MonthlyReportResponse{Month: month}

	// 1. Total Sales & Transactions
	salesQuery := `
		SELECT COALESCE(SUM(total), 0), COUNT(id)
		FROM transactions
		WHERE business_id = $1 AND status != 'void' AND TO_CHAR(created_at, 'YYYY-MM') = $2
	`
	_ = r.pool.QueryRow(ctx, salesQuery, businessID, month).Scan(&report.TotalGrossSales, &report.TotalTransactions)

	// 2. Gross Profit
	profitQuery := `
		SELECT COALESCE(SUM((ti.price - ti.cost_price) * ti.qty), 0)
		FROM transaction_items ti
		JOIN transactions t ON ti.transaction_id = t.id
		WHERE t.business_id = $1 AND t.status != 'void' AND TO_CHAR(t.created_at, 'YYYY-MM') = $2
	`
	_ = r.pool.QueryRow(ctx, profitQuery, businessID, month).Scan(&report.GrossProfit)

	// 3. Expenses
	expenseQuery := `
		SELECT COALESCE(SUM(amount), 0)
		FROM expenses
		WHERE business_id = $1 AND TO_CHAR(spent_at, 'YYYY-MM') = $2
	`
	_ = r.pool.QueryRow(ctx, expenseQuery, businessID, month).Scan(&report.TotalExpenses)

	// 4. Total Unpaid Debt
	debtQuery := `
		SELECT COALESCE(SUM(remaining), 0)
		FROM debts
		WHERE business_id = $1 AND status = 'open'
	`
	_ = r.pool.QueryRow(ctx, debtQuery, businessID).Scan(&report.TotalUnpaidDebt)

	report.NetProfit = report.GrossProfit - report.TotalExpenses
	return &report, nil
}

func (r *reportRepository) GetTopProducts(ctx context.Context, businessID string, limit int) ([]model.TopProductItem, error) {
	if limit <= 0 {
		limit = 5
	}

	query := `
		SELECT COALESCE(ti.product_id::text, ''), ti.product_name, SUM(ti.qty), SUM(ti.subtotal)
		FROM transaction_items ti
		JOIN transactions t ON ti.transaction_id = t.id
		WHERE t.business_id = $1 AND t.status != 'void'
		GROUP BY ti.product_id, ti.product_name
		ORDER BY SUM(ti.qty) DESC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, businessID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.TopProductItem
	for rows.Next() {
		var item model.TopProductItem
		if err := rows.Scan(&item.ProductID, &item.ProductName, &item.TotalQty, &item.TotalSales); err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	if list == nil {
		list = []model.TopProductItem{}
	}

	return list, nil
}
