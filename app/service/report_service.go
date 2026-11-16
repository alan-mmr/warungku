package service

import (
	"warungku/app/repository"
	"warungku/helper"
	"warungku/middleware"

	"github.com/gofiber/fiber/v2"
)

type ReportService interface {
	DailyReport(c *fiber.Ctx) error
	MonthlyReport(c *fiber.Ctx) error
	TopProducts(c *fiber.Ctx) error
}

type reportService struct {
	reportRepo repository.ReportRepository
}

func NewReportService(reportRepo repository.ReportRepository) ReportService {
	return &reportService{reportRepo: reportRepo}
}

func (s *reportService) DailyReport(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	date := c.Query("date")
	report, err := s.reportRepo.GetDailyReport(c.Context(), businessID, date)
	if err != nil {
		return helper.TranslateError(c, err)
	}

	return c.JSON(helper.SuccessResponse(report))
}

func (s *reportService) MonthlyReport(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	month := c.Query("month")
	report, err := s.reportRepo.GetMonthlyReport(c.Context(), businessID, month)
	if err != nil {
		return helper.TranslateError(c, err)
	}

	return c.JSON(helper.SuccessResponse(report))
}

func (s *reportService) TopProducts(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	list, err := s.reportRepo.GetTopProducts(c.Context(), businessID, 5)
	if err != nil {
		return helper.TranslateError(c, err)
	}

	return c.JSON(helper.SuccessResponse(list))
}
