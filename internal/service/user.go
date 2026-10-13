package service

import (
	"Confetki/internal/model"
	"Confetki/internal/repository"
	"Confetki/pkg/auth"
	"context"
	"errors"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewServiceAuth(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

func (a *AuthService) Register(ctx context.Context, name string, email string, password string) (*model.User, error) {
	hashPassword, err := auth.HashPassword(password)
	if err != nil {
		return &model.User{}, err
	}

	user := model.User{
		Name:         name,
		PasswordHash: hashPassword,
		Email:        email,
		Role:         "user",
	}

	err = a.userRepo.Create(ctx, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (a *AuthService) Login(ctx context.Context, email string, password string) (string, error) {
	user, err := a.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	if !auth.CheckPasswordHash(password, user.PasswordHash) {
		return "", errors.New("invalid credentials")
	}

	token, err := auth.GenerateToken(user.ID, user.Role)
	if err != nil {
		return "", err
	}

	return token, nil
}
