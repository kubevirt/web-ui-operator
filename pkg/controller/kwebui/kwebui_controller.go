package kwebui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	stderrors "errors"
	"crypto/rand"

    extenstionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kubevirtv1alpha1 "kubevirt.io/web-ui-operator/pkg/apis/kubevirt/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const InventoryFilePattern = "/tmp/inventory_%s.ini"
const ConfigFilePattern = "/tmp/config_%s"
const PlaybookFile = "/kubevirt-web-ui-ansible/playbooks/kubevirt-web-ui/config.yml"
const WebUIContainerName = "console"

const PhaseFreshProvision = "PROVISION_STARTED"
const PhaseProvisioned = "PROVISIONED"
const PhaseProvisionFailed = "PROVISION_FAILED"
const PhaseDeprovision = "DEPROVISION_STARTED"
const PhaseDeprovisioned = "DEPROVISIONED"
const PhaseDeprovisionFailed = "DEPROVISION_FAILED"
const PhaseOtherError = "OTHER_ERROR"
const PhaseNoDeployment = "NOT_DEPLOYED"

var log = logf.Log.WithName("controller_kwebui")

// Add creates a new KWebUI Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKWebUI{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("kwebui-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource KWebUI
	err = c.Watch(&source.Kind{Type: &kubevirtv1alpha1.KWebUI{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

/*
	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner KWebUI
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &kubevirtv1alpha1.KWebUI{},
	})
	if err != nil {
		return err
	}
*/
	return nil
}

var _ reconcile.Reconciler = &ReconcileKWebUI{}

// ReconcileKWebUI reconciles a KWebUI object
type ReconcileKWebUI struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a KWebUI object and makes changes based on the state read
// and what is in the KWebUI.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKWebUI) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// TODO: in case of error wait before reconciling again, see
	// following does not work: return reconcile.Result{RequeueAfter: RequeueDelay}, err
	// for reason, see: vendor/sigs.k8s.io/controller-runtime/pkg/internal/controller/controller.go

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling KWebUI")

	// Fetch the KWebUI instance
	instance := &kubevirtv1alpha1.KWebUI{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			// TODO: use finalizer if the KWebUI CR is deleted
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	reqLogger.Info("Desired kubevirt-web-ui version: ", "instance.Spec.Version", instance.Spec.Version)

	// Fetch the kubevirt-web-ui Deployment
	deployment := &extenstionsv1beta1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "console", Namespace: request.Namespace}, deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			return freshProvision(r, request.Namespace, instance)
		}
		reqLogger.Info("kubevirt-web-ui Deployment failed to be retrieved. Re-trying in a moment.", "error", err)
		updateStatus(r, instance, PhaseOtherError, "Failed to retrieve kubevirt-web-ui Deployment object.")
		return reconcile.Result{}, err
	}

	// Deployment found
	return reconcileExistingDeployment(r, request.Namespace, instance, deployment)
}

func runPlaybookWithSetup(namespace string, instance *kubevirtv1alpha1.KWebUI, action string) (reconcile.Result, error) {
	configFile, err := loginClient(namespace)
	if err != nil {
		return reconcile.Result{}, err
	}
	defer removeFile(configFile)

	inventoryFile, err := generateInventory(instance, namespace, action)
	if err != nil {
		return reconcile.Result{}, err
	}
	defer removeFile(inventoryFile)

	err = runPlaybook(inventoryFile, configFile)
	return reconcile.Result{}, err
}

func freshProvision(r *ReconcileKWebUI, namespace string, instance *kubevirtv1alpha1.KWebUI) (reconcile.Result, error) {
	if instance.Spec.Version == "" {
		log.Info("Removal of kubevirt-web-ui deploymnet is requested but no kubevirt-web-ui deployment found. ")
		updateStatus(r, instance, PhaseNoDeployment, "")
		return reconcile.Result{}, nil
	}

	// Kubevirt-web-ui deployment is not present yet
	log.Info("kubevirt-web-ui Deployment is not present. Ansible playbook will be executed to provision it.")
	updateStatus(r, instance, PhaseFreshProvision, fmt.Sprintf("Target version: %s", instance.Spec.Version))
	res, err := runPlaybookWithSetup(namespace, instance, "provision")
	if err == nil {
		// TODO: consider setting owner reference
		updateStatus(r, instance, PhaseProvisioned, "Provision finished.")
	} else {
		updateStatus(r, instance, PhaseProvisionFailed, "Failed to provision Kubevirt Web UI. See operator's log for more details.")
	}
	return res, err
}

func deprovision(r *ReconcileKWebUI, namespace string, instance *kubevirtv1alpha1.KWebUI) (reconcile.Result, error) {
	log.Info("Existing kubevirt-web-ui deployment is about to be deprovisioned.")
	updateStatus(r, instance, PhaseDeprovision, "")
	res, err := runPlaybookWithSetup(namespace, instance, "deprovision")
	if err == nil {
		updateStatus(r, instance, PhaseDeprovisioned, "Deprovision finished.")
	} else {
		updateStatus(r, instance, PhaseDeprovisionFailed, "Failed to deprovision Kubevirt Web UI. See operator's log for more details.")
	}

	return res, err
}

