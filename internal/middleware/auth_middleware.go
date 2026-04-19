package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/keykibatyr/triad-chat/internal/auth"
	"github.com/keykibatyr/triad-chat/internal/repository"
)

type AuthMiddleware struct {
	JWT         auth.JWTService
	AuthService repository.AuthService
}

func (m *AuthMiddleware) AuthAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("access_token")
		log.Print(tokenString)
		if err == nil {
			tokenClaims, err := m.JWT.ValidateAccessToken(tokenString)
			if err == nil {
				userID, err := strconv.ParseInt(tokenClaims.Subject, 10, 64)
				if err != nil {
					c.Redirect(http.StatusFound, "/login")
					c.Abort()
					return
				}

				c.Set("userID", int64(userID))
				c.Set("role", string(tokenClaims.Role))
				c.Next()
				return
			}
		}

		refreshTokenString, err := c.Cookie("refresh_token")
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		tokenPair, err := m.AuthService.RefreshToken(c, refreshTokenString)
		if err != nil {
			c.SetCookie("access_token", "", -1, "/", "localhost", false, true)
			c.SetCookie("refresh_token", "", -1, "/", "localhost", false, true)
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.SetCookie("access_token", tokenPair.AccessToken,
			int(time.Until(tokenPair.AccessExpiresAt).Seconds()),
			"/", "", false, true)
		c.SetCookie("refresh_token", tokenPair.RefreshToken,
			int(time.Until(tokenPair.RefreshExpiresAt).Seconds()),
			"/", "", false, true)

		tokenClaims, err := m.JWT.ValidateAccessToken(tokenPair.AccessToken)
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		userID, err := strconv.ParseInt(tokenClaims.Subject, 10, 64)
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Set("userID", int64(userID))
		c.Set("role", string(tokenClaims.Role))
		c.Next()
	}
}

func CurrentUser(c *gin.Context) (int64, string, error) {
	val1, exists := c.Get("userID")
	if !exists {
		log.Print("bug1")
		return 0, "", fmt.Errorf("failed getting val from context: %t", exists)
	}

	userID, ok := val1.(int64)
	if !ok {
		log.Print("bug2")
		return 0, "", fmt.Errorf("failed converting to int64")
	}

	val2, exists := c.Get("role")
	if !exists {
		log.Print("bug3")
		return 0, "", fmt.Errorf("failed getting val from context: %t", exists)
	}

	role, ok := val2.(string)
	if !ok {
		log.Print("bug4")
		return 0, "", fmt.Errorf("failed converting to string")
	}

	return userID, role, nil
}



// func (m *AuthMiddleware) AuthAccess() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		authHeader := c.GetHeader("Authorization")
// 		if authHeader == "" {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
// 			c.Abort()
// 			return
// 		}

// 		parts := strings.Split(authHeader, "Bearer ")
// 		if len(parts) != 2 {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "unautorized"})
// 			c.Abort()
// 			return
// 		}

// 		tokenString := parts[1]

// 		tokenClaims, err := m.JWT.ValidateAccessToken(tokenString)
// 		if err != nil {
// 			c.Redirect(http.StatusFound, "/login")
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
// 			c.Abort()
// 			return
// 		}

// 		userID, err := strconv.ParseInt(tokenClaims.Subject, 10, 64)
// 		if err != nil {
// 			c.Redirect(http.StatusFound, "/login")
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not parse uuid"})
// 			c.Abort()
// 			return
// 		}

// 		c.Set("userID", userID)
// 		c.Set("role", tokenClaims.Role)

// 		c.Next()
// 	}
// }
