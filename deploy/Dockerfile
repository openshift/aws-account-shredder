FROM registry.svc.ci.openshift.org/openshift/release:golang-1.13 AS builder

RUN mkdir -p /workdir
WORKDIR /workdir
COPY go.mod go.sum ./
RUN go mod vendor
COPY . .
RUN go build main.go

####
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
COPY --from=builder /workdir/* /usr/local/bin/
CMD main