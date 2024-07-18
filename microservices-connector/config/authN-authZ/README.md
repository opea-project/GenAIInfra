# Use GMC and Istio to compose an OPEA Pipeline with authentication and authorization enabled

Authentication and authorization are essential measures that ensure the secure operation of our GenAI workload. Currently we provide two options to implement the task : via bearer JWT token and via keycloak.
And here we use the chatQnA pipeline as an example.

## Prerequisite

Before composing an OPEA pipeline with authN & authZ using GMC, user need to install Istio to support this feature. Please follow the steps [here](https://istio.io/latest/docs/setup/install/istioctl/) for installation.

**Deploy chatQnA GMC custom resource and enable Istio sidecar injection**

```sh
kubectl create ns chatqa
kubectl label ns chatqa istio-injection=enabled
kubectl apply -f $(pwd)/config/samples/chatQnA_xeon.yaml
```

## Perform authentication and authorization via bearer JWT tokens

**Apply authentication and authorization policies to the pipeline endpoint based on raw JWT tokens**

```sh
# apply the yaml to request authentication using JWT token
kubectl apply -f $(pwd)/config/authN-authZ/chatQnA_authZ_fakejwt.yaml

# apply the yaml file to request that only JWT token with
# issuer & sub == "testing@secure.istio.io" and groups belongs to group1
# can access the endpoint of chatQnA service
kubectl apply -f $(pwd)/config/authN-authZ/chatQnA_authN_fakejwt.yaml
```

After applying these two yaml files, we have setup the policy that only user with a valid JWT token (with valid issuer and claims) could access the pipeline endpoint.

**Generate JWT tokens for testing**

Use the JWT tools provided by the Istio community to generate different JWT tokens to test the authentication and authorization cases.

Download the jwt tools and run the steps to generate tokens. Note: the JWT tool requires [python](https://www.python.org/downloads/) and [jwcrypto](https://pypi.org/project/jwcrypto/) as its dependency. Please download these if they didn't exist on your machine.

```sh
cd $(pwd)
wget --no-verbose https://raw.githubusercontent.com/istio/istio/release-1.22/security/tools/jwt/samples/gen-jwt.py
wget --no-verbose https://raw.githubusercontent.com/istio/istio/release-1.22/security/tools/jwt/samples/key.pem

# create an invalid jwt token with default issuer & sub = "testing@secure.istio.io" and claims=groups:group2
export TOKEN1=$(python3 -W ignore gen-jwt.py -claims=groups:group2 key.pem)

# create a valid jwt token with default issuer, sub and claims=groups:group1
export TOKEN2=$(python3 -W ignore gen-jwt.py -claims=groups:group1 key.pem)
```

**Validate authentication and authorization with different JWT tokens**

Deploy one client pod and test the chatQnA application with different tokens

```bash
kubectl create deployment client-test -n chatqa --image=python:3.8.13 -- sleep infinity

export CLIENT_POD=$(kubectl get pod -n chatqa -l app=client-test -o jsonpath={.items..metadata.name})
export accessUrl=$(kubectl get gmc -n chatqa -o jsonpath="{.items[?(@.metadata.name=='chatqa')].status.accessUrl}")

# try with an invalid token. Shall get response: "RBAC: access denied 403"
kubectl exec -it -n chatqa $CLIENT_POD -- curl -X POST $accessUrl -d '{"text":"What is the revenue of Nike in 2023?","parameters":{"max_new_tokens":17, "do_sample": true}}' -sS -H 'Authorization: Bearer $TOKEN1' -H 'Content-Type: application/json' -w " %{http_code}\n"

# try with the valid token. Shall get the correct response
kubectl exec -it -n chatqa $CLIENT_POD -- curl -X POST $accessUrl -d '{"text":"What is the revenue of Nike in 2023?","parameters":{"max_new_tokens":17, "do_sample": true}}' -sS -H 'Authorization: Bearer $TOKEN2' -H 'Content-Type: application/json'

```
