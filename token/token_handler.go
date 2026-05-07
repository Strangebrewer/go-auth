package token

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Strangebrewer/go-auth/tracer"
)

type Handler struct {
	svc    *Service
	tracer *tracer.Client
}

func NewHandler(svc *Service, tc *tracer.Client) *Handler {
	return &Handler{svc: svc, tracer: tc}
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	plain, ok := bearerToken(r)
	if !ok {
		http.Error(w, "missing bearer token", http.StatusUnauthorized)
		return
	}

	traceID := r.Header.Get("X-Trace-ID")
	start := time.Now()

	result, err := h.svc.Exchange(r.Context(), plain)
	if err != nil {
		end := time.Now()
		if errors.Is(err, ErrInvalidToken) {
			h.tracer.SendErrorSpan(traceID, "refresh_token", "invalid token", start, end)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		h.tracer.SendErrorSpan(traceID, "refresh_token", "internal server error", start, end)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
	end := time.Now()
	h.tracer.SendSpan(traceID, "refresh_token", start, end)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	traceID := r.Header.Get("X-Trace-ID")
	start := time.Now()
	plain, ok := bearerToken(r)
	if !ok {
		end := time.Now()
		h.tracer.SendErrorSpan(traceID, "logout", "missing bearer token", start, end)
		http.Error(w, "missing bearer token", http.StatusUnauthorized)
		return
	}

	if err := h.svc.Revoke(r.Context(), plain); err != nil {
		end := time.Now()
		h.tracer.SendErrorSpan(traceID, "logout", "invalid token", start, end)
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	end := time.Now()
	h.tracer.SendSpan(traceID, "logout", start, end)

	w.WriteHeader(http.StatusNoContent)
}

func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return "", false
	}
	return strings.TrimPrefix(header, "Bearer "), true
}
