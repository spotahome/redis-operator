package redis

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	rediscli "github.com/go-redis/redis"
)

// Client defines the functions neccesary to connect to redis and sentinel to get or set what we nned
type Client interface {
	GetNumberSentinelsInMemory(ip string) (int32, error)
	GetNumberSentinelSlavesInMemory(ip string) (int32, error)
	ResetSentinel(ip string) error
	GetSlaveOf(ip, password string) (string, error)
	IsMaster(ip, password string) (bool, error)
	MonitorRedis(ip, monitor, quorum, password string) error
	MonitorRedisWithPort(ip, monitor, port, quorum, password string) error
	MakeMaster(ip, password string) error
	MakeSlaveOf(ip, masterIP, password string) error
	MakeSlaveOfWithPort(ip, masterIP, masterPort, password string) error
	GetSentinelMonitor(ip string) (string, string, error)
	SetCustomSentinelConfig(ip string, configs []string) error
	SetCustomRedisConfig(ip string, configs []string, password string) error
	SlaveIsReady(ip, password string) (bool, error)
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
		Addr:     fmt.Sprintf("%s:%s", ip, sentinelPort),
		Password: "",
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	info, err := rClient.Info("sentinel").Result()
	if err != nil {
		return 0, err
	}
	if err2 := isSentinelReady(info); err2 != nil {
		return 0, err2
	}
	match := sentinelNumberRE.FindStringSubmatch(info)
	if len(match) == 0 {
		return 0, errors.New("Seninel regex not found")
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
		Addr:     fmt.Sprintf("%s:%s", ip, sentinelPort),
		Password: "",
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	info, err := rClient.Info("sentinel").Result()
	if err != nil {
		return 0, err
	}
	if err2 := isSentinelReady(info); err2 != nil {
		return 0, err2
	}
	match := slaveNumberRE.FindStringSubmatch(info)
	if len(match) == 0 {
		return 0, errors.New("Slaves regex not found")
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
		return errors.New("Sentinels not ready")
	}
	return nil
}

// ResetSentinel sends a sentinel reset * for the given sentinel
func (c *client) ResetSentinel(ip string) error {
	options := &rediscli.Options{
		Addr:     fmt.Sprintf("%s:%s", ip, sentinelPort),
		Password: "",
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	cmd := rediscli.NewIntCmd("SENTINEL", "reset", "*")
	rClient.Process(cmd)
	_, err := cmd.Result()
	if err != nil {
		return err
	}
	return nil
}

// GetSlaveOf returns the master of the given redis, or nil if it's master
func (c *client) GetSlaveOf(ip, password string) (string, error) {
	options := &rediscli.Options{
		Addr:     fmt.Sprintf("%s:%s", ip, redisPort),
		Password: password,
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	info, err := rClient.Info("replication").Result()
	if err != nil {
		return "", err
	}
	match := redisMasterHostRE.FindStringSubmatch(info)
	if len(match) == 0 {
		return "", nil
	}
	return match[1], nil
}

func (c *client) IsMaster(ip, password string) (bool, error) {
	options := &rediscli.Options{
		Addr:     fmt.Sprintf("%s:%s", ip, redisPort),
		Password: password,
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	info, err := rClient.Info("replication").Result()
	if err != nil {
		return false, err
	}
	return strings.Contains(info, redisRoleMaster), nil
}

func (c *client) MonitorRedis(ip, monitor, quorum, password string) error {
	return c.MonitorRedisWithPort(ip, monitor, redisPort, quorum, password)
}

func (c *client) MonitorRedisWithPort(ip, monitor, port, quorum, password string) error {
	options := &rediscli.Options{
		Addr:     fmt.Sprintf("%s:%s", ip, sentinelPort),
		Password: "",
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	cmd := rediscli.NewBoolCmd("SENTINEL", "REMOVE", masterName)
	rClient.Process(cmd)
	// We'll continue even if it fails, the priority is to have the redises monitored
	cmd = rediscli.NewBoolCmd("SENTINEL", "MONITOR", masterName, monitor, port, quorum)
	rClient.Process(cmd)
	_, err := cmd.Result()
	if err != nil {
		return err
	}

	if password != "" {
		cmd = rediscli.NewBoolCmd("SENTINEL", "SET", masterName, "auth-pass", password)
		rClient.Process(cmd)
		_, err = cmd.Result()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *client) MakeMaster(ip string, password string) error {
	options := &rediscli.Options{
		Addr:     fmt.Sprintf("%s:%s", ip, redisPort),
		Password: password,
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	if res := rClient.SlaveOf("NO", "ONE"); res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (c *client) MakeSlaveOf(ip, masterIP, password string) error {
	return c.MakeSlaveOfWithPort(ip, masterIP, redisPort, password)
}

func (c *client) MakeSlaveOfWithPort(ip, masterIP, masterPort, password string) error {
	options := &rediscli.Options{
		Addr:     fmt.Sprintf("%s:%s", ip, redisPort), // this is IP and Port for the RedisFailover redis
		Password: password,
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	if res := rClient.SlaveOf(masterIP, masterPort); res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (c *client) GetSentinelMonitor(ip string) (string, string, error) {
	options := &rediscli.Options{
		Addr:     fmt.Sprintf("%s:%s", ip, sentinelPort),
		Password: "",
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	cmd := rediscli.NewSliceCmd("SENTINEL", "master", masterName)
	rClient.Process(cmd)
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
		Addr:     fmt.Sprintf("%s:%s", ip, sentinelPort),
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

func (c *client) SetCustomRedisConfig(ip string, configs []string, password string) error {
	options := &rediscli.Options{
		Addr:     fmt.Sprintf("%s:%s", ip, redisPort),
		Password: password,
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
	result := rClient.ConfigSet(parameter, value)
	return result.Err()
}

func (c *client) applySentinelConfig(parameter string, value string, rClient *rediscli.Client) error {
	cmd := rediscli.NewStatusCmd("SENTINEL", "set", masterName, parameter, value)
	rClient.Process(cmd)
	return cmd.Err()
}

func (c *client) getConfigParameters(config string) (parameter string, value string, err error) {
	s := strings.Split(config, " ")
	if len(s) < 2 {
		return "", "", fmt.Errorf("configuration '%s' malformed", config)
	}
	return s[0], strings.Join(s[1:], " "), nil
}

func (c *client) SlaveIsReady(ip, password string) (bool, error) {
	options := &rediscli.Options{
		Addr:     fmt.Sprintf("%s:%s", ip, redisPort),
		Password: password,
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	info, err := rClient.Info("replication").Result()
	if err != nil {
		return false, err
	}

	ok := !strings.Contains(info, redisSyncing) &&
		!strings.Contains(info, redisMasterSillPending) &&
		strings.Contains(info, redisLinkUp)

	return ok, nil
}
