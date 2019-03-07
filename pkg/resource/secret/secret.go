package secret

import (
	"k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"redis-operator/pkg/apis/redis/v1alpha1"
	"redis-operator/pkg/resource/constant"
)

func New(redis *v1alpha1.Redis) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: v12.ObjectMeta{
			Name:      constant.SecretName,
			Namespace: redis.Namespace,
		},

		StringData: map[string]string{
			"password": *redis.Spec.Password,
		},
	}
}
