package v1alpha2

import (
	"errors"
)

// Validate set the values by default if not defined and checks if the values given are valid
func (r *RedisFailover) Validate() error {
	if r.Spec.Redis.Replicas == 0 {
		r.Spec.Redis.Replicas = defaultRedisNumber
	} else if r.Spec.Redis.Replicas < defaultRedisNumber {
		return errors.New("number of redises in spec is less than the minimum")
	}

	if r.Spec.Sentinel.Replicas == 0 {
		r.Spec.Sentinel.Replicas = defaultSentinelNumber
	} else if r.Spec.Sentinel.Replicas < defaultSentinelNumber {
		return errors.New("number of sentinels in spec is less than the minimum")
	}

	if r.Spec.Redis.Image == "" {
		r.Spec.Redis.Image = defaultRedisImage
	}

	if r.Spec.Redis.Version == "" {
		r.Spec.Redis.Version = defaultRedisImageVersion
	}

	if r.Spec.Redis.ExporterImage == "" {
		r.Spec.Redis.ExporterImage = defaultExporterImage
	}

	if r.Spec.Redis.ExporterVersion == "" {
		r.Spec.Redis.ExporterVersion = defaultExporterImageVersion
	}

	if len(r.Spec.Sentinel.CustomConfig) == 0 {
		r.Spec.Sentinel.CustomConfig = defaultSentinelCustomConfig
	}

	return nil
}
