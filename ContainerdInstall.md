## Download Runtimes

### Containerd:

	wget "https://github.com/containerd/containerd/releases/download/v1.7.23/containerd-1.7.23-linux-amd64.tar.gz"
	
	sudo tar Cxzvf /usr/local containerd-1.7.23-linux-amd64.tar.gz

### RunC:

	wget "https://github.com/opencontainers/runc/releases/download/v1.1.15/runc.amd64"
	
	sudo install -m 755 runc.amd64 /usr/local/sbin/runc

### CNI plugins:

	wget "https://github.com/containernetworking/plugins/releases/download/v1.6.0/cni-plugins-linux-amd64-v1.6.0.tgz"
	
	mkdir -p /opt/cni/bin
	
	sudo tar Cxzvf /opt/cni/bin cni-plugins-linux-amd64-v1.6.0.tgz
In the file /etc/cni/net.d/ add subnet if it does not exist 
	
	"ipam": {
	 "type": "host-local",
	 "subnet": "10.244.0.0/16"
	}

### Kata

	wget "https://github.com/kata-containers/kata-containers/releases/download/3.10.1/kata-static-3.10.1-amd64.tar.xz"
	
	sudo mkdir -p /opt/kata/

	sudo tar Cxf / kata-static-3.10.1-amd64.tar.xz

	sudo ln -s /opt/kata/bin/containerd-shim-kata-v2 /usr/local/bin/containerd-shim-kata-v2
	sudo ln -s /opt/kata/bin/kata-collect-data.sh /usr/local/bin/kata-collect-data
	sudo ln -s /opt/kata/bin/kata-runtime /usr/local/bin/system/kata-runtime

	
## Configure
### Containerd:
Create configuration file:

	sudo mkdir -p /etc/containerd
	containerd config default | sudo tee /etc/containerd/config.toml
Set runtime to use in config.toml:

	[plugins]  
		[plugins."io.containerd.grpc.v1.cri"]  
			[plugins."io.containerd.grpc.v1.cri".containerd]  
				default_runtime_name = "REPLACEWITHRUNTIMENAME"
				[plugins."io.containerd.grpc.v1.cri".containerd.runtimes]
					[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.kata]
				          runtime_type = "io.containerd.kata.v2"
	
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

	
