#!/bin/bash
set -ex

git diff-index --quiet HEAD || (echo Commit your changes first ; false) # fail if uncomitted changes

GIT_REMOTE_NAME=upstream # or origin

CSV_VERSION=0.1.2
VERSION=v${CSV_VERSION}
RELEASE=1 # see https://quay.io/repository/kubevirt/kubevirt-web-ui-operator?tab=tags

UNIQUE=`date +"%Y-%m-%d_%H-%M-%S"`

TAG1=${VERSION}-${RELEASE}
TAG2=${VERSION}

sleep 5

git checkout master && git fetch --all && git reset --hard ${GIT_REMOTE_NAME}/master
git status
git checkout -b olm-${CSV_VERSION}-${UNIQUE}
./hack/make-olm.sh
git status

operator-sdk build quay.io/kubevirt/kubevirt-web-ui-operator:$TAG1
docker push quay.io/kubevirt/kubevirt-web-ui-operator:$TAG1 

docker tag quay.io/kubevirt/kubevirt-web-ui-operator:$TAG1 quay.io/kubevirt/kubevirt-web-ui-operator:$TAG2
docker push quay.io/kubevirt/kubevirt-web-ui-operator:$TAG2

docker tag quay.io/kubevirt/kubevirt-web-ui-operator:$TAG1 quay.io/kubevirt/kubevirt-web-ui-operator:latest 
docker push quay.io/kubevirt/kubevirt-web-ui-operator:latest

# go build -o build/_output/bin/web-ui-operator cmd/manager/main.go
