package redisfailover_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	redisauth "github.com/spotahome/redis-operator/operator/redisfailover/auth"

	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	redisfailoverv1 "github.com/spotahome/redis-operator/api/redisfailover/v1"
	"github.com/spotahome/redis-operator/log"
	"github.com/spotahome/redis-operator/service/k8s"
)

const (
	namespaceAuthV2 = "rf-integration-tests-authv2"
	adminUser       = "admin"
	adminPassword   = "adminpassword"
)

var (
	redisObject = &redisfailoverv1.RedisFailover{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespaceAuthV2,
		},
		Spec: redisfailoverv1.RedisFailoverSpec{
			Redis: redisfailoverv1.RedisSettings{
				Replicas: redisSize,
				Exporter: redisfailoverv1.Exporter{
					Enabled: true,
				},
			},
			Sentinel: redisfailoverv1.SentinelSettings{
				Replicas: sentinelSize,
			},
			Auth: redisfailoverv1.AuthSettings{
				SecretPath: authSecretPath,
			},
			AuthV2: redisfailoverv1.AuthV2Settings{
				Enabled: true,
				Users: map[string]redisfailoverv1.UserSpec{
					"admin": {
						Passwords: []redisfailoverv1.Password{
							{
								Value: "testPassword",
							},
						},
					},
				},
			},
		},
	}

	adminPasswordSecret = &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      authSecretPath,
			Namespace: namespaceAuthV2,
		},
		Data: map[string][]byte{
			"password": []byte(adminPassword),
		},
	}

	newUserName = "newuser"
	newUserSpec = redisfailoverv1.UserSpec{
		Passwords: []redisfailoverv1.Password{
			{Value: "newpassword"},
		},
	}
	newPassword = "testpasswordaddition"
)

func (c *clients) prepareNSAuthV2() error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceAuthV2,
		},
	}
	_, err := c.k8sClient.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
	return err
}

func (c *clients) cleanupAuthV2(stopC chan struct{}) {
	c.k8sClient.CoreV1().Namespaces().Delete(context.Background(), namespaceAuthV2, metav1.DeleteOptions{})
	close(stopC)
}

func (c *clients) testCRCreationAuthV2(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)
	toCreateAuthV2 := redisObject

	// create namespace
	prepErr := c.prepareNSAuthV2()
	require.NoError(prepErr)

	_, err := c.k8sClient.CoreV1().Secrets(namespaceAuthV2).Create(context.Background(), adminPasswordSecret, metav1.CreateOptions{})
	require.NoError(err)

	c.rfClient.DatabasesV1().RedisFailovers(namespaceAuthV2).Create(context.Background(), toCreateAuthV2, metav1.CreateOptions{})
	gotRF, err := c.rfClient.DatabasesV1().RedisFailovers(namespaceAuthV2).Get(context.Background(), name, metav1.GetOptions{})
	require.NoError(err)

	assert.Equal(toCreateAuthV2.Spec, gotRF.Spec)
}

func (c *clients) testUserCreationAtRuntime(t *testing.T) {

	assert := assert.New(t)

	// Obtain deployed userspec
	toUpdate, err := c.rfClient.DatabasesV1().RedisFailovers(namespaceAuthV2).Get(context.Background(), name, metav1.GetOptions{})
	assert.NoError(err, fmt.Sprintf("Unable to get deployed redisfailover object%v", err))

	// Add new user to authV2 spec
	toUpdate.Spec.AuthV2.Users[newUserName] = newUserSpec
	_, err = c.rfClient.DatabasesV1().RedisFailovers(namespaceAuthV2).Update(context.Background(), toUpdate, metav1.UpdateOptions{})
	assert.NoError(err, fmt.Sprintf("Unable to update redisfailover object with new users: %v", err))

	k8sservice := k8s.New(c.k8sClient, c.rfClient, c.aeClient, log.Dummy)
	authProvider := redisauth.GetAuthProvider(toUpdate, k8sservice)
	username, password, err := authProvider.GetAdminCredentials()
	assert.NoError(err, fmt.Sprintf("Unabled to obtain admin user credentials to make calls to redis: %v", err))

	// Giving time to the operator to create users
	time.Sleep(1 * time.Minute)
	masterIP := getMasterIP(c)

	users, err := c.redisClient.GetUsers(masterIP, "6379", username, password)
	assert.NoError(err, fmt.Sprintf("Unable to get existing users from redis: %v", err))

	userFound := false
	for _, user := range users {
		if strings.Contains(user, newUserName) {
			userFound = true
		}
	}
	assert.True(userFound, "New user is not being updated in redis.")
}

