package auth

//REFACTOR ADD ERR VARS

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/keykibatyr/triad-chat/internal/config"
)

type JWTService interface {
	GenerateAccessToken(userID int64, role string) (accessToken string, err error)
	GenerateRefreshToken(userID int64) (refreshToken string, err error)
	GenerateTokenPair(userID int64, role string) (*TokenPair, error)
	ValidateAccessToken(tokenString string) (*CustomClaims, error)
	ValidateRefreshToken(tokenString string) (*CustomClaims, error)
}

type jwtService struct {
	AccessSecret  []byte
	RefreshSecret []byte
	config        *config.Config
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	AccessExpiresAt   time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

func NewJWTService(
	accessSecret []byte,
	refreshSecret []byte,
	cfg *config.Config,
) JWTService {
	return &jwtService{
		AccessSecret:  accessSecret,
		RefreshSecret: refreshSecret,
		config:        cfg,
	}
}

func (j *jwtService) GenerateAccessToken(userID int64, role string) (accessToken string, err error) {
	now := time.Now()

	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.config.JWT.AccessTokenTTL)),
		},
		Role:      role,
		TokenType: "AccessToken",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(j.AccessSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return signedToken, nil
}

func (j *jwtService) GenerateRefreshToken(userID int64) (refreshToken string, err error) {
	now := time.Now()
	tokenID := uuid.New().String()


	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID: tokenID,
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.config.JWT.RefreshTokenTTL)),
		},
		TokenType: "RefreshToken",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(j.RefreshSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return signedToken, nil
}

func (j *jwtService) GenerateTokenPair(userID int64, role string) (*TokenPair, error) {
	now := time.Now()
	accessToken, err := j.GenerateAccessToken(userID, role)
	if err != nil {
		return nil, fmt.Errorf("could not generate the Access Token: %w", err)
	}

	refreshToken, err := j.GenerateRefreshToken(userID)
	if err != nil {
		return nil, fmt.Errorf("could not generate the Refresh Token: %w", err)
	}

	tokenPair := TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		AccessExpiresAt:    now.Add(j.config.JWT.AccessTokenTTL),
		RefreshExpiresAt: now.Add(j.config.JWT.RefreshTokenTTL),
	}

	return &tokenPair, nil
}

func (j *jwtService) ValidateAccessToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }

		return j.AccessSecret, nil

	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token is expired: %w", err)
		}

		return nil, fmt.Errorf("token is invalid: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}

	if claims.TokenType != "AccessToken" {
		return nil, fmt.Errorf("invalid TokenType: %w", err)
	}

	return claims, nil

}

func (j *jwtService) ValidateRefreshToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
		return j.RefreshSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("refresh token has expired: %w", err)
		}

		return nil, fmt.Errorf("refresh token is invalid: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}

	if claims.TokenType != "RefreshToken" {
		return nil, fmt.Errorf("invalid token type")
	}


	return claims, nil
}

