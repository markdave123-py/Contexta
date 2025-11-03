package services

import (
	"context"
	"errors"

	"github.com/markdave123-py/Contexta/internal/core"
	"github.com/markdave123-py/Contexta/internal/models"
)

type UserService struct {
	db core.DbClient
}

func NewUserService(db core.DbClient) *UserService {
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
