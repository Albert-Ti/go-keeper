package models

type User struct {
	UUID           string
	Email          string
	EmailCode      string
	IsConfirmEmail bool
	Password       string
}

type UpdateUserParams struct {
	UUID           string
	Email          *string
	EmailCode      *string
	IsConfirmEmail *bool
	Password       *string
}
