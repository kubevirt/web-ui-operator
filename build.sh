#!/bin/bash
set -ex

VERSION=v0.1.1
RELEASE=1 # see https://quay.io/repository/kubevirt/kubevirt-web-ui-operator?tab=tags

TAG1=${VERSION}-${RELEASE}
TAG2=${VERSION}

sleep 5

operator-sdk build quay.io/kubevirt/kubevirt-web-ui-operator:$TAG1
docker push quay.io/kubevirt/kubevirt-web-ui-operator:$TAG1 

docker tag quay.io/kubevirt/kubevirt-web-ui-operator:$TAG1 quay.io/kubevirt/kubevirt-web-ui-operator:$TAG2
docker push quay.io/kubevirt/kubevirt-web-ui-operator:$TAG2

docker tag quay.io/kubevirt/kubevirt-web-ui-operator:$TAG1 quay.io/kubevirt/kubevirt-web-ui-operator:latest 
docker push quay.io/kubevirt/kubevirt-web-ui-operator:latest

# go build -o build/_output/bin/web-ui-operator cmd/manager/main.go
