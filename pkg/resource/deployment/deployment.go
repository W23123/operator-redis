package deployment

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	redis "redis-operator/pkg/apis/redis/v1alpha1"
	"redis-operator/pkg/resource/constant"
	"redis-operator/pkg/resource/util"
)

func New(redis *redis.Redis, secret bool, configmap bool) *appsv1.Deployment {
	labels := util.LabelsForRedis(redis.Name)

	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      redis.Name,
			Namespace: redis.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: newContainers(redis, secret, configmap),
					Volumes:    newVolumes(redis, configmap),
				},
			},
		},
	}
	return dep
}

func newVolumes(redis *redis.Redis, configmap bool) []corev1.Volume {

	volumes := make([]corev1.Volume, 0)

	if configmap {
		volumes = append(volumes, corev1.Volume{
			Name: constant.ConfigMapName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: constant.ConfigMapName,
					},
					Items: []corev1.KeyToPath{
						{
							Key:  "password",
							Path: "redis.conf",
						},
					},
				},
			},
		})
	}

	if redis.Spec.Volume != nil {
		volumes = append(volumes, *redis.Spec.Volume)
	} else {
		volumes = append(volumes, corev1.Volume{
			Name: constant.DataVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	return volumes
}

func newContainers(redis *redis.Redis, secret bool, configmap bool) []corev1.Container {

	return [] corev1.Container{
		{
			Image:   constant.RedisImage,
			Command: getCommand(redis, secret, configmap),
			Name:    "redis",
			Env:     getEnv(secret),
			Ports: []corev1.ContainerPort{{
				ContainerPort: constant.RedisPort,
				Name:          "redis",
			}},
			VolumeMounts: getVolumeMounts(redis, configmap),
		},
	}
}

func getVolumeMounts(redis *redis.Redis, configmap bool) []corev1.VolumeMount {

	vm := make([]corev1.VolumeMount, 0)

	if configmap {
		vm = append(vm, corev1.VolumeMount{
			Name:      constant.ConfigMapName,
			MountPath: "/usr/local/etc/redis",
		})
	}

	if redis.Spec.Volume != nil {
		vm = append(vm, corev1.VolumeMount{
			Name:      redis.Spec.Volume.Name,
			MountPath: "/data",
		})
	} else {
		vm = append(vm, corev1.VolumeMount{
			Name:      constant.DataVolumeName,
			MountPath: "/data",
		})
	}
	return vm

}

func getEnv(secret bool) []corev1.EnvVar {

	env := make([]corev1.EnvVar, 0)

	if secret {
		env = append(env, corev1.EnvVar{
			Name: "REDIS_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: "password",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: constant.SecretName,
					},
				},
			},
		})
	}

	return env
}

func getCommand(redis *redis.Redis, secret bool, configmap bool) []string {

	commands := make([]string, 0)

	command := "redis-server  "

	if configmap {
		command = command + "/usr/local/etc/redis/redis.conf  "
	}

	if secret {
		command = command + "--requirepass \"${REDIS_PASSWORD}\" "
	}

	command = command + "--save 900"
	commands = append(commands, "/bin/sh", "-c", command)
	return commands
}