func reconcileExistingDeployment(r *ReconcileKWebUI, namespace string, instance *kubevirtv1alpha1.KWebUI, deployment *extenstionsv1beta1.Deployment) (reconcile.Result, error) {
	existingVersion := ""
	for _, container := range deployment.Spec.Template.Spec.Containers {
		if container.Name == WebUIContainerName {
			// quay.io/kubevirt/kubevirt-web-ui:v1.4
			existingVersion = afterLast(container.Image, ":")
			log.Info(fmt.Sprintf("Existing image tag: %s, from image: %s", existingVersion, container.Image))
			if existingVersion == "" {
				log.Info("Failed to read existing image tag")
				return reconcile.Result{}, stderrors.New("failed to read existing image tag")
			}
			break
		}
	}
	if existingVersion == "" {
		log.Info("Can not read deployed container version, giving up.")
		updateStatus(r, instance, PhaseOtherError, "Can not read deployed container version.")
		return reconcile.Result{}, nil
	}

	if instance.Spec.Version == existingVersion {
		msg := fmt.Sprintf("Existing version conform the requested one: %s. Nothing to do.", existingVersion)
		log.Info(msg)
		updateStatus(r, instance, PhaseProvisioned, msg)
		return reconcile.Result{}, nil
	}

	if instance.Spec.Version == "" { // deprovision only
		return deprovision(r, namespace, instance)
	}

	// requested and deployed version are different
	// It should be enough to just re-execute the provision process and restart kubevirt-web-ui pod to read the updated ConfigMap.
	// But deprovision is safer to address potential incompatible changes in the future.
	_ , err := deprovision(r, namespace, instance)
	if err != nil {
		log.Error(err, "Failed to deprovision existing deployment. Can not continue with provision of the requested one.")
		return reconcile.Result{}, err
	}

	return freshProvision(r, namespace, instance)
}

func loginClient(namespace string) (string, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed to get in-cluster config"))
		return "", err
	}

	configFile := fmt.Sprintf(ConfigFilePattern, unique())
	env := []string{fmt.Sprintf("KUBECONFIG=%s", configFile)}

	cmd, args := "oc", []string{
		"login",
		config.Host,
		fmt.Sprintf("--certificate-authority=%s", config.TLSClientConfig.CAFile),
		fmt.Sprintf("--token=%s", config.BearerToken),
	}

	anonymArgs := append([]string{}, args...)
	err = runCommand(cmd, args, env, anonymArgs)
	if err != nil {
		return "", err
	}

	cmd, args = "oc", []string{
		"project",
		namespace,
	}
	err = runCommand(cmd, args, env, args)
	if err != nil {
		return "", err
	}

	return configFile, nil
}

func generateInventory(instance *kubevirtv1alpha1.KWebUI, namespace string, action string) (string, error) {
	log.Info("Writing inventory file")
	inventoryFile := fmt.Sprintf(InventoryFilePattern, unique())
	f, err := os.Create(inventoryFile)
	if err != nil {
		log.Error(err, "Failed to write inventory file")
		return "", err
	}
	defer f.Close()

	// TODO: provide parameters if openshift-console project is not present
	f.WriteString("[OSEv3:children]\nmasters\n\n")
	f.WriteString("[OSEv3:vars]\n")
	f.WriteString("platform=openshift\n")
	f.WriteString(strings.Join([]string{"apb_action=", action, "\n"}, ""))
	f.WriteString(strings.Join([]string{"registry_url=", def(instance.Spec.RegistryUrl, "quay.io"), "\n"}, ""))
	f.WriteString(strings.Join([]string{"registry_namespace=", def(instance.Spec.RegistryNamespace, "kubevirt"), "\n"}, ""))
	f.WriteString(strings.Join([]string{"docker_tag=", def(instance.Spec.Version, "v1.4"), "\n"}, ""))
	f.WriteString(strings.Join([]string{"kubevirt_web_ui_namespace=", def(namespace, "kubevirt-web-ui"), "\n"}, ""))
	if action == "deprovision" {
		f.WriteString("preserve_namespace=true\n")
	}
	f.WriteString("\n")
	f.WriteString("[masters]\n")
	_, err = f.WriteString("127.0.0.1 ansible_connection=local\n")

	if err != nil {
		log.Error(err, "Failed to write into the inventory file")
		return "", err
	}
	f.Sync()
	log.Info("The inventory file is written.")
	return inventoryFile, nil
}

func runPlaybook(inventoryFile, configFile string) error {
	cmd, args := "ansible-playbook", []string{
		"-i",
		inventoryFile,
		PlaybookFile,
		"-vvv",
	}
	env := []string{fmt.Sprintf("KUBECONFIG=%s", configFile)}
	return runCommand(cmd, args, env, args)
}

func runCommand(cmd string, args []string, env []string, anonymArgs []string) error {
	command := exec.Command(cmd, args...)
	command.Env = append(os.Environ(), env...)
	out, err := command.CombinedOutput() // TODO: read in continuously
	if err != nil {
		log.Error(err, fmt.Sprintf("Execution failed: %s %s", cmd, strings.Join(anonymArgs," ")))
		return err
	}
	logPerLine("output:", string(out[:]))
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

func removeFile(name string) {
	// TODO: re-enable
	log.Info(fmt.Sprintf("Skiping removal of file for debug reasons: %s", name))
	/*
	err := os.Remove(name)
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed to remove file: %s", name))
	}
	*/
}

func afterLast(value string, a string) string {
	pos := strings.LastIndex(value, a)
	if pos == -1 {
		return ""
	}
	adjustedPos := pos + len(a)
	if adjustedPos >= len(value) {
		return ""
	}
	return value[adjustedPos:]
}

func updateStatus(r *ReconcileKWebUI, instance *kubevirtv1alpha1.KWebUI, phase string, msg string) {
	instance.Status.Phase = phase
	instance.Status.Message = msg

	err := r.client.Update(context.TODO(), instance)
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed to update KWebUI status. Intended to write phase: '%s', message: %s", phase, msg))
	}
}

func unique() string {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		return "abcde"
	}
	return fmt.Sprintf("%X", b)
}