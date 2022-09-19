package authv2

import (
	"fmt"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/log"
)

/*
gets index of a user identified by the username, in a given list of users of type redisfailoverv1.User

inputs:

	username (string)
	users ([]redisfailoverv1.User )

outputs:

	index of the identified user in the list; -1 if user is not found (int)
*/
func getIndexOfUser(username string, users []redisfailoverv1.User) int {
	for idx, user := range users {
		if user.Name == username {
			return idx
		}
	}
	return -1
}

/*
gets user object (referenced) in a given list of users, identified by given username

inputs:

	username (string)
	users ([]redisfailoverv1.User)

outputs:

	reference to the userobject in given slice, nil otherwise
*/
func getUser(username string, users []redisfailoverv1.User) *redisfailoverv1.User {
	for _, user := range users {
		if user.Name == username {
			return &user
		}
	}
	return nil
}

/*
add user to a given list of users

inputs:

	user object (redisfailoverv1.User)
	list of users ([]redisfailoverv1.User)

outputs:

	None
*/
func addUser(users []redisfailoverv1.User, user redisfailoverv1.User) {
	users = append(users, user)
}

/*
checks if a user identified by the username is present in list of users or not

inputs:

	users ([]redisfailoverv1.User)
	username (string)

outputs:

	bool - true if user is found, false otherwise
*/
func hasUser(users []redisfailoverv1.User, username string) bool {
	for _, user := range users {
		if user.Name == username {
			return true
		}
	}
	return false
}

/*
returns user object of "admin" user with default settings
inputs:

	none

outputs:

	reference to user object with default settings of admin user
*/
func getAdminUserWithDefaultSpec() *redisfailoverv1.User {

	return &redisfailoverv1.User{
		Name:      AdminUserName,
		Passwords: []string{defaultAdminUserPassword},
		ACL:       defaultAdminPermissions,
	}
}

/*
	Converts data of type redisfailoverv1.User to string that can be embedded into redis conf

Inputs:

	redisCommandMode (string)   : whether the caller wants the config to be redis-conf compatible format or cli-command compatible format
	*redisfailoverv1.User       : user whos config needs to be converted to string

Outputs:

	string                      : Config converted to string in appropriate format
	error                       : if any error is encountered , nil otherwise.
*/
func getUserSpecAs(redisCommandMode string, user *redisfailoverv1.User) (string, error) {

	passwordCmd := ""
	for _, password := range user.Passwords {
		passwordCmd += fmt.Sprintf(">%s ", password)
	}

	// process ACL
	userACL := user.ACL
	if userACL == "" {
		userACL = defaultUserPermissions
	}

	// placeholder for processing access control keys and channels in spec
	permittedKeys := defaultPermittedKeys
	permittedChannels := defaultPermittedChannels

	commandPrefix := ""
	if redisConfigCommand == redisCommandMode {
		commandPrefix = "user"
	} else if redisRuntimeCommand == redisCommandMode {
		commandPrefix = "acl setuser"
	} else {
		return "", fmt.Errorf("redis command mode not recognised: %s ; accepted modes - %s and %s", redisCommandMode, redisConfigCommand, redisRuntimeCommand)
	}

	return fmt.Sprintf("%s %s on %s %s %s %s", commandPrefix, user.Name, permittedKeys, permittedChannels, passwordCmd, userACL), nil
}

func GetUserPassword(username string, users []redisfailoverv1.User) (string, error) {
	user := getUser(username, users)
	if nil == user {
		log.Warnf("unable to process \"GetUserPassword\": user %s not found.", username)
		return "", nil
	}
	if len(user.Passwords) == 0 {
		log.Warnf("no password configured for %s user")
		return "", nil
	}
	return user.Passwords[0], nil
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
func getUsersSpecAs(redisCommandMode string /* "redisConfigCommand" or "redisRuntimeCommand" */, users []redisfailoverv1.User) (string, error) {

	var userCreationCmd string
	var usersToCreate string
	var err error

	if users != nil {
		for _, user := range users {

			usersToCreate, err = getUserSpecAs(redisCommandMode, &user)
			if nil != err {
				return "", fmt.Errorf("Unable to process userspec for %v : %v", user.Name, err)
			}
			userCreationCmd = fmt.Sprintf("%s\n%s", userCreationCmd, usersToCreate)
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
func GetAuthSpecAsRedisConf(users []redisfailoverv1.User) (string, error) {
	return getUsersSpecAs(redisConfigCommand, users)
}

/*
Converts AuthV2 spec into a string that can be run as commands via a redis client
Inputs:

	list of users whose spec should be converted to config ([]*redisfailoverv1.User)

Outputs:

	string                        : Config converted to string in appropriate format
	error                         : if error is encountered, nil otherwise
*/
func GetAuthSpecAsRedisCliCommands(users []redisfailoverv1.User) (string, error) {
	return getUsersSpecAs(redisRuntimeCommand, users)
}
