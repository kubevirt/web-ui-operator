package appservice

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/apimachinery/pkg/api/errors"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kubevirtv1alpha1 "kubevirt.io/kubevirt-web-ui-operator/pkg/apis/kubevirt/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
//	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const InventoryFile = "/tmp/inventory.ini"
var log = logf.Log.WithName("controller_appservice")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new AppService Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAppService{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("appservice-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource AppService
	err = c.Watch(&source.Kind{Type: &kubevirtv1alpha1.AppService{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

/*
	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner AppService
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &kubevirtv1alpha1.AppService{},
	})
	if err != nil {
		return err
	}
*/
	return nil
}

var _ reconcile.Reconciler = &ReconcileAppService{}

// ReconcileAppService reconciles a AppService object
type ReconcileAppService struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a AppService object and makes changes based on the state read
// and what is in the AppService.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAppService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling AppService")

	// Fetch the AppService instance
	instance := &kubevirtv1alpha1.AppService{}
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
	reqLogger.Info("Desired kubevirt-web-ui version", "instance.Spec.Version", instance.Spec.Version)

	// Fetch the kubevirt-web-ui ReplicaSet
	replicaSet := &corev1.ReplicationController{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "console", Namespace: request.Namespace}, replicaSet)
	if err != nil {
		if errors.IsNotFound(err) {
			// TODO: add error handling of inner-steps
			// Kubevirt-web-ui deployment is not present yet
			reqLogger.Info("kubevirt-web-ui ReplicaSet is not present. Ansible playbook will be executed to provision it.")
			loginClient()
			generateInventory(instance, "provision")
			provisionKubevirtWebUI()
			// TODO: start ansible playbook to provision
			// TODO: log ansible Log output
			// TODO: return based on exit code
			// TODO: consider setting owner reference
			return reconcile.Result{}, nil
		}
		reqLogger.Info("kubevirt-web-ui ReplicaSet failed to be retrieved. Re-trying in a moment.", "error", err)
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// ReplicaSet found
	// TODO: check installed version and optionally deprovision-provision
	// It should be enough to just re-execute the provision process and restart kubevirt-web-ui pod to read the updated ConfigMap. But deprovision is safe to address potential incompatible changes.

	return reconcile.Result{}, nil

	/*
		// Define a new Pod object
		pod := newPodForCR(instance)

		// Set AppService instance as the owner and controller
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
	*/
}

func loginClient() {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed to get in-cluster config"))
	}

	cmd, args := "oc", []string{
		"login",
		config.Host,
		fmt.Sprintf("--certificate-authority=%s", config.TLSClientConfig.CAFile),
		fmt.Sprintf("--token=%s", config.BearerToken),
	}
	env := []string{"KUBECONFIG=/tmp/config"}

	command := exec.Command(cmd, args...)
	command.Env = append(os.Environ(), env...)
	out, err := command.CombinedOutput()
	if err != nil {
		args[3] = "--token=[SECRET]"
		log.Error(err, fmt.Sprintf("Execution failed: %s %s", cmd, strings.Join(args," ")))
	}
	logPerLine("Login output:", string(out[:]))
}

func generateInventory(instance *kubevirtv1alpha1.AppService, action string) error {
	log.Info("Writing inventory file")
	f, err := os.Create(InventoryFile)
	if err != nil {
		log.Error(err, "Failed to write inventory file")
		return err
	}
	defer f.Close()
	f.WriteString("[OSEv3:children]\nmasters\n\n")
	f.WriteString("[OSEv3:vars]\n")
	f.WriteString("platform=openshift\n")
	f.WriteString(strings.Join([]string{"apb_action=", action, "\n"}, ""))
	f.WriteString(strings.Join([]string{"registry_url=", def(instance.Spec.RegistryUrl, "quay.io"), "\n"}, ""))
	f.WriteString(strings.Join([]string{"registry_namespace=", def(instance.Spec.RegistryNamespace, "kubevirt"), "\n"}, ""))
	f.WriteString(strings.Join([]string{"docker_tag=", def(instance.Spec.Version, "v1.4"), "\n"}, ""))
	f.WriteString("\n")
	f.WriteString("[masters]\n")
	_, err = f.WriteString("127.0.0.1 ansible_connection=local\n")
	if err != nil {
		log.Error(err, "Failed to write into the inventory file")
		return err
	}
	f.Sync()
	log.Info("The inventory file is written.")
	return nil
}

func provisionKubevirtWebUI() error {
	// TODO: create inventory file, set parameters
	// run ansible-playbook

	// Just for test:
	cmd, args := "oc", []string{
		"get",
		"pods",
	}
	env := []string{"KUBECONFIG=/tmp/config"}

	command := exec.Command(cmd, args...)
	command.Env = append(os.Environ(), env...)
	out, err := command.CombinedOutput()
	if err != nil {
		log.Error(err, fmt.Sprintf("Execution failed: %s %s", cmd, strings.Join(args," ")))
		return err
	}
	logPerLine("Test output:", string(out[:]))

	return nil
}

func logPerLine(header string, out string) {
	log.Info(header)
	for _,line := range strings.Split(out, "\n") {
		log.Info(line)
	}
}

func def(s string, defVal string) string {
	if s == "" {
		return defVal
	}
	return s
}


/*
// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *kubevirtv1alpha1.AppService) *corev1.Pod {
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
*/