package resources

import (
	"context"
	"fmt"
	"io/ioutil"

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
	odhNamespace      = "redhat-ods-applications"
	odhServiceAccount = "odh-manifests-test-sa"
	odhRoleBinding    = "odh-manifests-test-rb"
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
	_, err = clientset.CoreV1().ServiceAccounts(odhNamespace).Get(odhServiceAccount, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = clientset.CoreV1().ServiceAccounts(odhNamespace).Create(newSA())
		if err != nil {
			panic(err.Error())
		}

	}

	_, err = clientset.RbacV1().RoleBindings(odhNamespace).Get(odhRoleBinding, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = clientset.RbacV1().RoleBindings(odhNamespace).Create(newRoleBinding())
		if err != nil {
			panic(err.Error())
		}
	}

	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", odhLabels["app"]),
		Limit:         100,
	}
	joblist, err := clientset.BatchV1().Jobs(odhNamespace).List(listOptions)
	if err != nil {
		panic(err.Error())
	}
	if len(joblist.Items) > 0 {
		err = clientset.BatchV1().Jobs(odhNamespace).DeleteCollection(&metav1.DeleteOptions{}, listOptions)
		if err != nil {
			panic(err.Error())
		}
	}

	err = createJob(context.Background(), config)

	if err != nil {
		panic(err.Error())
	}

}

func newSA() *corev1.ServiceAccount {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      odhServiceAccount,
			Namespace: odhNamespace,
			Labels:    odhLabels,
		},
	}
	return sa
}

func newRoleBinding() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      odhRoleBinding,
			Namespace: odhNamespace,
			Labels:    odhLabels,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "admin",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: odhServiceAccount,
			},
		},
	}
}

func createJob(ctx context.Context, cfg *rest.Config) error {
	jobYaml, err := ioutil.ReadFile("/home/odh-manifest-test-job.yaml")

	if err != nil {
		return fmt.Errorf("Error reading job template file: %s, %v", jobYaml, err)
	}

	// 1. Prepare a RESTMapper to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	// 2. Prepare the dynamic client
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return err
	}

	// 3. Decode YAML manifest into unstructured.Unstructured
	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode([]byte(jobYaml), nil, obj)
	if err != nil {
		return err
	}

	// 4. Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
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

	_, err = dr.Create(obj, metav1.CreateOptions{})

	if err != nil {
		return err
	}

	return nil

}
