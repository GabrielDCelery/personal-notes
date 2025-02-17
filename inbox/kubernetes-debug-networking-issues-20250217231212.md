---
title: Kubernetes debug networking issues
author: GaborZeller
date: 2025-02-17T23-12-12Z
tags:
draft: true
---

# Kubernetes debug networking issues

## List pods and services to see what are thr IP addresses

```sh
❯ k get pods -o wide
NAME                                        READY   STATUS    RESTARTS   AGE   IP            NODE            NOMINATED NODE   READINESS GATES
mongo-express-deployment-7c86595bd4-cg4xq   1/1     Running   0          20m   10.244.0.51   talos-q9t-a11   <none>           <none>
mongodb-deployment-6d9d7c68f6-dmvbm         1/1     Running   0          45m   10.244.0.47   talos-q9t-a11   <none>           <none>
```

```sh
❯ k get services -o wide
NAME                    TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE    SELECTOR
kubernetes              ClusterIP   10.96.0.1       <none>        443/TCP          2d9h   <none>
mongo-express-service   NodePort    10.111.17.144   <none>        8081:30081/TCP   10m    app=mongo-express
mongodb-service         ClusterIP   10.98.174.244   <none>        27017/TCP        10m    app=mongodb
```

### Check nslookup from inside the container that is trying to connect

```sh
✗ kubectl exec mongo-express-deployment-7c86595bd4-cg4xq -- nslookup mongodb-service.default.svc.cluster.local
E0217 22:39:28.512290   71692 websocket.go:296] Unknown stream id 1, discarding message
Server:         10.96.0.10
Address:        10.96.0.10:53


Name:   mongodb-service.default.svc.cluster.local
Address: 10.98.78.54
```

### Check if /etc/resolv.conf has the expected values

```sh
mongo-express-deployment-7c86595bd4-s52bq:/app# cat /etc/resolv.conf
search default.svc.cluster.local svc.cluster.local cluster.local
nameserver 10.96.0.10
options ndots:5
```

### Use netstat to see if we have an established connection

```sh
✗ k exec -it mongo-express-deployment-7c86595bd4-cg4xq -- /bin/bash

mongo-express-deployment-7c86595bd4-cg4xq:/app# echo $ME_CONFIG_MONGODB_SERVER
mongodb-service

mongo-express-deployment-7c86595bd4-cg4xq:/app# netstat mongodb-service
Active Internet connections (w/o servers)
Proto Recv-Q Send-Q Local Address           Foreign Address         State
tcp        0      0 mongo-express-deployment-7c86595bd4-cg4xq:50296 mongodb-service.default.svc.cluster.local:27017 ESTABLISHED
tcp        0      0 mongo-express-deployment-7c86595bd4-cg4xq:50308 mongodb-service.default.svc.cluster.local:27017 ESTABLISHED
tcp        0      0 mongo-express-deployment-7c86595bd4-cg4xq:51444 mongodb-service.default.svc.cluster.local:27017 ESTABLISHED
Active UNIX domain sockets (w/o servers)
Proto RefCnt Flags       Type       State         I-Node Path
```
