```
cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF

sudo modprobe overlay
sudo modprobe br_netfilter

# sysctl params required by setup, params persist across reboots
cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF

# Apply sysctl params without reboot
sudo sysctl --system
```

```
sudo swapoff -a
(crontab -l 2>/dev/null; echo "@reboot /sbin/swapoff -a") | crontab - || true
```

```
sudo apt-get update
sudo apt-get install -y apt-transport-https ca-certificates curl gpg
```
```
ls /etc/apt/keyrings
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.30/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.30/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list
```

```
sudo apt-get update

#see latest kubeadmversion
apt-cache madison kubeadm | tac 

sudo apt-get install -y kubelet kubeadm kubectl

#prevent upgrades
sudo apt-mark hold kubelet kubeadm kubectl 
```

```
sudo apt-get install -y jq
local_ip="$(ip --json addr show eth0 | jq -r '.[0].addr_info[] | select(.family == "inet") | .local')"
cat > /etc/default/kubelet << EOF
KUBELET_EXTRA_ARGS=--node-ip=$local_ip
EOF
```

##Install Containerd
download containerd-<VERSION>-<OS>-<ARCH>.tar.gz at https://github.com/containerd/containerd/releases
I am downloading containerd-1.7.15-linux-amd64.tar.gz
```
mkdir Downloads

#On host machine cmd:
scp containerd-1.7.15-linux-amd64.tar.gz joel@192.168.1.10:~/Downloads
scp containerd-1.7.15-linux-amd64.tar.gz joel@192.168.1.11:~/Downloads
scp containerd-1.7.15-linux-amd64.tar.gz joel@192.168.1.12:~/Downloads

#On each node
sudo tar Cxzvf /usr/local Downloads/containerd-1.7.15-linux-amd64.tar.gz
```

```
sudo mkdir -p /usr/local/lib/systemd/system/
sudo nano /usr/local/lib/systemd/system/containerd.service
#Paste contents from containerd.service below
```

containerd.service:
https://raw.githubusercontent.com/containerd/containerd/main/containerd.service
```
# Copyright The containerd Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

[Unit]
Description=containerd container runtime
Documentation=https://containerd.io
After=network.target local-fs.target

[Service]
ExecStartPre=-/sbin/modprobe overlay
ExecStart=/usr/local/bin/containerd

Type=notify
Delegate=yes
KillMode=process
Restart=always
RestartSec=5

# Having non-zero Limit*s causes performance problems due to accounting overhead
# in the kernel. We recommend using cgroups to do container-local accounting.
LimitNPROC=infinity
LimitCORE=infinity

# Comment TasksMax if your systemd version does not supports it.
# Only systemd 226 and above support this version.
TasksMax=infinity
OOMScoreAdjust=-999

[Install]
WantedBy=multi-user.target
```

```
systemctl daemon-reload
systemctl enable --now containerd
```

##Install RunC
Download runc.<ARCH> at: https://github.com/opencontainers/runc/releases
I am downloading runc 1.1.12
```
#On host machine cmd:
scp runc.amd64 joel@192.168.1.10:~/Downloads
scp runc.amd64 joel@192.168.1.11:~/Downloads
scp runc.amd64 joel@192.168.1.12:~/Downloads

#On each node:
sudo mkdir /usr/local/sbin/runc
sudo install -m 755 Downloads/runc.amd64 /usr/local/sbin/runc
```
Download cni-plugins-<OS>-<ARCH>-<VERSION>.tgz at https://github.com/containernetworking/plugins/releases
I am downloading cni-plugins-linux-amd64-v1.4.1.tgz
```
#On host machine cmd:
scp cni-plugins-linux-amd64-v1.4.1.tgz joel@192.168.1.10:~/Downloads
scp cni-plugins-linux-amd64-v1.4.1.tgz joel@192.168.1.11:~/Downloads
scp cni-plugins-linux-amd64-v1.4.1.tgz joel@192.168.1.12:~/Downloads

#On each node:
mkdir -p /opt/cni/bin
sudo tar Cxzvf /opt/cni/bin Downloads/cni-plugins-linux-amd64-v1.4.1.tgz
```
Add config.toml file:
```
sudo mkdir /etc/containerd
sudo touch /etc/containerd/config.toml
sudo sh -c 'containerd config default > /etc/containerd/config.toml'

systemctl restart containerd
systemctl enable containerd
```


#ON MASTER NODE ONLY:
```
IPADDR="192.168.1.10"
NODENAME=$(hostname -s)
POD_CIDR="192.168.0.0/16"

sudo kubeadm init --apiserver-advertise-address=$IPADDR  --apiserver-cert-extra-sans=$IPADDR  --pod-network-cidr=$POD_CIDR --node-name $NODENAME --ignore-preflight-errors Swap
```



