package service

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"redis-operator/pkg/apis/redis/v1alpha1"
	"redis-operator/pkg/resource/util"
)

func New(redis *v1alpha1.Redis) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      redis.Name,
			Namespace: redis.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports:    addServicePort(redis),
			Selector: util.LabelsForRedis(redis.Name),
		},
	}
}

func addServicePort(redis *v1alpha1.Redis) []corev1.ServicePort {
	if redis.Spec.Ports != nil {
		return redis.Spec.Ports
	} else {
		return []corev1.ServicePort{
			{
				Protocol:   corev1.ProtocolTCP,
				Port:       6379,
				TargetPort: intstr.FromInt(6379),
			},
		}
	}
}
