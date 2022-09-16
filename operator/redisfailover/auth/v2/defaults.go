package authv2

const (
	redisConfigCommand         = "redisConfig"
	redisRuntimeCommand        = "redisRuntimeCommand"
	defaultUserPermissions     = "+@all" // backward compatibility.
	defaultAdminPermissions    = "-@all +@admin"
	defaultUserName            = "default"
	defaultDefaultUserPassword = "password"
	defaultAdminUserPassword   = "password"
	adminUserName              = "admin"
	defaultPermittedKeys       = "~*"
	defaultPermittedChannels   = "&*"
)
