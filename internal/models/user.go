package models

import "time"

type User struct {
	ID            string     `json:"id" db:"id"`
	Email         string     `json:"email" db:"email"`
	PasswordHash  string     `json:"-" db:"password_hash"`
	FullName      string     `json:"full_name" db:"full_name"`
	Company       string     `json:"company" db:"company"`
	Role          string     `json:"role" db:"role"`
	EmailVerified bool       `json:"email_verified" db:"email_verified"`
	Settings      JSON       `json:"settings" db:"settings"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt   *time.Time `json:"last_login_at" db:"last_login_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// UserCredentials для регистрации/логина
type UserCredentials struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// UserRegistration для регистрации
type UserRegistration struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Company  string `json:"company"`
}

// UserResponse для ответа API (без чувствительных данных)
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Company   string    `json:"company"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
