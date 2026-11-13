package service

import (
	"errors"
	"time"

	"warungku/app/model"
	"warungku/app/repository"
	"warungku/helper"
	"warungku/middleware"

	"github.com/gofiber/fiber/v2"
)

type ExpenseService interface {
	Create(c *fiber.Ctx) error
	FindAll(c *fiber.Ctx) error
	FindByID(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
}

type expenseService struct {
	expenseRepo repository.ExpenseRepository
}

func NewExpenseService(expenseRepo repository.ExpenseRepository) ExpenseService {
	return &expenseService{expenseRepo: expenseRepo}
}

func (s *expenseService) Create(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}
	userID := c.Locals(middleware.CtxUserID).(string)

	var req model.CreateExpenseRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.TranslateError(c, helper.ErrBadRequest)
	}

	if req.Amount <= 0 {
		return helper.TranslateError(c, errors.New("nominal pengeluaran harus lebih dari 0"))
	}
	if req.Category == "" {
		req.Category = "lainnya"
	}

	spentTime := time.Now()
	if req.SpentAt != nil && *req.SpentAt != "" {
		parsed, err := time.Parse("2006-01-02", *req.SpentAt)
		if err == nil {
			spentTime = parsed
		}
	}

	expense := model.Expense{
		BusinessID: businessID,
		Category:   req.Category,
		Amount:     req.Amount,
		Note:       req.Note,
		SpentAt:    spentTime,
		CreatedBy:  &userID,
	}

	if err := s.expenseRepo.Create(c.Context(), &expense); err != nil {
		return helper.TranslateError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(helper.CreatedResponse(expense))
}

func (s *expenseService) FindAll(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	category := c.Query("category")
	date := c.Query("date")

	expenses, err := s.expenseRepo.FindAll(c.Context(), businessID, category, date)
	if err != nil {
		return helper.TranslateError(c, err)
	}

	if expenses == nil {
		expenses = []model.Expense{}
	}

	return c.JSON(helper.SuccessResponse(expenses))
}

func (s *expenseService) FindByID(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	id := c.Params("id")
	expense, err := s.expenseRepo.FindByID(c.Context(), businessID, id)
	if err != nil {
		return helper.TranslateError(c, err)
	}

	return c.JSON(helper.SuccessResponse(expense))
}

func (s *expenseService) Update(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	id := c.Params("id")
	var req model.CreateExpenseRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.TranslateError(c, helper.ErrBadRequest)
	}

	if req.Amount <= 0 {
		return helper.TranslateError(c, errors.New("nominal pengeluaran harus lebih dari 0"))
	}

	spentTime := time.Now()
	if req.SpentAt != nil && *req.SpentAt != "" {
		parsed, err := time.Parse("2006-01-02", *req.SpentAt)
		if err == nil {
			spentTime = parsed
		}
	}

	expense := model.Expense{
		ID:         id,
		BusinessID: businessID,
		Category:   req.Category,
		Amount:     req.Amount,
		Note:       req.Note,
		SpentAt:    spentTime,
	}

	if err := s.expenseRepo.Update(c.Context(), &expense); err != nil {
		return helper.TranslateError(c, err)
	}

	return c.JSON(helper.SuccessResponse(expense))
}

func (s *expenseService) Delete(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	id := c.Params("id")
	if err := s.expenseRepo.Delete(c.Context(), businessID, id); err != nil {
		return helper.TranslateError(c, err)
	}

	return c.JSON(helper.SuccessResponse(fiber.Map{
		"message": "pengeluaran berhasil dihapus",
		"id":      id,
	}))
}
