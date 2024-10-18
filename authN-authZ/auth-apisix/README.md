# Authentication and Authorization with APISIX and OIDC based Identity provider (Keycloak)

Follow the steps to enable authentication and authorization of OPEA services using APISIX api gateway and Identity provider like keycloak

## Start ChatQnA service

Please refer to [GenAIExamples ChatQnA](https://github.com/opea-project/GenAIExamples/tree/main/ChatQnA/kubernetes) to start `chatqna` megaservice.

## Starting and Configuring Keycloak

In this stage, we run keycloak, setup a realm with OIDC based authentication. For demonstration, we will add a realm called `apisix` and add a user called `mary` with password. In the authentication step, only the user from `apisix` realm can access the chatQnA pipeline.

### Step 1: Prerequisite.

Create a PersistentVolume of 9Gi for keycloak-postgress with RWO access (to persist updated keycloak configuration)

```sh
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

### Step 2: Install keycloak

```bash
# (Option 1) install keycloak with helm by setting, you can change the user and password to your customized setting
helm install keycloak oci://registry-1.docker.io/bitnamicharts/keycloak --version 22.1.0 --set auth.adminUser=admin --set auth.adminPassword=admin

#(Option 2) install keycloak with kubectl and configuration file
cd ..
kubectl apply -f ./keycloak_install.yaml
```

### Step 3: Determine keycloak service ip and port.

```bash
export HOST_IP=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}' | cut -d '/' -f3 | cut -d ':' -f1)
export KEYCLOAK_PORT=$(kubectl get svc keycloak -o jsonpath='{.spec.ports[0].nodePort}')
export KEYCLOAK_ADDR=${HOST_IP}:${KEYCLOAK_PORT}
```

**Note:** Double check if the host ip captured is the correct ip.

### Step 4. Configuration.

Access the Keycloak admin console through the `KEYCLOAK_ADDR` and use the username and password specified in step 2 to login. Then we configure the users, here we create a user named "mary" and assign "user" role to her.

The user management is done via Keycloak and the configuration steps look like this:

1. Create a new realm named `apisix` (change if needed) within Keycloak.

2. Create a new client called `apisix` (change if needed) , set `Client authentication` to `On`, `Save` setting and go to the `Credentials` page, copy `Client Secret` for later use.

3. From the left pane select the Realm roles and create a new role name as `user`.

4. Create a new user name as `mary`, set password for her (set 'Temporary' to 'Off'). Select Role mapping on the top, assign the `user` role to `mary`.

5. Turn off all 'Required actions' buttons under the 'Authentication' section in Keycloak

Then set some environment variables.

```bash
export USER='mary'
export PASSWORD=<password>
export KEYCLOAK_REALM='apisix'  # change if needed
export KEYCLOAK_CLIENT_ID='apisix' # change if needed
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

Then goto the Keycloak console and find the "Realm setting" for `apisix` realm, set "Require SSL" to "None".

## Update values

Update the following values in values.yaml, you can take values_megaservice.yaml as a reference.

1. Update all the entries in oidc config pertaining to your identity provider

2. Update all the entries in API specific configs

## Install APISIX

```sh
# Install APISIX api gateway and ingress controller [reference: https://apisix.apache.org/docs/apisix/installation-guide/]
helm repo add apisix https://charts.apiseven.com
helm repo update
helm install auth-apisix apisix/apisix -f values_apisix_gw.yaml --create-namespace --namespace auth-apisix

# WAIT UNTIL the apisix-ingress-controller POD IS READY by checking it's status with 'kubectl get -n auth-apisix pods'
# The pod is ready when READY status shows 1/1

# Publish authenticated APIs in APISIX gateway
helm install auth-apisix-crds . --namespace auth-apisix
```

</br>
APISIX helm chart provides configs to change the service type to other options like LoadBalancer (apisix.service.type) and externalTrafficPolicy to 'local'(apisix.service.externalTrafficPolicy). These can be added in values_apisix_gw.yaml </br></br>

## Usage

The published APIs in APISIX gateway are accessible through auth-apisix-gateway kubernetes service. By default, it is a NodePort service and can be accessed as:

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
#Invoke Keycloak's OIDC token endpoint to get access token, refresh token and expiry  times. (Only Access token is used in the example below)
export TOKEN=$(curl -X POST http://${KEYCLOAK_ADDR}/realms/${KEYCLOAK_REALM}/protocol/openid-connect/token -H 'Content-Type: application/x-www-form-urlencoded' -d "grant_type=password&client_id=${KEYCLOAK_CLIENT_ID}&client_secret=${KEYCLOAK_CLIENT_SECRET}&username=${USER}&password=${PASSWORD}" | jq -r .access_token)

# try without token. Shall get response: "Authorization required 401 error"
curl -X POST $accessUrl -d '{"messages": "What is the revenue of Nike in 2023?", "max_new_tokens":17, "do_sample": true}' -sS -H 'Content-Type: application/json' -w " %{http_code}\n"

# try with token. Shall get the correct response from ChatQnA with http code 200
curl -X POST $accessUrl -d '{"messages": "What is the revenue of Nike in 2023?", "max_new_tokens":17, "do_sample": true}' -sS -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -w " %{http_code}\n"

```

## Uninstall

```sh
# Uninstall APISIX
helm uninstall auth-apisix-crds --namespace auth-apisix
helm uninstall auth-apisix --namespace auth-apisix
```

The crds installed by APISIX won't be deleted by helm uninstall. Need to manually delete those crds </br>
All APISIX specific crds can be obtained by 'kubectl get crds | grep apisix' </br>
Each crd can be manually deleted by 'kubectl delete crd/\<crd name\>' </br>
