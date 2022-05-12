##################################################
# HTPASSWD IDENTITY PROVIDER
##################################################
# You can also use "ocm create idp" instead of using the REST API
echo "Checking if an HTPASSWD identity provider is enabled"
# NOTE: Only one htpasswd identity is allowed per cluster
HTPASSWD_IDP_STATE=$(ocm list idps --cluster ${CLUSTER_ID} | grep -q htpasswd)
if [[ "$?" -ne 0 ]]; then
ocm post /api/clusters_mgmt/v1/clusters/${CLUSTER_ID}/identity_providers << EOF
{
  "type": "HTPasswdIdentityProvider",
  "name": "${HTPASSWD_IDP_NAME}",
  "htpasswd": {
    "username": "${HTPASSWD_USER}",
    "password": "${HTPASSWD_PASSWORD}"
  }
}
EOF

  if [[ "$?" -ne 0 ]]; then
    echo "ERROR adding htpasswd IDP. EXITING"
    exit 1
  fi
fi
#-------------------------------------------------
# HTPASSWD CLUSTER ADMIN
#-------------------------------------------------
# Grant HTPASSWD user cluster-admin access
FOUND_USER=$(ocm list users --cluster ${CLUSTER_ID} | grep -q ${HTPASSWD_USER})
if [[ "$?" -ne 0 ]]; then
  ocm create user ${HTPASSWD_USER} --cluster ${CLUSTER_ID} --group=cluster-admins
fi
