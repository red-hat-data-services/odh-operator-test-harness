#!/bin/sh -x

# Run original test harness
/odh-operator-test-harness.test

pushd ${HOME}

# Set up variables for ods-ci htpasswd provider
export HTPASSWD_IDP_NAME='htpasswd'
export HTPASSWD_USER="$(pwgen 16 1)"
export HTPASSWD_PASSWORD="$(pwgen 16 1)"
# CLUSTER_ID should be passed in via the harness, if not attempt to set it
if [ x${CLUSTER_ID} == x ]; then
  export CLUSTER_ID="$(oc get clusterversion -o jsonpath='{.items[].spec.clusterID}')"
fi

ocm login --token "$(oc get secrets ci-secrets -n osde2e-ci-secrets -o json | jq -r '.data.ocm-token-refresh')"

# Add htpasswd provider for ods-ci
./ocm-htpasswd.sh

# Set up variables for ods-ci
echo "Setting up ods-ci variables"
sed -i "s%VAR_OCP_CONSOLE_URL%https://$(oc -n openshift-console get routes console -o jsonpath='{.spec.host}')%" test-variables.yml
sed -i "s%VAR_ODH_DASHBOARD_URL%https://$(oc -n redhat-ods-applications get routes rhods-dashboard -o jsonpath='{.spec.host}')%" test-variables.yml
sed -i "s%VAR_OCP_API_URL%$(oc whoami --show-server)%" test-variables.yml
sed -i "s%VAR_RHODS_PROMETHEUS_URL%https://$(oc -n openshift-monitoring get routes prometheus-k8s -o jsonpath='{.spec.host}')%" test-variables.yml

sed -i "s%VAR_PROMETHEUS_TOKEN%$(oc serviceaccounts get-token prometheus -n redhat-ods-monitoring)%" test-variables.yml
sed -i "s%VAR_TEST_USER_AUTH_TYPE%${HTPASSWD_IDP_NAME}%" test-variables.yml
sed -i "s%VAR_TEST_USER_USERNAME%${HTPASSWD_USER}%" test-variables.yml
sed -i "s%VAR_TEST_USER_PASSWORD%${HTPASSWD_PASSWORD}%" test-variables.yml
sed -i "s%VAR_ADMIN_USER_AUTH_TYPE%${HTPASSWD_IDP_NAME}%" test-variables.yml
sed -i "s%VAR_ADMIN_USER_USERNAME%${HTPASSWD_USER}%" test-variables.yml
sed -i "s%VAR_ADMIN_USER_PASSWORD%${HTPASSWD_PASSWORD}%" test-variables.yml

# Ods-ci RBAC
oc apply -f ods_ci_rbac.yaml
# Create ods-ci variables secret
oc delete secret ods-ci-test-variables
oc create secret generic ods-ci-test-variables --from-file test-variables.yml || echo "Already there"
# Launch ods-ci pod
oc delete -f ods-ci.yaml
oc apply -f ods-ci.yaml

while ! oc get pods ods-ci | grep Running; do sleep 5; done

# Rsh into the pod, and check if the completed file exists then proceed
oc rsh ods-ci /bin/bash -c "while ! [ -f /tmp/completed ]; do echo waiting for ods-ci completion; sleep 5; done; exit"
# Copy over all the results into /test-run-results
echo 'osd-ci complete, copying'
oc cp ods-ci:/tmp/ods-ci/test-output/ /test-run-results/
echo 'test complete!'
