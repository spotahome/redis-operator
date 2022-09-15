// used to process user spec; aggregate content from secret(s)
package util

import (
	"encoding/json"
	"fmt"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/service/k8s"
)

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

/*
Converts AuthV2 spec into a string that can be embedded into a redis config file
Inputs:

	*redisfailoverv1.RedisFailover: Custom resource
	k8s.Services                  : k8s client to retrieve data from kubernetes secret

Outputs:

	string                        : Config converted to string in appropriate format
	error                         : if error is encountered, nil otherwise
*/
func GetAuthV2SpecAsRedisConf(k8sServices k8s.Services, rf *redisfailoverv1.RedisFailover) (string, error) {
	var redisFailOver redisfailoverv1.RedisFailover
	redisFailOver = *rf
	if user := getUser(adminUserName, &redisFailOver); user != nil {
		updatePermissionsOfUser(adminUserName, &redisFailOver, defaultAdminPermissions)
		//adminUserSpec, _ = getUserSpecAs(redisConfigCommand, getUser(adminUserName, &redisFailOver))
	} else {
		redisFailOver.Spec.AuthV2.Users = append(redisFailOver.Spec.AuthV2.Users, *getDefaultAdminUserSpec(&redisFailOver))
	}
	authSpec, err := getAuthV2SpecAs(redisConfigCommand, k8sServices, &redisFailOver)
	return fmt.Sprintf("%s", authSpec), err
}

/*
Converts AuthV2 spec into a string that can be run on a redis server
Inputs:

	*redisfailoverv1.RedisFailover: Custom resource
	k8s.Services                  : k8s client to retrieve data from kubernetes secret

Outputs:

	string                        : Config converted to string in appropriate format
	error                         : if any error is encountered, nil otherwise
*/
func GetAuthV2SpecAsRuntimeCommand(k8sServices k8s.Services, rf *redisfailoverv1.RedisFailover) (string, error) {
	var redisFailOver redisfailoverv1.RedisFailover
	redisFailOver = *rf
	if user := getUser(adminUserName, &redisFailOver); user != nil {
		updatePermissionsOfUser(adminUserName, &redisFailOver, defaultAdminPermissions)
	} else {
		redisFailOver.Spec.AuthV2.Users = append(redisFailOver.Spec.AuthV2.Users, *getDefaultAdminUserSpec(&redisFailOver))
	}
	authSpec, err := getAuthV2SpecAs(redisRuntimeCommand, k8sServices, &redisFailOver)
	return fmt.Sprintf("%s", authSpec), err
}

/*
Converts AuthV2 spec into a string in a given format determined by "redisCommandMode" input
Inputs:

	redisCommandMode (string)     : whether the caller wants the config to be redis-conf compatible format or cli-command compatible format
	*redisfailoverv1.RedisFailover: Custom resource
	k8s.Services                  : k8s client to retrieve data from kubernetes secret

Outputs:

	string                        : Config converted to string in appropriate format
	error                         : if any error is encountered, nil otherwise
*/
func getAuthV2SpecAs(redisCommandMode string /* "redisConfigCommand" or "redisRuntimeCommand" */, k8sServices k8s.Services, rf *redisfailoverv1.RedisFailover) (string, error) {

	var userCreationCmd string
	var usersToCreate string
	var err error

	if rf.Spec.AuthV2.Users != nil {
		for _, user := range rf.Spec.AuthV2.Users {
			if !(user.Name != "" && user.Passwords != nil) {
				loadUserConfigFromSecrets(&user, rf.GetObjectMeta().GetNamespace(), k8sServices)
			}
			usersToCreate, err = getUserSpecAs(redisCommandMode, &user)
			if nil != err {
				return "", fmt.Errorf("Unable to process userspec for %v : %v", user, err)
			}
			userCreationCmd = fmt.Sprintf("%s\n%s", userCreationCmd, usersToCreate)
		}
	}
	return userCreationCmd, nil
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
	}

	return fmt.Sprintf("%s %s on %s %s %s %s", commandPrefix, user.Name, permittedKeys, permittedChannels, passwordCmd, userACL), nil
}

/*
Parses user spec and loads user data from secret if necessary.

Assumptions: Plain text spec provided in user field takes precedence over spec provided as a reference to secret

Inputs:

	*redisfailoverv1.User  : user whose spec needs to be parsed
	k8s.Services           : k8s client to fetch secret data

Outputs:

	error                  : if any error is encounted. nil otherwise
*/
func loadUserConfigFromSecrets(user *redisfailoverv1.User, namespace string, k8sServices k8s.Services) error {
	if user.SecretKey != "" && user.SecretName != "" {

		userSpecSecret, err := k8sServices.GetSecret(namespace, user.SecretName)
		if nil != err {
			return fmt.Errorf("Unable to process userspec : %s", err.Error())
		}
		userSpecData, ok := userSpecSecret.Data[user.SecretKey]
		if !ok {
			return fmt.Errorf("Unable to process userspec : secret key %s not found in secret %s", user.SecretKey, user.SecretName)
		}
		err = json.Unmarshal(userSpecData, &user)
		if nil != err {
			return fmt.Errorf("Unable to process userspec : %s", err.Error())
		}
		return nil
	}
	return fmt.Errorf("Could not parse user spec; either plaintext config should be present, or config secret must be specified")
}

func getDefaultAdminUserSpec(rf *redisfailoverv1.RedisFailover) *redisfailoverv1.User {
	var newAdminPasswords []string
	if !hasUser(rf, adminUserName) { // if admin user is already specified by user, then dont override password
		newAdminPasswords = []string{defaultAdminUserPassword}
	}
	return &redisfailoverv1.User{
		Name:      adminUserName,
		Passwords: newAdminPasswords,
		ACL:       defaultAdminPermissions,
	}
}

func hasUser(rf *redisfailoverv1.RedisFailover, username string) bool {
	users := rf.Spec.AuthV2.Users
	for _, user := range users {
		if user.Name == username {
			return true
		}
	}
	return false
}

func addUser(rf *redisfailoverv1.RedisFailover, user redisfailoverv1.User) {
	rf.Spec.AuthV2.Users = append(rf.Spec.AuthV2.Users, user)
}

func getUser(username string, rf *redisfailoverv1.RedisFailover) *redisfailoverv1.User {
	for _, user := range rf.Spec.AuthV2.Users {
		if user.Name == username {
			return &user
		}
	}
	return nil
}

func getIndexOfUser(username string, rf *redisfailoverv1.RedisFailover) int {
	for idx, user := range rf.Spec.AuthV2.Users {
		if user.Name == username {
			return idx
		}
	}
	return -1
}

func updatePermissionsOfUser(username string, rf *redisfailoverv1.RedisFailover, newPermissionSet string) {
	idx := getIndexOfUser(username, rf)
	rf.Spec.AuthV2.Users[idx].ACL = newPermissionSet
}
