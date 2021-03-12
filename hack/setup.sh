oc project redhat-ods-applications

echo "
apiVersion: v1
kind: ServiceAccount
metadata:
  name: odh-manifests-test-sa
  namespace: redhat-ods-applications" | oc create -f -

oc adm policy add-cluster-role-to-user cluster-admin -z odh-manifests-test-sa -n redhat-ods-applications
