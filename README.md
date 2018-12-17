# Kubevirt Web UI Operator
The kubernetes [operator](https://github.com/operator-framework) for managing [Kubevirt Web UI](https://github.com/kubevirt/web-ui) deployment.
Leverages the [operator-sdk](https://github.com/operator-framework/operator-sdk/).

Kubevirt-web-ui image repository on quay.io: [quay.io/repository/kubevirt/kubevirt-web-ui](https://quay.io/repository/kubevirt/kubevirt-web-ui?tab=tags)

## How to Run
Depending on your OpenShift cluster installation, please choose from the two variants bellow.

If `Cluster Console` (in `openshift-console` project) is deployed (as by default), optional parameters can be automatically retrieved from its ConfigMap (follow Variant 1).
Otherwise they need to be explicitely provided (Variant 2).

### Variant 1: The openshift-console Is Installed
To ease deployment, parameters of the cluster deployment can be  automatically retrieved from the `openshift-console` ConfigMap, if present.

To do so, the operator's service account needs to be granted to access the `openshift-console` namespace.

```angular2
oc new-project kubevirt-web-ui

oc apply -f deploy/service_account.yaml
oc adm policy add-scc-to-user anyuid -z kubevirt-web-ui-operator

oc apply -f deploy/role.yaml
oc apply -f deploy/role_extra_for_console.yaml
oc apply -f deploy/role_binding.yaml
oc apply -f deploy/role_binding_extra_for_console.yaml
```

### Variant 2: The openshift-console Is Not Installed
In `deploy/crds/kubevirt_v1alpha1_kwebui_cr.yaml`, add following under `spec` section based on your actual OpenShift cluster deployment: 

- `openshift_master_default_subdomain=[SUBDOMAIN FOR APPLICATIONS]`
  - example: `router.default.svc.cluster.local`
  - Used for composition of web-ui's public URL

- `public_master_hostname=[FQDN:port]`
  - example: `master.your.domain.com:8443`
  - Public URL of your first master node, used for composition of public `console` URL for redirects

Then execute:

```angular2
oc new-project kubevirt-web-ui

oc apply -f deploy/service_account.yaml
oc adm policy add-scc-to-user anyuid -z kubevirt-web-ui-operator

oc apply -f deploy/role.yaml
oc apply -f deploy/role_binding.yaml
```


### Kubevirt Web UI Version to Install
To actually deploy the Kubevirt Web UI, choose it's version by editting `spec.version` in `deploy/crds/kubevirt_v1alpha1_kwebui_cr.yaml`.

Example:
```angular2
spec:
  version: "1.4.0-4"
``` 

The image repository can be farther tweaked by using the `spec.registry_url` and `spec.registry_namespace` parameters. 

To **undeploy** the Web UI, set `spec.version` to empty string (`""`).
By providing non-empty value here, the Web UI deployment is **upgraded**/**downgraded**.

Please note, the `version` needs to match Web UI's docker image tag in the specified repository (seed [default quay repo](https://quay.io/repository/kubevirt/kubevirt-web-ui?tab=tags)).

### Fire Web UI Deployment
Once `spec.version` in the CR is set:

```angular2
oc apply -f deploy/crds/kubevirt_v1alpha1_kwebui_crd.yaml
oc apply -f deploy/operator.yaml 
```

### Kubevirt Web UI Deployment
Actual [Kubevirt Web UI](https://github.com/kubevirt/web-ui) deployment is managed via `KWebUI` custom resource edited in previous step:
```angular2
oc apply -f deploy/crds/kubevirt_v1alpha1_kwebui_cr.yaml
```

### Status
Processing status can be observed within the `KWebUI` custom resource's `status` section:
- `status.phase` - contains one of the string constants for automatization
- `status.message` - human readable details

In case of errors, watch operator's pod logs, sort of:
```angular2
oc logs kubevirt-web-ui-operator-85ffcdd9d5-8lt9g
```

## How to Build
See [operator-sdk](https://github.com/operator-framework/operator-sdk/) for the tooling installation instructions.

The operator is built using:
```angular2
operator-sdk build quay.io/[YOUR_REPO]/kubevirt-web-ui-operator
```

## Note About Internals
To achieve full parity with the Kubevirt RPM installation, the operator reuses the [kubevirt-ansible playbook](https://github.com/kubevirt/kubevirt-ansible/tree/master/playbooks/kubevirt-web-ui).

This decision has several consequences, namely
- requirement for the `oc` and `ansible-playbook` binaries within the docker image (increases size of the image significantly)
- requirement for the `anyuid` SCC for the service account as the ansible requires actual user to run under (in particular, container's `root` is recycled)

The project is intentionally not based on the [ansible operator-sdk](https://github.com/operator-framework/operator-sdk/blob/master/doc/ansible/user-guide.md) as there is still plan to remove the ansible code completely once the (de)provision logic can live in a single project only. 

Copy of the ansible playbook is stored under `build/kubevirt-web-ui-ansible` directory.
This playbook is extracted from the kubevirt-ansible project.

Please run following command to update it in this project:

```angular2
$ ./update.ansible.sh [RELEASE]
```

By design, the `kubevirt-web-ui-ansible` uses the `oc` client to perform particular installation steps.
To make it work, kubeconfig is recomposed by the operator based on in-cluster-config secrets.

## Authors
- Marek Libra
