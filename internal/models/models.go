package models

type User struct {
	UUID           string
	Email          string
	EmailToken     string
	IsConfirmEmail bool
	Password       string
}
