/*
Copyright 2022.

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

package controllers

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cc "github.com/GoogleCloudPlatform/k8s-config-connector/pkg/apis/k8s/v1alpha1"
	"github.com/mortent/config-connector-test/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	defaultCondition = cc.Condition{
		Type:    "Ready",
		Status:  corev1.ConditionFalse,
		Reason:  "Updating",
		Message: "Update in progress",
	}
)

// StubReconciler reconciles a Stub object
type StubReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=test.cnrm.cloud.google.com,resources=stubs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=test.cnrm.cloud.google.com,resources=stubs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=test.cnrm.cloud.google.com,resources=stubs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Stub object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *StubReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var stub v1beta1.Stub
	if err := r.Get(ctx, req.NamespacedName, &stub); err != nil {
		logger.Error(err, "unable to fetch stub")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	condition, nextChange := r.currentCondition(ctx, &stub)

	var needsUpdate bool
	if len(stub.Status.Conditions) == 0 {
		condition.LastTransitionTime = metav1.Now().Format(time.RFC3339)
		stub.Status.Conditions = append(stub.Status.Conditions, condition)
		logger.Info("Setting Ready condition", "condition", condition)
		needsUpdate = true
	} else {
		for i := range stub.Status.Conditions {
			c := stub.Status.Conditions[i]
			if c.Type != condition.Type {
				continue
			}
			if c.Message != condition.Message ||
				c.Status != condition.Status ||
				c.Reason != condition.Reason {
				condition.LastTransitionTime = metav1.Now().Format(time.RFC3339)
				stub.Status.Conditions[i] = condition
				logger.Info("Updating Ready condition", "condition", condition)
				needsUpdate = true
			}
		}
	}
	stub.Status.ObservedGeneration = stub.GetObjectMeta().GetGeneration()

	if needsUpdate {
		logger.Info("Updating stub")
		if err := r.Status().Update(ctx, &stub); err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}

	if nextChange.IsZero() {
		logger.Info("No further changes")
		return ctrl.Result{}, nil
	}

	logger.Info("Next reconcile scheduled", "time", nextChange.Sub(time.Now()))
	return ctrl.Result{RequeueAfter: nextChange.Sub(time.Now())}, nil
}

func (r *StubReconciler) currentCondition(_ context.Context, stub *v1beta1.Stub) (cc.Condition, time.Time) {
	sequence := stub.Spec.ConditionSequence
	currentTime := time.Now()
	startTime := stub.ObjectMeta.CreationTimestamp.Time
	// If empty sequence, just set the default condition and don't provide a new update time.
	if len(sequence) == 0 {
		return defaultCondition, time.Time{}
	}

	// If we are not at time for the first update in the sequence yet, just
	// set the default condition and provide an update time when it should be updated.
	firstChangeTime := startTime.Add(time.Second * time.Duration(sequence[0].LatencySeconds))
	if currentTime.Before(firstChangeTime) {
		return defaultCondition, firstChangeTime
	}

	var latencyCounter int64
	for i := 0; i < len(sequence)-1; i++ {
		latencyCounter = latencyCounter + sequence[i].LatencySeconds
		lower := startTime.Add(time.Second * time.Duration(latencyCounter))
		upper := startTime.Add(time.Second * time.Duration(latencyCounter+sequence[i+1].LatencySeconds))
		if lower.Before(currentTime) && upper.After(currentTime) {
			return sequence[i].Condition, upper
		}
	}
	return sequence[len(sequence)-1].Condition, time.Time{}
}

// SetupWithManager sets up the controller with the Manager.
func (r *StubReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Stub{}).
		Complete(r)
}
