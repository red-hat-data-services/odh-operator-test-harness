#!/bin/bash
source ./env.sh
oc project ${ODS_NAMESPACE}

echo "
apiVersion: v1
kind: ServiceAccount
metadata:
  name: odh-manifests-test-sa
  namespace: ${ODS_NAMESPACE}" | oc create -f -

oc adm policy add-cluster-role-to-user cluster-admin -z odh-manifests-test-sa -n  ${ODS_NAMESPACE}
