// Package http provides the REST transport layer for the cart service.
package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/marketplace/cart-service/internal/domain"
	"github.com/marketplace/cart-service/internal/service"
	"github.com/marketplace/cart-service/pkg/metrics"
)

// Handler holds dependencies for HTTP request handling.
type Handler struct {
	svc     service.CartService
	metrics *metrics.Metrics
	tracer  trace.Tracer
}

// NewHandler creates a new HTTP handler.
func NewHandler(svc service.CartService, m *metrics.Metrics) *Handler {
	return &Handler{
		svc:     svc,
		metrics: m,
		tracer:  otel.Tracer("cart/handler/http"),
	}
}

// RegisterRoutes registers all cart API routes on mux.
// Metrics middleware is applied to all routes.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/cart/{userID}/items", h.metricsMiddleware(http.HandlerFunc(h.addItem)))
	mux.Handle("DELETE /api/v1/cart/{userID}/items/{productID}", h.metricsMiddleware(http.HandlerFunc(h.removeItem)))
	mux.Handle("PATCH /api/v1/cart/{userID}/items/{productID}", h.metricsMiddleware(http.HandlerFunc(h.updateQuantity)))
	mux.Handle("GET /api/v1/cart/{userID}", h.metricsMiddleware(http.HandlerFunc(h.getCart)))
	mux.Handle("DELETE /api/v1/cart/{userID}", h.metricsMiddleware(http.HandlerFunc(h.clearCart)))
	mux.HandleFunc("GET /health", h.health)
}

// addItemBody is the JSON body for POST /cart/{userID}/items.
type addItemBody struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

// updateQuantityBody is the JSON body for PATCH /cart/{userID}/items/{productID}.
type updateQuantityBody struct {
	Quantity int `json:"quantity"`
}

func (h *Handler) addItem(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "http.AddItem")
	defer span.End()

	userID := r.PathValue("userID")

	var body addItemBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item := &domain.Item{
		ProductID: body.ProductID,
		Name:      body.Name,
		Price:     body.Price,
		Quantity:  body.Quantity,
	}

	cart, err := h.svc.AddItem(ctx, userID, item)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cart)
}

func (h *Handler) removeItem(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "http.RemoveItem")
	defer span.End()

	cart, err := h.svc.RemoveItem(ctx, r.PathValue("userID"), r.PathValue("productID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cart)
}

func (h *Handler) updateQuantity(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "http.UpdateQuantity")
	defer span.End()

	var body updateQuantityBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cart, err := h.svc.UpdateQuantity(ctx, r.PathValue("userID"), r.PathValue("productID"), body.Quantity)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cart)
}

func (h *Handler) getCart(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "http.GetCart")
	defer span.End()

	cart, err := h.svc.GetCart(ctx, r.PathValue("userID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cart)
}

func (h *Handler) clearCart(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "http.ClearCart")
	defer span.End()

	if err := h.svc.ClearCart(ctx, r.PathValue("userID")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrCartNotFound), errors.Is(err, domain.ErrItemNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidQuantity),
		errors.Is(err, domain.ErrInvalidUserID),
		errors.Is(err, domain.ErrInvalidProductID):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
