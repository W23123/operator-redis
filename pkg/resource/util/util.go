package util

func LabelsForRedis(name string) map[string]string {
	return map[string]string{"app": "redis", "redis_cr": name}
}
