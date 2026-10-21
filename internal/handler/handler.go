package handler

import (
	"Confetki/internal/middleware"
	"Confetki/internal/model"
	"Confetki/internal/repository"
	"Confetki/internal/service"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type Handler struct {
	productService     *service.ProductService
	masterClassService *service.MasterClassService
	orderService       *service.OrderService
}

func NewHandler(productService *service.ProductService, masterClassService *service.MasterClassService, orderService *service.OrderService) *Handler {
	return &Handler{
		productService:     productService,
		masterClassService: masterClassService,
		orderService:       orderService,
	}
}

func Home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, " server is running")
}

func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.productService.GetAllProducts(r.Context())
	if err != nil {
		http.Error(w, "failed to get products", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)

}

func (h *Handler) GetProductById(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}
	product, err := h.productService.GetProductByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "product not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get product", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)

}

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product model.Product
	err := json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	createdProduct, err := h.productService.CreateProduct(r.Context(), product)
	if err != nil {
		http.Error(w, "failed to create product", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdProduct)
}

func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}
	err = h.productService.DeleteProduct(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			http.Error(w, "product not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete product", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CreateMasterClass(w http.ResponseWriter, r *http.Request) {
	var masterClass model.MasterClass
	err := json.NewDecoder(r.Body).Decode(&masterClass)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	createdMasterClass, err := h.masterClassService.CreateMasterClass(r.Context(), masterClass)
	if err != nil {
		http.Error(w, "failed to create masteClass", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application'json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdMasterClass)
}

func (h *Handler) GetMasterClass(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	masterClass, err := h.masterClassService.GetMasterClassByProductID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			http.Error(w, "masterClass not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get masterClass", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(masterClass)
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var request model.CreateOrderRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	request.UserID = userID

	createdOrder, err := h.orderService.CreateOrder(r.Context(), request)
	if err != nil {
		var valErr *service.ValidationError

		if errors.As(err, &valErr) {
			http.Error(w, valErr.Message, http.StatusInternalServerError)
			return
		}

		if errors.Is(err, repository.ErrNotEnoughProduct) {
			http.Error(w, "not enough product", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrProductNotFound) {
			http.Error(w, "product not found", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrInvalidQuantity) {
			http.Error(w, "quantity <0 ", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrEmptyOrderItems) {
			http.Error(w, "empty order items", http.StatusBadRequest)
			return
		}

		http.Error(w, "failed to create order", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdOrder)
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userRole, ok := r.Context().Value(middleware.RoleKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idString := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	order, err := h.orderService.GetOrderByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrInvalidID) {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrOrderNotFound) {
			http.Error(w, "order not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get order", http.StatusInternalServerError)
		return
	}
	if userRole != "admin" && order.UserID != userID {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (h *Handler) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userRole, ok := r.Context().Value(middleware.RoleKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	orders := make([]model.Order, 0)
	var err error

	if userRole == "admin" {
		orders, err = h.orderService.GetAllOrders(r.Context())
		if err != nil {
			http.Error(w, "failed to get orders", http.StatusInternalServerError)
			return
		}
	} else {
		orders, err = h.orderService.GetOrderByUserID(r.Context(), userID)
		if err != nil {
			http.Error(w, "failed to get orders", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req model.UpdateOrderStatusRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err = h.orderService.UpdateStatus(r.Context(), id, req.Status)
	if err != nil {
		if errors.Is(err, service.ErrBadStatus) {
			http.Error(w, "invalid status", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrInvalidID) {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrOrderNotFound) {
			http.Error(w, "order not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, repository.ErrCannotCancelOrder) {
			http.Error(w, "can not cancel order (mb its completed)", http.StatusBadRequest)
			return
		}
		if errors.Is(err, repository.ErrOrderAlreadyCanceled) {
			http.Error(w, "order has already canceled", http.StatusBadRequest)
			return
		}
		log.Printf("UpdateStatus error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return

	}
	w.WriteHeader(http.StatusNoContent)
}
