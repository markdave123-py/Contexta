package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"github.com/markdave123-py/Contexta/internal/core"
	"github.com/markdave123-py/Contexta/internal/models"
)

type AuthHandler struct {
	dbclient core.DbClient
}

func NewAuthHandler(dbclient core.DbClient) *AuthHandler {
	return &AuthHandler{dbclient: dbclient}
}

type signupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", 400)
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	user := &models.User{
		ID:           uuid.NewString(),
		Email:        req.Email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}

	if err := h.dbclient.CreateUser(context.Background(), user); err != nil {
		http.Error(w, "user exists", 409)
		return
	}

	token := generateJWT(user.ID)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", 400)
		return
	}

	user, err := h.dbclient.GetUserByEmail(context.Background(), req.Email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		http.Error(w, "invalid credentials", 401)
		return
	}

	token := generateJWT(user.ID)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// generateJWT creates a signed token with user ID claim
func generateJWT(userID string) string {
	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, _ := tok.SignedString([]byte(secret))
	return token
}
