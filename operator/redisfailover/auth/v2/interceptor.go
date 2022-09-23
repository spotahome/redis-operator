/* A wrapper over mutators and validators */
package authv2

import (
	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/service/k8s"
)

/*
1) Check if admin user is present in the spec; if yes, update its ACL to set it to desired default

2) in users defined in rf.spec.authv2, convert all secret references to plaintext fields (for uniform operabilisty)

inputs:

 1. list of users from CR    ([]redisfailoverv1.User)
 2. namespace of CR          (string)
 3. k8s service access utils (k8s.Services)

ouputs:
 1. copy of users with mutations ([]redisfailoverv1.User)
 2. error, if any encountered
*/
func InterceptUsers(crUsers /*cr = users from CR */ map[string]redisfailoverv1.UserSpec, namespace string, k8sServices k8s.Services) (map[string]redisfailoverv1.UserSpec, error) {
	// make deep copy of users - *must* not update spec directly - which will lead to perpetual reconciliation cycle
	users := make(map[string]redisfailoverv1.UserSpec, len(crUsers))
	users = crUsers
	// ------- Admin User Mutators -------- //
	// update admin user config
	adminUserSpec, ok := users[AdminUserName]
	if !ok { // add admin user list of users
		users[AdminUserName] = getDefaultAdminUserSpec()
	} else { // enforce admin permissions
		updatePermissionsOfUser(&adminUserSpec, DefaultAdminPermissions)
		users[AdminUserName] = adminUserSpec
	}

	// ------- Default User Mutators -------- //
	// update default user config
	defaultUserSpec, ok := users[DefaultUserName]
	if !ok { // add default user list of users
		users[DefaultUserName] = getDefaultDefaultUserSpec()
	} else { // enforce default permissions
		updatePermissionsOfUser(&defaultUserSpec, DefaultUserPermissions)
		users[DefaultUserName] = defaultUserSpec
	}

	// ------- Pinger User Mutators -------- //
	// update pinger user config
	pingerUserSpec, ok := users[PingerUserName]
	if !ok { // add pinger user list of users
		users[PingerUserName] = getDefaultPingerUserSpec()
	} else { // enforce pinger permissions
		updatePermissionsOfUser(&pingerUserSpec, PingerUserPermissions)
		users[PingerUserName] = pingerUserSpec
	}

	// -------- User spec mutators --------- //
	for _, userSpec := range users {

		err := loadUserConfig(&userSpec, namespace, k8sServices /* used to access secrets */)
		if nil != err {
			return nil, err
		}
	}
	return users, nil
}
