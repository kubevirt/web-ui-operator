Please note, this project is currently in early development phase.

---

# Kubevirt Web UI Operator
The kubernetes [operator](https://github.com/operator-framework) for managing [Kubevirt Web UI](https://github.com/kubevirt/web-ui).
Leverages the [operator-sdk](https://github.com/operator-framework/operator-sdk/).

## kubevirt-ansible
To achieve full parity with the kubevirt RPM installation, the operator reuses the [kubevirt-ansible playbook](https://github.com/kubevirt/kubevirt-ansible/tree/master/playbooks/kubevirt-web-ui).
The ansible playbook lives under `build/kubevirt-web-ui-ansible` directory.

The playbook is extracted from the kubevirt-ansible project.
Please run following command to update it in this project:

```angular2
$ ./update.ansible.sh [RELEASE]
```

By design, the `kubevirt-web-ui-ansible` uses the `oc` client to perform particular installation steps.
To make it work, kubeconfig is recomposed by the operator based on in-cluster-config secrets.

## How to Build
TBD

## How to Run
```angular2
oc new-project kubevirt-web-ui
```

```angular2
oc apply -f deploy/service_account.yaml
oc adm policy add-scc-to-user anyuid -z kubevirt-web-ui-operator
oc apply -f deploy/role.yaml
oc apply -f deploy/role_binding.yaml
oc apply -f deploy/crds/kubevirt_v1alpha1_appservice_crd.yaml
oc apply -f deploy/operator.yaml 
```

```angular2
oc apply -f deploy/crds/kubevirt_v1alpha1_appservice_cr.yaml
```

TBD: generic URL of yaml files
TBD: change version

## Authors
- Marek Libra
