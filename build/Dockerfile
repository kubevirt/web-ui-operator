FROM openshift/origin-cli:v3.11

# TODO: find better base image/creation to minimalize image size
# required binnaries:
#   - oc
#   - ansible-playbook

RUN yum install -y ansible && yum clean all

USER root

ADD build/kubevirt-web-ui-ansible /kubevirt-web-ui-ansible
ADD build/_output/bin/kubevirt-web-ui-operator /usr/local/bin/kubevirt-web-ui-operator
