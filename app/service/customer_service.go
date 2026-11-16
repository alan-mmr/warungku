package service

import (
	"errors"
	"strings"

	"warungku/app/model"
	"warungku/app/repository"
	"warungku/helper"
	"warungku/middleware"

	"github.com/gofiber/fiber/v2"
)

type CustomerService interface {
	Create(c *fiber.Ctx) error
	FindAll(c *fiber.Ctx) error
	FindByID(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
}

type customerService struct {
	customerRepo repository.CustomerRepository
}

func NewCustomerService(customerRepo repository.CustomerRepository) CustomerService {
	return &customerService{customerRepo: customerRepo}
}

func (s *customerService) Create(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	var req model.CreateCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.TranslateError(c, helper.ErrBadRequest)
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return helper.TranslateError(c, errors.New("nama pelanggan wajib diisi"))
	}

	customer := model.Customer{
		BusinessID: businessID,
		Name:       req.Name,
		Phone:      req.Phone,
		Note:       req.Note,
	}

	if err := s.customerRepo.Create(c.Context(), &customer); err != nil {
		return helper.TranslateError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(helper.CreatedResponse(customer))
}

func (s *customerService) FindAll(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	customers, err := s.customerRepo.FindAll(c.Context(), businessID)
	if err != nil {
		return helper.TranslateError(c, err)
	}

	if customers == nil {
		customers = []model.Customer{}
	}

	return c.JSON(helper.SuccessResponse(customers))
}

func (s *customerService) FindByID(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	id := c.Params("id")
	customer, err := s.customerRepo.FindByID(c.Context(), businessID, id)
	if err != nil {
		return helper.TranslateError(c, err)
	}

	return c.JSON(helper.SuccessResponse(customer))
}

func (s *customerService) Update(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	id := c.Params("id")
	var req model.UpdateCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.TranslateError(c, helper.ErrBadRequest)
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return helper.TranslateError(c, errors.New("nama pelanggan wajib diisi"))
	}

	customer := model.Customer{
		ID:         id,
		BusinessID: businessID,
		Name:       req.Name,
		Phone:      req.Phone,
		Note:       req.Note,
	}

	if err := s.customerRepo.Update(c.Context(), &customer); err != nil {
		return helper.TranslateError(c, err)
	}

	return c.JSON(helper.SuccessResponse(customer))
}
