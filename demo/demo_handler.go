package demo

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/Strangebrewer/go-auth/pubsub"
	"github.com/Strangebrewer/go-auth/token"
	"github.com/Strangebrewer/go-auth/user"
)

const (
	demoTTL      = 2 * time.Hour
	ipLimit      = 3
	passwordLen  = 16
	passwordChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type Handler struct {
	userStore    *user.Store
	tokenService *token.Service
	demoStore    *Store
	publisher    *pubsub.Publisher
	topicID      string
}

func NewHandler(userStore *user.Store, tokenService *token.Service, demoStore *Store, publisher *pubsub.Publisher, topicID string) *Handler {
	return &Handler{
		userStore:    userStore,
		tokenService: tokenService,
		demoStore:    demoStore,
		publisher:    publisher,
		topicID:      topicID,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	ip := extractIP(r)

	limitReached, err := h.demoStore.CheckAndIncrementIP(r.Context(), ip, ipLimit)
	if err != nil {
		slog.Error("demo register: check ip limit", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if limitReached {
		http.Error(w, "demo account limit reached for this IP", http.StatusTooManyRequests)
		return
	}

	username := generateUsername()
	password := generatePassword()
	expiresAt := time.Now().UTC().Add(demoTTL)

	u, err := h.userStore.CreateDemo(r.Context(), username, password, expiresAt)
	if err != nil {
		slog.Error("demo register: create user", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	result, err := h.tokenService.IssueForDemoUser(r.Context(), u.ID)
	if err != nil {
		slog.Error("demo register: issue tokens", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if h.publisher != nil {
		h.publisher.Publish(h.topicID, pubsub.DemoRegisteredPayload{
			UserID:    u.ID.String(),
			ExpiresAt: expiresAt,
			TraceID:   r.Header.Get("X-Trace-ID"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(DemoRegisterResponse{
		Username:     username,
		Password:     password,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	})
}

func extractIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		// X-Forwarded-For may be a comma-separated list; first value is the client IP
		return strings.TrimSpace(strings.SplitN(fwd, ",", 2)[0])
	}
	// RemoteAddr is "host:port" — strip the port
	addr := r.RemoteAddr
	if i := strings.LastIndex(addr, ":"); i != -1 {
		return addr[:i]
	}
	return addr
}

func generateUsername() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("demo-%x@demo.local", b)
}

func generatePassword() string {
	b := make([]byte, passwordLen)
	for i := range b {
		b[i] = passwordChars[rand.Intn(len(passwordChars))]
	}
	return string(b)
}
