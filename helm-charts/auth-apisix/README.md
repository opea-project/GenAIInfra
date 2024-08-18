# Authentication anid Authorization with APISIX and OIDC based Identity provider (Keycloak)
Follow the steps to enable authentication and authorization of OPEA services using APISIX api gateway and Identity provider like keycloak

## Prerequisites

1. Make sure ChatQnA service is running and accessible locally

2. Run keycloak, setup a realm with OIDC based authentication and add users with passwords for each user. 

Steps to start keycloak from official keycloak helm chart
```sh
# Prerequisite: Create a PersistentVolume of 9Gi for keycloak-postgress with RWO access (to persist updated keycloak configuration)

# install keycloak with helm by setting 
helm install keycloak oci://registry-1.docker.io/bitnamicharts/keycloak --version 22.1.0 --set auth.adminUser=admin --set auth.adminPassword=admin

# Access keycloak UI through service/keycloak to do the necessary configurations
```
Once the keycloak pod is up and running, access the UI through the Keycloak's NodePort service and set the necessary configuration

## Update values
Update the following values in values.yaml

1. Update all the entries in oidc config in apis-crd-helm/values.yaml

2. Update all the entries in API specific configs in apis-crd-helm/values.yaml

## Install 

```sh
cd apisix-helm

# Get dependencies (apisix helm chart)
helm dependency update

# Install apisix 
helm install auth-apisix . --create-namespace --namespace auth-apisix

# WAIT UNTIL apisix-ingress-controller POD IS READY

# Apply API configs to apisix-ingress-controller 
cd ../apis-crd-helm
helm install auth-apisix-crds . --namespace auth-apisix
```

## Uninstall
```sh
# Uninstall apisix
helm uninstall auth-apisix-crds --namespace auth-apisix
helm uninstall auth-apisix --namespace auth-apisix
```

