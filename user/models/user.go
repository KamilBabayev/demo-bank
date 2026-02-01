package models

import (
	"time"
)

type User struct {
	ID                  int64      `json:"id"`
	Username            string     `json:"username"`
	Email               string     `json:"email"`
	PasswordHash        string     `json:"-"` // Never expose password hash in JSON
	FirstName           string     `json:"first_name"`
	LastName            string     `json:"last_name"`
	Phone               *string    `json:"phone,omitempty"`
	Role                string     `json:"role"`
	Status              string     `json:"status"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	LastLoginAt         *time.Time `json:"last_login_at,omitempty"`
	FailedLoginAttempts int        `json:"-"` // Internal security field
	LockedUntil         *time.Time `json:"-"` // Internal security field
}

type CreateUserRequest struct {
	Username  string  `json:"username" binding:"required,min=3,max=50"`
	Email     string  `json:"email" binding:"required,email"`
	Password  string  `json:"password" binding:"required,min=8"`
	FirstName string  `json:"first_name" binding:"required"`
	LastName  string  `json:"last_name" binding:"required"`
	Phone     *string `json:"phone,omitempty"`
	Role      string  `json:"role" binding:"required,oneof=customer admin"`
}

type UpdateUserRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Status    *string `json:"status,omitempty" binding:"omitempty,oneof=active suspended closed"`
}

type UserListResponse struct {
	Users []User `json:"users"`
	Total int64  `json:"total"`
}
