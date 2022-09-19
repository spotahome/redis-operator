package authv2

const (
	redisConfigCommand       = "redisConfig"
	redisRuntimeCommand      = "redisRuntimeCommand"
	defaultUserPermissions   = "+@all" // backward compatibility.
	defaultAdminPermissions  = "-@all +@admin"
	defaultAdminUserPassword = "password"
	AdminUserName            = "admin"
	defaultPermittedKeys     = "~*"
	defaultPermittedChannels = "&*"
)
