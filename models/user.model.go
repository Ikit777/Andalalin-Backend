package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID               uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	Name             string    `gorm:"type:varchar(255);not null"`
	Email            string    `gorm:"uniqueIndex;not null"`
	Password         string    `gorm:"not null"`
	Role             string    `sql:"type:enum('Super Admin', 'Dinas Perhubungan', 'Admin', 'Operator', 'Petugas', 'User');default:'User';not null"`
	Photo            []byte
	VerificationCode string `gorm:"type:varchar(255);not null"`
	Verified         bool   `gorm:"not null"`
	ResetToken       string
	ResetAt          time.Time
	PushToken        string
	CreatedAt        string `gorm:"not null"`
	UpdatedAt        string `gorm:"not null"`
}

type UserSignUp struct {
	Name            string `json:"name" binding:"required"`
	Email           string `json:"email" binding:"required"`
	Password        string `json:"password" binding:"required,min=8"`
	PasswordConfirm string `json:"passwordConfirm" binding:"required"`
}

type Notifikasi struct {
	IdUser uuid.UUID `gorm:"not null"`
	Title  string    `gorm:"not null"`
	Body   string    `gorm:"not null"`
}

type NotifikasiRespone struct {
	IdUser uuid.UUID `json:"id_user,omitempty"`
	Title  string    `json:"title,omitempty"`
	Body   string    `json:"body,omitempty"`
}

type UserAdd struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Role     string `json:"role" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type UserSignIn struct {
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
	PushToken string `json:"push_token"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id,omitempty"`
	Name      string    `json:"name,omitempty"`
	Email     string    `json:"email,omitempty"`
	Role      string    `json:"role,omitempty"`
	Photo     []byte    `json:"photo,omitempty"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

type Delete struct {
	ID   uuid.UUID `json:"id" binding:"required"`
	Role string    `json:"role" binding:"required"`
}

type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required"`
}

type ForgotPasswordRespone struct {
	ResetToken string    `json:"reset_token"`
	ResetAt    time.Time `json:"reset_at,omitempty"`
}

type ResetPasswordInput struct {
	Password        string `json:"password" binding:"required"`
	PasswordConfirm string `json:"passwordConfirm" binding:"required"`
}
