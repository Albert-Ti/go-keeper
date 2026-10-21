package models

import "time"

type User struct {
	UUID           string
	Email          string
	EmailCode      string
	IsConfirmEmail bool
	Pass           string
}

type UpdateUserParams struct {
	UUID           string
	Email          string
	EmailCode      *string
	IsConfirmEmail *bool
}

type Profile struct {
	Email     string
	Pass      string
	CreatedAt time.Time
}
