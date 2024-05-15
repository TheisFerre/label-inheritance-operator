/*
Copyright 2024.

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

package controller

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	labelsv1 "github.com/theisferre/label-inheritance-operator/api/v1"
)

// InheritorReconciler reconciles a Inheritor object
type InheritorReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=labels.theisferre,resources=inheritors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=labels.theisferre,resources=inheritors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=labels.theisferre,resources=inheritors/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
//+kubebuilder:rbac:groups="*",resources="*",verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Inheritor object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *InheritorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	inheritor := &labelsv1.Inheritor{}
	log.Info("Reconciling Inheritor", "name", req.Name)
	err := r.Client.Get(ctx, req.NamespacedName, inheritor)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Inheritor not found, creating")
			// The Inheritor object is not found, create it and initialize the Status field
			inheritor = &labelsv1.Inheritor{
				Status: labelsv1.InheritorStatus{
					Namespaces: make(map[string]labelsv1.NamespaceStatus),
				},
			}
			if err := r.Client.Create(ctx, inheritor); err != nil {
				log.Error(err, "unable to create Inheritor")
				return ctrl.Result{RequeueAfter: time.Minute * 2}, err
			}
		} else {
			log.Error(err, "unable to fetch Inheritor")
			return ctrl.Result{RequeueAfter: time.Minute * 2}, client.IgnoreNotFound(err)
		}
	}

	if inheritor.Status.Namespaces == nil {
		// Initialize the Namespaces field
		log.Info("Initializing Namespaces field")
		inheritor.Status.Namespaces = make(map[string]labelsv1.NamespaceStatus)
	}

	for _, selector := range inheritor.Spec.Selectors {
		log.Info("Processing selector", "selector", selector)

		nsSelecter := labels.SelectorFromSet(selector.NamespaceSelector.MatchLabels)

		// List all namespaces
		// Filter namespaces by selector
		nsList := &corev1.NamespaceList{}
		err := r.Client.List(ctx, nsList, &client.ListOptions{LabelSelector: nsSelecter})
		if err != nil {
			log.Error(err, "unable to list namespaces with selector", "selector", selector)
			return ctrl.Result{RequeueAfter: time.Minute * 2}, err
		}

		for _, ns := range nsList.Items {

			nsItem := ns

			namespace := ns.Name
			log.Info("Processing namespace", "namespace", namespace)

			// call the function to sync labels
			log.Info("Syncing Configmap Labels")
			err := r.syncConfigMapLabels(ctx, nsItem, selector.IncludeLabels)

			if err != nil {
				inheritor.Status.Namespaces[namespace] = labelsv1.NamespaceStatus{Name: namespace, LabelsSynced: false}
				log.Info("Error syncing Configmap Labels", "namespace", namespace, "error", err)
				return ctrl.Result{RequeueAfter: time.Minute * 2}, err
			}

			log.Info("Syncing Pod Labels")
			err = r.syncPodLabels(ctx, nsItem, selector.IncludeLabels)
			if err != nil {
				inheritor.Status.Namespaces[namespace] = labelsv1.NamespaceStatus{Name: namespace, LabelsSynced: false}
				log.Info("Error syncing Pod Labels", "namespace", namespace, "error", err)
				return ctrl.Result{RequeueAfter: time.Minute * 2}, err
			}

			inheritor.Status.Namespaces[namespace] = labelsv1.NamespaceStatus{Name: namespace, LabelsSynced: true}

			err = r.Client.Status().Update(ctx, inheritor)
			if err != nil {
				log.Error(err, "unable to update Inheritor status")
				return ctrl.Result{RequeueAfter: time.Minute * 2}, err
			}

		}
	}
	return ctrl.Result{RequeueAfter: time.Minute * 2}, nil
}

func (r *InheritorReconciler) syncPodLabels(ctx context.Context, ns corev1.Namespace, includeLabels []string) error {

	log := log.FromContext(ctx)

	// Get pods in namespace
	podList := &corev1.PodList{}
	err := r.Client.List(ctx, podList, &client.ListOptions{Namespace: ns.Name})
	if err != nil {
		return err
	}

	if podList.Items == nil {
		return nil
	}

	// Add labels to pods
	log.Info("Pods in namespace", "namespace", ns.Name, "count", len(podList.Items))
	for _, pod := range podList.Items {
		if pod.Labels == nil {
			pod.Labels = make(map[string]string)
		}
		for _, label := range includeLabels {
			pod.Labels[label] = ns.Labels[label]
		}
		err := r.Client.Update(ctx, &pod)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *InheritorReconciler) syncConfigMapLabels(ctx context.Context, ns corev1.Namespace, includeLabels []string) error {

	log := log.FromContext(ctx)

	// get configmaps in namespace
	cmList := &corev1.ConfigMapList{}
	err := r.Client.List(ctx, cmList, &client.ListOptions{Namespace: ns.Name})
	if err != nil {
		return err
	}

	if cmList.Items == nil {
		return nil
	}

	log.Info("Configmaps in namespace", "namespace", ns.Name, "count", len(cmList.Items))
	// Add labels to configmaps
	for _, cm := range cmList.Items {
		if cm.Labels == nil {
			cm.Labels = make(map[string]string)
		}
		for _, label := range includeLabels {
			cm.Labels[label] = ns.Labels[label]
		}
		err := r.Client.Update(ctx, &cm)
		if err != nil {
			return err
		}
	}

	//fmt.Println("Syncing labels for namespace", ns.Name)
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *InheritorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&labelsv1.Inheritor{
			Status: labelsv1.InheritorStatus{
				Namespaces: make(map[string]labelsv1.NamespaceStatus),
			},
		}).
		Complete(r)
}
