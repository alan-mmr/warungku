package model

type DailyReportResponse struct {
	Date             string  `json:"date"`
	TotalGrossSales  float64 `json:"total_gross_sales"`
	TotalTransactions int     `json:"total_transactions"`
	GrossProfit      float64 `json:"gross_profit"`
	TotalExpenses    float64 `json:"total_expenses"`
	NetProfit        float64 `json:"net_profit"`
}

type MonthlyReportResponse struct {
	Month            string  `json:"month"`
	TotalGrossSales  float64 `json:"total_gross_sales"`
	TotalTransactions int     `json:"total_transactions"`
	GrossProfit      float64 `json:"gross_profit"`
	TotalExpenses    float64 `json:"total_expenses"`
	NetProfit        float64 `json:"net_profit"`
	TotalUnpaidDebt  float64 `json:"total_unpaid_debt"`
}

type TopProductItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	TotalQty    float64 `json:"total_qty"`
	TotalSales  float64 `json:"total_sales"`
}
