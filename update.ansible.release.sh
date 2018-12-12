#!/bin/bash

echo This script updates kubevirt-web-ui ansible playbook by extracting it from kubevirt-ansible official release

function usage {
  # https://github.com/kubevirt/kubevirt-ansible/archive/v0.9.2.tar.gz
  echo Usage: $0 [kubevirt-ansible release]
  echo Example: $0 0.9.2
}

RELEASE=$1

if [ x${RELEASE} = x ] ; then
  usage
  exit 1
fi

set -ex
TMP=`mktemp -d`

pushd $TMP
wget -O kubevirt-ansible.tgz https://github.com/kubevirt/kubevirt-ansible/archive/v${RELEASE}.tar.gz
tar -xzf kubevirt-ansible.tgz

mkdir -p kubevirt-web-ui-ansible/playbooks
mkdir -p kubevirt-web-ui-ansible/roles
cp -r kubevirt-ansible-${RELEASE}/playbooks/kubevirt-web-ui kubevirt-web-ui-ansible/playbooks
cp -r kubevirt-ansible-${RELEASE}/roles/kubevirt_web_ui kubevirt-web-ui-ansible/roles
popd

git rm -rf build/kubevirt-web-ui-ansible || true
mv $TMP/kubevirt-web-ui-ansible build/
git add build/kubevirt-web-ui-ansible
git commit -m "Bump kubevirt-web-ui-ansible to v${RELEASE}"

