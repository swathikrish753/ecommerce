package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/swathikrish753/ecommerce/services/catalog/internal/domain"
	"github.com/swathikrish753/ecommerce/services/catalog/internal/service"
)

type CatalogHandler struct {
	svc *service.Catalog
}

func NewCatalogHandler(svc *service.Catalog) *CatalogHandler {
	return &CatalogHandler{svc: svc}
}

func (h *CatalogHandler) Register(e *echo.Echo) {
	e.POST("/products", h.create)
	e.GET("/products/:id", h.get)
	e.GET("/products", h.list)
}

type productIn struct {
	SKU         string            `json:"sku"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	PriceCents  int64             `json:"price_cents"`
	Attributes  map[string]string `json:"attributes"`
}

func (h *CatalogHandler) create(c echo.Context) error {
	var in productIn
	if err := c.Bind(&in); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON body")
	}
	p := &domain.Product{
		SKU: in.SKU, Name: in.Name, Description: in.Description,
		Category: in.Category, PriceCents: in.PriceCents, Attributes: in.Attributes,
	}
	if err := h.svc.Create(c.Request().Context(), p); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidProduct):
			return echo.NewHTTPError(http.StatusBadRequest, "sku, name required; price >= 0")
		case errors.Is(err, domain.ErrSKUTaken):
			return echo.NewHTTPError(http.StatusConflict, "sku already exists")
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, "could not create product")
		}
	}
	return c.JSON(http.StatusCreated, p)
}

func (h *CatalogHandler) get(c echo.Context) error {
	p, err := h.svc.Get(c.Request().Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "product not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "could not get product")
	}
	return c.JSON(http.StatusOK, p)
}

func (h *CatalogHandler) list(c echo.Context) error {
	category := c.QueryParam("category")
	after := c.QueryParam("after")
	limit, _ := strconv.ParseInt(c.QueryParam("limit"), 10, 64)

	items, err := h.svc.List(c.Request().Context(), category, after, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not list products")
	}

	var next string
	if n := len(items); n > 0 {
		next = items[n-1].ID
	}
	return c.JSON(http.StatusOK, map[string]any{
		"items":      items,
		"next_after": next,
	})
}
