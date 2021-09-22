/*
Package mocks will have all the mocks of the application.
*/
package mocks // import "github.com/spotahome/redis-operator/mocks"

// Logger mocks
//go:generate mockery --output log --dir ../log --name Logger

// RedisClient mocks
//go:generate mockery --output service/redis --dir ../service/redis --name Client

// K8SClient mocks
//go:generate mockery --output service/k8s --dir ../service/k8s --name Services

// CRD mocks
//go:generate mockery --output service/k8s --dir ../service/k8s --name CRD

// RedisFailover mocks
//go:generate mockery --output operator/redisfailover --dir ../service/k8s --name RedisFailover

// RedisFailover Operator service Checker mocks
//go:generate mockery --output operator/redisfailover/service --dir ../operator/redisfailover/service --name RedisFailoverCheck

// RedisFailover Operator service Client mocks
//go:generate mockery --output operator/redisfailover/service --dir ../operator/redisfailover/service --name RedisFailoverClient

// RedisFailover Operator service Healer mocks
//go:generate mockery --output operator/redisfailover/service --dir ../operator/redisfailover/service --name RedisFailoverHeal
