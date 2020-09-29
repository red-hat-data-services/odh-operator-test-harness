package tests

import (
	"github.com/crobby/odh-operator-test-harness/pkg/metadata"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var _ = ginkgo.Describe("ODH Operator Tests", func() {
	defer ginkgo.GinkgoRecover()
	config, err := rest.InClusterConfig()

	if err != nil {
		panic(err)
	}

	ginkgo.It("kfdefs.kfdef.apps.kubeflow.org CRD exists", func() {
		apiextensions, err := clientset.NewForConfig(config)
		Expect(err).NotTo(HaveOccurred())

		// Make sure the CRD exists
		_, err = apiextensions.ApiextensionsV1beta1().CustomResourceDefinitions().Get("kfdefs.kfdef.apps.kubeflow.org", v1.GetOptions{})

		if err != nil {
			metadata.Instance.FoundCRD = false
		} else {
			metadata.Instance.FoundCRD = true
		}

		Expect(err).NotTo(HaveOccurred())
	})
})
