package models

import "time"

type DataStatus uint

const (
	DataStatusUploading DataStatus = iota
	DataStatusSuccess
	DataStatusFailed
)

type ArbitraryData struct {
	ID         int64
	Name       []byte
	Type       string
	Status     int64
	ObjectKey  string
	OSPath     []byte
	ClientSize int64
	TotalSize  int64
	CreatedAt  time.Time
}

type UpdateArbitraryDataParams struct {
	ID        int64
	Status    *uint
	TotalSize *int64
}
