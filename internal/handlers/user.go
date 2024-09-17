package handlers

import (
	"time"
)

type User struct {
	ID                    int       `json:"id"`
	Email                 string    `json:"email"`
	Password              string    `json:"password"`
	ExpiresInSeconds      int       `json:"expires_in_seconds"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	IsChirpyRed           bool      `json:"is_chirpy_red"`
}

func (user User) GetID() (ID int) {
	return user.ID
}

func (user User) GetUniqueIdentifier() (uniqueIdentifier string) {
	return user.Email
}
