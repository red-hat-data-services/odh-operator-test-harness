FROM registry.access.redhat.com/ubi8/go-toolset AS builder

USER root

ENV PKG=/go/src/github.com/crobby/odh-operator-test-harness/
WORKDIR ${PKG}
RUN chmod -R 755 ${PKG}

# compile test binary
COPY . .
RUN make

FROM registry.access.redhat.com/ubi7/ubi-minimal:latest

RUN mkdir -p ${HOME} &&\
    chown 1001:0 ${HOME} &&\
    chmod ug+rwx ${HOME}

COPY --from=builder /go/src/github.com/crobby/odh-operator-test-harness/odh-operator-test-harness.test odh-operator-test-harness.test
RUN chmod +x odh-operator-test-harness.test

ENTRYPOINT [ "/odh-operator-test-harness.test" ]

USER 1001