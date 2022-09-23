package authv1

const (
	DefaultUserName            = "default"
	DefaultUserPermissions     = "+@all"
	DefaultDefaultUserPassword = ""

	PingerUserName            = "pinger"
	PingerUserPermissions     = "-@all +ping +info|replication"
	DefaultPingerUserPassword = "pingpass"

	DefaultPermittedKeys     = "~*"
	DefaultPermittedChannels = "&*"
)
