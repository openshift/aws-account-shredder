FROM registry.svc.ci.openshift.org/openshift/release:golang-1.13 AS builder

RUN mkdir -p /workdir
WORKDIR /workdir
COPY . .
RUN make gobuild

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
COPY --from=builder /workdir/* /usr/local/bin/
CMD main