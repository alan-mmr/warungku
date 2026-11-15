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

type CategoryService interface {
	Create(c *fiber.Ctx) error
	FindAll(c *fiber.Ctx) error
	FindByID(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
}

type categoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(categoryRepo repository.CategoryRepository) CategoryService {
	return &categoryService{categoryRepo: categoryRepo}
}

func (s *categoryService) Create(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	var req model.CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.TranslateError(c, helper.ErrBadRequest)
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return helper.TranslateError(c, errors.New("nama kategori wajib diisi"))
	}

	category := model.Category{
		BusinessID: businessID,
		Name:       req.Name,
	}

	if err := s.categoryRepo.Create(c.Context(), &category); err != nil {
		return helper.TranslateError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(helper.CreatedResponse(category))
}

func (s *categoryService) FindAll(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	categories, err := s.categoryRepo.FindAll(c.Context(), businessID)
	if err != nil {
		return helper.TranslateError(c, err)
	}

	if categories == nil {
		categories = []model.Category{}
	}

	return c.JSON(helper.SuccessResponse(categories))
}

func (s *categoryService) FindByID(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	id := c.Params("id")
	category, err := s.categoryRepo.FindByID(c.Context(), businessID, id)
	if err != nil {
		return helper.TranslateError(c, err)
	}

	return c.JSON(helper.SuccessResponse(category))
}

func (s *categoryService) Update(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	id := c.Params("id")
	var req model.UpdateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.TranslateError(c, helper.ErrBadRequest)
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return helper.TranslateError(c, errors.New("nama kategori wajib diisi"))
	}

	category := model.Category{
		ID:         id,
		BusinessID: businessID,
		Name:       req.Name,
	}

	if err := s.categoryRepo.Update(c.Context(), &category); err != nil {
		return helper.TranslateError(c, err)
	}

	return c.JSON(helper.SuccessResponse(category))
}

func (s *categoryService) Delete(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	id := c.Params("id")
	if err := s.categoryRepo.DeleteSoft(c.Context(), businessID, id); err != nil {
		return helper.TranslateError(c, err)
	}

	return c.JSON(helper.SuccessResponse(fiber.Map{
		"message": "kategori berhasil dinonaktifkan",
		"id":      id,
	}))
}
