package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/swathikrish753/ecommerce/services/auth/internal/domain"
	"github.com/swathikrish753/ecommerce/services/auth/internal/service"
)

type AuthHandler struct {
	svc *service.Auth
}

func NewAuthHandler(svc *service.Auth) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(e *echo.Echo) {
	e.POST("/register", h.register)
	e.POST("/login", h.login)
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) register(c echo.Context) error {
	var in credentials
	if err := c.Bind(&in); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON body")
	}
	if in.Email == "" || len(in.Password) < 8 {
		return echo.NewHTTPError(http.StatusBadRequest,
			"email required and password must be >= 8 chars")
	}

	u, err := h.svc.Register(c.Request().Context(), in.Email, in.Password)
	if err != nil {
		if errors.Is(err, domain.ErrEmailTaken) {
			return echo.NewHTTPError(http.StatusConflict, "email already registered")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "could not register")
	}
	return c.JSON(http.StatusCreated, map[string]string{"id": u.ID, "email": u.Email})
}

func (h *AuthHandler) login(c echo.Context) error {
	var in credentials
	if err := c.Bind(&in); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON body")
	}

	token, err := h.svc.Login(c.Request().Context(), in.Email, in.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "could not log in")
	}
	return c.JSON(http.StatusOK, map[string]string{"token": token})
}
