#!/bin/bash

echo This script updates kubevirt-web-ui ansible playbook by extracting it from kubevirt-ansible github branch

function usage {
  # https://github.com/kubevirt/kubevirt-ansible/archive/v0.9.2.tar.gz
  echo Usage: $0 [kubevirt-ansible tag/branch]
  echo Example: $0 master
}

BRANCH=$1

if [ x${BRANCH} = x ] ; then
  usage
  exit 1
fi

set -ex
TMP=`mktemp -d`

pushd $TMP
rm -rf kubevirt-ansible || true
git clone https://github.com/kubevirt/kubevirt-ansible.git
(cd kubevirt-ansible && git checkout $BRANCH)

mkdir -p kubevirt-web-ui-ansible/playbooks
mkdir -p kubevirt-web-ui-ansible/roles
cp -r kubevirt-ansible/playbooks/kubevirt-web-ui kubevirt-web-ui-ansible/playbooks
cp -r kubevirt-ansible/roles/kubevirt_web_ui kubevirt-web-ui-ansible/roles
popd

git rm -rf build/kubevirt-web-ui-ansible || true
mv $TMP/kubevirt-web-ui-ansible build/
git add build/kubevirt-web-ui-ansible
git commit -m "Bump kubevirt-web-ui-ansible to v${RELEASE}"

