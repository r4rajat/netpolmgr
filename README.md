<h1 align="center">Netpolmgr - Backup and Restore PVC Custom k8s Validation Webhook</h1>

---


## 📝 Table of Contents

- [About](#about)
- [Getting Started](#getting_started)
- [Running the Code](#run)
- [Authors](#authors)
- [Acknowledgments](#acknowledgement)

## 🧐 About <a name = "about"></a>

The Netpolmgr custom kubernetes Validation Webhook is written primarily in go lang. This Validation Webhook Validates in case a label of a pod is edited, and it exists in some network policy, it doesn't let user edit that label.

## 🏁 Getting Started <a name = "getting_started"></a>

These instructions will get you the project up and running on your local machine for development and testing purposes. See [Running the Code](#run) for notes on how to deploy the project on a Local System or on a Kubernetes Server.

### Prerequisites

To run/test the Netpolmgr Validation Webhook on Minikube, first we need to install following Software Dependencies.

- [Go](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/)
- [Minikube](https://minikube.sigs.k8s.io/docs/start/)

Once above Dependencies are installed we can move with [further steps](#installing)

### Installing <a name = "installing"></a>

A step by step series of examples that tell you how to get a development env running.

#### Step 1: Install Project related Dependencies
```
go mod tidy
```

#### Step 2: Running a 2 Node Mock Kubernetes Server Locally using minikube
```
minikube start --nodes 2
```

#### Step 3: Create Service Account, Role and Role Binding:
```
kubectl create -f manifests/sa.yaml
kubectl create -f manifests/role.yaml
kubectl create -f manifests/rb.yaml
```

#### Step 4: Setting Up Certificates for HTTPS

```
kubectl create -f manifests/certs/secret.yaml
```


#### Step 5: Creating Deployments and Service for Netpolmgr
```
kubectl create -f manifests/netpolmgr.yaml
kubectl create -f manifests/service.yaml
```

#### Step 6: Creating Validation Webhook 
```
kubectl create -f manifests/validation-pod-label.yaml
```

#### Step 7: Creating Test deployment and network policy
```
kubectl create -f manifests/allow-network-policy.yaml
kubectl create -f manifests/nginx.yaml
```



## 🔧 Running the Code <a name = "run"></a>

```
kubectl edit pod/nginx
```
Try to edit label ```app: nginx``` to ```app: nginx-test```
Netpolmgr will restrict the changes as app:nginx is mentioned in network policy which we created.
<br>
Try to add some new labels in the nginx pod ```role: frontend```, it will allow to add this

## ✍️ Authors <a name = "authors"></a>

- [@r4rajat](https://github.com/r4rajat) - Implementation

## 🎉 Acknowledgements <a name = "acknowledgement"></a>

- References
    - https://pkg.go.dev/k8s.io/client-go
    - https://pkg.go.dev/k8s.io/apimachinery
    - https://pkg.go.dev/k8s.io/apiserver