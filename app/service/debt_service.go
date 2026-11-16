package service

import (
	"errors"

	"warungku/app/model"
	"warungku/app/repository"
	"warungku/helper"
	"warungku/middleware"

	"github.com/gofiber/fiber/v2"
)

type DebtService interface {
	FindAll(c *fiber.Ctx) error
	FindByID(c *fiber.Ctx) error
	Pay(c *fiber.Ctx) error
}

type debtService struct {
	debtRepo repository.DebtRepository
}

func NewDebtService(debtRepo repository.DebtRepository) DebtService {
	return &debtService{debtRepo: debtRepo}
}

func (s *debtService) FindAll(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	status := c.Query("status")
	debts, err := s.debtRepo.FindAll(c.Context(), businessID, status)
	if err != nil {
		return helper.TranslateError(c, err)
	}

	if debts == nil {
		debts = []model.Debt{}
	}

	return c.JSON(helper.SuccessResponse(debts))
}

func (s *debtService) FindByID(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	id := c.Params("id")
	debt, err := s.debtRepo.FindByID(c.Context(), businessID, id)
	if err != nil {
		return helper.TranslateError(c, err)
	}

	return c.JSON(helper.SuccessResponse(debt))
}

func (s *debtService) Pay(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}
	userID := c.Locals(middleware.CtxUserID).(string)

	debtID := c.Params("id")
	var req model.PayDebtRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.TranslateError(c, helper.ErrBadRequest)
	}

	if req.Amount <= 0 {
		return helper.TranslateError(c, errors.New("nominal pembayaran harus lebih dari 0"))
	}
	if req.Method == "" {
		req.Method = "cash"
	}

	payment := model.DebtPayment{
		Amount:     req.Amount,
		Method:     req.Method,
		Note:       req.Note,
		ReceivedBy: &userID,
	}

	if err := s.debtRepo.PayDebt(c.Context(), businessID, debtID, &payment); err != nil {
		return helper.TranslateError(c, err)
	}

	return c.JSON(helper.SuccessResponse(payment))
}
