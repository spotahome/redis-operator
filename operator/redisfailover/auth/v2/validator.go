package authv2

import (
	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
)

/*
checks if admin user is present in given list of users

inputs:

	none

outputs:

	bool (true if admin user is present; false otherwise )
*/
func isAdminUserPresent(users []redisfailoverv1.User) bool {
	if nil != getUser(AdminUserName, users) {
		return true
	}
	return false
}

/*
parsed CR and decides if authV2 is enabled or not.
returns response as boolean accordingly.
*/
func IsEnabled(rf redisfailoverv1.RedisFailover) bool {
	if rf.Spec.AuthV2.Enabled {
		return true
	}
	return false
}
