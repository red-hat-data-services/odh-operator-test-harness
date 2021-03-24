TEST_HARNESS_NAME=odh-operator-test-harness
ODH_MANIFESTS_TEST=odh-manifests-test
# ODS_NAMESPACE=  //update env.sh for ODS_NAMESPACE
DEFAULT_IMAGE_REGISTRY=quay.io
DEFAULT_REGISTRY_NAMESPACE=modh
DEFAULT_IMAGE_TAG=latest
IMAGE_REGISTRY ?=$(DEFAULT_IMAGE_REGISTRY)
REGISTRY_NAMESPACE ?=$(DEFAULT_REGISTRY_NAMESPACE)
IMAGE_TAG ?=$(DEFAULT_IMAGE_TAG)
TEST_HARNESS_FULL_IMAGE_NAME=$(IMAGE_REGISTRY)/$(REGISTRY_NAMESPACE)/$(TEST_HARNESS_NAME):$(IMAGE_TAG)

DIR := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
OUT_FILE := "$(DIR)$(TEST_HARNESS_NAME)"

include $(shell pwd)/env
build:
	CGO_ENABLED=0 go test -v -c

build-image:
	@echo "Building the $(TEST_HARNESS_NAME)"
	podman build --format docker -t $(TEST_HARNESS_FULL_IMAGE_NAME) -f $(shell pwd)/Dockerfile .

push-image:
	@echo "Pushing the $(TEST_HARNESS_NAME) image to $(IMAGE_REGISTRY)/$(REGISTRY_NAMESPACE)"
	podman push $(TEST_HARNESS_FULL_IMAGE_NAME)

image: build-image push-image

test-setup:
	./hack/setup.sh

job-test:
	oc delete job $(ODH_MANIFESTS_TEST)-job -n $(ODS_NAMESPACE) --ignore-not-found
	oc get sa $(ODH_MANIFESTS_TEST)-sa -n $(ODS_NAMESPACE) || $(MAKE) test-setup
	oc create -f ./template/odh-manifests-test-job.yaml -n $(ODS_NAMESPACE) 

job-test-clean:
	oc delete sa $(ODH_MANIFESTS_TEST)-sa -n $(ODS_NAMESPACE) --ignore-not-found
	oc delete rolebinding $(ODH_MANIFESTS_TEST)-rb -n $(ODS_NAMESPACE) --ignore-not-found
	oc delete job $(ODH_MANIFESTS_TEST)-job -n $(ODS_NAMESPACE) --ignore-not-found
	oc delete pod -l job_name=$(ODH_MANIFESTS_TEST)-job -n $(ODS_NAMESPACE) --ignore-not-found
	oc delete pod jupyterhub-nb-admin  -n $(ODS_NAMESPACE) --ignore-not-found --force --grace-period=0
	oc delete pvc jupyterhub-nb-admin-pvc -n $(ODS_NAMESPACE) --ignore-not-found

cluster-test:
	oc delete pod $(TEST_HARNESS_NAME)-pod -n $(ODS_NAMESPACE) --ignore-not-found
	oc get sa $(ODH_MANIFESTS_TEST)-sa -n $(ODS_NAMESPACE) || $(MAKE) test-setup

	# oc run $(ODH_MANIFESTS_TEST)-pod --image=$(TEST_HARNESS_FULL_IMAGE_NAME) --restart=Never --attach -i --tty --serviceaccount $(ODH_MANIFESTS_TEST)-sa -n $(ODS_NAMESPACE) --env=JOB_PATH=/home/odh-manifest-test-job-pvc.yaml
	oc create -f ./hack/odh-operator-test-harness-pod.yaml -n $(ODS_NAMESPACE)
	# oc logs odh-operator-test-harness-pod -c odh -f

cluster-test-clean:
	oc delete -f ./hack/odh-operator-test-harness-pod.yaml -n $(ODS_NAMESPACE) --ignore-not-found
	oc delete sa $(ODH_MANIFESTS_TEST)-sa -n $(ODS_NAMESPACE) --ignore-not-found
	oc delete rolebinding $(ODH_MANIFESTS_TEST)-rb -n $(ODS_NAMESPACE) --ignore-not-found
	oc delete job $(ODH_MANIFESTS_TEST)-job -n $(ODS_NAMESPACE) --ignore-not-found
	oc delete pod -l job_name=$(ODH_MANIFESTS_TEST)-job -n $(ODS_NAMESPACE) --ignore-not-found
	oc delete pod jupyterhub-nb-admin  -n $(ODS_NAMESPACE) --ignore-not-found --force --grace-period=0
	oc delete pvc jupyterhub-nb-admin-pvc -n $(ODS_NAMESPACE) --ignore-not-found

