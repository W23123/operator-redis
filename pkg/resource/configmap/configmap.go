package configmap

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	redis "redis-operator/pkg/apis/redis/v1alpha1"
	"redis-operator/pkg/resource/constant"
)

func New(redis *redis.Redis) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constant.ConfigMapName,
			Namespace: redis.Namespace,
		},
		Data: map[string]string{
			"password": *redis.Spec.Config,
		},
	}
}
