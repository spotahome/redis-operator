package authv2

const (
	redisConfigCommand         = "redisConfig"
	redisRuntimeCommand        = "redisRuntimeCommand"
	defaultUserPermissions     = "+@all" // backward compatibility.
	defaultAdminPermissions    = "-@all +@admin"
	DefaultUserName            = "default"
	defaultDefaultUserPassword = "password"
	defaultAdminUserPassword   = "password"
	AdminUserName              = "admin"
	defaultPermittedKeys       = "~*"
	defaultPermittedChannels   = "&*"
)
