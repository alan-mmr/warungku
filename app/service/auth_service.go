package service

import (
	"errors"
	"strings"

	"warungku/app/model"
	"warungku/app/repository"
	"warungku/helper"
	"warungku/middleware"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	RegisterOwner(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	GetMe(c *fiber.Ctx) error
}

type authService struct {
	userRepo   repository.UserRepository
	jwtSecret  string
	expMinutes int
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string, expMinutes int) AuthService {
	return &authService{
		userRepo:   userRepo,
		jwtSecret:  jwtSecret,
		expMinutes: expMinutes,
	}
}

func (s *authService) RegisterOwner(c *fiber.Ctx) error {
	var req model.RegisterOwnerRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.TranslateError(c, helper.ErrBadRequest)
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" || req.Name == "" || req.BusinessName == "" {
		return helper.TranslateError(c, errors.New("nama, email, password, dan nama usaha wajib diisi"))
	}

	// Cek apakah email sudah terdaftar
	existing, err := s.userRepo.FindByEmail(c.Context(), req.Email)
	if err == nil && existing != nil {
		return helper.TranslateError(c, helper.ErrConflict)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return helper.TranslateError(c, helper.ErrInternalServer)
	}

	user := model.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         "owner",
		IsActive:     true,
	}

	business := model.Business{
		Name:         req.BusinessName,
		BusinessType: req.BusinessType,
		Address:      req.Address,
		Phone:        req.Phone,
		IsActive:     true,
	}

	if err := s.userRepo.CreateUserWithBusiness(c.Context(), &user, &business); err != nil {
		return helper.TranslateError(c, err)
	}

	// Generate JWT Access Token
	token, err := helper.GenerateAccessToken(helper.JWTPayload{
		UserID:     user.ID,
		BusinessID: user.BusinessID,
		Role:       user.Role,
	}, s.jwtSecret, s.expMinutes)
	if err != nil {
		return helper.TranslateError(c, helper.ErrInternalServer)
	}

	return c.Status(fiber.StatusCreated).JSON(helper.CreatedResponse(model.LoginResponse{
		Token: token,
		User:  user,
	}))
}

func (s *authService) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.TranslateError(c, helper.ErrBadRequest)
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	user, err := s.userRepo.FindByEmail(c.Context(), req.Email)
	if err != nil {
		return helper.TranslateError(c, helper.ErrUnauthorized)
	}

	if !user.IsActive {
		return helper.TranslateError(c, errors.New("akun dinonaktifkan"))
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return helper.TranslateError(c, helper.ErrUnauthorized)
	}

	token, err := helper.GenerateAccessToken(helper.JWTPayload{
		UserID:     user.ID,
		BusinessID: user.BusinessID,
		Role:       user.Role,
	}, s.jwtSecret, s.expMinutes)
	if err != nil {
		return helper.TranslateError(c, helper.ErrInternalServer)
	}

	return c.JSON(helper.SuccessResponse(model.LoginResponse{
		Token: token,
		User:  *user,
	}))
}

func (s *authService) GetMe(c *fiber.Ctx) error {
	userID, ok := c.Locals(middleware.CtxUserID).(string)
	if !ok || userID == "" {
		return helper.TranslateError(c, helper.ErrUnauthorized)
	}

	user, err := s.userRepo.FindByID(c.Context(), userID)
	if err != nil {
		return helper.TranslateError(c, err)
	}

	return c.JSON(helper.SuccessResponse(user))
}
