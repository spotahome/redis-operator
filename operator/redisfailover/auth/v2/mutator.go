/* Mutations to redis CRs w.r.t auth is done here.
 */
package authv2

import (
	"encoding/json"
	"fmt"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/service/k8s"
)

// updates ACL of a user, given a user of type *redisfailoverv1.User
func updatePermissionsOfUser(user *redisfailoverv1.User, newPermissionSet string) {
	user.ACL = newPermissionSet
}

/*
Parses k8s secret and loads user data into plaintext(string) fields

inputs:

	*redisfailoverv1.User  : user whose spec needs to be parsed
	k8s.Services           : k8s client to fetch secret data

outputs:

	error                  : if any error is encountered. nil otherwise
*/
func loadUserConfigFromSecrets(user *redisfailoverv1.User, namespace string, k8sServices k8s.Services) error {
	log.Infof("recieved request to load user spec from secrets for %s user", user.Name)
	userSpecSecret, err := k8sServices.GetSecret(namespace, user.SecretName)
	if nil != err {
		return fmt.Errorf("Unable to process userspec of %v : %s", user.Name, err.Error())
	}
	userSpecData, ok := userSpecSecret.Data[user.SecretKey]
	if !ok {
		return fmt.Errorf("Unable to process userspec of %v: secret key %s not found in secret %s", user.Name, user.SecretKey, user.SecretName)
	}
	err = json.Unmarshal(userSpecData, &user)
	if nil != err {
		return fmt.Errorf("Unable to process userspec : %s", err.Error())
	}
	return nil
}
