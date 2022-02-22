FROM registry.access.redhat.com/ubi8/go-toolset AS builder

USER root

ENV PKG=/go/src/github.com/red-hat-data-services/odh-operator-test-harness/
WORKDIR ${PKG}
RUN chmod -R 755 ${PKG}

# compile test binary
COPY . .
RUN make

FROM registry.access.redhat.com/ubi7/ubi:latest
ENV HOME=/tmp

RUN mkdir -p ${HOME} &&\
    chown 1001:0 ${HOME} &&\
    chmod ug+rwx ${HOME}

RUN mkdir -p /test-run-results &&\
    chown 1001:0 /test-run-results &&\
    chmod ug+rwx /test-run-results

RUN /usr/bin/yum -y install 'https://dl.fedoraproject.org/pub/epel/epel-release-latest-7.noarch.rpm'; /usr/bin/yum -y install pwgen jq; yum clean all
RUN curl -s https://mirror.openshift.com/pub/openshift-v4/x86_64/clients/ocp/latest/openshift-client-linux.tar.gz | tar -xvz -C /bin
RUN curl -L -o /usr/bin/ocm https://github.com/openshift-online/ocm-cli/releases/download/v0.1.61/ocm-linux-amd64
RUN chmod +x /usr/bin/ocm

COPY --from=builder /go/src/github.com/red-hat-data-services/odh-operator-test-harness/odh-operator-test-harness.test /odh-operator-test-harness.test
RUN chmod +x odh-operator-test-harness.test
COPY ods-ci.yaml /tmp/
COPY ods_ci_rbac.yaml /tmp/
COPY test-variables.yml /tmp/
COPY ocm-htpasswd.sh /tmp/
COPY run-tests.sh /tmp/

ENTRYPOINT [ "/tmp/run-tests.sh" ]

USER 1001
