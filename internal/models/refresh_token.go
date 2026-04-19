package models

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        int64
	JTI       uuid.UUID
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
}
