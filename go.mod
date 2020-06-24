module github.com/openshift/aws-account-shredder

go 1.13

require (
	github.com/aws/aws-sdk-go v1.31.13
	github.com/openshift/aws-account-operator v0.0.0-20200610163429-768659a7cd0c
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v0.18.2
	sigs.k8s.io/controller-runtime v0.6.0
	sigs.k8s.io/structured-merge-diff v0.0.0-20190525122527-15d366b2352e // indirect
)
