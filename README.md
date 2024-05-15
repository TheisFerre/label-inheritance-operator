# label-inheritance-operator
This is a Kubernetes operator that allows you to inherit labels from a namespace to the resources within it. The operator only supports the following resources for now:
- pods
- configmaps

## Description
The operator is built using the Kubebuilder framework. It watches for the Namespace resources and updates the labels of the resources within the namespace based on the labels of the namespace. It watches for the Namespace resources and updates the labels of the resources within the namespace based on the labels of the namespace.

### How it works

The Operator watches the Custom Resource Definition (CRD) `inheritors.labels.theisferre`. A sample CR is as follows:

```yaml
apiVersion: labels.theisferre/v1
kind: Inheritor
metadata:
  name: inheritor-sample
spec:
  selectors:
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: inheritor
      includeLabels:
        - app
        - foo
```

This CR is used to specify the namespaces that need to be watched and the labels that need to be inherited. The operator watches for the namespaces that match the `namespaceSelector` and updates the labels of the resources within the namespace based on the `includeLabels`.

## Deploying the Operator to a KIND cluster

### Prerequisites
- [Docker](https://docs.docker.com/get-docker/)
- [Kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
- [Kustomize](https://kubectl.docs.kubernetes.io/installation/kustomize/)

### Steps

1. Create a KIND cluster

```sh
kind create cluster
```

2. Build and load the Docker image into the KIND cluster

```sh
make docker-build 
kind load docker-image label-inheritance-operator:0.0.1
```

3. Deploy the operator to the KIND cluster

```sh
make deploy IMG=label-inheritance-operator:0.0.1
```

4. Install the CRDs

```sh
make install
```

5. Create namespace `inheritor` and a sample pod

```sh
kubectl create namespace inheritor
kubectl label namespace inheritor app=sample foo=bar
kubectl apply -f config/samples/pod.yaml
```

6. Apply the sample Custom Resource (CR) 

The sample CR is in the `config/samples` directory. It listens for the namespace `inheritor` and inherits the labels `app` and `foo` to the resources within the namespace.

```sh
kubectl apply -f config/samples/labels_v1_inheritor.yaml
```

7. Verify that the labels are inherited by the pod

```sh
kubectl get pod -n inheritor -o jsonpath='{.items[*].metadata.labels}'
```

## Running tests

To run the tests, run the following command:

```sh
make test
```

