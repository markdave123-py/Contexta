package services

import (
	"context"
	"errors"

	db "github.com/markdave123-py/Contexta/internal/core/database"
	"github.com/markdave123-py/Contexta/internal/models"
)

type UserService struct {
	db db.DbClient
}

func NewUserService(db db.DbClient) *UserService {
	return &UserService{db: db}
}

func (s *UserService) Create(ctx context.Context, u *models.User) error {
	if u == nil || u.Email == "" || u.PasswordHash == "" {
		return errors.New("invalid user payload")
	}
	return s.db.CreateUser(ctx, u)
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.db.GetUserByEmail(ctx, email)
}
