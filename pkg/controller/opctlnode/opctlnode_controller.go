/*
Copyright 2018 Zach Puckett.

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

package opctlnode

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"

	opctlv1beta1 "github.com/zachpuck/opctl-operator/pkg/apis/opctl/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

const (
	// SuccessSynced is used as part of the Event 'reason' when a OpctlNode is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a OpctlNode fails
	// to sync due to a resource of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a resource already existing
	MessageResourceExists = "Resource %q already exists and is not managed by OpctlNode"
	// MessageResourceSynced is the message used for an Event fired when a OpctlNode
	// is synced successfully
	MessageResourceSynced = "OpctlNode synced successfully"
)

const (
	OpctlNodePort       = 42224
	OpctlNodeReplicas   = 1
	OpctlNodePullPolicy = "Always"
)

// Add creates a new OpctlNode Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
// USER ACTION REQUIRED: update cmd/manager/main.go to call this opctl.Add(mgr) to install this Controller
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileOpctlNode{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("opctlnode-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to OpctlNode
	err = c.Watch(&source.Kind{Type: &opctlv1beta1.OpctlNode{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by OpctlNode - change this for objects you create
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &opctlv1beta1.OpctlNode{},
	})
	if err != nil {
		return err
	}

	// Watch a Service created by OpctlNode
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &opctlv1beta1.OpctlNode{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileOpctlNode{}

// ReconcileOpctlNode reconciles a OpctlNode object
type ReconcileOpctlNode struct {
	client.Client
	record.EventRecorder
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a OpctlNode object and makes changes based on the state read
// and what is in the OpctlNode.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=opctl.opctloperator.opctl.io,resources=opctlnodes,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileOpctlNode) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the OpctlNode instance
	opctlNode := &opctlv1beta1.OpctlNode{}
	err := r.Get(context.TODO(), request.NamespacedName, opctlNode)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	deploy, err := r.newDeployment(opctlNode)
	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could bave been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		glog.Errorf("Error creating deployment: %s", err)
		return reconcile.Result{}, err
	}

	service, err := r.newService(opctlNode)
	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		glog.Errorf("Error creating service: %s", err)
		return reconcile.Result{}, err
	}

	// Finally, we update the status block of the OpctlNode resource to reflect the
	// current state of the world
	// fmt.Println(deploy, service)
	err = r.updateOpctlNodeStatus(opctlNode, deploy, service)
	if err != nil {
		return reconcile.Result{}, err
	}

	// r.Event(opctlNode, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)

	return reconcile.Result{}, nil
}

// update status fields of the opctl node object and emit events
func (r *ReconcileOpctlNode) updateOpctlNodeStatus(opctlNode *opctlv1beta1.OpctlNode, deployment *appsv1.Deployment, service *corev1.Service) error {
	// func (r *ReconcileOpctlNode) updateOpctlNodeStatus(opctlNode *opctlv1beta1.OpctlNode) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	opctlNodeCopy := opctlNode.DeepCopy()

	// TODO: update other status fields (ex: apiEndpoint)
	opctlNodeCopy.Status.Phase = "Ready"

	// update the Status block of the OpctlNode resource
	var err error
	err = r.Client.Update(context.TODO(), opctlNodeCopy)
	return err
}

// newDeployment creates a new Deployment for an OpctlNode resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the OpctlNode resource that 'owns' it.
func (r *ReconcileOpctlNode) newDeployment(opctlNode *opctlv1beta1.OpctlNode) (*appsv1.Deployment, error) {
	// Get the deployment with the name specified in OpctlNode.spec
	deployment := &appsv1.Deployment{}
	err := r.Client.Get(
		context.TODO(), types.NewNamespacedNameFromString(
			fmt.Sprintf("%s%c%s", opctlNode.GetNamespace(), types.Separator,
				opctlNode.GetName())),
		deployment)

	// if the resource doesn't exist, we'll create it
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}
	} else {
		// if the Deployment is not controlled by this OpctlNode resource, we should log
		// a warning to the event recorder and return
		if !metav1.IsControlledBy(deployment, opctlNode) {
			msg := fmt.Sprintf(MessageResourceExists, deployment.GetName())
			r.Event(opctlNode, corev1.EventTypeWarning, ErrResourceExists, msg)
			return deployment, fmt.Errorf(msg)
		}

		return deployment, nil
	}

	labels := map[string]string{
		"app":        "opctl",
		"controller": opctlNode.GetName(),
		"component":  string(opctlNode.UID),
	}

	// create environment variables
	var env []corev1.EnvVar

	// create security context variable
	privilegedContainer := true
	securityContext := &corev1.SecurityContext{
		Privileged: &privilegedContainer,
	}

	var replicas int32 = OpctlNodeReplicas
	deployment = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        opctlNode.GetName(),
			Namespace:   opctlNode.GetNamespace(),
			Annotations: opctlNode.Spec.Annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "opctl",
							Image: opctlNode.Spec.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "node",
									ContainerPort: OpctlNodePort,
									HostPort:      OpctlNodePort,
									Protocol:      "TCP",
								},
							},
							Env:             env,
							ImagePullPolicy: OpctlNodePullPolicy,
							// TODO: How can we have opctl not need privilaged access? (possible run k8s jobs instead of container?)
							SecurityContext: securityContext,
						},
					},
				},
			},
		},
	}

	err = controllerutil.SetControllerReference(opctlNode, deployment, r.scheme)
	if err != nil {
		return nil, err
	}

	err = r.Client.Create(context.TODO(), deployment)
	return deployment, err
}

// newService creates a new Service for an OpctlNode resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the OpctlNode resources that 'owns' it.
func (r *ReconcileOpctlNode) newService(opctlNode *opctlv1beta1.OpctlNode) (*corev1.Service, error) {
	// Get the service with the name specified in OpctlNode.spec
	serviceName := opctlNode.GetName()
	if opctlNode.Spec.Service != nil && opctlNode.Spec.Service.Name != "" {
		serviceName = opctlNode.Spec.Service.Name
	}
	service := &corev1.Service{}
	err := r.Client.Get(
		context.TODO(),
		types.NewNamespacedNameFromString(
			fmt.Sprintf("%s%c%s", opctlNode.GetNamespace(), types.Separator, serviceName)),
		service)
	// If the resource doesn't exist, we'll create it
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}
	} else {
		// If the Service is not controlled by this OpctlNode resource, we should log
		// a warning to the event recorder and return
		if !metav1.IsControlledBy(service, opctlNode) {
			msg := fmt.Sprintf(MessageResourceExists, service.GetName())
			r.Event(opctlNode, corev1.EventTypeWarning, ErrResourceExists, msg)
			return service, fmt.Errorf(msg)
		}

		return service, nil
	}

	labels := map[string]string{
		"app":        "opctl",
		"controller": opctlNode.GetName(),
		"component":  string(opctlNode.UID),
	}

	service = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: opctlNode.GetNamespace(),
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "node",
					Protocol: "TCP",
					Port:     OpctlNodePort,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: OpctlNodePort,
					},
				},
			},
			Selector: map[string]string{
				"component": string(opctlNode.UID),
			},
		},
	}

	if opctlNode.Spec.Service != nil {
		service.ObjectMeta.Annotations = opctlNode.Spec.Service.Annotations
		if opctlNode.Spec.Service.ServiceType != "" {
			service.Spec.Type = opctlNode.Spec.Service.ServiceType
		}
	}

	err = controllerutil.SetControllerReference(opctlNode, service, r.scheme)
	if err != nil {
		return nil, err
	}

	err = r.Client.Create(context.TODO(), service)
	return service, nil
}
