# CodeGen

Helm chart for deploying CodeGen service on Red Hat OpenShift with Red Hat OpenShift AI.

Serving runtime template in this example uses model _ise-uiuc/Magicoder-S-DS-6.7B_ for Xeon and _meta-llama/CodeLlama-7b-hf_ for Gaudi.

## Prerequisites

1. **Red Hat OpenShift Cluster** with dynamic _StorageClass_ to provision _PersistentVolumes_ e.g. **OpenShift Data Foundation**) and installed Operators: **Red Hat - Authorino (Technical Preview)**, **Red Hat OpenShift Service Mesh**, **Red Hat OpenShift Serverless** and **Red Hat Openshift AI**.
2. Image registry to push there docker images (https://docs.openshift.com/container-platform/4.16/registry/securing-exposing-registry.html).
3. Access to S3-compatible object storage bucket (e.g. **OpenShift Data Foundation**, **AWS S3**) and values of access and secret access keys and S3 endpoint (https://docs.redhat.com/en/documentation/red_hat_openshift_data_foundation/4.16/html/managing_hybrid_and_multicloud_resources/accessing-the-multicloud-object-gateway-with-your-applications_rhodf#accessing-the-multicloud-object-gateway-with-your-applications_rhodf).
4. Account on https://huggingface.co/, access to model _ise-uiuc/Magicoder-S-DS-6.7B_ (for Xeon) or _meta-llama/CodeLlama-7b-hf_ (for Gaudi) and token with Read permissions.

## Deploy model in Red Hat Openshift AI

1. Login to OpenShift CLI and run following commands to create new serving runtime and _hf-token_ secret.

```
cd GenAIInfra/helm-charts/codegen-openshift-rhoai/
export HFTOKEN="insert-your-huggingface-token-here"

On Xeon:
helm install servingruntime tgi --set global.huggingfacehubApiToken=${HFTOKEN}

On Gaudi:
helm install servingruntime tgi --set global.huggingfacehubApiToken=${HFTOKEN} --values tgi/gaudi-values.yaml
```

Verify if template has been created with `oc get template -n redhat-ods-applications` command.

2. Find the route for **Red Hat OpenShift AI** dashboard with below command and open it in the browser:

```
oc get routes -A | grep rhods-dashboard
```

3. Go to **Data Science Project** and click **Create data science project**. Fill the **Name** and click **Create**.
4. Go to **Workbenches** tab and click **Create workbench**. Fill the **Name**, under **Notebook image** choose _Standard Data Science_, under **Cluster storage** choose _Create new persistent storage_ and change **Persistent storage size** to 40 GB. Click **Create workbench**.
5. Open newly created Jupiter notebook and run following commands to download the model and upload it on s3:

```
%env S3_ENDPOINT=<S3_RGW_ROUTE>
%env S3_ACCESS_KEY=<AWS_ACCESS_KEY_ID>
%env S3_SECRET_KEY=<AWS_SECRET_ACCESS_KEY>
%env HF_TOKEN=<PASTE_HUGGINGFACE_TOKEN>
```

```
!pip install huggingface-hub
```

```
import os
import boto3
import botocore
import glob
from huggingface_hub import snapshot_download
bucket_name = 'first.bucket'
s3_endpoint = os.environ.get('S3_ENDPOINT')
s3_accesskey = os.environ.get('S3_ACCESS_KEY')
s3_secretkey = os.environ.get('S3_SECRET_KEY')
path = 'models'
hf_token = os.environ.get('HF_TOKEN')
session = boto3.session.Session()
s3_resource = session.resource('s3',
                               endpoint_url=s3_endpoint,
                               verify=False,
                               aws_access_key_id=s3_accesskey,
                               aws_secret_access_key=s3_secretkey)
bucket = s3_resource.Bucket(bucket_name)
```

For Xeon download _ise-uiuc/Magicoder-S-DS-6.7B_:

```
snapshot_download("ise-uiuc/Magicoder-S-DS-6.7B", cache_dir=f'./models', token=hf_token)
```

For Gaudi download _meta-llama/CodeLlama-7b-hf_:

```
snapshot_download("meta-llama/CodeLlama-7b-hf", cache_dir=f'./models', token=hf_token)
```

Upload the downloaded model to S3:

```
files = (file for file in glob.glob(f'{path}/**/*', recursive=True) if os.path.isfile(file) and "snapshots" in file)
for filename in files:
    s3_name = filename.replace(path, '')
    print(f'Uploading: {filename} to {path}{s3_name}')
    bucket.upload_file(filename, f'{path}{s3_name}')
```

6. Go to your project in **Red Hat OpenShift AI** dashboard, then "Models" tab and click **Deploy model** under _Single-model serving platform_. Fill the **Name**, choose newly created **Serving runtime**: _Text Generation Inference Magicoder-S-DS-6.7B on CPU_ (for Xeon) or _Text Generation Inference CodeLlama-7b-hf on Gaudi_ (for Gaudi), **Model framework**: _llm_ and change **Model server size** to _Custom_: 16 CPUs and 64 Gi memory. For deployment with Gaudi select proper **Accelerator**. Click the checkbox to create external route in **Model route** section and uncheck the **Token authentication**. Under **Model location** choose _New data connection_ and fill all required fields for s3 access, **Bucket** _first.bucket_ and **Path**: _models_. Click **Deploy**. It takes about 10 minutes to get _Loaded_ status.\
   If it's not going to _Loaded_ status and revision changed status to "ProgressDeadlineExceeded" (`oc get revision`), scale model deployment to 0 and than to 1 with command `oc scale deployment.apps/<model_deployment_name> --replicas=1` and wait about 10 minutes for deployment.

## Install the Chart

To install the chart, back to OpenShift CLI, go to your project and run the following:

```console
cd GenAIInfra/helm-charts/

export NAMESPACE="insert-your-namespace-here"
export CLUSTERDOMAIN="$(oc get Ingress.config.openshift.io/cluster -o jsonpath='{.spec.domain}' | sed 's/^apps.//')"
export MODELNAME="insert-name-of-deployed-model-here" (it refers to the *Name* from step 6 in **Deploy model in Red Hat Openshift AI**)
export PROJECT="insert-project-name-where-model-is-deployed"

sed -i "s/insert-your-namespace-here/${NAMESPACE}/g" codegen-openshift-rhoai/llm-uservice/values.yaml

./update_dependency.sh
helm dependency update codegen-openshift-rhoai

helm install codegen codegen-openshift-rhoai --set image.repository=image-registry.openshift-image-registry.svc:5000/${NAMESPACE}/codegen --set llm-uservice.image.repository=image-registry.openshift-image-registry.svc:5000/${NAMESPACE}/llm-tgi --set react-ui.image.repository=image-registry.openshift-image-registry.svc:5000/${NAMESPACE}/react-ui --set global.clusterDomain=${CLUSTERDOMAIN} --set global.huggingfacehubApiToken=${HFTOKEN} --set llm-uservice.servingRuntime.name=${MODELNAME} --set llm-uservice.servingRuntime.namespace=${PROJECT}
```

## Verify

To verify the installation, run the command `oc get pods` to make sure all pods are running. Wait about 5 minutes for building images. When 4 pods achieve _Completed_ status, the rest with services should go to _Running_.

## Launch the UI

To access the frontend, find the route for _react-ui_ with command `oc get routes` and open it in the browser.
