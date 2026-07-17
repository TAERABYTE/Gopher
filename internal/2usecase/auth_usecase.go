package usecase

import (
	"context"
	"errors"
	"time"

	"go-minimal-backend/internal/4domain"
	appjwt "go-minimal-backend/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type authUsecase struct {
	userRepo   domain.UserRepository
	jwtSecret  string
	tokenExpir time.Duration
}

func NewAuthUsecase(ur domain.UserRepository, secret string, expir time.Duration) domain.AuthUsecase {
	return &authUsecase{
		userRepo:   ur,
		jwtSecret:  secret,
		tokenExpir: expir,
	}
}

func (a *authUsecase) ValidateCredentials(ctx context.Context, username, password string) (*domain.User, error) {
	user, err := a.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidCreds
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, domain.ErrInvalidCreds
	}

	return user, nil
}

func (a *authUsecase) Login(ctx context.Context, username, password string) (*domain.User, string, error) {
	user, err := a.ValidateCredentials(ctx, username, password)
	if err != nil {
		return nil, "", err
	}

	token, err := appjwt.GenerateToken(user.ID, user.Username, user.Role, a.jwtSecret, a.tokenExpir)
	if err != nil {
		return nil, "", domain.ErrInternalServer
	}

	return user, token, nil
}

func (a *authUsecase) Register(ctx context.Context, username, password, role string) error {
	// check if user already exists
	_, err := a.userRepo.GetByUsername(ctx, username)
	if err == nil {
		return domain.ErrUserExists
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.ErrInternalServer
	}

	if role != domain.RoleAdmin {
		role = domain.RoleUser
	}

	user := &domain.User{
		Username: username,
		Password: string(hashedPassword),
		Role:     role,
	}

	return a.userRepo.Create(ctx, user)
}
