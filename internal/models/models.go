package models

type User struct {
	ID             string
	Email          string
	EmailToken     string
	IsConfirmEmail bool
	Password       string
}
