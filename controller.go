/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/time/rate"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	configmaplisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	"MinecraftServerController/internal"
	mcsv1alpha1 "MinecraftServerController/pkg/apis/minecraftserver/v1alpha1"

	listers "MinecraftServerController/pkg/generated/listers/minecraftserver/v1alpha1"

	informers "MinecraftServerController/pkg/generated/informers/externalversions/minecraftserver/v1alpha1"

	clientset "MinecraftServerController/pkg/generated/clientset/versioned"

	samplescheme "k8s.io/sample-controller/pkg/generated/clientset/versioned/scheme"
)

const controllerAgentName = "sample-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Foo"
	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSynced = "Foo synced successfully"
)

// Controller is the controller implementation for Foo resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	sampleclientset clientset.Interface

	deploymentsLister appslisters.DeploymentLister

	statefulsetLister appslisters.StatefulSetLister
	statefulsetSynced cache.InformerSynced

	configMapLister configmaplisters.ConfigMapLister
	configMapSynced cache.InformerSynced

	deploymentsSynced cache.InformerSynced
	//foosLister        listers.FooLister
	//foosSynced        cache.InformerSynced

	minecraftServerLister listers.MinecraftServerLister
	minecraftServerSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new sample controller
func NewController(
	ctx context.Context,
	kubeclientset kubernetes.Interface,
	sampleclientset clientset.Interface,
	deploymentInformer appsinformers.DeploymentInformer,
	statefulsetInformer appsinformers.StatefulSetInformer,

	minecraftServerInformer informers.MinecraftServerInformer) *Controller {
	logger := klog.FromContext(ctx)

	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(samplescheme.AddToScheme(scheme.Scheme))
	logger.V(4).Info("Creating event broadcaster")

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	ratelimiter := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 1000*time.Second),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(50), 300)},
	)

	controller := &Controller{
		kubeclientset:         kubeclientset,
		sampleclientset:       sampleclientset,
		deploymentsLister:     deploymentInformer.Lister(),
		deploymentsSynced:     deploymentInformer.Informer().HasSynced,
		minecraftServerLister: minecraftServerInformer.Lister(),
		minecraftServerSynced: minecraftServerInformer.Informer().HasSynced,
		statefulsetLister:     statefulsetInformer.Lister(),
		statefulsetSynced:     statefulsetInformer.Informer().HasSynced,
		//foosLister:            fooInformer.Lister(),
		//foosSynced:            fooInformer.Informer().HasSynced,
		workqueue: workqueue.NewRateLimitingQueue(ratelimiter),
		recorder:  recorder,
	}

	logger.Info("Setting up event handlers")
	// Set up an event handler for when Foo resources change
	minecraftServerInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueFoo,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueFoo(new)
		},
	})

	// Set up an event handler for when Deployment resources change. This
	// handler will lookup the owner of the given Deployment, and if it is
	// owned by a Foo resource then the handler will enqueue that Foo resource for
	// processing. This way, we don't need to implement custom logic for
	// handling Deployment resources. More info on this pattern:
	// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*appsv1.Deployment)
			oldDepl := old.(*appsv1.Deployment)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {
				// Periodic resync will send update events for all known Deployments.
				// Two different versions of the same Deployment will always have different RVs.
				return
			}
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(ctx context.Context, workers int) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()
	logger := klog.FromContext(ctx)

	// Start the informer factories to begin populating the informer caches
	logger.Info("Starting Foo controller")

	// Wait for the caches to be synced before starting workers
	logger.Info("Waiting for informer caches to sync")

	if ok := cache.WaitForCacheSync(ctx.Done(), c.minecraftServerSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	logger.Info("Starting workers", "count", workers)
	// Launch two workers to process Foo resources
	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, c.runWorker, time.Second)
	}

	logger.Info("Started workers")
	<-ctx.Done()
	logger.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem(ctx context.Context) bool {
	obj, shutdown := c.workqueue.Get()
	logger := klog.FromContext(ctx)

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.syncHandler(ctx, key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		logger.Info("Successfully synced", "resourceName", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *Controller) syncHandler(ctx context.Context, key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	logger := klog.LoggerWithValues(klog.FromContext(ctx), "resourceName", key)

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Foo resource with this namespace/name
	foo, err := c.minecraftServerLister.MinecraftServers(namespace).Get(name)
	if err != nil {
		// The Foo resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("foo '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	mcs, err := c.minecraftServerLister.MinecraftServers(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("minecraftserver '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}

	/*
		deploymentName := foo.Spec.DeploymentName
		if deploymentName == "" {
			// We choose to absorb the error here as the worker would requeue the
			// resource otherwise. Instead, the next time the resource is updated
			// the resource will be queued again.
			utilruntime.HandleError(fmt.Errorf("%s: deployment name must be specified", key))
			return nil
		}


		// Get the deployment with the name specified in Foo.spec
		deployment, err := c.deploymentsLister.Deployments(foo.Namespace).Get(deploymentName)
		// If the resource doesn't exist, we'll create it
		if errors.IsNotFound(err) {
			deployment, err = c.kubeclientset.AppsV1().Deployments(foo.Namespace).Create(context.TODO(), newDeployment(foo), metav1.CreateOptions{})
		}
	*/

	logger.Info("Statefulset", "namespace", mcs.Namespace)
	logger.Info("Statefulset", "name", "google")
	logger.Info("MCS", mcs)

	statefulset, err := c.statefulsetLister.StatefulSets(mcs.Namespace).Get(mcs.Name)
	if err != nil {
		logger.Error(err, "Statefulset error")
	}
	if errors.IsNotFound(err) {
		logger.Info("Creating Statefulset", "namespace", mcs.Namespace)
		statefulset, err = c.kubeclientset.AppsV1().StatefulSets(mcs.Namespace).Create(context.TODO(), newStatefulset(mcs), metav1.CreateOptions{})
		if err != nil {
			logger.Error(err, "Statefulset error")
		}
	}

	/* TO DO
	configmap, err := c.configMapLister.ConfigMaps(mcs.Namespace).Get("cm-" + mcs.Name)
	if errors.IsNotFound(err) {
		configmap, err = c.kubeclientset.CoreV1().ConfigMaps(foo.Namespace).Create(context.TODO(), newConfigMap(mcs), metav1.CreateOptions{})
	}

	*/

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	/* TODO
	if !metav1.IsControlledBy(configmap, foo) {
		msg := fmt.Sprintf(MessageResourceExists, configmap.Name)
		c.recorder.Event(foo, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf("%s", msg)
	}
	*/

	if !metav1.IsControlledBy(statefulset, foo) {
		msg := fmt.Sprintf(MessageResourceExists, statefulset.Name)
		c.recorder.Event(foo, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf("%s", msg)
	}

	/*
		// If the Deployment is not controlled by this Foo resource, we should log
		// a warning to the event recorder and return error msg.
		if !metav1.IsControlledBy(deployment, foo) {
			msg := fmt.Sprintf(MessageResourceExists, deployment.Name)
			c.recorder.Event(foo, corev1.EventTypeWarning, ErrResourceExists, msg)
			return fmt.Errorf("%s", msg)
		}
	*/

	/*
		// If this number of the replicas on the Foo resource is specified, and the
		// number does not equal the current desired replicas on the Deployment, we
		// should update the Deployment resource.
		if foo.Spec.Replicas != nil && *foo.Spec.Replicas != *deployment.Spec.Replicas {
			logger.V(4).Info("Update deployment resource", "currentReplicas", *foo.Spec.Replicas, "desiredReplicas", *deployment.Spec.Replicas)
			deployment, err = c.kubeclientset.AppsV1().Deployments(foo.Namespace).Update(context.TODO(), newDeployment(foo), metav1.UpdateOptions{})
		}
	*/
	if mcs.Spec.Replicas != 1 {
		logger.V(4).Info("Update statefulset resource", "currentReplicas", mcs.Spec.Replicas, "desiredReplicas", 1)
		//mcs.Spec.Statefulset.Replicas = 1
		//statefulset, err = c.kubeclientset.AppsV1().StatefulSets(mcs.Namespace).Update(context.TODO(), newStatefulset(mcs), metav1.UpdateOptions{})
	}

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// Finally, we update the status block of the Foo resource to reflect the
	// current state of the world
	err = c.updateFooStatus(mcs, statefulset)
	if err != nil {
		return err
	}

	c.recorder.Event(foo, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) updateFooStatus(mcs *mcsv1alpha1.MinecraftServer, statefulset *appsv1.StatefulSet) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	fooCopy := mcs.DeepCopy()
	fooCopy.Status.AvailableReplicas = statefulset.Status.AvailableReplicas
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Foo resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := c.sampleclientset.MinecraftserverV1alpha1().MinecraftServers(mcs.Namespace).Update(context.TODO(), fooCopy, metav1.UpdateOptions{})
	//_, err := c.sampleclientset.MinecraftserverV1Alpha1().MinecraftServers(mcs.Namespace).Update(context.TODO(), fooCopy, metav1.UpdateOptions{})
	//_, err := c.sampleclientset.SamplecontrollerV1alpha1().Foos(foo.Namespace).UpdateStatus(context.TODO(), fooCopy, metav1.UpdateOptions{})
	return err
}

// enqueueFoo takes a Foo resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Foo.
func (c *Controller) enqueueFoo(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the Foo resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that Foo resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	logger := klog.FromContext(context.Background())
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		logger.V(4).Info("Recovered deleted object", "resourceName", object.GetName())
	}
	logger.V(4).Info("Processing object", "object", klog.KObj(object))
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a Foo, we should not do anything more
		// with it.
		if ownerRef.Kind != "Foo" {
			return
		}

		foo, err := c.minecraftServerLister.MinecraftServers(object.GetNamespace()).Get(ownerRef.Name)
		//foo, err := c.foosLister.Foos(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			logger.V(4).Info("Ignore orphaned object", "object", klog.KObj(object), "foo", ownerRef.Name)
			return
		}

		c.enqueueFoo(foo)
		return
	}
}

func newConfigMap(mcs *mcsv1alpha1.MinecraftServer) *corev1.ConfigMap {
	properties := internal.NewServerProperties().Template()

	data := make(map[string]string)
	data["properties"] = properties

	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cm-" + mcs.Name,
			Namespace: mcs.Namespace,
		},
		Data: data,
	}
	return cm
}

