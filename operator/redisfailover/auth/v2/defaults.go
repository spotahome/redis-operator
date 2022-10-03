package authv2

const (
	RedisConfigCommand  = "redisConfig"
	RedisRuntimeCommand = "redisRuntimeCommand"

	AdminUserName            = "admin"
	DefaultAdminPermissions  = "+client +ping +multi +info +slaveof +exec +subscribe +config|rewrite +role +publish +client|setname +client|kill +script|kill +config|get +cluster|info +slowlog +latency +memory +select +get +scan +xinfo +type +pfcount +strlen +llen +scard +zcard +hlen +xlen +eval +@admin"
	DefaultAdminUserPassword = "password"

	DefaultUserName            = "default"
	DefaultUserPermissions     = "+@all" // backward compatibility.
	DefaultDefaultUserPassword = "password"

	PingerUserName            = "pinger"
	PingerUserPermissions     = "-@all +ping +info|replication" // backward compatibility.
	DefaultPingerUserPassword = "pingpass"

	DefaultPermittedKeys     = "~*"
	DefaultPermittedChannels = "&*"
)
