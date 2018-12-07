FROM alpine:3.8

RUN apk upgrade --update --no-cache

USER nobody

ADD build/_output/bin/kubevirt-web-ui-operator /usr/local/bin/kubevirt-web-ui-operator