func newService(mcs *mcsv1alpha1.MinecraftServer) *corev1.Service {
	labels := map[string]string{
		"app": mcs.Name,
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mcs.ObjectMeta.Name,
			Namespace: mcs.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mcs, mcsv1alpha1.SchemeGroupVersion.WithKind("MinecraftServer")),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "minecraft",
					Port: 25565,
				},
			},
			Selector: labels,
		},
	}

}

func newPersistentVolumeClaim(mcs *mcsv1alpha1.MinecraftServer) *corev1.PersistentVolumeClaim {
	diskSize := resource.NewQuantity(5*1000*1000*1000, resource.DecimalSI) // 5Gi

	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kc-" + mcs.Name,
			Namespace: mcs.Namespace,
			Labels: map[string]string{
				"app":        "dude",
				"controller": mcs.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mcs, mcsv1alpha1.SchemeGroupVersion.WithKind("MinecraftServer")),
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: *diskSize,
				},
			},
		},
	}
}

func newStatefulset(mcs *mcsv1alpha1.MinecraftServer) *appsv1.StatefulSet {
	labels := map[string]string{
		"app": mcs.Name,
	}

	if mcs.Spec.Replicas == 0 {
		mcs.Spec.Replicas = 1
	}

	if mcs.Spec.ServiceAccountName == "" {
		mcs.Spec.ServiceAccountName = "default"
	}

	fmt.Println("------------------------------------------------------------")
	fmt.Println("ServerVersion", mcs.Spec.ServerVersion)
	fmt.Println("------------------------------------------------------------")

	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mcs.ObjectMeta.Name,
			Namespace: mcs.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(mcs, mcsv1alpha1.SchemeGroupVersion.WithKind("MinecraftServer")),
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &mcs.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: mcs.Spec.ServiceAccountName,
					InitContainers: []corev1.Container{
						{
							Name:  "init",
							Image: "kubecraft/kubecraft-cli:v0.1.0",
							Command: []string{
								"/kubecraft-cli"},
							Args: []string{
								"--version",
								mcs.Spec.ServerVersion,
								"--output",
								"/data",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "minecraft-data",
									MountPath: "/data",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:    "minecraft",
							Image:   "kubecraft/minecraft-server:v0.1.0",
							Command: []string{"/usr/local/bin/box64", "/data/bedrock_server"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "minecraft-data",
									MountPath: "/data",
								},
								/*
									{
										Name:      "server-properties",
										MountPath: "/data/server.properties",
									},
								*/
							},
						},
					},
					/*
						TODO: Add volumes for server.properties

						Volumes: []corev1.Volume{
							{
								Name: "server-properties",
								VolumeSource: corev1.VolumeSource{
									ConfigMap: &corev1.ConfigMapVolumeSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "cm-" + mcs.Name,
										},
									},
								},
							},
						},
					*/
				},
			},
		},
	}

	emptyPersistentVolumeClaim := corev1.PersistentVolumeClaimSpec{}

	if !cmp.Equal(mcs.Spec.DataVolume.PersistentVolumeClaim, emptyPersistentVolumeClaim) {
		klog.Infof("---- %s Data: PersistentVolumeClaimTemplate provided ----", mcs.ObjectMeta.Name)
		statefulset.Spec.VolumeClaimTemplates = append(
			statefulset.Spec.VolumeClaimTemplates,
			corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "minecraft-data",
				},
				Spec: mcs.Spec.DataVolume.PersistentVolumeClaim,
			},
		)
	} else {
		klog.Infof("---- %s Data: PersistentVolumeClaimTemplate not provided ----", mcs.ObjectMeta.Name)
	}

	if mcs.Spec.DataVolume.Volume != (corev1.Volume{}) {
		klog.Infof("---- %s Data: Volume provided ----", mcs.ObjectMeta.Name)
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, corev1.Volume{
			Name:         "minecraft-data",
			VolumeSource: mcs.Spec.DataVolume.Volume.VolumeSource,
		})
	} else {
		klog.Infof("---- %s Data: Volume not provided ----", mcs.ObjectMeta.Name)
	}

	return statefulset

}
