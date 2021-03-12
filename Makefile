TEST_HARNESS_NAME=odh-operator-test-harness
ODH_MANIFESTS_TEST=odh-manifests-test
OSD_NAMESPACE=redhat-ods-applications
DEFAULT_IMAGE_REGISTRY=quay.io
DEFAULT_REGISTRY_NAMESPACE=opendatahub
DEFAULT_IMAGE_TAG=latest
IMAGE_REGISTRY ?=$(DEFAULT_IMAGE_REGISTRY)
REGISTRY_NAMESPACE ?=$(DEFAULT_REGISTRY_NAMESPACE)
IMAGE_TAG ?=$(DEFAULT_IMAGE_TAG)
TEST_HARNESS_FULL_IMAGE_NAME=$(IMAGE_REGISTRY)/$(REGISTRY_NAMESPACE)/$(TEST_HARNESS_NAME):$(IMAGE_TAG)

DIR := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
OUT_FILE := "$(DIR)$(TEST_HARNESS_NAME)"

build:
	CGO_ENABLED=0 go test -v -c

build-image:
	@echo "Building the $(TEST_HARNESS_NAME)"
	podman build --format docker -t $(TEST_HARNESS_FULL_IMAGE_NAME) -f $(shell pwd)/Dockerfile .

push-image:
	@echo "Pushing the $(TEST_HARNESS_NAME) image to $(IMAGE_REGISTRY)/$(REGISTRY_NAMESPACE)"
	podman push $(TEST_HARNESS_FULL_IMAGE_NAME)

image: build-image push-image

cluster-test-setup:
	./hack/setup.sh

cluster-test:
	oc delete pod $(ODH_MANIFESTS_TEST)-pod --ignore-not-found
	oc get sa $(ODH_MANIFESTS_TEST)-sa -n $(OSD_NAMESPACE) || $(MAKE) cluster-test-setup

	oc run $(ODH_MANIFESTS_TEST)-pod --image=$(TEST_HARNESS_FULL_IMAGE_NAME) --restart=Never --attach -i --tty --serviceaccount $(ODH_MANIFESTS_TEST)-sa -n $(OSD_NAMESPACE)

cluster-test-clean:
	oc delete pod $(ODH_MANIFESTS_TEST)-pod -n $(OSD_NAMESPACE) --ignore-not-found
	oc delete sa $(ODH_MANIFESTS_TEST)-sa -n $(OSD_NAMESPACE) --ignore-not-found
	oc delete rolebinding $(ODH_MANIFESTS_TEST)-rb -n $(OSD_NAMESPACE) --ignore-not-found
	oc delete job $(ODH_MANIFESTS_TEST)-job -n $(OSD_NAMESPACE) --ignore-not-found
