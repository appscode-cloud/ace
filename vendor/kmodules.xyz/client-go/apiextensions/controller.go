/*
Copyright AppsCode Inc. and Contributors

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

package apiextensions

import (
	"context"
	"sync"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type (
	SetupFn func(ctx context.Context, mgr ctrl.Manager)
	TestFn  func(*apiextensionsv1.CustomResourceDefinition) bool
)

type setupGroup struct {
	gks []schema.GroupKind
	fn  SetupFn
}

var (
	setupFns  = make(map[schema.GroupKind]setupGroup)
	testFns   = make(map[schema.GroupKind]TestFn)
	setupDone = map[schema.GroupKind]bool{}
	CRDParam  = struct{}{}
	mu        sync.Mutex
)

type Reconciler struct {
	ctx context.Context
	mgr ctrl.Manager
}

func NewReconciler(ctx context.Context, mgr ctrl.Manager) *Reconciler {
	return &Reconciler{ctx: ctx, mgr: mgr}
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	var crd apiextensionsv1.CustomResourceDefinition
	if err := r.mgr.GetClient().Get(ctx, req.NamespacedName, &crd); err != nil {
		log.Error(err, "unable to fetch CustomResourceDefinition")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	gk := schema.GroupKind{
		Group: crd.Spec.Group,
		Kind:  crd.Spec.Names.Kind,
	}
	mu.Lock()
	defer mu.Unlock()
	_, found := setupDone[gk]
	if found {
		return ctrl.Result{}, nil
	}

	setup, setupFnExists := setupFns[gk]
	if !setupFnExists {
		return ctrl.Result{}, nil
	}
	if !testFns[gk](&crd) {
		return ctrl.Result{}, nil
	}

	ctxSetup := context.WithValue(r.ctx, CRDParam, &crd)
	setup.fn(ctxSetup, r.mgr)

	for _, gk := range setup.gks {
		setupDone[gk] = true
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiextensionsv1.CustomResourceDefinition{}).
		Complete(r)
}

func RegisterSetup(gk schema.GroupKind, fn SetupFn, tn ...TestFn) {
	mu.Lock()
	defer mu.Unlock()

	setupFns[gk] = setupGroup{
		gks: []schema.GroupKind{gk},
		fn:  fn,
	}
	testFns[gk] = andTestFn(tn...)
}

func MultiRegisterSetup(gks []schema.GroupKind, fn SetupFn, tn ...TestFn) {
	mu.Lock()
	defer mu.Unlock()

	testFN := andTestFn(tn...)
	for _, gk := range gks {
		setupFns[gk] = setupGroup{
			gks: gks,
			fn:  fn,
		}
		testFns[gk] = testFN
	}
}

func andTestFn(fns ...TestFn) TestFn {
	return func(crd *apiextensionsv1.CustomResourceDefinition) bool {
		for _, fn := range fns {
			if !fn(crd) {
				return false
			}
		}
		return true
	}
}
