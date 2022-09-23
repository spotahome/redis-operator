/* Mutations to redis CRs w.r.t auth is done here.
 */
package authv2

import (
	"encoding/json"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/service/k8s"
)

// updates ACL of a user, given a user of type *redisfailoverv1.User
func updatePermissionsOfUser(userSpec *redisfailoverv1.UserSpec, newPermissionSet string) {
	userSpec.ACL = redisfailoverv1.ACL{Value: newPermissionSet}
}

/*
Parses k8s secret and loads user data into plaintext(string) fields

inputs:

	*redisfailoverv1.User  : user whose spec needs to be parsed
	k8s.Services           : k8s client to fetch secret data

outputs:

	error                  : if any error is encountered. nil otherwise
*/
func loadUserConfig(userSpec *redisfailoverv1.UserSpec, namespace string, k8sServices k8s.Services) error {

	for idx, password := range userSpec.Passwords {
		// if password has value field populated, perfer using it; otherwise load password from Valuefrom
		if password.Value == "" && nil != password.ValueFrom {
			secretName := password.ValueFrom.LocalObjectReference.Name
			secretKey := password.ValueFrom.Key
			secret, err := k8sServices.GetSecret(namespace, secretName)
			if nil != err {
				log.Warnf("unable to read password from secret %v in key %v", secretName, secretKey)
				return err
			}
			userSpec.Passwords[idx].Value = string(secret.Data[secretKey])

		}
		userSpec.Passwords[idx].HashedValue = GetHashedPassword(userSpec.Passwords[idx].Value)
	}

	// if password has value field populated, perfer using it; otherwise load password from Valuefrom
	if userSpec.ACL.Value == "" && nil != userSpec.ACL.ValueFrom {
		var aclValue string
		secretName := userSpec.ACL.ValueFrom.LocalObjectReference.Name
		secretKey := userSpec.ACL.ValueFrom.Key
		secret, err := k8sServices.GetSecret(namespace, secretName)
		if nil != err {
			return err
		}
		err = json.Unmarshal(secret.Data[secretKey], &aclValue)
		userSpec.ACL.Value = aclValue
	}
	return nil
}
