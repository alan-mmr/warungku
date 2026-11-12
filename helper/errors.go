package helper

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

var (
	ErrNotFound       = errors.New("data tidak ditemukan")
	ErrUnauthorized   = errors.New("kredensial tidak valid atau belum terautentikasi")
	ErrForbidden      = errors.New("akses ditolak: tidak memiliki hak akses")
	ErrBadRequest     = errors.New("permintaan tidak valid")
	ErrConflict       = errors.New("terjadi konflik data atau data duplikat")
	ErrOutOfStock     = errors.New("stok produk tidak mencukupi")
	ErrInternalServer = errors.New("terjadi kesalahan internal pada server")
)

func TranslateError(c *fiber.Ctx, err error) error {
	var statusCode int
	var statusMsg string

	switch {
	case errors.Is(err, ErrNotFound):
		statusCode = fiber.StatusNotFound
		statusMsg = "NOT_FOUND"
	case errors.Is(err, ErrUnauthorized):
		statusCode = fiber.StatusUnauthorized
		statusMsg = "UNAUTHORIZED"
	case errors.Is(err, ErrForbidden):
		statusCode = fiber.StatusForbidden
		statusMsg = "FORBIDDEN"
	case errors.Is(err, ErrConflict):
		statusCode = fiber.StatusConflict
		statusMsg = "CONFLICT"
	case errors.Is(err, ErrOutOfStock), errors.Is(err, ErrBadRequest):
		statusCode = fiber.StatusBadRequest
		statusMsg = "BAD_REQUEST"
	default:
		statusCode = fiber.StatusInternalServerError
		statusMsg = "INTERNAL_SERVER_ERROR"
	}

	return c.Status(statusCode).JSON(WebResponse{
		Code:   statusCode,
		Status: statusMsg,
		Errors: err.Error(),
	})
}
