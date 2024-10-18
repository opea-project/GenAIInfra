# Authentication and Authorization with APISIX and OIDC based Identity provider (Keycloak)

Follow the steps to enable authentication and authorization of OPEA services using APISIX api gateway and Identity provider like keycloak

## Start ChatQnA service
Please refer to [GenAIExamples ChatQnA](https://github.com/opea-project/GenAIExamples/tree/main/ChatQnA/kubernetes/intel) to start `chatqna` megaservice.

## Start Keycloak and configuration
In this step, we run keycloak, setup a realm with OIDC based authentication and add a user with password.

Steps to start keycloak.

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
```

Then install keycloak.
```bash
# (Option 1) install keycloak with helm by setting, you can change the user and password to your customized setting
helm install keycloak oci://registry-1.docker.io/bitnamicharts/keycloak --version 22.1.0 --set auth.adminUser=admin --set auth.adminPassword=admin

#(Option 2) install keycloak with kubectl and configuration file
kubectl apply -f ./keycloak_install.yaml
```
Get the ip and port to access keycloak.
```bash
export HOST_IP=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}' | cut -d '/' -f3 | cut -d ':' -f1)
export KEYCLOAK_PORT=$(kubectl get svc keycloak -o jsonpath='{.spec.ports[0].nodePort}')
export KEYCLOAK_ADDR=${HOST_IP}:${KEYCLOAK_PORT}
```

**Note:** Double check if the host ip captured is the correct ip.

Access the Keycloak admin console through the `KEYCLOAK_ADDR` and use the username and password specified before to login. Then we configure the users, here we create a user named "mary" and assign "user" role to her. 

The user management is done via Keycloak and the configuration steps look like this:

1. Create a new realm named `apisix` within Keycloak.

2. Create a new client called `apisix`, set `Client authentication` to `On`,  `Save` settting and go to `Credentials` page, copy `Client Secret` for later usage.

3. From the left pane select the Realm roles and create a new role name as `user`.

4. Create a new user name as `mary`, set passwords for her (set 'Temporary' to 'Off'). Select Role mapping on the top, assign the `user` role to `mary`.

5. Turn off the all the 'Required actions' under the 'Authentication' section in Keycloak

Then set some environment variables.
```bash
export USER='mary'
export PASSWORD=<password>
export KEYCLOAK_REALM='apisix'
export KEYCLOAK_CLIENT_ID='apisix'
export KEYCLOAK_CLIENT_SECRET=<keycloak client secret>
```

**Trouble Shooting: https required** 

If you meet "https required" issue when you open the console, you can fix with the following steps:
```bash
kubectl exec -it ${keycloak_pods_id} -- /bin/bash
cd /opt/keycloak/bin/
./kcadm.sh config credentials --server ${KEYCLOAK_ADDR} --realm master --user admin ## need to type in password set before
./kcadm.sh update realms/master -s sslRequired=NONE --server ${KEYCLOAK_ADDR}
```
Then after open the console and create `apisix` realm, go to "Realm setting", set "Require SSL" to "None"

## Install Apisix

```sh
# Install apisix api gateway and ingress controller [reference: https://apisix.apache.org/docs/apisix/installation-guide/]
helm repo add apisix https://charts.apiseven.com
helm repo update
helm install auth-apisix apisix/apisix -f values_apisix_gw.yaml --create-namespace --namespace auth-apisix

# WAIT UNTIL apisix-ingress-controller POD IS READY by checking status with 'kubectl get -n auth-apisix pods'
# The pod is ready when READY status shows 1/1
```
</br>
Apisix helm chart provides configs to change the service type to other options like LoadBalancer (apisix.service.type) and externalTrafficPolicy to 'local'(apisix.service.externalTrafficPolicy). These can be added in values_apisix_gw.yaml </br></br>

### Publish authenticated APIs in APISIX gateway
Check and update the oidc and megaservice config values in `apisix-chatqna-route.yaml` with environment variables, you can refer more in `templates` folder. 

For other OIDC Keycloak related information, you can get more at its discrovery endpoint `http://${KEYCLOAK_ADDR}/realms/apisix/.well-known/openid-configuration`

Then apply the authentication for ChatQnA megaservice.
```bash
kubectl apply -f apisix-chatqna-route.yaml
```

## Usage

The published APIs in apisix gateway are accessible through auth-apisix-gateway kubernetes service. By default, it is a NodePort service and can be accessed as:

```sh
export NODE_PORT=$(kubectl get --namespace auth-apisix -o jsonpath="{.spec.ports[0].nodePort}" services auth-apisix-gateway)
export NODE_IP=$(kubectl get nodes --namespace auth-apisix -o jsonpath="{.items[0].status.addresses[0].address}")

# the authenticated endpoint published in APISIX gateway can be accessed as: http://$NODE_IP:$NODE_PORT/<published endpoint uri>
export accessUrl="http://$NODE_IP:$NODE_PORT/chatqna-oidc"
```

While accessing the published APIs, the HTTP Authorization header of the request should contain the Access token provided by Identity provider as 'Bearer &ltAccess Token>'. </br></br>
The access token, refresh token, userinfo, user roles and OIDC scopes assigned to user can be obtained by invoking OIDC auth endpoint through UI or token endpoint through curl and providing user credentials. </br></br>

Below steps can be followed to get access token from keycloak and access the APISIX published ChatQnA API through curl

```sh
#Invoke Keycloak's OIDC token endpoint to get access token, refresh token and expirt times. (Only Access token is used in the example below)
export TOKEN=$(curl -X POST http://${KEYCLOAK_ADDR}/realms/${KEYCLOAK_REALM}/protocol/openid-connect/token -H 'Content-Type: application/x-www-form-urlencoded' -d "grant_type=password&client_id=${KEYCLOAK_CLIENT_ID}&client_secret=${KEYCLOAK_CLIENT_SECRET}&username=${USER}&password=${PASSWORD}" | jq -r .access_token)

# try without token. Shall get response: "Authorization required 401 error"
curl -X POST $accessUrl -d '{"messages": "What is the revenue of Nike in 2023?", "max_new_tokens":17, "do_sample": true}' -sS -H 'Content-Type: application/json' -w " %{http_code}\n"

# try with token. Shall get the correct response from ChatQnA with http code 200
curl -X POST $accessUrl -d '{"messages": "What is the revenue of Nike in 2023?", "max_new_tokens":17, "do_sample": true}' -sS -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -w " %{http_code}\n"

```

## Uninstall

```sh
# Uninstall apisix
kubectl delete -f apisix-chatqna-route.yaml
helm uninstall auth-apisix --namespace auth-apisix
```

The crds installed by apisix won't be deleted by helm uninstall. Need to manually delete those crds </br>
All APISIX specific crds can be obtained by 'kubectl get crds | grep apisix' </br>
Each crd can be manually deleted by 'kubectl delete crd/\<crd name\>' </br>
