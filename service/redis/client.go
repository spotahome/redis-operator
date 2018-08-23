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
	GetSlaveOf(ip string) (string, error)
	IsMaster(ip string) (bool, error)
	MonitorRedis(ip string, monitor string, quorum string) error
	MakeMaster(ip string) error
	MakeSlaveOf(ip string, masterIP string) error
	GetSentinelMonitor(ip string) (string, error)
	SetCustomSentinelConfig(ip string, configs []string) error
	SetCustomRedisConfig(ip string, configs []string) error
}

type client struct{}

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
	redisPort               = "6379"
	sentinelPort            = "26379"
	masterName              = "mymaster"
	sentinelSetCommand      = "SENTINEL set %s %s"
	redisSetCommand         = "CONFIG set %s"
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
func (c *client) GetSlaveOf(ip string) (string, error) {
	options := &rediscli.Options{
		Addr:     fmt.Sprintf("%s:%s", ip, redisPort),
		Password: "",
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

func (c *client) IsMaster(ip string) (bool, error) {
	options := &rediscli.Options{
		Addr:     fmt.Sprintf("%s:%s", ip, redisPort),
		Password: "",
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

func (c *client) MonitorRedis(ip string, monitor string, quorum string) error {
	options := &rediscli.Options{
		Addr:     fmt.Sprintf("%s:%s", ip, sentinelPort),
		Password: "",
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	cmd := rediscli.NewBoolCmd("SENTINEL", "REMOVE", masterName)
	rClient.Process(cmd)
	// We'll continue even if it fails, the priotity is to have the redises monitored
	cmd = rediscli.NewBoolCmd("SENTINEL", "MONITOR", masterName, monitor, redisPort, quorum)
	rClient.Process(cmd)
	_, err := cmd.Result()
	if err != nil {
		return err
	}
	return nil
}

func (c *client) MakeMaster(ip string) error {
	options := &rediscli.Options{
		Addr:     fmt.Sprintf("%s:%s", ip, redisPort),
		Password: "",
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	if res := rClient.SlaveOf("NO", "ONE"); res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (c *client) MakeSlaveOf(ip string, masterIP string) error {
	options := &rediscli.Options{
		Addr:     fmt.Sprintf("%s:%s", ip, redisPort),
		Password: "",
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()
	if res := rClient.SlaveOf(masterIP, redisPort); res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (c *client) GetSentinelMonitor(ip string) (string, error) {
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
		return "", err
	}
	masterIP := res[3].(string)
	return masterIP, nil
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
		setCommand := fmt.Sprintf(sentinelSetCommand, masterName, config)
		if err := c.applyConfig(setCommand, rClient); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) SetCustomRedisConfig(ip string, configs []string) error {
	options := &rediscli.Options{
		Addr:     fmt.Sprintf("%s:%s", ip, redisPort),
		Password: "",
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	defer rClient.Close()

	for _, config := range configs {
		setCommand := fmt.Sprintf(redisSetCommand, config)
		if err := c.applyConfig(setCommand, rClient); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) applyConfig(command string, rClient *rediscli.Client) error {
	sc := strings.Split(command, " ")
	// Required conversion due to language specifications
	// https://golang.org/doc/faq#convert_slice_of_interface
	s := make([]interface{}, len(sc))
	for i, v := range sc {
		s[i] = v
	}

	cmd := rediscli.NewBoolCmd(s...)
	rClient.Process(cmd)
	if _, err := cmd.Result(); err != nil {
		return err
	}
	return nil
}
