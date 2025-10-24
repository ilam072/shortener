package domain

import (
	"github.com/google/uuid"
	"time"
)

type Link struct {
	ID        uuid.UUID
	URL       string
	Alias     string
	CreatedAt time.Time
}
