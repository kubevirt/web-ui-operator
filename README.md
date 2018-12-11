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
TBD

## Authors
- Marek Libra