// Delete a user from spec; the user must be removed from redis at runtime as well.
func (c *clients) testUserDeletionAtRuntime(t *testing.T) {

	assert := assert.New(t)

	// Obtain deployed userspec
	toUpdate, err := c.rfClient.DatabasesV1().RedisFailovers(namespaceAuthV2).Get(context.Background(), name, metav1.GetOptions{})
	assert.NoError(err, fmt.Sprintf("Unable to get deployed redisfailover object%v", err))

	// Delete new user from authV2 spec
	delete(toUpdate.Spec.AuthV2.Users, newUserName)
	_, err = c.rfClient.DatabasesV1().RedisFailovers(namespaceAuthV2).Update(context.Background(), toUpdate, metav1.UpdateOptions{})
	assert.NoError(err, fmt.Sprintf("Unable to update redisfailover object with new users: %v", err))

	k8sservice := k8s.New(c.k8sClient, c.rfClient, c.aeClient, log.Dummy)
	authProvider := redisauth.GetAuthProvider(toUpdate, k8sservice)
	username, password, err := authProvider.GetAdminCredentials()
	assert.NoError(err, fmt.Sprintf("Unabled to obtain admin user credentials to make calls to redis: %v", err))

	// Giving time to the operator to delete usersxcm.
	time.Sleep(1 * time.Minute)
	masterIP := getMasterIP(c)

	users, err := c.redisClient.GetUsers(masterIP, "6379", username, password)
	assert.NoError(err, fmt.Sprintf("Unable to get existing users from redis: %v", err))

	userFound := false
	for _, user := range users {
		if strings.Contains(user, newUserName) {
			userFound = true
		}
	}
	assert.False(userFound, "Deleted user is still present in redis.")
}

// Delete a user from spec; the user must be removed from redis at runtime as well.
func (c *clients) testPasswordAdditionAtRuntime(t *testing.T) {

	assert := assert.New(t)

	// Obtain deployed userspec
	toUpdate, err := c.rfClient.DatabasesV1().RedisFailovers(namespaceAuthV2).Get(context.Background(), name, metav1.GetOptions{})
	assert.NoError(err, fmt.Sprintf("Unable to get deployed redisfailover object%v", err))

	// Append new password to existing user's spec
	currentUserSpec := toUpdate.Spec.AuthV2.Users[newUserName]
	updatedPasswords := append(currentUserSpec.Passwords, redisfailoverv1.Password{
		Value: newPassword,
	})
	currentUserSpec.Passwords = updatedPasswords
	toUpdate.Spec.AuthV2.Users[newUserName] = currentUserSpec

	_, err = c.rfClient.DatabasesV1().RedisFailovers(namespaceAuthV2).Update(context.Background(), toUpdate, metav1.UpdateOptions{})
	assert.NoError(err, fmt.Sprintf("Unable to update redisfailover object: %v", err))

	k8sservice := k8s.New(c.k8sClient, c.rfClient, c.aeClient, log.Dummy)
	authProvider := redisauth.GetAuthProvider(toUpdate, k8sservice)
	username, password, err := authProvider.GetAdminCredentials()
	assert.NoError(err, fmt.Sprintf("Unabled to obtain admin user credentials to make calls to redis: %v", err))

	// Giving time to the operator to apply new password for new user
	time.Sleep(1 * time.Minute)
	masterIP := getMasterIP(c)

	users, err := c.redisClient.GetUsers(masterIP, "6379", username, password)
	assert.NoError(err, fmt.Sprintf("Unable to get existing users from redis: %v", err))

	fmt.Printf("users after password update: %v", users)

	newPasswordFound := false
	// search of the given user, and check if the user has new password is added to its ACL in redis.
	for _, user := range users {
		if strings.Contains(user, newUserName) {
			desiredPasswordsHashed := authProvider.GetHashedPasswords(currentUserSpec)
			for _, desiredPassword := range desiredPasswordsHashed {
				if strings.Contains(user, desiredPassword) {
					newPasswordFound = true
				}
			}
		}
	}
	assert.True(newPasswordFound, "New password is not updated in redis.")
}

func getMasterIP(c *clients) string {
	masters := []string{}
	// ToDo: Handle error
	sentinelD, _ := c.k8sClient.AppsV1().Deployments(namespaceAuthV2).Get(context.Background(), fmt.Sprintf("rfs-%s", name), metav1.GetOptions{})

	listOptions := metav1.ListOptions{
		LabelSelector: labels.FormatLabels(sentinelD.Spec.Selector.MatchLabels),
	}
	// ToDo: Handle error
	sentinelPodList, _ := c.k8sClient.CoreV1().Pods(namespaceAuthV2).List(context.Background(), listOptions)

	for _, pod := range sentinelPodList.Items {
		ip := pod.Status.PodIP
		master, _, _ := c.redisClient.GetSentinelMonitor(ip)
		masters = append(masters, master)
	}
	return masters[0]
}
