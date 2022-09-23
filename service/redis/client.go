package redis

import (
	"context"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	rediscli "github.com/go-redis/redis/v8"
	"github.com/spotahome/redis-operator/log"
)

// Client defines the functions neccesary to connect to redis and sentinel to get or set what we nned
type Client interface {
	GetNumberSentinelsInMemory(ip string) (int32, error)
	GetNumberSentinelSlavesInMemory(ip string) (int32, error)
	ResetSentinel(ip string) error
	GetSlaveOf(ip, port, username string, password string) (string, error)
	IsMaster(ip, port, username string, password string) (bool, error)
	MonitorRedis(ip, monitor, quorum, username string, password string) error
	MonitorRedisWithPort(ip, monitor, port, quorum, username string, password string) error
	MakeMaster(ip, port, username string, password string) error
	MakeSlaveOf(ip, masterIP, username string, password string) error
	MakeSlaveOfWithPort(ip, masterIP, masterPort, username string, password string) error
	GetSentinelMonitor(ip string) (string, string, error)
	SetCustomSentinelConfig(ip string, configs []string) error
	SetCustomRedisConfig(ip string, port string, configs []string, username string, password string) error
	SlaveIsReady(ip, port, username string, password string) (bool, error)
	// acl user management
	GetUsers(ip /* IP Of Master Instance */ string, port /* Port Of Master Instance */ string, username /* Admin Username */ string, password /* Admin Password */ string) ([]string /* Redis client output (un processed) */, error)
	DeleteUser(ip, port, adminUsername string, adminPassword string, username string) error
	ACLSetUser(ip, port, adminUsername string, adminPassword string, username string, permissionSpaces []string, passwords []string, permissions []string) error
}

type client struct {
}

// New returns a redis client
func New() Client {
	return &client{}
}

const (
	sentinelsNumberREString = "sentinels=([0-9]+)"
	slaveNumberREString     = "slaves=([0-9]+)"
	sentinelStatusREString  = "status=([a-z]+)"
	redisMasterHostREString = "master_host:([0-9.]+)"
	redisRoleMaster         = "role:master"
	redisSyncing            = "master_sync_in_progress:1"
	redisMasterSillPending  = "master_host:127.0.0.1"
	redisLinkUp             = "master_link_status:up"
	redisPort               = "6379"
	sentinelPort            = "26379"
	masterName              = "mymaster"
)

var (
	sentinelNumberRE  = regexp.MustCompile(sentinelsNumberREString)
	sentinelStatusRE  = regexp.MustCompile(sentinelStatusREString)
	slaveNumberRE     = regexp.MustCompile(slaveNumberREString)
	redisMasterHostRE = regexp.MustCompile(redisMasterHostREString)
)

// GetNumberSentinelsInMemory return the number of sentinels that the requested sentinel has
func (c *client) GetNumberSentinelsInMemory(ip string) (int32, error) {
	options := &rediscli.Options{
		Addr:     net.JoinHostPort(ip, sentinelPort),
		Password: "",
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	info, err := rClient.Info(context.TODO(), "sentinel").Result()
	if err != nil {
		return 0, err
	}
	if err2 := isSentinelReady(info); err2 != nil {
		return 0, err2
	}
	match := sentinelNumberRE.FindStringSubmatch(info)
	if len(match) == 0 {
		return 0, errors.New("seninel regex not found")
	}
	nSentinels, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, err
	}
	return int32(nSentinels), nil
}

// GetNumberSentinelsInMemory return the number of sentinels that the requested sentinel has
func (c *client) GetNumberSentinelSlavesInMemory(ip string) (int32, error) {
	options := &rediscli.Options{
		Addr:     net.JoinHostPort(ip, sentinelPort),
		Password: "",
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	info, err := rClient.Info(context.TODO(), "sentinel").Result()
	if err != nil {
		return 0, err
	}
	if err2 := isSentinelReady(info); err2 != nil {
		return 0, err2
	}
	match := slaveNumberRE.FindStringSubmatch(info)
	if len(match) == 0 {
		return 0, errors.New("slaves regex not found")
	}
	nSlaves, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, err
	}
	return int32(nSlaves), nil
}

func isSentinelReady(info string) error {
	matchStatus := sentinelStatusRE.FindStringSubmatch(info)
	if len(matchStatus) == 0 || matchStatus[1] != "ok" {
		return errors.New("sentinels not ready")
	}
	return nil
}

