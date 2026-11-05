package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/swathikrish753/ecommerce/services/cart/internal/domain"
	"github.com/swathikrish753/ecommerce/services/cart/internal/service"
)

// userHeader is set by the API Gateway (Day 6) after it verifies the JWT.
const userHeader = "X-User-ID"

type CartHandler struct {
	svc *service.Cart
}

func NewCartHandler(svc *service.Cart) *CartHandler {
	return &CartHandler{svc: svc}
}

func (h *CartHandler) Register(e *echo.Echo) {
	e.GET("/cart", h.get)
	e.DELETE("/cart", h.clear)
	e.POST("/cart/items", h.addItem)
	e.PATCH("/cart/items/:sku", h.setQuantity)
	e.DELETE("/cart/items/:sku", h.removeItem)
}

type addItemIn struct {
	SKU      string `json:"sku"`
	Quantity int64  `json:"quantity"`
}

type setQuantityIn struct {
	Quantity int64 `json:"quantity"`
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, service.ErrNoUser):
		return echo.NewHTTPError(http.StatusUnauthorized, "missing X-User-ID")
	case errors.Is(err, service.ErrInvalidItem):
		return echo.NewHTTPError(http.StatusBadRequest, "sku required; quantity 1-99")
	case errors.Is(err, domain.ErrItemNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "item not in cart")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "cart error")
	}
}

func (h *CartHandler) get(c echo.Context) error {
	cart, err := h.svc.Get(c.Request().Context(), c.Request().Header.Get(userHeader))
	if err != nil {
		return mapErr(err)
	}
	return c.JSON(http.StatusOK, cart)
}

func (h *CartHandler) addItem(c echo.Context) error {
	var in addItemIn
	if err := c.Bind(&in); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON body")
	}
	qty, err := h.svc.AddItem(c.Request().Context(), c.Request().Header.Get(userHeader), in.SKU, in.Quantity)
	if err != nil {
		return mapErr(err)
	}
	return c.JSON(http.StatusOK, domain.CartItem{SKU: in.SKU, Quantity: qty})
}

func (h *CartHandler) setQuantity(c echo.Context) error {
	var in setQuantityIn
	if err := c.Bind(&in); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON body")
	}
	sku := c.Param("sku")
	if err := h.svc.SetQuantity(c.Request().Context(), c.Request().Header.Get(userHeader), sku, in.Quantity); err != nil {
		return mapErr(err)
	}
	return c.JSON(http.StatusOK, domain.CartItem{SKU: sku, Quantity: in.Quantity})
}

func (h *CartHandler) removeItem(c echo.Context) error {
	if err := h.svc.RemoveItem(c.Request().Context(), c.Request().Header.Get(userHeader), c.Param("sku")); err != nil {
		return mapErr(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *CartHandler) clear(c echo.Context) error {
	if err := h.svc.Clear(c.Request().Context(), c.Request().Header.Get(userHeader)); err != nil {
		return mapErr(err)
	}
	return c.NoContent(http.StatusNoContent)
}
