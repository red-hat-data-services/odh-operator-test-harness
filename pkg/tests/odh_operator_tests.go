package tests

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	routeclientset "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"

	"github.com/red-hat-data-services/odh-operator-test-harness/pkg/metadata"
	"github.com/red-hat-data-services/odh-operator-test-harness/pkg/resources"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var config *rest.Config

func init() {
	// Try inClusterConfig, fallback to using ~/.kube/config
	runtimeConfig, err := rest.InClusterConfig()
	if err != nil {
		var kubeconfig *string

		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err.Error())
		}
	}else {
		config = runtimeConfig
	}
	
}

var _ = ginkgo.BeforeSuite(func() {
	defer ginkgo.GinkgoRecover()
	fmt.Println("---------------------------------------")
	fmt.Println("Wait for Jupyterhub notebook is ready.")
	fmt.Println("...")
	fmt.Println("")

	// Get Route Host
	routeClientset, err := routeclientset.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	jupyterRoute := &routev1.Route{}
	for {
		tempJupyterRoute, err := routeClientset.Routes(resources.OdhNamespace).Get(context.Background(), "jupyterhub", v1.GetOptions{})
		if err != nil {
			fmt.Println("-------------")
			fmt.Printf("Jupyterhub route does not exist: %v\n", err)
			fmt.Println("Check it again after 5 secs")
			fmt.Println("")
			time.Sleep(10 * time.Second)
		} else {
			jupyterRoute = tempJupyterRoute
			fmt.Println("Jupyterhub route created")
			break
		}
	}

	jupyterRouteHost := jupyterRoute.Spec.Host
	// fmt.Printf("%v", jupyterRoute.Spec.Host)

	//Wait until Jupyterhub route return 200 OK
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
	}
	client := &http.Client{Transport: transCfg}

	for {
		response, err := client.Get("https://" + jupyterRouteHost)
		if err != nil {
			fmt.Printf("%v", err)
		}

		if response.StatusCode == http.StatusOK {
			fmt.Println("Juypterhub is Ready so test starts")
			break
		} else {
			fmt.Println("-------------")
			fmt.Println("Juypterhub is not Ready")
			fmt.Printf("Jupyter notebook URL response code: %v\n", response.StatusCode)
			fmt.Println("Check it again after 5 secs")
			fmt.Println("")
			time.Sleep(10 * time.Second)
		}
	}
})

var _ = ginkgo.Describe("ODH Operator Tests", func() {

	ginkgo.It("kfdefs.kfdef.apps.kubeflow.org CRD exists", func() {
		apiextensions, err := clientset.NewForConfig(config)
		Expect(err).NotTo(HaveOccurred())

		// Make sure the CRD exists
		_, err = apiextensions.ApiextensionsV1().CustomResourceDefinitions().Get(context.Background(), "kfdefs.kfdef.apps.kubeflow.org", v1.GetOptions{})

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
			job, err := clientset.BatchV1().Jobs(resources.OdhNamespace).Get(context.Background(), "odh-manifests-test-job", v1.GetOptions{})
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
