package service

import (
	"errors"

	"warungku/app/model"
	"warungku/app/repository"
	"warungku/helper"
	"warungku/middleware"

	"github.com/gofiber/fiber/v2"
)

type StockService interface {
	StockIn(c *fiber.Ctx) error
	StockAdjust(c *fiber.Ctx) error
	GetMovements(c *fiber.Ctx) error
	GetLowStock(c *fiber.Ctx) error
}

type stockService struct {
	stockRepo repository.StockRepository
}

func NewStockService(stockRepo repository.StockRepository) StockService {
	return &stockService{stockRepo: stockRepo}
}

func (s *stockService) StockIn(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}
	userID := c.Locals(middleware.CtxUserID).(string)

	var req model.StockInRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.TranslateError(c, helper.ErrBadRequest)
	}

	if req.ProductID == "" || req.Qty <= 0 {
		return helper.TranslateError(c, errors.New("product_id dan kuantitas valid (> 0) wajib diisi"))
	}

	if err := s.stockRepo.StockIn(c.Context(), businessID, req, userID); err != nil {
		return helper.TranslateError(c, err)
	}

	return c.JSON(helper.SuccessResponse(fiber.Map{
		"message":    "stok masuk berhasil dicatat",
		"product_id": req.ProductID,
		"qty":        req.Qty,
	}))
}

func (s *stockService) StockAdjust(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}
	userID := c.Locals(middleware.CtxUserID).(string)

	var req model.StockAdjustRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.TranslateError(c, helper.ErrBadRequest)
	}

	if req.ProductID == "" || req.ActualStock < 0 {
		return helper.TranslateError(c, errors.New("product_id dan actual_stock (>= 0) wajib diisi"))
	}

	if err := s.stockRepo.StockAdjust(c.Context(), businessID, req, userID); err != nil {
		return helper.TranslateError(c, err)
	}

	return c.JSON(helper.SuccessResponse(fiber.Map{
		"message":      "penyesuaian stok opname berhasil",
		"product_id":   req.ProductID,
		"actual_stock": req.ActualStock,
	}))
}

func (s *stockService) GetMovements(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	productID := c.Query("product_id")
	movements, err := s.stockRepo.GetMovements(c.Context(), businessID, productID)
	if err != nil {
		return helper.TranslateError(c, err)
	}

	if movements == nil {
		movements = []model.StockMovement{}
	}

	return c.JSON(helper.SuccessResponse(movements))
}

func (s *stockService) GetLowStock(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	lowStocks, err := s.stockRepo.GetLowStock(c.Context(), businessID)
	if err != nil {
		return helper.TranslateError(c, err)
	}

	if lowStocks == nil {
		lowStocks = []model.Product{}
	}

	return c.JSON(helper.SuccessResponse(lowStocks))
}
