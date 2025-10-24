package domain

import (
	"github.com/google/uuid"
)

type Click struct {
	ID        uuid.UUID
	Alias     string
	UserAgent string
	Client    string
	Device    string
	IP        string
}
