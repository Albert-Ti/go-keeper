package main

import "time"

const DOWNLOAD_PATHNAME = "download"

const (
	expiryLeeway   = 30 * time.Second // обновляем чуть раньше реального истечения
	refreshTimeout = 5 * time.Second
)
