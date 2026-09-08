package models

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
