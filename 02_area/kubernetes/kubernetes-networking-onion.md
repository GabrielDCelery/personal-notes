---
title: Kubernetes networking onion
tags:
  - kubernetes
  - networking
---

# The problem

Wanted to get a better understandig of how traffic traverses in kubernetes (broken down layer-by-layer). Was using `minikube` when testing this.

# The setup

First created a series of nginx pods and a NodePort to expose them.

```sh
kubectl create deployment nginx --image=nginx --replicas=3
kubectl expose deployment nginx --port 80 --type NodePort
```

> [!WARN]
> Because I was using minikube with a docker driver there is a gotcha, had to run `minikube service nginx --url` to add an extra layer to the mix, which established a mapping between the host and docker but that is not an issue for most use cases

This created the following setup

```sh
# kubectl get deployment nginx -o wide
NAME    READY   UP-TO-DATE   AVAILABLE   AGE    CONTAINERS   IMAGES   SELECTOR
nginx   3/3     3            3           136m   nginx        nginx    app=nginx

# kubectl get service nginx -o wide
NAME    TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)        AGE   SELECTOR
nginx   NodePort   10.100.63.134   <none>        80:30761/TCP   63m   app=nginx

# kubectl get pods -o wide
NAME                     READY   STATUS    RESTARTS   AGE    IP           NODE       NOMINATED NODE   READINESS GATES
nginx-66686b6766-86xzs   1/1     Running   0          136m   10.244.0.6   minikube   <none>           <none>
nginx-66686b6766-qhg9f   1/1     Running   0          136m   10.244.0.7   minikube   <none>           <none>
nginx-66686b6766-zfd9r   1/1     Running   0          136m   10.244.0.8   minikube   <none>           <none>

# kubectl get endpoints nginx -o wide
NAME    ENDPOINTS                                   AGE
nginx   10.244.0.6:80,10.244.0.7:80,10.244.0.8:80   64m
```

# The network layer breakdown

## Layer 1: Container Port (Inside the Container)

Nginx process listening on: 0.0.0.0:80

- The nginx application inside the container listens on port 80
- This is what nginx is actually configured to use (default nginx port)
- This port is ONLY accessible inside the container itself

---

## Layer 2: Pod Network (Container → Pod)

```txt
Pod 1: 10.244.0.6:80 ─┐
Pod 2: 10.244.0.7:80 ─┼─ Each pod gets its own IP in the pod network
Pod 3: 10.244.0.8:80 ─┘
```

- Each pod gets a unique IP from the pod network (10.244.0.0/16)
- The container's port 80 is exposed on the pod's IP
- These IPs are only accessible within the Kubernetes cluster
- Pods can talk to each other directly using these IPs

How it works:

- Container port 80 is mapped to Pod IP:80
- No port translation happens here
- targetPort: 80 in the service refers to this

---

## Layer 3: Service (ClusterIP) - Load Balancer

```txt
Service ClusterIP: 10.100.63.134:80
            ↓
     [Load Balancer]
     /      |      \
    ↓       ↓       ↓
10.244.0.6:80  10.244.0.7:80  10.244.0.8:80
```

The Service provides:

1. Stable virtual IP: 10.100.63.134

- This IP never changes even if pods die/restart
- Allocated from the service CIDR (10.96.0.0/12)

2. Load balancing via kube-proxy

- kube-proxy runs on every node
- It watches the Endpoints object: 10.244.0.6:80, 10.244.0.7:80, 10.244.0.8:80
- Creates iptables rules (or IPVS) that randomly distribute traffic (round-robin or random selection)

3. Port mapping:

- Service listens on: port 80 (port: 80)
- Forwards to pods on: port 80 (targetPort: 80)

---

## Layer 4: NodePort (Cluster → Outside World)

```txt
Node IP: 192.168.49.2:30761
           ↓
      [kube-proxy]
           ↓
Service ClusterIP: 10.100.63.134:80
           ↓
       [Load balances to pods]
```

What NodePort does:

1. Opens port 30761 on the node (minikube machine at 192.168.49.2)

- Port range: 30000-32767
- Listens on the actual node's network interface

2. Forwards to ClusterIP:

- nodePort: 30761 → port: 80 (ClusterIP)
- Then ClusterIP load balances to pods

3. kube-proxy creates iptables rules:

```txt
Simplified iptables logic:
Traffic to 192.168.49.2:30761
  → DNAT to 10.100.63.134:80
  → Load balance to one of the pod IPs
```
