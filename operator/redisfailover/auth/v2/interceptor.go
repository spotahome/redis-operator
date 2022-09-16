/* A wrapper over mutators and validators */
package authv2

import (
	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/service/k8s"
)

/*
1) Check if admin user is present in the spec; if yes, update its ACL to set it to desired default

2) in users defined in rf.spec.authv2, convert all secret references to plaintext fields (for uniform operability)

inputs:

 1. list of users from CR    ([]redisfailoverv1.User)
 2. namespace of CR          (string)
 3. k8s service access utils (k8s.Services)

ouputs:
 1. copy of users with mutations ([]redisfailoverv1.User)
 2. error, if any encountered
*/
func InterceptUsers(crUsers /*cr = users from CR */ []redisfailoverv1.User, namespace string, k8sServices k8s.Services) ([]redisfailoverv1.User, error) {
	// make deep copy of users - *must* not update spec directly - which will lead to perpetual reconciliation cycle
	users := make([]redisfailoverv1.User, len(crUsers))
	copy(users, crUsers)
	log.Infof("Users [before interception]: %s", users)
	// ------- Admin User Mutators -------- //
	// update admin user config
	adminUser := getUser(adminUserName, users)
	if nil != adminUser {
		updatePermissionsOfUser(adminUser, defaultAdminPermissions)
	} else { // add admin user list of users
		addUser(users, *getAdminUserWithDefaultSpec())
	}
	log.Infof("Users [post admin user mutation]: %s", users)
	// -------- User spec mutators --------- //
	for idx, user := range users {
		if user.Name == "" && user.Passwords == nil {
			err := loadUserConfigFromSecrets(&users[idx], namespace, k8sServices /* used to access secrets */)
			if nil != err {
				return nil, err
			}
		}
	}
	log.Infof("Users [post secrets load mutation]: %s", users)
	return users, nil
}
