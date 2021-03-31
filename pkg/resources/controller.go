package resources

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	corev1 "k8s.io/api/core/v1"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

const (
	// OdhNamespace      = "opendatahub"
	OdhNamespace      = "redhat-ods-applications"
	odhServiceAccount = "odh-manifests-test-sa"
	odhClusterRoleBinding    = "odh-manifests-test-rb"
)

var (
	odhLabels = map[string]string{
		"app":  "odh-manifests-test",
		"test": "osd-e2e-test",
	}
	decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
)

//PrepareTest is for creating SA,RoleBinding,Job. Those objects are needed to execute odh manifests test image
func PrepareTest(config *rest.Config) {

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// create SA for odh manifests test job
	_, err = clientset.CoreV1().ServiceAccounts(OdhNamespace).Get(context.Background(), odhServiceAccount, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = clientset.CoreV1().ServiceAccounts(OdhNamespace).Create(context.Background(), newSA(), metav1.CreateOptions{})
		if err != nil {
			panic(err.Error())
		}

	}

	_, err = clientset.RbacV1().ClusterRoleBindings().Get(context.Background(), odhClusterRoleBinding, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = clientset.RbacV1().ClusterRoleBindings().Create(context.Background(), newClusterRoleBinding(), metav1.CreateOptions{})
		if err != nil {
			panic(err.Error())
		}
	}

	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", odhLabels["app"]),
		Limit:         100,
	}
	joblist, err := clientset.BatchV1().Jobs(OdhNamespace).List(context.Background(), listOptions)
	if err != nil {
		panic(err.Error())
	}
	if len(joblist.Items) > 0 {
		err = clientset.BatchV1().Jobs(OdhNamespace).DeleteCollection(context.Background(), metav1.DeleteOptions{}, listOptions)
		if err != nil {
			panic(err.Error())
		}
	}

	err = createJob(context.Background(), config)

	if err != nil {
		fmt.Printf("Failed to create Job: %v\n", err)
		panic(err.Error())
	}

}

func newSA() *corev1.ServiceAccount {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      odhServiceAccount,
			Namespace: OdhNamespace,
			Labels:    odhLabels,
		},
	}
	return sa
}

func newClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      odhClusterRoleBinding,			
			Labels:    odhLabels,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: odhServiceAccount,
				Namespace: OdhNamespace,
			},
		},
	}
}

func createJob(ctx context.Context, cfg *rest.Config) error {
	job_yaml := "/home/odh-manifests-test-job.yaml"
	if os.Getenv("JOB_PATH") != "" {
		job_yaml = os.Getenv("JOB_PATH")
	}

	jobYaml, err := ioutil.ReadFile(job_yaml)

	if err != nil {
		return fmt.Errorf("Error reading job template file: %s, %v", jobYaml, err)
	}

	// 1. Prepare a RESTMapper to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return fmt.Errorf("Error creating discoveryclient: %v", err)
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	// 2. Prepare the dynamic client
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("Error creating dynamic client: %v", err)
	}

	// 3. Decode YAML manifest into unstructured.Unstructured
	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode([]byte(jobYaml), nil, obj)
	if err != nil {
		return fmt.Errorf("Error getting unstructed data: %v", err)
	}

	// 4. Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return fmt.Errorf("Error finding GVR: %v", err)
	}

	// 5. Obtain REST interface for the GVR
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		dr = dyn.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		// for cluster-wide resources
		dr = dyn.Resource(mapping.Resource)
	}

	_, err = dr.Create(context.Background(), obj, metav1.CreateOptions{})

	if err != nil {
		return fmt.Errorf("Error creating a resource(Job): %v", err)
	}

	return nil

}

func WriteLogFromPod(jobName string, clientset *kubernetes.Clientset) error {
	count := int64(200)
	podList, err := clientset.CoreV1().Pods(OdhNamespace).List(context.Background(), metav1.ListOptions{LabelSelector: "job-name=" + jobName})
	if err != nil {
		fmt.Printf("job pod does not exist: %v", err)
		return err
	}

	sortedPodList := podList.Items
	sort.Slice(sortedPodList, func(i, j int) bool {
		return sortedPodList[i].Status.StartTime.Before(sortedPodList[j].Status.StartTime)
	})

	var targetPod corev1.Pod

	for _, p := range sortedPodList {
		if p.Status.Phase != "Running" {
			targetPod = p
			break
		}
	}

	podLogOptions := corev1.PodLogOptions{
		TailLines: &count,
	}
	podLogRequest := clientset.CoreV1().
		Pods(OdhNamespace).
		GetLogs(targetPod.Name, &podLogOptions)

	stream, err := podLogRequest.Stream(context.Background())
	if err != nil {
		fmt.Printf("Can not read the failed pod log: %v", err)
		return err
	}
	// defer stream.Close()

	logFile, err := os.Create(filepath.Join("/test-run-results", "error.log"))
	if err != nil {
		return err
	}

	// defer logFile.Close()
	bufferedLogFile := bufio.NewWriter(logFile)

	for {
		buf := make([]byte, 2000)
		numBytes, err := stream.Read(buf)
		if numBytes == 0 {
			break
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		message := string(buf[:numBytes])
		fmt.Print(message)
		bufferedLogFile.WriteString(message)
		bufferedLogFile.Flush()
	}
	logFile.Close()
	stream.Close()
	return nil
}
