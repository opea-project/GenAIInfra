# Authentication anid Authorization with APISIX and OIDC based Identity provider (Keycloak)
Follow the steps to enable authentication and authorization of OPEA services using APISIX api gateway and Identity provider like keycloak

## Prerequisites

1. Make sure ChatQnA service is running and accessible locally

2. Run keycloak, setup a realm with OIDC based authentication and add users with passwords for each user. 

Steps to start keycloak from official keycloak helm chart
```sh
# Prerequisite: Create a PersistentVolume of 9Gi for keycloak-postgress with RWO access (to persist updated keycloak configuration)
# Below is a reference to create PersistentVolume which will be used by keycloak-postgress
kubectl apply -f - <<EOF
apiVersion: v1
kind: PersistentVolume
metadata:
  name: data-keycloak-postgresql-0
spec:
  capacity:
    storage: 9Gi
  hostPath:
    path: "/mnt/keycloak-vol"
  accessModes:
    - ReadWriteOnce
EOF

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

## Usage
The published APIs in apisix gateway are accessible through auth-apisix-gateway kubernetes service. By default, it is a NodePort service and accessible on host through http://\<auth-apisix-gateway service name\>:\<forwarded host port\>. </br></br>
Apisix helm chart provides configs to change the service type to other options like LoadBalancer (apisix.service.type) and externalTrafficPolicy to 'local'(apisix.service.externalTrafficPolicy). These can be added in apisix-helm/values.yaml </br></br>
While accessing the published APIs, the HTTP Authorization header of the request should contain the Access token provided by Identity provider as 'Bearer \<Access Token\>' </br></br>
The access token, refresh token, userinfo and user roles can be obtained by invoking OIDC auth endpoint through UI or token endpoint through curl and providing user credentials. </br>

## Uninstall
```sh
# Uninstall apisix
helm uninstall auth-apisix-crds --namespace auth-apisix
helm uninstall auth-apisix --namespace auth-apisix
```

