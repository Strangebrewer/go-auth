package user

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/Strangebrewer/go-auth/middleware"
	"github.com/Strangebrewer/go-auth/token"
	"github.com/Strangebrewer/go-auth/tracer"
)

type Handler struct {
	store        *Store
	tokenService *token.Service
	tracer       *tracer.Client
}

func NewHandler(store *Store, tokenService *token.Service, tc *tracer.Client) *Handler {
	return &Handler{store: store, tokenService: tokenService, tracer: tc}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	traceID := r.Header.Get("X-Trace-ID")
	start := time.Now()

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		end := time.Now()
		h.tracer.SendErrorSpan(traceID, "register_user", "invalid json", start, end)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	req.Email = normalizeEmail(req.Email)
	if req.Email == "" || len(req.Email) > 254 {
		end := time.Now()
		h.tracer.SendErrorSpan(traceID, "register_user", "invalid email", start, end)
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	}
	if utf8.RuneCountInString(req.Password) < 12 {
		end := time.Now()
		h.tracer.SendErrorSpan(traceID, "register_user", "password too short", start, end)
		http.Error(w, "password must be at least 12 characters", http.StatusBadRequest)
		return
	}

	u, err := h.store.Create(r.Context(), req.Email, req.Password)
	if err != nil {
		end := time.Now()
		if errors.Is(err, ErrEmailExists) {
			h.tracer.SendErrorSpan(traceID, "register_user", "email already in use", start, end)
			http.Error(w, "email already in use", http.StatusConflict)
			return
		}
		slog.Error("register: create user", "error", err)
		h.tracer.SendErrorSpan(traceID, "register_user", "internal server error", start, end)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(publicUser(u.ID, u.Email))
	end := time.Now()
	h.tracer.SendSpan(traceID, "register_user", start, end)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	traceID := r.Header.Get("X-Trace-ID")
	start := time.Now()

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		end := time.Now()
		h.tracer.SendErrorSpan(traceID, "login_user", "invalid json", start, end)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		end := time.Now()
		h.tracer.SendErrorSpan(traceID, "login_user", "email and password required", start, end)
		http.Error(w, "email and password required", http.StatusBadRequest)
		return
	}

	u, err := h.store.FindByEmail(r.Context(), req.Email)
	if err != nil {
		end := time.Now()
		h.tracer.SendErrorSpan(traceID, "login_user", "invalid credentials", start, end)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if u.Disabled {
		end := time.Now()
		h.tracer.SendErrorSpan(traceID, "login_user", "invalid credentials", start, end)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	ok, err := verifyPassword(req.Password, u.PasswordHash)
	if err != nil || !ok {
		end := time.Now()
		h.tracer.SendErrorSpan(traceID, "login_user", "invalid credentials", start, end)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	result, err := h.tokenService.IssueForUser(r.Context(), u.ID)
	if err != nil {
		end := time.Now()
		slog.Error("login: issue tokens", "error", err)
		h.tracer.SendErrorSpan(traceID, "login_user", "internal server error", start, end)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(LoginResponse{
		User:         publicUser(u.ID, u.Email),
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	})
	end := time.Now()
	h.tracer.SendSpan(traceID, "login_user", start, end)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	traceID := r.Header.Get("X-Trace-ID")
	start := time.Now()

	idStr, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		end := time.Now()
		h.tracer.SendErrorSpan(traceID, "get_current_user", "unauthorized", start, end)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		end := time.Now()
		h.tracer.SendErrorSpan(traceID, "get_current_user", "unauthorized", start, end)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	u, err := h.store.FindByID(r.Context(), id)
	if err != nil {
		end := time.Now()
		if errors.Is(err, ErrNotFound) {
			h.tracer.SendErrorSpan(traceID, "get_current_user", "not found", start, end)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		slog.Error("me: find user", "error", err)
		h.tracer.SendErrorSpan(traceID, "get_current_user", "internal server error", start, end)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(publicUser(u.ID, u.Email))
	end := time.Now()
	h.tracer.SendSpan(traceID, "get_current_user", start, end)
}
