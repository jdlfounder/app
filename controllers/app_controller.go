/*
Copyright 2022 jdlfounder.

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
	//"sigs.k8s.io/controller-runtime/pkg/log"
	//
	"context"
	"encoding/json"
	"reflect"

	"github.com/jdlfounder/app/resources/deployment"
	"github.com/jdlfounder/app/resources/service"
	"k8s.io/apimachinery/pkg/api/errors"

	//"github.com/go-logr/logr"
	//"sigs.k8s.io/controller-runtime/pkg/log"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appv1 "github.com/jdlfounder/app/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// AppReconciler reconciles a App object
type AppReconciler struct {
	client.Client
	//Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=app.example.com,resources=apps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.example.com,resources=apps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.example.com,resources=apps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the App object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *AppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//_ = log.FromContext(ctx)
	//logger := log.FromContext(ctx)
	//_ = r.Log.WithValues("app", req.NamespacedName)
	// TODO(user): your logic here
	// 获取 crd 资源
	instance := &appv1.App{}
	if err := r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// crd 资源已经标记为删除
	if instance.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	oldDeploy := &appsv1.Deployment{}
	if err := r.Client.Get(ctx, req.NamespacedName, oldDeploy); err != nil {
		// deployment 不存在，创建
		if errors.IsNotFound(err) {
			// 创建deployment
			if err := r.Client.Create(ctx, deployment.New(instance)); err != nil {
				return ctrl.Result{}, err
			}

			// 创建service
			if err := r.Client.Create(ctx, service.New(instance)); err != nil {
				return ctrl.Result{}, err
			}

			// 更新 crd 资源的 Annotations
			data, _ := json.Marshal(instance.Spec)
			if instance.Annotations != nil {
				instance.Annotations["spec"] = string(data)
			} else {
				instance.Annotations = map[string]string{"spec": string(data)}
			}
			if err := r.Client.Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, err
		}
	} else {
		// deployment 存在，更新
		oldSpec := appv1.AppSpec{}
		if err := json.Unmarshal([]byte(instance.Annotations["spec"]), &oldSpec); err != nil {
			return ctrl.Result{}, err
		}

		if !reflect.DeepEqual(instance.Spec, oldSpec) {
			// 更新deployment
			newDeploy := deployment.New(instance)
			oldDeploy.Spec = newDeploy.Spec
			if err := r.Client.Update(ctx, oldDeploy); err != nil {
				return ctrl.Result{}, err
			}

			// 更新service
			newService := service.New(instance)
			oldService := &corev1.Service{}
			if err := r.Client.Get(ctx, req.NamespacedName, oldService); err != nil {
				return ctrl.Result{}, err
			}
			clusterIP := oldService.Spec.ClusterIP // 更新 service 必须设置老的 clusterIP
			oldService.Spec = newService.Spec
			oldService.Spec.ClusterIP = clusterIP
			if err := r.Client.Update(ctx, oldService); err != nil {
				return ctrl.Result{}, err
			}

			// 更新 crd 资源的 Annotations
			data, _ := json.Marshal(instance.Spec)
			if instance.Annotations != nil {
				instance.Annotations["spec"] = string(data)
			} else {
				instance.Annotations = map[string]string{"spec": string(data)}
			}
			if err := r.Client.Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1.App{}).
		Complete(r)
}
