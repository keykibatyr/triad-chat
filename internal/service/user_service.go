package service

import (
	"context"
	"fmt"

	"github.com/keykibatyr/triad-chat/internal/models"
	"github.com/keykibatyr/triad-chat/internal/repository"
)

type UserService struct {
	UserRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) repository.UserService {
	return &UserService{
		UserRepo: userRepo,
	}
}

func (s *UserService) GetByID(ctx context.Context, userID int64) (*models.User, error) {
	user, err := s.UserRepo.GetById(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("could not find the user")
	}

	return user, nil
}
