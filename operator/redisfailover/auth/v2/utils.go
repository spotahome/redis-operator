package authv2

import (
	"crypto/sha256"
	"fmt"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/log"
)

func GetHashedPassword(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum(nil))
}

/*
checks if a user identified by the username is present in list of users or not

inputs:

	users ([]redisfailoverv1.User)
	username (string)

outputs:

	bool - true if user is found, false otherwise
*/
func hasUser(users map[string]redisfailoverv1.UserSpec, username string) bool {
	_, ok := users[username]
	return ok
}

/*
returns user object of "admin" user with default settings
inputs:

	none

outputs:

	reference to user object with default settings of admin user
*/
func getDefaultAdminUserSpec() redisfailoverv1.UserSpec {

	return redisfailoverv1.UserSpec{
		Passwords: []redisfailoverv1.Password{
			{
				Value: DefaultAdminUserPassword,
			},
		},
		ACL: redisfailoverv1.ACL{
			Value: DefaultAdminPermissions,
		},
	}
}

/*
returns user object of "default" user with default settings
inputs:

	none

outputs:

	reference to user object with default settings of admin user
*/
func getDefaultDefaultUserSpec() redisfailoverv1.UserSpec {

	return redisfailoverv1.UserSpec{
		Passwords: []redisfailoverv1.Password{
			{
				Value: DefaultDefaultUserPassword,
			},
		},
		ACL: redisfailoverv1.ACL{
			Value: DefaultUserPermissions,
		},
	}
}

/*
returns user object of "default" user with default settings
inputs:

	none

outputs:

	reference to user object with default settings of admin user
*/
func getDefaultPingerUserSpec() redisfailoverv1.UserSpec {

	return redisfailoverv1.UserSpec{
		Passwords: []redisfailoverv1.Password{
			{
				Value: DefaultPingerUserPassword,
			},
		},
		ACL: redisfailoverv1.ACL{
			Value: PingerUserPermissions,
		},
	}
}

/*
	Converts data of type redisfailoverv1.User to string that can be embedded into redis conf or run as redis CLI command

Inputs:

	redisCommandMode (string)   : whether the caller wants the config to be redis-conf compatible format or cli-command compatible format
	*redisfailoverv1.User       : user whos config needs to be converted to string

Outputs:

	string                      : Config converted to string in appropriate format
	error                       : if any error is encountered , nil otherwise.
*/
func getUserSpecAs(redisCommandMode string, username string, userSpec redisfailoverv1.UserSpec) (string, error) {

	passwordCmd := ""
	for _, password := range userSpec.Passwords {
		if password.HashedValue != "" {
			passwordCmd += fmt.Sprintf("#%s ", password.HashedValue)
		}
	}
	// process ACL
	userACL := userSpec.ACL

	/* Should we really add default ACL for new users? if yes, we need to uncomment this.
	if userACL.Value == "" {
		userACL = redisfailoverv1.ACL{Value: defaultUserPermissions}
	}*/

	// placeholder for processing access control keys and channels in spec
	permittedKeys := DefaultPermittedKeys
	permittedChannels := DefaultPermittedChannels

	// format command based on selected mode
	commandPrefix := ""
	if RedisConfigCommand == redisCommandMode {
		commandPrefix = "user"
	} else if RedisRuntimeCommand == redisCommandMode {
		commandPrefix = "acl setuser"
	} else {
		return "", fmt.Errorf("redis command mode not recognised: %s ; accepted modes - %s and %s", redisCommandMode, RedisConfigCommand, RedisRuntimeCommand)
	}
	// return string in the for of redis-compatible command
	return fmt.Sprintf("%s %s on %s %s %s %s", commandPrefix, username, permittedKeys, permittedChannels, passwordCmd, userACL.Value), nil

}

// returns first of the list of passwords configured for the users.
// returns empty string if no password is configured.
func GetUserPassword(username string, users map[string]redisfailoverv1.UserSpec) (string, error) {
	userSpec, ok := users[username]
	if !ok {
		log.Warnf("unable to process \"GetUserPassword\": user %s not found.", username)
		return "", nil
	}
	if len(userSpec.Passwords) == 0 {
		log.Warnf("no password configured for %s user", username)
		return "", nil
	}
	return userSpec.Passwords[0].Value, nil
}

/*
Converts AuthV2 spec of a list of users into a string in a given format determined by "redisCommandMode" input
Inputs:

	whether the caller wants the config to be redis-conf compatible format or cli-command compatible format : (string)
	list of user objects                                                                                    : ([]*redisfailoverv1.User)
	k8s client to retrieve data from kubernetes                                                             : secretk8s.Services

Outputs:

	string                        : Config converted to string in appropriate format
	error                         : if any error is encountered, nil otherwise
*/
func getUsersSpecAs(redisCommandMode string /* "RedisConfigCommand" or "redisRuntimeCommand" */, users map[string]redisfailoverv1.UserSpec) (string, error) {

	var userCreationCmd string
	var usersToProcess string
	var err error

	if users != nil {
		for username, userSpec := range users {

			usersToProcess, err = getUserSpecAs(redisCommandMode, username, userSpec)
			if nil != err {
				return "", fmt.Errorf("Unable to process userspec for %v : %v", username, err)
			}
			userCreationCmd = fmt.Sprintf("%s\n%s", userCreationCmd, usersToProcess)
		}
	}
	return userCreationCmd, nil
}

/*
Converts AuthV2 spec into a string that can be embedded into a redis config file
Inputs:

	list of users whose spec should be converted to config ([]*redisfailoverv1.User)

Outputs:

	string                        : Config converted to string in appropriate format
	error                         : if error is encountered, nil otherwise
*/
func GetAuthSpecAsRedisConf(users map[string]redisfailoverv1.UserSpec) (string, error) {
	return getUsersSpecAs(RedisConfigCommand, users)
}

/*
Converts AuthV2 spec into a string that can be run as commands via a redis client
Inputs:

	list of users whose spec should be converted to config ([]*redisfailoverv1.User)

Outputs:

	string                        : Config converted to string in appropriate format
	error                         : if error is encountered, nil otherwise
*/
func GetAuthSpecAsRedisCliCommands(users map[string]redisfailoverv1.UserSpec) (string, error) {
	return getUsersSpecAs(RedisRuntimeCommand, users)
}
