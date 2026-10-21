package consts

var PublicMethods = map[string]bool{
	"/gokeeper.GoKeeperService/Register":     true,
	"/gokeeper.GoKeeperService/Login":        true,
	"/gokeeper.GoKeeperService/ConfirmEmail": true,
	"/gokeeper.GoKeeperService/RefreshToken": true,
}
