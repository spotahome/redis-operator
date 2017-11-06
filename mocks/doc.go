/*
Package mocks will have all the mocks of the application.
*/
package mocks // import "github.com/spotahome/redis-operator/mocks"

// EventHandler mocks
//go:generate mockery -output . -dir ../pkg/tpr -name EventHandler

// RedisFailoverClient mocks
//go:generate mockery -output . -dir ../pkg/failover -name RedisFailoverClient

// Logger mocks
//go:generate mockery -output . -dir ../pkg/log -name Logger

// Clock mocks
//go:generate mockery -output . -dir ../pkg/clock -name Clock

// Transformer mocks
//go:generate mockery -output . -dir ../pkg/failover -name Transformer

// Check mocks
//go:generate mockery -output . -dir ../pkg/failover -name RedisFailoverCheck

// redisClient mocks
//go:generate mockery -output . -dir ../pkg/redis -name Client
