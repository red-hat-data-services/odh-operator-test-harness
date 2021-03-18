package tests

import (
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/red-hat-data-services/odh-operator-test-harness/pkg/metadata"
	"github.com/red-hat-data-services/odh-operator-test-harness/pkg/resources"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
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
		clientset, err := kubernetes.NewForConfig(config)
		Expect(err).NotTo(HaveOccurred())
		var checkErr error = nil
		retry := 0

		for {
			job, err := clientset.BatchV1().Jobs(resources.OdhNamespace).Get("odh-manifests-test-job", v1.GetOptions{})
			if err != nil {
				//Failed
				fmt.Printf("Job Error: %v", err)
				checkErr = err
				metadata.Instance.SucceedJuypterHubLoadTest = false
				if retry == 20 {
					fmt.Println("Failed - Timeout 20mins")
					break
				}
			}
			fmt.Printf("job.Status.Succeeded: %d\n", job.Status.Succeeded)
			fmt.Printf("job.Status.Failed: %d\n", job.Status.Failed)
			if job.Status.Succeeded >= 1 {
				// Succeeded
				metadata.Instance.SucceedJuypterHubLoadTest = true
				fmt.Println("Job is successfully finished.")
				break
			}

			if job.Status.Failed >= 2 {
				checkErr = fmt.Errorf("Job failed more than 2 times")
				metadata.Instance.SucceedJuypterHubLoadTest = false
				if err := resources.WriteLogFromPod(job.Name, clientset); err != nil {
					checkErr = fmt.Errorf("Writing log failed")
					break
				}
				break
			}
			fmt.Println("Job is not finished")
			fmt.Printf("You waited for %d Mins\n", retry)
			time.Sleep(1 * time.Minute)
			retry++
			fmt.Println("")
			fmt.Println("---------------------")
		}
		if checkErr != nil {
			fmt.Println("Job failed.")
		} else {
			fmt.Println("Job finished successfully.")
		}
		Expect(checkErr).NotTo(HaveOccurred())
	})
})
