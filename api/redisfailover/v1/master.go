package v1

func (r *RedisFailover) MasterName() string {
	if r.Spec.Sentinel.DisableMyMaster {
		return r.Name
	} else {
		return "mymaster"
	}
}
