package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/swathikrish753/ecommerce/services/auth/internal/domain"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

// Auth depends on the repository INTERFACE, not on Postgres.
type Auth struct {
	repo      domain.UserRepository
	jwtSecret []byte
	jwtTTL    time.Duration
}

func NewAuth(repo domain.UserRepository, jwtSecret string, ttl time.Duration) *Auth {
	return &Auth{repo: repo, jwtSecret: []byte(jwtSecret), jwtTTL: ttl}
}

func (a *Auth) Register(ctx context.Context, email, password string) (*domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := &domain.User{Email: email, PasswordHash: string(hash)}
	if err := a.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (a *Auth) Login(ctx context.Context, email, password string) (string, error) {
	u, err := a.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	return a.issueToken(u)
}

func (a *Auth) issueToken(u *domain.User) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   u.ID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.jwtTTL)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "ecommerce-auth",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(a.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}
