package auth

import (
	"encoding/json"
	"errors"
	"net/http"
)

const maxBodyBytes = 4 << 10 // 4 KiB; auth payloads are tiny

type Handler struct {
	store UserStorer
}

func NewHandler(store UserStorer) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Username == "" || body.Password == "" {
		http.Error(w, "username and password required", http.StatusBadRequest)
		return
	}

	u, err := h.store.Create(body.Username, body.Password)
	if errors.Is(err, ErrUserExists) {
		http.Error(w, "username already taken", http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	token, err := GenerateToken(u.ID, u.Username)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	SetAuthCookie(w, token)
	writeJSON(w, http.StatusCreated, map[string]string{"userId": u.ID, "username": u.Username})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	u, err := h.store.Authenticate(body.Username, body.Password)
	if errors.Is(err, ErrInvalidCredentials) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	token, err := GenerateToken(u.ID, u.Username)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	SetAuthCookie(w, token)
	writeJSON(w, http.StatusOK, map[string]string{"userId": u.ID, "username": u.Username})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	ClearAuthCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, err := TokenFromRequest(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"userId": claims.UserID, "username": claims.Username})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
