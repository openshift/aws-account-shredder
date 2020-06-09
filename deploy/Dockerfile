FROM registry.svc.ci.openshift.org/openshift/release:golang-1.14 AS builder
COPY . /go/src/github.com/openshift/aws-account-shredder
WORKDIR /go/src/github.com/openshift/aws-account-shredder
RUN go get ./...

LABEL io.openshift.managed.name="aws-account-shredder" \
      io.openshift.managed.description="this is a description"

RUN go build main.go
CMD go run main.go