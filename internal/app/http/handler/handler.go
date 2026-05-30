package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/marketplace/marketplace-bucket/internal/core/domain"
	"github.com/marketplace/marketplace-bucket/internal/core/port"
	"github.com/marketplace/marketplace-bucket/internal/platform/logger"
	"github.com/marketplace/marketplace-bucket/internal/platform/metrics"
)

type Handler struct {
	svc     port.CartService
	metrics *metrics.Metrics
	log     *slog.Logger
}

func New(svc port.CartService, m *metrics.Metrics, log *slog.Logger) *Handler {
	return &Handler{svc: svc, metrics: m, log: log}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/cart/{userID}/items", h.addItem)
	mux.HandleFunc("DELETE /api/v1/cart/{userID}/items/{productID}", h.removeItem)
	mux.HandleFunc("PATCH /api/v1/cart/{userID}/items/{productID}", h.updateQuantity)
	mux.HandleFunc("GET /api/v1/cart/{userID}", h.getCart)
	mux.HandleFunc("DELETE /api/v1/cart/{userID}", h.clearCart)
	mux.HandleFunc("POST /api/v1/cart/{userID}/merge", h.mergeCart)
}

type addItemBody struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

type updateQuantityBody struct {
	Quantity int `json:"quantity"`
}

func (h *Handler) addItem(w http.ResponseWriter, r *http.Request) {
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

	cart, err := h.svc.AddItem(r.Context(), userID, item)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	h.metrics.CartOperationsTotal.WithLabelValues("add_item").Inc()
	writeJSON(w, http.StatusOK, cart)
}

func (h *Handler) removeItem(w http.ResponseWriter, r *http.Request) {
	cart, err := h.svc.RemoveItem(r.Context(), r.PathValue("userID"), r.PathValue("productID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}

	h.metrics.CartOperationsTotal.WithLabelValues("remove_item").Inc()
	writeJSON(w, http.StatusOK, cart)
}

func (h *Handler) updateQuantity(w http.ResponseWriter, r *http.Request) {
	var body updateQuantityBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cart, err := h.svc.UpdateQuantity(r.Context(), r.PathValue("userID"), r.PathValue("productID"), body.Quantity)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	h.metrics.CartOperationsTotal.WithLabelValues("update_quantity").Inc()
	writeJSON(w, http.StatusOK, cart)
}

func (h *Handler) getCart(w http.ResponseWriter, r *http.Request) {
	cart, err := h.svc.GetCart(r.Context(), r.PathValue("userID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}

	logger.FromContext(r.Context(), h.log).Debug("cart fetched", "user_id", r.PathValue("userID"))
	writeJSON(w, http.StatusOK, cart)
}

func (h *Handler) clearCart(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.ClearCart(r.Context(), r.PathValue("userID")); err != nil {
		writeServiceError(w, err)
		return
	}

	h.metrics.CartOperationsTotal.WithLabelValues("clear_cart").Inc()
	w.WriteHeader(http.StatusNoContent)
}

type mergeCartBody struct {
	GuestUserID string `json:"guest_user_id"`
}

func (h *Handler) mergeCart(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")

	var body mergeCartBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.GuestUserID == "" {
		writeError(w, http.StatusBadRequest, "guest_user_id is required")
		return
	}
	if body.GuestUserID == userID {
		writeError(w, http.StatusBadRequest, "guest_user_id must differ from user id")
		return
	}

	cart, err := h.svc.MergeCart(r.Context(), body.GuestUserID, userID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	h.metrics.CartOperationsTotal.WithLabelValues("merge_cart").Inc()
	writeJSON(w, http.StatusOK, cart)
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
