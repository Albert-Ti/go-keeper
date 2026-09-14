package models

import "time"

type User struct {
	UUID           string
	Email          string
	EmailCode      string
	IsConfirmEmail bool
	Password       string
}

type UpdateUserParams struct {
	Email          string
	EmailCode      *string
	IsConfirmEmail *bool
	Password       *string
}

type Profile struct {
	Email     string
	Password  string
	CreatedAt time.Time
}

type Card struct {
	CardNumber []byte
	ExpiryDate time.Time
	Active     bool
}
