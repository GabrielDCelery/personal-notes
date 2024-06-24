### Chapter 20 - Configuring NFS server

#### Basics of NFS server

The idea of an NFS server is to be able to share file systems across different machines on a private network. To set up a server you generally need the followings:

1. set up the network
2. start NFS service (nfs-server on Linux)
3. choose what to share on the file server
4. set up security on the server (there are many fine grated solutions you can choose from, you can specify which computers can mount what and whether they can have read/write access etc...)
5. mount file system on the client

#### Installing NFS server on host

To make the NFS server work first some packages need to be installed.

```sh
apt update
apt install nfs-kernel-server
```

Once installed `systemctl` should automatically start the service:

```sh
$ systemctl status nfs-kernel-server.service

● nfs-server.service - NFS server and services
     Loaded: loaded (/usr/lib/systemd/system/nfs-server.service; enabled; preset: enabled)
     Active: active (exited) since Mon 2024-06-24 19:19:40 UTC; 30min ago
   Main PID: 2250 (code=exited, status=0/SUCCESS)
        CPU: 6ms

Jun 24 19:19:39 ip-172-31-19-54 systemd[1]: Starting nfs-server.service - NFS server and services...
Jun 24 19:19:39 ip-172-31-19-54 exportfs[2248]: exportfs: can't open /etc/exports for reading
Jun 24 19:19:40 ip-172-31-19-54 systemd[1]: Finished nfs-server.service - NFS server and services.
```

#### Configuring NFS server on host

In order to configure the host to share directories it is done through configuring the `/etc/exports` file. The structure looks like this:

```sh
directory_to_share    client(share_option1,...,share_optionN)

# Examples
# /var/nfs/general    client_ip(rw,sync,no_subtree_check)
# /home               client_ip(rw,sync,no_root_squash,no_subtree_check)
# /srv/nfs4           gss/krb5i(rw,sync,fsid=0,crossmnt,no_subtree_check)
# /srv/nfs4/homes     gss/krb5i(rw,sync,no_subtree_check)
# /cal                *linuxtoys.net(rw)
# /pub                *(ro, insecure,all_squash)
# /home               maple(rw,root_squash) spruce(rw,root_squash)
```

A few details on some of the options:

- `rw` - option gives read/write access to the client
- `sync` - this option forces NFS to write the changes to the file system before replying to the client (recommended for stability)
- `no_subtree_check` - subtree checking is the process where the host has to ensure on every request if the file is still available. This can cause issues when a file gets renamed while the client still has it open so generally it is a safe practice to disable it
- `no_root_squash` - by default the NFS file system will "convert" a client root user to a non-privilidged user to prevent the client to be able to use the file system as a root user. This setting is there to disable this default behaviour
- `insecure` - allows computers that don't even have a secure NFS port to have access to the directory
- `all_squash` - squashes every user to be a nobody user on the shared file system

Once we edited the `/etc/exports` file to our liking we have to restart the NFS server.

```sh
systemctl restart nfs-kernel-server
```

Some tips and tricks:

- In the above examples if you have a machine called `oak` and you mount the `maple` and `spruce` machine's `/host` folders on oak's `/host` then if you set up the same users with the same user IDs on all amchines then the same users can have access to their own files regardless on which machine they are using.
- In the above examples `*` allows for any machine to connect to the file share in read only mode and where all priviliges are being removed. Useful settings for publicly shared directories

##### Hostnames in /etc/exports

There are many ways to specify allowed client host namaes in a NFS.

1. `Individual domain` - use a TCP/IP hostname or IP address if the client machine is on your local network

   - maple
   - maple.handsonhistory.com
   - 10.0.0.11

2. `IP network` - you can speccify a range of IP adresses or netmask

   - 10.0.0.0/255.0.0.0 172.16.0.0/255.255.0.0
   - 192.168.18.0/24

3. `TCP/IP domain` - you can use domain levels to allow clients to connect to the NFS

   - \*.handsonhistory.com
   - \*craft.handsonhistory.com
   - ???.handsonhistory.com
