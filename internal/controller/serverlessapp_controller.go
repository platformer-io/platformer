package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	platformerv1 "github.com/platformer-io/platformer/api/v1"
	"github.com/platformer-io/platformer/internal/provider"
)

const finalizer = "platformer.io/cleanup"

// ServerlessAppReconciler reconciles ServerlessApp objects.
// All cloud calls go through CloudProvider — no SDK imports here.
type ServerlessAppReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Provider provider.CloudProvider
}

// +kubebuilder:rbac:groups=platformer.io,resources=serverlessapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=platformer.io,resources=serverlessapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=platformer.io,resources=serverlessapps/finalizers,verbs=update

func (r *ServerlessAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	app := &platformerv1.ServerlessApp{}
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Handle deletion via finalizer.
	if !app.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, app)
	}

	// Register finalizer on first reconcile.
	if !controllerutil.ContainsFinalizer(app, finalizer) {
		controllerutil.AddFinalizer(app, finalizer)
		if err := r.Update(ctx, app); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Route: create (never been reconciled) vs update (spec changed).
	if app.Status.ObservedGeneration == 0 && app.Status.FunctionARN == "" {
		logger.Info("creating cloud resources", "app", req.NamespacedName)
		return r.reconcileCreate(ctx, app)
	}

	if app.Status.ObservedGeneration < app.Generation {
		logger.Info("updating cloud resources", "app", req.NamespacedName,
			"observedGeneration", app.Status.ObservedGeneration,
			"generation", app.Generation)
		return r.reconcileUpdate(ctx, app)
	}

	// Already up-to-date — nothing to do.
	logger.V(1).Info("already reconciled, skipping", "app", req.NamespacedName)
	return ctrl.Result{}, nil
}

// ── Create ────────────────────────────────────────────────────────────────────

