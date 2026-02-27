package models

import (
	"time"
)

// JSON type for PostgreSQL
type JSON map[string]interface{}

type Project struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	APIKey    string    `json:"api_key" db:"api_key"`
	UserID    string    `json:"user_id" db:"user_id"`
	Settings  JSON      `json:"settings" db:"settings"`
	Active    bool      `json:"active" db:"active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type APIKey struct {
	Key       string     `json:"key" db:"key"`
	ProjectID string     `json:"project_id" db:"project_id"`
	Name      string     `json:"name" db:"name"`
	LastUsed  *time.Time `json:"last_used" db:"last_used"`
	ExpiresAt *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

type Claims struct {
	UserID    string `json:"user_id"`
	ProjectID string `json:"project_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
}

type Session struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	RefreshToken string    `json:"refresh_token" db:"refresh_token"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
