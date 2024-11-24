## Download Runtimes

### Containerd:

	

~~wget "https://github.com/containerd/containerd/releases/download/v1.7.23/containerd-1.7.23-linux-amd64.tar.gz"~~

    wget "https://github.com/containerd/containerd/releases/download/v2.0.0/containerd-2.0.0-linux-amd64.tar.gz"

	
~~sudo tar Cxzvf /usr/local containerd-1.7.23-linux-amd64.tar.gz~~

    sudo tar Cxzvf /usr/local containerd-2.0.0-linux-amd64.tar.gz

	sudo mkdir -p /usr/local/lib/systemd/system
	cd /usr/local/lib/systemd/system
	sudo wget "https://raw.githubusercontent.com/containerd/containerd/main/containerd.service"
	
	sudo mkdir -p /etc/containerd
	sudo sh -c "containerd config default > /etc/containerd/config.toml"

	systemctl daemon-reload
	systemctl enable --now containerd

	sudo iptables -P FORWARD ACCEPT

Useful links:

- https://github.com/containerd/containerd/blob/main/docs/getting-started.md
- https://github.com/containerd/containerd/issues/7975
- https://github.com/containerd/containerd/blob/main/docs/man/containerd-config.toml.5.md

### RunC:

	wget "https://github.com/opencontainers/runc/releases/download/v1.1.15/runc.amd64"
	
	sudo install -m 755 runc.amd64 /usr/local/sbin/runc

### CNI plugins:

	wget "https://github.com/containernetworking/plugins/releases/download/v1.6.0/cni-plugins-linux-amd64-v1.6.0.tgz"
	
	sudo mkdir -p /opt/cni/bin
	
	sudo tar Cxzvf /opt/cni/bin cni-plugins-linux-amd64-v1.6.0.tgz
~~In the file /etc/cni/net.d/ add subnet if it does not exist~~ 
	
	"ipam": {
	 "type": "host-local",
	 "subnet": "10.244.0.0/16"
	}

### Kata

	wget "https://github.com/kata-containers/kata-containers/releases/download/3.10.1/kata-static-3.10.1-amd64.tar.xz"
	
	

    sudo mkdir -p /opt/kata/
    sudo tar -C /opt/kata/ -xf kata-static-3.10.1-amd64.tar.xz
    sudo ln -s /opt/kata/bin/containerd-shim-kata-v2 /usr/local/bin/containerd-shim-kata-v2
    sudo ln -s /opt/kata/bin/kata-collect-data.sh /usr/local/bin/kata-collect-data
    sudo ln -s /opt/kata/bin/kata-runtime /usr/local/bin/kata-runtime

Check if system can use kata containers:

	sudo /opt/kata/bin/kata-runtime check
Useful links:

 - https://github.com/kata-containers/kata-containers/blob/main/docs/install/container-manager/containerd/containerd-install.md
 - https://github.com/kata-containers/kata-containers/blob/main/docs/how-to/containerd-kata.md

	

### Crun

	wget "https://github.com/containers/crun/releases/download/1.18/crun-1.18-linux-amd64"
	
## Configure
### Containerd:
Create configuration file:

	sudo mkdir -p /etc/containerd
	containerd config default | sudo tee /etc/containerd/config.toml
	sudo nano /etc/containerd/config.toml
Set runtime to use in config.toml:

	[plugins]  
		[plugins."io.containerd.grpc.v1.cri"]  
			[plugins."io.containerd.grpc.v1.cri".containerd]  
				default_runtime_name = "REPLACEWITHRUNTIMENAME"
				[plugins."io.containerd.grpc.v1.cri".containerd.runtimes]
				
					[plugins.'io.containerd.cri.v1.runtime'.containerd.runtimes.kata]
				          runtime_type = "io.containerd.kata.v2"
				          privileged_without_host_devices = true
				          privileged_without_host_devices_all_devices_allowed = true
				
				          [plugins.'io.containerd.cri.v1.runtime'.containerd.runtimes.kata.options]
				            BinaryName = '/usr/local/bin/containerd-shim-kata-v2'
				            SystemdCgroup = true
				          
				    [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.crun]
						runtime_type = "io.containerd.runc.v2"
						[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.crun.options]
							BinaryName = "/usr/local/sbin/crun"
	
	# Runc
	io.containerd.runc.v2
	# Kata
	io.containerd.kata.v2


Start:

	cd /usr/lib/systemd/system
	sudo wget "https://raw.githubusercontent.com/containerd/containerd/main/containerd.service"

	systemctl daemon-reload
	systemctl enable --now containerd
Restart:

	systemctl daemon-reload
	sudo systemctl restart containerd