func (r *ServerlessAppReconciler) reconcileCreate(ctx context.Context, app *platformerv1.ServerlessApp) (ctrl.Result, error) {
	if err := r.setPhase(ctx, app, "Provisioning"); err != nil {
		return ctrl.Result{}, err
	}

	fnName := resourceName(app)
	logGroup := fmt.Sprintf("/aws/lambda/%s", fnName)

	// 1. IAM execution role.
	role, err := r.Provider.CreateExecutionRole(ctx, provider.RoleSpec{
		Name:        fnName,
		DatabaseARN: "", // populated in a future session when DynamoDB support is wired
	})
	if err != nil {
		return r.failWith(ctx, app, "CreateExecutionRole", err)
	}
	app.Status.ExecutionRoleARN = role.ID

	// 2. CloudWatch log group.
	if err := r.Provider.CreateLogGroup(ctx, logGroup); err != nil {
		return r.failWith(ctx, app, "CreateLogGroup", err)
	}
	app.Status.LogGroupName = logGroup

	// 3. Lambda function.
	fn, err := r.Provider.CreateFunction(ctx, provider.FunctionSpec{
		Name:          fnName,
		Runtime:       app.Spec.Runtime,
		MemoryMB:      int(app.Spec.MemoryMB),
		TimeoutSecs:   int(app.Spec.TimeoutSecs),
		ExecutionRole: role.ID,
		Environment:   app.Spec.Environment,
		CodeBucket:    app.Spec.Code.S3Bucket,
		CodeKey:       app.Spec.Code.S3Key,
	})
	if err != nil {
		return r.failWith(ctx, app, "CreateFunction", err)
	}
	app.Status.FunctionARN = fn.ID
	app.Status.FunctionVersion = fn.Version

	// 4. API Gateway (optional).
	if app.Spec.API != nil && app.Spec.API.Enabled {
		api, err := r.Provider.CreateAPIEndpoint(ctx, provider.APISpec{
			Name:     fnName,
			Protocol: "HTTP",
			TargetID: fn.ID,
			Stage:    app.Spec.API.Stage,
		})
		if err != nil {
			return r.failWith(ctx, app, "CreateAPIEndpoint", err)
		}
		app.Status.APIEndpoint = api.Endpoint
		app.Status.APIID = api.ID
	}

	// 5. DynamoDB tables (optional).
	if app.Spec.Database != nil {
		if _, err := r.Provider.CreateDatabase(ctx, toProviderDatabaseSpec(fnName, app)); err != nil {
			return r.failWith(ctx, app, "CreateDatabase", err)
		}
	}

	return r.markReady(ctx, app)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *ServerlessAppReconciler) reconcileUpdate(ctx context.Context, app *platformerv1.ServerlessApp) (ctrl.Result, error) {
	if err := r.setPhase(ctx, app, "Provisioning"); err != nil {
		return ctrl.Result{}, err
	}

	fnName := resourceName(app)

	// 1. Update Lambda code + configuration.
	fn, err := r.Provider.UpdateFunction(ctx, provider.FunctionSpec{
		Name:          fnName,
		Runtime:       app.Spec.Runtime,
		MemoryMB:      int(app.Spec.MemoryMB),
		TimeoutSecs:   int(app.Spec.TimeoutSecs),
		ExecutionRole: app.Status.ExecutionRoleARN,
		Environment:   app.Spec.Environment,
		CodeBucket:    app.Spec.Code.S3Bucket,
		CodeKey:       app.Spec.Code.S3Key,
	})
	if err != nil {
		return r.failWith(ctx, app, "UpdateFunction", err)
	}
	app.Status.FunctionARN = fn.ID
	app.Status.FunctionVersion = fn.Version

	// 2. Update API Gateway if it exists.
	if app.Spec.API != nil && app.Spec.API.Enabled && app.Status.APIID != "" {
		api, err := r.Provider.UpdateAPIEndpoint(ctx, provider.APISpec{
			Name:     fnName,
			APIID:    app.Status.APIID,
			Protocol: "HTTP",
			TargetID: fn.ID,
			Stage:    app.Spec.API.Stage,
		})
		if err != nil {
			return r.failWith(ctx, app, "UpdateAPIEndpoint", err)
		}
		app.Status.APIEndpoint = api.Endpoint
	}

	// 3. API was newly enabled — create it now.
	if app.Spec.API != nil && app.Spec.API.Enabled && app.Status.APIID == "" {
		api, err := r.Provider.CreateAPIEndpoint(ctx, provider.APISpec{
			Name:     fnName,
			Protocol: "HTTP",
			TargetID: fn.ID,
			Stage:    app.Spec.API.Stage,
		})
		if err != nil {
			return r.failWith(ctx, app, "CreateAPIEndpoint", err)
		}
		app.Status.APIEndpoint = api.Endpoint
		app.Status.APIID = api.ID
	}

	// 4. API was disabled — remove it.
	if (app.Spec.API == nil || !app.Spec.API.Enabled) && app.Status.APIID != "" {
		if err := r.Provider.DeleteAPIEndpoint(ctx, app.Status.APIID, app.Status.FunctionARN); err != nil {
			return r.failWith(ctx, app, "DeleteAPIEndpoint", err)
		}
		app.Status.APIEndpoint = ""
		app.Status.APIID = ""
	}

	return r.markReady(ctx, app)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *ServerlessAppReconciler) reconcileDelete(ctx context.Context, app *platformerv1.ServerlessApp) (ctrl.Result, error) {
	_ = r.setPhase(ctx, app, "Deleting")

	fnName := resourceName(app)

	// Best-effort cleanup — log errors but don't block finalizer removal.
	if app.Status.APIID != "" {
		_ = r.Provider.DeleteAPIEndpoint(ctx, app.Status.APIID, app.Status.FunctionARN)
	}
	_ = r.Provider.DeleteFunction(ctx, fnName)
	if app.Status.LogGroupName != "" {
		_ = r.Provider.DeleteLogGroup(ctx, app.Status.LogGroupName)
	}
	_ = r.Provider.DeleteExecutionRole(ctx, fnName)

	if app.Spec.Database != nil {
		for _, t := range app.Spec.Database.Tables {
			_ = r.Provider.DeleteDatabase(ctx, t.Name)
		}
	}

	controllerutil.RemoveFinalizer(app, finalizer)
	return ctrl.Result{}, r.Update(ctx, app)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func resourceName(app *platformerv1.ServerlessApp) string {
	return fmt.Sprintf("platformer-%s-%s", app.Namespace, app.Name)
}

func (r *ServerlessAppReconciler) setPhase(ctx context.Context, app *platformerv1.ServerlessApp, phase string) error {
	app.Status.Phase = phase
	return r.Status().Update(ctx, app)
}

func (r *ServerlessAppReconciler) markReady(ctx context.Context, app *platformerv1.ServerlessApp) (ctrl.Result, error) {
	app.Status.Phase = "Ready"
	app.Status.ObservedGeneration = app.Generation
	return ctrl.Result{}, r.Status().Update(ctx, app)
}

func (r *ServerlessAppReconciler) failWith(ctx context.Context, app *platformerv1.ServerlessApp, op string, err error) (ctrl.Result, error) {
	app.Status.Phase = "Failed"
	_ = r.Status().Update(ctx, app)
	return ctrl.Result{}, fmt.Errorf("reconcile %s: %w", op, err)
}

func toProviderDatabaseSpec(name string, app *platformerv1.ServerlessApp) provider.DatabaseSpec {
	tables := make([]provider.TableSpec, len(app.Spec.Database.Tables))
	for i, t := range app.Spec.Database.Tables {
		tables[i] = provider.TableSpec{Name: t.Name}
	}
	return provider.DatabaseSpec{Name: name, Tables: tables}
}

// SetupWithManager registers the controller with the manager.
func (r *ServerlessAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&platformerv1.ServerlessApp{}).
		Complete(r)
}
