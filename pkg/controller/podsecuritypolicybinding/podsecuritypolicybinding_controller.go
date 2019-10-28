package podsecuritypolicybinding

import (
	"context"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"

	securityv1alpha1 "github.com/cloudfoundry/security-operator/pkg/apis/security/v1alpha1"
	"github.com/cloudfoundry/security-operator/pkg/policies"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_podsecuritypolicybinding")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new PodSecurityPolicyBinding Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePodSecurityPolicyBinding{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("podsecuritypolicybinding-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource PodSecurityPolicyBinding
	err = c.Watch(&source.Kind{Type: &securityv1alpha1.PodSecurityPolicyBinding{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner PodSecurityPolicyBinding
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &securityv1alpha1.PodSecurityPolicyBinding{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcilePodSecurityPolicyBinding implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcilePodSecurityPolicyBinding{}

// ReconcilePodSecurityPolicyBinding reconciles a PodSecurityPolicyBinding object
type ReconcilePodSecurityPolicyBinding struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

func (r *ReconcilePodSecurityPolicyBinding) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling PodSecurityPolicyBinding")

	pspBinding := &securityv1alpha1.PodSecurityPolicyBinding{}
	err := r.client.Get(context.TODO(), request.NamespacedName, pspBinding)

	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if pspBinding.Spec.Policy != policies.Unprivileged {
		return reconcile.Result{}, fmt.Errorf("policy name %q invalid - must be one of [%q]", pspBinding.Spec.Policy, policies.Unprivileged)
	}

	for _, subject := range pspBinding.Spec.Subjects {
		namespace := subject.Namespace
		err = r.ensurePodSecurityPolicyIsInstalled(reqLogger, namespace)
		if err != nil {
			return reconcile.Result{}, err
		}

		err = r.ensureClusterRoleExists(reqLogger, namespace, pspBinding.Spec.Policy)
		if err != nil {
			return reconcile.Result{}, err
		}

		err = r.ensureRoleBindingExists(reqLogger, pspBinding, subject, pspBinding.Spec.Policy)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *ReconcilePodSecurityPolicyBinding) ensureRoleBindingExists(reqLogger logr.Logger, cr *securityv1alpha1.PodSecurityPolicyBinding, subject rbacv1.Subject, policyName string) error {
	existingRoleBinding := &rbacv1.RoleBinding{}
	bindingName := fmt.Sprintf("%s_%s_%s", policyName, subject.Name, subject.Kind)
	err := r.client.Get(context.TODO(), client.ObjectKey{Name: bindingName, Namespace: subject.Namespace}, existingRoleBinding)
	if err == nil {
		reqLogger.Info("role binding already exists", "binding", bindingName)
		return nil
	}
	if errors.IsNotFound(err) {
		return r.createRoleBinding(reqLogger, cr, subject, bindingName, policyName)
	}
	return err
}

func (r *ReconcilePodSecurityPolicyBinding) createRoleBinding(reqLogger logr.Logger, cr *securityv1alpha1.PodSecurityPolicyBinding, subject rbacv1.Subject, bindingName, policyName string) error {
	binding := rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bindingName,
			Namespace: subject.Namespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     policyName,
		},
		Subjects: []rbacv1.Subject{subject},
	}
	if err := controllerutil.SetControllerReference(cr, &binding, r.scheme); err != nil {
		return err
	}
	return r.client.Create(context.TODO(), &binding)
}

func (r *ReconcilePodSecurityPolicyBinding) ensureClusterRoleExists(reqLogger logr.Logger, namespace, policyName string) error {
	existingRole := &rbacv1.ClusterRole{}
	err := r.client.Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: policyName}, existingRole)
	if err == nil {
		return nil
	}

	if errors.IsNotFound(err) {
		reqLogger.Info("Creating the missing ClusterRole", "clusterRoleName", policyName)
		return r.createClusterRole(reqLogger, namespace, policyName)
	}
	return err
}

func (r *ReconcilePodSecurityPolicyBinding) createClusterRole(reqLogger logr.Logger, namespace, policyName string) error {
	role := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      policyName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{"policy"},
				Resources:     []string{"podsecuritypolicy"},
				ResourceNames: []string{policyName},
				Verbs:         []string{"use"},
			},
		},
	}
	return r.client.Create(context.TODO(), role)
}

func (r *ReconcilePodSecurityPolicyBinding) ensurePodSecurityPolicyIsInstalled(reqLogger logr.Logger, pspNamespace string) error {
	foundPolicy := &policy.PodSecurityPolicy{}

	err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: pspNamespace, Name: policies.Unprivileged}, foundPolicy)
	if err == nil {
		reqLogger.Info("The unprivileged PodSecurityPolicy already exists!")
		return nil
	}

	if errors.IsNotFound(err) {
		reqLogger.Info("Installing the unprivileged PodSecurityPolicy")
		return r.client.Create(context.TODO(), policies.NewPodSecurityPolicy(pspNamespace))
	}
	return err
}

func (r *ReconcilePodSecurityPolicyBinding) ExampleReconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling PodSecurityPolicyBinding")

	// Fetch the PodSecurityPolicyBinding instance
	instance := &securityv1alpha1.PodSecurityPolicyBinding{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
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

	// Define a new Pod object
	pod := newPodForCR(instance)

	// Set PodSecurityPolicyBinding instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Pod already exists - don't requeue
	reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *securityv1alpha1.PodSecurityPolicyBinding) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
