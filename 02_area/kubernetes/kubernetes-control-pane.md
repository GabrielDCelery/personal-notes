# Install WSL2

# Install Ubuntu on WSL2

```sh
wsl --list
wsl --unregister <<distrotounregister>
```

# Install Docker desktop

"Use the WSL2 Base engine"

```sh
sudo usermod -aG docker $USER
# after running the command reboot the terminal
```

# Install Rancher desktop

In preferences -> integrations -> set Ubuntu WSL

# Open ports

```sh

- "6443" # Kubernetes API server
- "2379" # etcd server client API
- "2380" # etcd server client API
- "10250" # Kubelet API
- "10259" # kube-scheduler
- "10257" # kube-controller-manager
```

```sh
nc -zv 10.83.16.98 6443

kubectl proxy --address='0.0.0.0' --port=8001 --accept-hosts='.*'
- `kubectl proxy`: Starts a proxy server that acts as an intermediary between your local machine and the Kubernetes API server

- `--address='0.0.0.0'`: Makes the proxy listen on all network interfaces of the machine
  - This means the proxy will be accessible from any IP address that can reach the machine

- `--port=8001`: Sets the proxy server to listen on port 8001

- `--accept-hosts='*'`: Configures the proxy to accept requests from any hostname
  - This is a permissive setting that doesn't restrict which hosts can connect to the proxy
   ssh -L 6443:localhost:6443 user@DOCKER_DESKTOP_IP


- `ssh`: The SSH client command
- `-L 6443:localhost:6443`: This is the port forwarding configuration
  - `6443`: The local port on your machine
  - `localhost`: The hostname as seen from the SSH server (Docker Desktop)
  - `6443`: The remote port on the Docker Desktop machine
- `user@DOCKER_DESKTOP_IP`: The username and IP address of the Docker Desktop machine you're connecting to


```

```sh
sudo apt update
sudo apt install openssh-server -y
sudo systemctl enable ssh
sudo sustemctl start ssh
sudo systemctl status ssh
```

```sh
ssh-copy-id <user>@<machine>
```

```sh
sudo visudo -f /etc/sudoers.d/username
username ALL=(ALL) NOPASSWD: ALL
```

```sh
cloud-localds cloud-init.iso user-data.yaml meta-data.yaml
```

## Open windows port using Windows Firewall with Advanced Security GUI

```sh
Get-Service WinRM
Start-Service WinRM

Set-Service -Name WinRM -StartupType Automatic

winrm quickconfig


PS C:\WINDOWS\system32> winrm quickconfig
WinRM service is already running on this machine.
WinRM is not set up to allow remote access to this machine for management.
The following changes must be made:

Enable the WinRM firewall exception.
Configure LocalAccountTokenFilterPolicy to grant administrative rights remotely to local users.

Make these changes [y/n]? y

WinRM has been updated for remote management.

WinRM firewall exception enabled.
Configured LocalAccountTokenFilterPolicy to grant administrative rights remotely to local users.
```

```sh
1. Press **Win + R**, type **wf.msc** and press Enter
2. Select **Inbound Rules** from the left panel
3. Click **New Rule** from the right panel
4. Select **Port** and click **Next**
5. Choose **TCP** and enter **5985** in the "Specific local ports" field
6. Click **Next**
7. Select **Allow the connection** and click **Next**
8. Select which network profiles this rule applies to (Domain/Private/Public)
9. Name the rule (e.g., "WinRM HTTP 5985") and click **Finish**

```
