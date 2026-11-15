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

type ProductService interface {
	Create(c *fiber.Ctx) error
	FindAll(c *fiber.Ctx) error
	FindByID(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
	CreateVariant(c *fiber.Ctx) error
}

type productService struct {
	productRepo repository.ProductRepository
}

func NewProductService(productRepo repository.ProductRepository) ProductService {
	return &productService{productRepo: productRepo}
}

func (s *productService) Create(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	var req model.CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.TranslateError(c, helper.ErrBadRequest)
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return helper.TranslateError(c, errors.New("nama produk wajib diisi"))
	}
	if req.Price < 0 || req.CostPrice < 0 {
		return helper.TranslateError(c, errors.New("harga jual dan modal tidak boleh negatif"))
	}
	if req.Unit == "" {
		req.Unit = "pcs"
	}

	product := model.Product{
		BusinessID: businessID,
		CategoryID: req.CategoryID,
		Name:       req.Name,
		SKU:        req.SKU,
		Unit:       req.Unit,
		Price:      req.Price,
		CostPrice:  req.CostPrice,
		TrackStock: req.TrackStock,
		Stock:      req.Stock,
		MinStock:   req.MinStock,
	}

	if err := s.productRepo.Create(c.Context(), &product); err != nil {
		return helper.TranslateError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(helper.CreatedResponse(product))
}

func (s *productService) FindAll(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	categoryID := c.Query("category_id")
	search := c.Query("search")

	products, err := s.productRepo.FindAll(c.Context(), businessID, categoryID, search)
	if err != nil {
		return helper.TranslateError(c, err)
	}

	if products == nil {
		products = []model.Product{}
	}

	return c.JSON(helper.SuccessResponse(products))
}

func (s *productService) FindByID(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	id := c.Params("id")
	product, err := s.productRepo.FindByID(c.Context(), businessID, id)
	if err != nil {
		return helper.TranslateError(c, err)
	}

	return c.JSON(helper.SuccessResponse(product))
}

func (s *productService) Update(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	id := c.Params("id")
	var req model.UpdateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.TranslateError(c, helper.ErrBadRequest)
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return helper.TranslateError(c, errors.New("nama produk wajib diisi"))
	}
	if req.Price < 0 || req.CostPrice < 0 {
		return helper.TranslateError(c, errors.New("harga jual dan modal tidak boleh negatif"))
	}

	product := model.Product{
		ID:         id,
		BusinessID: businessID,
		CategoryID: req.CategoryID,
		Name:       req.Name,
		SKU:        req.SKU,
		Unit:       req.Unit,
		Price:      req.Price,
		CostPrice:  req.CostPrice,
		TrackStock: req.TrackStock,
		MinStock:   req.MinStock,
	}

	if err := s.productRepo.Update(c.Context(), &product); err != nil {
		return helper.TranslateError(c, err)
	}

	return c.JSON(helper.SuccessResponse(product))
}

func (s *productService) Delete(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	id := c.Params("id")
	if err := s.productRepo.DeleteSoft(c.Context(), businessID, id); err != nil {
		return helper.TranslateError(c, err)
	}

	return c.JSON(helper.SuccessResponse(fiber.Map{
		"message": "produk berhasil dinonaktifkan",
		"id":      id,
	}))
}

func (s *productService) CreateVariant(c *fiber.Ctx) error {
	businessID, ok := c.Locals(middleware.CtxBusinessID).(string)
	if !ok || businessID == "" {
		return helper.TranslateError(c, helper.ErrForbidden)
	}

	productID := c.Params("id")
	var req model.CreateVariantRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.TranslateError(c, helper.ErrBadRequest)
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return helper.TranslateError(c, errors.New("nama varian wajib diisi"))
	}

	variant := model.ProductVariant{
		ProductID:  productID,
		Name:       req.Name,
		ExtraPrice: req.ExtraPrice,
	}

	if err := s.productRepo.CreateVariant(c.Context(), businessID, &variant); err != nil {
		return helper.TranslateError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(helper.CreatedResponse(variant))
}
