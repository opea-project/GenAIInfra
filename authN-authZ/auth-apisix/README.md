# Authentication and Authorization with APISIX and OIDC based Identity provider (Keycloak)

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

1. Update all the entries in oidc config pertaining to your identity provider

2. Update all the entries in API specific configs

## Install

```sh
# Install apisix api gateway and ingress controller [reference: https://apisix.apache.org/docs/apisix/installation-guide/]
helm repo add apisix https://charts.apiseven.com
helm repo update
helm install auth-apisix apisix/apisix -f values_apisix_gw.yaml --create-namespace --namespace auth-apisix

# WAIT UNTIL apisix-ingress-controller POD IS READY by checking status with 'kubectl get -n auth-apisix pods'
# The pod is ready when READY status shows 1/1

# Publish authenticated APIs in APISIX gateway
helm install auth-apisix-crds . --namespace auth-apisix
```

## Usage

The published APIs in apisix gateway are accessible through auth-apisix-gateway kubernetes service. By default, it is a NodePort service and can be accessed as:

```sh
export NODE_PORT=$(kubectl get --namespace auth-apisix -o jsonpath="{.spec.ports[0].nodePort}" services auth-apisix-gateway)
export NODE_IP=$(kubectl get nodes --namespace auth-apisix -o jsonpath="{.items[0].status.addresses[0].address}")

# the authenticated endpoint published in APISIX gateway can be accessed as: http://$NODE_IP:$NODE_PORT/<published endpoint uri>
export accessUrl=http://$NODE_IP:$NODE_PORT/<your published endpoint uri>


```

</br>
Apisix helm chart provides configs to change the service type to other options like LoadBalancer (apisix.service.type) and externalTrafficPolicy to 'local'(apisix.service.externalTrafficPolicy). These can be added in values_apisix_gw.yaml </br></br>
While accessing the published APIs, the HTTP Authorization header of the request should contain the Access token provided by Identity provider as 'Bearer &ltAccess Token>'. </br></br>
The access token, refresh token, userinfo, user roles and OIDC scopes assigned to user can be obtained by invoking OIDC auth endpoint through UI or token endpoint through curl and providing user credentials. </br></br>

Below steps can be followed to get access token from keycloak and access the APISIX published ChatQnA API through curl

```sh
# Get access token for specified user from keycloak
export USER=<username>
export PASSWORD=<password>
export KEYCLOAK_ADDR=<keycloak url>
export KEYCLOAK_REALM=<keycloak realm>
export KEYCLOAK_CLIENT_ID=<keycloak client id>
export KEYCLOAK_CLIENT_SECRET=<keycloak client secret>

#Invoke Keycloak's OIDC token endpoint to get access token, refresh token and expirt times. (Only Access token is used in the example below)
export TOKEN=$(curl -X POST http://${KEYCLOAK_ADDR}/realms/${KEYCLOAK_REALM}/protocol/openid-connect/token -H 'Content-Type: application/x-www-form-urlencoded' -d "grant_type=password&client_id=${KEYCLOAK_CLIENT_ID}&client_secret=${KEYCLOAK_CLIENT_SECRET}&username=${USER}&password=${PASSWORD}" | jq -r .access_token)

# follow instructions above to fetch the NODE_IP and NODE_PORT
export accessUrl="http://$NODE_IP:$NODE_PORT/chatqna-oidc"

# try without token. Shall get response: "Authorization required 401 error"
curl -X POST $accessUrl -d '{"text":"What is the revenue of Nike in 2023?","parameters":{"max_new_tokens":17, "do_sample": true}}' -sS -H 'Content-Type: application/json' -w " %{http_code}\n"

# try with token. Shall get the correct response from ChatQnA with http code 200
curl -X POST $accessUrl -d '{"text":"What is the revenue of Nike in 2023?","parameters":{"max_new_tokens":17, "do_sample": true}}' -sS -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -w " %{http_code}\n"

```

## Uninstall

```sh
# Uninstall apisix
helm uninstall auth-apisix-crds --namespace auth-apisix
helm uninstall auth-apisix --namespace auth-apisix
```

The crds installed by apisix won't be deleted by helm uninstall. Need to manually delete those crds </br>
All APISIX specific crds can be obtained by 'kubectl get crds | grep apisix' </br>
Each crd can be manually deleted by 'kubectl delete crd/\<crd name\>' </br>
