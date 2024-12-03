package controller

import (
	"context"
	_ "fmt"
	"k8s.io/apimachinery/pkg/runtime"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	opsv1beta1 "github.com/fanny7d/ops-operator/api/v1beta1"
)

// AppsReconciler reconciles a Apps object
type AppsReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile handles the reconciliation logic
func (r *AppsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the Apps resource
	app := &opsv1beta1.Apps{}
	if err := r.Get(ctx, req.NamespacedName, app); client.IgnoreNotFound(err) != nil {
		logger.Error(err, "Failed to fetch Apps resource")
		return ctrl.Result{}, err
	} else if err != nil {
		logger.Info("Apps resource not found. Ignoring since it must be deleted.")
		return ctrl.Result{}, nil
	}

	// Ensure Deployment
	deployment := &appsv1.Deployment{}
	if err := r.Get(ctx, client.ObjectKey{Name: app.Name, Namespace: app.Namespace}, deployment); client.IgnoreNotFound(err) != nil {
		logger.Error(err, "Failed to fetch Deployment")
		return ctrl.Result{}, err
	} else if err != nil {
		// Create Deployment
		deployment = createDeployment(app)
		if err := controllerutil.SetControllerReference(app, deployment, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, deployment); err != nil {
			logger.Error(err, "Failed to create Deployment")
			return ctrl.Result{}, err
		}
		logger.Info("Deployment created successfully", "name", deployment.Name)
	}

	// Ensure Service
	service := &corev1.Service{}
	if err := r.Get(ctx, client.ObjectKey{Name: app.Name, Namespace: app.Namespace}, service); client.IgnoreNotFound(err) != nil {
		logger.Error(err, "Failed to fetch Service")
		return ctrl.Result{}, err
	} else if err != nil {
		// Create Service
		service = createService(app)
		if err := controllerutil.SetControllerReference(app, service, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, service); err != nil {
			logger.Error(err, "Failed to create Service")
			return ctrl.Result{}, err
		}
		logger.Info("Service created successfully", "name", service.Name)
	}

	// Handle Ingress
	ingress := &networkingv1.Ingress{}
	err := r.Get(ctx, client.ObjectKey{Name: app.Name, Namespace: app.Namespace}, ingress)

	if app.Spec.IngressHost == "" {
		// If IngressHost is empty, delete existing Ingress
		if err == nil {
			if err := r.Delete(ctx, ingress); err != nil {
				logger.Error(err, "Failed to delete Ingress")
				return ctrl.Result{}, err
			}
			logger.Info("Ingress deleted successfully", "name", ingress.Name)
		}
	} else {
		// If IngressHost is set, create or update Ingress
		if err != nil {
			// Create Ingress if not found
			ingress = createIngress(app)
			if err := controllerutil.SetControllerReference(app, ingress, r.Scheme); err != nil {
				return ctrl.Result{}, err
			}
			if err := r.Create(ctx, ingress); err != nil {
				logger.Error(err, "Failed to create Ingress")
				return ctrl.Result{}, err
			}
			logger.Info("Ingress created successfully", "name", ingress.Name, "host", app.Spec.IngressHost)
		}
	}

	// Update the Status
	app.Status.AvailableReplicas = deployment.Status.ReadyReplicas
	if err := r.Status().Update(ctx, app); err != nil {
		logger.Error(err, "Failed to update Apps status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager registers the controller with the manager
func (r *AppsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1beta1.Apps{}).
		Complete(r)
}

// Helper function to create Deployment
func createDeployment(app *opsv1beta1.Apps) *appsv1.Deployment {
	labels := map[string]string{"app": app.Name}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &app.Spec.Replicas,
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
							Name:  app.Name,
							Image: app.Spec.Image,
							Ports: []corev1.ContainerPort{
								{ContainerPort: app.Spec.Port},
							},
						},
					},
				},
			},
		},
	}
}

// Helper function to create Service
func createService(app *opsv1beta1.Apps) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": app.Name},
			Ports: []corev1.ServicePort{
				{Port: app.Spec.Port},
			},
		},
	}
}

// Helper function to create Ingress
func createIngress(app *opsv1beta1.Apps) *networkingv1.Ingress {
	pathType := networkingv1.PathTypePrefix
	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: app.Spec.IngressHost,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: app.Name,
											Port: networkingv1.ServiceBackendPort{
												Number: app.Spec.Port,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
