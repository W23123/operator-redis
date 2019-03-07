package constant

const (
	RedisPort      = 6379
	RedisImage     = "hub.c.163.com/library/redis:latest"
	ConfigMapName  = "redis-config"
	SecretName     = "redis-secret"
	DataVolumeName = "data"
)
