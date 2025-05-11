---
title: Connect kubernetes talos worker node to mini pc control plane
author: GaborZeller
date: 2025-05-11T10-20-59Z
tags:
draft: true
---

# Connect kubernetes talos worker node to mini pc control plane

To connect a Talos worker node to your Rancher-managed Kubernetes control plane, you'll need to follow these steps:

1. First, on your Rancher (Mini PC) side:

```bash
# Get the control plane endpoint (IP address of your mini PC)
CONTROL_PLANE_IP=<your-mini-pc-ip>

# Get the Kubernetes join token from Rancher
# You can find this in the Rancher UI under Cluster -> Registration
CLUSTER_TOKEN=<your-cluster-token>
```

2. Create a Talos configuration file for the worker node:

```bash
# Generate the Talos worker configuration
talosctl gen config \
    --with-docs=false \
    --with-examples=false \
    your-cluster-name \
    https://${CONTROL_PLANE_IP}:6443

# This will create two files:
# - controlplane.yaml
# - worker.yaml
```

3. Edit the worker.yaml file to include the proper settings:

```yaml
machine:
  type: worker
  certSANs:
    - ${WORKER_IP}
cluster:
  controlPlane:
    endpoint: https://${CONTROL_PLANE_IP}:6443
  clusterName: your-cluster-name
  network:
    cni:
      name: custom
  token: ${CLUSTER_TOKEN}
  discovery:
    enabled: true
```

4. Apply the configuration to your Talos worker node:

```bash
# Set the worker node IP
WORKER_IP=<your-worker-node-ip>

# Apply the configuration
talosctl --talosconfig=./talosconfig \
    --nodes ${WORKER_IP} \
    apply-config --insecure \
    --file worker.yaml
```

5. Bootstrap the worker node:

```bash
# Bootstrap the node
talosctl --talosconfig=./talosconfig \
    --nodes ${WORKER_IP} \
    bootstrap
```

6. Verify the connection:

```bash
# Check node status in Rancher
kubectl get nodes
```

Important notes:

- Make sure both machines can communicate over the network
- Ensure required ports are open (6443 for API server, 50000 for Talos)
- The Rancher cluster must be configured to accept new nodes
- You might need to adjust firewall settings on both machines

If you encounter issues:

- Check Talos logs: `talosctl logs`
- Verify network connectivity between nodes
- Ensure the cluster token is correct
- Check if all required ports are open

Remember to replace placeholder values:

- `${CONTROL_PLANE_IP}` with your mini PC's IP
- `${WORKER_IP}` with your Talos machine's IP
- `${CLUSTER_TOKEN}` with the actual token from Rancher
- `your-cluster-name` with your desired cluster name
