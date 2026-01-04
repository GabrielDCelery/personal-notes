---
title: Delete all kubernetes resources in namespace
tags:
  - kubernetes
---

# The problem

Just a handy reminder on how to delete all resources from a namespace.

# The solution

1. List everything

```sh
kubectl get all -n <namespace>

NAME                         READY   STATUS    RESTARTS   AGE
pod/nginx-66686b6766-86xzs   1/1     Running   0          162m
pod/nginx-66686b6766-qhg9f   1/1     Running   0          162m
pod/nginx-66686b6766-zfd9r   1/1     Running   0          162m

NAME                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)        AGE
service/kubernetes   ClusterIP   10.96.0.1       <none>        443/TCP        3h12m
service/nginx        NodePort    10.100.63.134   <none>        80:30761/TCP   88m

NAME                    READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nginx   3/3     3            3           162m

NAME                               DESIRED   CURRENT   READY   AGE
replicaset.apps/nginx-66686b6766   3         3         3       162m
```

---

2. Delete everything

```sh
kubectl delete all --all -n <namespace>
```
