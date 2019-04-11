FROM openshift/origin-cli:v3.11

# TODO: find better base image/creation to minimalize image size
# required binnaries:
#   - oc
#   - ansible-playbook

RUN yum install -y ansible && yum clean all

# Workaround openshift issue with ansible-playbook's requirement for a named user
# Does not work - allowed range is changing
# RUN echo "weboperator:x:1000420123:0:root:/tmp:/bin/bash" >> /etc/passwd

ADD build/kubevirt-web-ui-ansible /kubevirt-web-ui-ansible
ADD build/_output/bin/web-ui-operator /usr/local/bin/kubevirt-web-ui-operator
