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
	ResetSentinel(ip string) error
	GetSlaveOf(ip string) (string, error)
	IsMaster(ip string) (bool, error)
}

type client struct{}

// New returns a redis client
func New() Client {
	return &client{}
}

const (
	sentinelsNumberREString = "sentinels=([0-9]+)"
	sentinelStatusREString  = "status=([a-z]+)"
	redisMasterHostREString = "master_host:([0-9.]+)"
	redisRoleMaster         = "role:master"
	redisPort               = "6379"
	sentinelPort            = "26379"
)

var (
	sentinelNumberRE  = regexp.MustCompile(sentinelsNumberREString)
	sentinelStatusRE  = regexp.MustCompile(sentinelStatusREString)
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
	info, err := rClient.Info("sentinel").Result()
	rClient.Close()
	if err != nil {
		return 0, err
	}
	match := sentinelNumberRE.FindStringSubmatch(info)
	if len(match) == 0 {
		return 0, errors.New("Seninel regex not found")
	}
	matchStatus := sentinelStatusRE.FindStringSubmatch(info)
	if len(matchStatus) == 0 || matchStatus[1] != "ok" {
		return 0, errors.New("Sentinels not ready")
	}
	nSentinels, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, err
	}
	return int32(nSentinels), nil
}

// ResetSentinel sends a sentinel reset * for the given sentinel
func (c *client) ResetSentinel(ip string) error {
	options := &rediscli.Options{
		Addr:     fmt.Sprintf("%s:%s", ip, sentinelPort),
		Password: "",
		DB:       0,
	}
	rClient := rediscli.NewClient(options)
	cmd := rediscli.NewIntCmd("SENTINEL", "reset", "*")
	rClient.Process(cmd)
	rClient.Close()
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
	info, err := rClient.Info("replication").Result()
	rClient.Close()
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
	info, err := rClient.Info("replication").Result()
	rClient.Close()
	if err != nil {
		return false, err
	}
	return strings.Contains(info, redisRoleMaster), nil
}
