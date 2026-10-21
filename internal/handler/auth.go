package handler

import (
	"Confetki/internal/service"
	"encoding/json"
	"net/http"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := a.authService.Register(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		http.Error(w, "failed to register(((", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var log LoginRequest

	err := json.NewDecoder(r.Body).Decode(&log)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	token, err := a.authService.Login(r.Context(), log.Email, log.Password)
	if err != nil {
		http.Error(w, "failed to login", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
