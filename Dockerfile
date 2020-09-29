FROM registry.redhat.io/ubi8/go-toolset AS builder

ENV PKG=/go/src/github.com/crobby/odh-operator-test-harness/
WORKDIR ${PKG}

# compile test binary
COPY . .
RUN make

FROM registry.access.redhat.com/ubi7/ubi-minimal:latest

COPY --from=builder /go/src/github.com/crobby/odh-operator-test-harness/odh-operator-test-harness.test odh-operator-test-harness.test

ENTRYPOINT [ "/odh-operator-test-harness.test" ]

