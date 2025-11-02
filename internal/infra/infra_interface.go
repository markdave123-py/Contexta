package infra

import "context"
import "github.com/markdave123-py/Contexta/internal/models"

type DbClient interface {
	CreateUser(ctx context.Context, user *models.User) (err error)
	// GetUserByEmail(ctx context.Context, email string) (user *models.User, err error)
}