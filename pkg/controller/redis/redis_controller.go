package redis

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	redisv1alpha1 "redis-operator/pkg/apis/redis/v1alpha1"
	"redis-operator/pkg/resource/configmap"
	"redis-operator/pkg/resource/constant"
	"redis-operator/pkg/resource/deployment"
	"redis-operator/pkg/resource/secret"
	"redis-operator/pkg/resource/service"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_redis")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Redis Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileRedis{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("redis-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Redis
	err = c.Watch(&source.Kind{Type: &redisv1alpha1.Redis{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Redis
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &redisv1alpha1.Redis{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileRedis{}

// ReconcileRedis reconciles a Redis object
type ReconcileRedis struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Redis object and makes changes based on the state read
// and what is in the Redis.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileRedis) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Redis")

	// Fetch the Redis
	redis := &redisv1alpha1.Redis{}
	err := r.client.Get(context.TODO(), request.NamespacedName, redis)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	f := &appsv1.Deployment{}

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: redis.Name, Namespace: redis.Namespace}, f)

	if err != nil && errors.IsNotFound(err) {

		if err := r.createSecret(redis); err != nil {
			reqLogger.Error(err, "Failed to create new Secret", "Namespace", redis.Namespace)
			return reconcile.Result{}, err
		}

		if err := r.createConfigMap(redis); err != nil {
			reqLogger.Error(err, "Failed to create new ConfigMap", "Namespace", redis.Namespace)
			return reconcile.Result{}, err
		}

		if err := r.createDeployment(redis); err != nil {
			reqLogger.Error(err, "Failed to create new Deployment", "Namespace", redis.Namespace)
			return reconcile.Result{}, err
		}
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Deployment")
		return reconcile.Result{}, err
	} else {
		//检查 pod 数量
	}

	if err = r.svc(redis); err != nil {
		reqLogger.Error(err, "Failed to update new Service", "Namespace", redis.Namespace)
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileRedis) createSecret(redis *redisv1alpha1.Redis) error {
	if redis.Spec.Password != nil {
		s := secret.New(redis)
		if err := controllerutil.SetControllerReference(redis, s, r.scheme); err != nil {
			return err
		}
		if err := r.client.Create(context.TODO(), s); err != nil {
			return err
		}
	}
	return nil
}

func (r *ReconcileRedis) createConfigMap(redis *redisv1alpha1.Redis) error {
	if redis.Spec.Config != nil {
		cm := configmap.New(redis)
		if err := controllerutil.SetControllerReference(redis, cm, r.scheme); err != nil {
			return err
		}

		if err := r.client.Create(context.TODO(), cm); err != nil {
			return err
		}
	}
	return nil
}

func (r *ReconcileRedis) createDeployment(redis *redisv1alpha1.Redis) error {

	//查看有没有secret
	sc := &corev1.Secret{}
	err1 := r.client.Get(context.TODO(), types.NamespacedName{Name: constant.SecretName, Namespace: redis.Namespace}, sc)

	//查看 configMap
	cm := &corev1.ConfigMap{}
	err2 := r.client.Get(context.TODO(), types.NamespacedName{Name: constant.ConfigMapName, Namespace: redis.Namespace}, cm)

	dc := deployment.New(redis, err1 == nil, err2 == nil)

	if err := controllerutil.SetControllerReference(redis, dc, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), dc); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileRedis) svc(redis *redisv1alpha1.Redis) error {
	s := &corev1.Service{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: redis.Name, Namespace: redis.Namespace}, s); err != nil && errors.IsNotFound(err) {
		svc := service.New(redis)
		if err := controllerutil.SetControllerReference(redis, svc, r.scheme); err != nil {
			return err
		}
		if err = r.client.Create(context.TODO(), svc); err != nil {
			return err
		}

	} else if err != nil {
		return err
	} else {
		//查看 有没有变动
		ports := redis.Spec.Ports
		if (!reflect.DeepEqual(ports, s.Spec.Ports)) {
			s.Spec.Ports = ports
			if err = r.client.Update(context.TODO(), s); err != nil {
				return err
			}
		}
	}
	return nil
}
