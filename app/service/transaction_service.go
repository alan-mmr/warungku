package service

import (
	"errors"

	"warungku/app/model"
	"warungku/app/repository"
	"warungku/helper"
	"warungku/middleware"

	"github.com/gofiber/fiber/v2"
)

type TransactionService interface {
	Create(c *fiber.Ctx) error
	FindAll(c *fiber.Ctx) error
	FindByID(c *fiber.Ctx) error
	Void(c *fiber.Ctx) error
}

type transactionService struct {
	txRepo repository.TransactionRepository
}

func NewTransactionService(txRepo repository.TransactionRepository) TransactionService {
	return &transactionService{txRepo: txRepo}
}

func (s *transactionService) Create(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}
	userID := c.Locals(middleware.CtxUserID).(string)

	var req model.CreateTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.TranslateError(c, helper.ErrBadRequest)
	}

	if len(req.Items) == 0 {
		return helper.TranslateError(c, errors.New("daftar item transaksi tidak boleh kosong"))
	}

	if req.PaymentMethod == "" {
		req.PaymentMethod = "cash"
	}

	var items []model.TransactionItem
	for _, it := range req.Items {
		if it.Qty <= 0 {
			return helper.TranslateError(c, errors.New("kuantitas item harus lebih dari 0"))
		}
		items = append(items, model.TransactionItem{
			ProductID: it.ProductID,
			VariantID: it.VariantID,
			Qty:       it.Qty,
		})
	}

	tx := model.Transaction{
		BusinessID:    businessID,
		CashierID:     &userID,
		CustomerID:    req.CustomerID,
		Discount:      req.Discount,
		PaidAmount:    req.PaidAmount,
		PaymentMethod: req.PaymentMethod,
		Note:          req.Note,
		Items:         items,
	}

	if err := s.txRepo.CreateTransactionAtomic(c.Context(), &tx, req.DueDate); err != nil {
		return helper.TranslateError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(helper.CreatedResponse(tx))
}

func (s *transactionService) FindAll(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	status := c.Query("status")
	date := c.Query("date")

	transactions, err := s.txRepo.FindAll(c.Context(), businessID, status, date)
	if err != nil {
		return helper.TranslateError(c, err)
	}

	if transactions == nil {
		transactions = []model.Transaction{}
	}

	return c.JSON(helper.SuccessResponse(transactions))
}

func (s *transactionService) FindByID(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	id := c.Params("id")
	tx, err := s.txRepo.FindByID(c.Context(), businessID, id)
	if err != nil {
		return helper.TranslateError(c, err)
	}

	return c.JSON(helper.SuccessResponse(tx))
}

func (s *transactionService) Void(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}
	userID := c.Locals(middleware.CtxUserID).(string)

	id := c.Params("id")
	if err := s.txRepo.VoidTransaction(c.Context(), businessID, id, userID); err != nil {
		return helper.TranslateError(c, err)
	}

	return c.JSON(helper.SuccessResponse(fiber.Map{
		"message": "transaksi berhasil dibatalkan dan stok dikembalikan",
		"id":      id,
	}))
}
