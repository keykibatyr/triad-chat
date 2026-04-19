package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/keykibatyr/triad-chat/internal/auth"
	"github.com/keykibatyr/triad-chat/internal/models"
	"github.com/keykibatyr/triad-chat/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepo    repository.UserRepository
	RefreshRepo repository.RefreshTokenRepository
	JWT         auth.JWTService
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshRepo repository.RefreshTokenRepository,
	jwt auth.JWTService,
) repository.AuthService {
	return &AuthService{
		UserRepo:    userRepo,
		RefreshRepo: refreshRepo,
		JWT:         jwt,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, username string) (*models.User, *auth.TokenPair, error) {

	email = strings.ToLower(email)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("could not create passwordHash: %w", err)
	}

	user := models.User{
		Email:        email,
		PasswordHash: string(passwordHash),
		Username:     username,
		Role:         "user",
	}

	err = s.UserRepo.Create(ctx, &user)
	if err != nil {
		return nil, nil, fmt.Errorf("could not create the User")
	}

	tokenPair, err := s.JWT.GenerateTokenPair(int64(user.ID), user.Role)
	if err != nil {
		return nil, nil, fmt.Errorf("could not generate the tokenPair:%w", err)
	}

	err = s.refreshStore(ctx, tokenPair, int64(user.ID))
	if err != nil {
		return nil, nil, fmt.Errorf("could not store refresh token in db: %w", err)
	}

	return &user, tokenPair, nil

}

func (s *AuthService) Login(ctx context.Context, email, password string) (*models.User, *auth.TokenPair, error) {
	email = strings.ToLower(email)

	user, err := s.UserRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get the user by email:%w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, nil, fmt.Errorf("passwordhash comparison fail:%w", err)
	}

	tokenPair, err := s.JWT.GenerateTokenPair(int64(user.ID), user.Role)
	if err != nil {
		return nil, nil, fmt.Errorf("could not generate the tokenPair:%w", err)
	}

	err = s.refreshStore(ctx, tokenPair, int64(user.ID))
	if err != nil {
		return nil, nil, fmt.Errorf("could not store refresh token in db: %w", err)
	}

	log.Println(tokenPair)
	log.Println(user.Role)

	return user, tokenPair, nil

}

func (s *AuthService) refreshStore(ctx context.Context, tokenPair *auth.TokenPair, userID int64) error {
	refreshClaims, err := s.JWT.ValidateRefreshToken(tokenPair.RefreshToken)
	if err != nil {
		return fmt.Errorf("claim is not valid")
	}

	uuidTokenID, err := uuid.Parse(refreshClaims.ID)
	if err != nil {
		return fmt.Errorf("failed parsing the uuidToken from Claims: %w", err)
	}

	data := []byte(tokenPair.RefreshToken)

	sum := sha256.Sum256(data)

	tokenHash := hex.EncodeToString(sum[:])

	refreshTokenStruct := models.RefreshToken{
		JTI:       uuidTokenID,
		UserID:    int64(userID),
		TokenHash: tokenHash,
		ExpiresAt: tokenPair.RefreshExpiresAt,
	}

	err = s.RefreshRepo.Create(ctx, &refreshTokenStruct)

	return nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenString string) (*auth.TokenPair, error) {

	data := []byte(refreshTokenString)

	sum := sha256.Sum256(data)

	hashedRefreshToken := hex.EncodeToString(sum[:])

	refreshToken, err := s.RefreshRepo.GetByTokenHash(ctx, hashedRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("could not retrive the tokenhash: %w", err)
	}

	if (time.Now()).After(refreshToken.ExpiresAt) {
		_ = s.RefreshRepo.DeleteById(ctx, refreshToken.JTI)
		return nil, fmt.Errorf("Token is Expired")
	}

	refreshClaims, err := s.JWT.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("refresh claims are invalid: %w", err)
	}

	uuidTokenID, err := uuid.Parse(refreshClaims.ID)
	if err != nil {
		return nil, fmt.Errorf("failed parsing the uuidToken from Claims: %w", err)
	}

	if uuidTokenID != refreshToken.JTI {
		return nil, fmt.Errorf("refreshToken JTI is invalid")
	}

	err = s.RefreshRepo.DeleteById(ctx, uuidTokenID)
	if err != nil {
		return nil, fmt.Errorf("could not delete refreshToken by jti: %w", err)
	}

	tokenPair, err := s.JWT.GenerateTokenPair(refreshToken.UserID, refreshClaims.Role)
	if err != nil {
		return nil, fmt.Errorf("could not generate tokenPair: %w", err)
	}

	err = s.refreshStore(ctx, tokenPair, refreshToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("could not store refresh token in db: %w", err)
	}

	return tokenPair, nil
}

func (s AuthService) Logout(ctx context.Context, refreshTokenString string) error {
	refreshClaims, err := s.JWT.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return fmt.Errorf("refresh claims are invalid: %w", err)
	}

	uuidTokenID, err := uuid.Parse(refreshClaims.ID)
	if err != nil {
		return fmt.Errorf("failed parsing the uuidToken from Claims: %w", err)
	}

	err = s.RefreshRepo.DeleteById(ctx, uuidTokenID)
	if err != nil {
		return fmt.Errorf("could not delete refreshToken by jti: %w", err)
	}

	return nil
}
