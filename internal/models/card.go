package models

import "time"

type Card struct {
	ID         int64
	CardNumber []byte
	ExpiryDate time.Time
	Active     bool
}
