package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/warungos/auth-service/model"
	"github.com/warungos/auth-service/repository"
	rdc "github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler holds dependencies for auth endpoints.
type AuthHandler struct {
	userRepo  *repository.UserRepository
	pool      *pgxpool.Pool
	rdb       *rdc.Client
	jwtSecret []byte
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(pool *pgxpool.Pool, rdb *rdc.Client, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		userRepo:  repository.NewUserRepository(pool),
		pool:      pool,
		rdb:       rdb,
		jwtSecret: []byte(jwtSecret),
	}
}

// Login authenticates a user and returns JWT tokens.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}

	// Rate limit: 5 attempts per 15 minutes per email
	key := fmt.Sprintf("auth:login:%s", req.Email)
	attempts, _ := h.rdb.Incr(r.Context(), key).Result()
	if attempts == 1 {
		h.rdb.Expire(r.Context(), key, 15*time.Minute)
	}
	if attempts > 5 {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "too many login attempts, try again later"})
		return
	}

	user, err := h.userRepo.FindByEmail(r.Context(), req.Email)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	// Successful login — reset rate limit
	h.rdb.Del(r.Context(), key)

	tokens, err := h.generateTokens(user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate tokens"})
		return
	}

	writeJSON(w, http.StatusOK, tokens)
}

// Register creates a new user account.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Email == "" || req.Password == "" || req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email, name, and password are required"})
		return
	}

	// Check for duplicate email
	existing, _ := h.userRepo.FindByEmail(r.Context(), req.Email)
	if existing != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "email already registered"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
		return
	}

	role := req.Role
	if role == "" {
		role = "cashier"
	}

	user := &model.User{
		ID:           generateUUID(),
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: string(hash),
		Role:         role,
		BranchID:     req.BranchID,
	}

	if err := h.userRepo.Create(r.Context(), user); err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "failed to create user"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
		"role":  user.Role,
	})
}

// Refresh rotates tokens using a valid refresh token.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req model.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.RefreshToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "refresh_token is required"})
		return
	}

	// Parse and validate the refresh token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return h.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid claims"})
		return
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "token is not a refresh token"})
		return
	}

	// Check if token is blacklisted (token rotation)
	tokenStr, _ := claims["jti"].(string)
	if tokenStr != "" {
		blacklisted, _ := h.rdb.Get(r.Context(), "auth:blacklist:"+tokenStr).Result()
		if blacklisted != "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "refresh token has been revoked"})
			return
		}
		// Blacklist old refresh token
		h.rdb.Set(r.Context(), "auth:blacklist:"+tokenStr, "1", 7*24*time.Hour)
	}

	userID, _ := claims["user_id"].(string)
	user, err := h.userRepo.FindByID(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "user not found"})
		return
	}

	tokens, err := h.generateTokens(user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate tokens"})
		return
	}

	writeJSON(w, http.StatusOK, tokens)
}

// Me returns the current authenticated user's profile.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		return
	}

	user, err := h.userRepo.FindByID(r.Context(), userID.(string))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"id":        user.ID,
		"email":     user.Email,
		"name":      user.Name,
		"role":      user.Role,
		"branch_id": user.BranchID,
	})
}

// generateTokens creates an access token (15 min) and refresh token (7 days).
func (h *AuthHandler) generateTokens(user *model.User) (*model.TokenResponse, error) {
	now := time.Now()

	// Access token — 15 minutes
	accessClaims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"type":    "access",
		"exp":     now.Add(15 * time.Minute).Unix(),
		"iat":     now.Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString(h.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	// Refresh token — 7 days
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"type":    "refresh",
		"jti":     generateUUID(),
		"exp":     now.Add(7 * 24 * time.Hour).Unix(),
		"iat":     now.Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString(h.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &model.TokenResponse{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
		ExpiresIn:    900,
		TokenType:    "Bearer",
	}, nil
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// generateUUID produces a simple UUID v4 string.
func generateUUID() string {
	b := make([]byte, 16)
	now := time.Now().UnixNano()
	for i := 0; i < 16; i++ {
		b[i] = byte((now >> (uint(i) * 4)) & 0xff)
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 1
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
