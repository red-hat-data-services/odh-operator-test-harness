package tests

import (
	"flag"
	"path/filepath"
	"fmt"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/red-hat-data-services/odh-operator-test-harness/pkg/metadata"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"github.com/red-hat-data-services/odh-operator-test-harness/pkg/resources"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var _ = ginkgo.Describe("ODH Operator Tests", func() {
	defer ginkgo.GinkgoRecover()
	// Try inClusterConfig, fallback to using ~/.kube/config
	config, err := rest.InClusterConfig()
	if err != nil {
		var kubeconfig *string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		}
		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	}
	if err != nil {
		panic(err.Error())
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

	ginkgo.It("Jupyterhub Load Test for Tensorflow and Pytorch", func() {
		resources.PrepareTest(config)
		client, err := kubernetes.NewForConfig(config)
		Expect(err).NotTo(HaveOccurred())
		var checkErr error = nil

		for {
			job, err := client.BatchV1().Jobs("redhat-ods-applications").Get("odh-manifest-test-job", v1.GetOptions{})
			if err != nil {
				//Failed
				checkErr= err
				metadata.Instance.JuypterHubLoadTest = "Failed"
				time.Sleep(1 * time.Minute)
			}
			if job.Status.Succeeded >= 1 {
					// Succeeded
					metadata.Instance.JuypterHubLoadTest = "Succeeded"
					break
			}
			if job.Status.Failed >= 2{
				checkErr= fmt.Errorf("job failed")
				metadata.Instance.JuypterHubLoadTest = "Failed"
				break
			}

		}
		Expect(checkErr).NotTo(HaveOccurred())
	})
})
