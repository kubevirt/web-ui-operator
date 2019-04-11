FROM openshift/origin-cli:v3.11

ENV USER_NAME=kwebui \
    USER_UID=1001 \
    BASE_DIR=/opt/kwebui
ENV HOME=${BASE_DIR}

# TODO: find better base image/creation to minimalize image size
# required binnaries:
#   - oc
#   - ansible-playbook
RUN yum install -y ansible && yum clean all

ADD build/kubevirt-web-ui-ansible ${BASE_DIR}/kubevirt-web-ui-ansible
ADD build/_output/bin/web-ui-operator /usr/local/bin/kubevirt-web-ui-operator
COPY build/entrypoint.sh /usr/local/bin/entrypoint.sh

ENV PATH=/usr/local/bin:${PATH}
RUN useradd -u ${USER_UID} -r -g 0 -M -d ${BASE_DIR} -b ${BASE_DIR} -s /sbin/nologin -c "kubevirt web-ui user" ${USER_NAME} && \
    chmod -R u+x /usr/local/bin && \
    chown -R ${USER_NAME}:0 ${BASE_DIR} /usr/local/bin && \
    chmod -R g=u ${BASE_DIR} /usr/local/bin /etc/passwd

ENTRYPOINT ["entrypoint.sh"]
CMD kubevirt-web-ui-operator