// ResetSentinel sends a sentinel reset * for the given sentinel
func (c *client) ResetSentinel(ip string) error {
	options := &rediscli.Options{
		Addr:     net.JoinHostPort(ip, sentinelPort),
		Password: "",
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	cmd := rediscli.NewIntCmd(context.TODO(), "SENTINEL", "reset", "*")
	err := rClient.Process(context.TODO(), cmd)
	if err != nil {
		return err
	}
	_, err = cmd.Result()
	if err != nil {
		return err
	}
	return nil
}

// GetSlaveOf returns the master of the given redis, or nil if it's master
func (c *client) GetSlaveOf(ip, port, adminUsername string, adminPassword string) (string, error) {
	options := &rediscli.Options{
		Addr:     net.JoinHostPort(ip, port),
		Username: adminUsername,
		Password: adminPassword,
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	info, err := rClient.Info(context.TODO(), "replication").Result()
	if err != nil {
		return "", err
	}
	match := redisMasterHostRE.FindStringSubmatch(info)
	if len(match) == 0 {
		return "", nil
	}
	return match[1], nil
}

func (c *client) IsMaster(ip, port, adminUsername string, adminPassword string) (bool, error) {
	options := &rediscli.Options{
		Addr:     net.JoinHostPort(ip, port),
		Username: adminUsername,
		Password: adminPassword,
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	info, err := rClient.Info(context.TODO(), "replication").Result()
	if err != nil {
		return false, err
	}
	return strings.Contains(info, redisRoleMaster), nil
}

func (c *client) MonitorRedis(ip, monitor, quorum, adminUsername string, adminPassword string) error {
	return c.MonitorRedisWithPort(ip, monitor, redisPort, quorum, adminUsername, adminPassword)
}

func (c *client) MonitorRedisWithPort(ip, monitor, port, quorum, adminUsername string, adminPassword string) error {
	options := &rediscli.Options{
		Addr:     net.JoinHostPort(ip, sentinelPort),
		Password: "",
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	cmd := rediscli.NewBoolCmd(context.TODO(), "SENTINEL", "REMOVE", masterName)
	_ = rClient.Process(context.TODO(), cmd)
	// We'll continue even if it fails, the priority is to have the redises monitored
	cmd = rediscli.NewBoolCmd(context.TODO(), "SENTINEL", "MONITOR", masterName, monitor, port, quorum)
	err := rClient.Process(context.TODO(), cmd)
	if err != nil {
		return err
	}
	_, err = cmd.Result()
	if err != nil {
		return err
	}
	// set password
	if adminPassword != "" {

		cmd = rediscli.NewBoolCmd(context.TODO(), "SENTINEL", "SET", masterName, "auth-pass", adminPassword)
		err := rClient.Process(context.TODO(), cmd)
		if err != nil {
			return err
		}
		_, err = cmd.Result()
		if err != nil {
			return err
		}
	}
	// set user
	if adminUsername != "" {

		cmd = rediscli.NewBoolCmd(context.TODO(), "SENTINEL", "SET", masterName, "auth-user", adminUsername)
		err = rClient.Process(context.TODO(), cmd)
		if err != nil {
			return err
		}
		_, err = cmd.Result()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *client) MakeMaster(ip string, port string, adminUsername string, adminPassword string) error {
	options := &rediscli.Options{
		Addr:     net.JoinHostPort(ip, port),
		Username: adminUsername,
		Password: adminPassword,
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	if res := rClient.SlaveOf(context.TODO(), "NO", "ONE"); res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (c *client) MakeSlaveOf(ip, masterIP, adminUsername string, adminPassword string) error {
	return c.MakeSlaveOfWithPort(ip, masterIP, redisPort, adminUsername, adminPassword)
}

func (c *client) MakeSlaveOfWithPort(ip, masterIP, masterPort, adminUsername string, adminPassword string) error {
	options := &rediscli.Options{
		Addr:     net.JoinHostPort(ip, masterPort), // this is IP and Port for the RedisFailover redis
		Username: adminUsername,
		Password: adminPassword,
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	if res := rClient.SlaveOf(context.TODO(), masterIP, masterPort); res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (c *client) GetSentinelMonitor(ip string) (string, string, error) {
	options := &rediscli.Options{
		Addr:     net.JoinHostPort(ip, sentinelPort),
		Password: "",
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	cmd := rediscli.NewSliceCmd(context.TODO(), "SENTINEL", "master", masterName)
	err := rClient.Process(context.TODO(), cmd)
	if err != nil {
		return "", "", err
	}
	res, err := cmd.Result()
	if err != nil {
		return "", "", err
	}
	masterIP := res[3].(string)
	masterPort := res[5].(string)
	return masterIP, masterPort, nil
}

func (c *client) SetCustomSentinelConfig(ip string, configs []string) error {
	options := &rediscli.Options{
		Addr:     net.JoinHostPort(ip, sentinelPort),
		Password: "",
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()

	for _, config := range configs {
		param, value, err := c.getConfigParameters(config)
		if err != nil {
			return err
		}
		if err := c.applySentinelConfig(param, value, rClient); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) SetCustomRedisConfig(ip string, port string, configs []string, adminUsername string, adminPassword string) error {
	options := &rediscli.Options{
		Addr:     net.JoinHostPort(ip, port),
		Username: adminUsername,
		Password: adminPassword,
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()

	for _, config := range configs {
		param, value, err := c.getConfigParameters(config)
		if err != nil {
			return err
		}
		if err := c.applyRedisConfig(param, value, rClient); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) applyRedisConfig(parameter string, value string, rClient *rediscli.Client) error {
	result := rClient.ConfigSet(context.TODO(), parameter, value)
	return result.Err()
}

func (c *client) applySentinelConfig(parameter string, value string, rClient *rediscli.Client) error {
	cmd := rediscli.NewStatusCmd(context.TODO(), "SENTINEL", "set", masterName, parameter, value)
	err := rClient.Process(context.TODO(), cmd)
	if err != nil {
		return err
	}
	return cmd.Err()
}

func (c *client) getConfigParameters(config string) (parameter string, value string, err error) {
	s := strings.Split(config, " ")
	if len(s) < 2 {
		return "", "", fmt.Errorf("configuration '%s' malformed", config)
	}
	return s[0], strings.Join(s[1:], " "), nil
}

func (c *client) SlaveIsReady(ip, port, adminUsername string, adminPassword string) (bool, error) {
	options := &rediscli.Options{
		Addr:     net.JoinHostPort(ip, port),
		Username: adminUsername,
		Password: adminPassword,
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	info, err := rClient.Info(context.TODO(), "replication").Result()
	if err != nil {
		return false, err
	}

	ok := !strings.Contains(info, redisSyncing) &&
		!strings.Contains(info, redisMasterSillPending) &&
		strings.Contains(info, redisLinkUp)

	return ok, nil
}

// returns list of users in for of []string, where each string is of the format provided by `acl list` redis command
func (c *client) GetUsers(ip, port, adminUsername string, adminPassword string) ([]string, error) {
	options := &rediscli.Options{
		Addr:     net.JoinHostPort(ip, port),
		Username: adminUsername,
		Password: adminPassword,
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	items, err := rClient.Do(context.TODO(), "acl", "list").Result()
	if err != nil {
		return nil, err
	}
	usersAsString := []string{}
	for _, user := range items.([]interface{}) {
		usersAsString = append(usersAsString, user.(string))
	}
	return usersAsString, nil
}

// Deletes given user from redis instance
func (c *client) DeleteUser(ip, port, adminUsername string, adminPassword string, username string) error {
	options := &rediscli.Options{
		Addr:     net.JoinHostPort(ip, port),
		Username: adminUsername,
		Password: adminPassword,
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	result, err := rClient.Do(context.TODO(), "acl", "deluser", username).Result()
	log.Debugf("delete user %v resulted with: %v", username, result)
	return err
}

// applies ACL to given user, equivalent to `ACL SETUSER redis command
// `permissionSpaces` are keyspaces and channels
func (c *client) ACLSetUser(ip, port, adminUsername string, adminPassword string, username string, permissionSpaces []string, passwords []string, permissions []string) error {
	options := &rediscli.Options{
		Addr:     net.JoinHostPort(ip, port),
		Username: adminUsername,
		Password: adminPassword,
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()

	cmdPrefix := []string{"acl", "setuser", username, "on"}
	cmdList := append(append(append(cmdPrefix, permissionSpaces...), passwords...), permissions...)
	var aclSetUserCmd []interface{} = make([]interface{}, len(cmdList))

	for i, aclSetting := range cmdList {
		aclSetUserCmd[i] = aclSetting
	}
	_, err := rClient.Do(context.Background(), aclSetUserCmd...).Result()
	if nil != err {
		log.Errorf("Error: %v", err)
	}

	return nil
}
