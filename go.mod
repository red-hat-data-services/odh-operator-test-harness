module github.com/red-hat-data-services/odh-operator-test-harness

go 1.14

require (
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/openshift/api v0.0.0-20210202165416-a9e731090f5e // indirect
	github.com/openshift/client-go v0.0.0-20210112165513-ebc401615f47
	k8s.io/api v0.20.0
	k8s.io/apiextensions-apiserver v0.17.0
	k8s.io/apimachinery v0.20.0
	k8s.io/client-go v0.20.0
)

// Fixed version for OpenShift 4.7 / Kubernetes 0.20.0
replace (
	github.com/openshift/api v3.9.0+incompatible => github.com/openshift/api v0.0.0-20210202165416-a9e731090f5e
	github.com/openshift/client-go v3.9.0+incompatible => github.com/openshift/client-go v0.0.0-20210112165513-ebc401615f47
	k8s.io/apiextensions-apiserver v0.17.0 => k8s.io/apiextensions-apiserver v0.20.0
)
