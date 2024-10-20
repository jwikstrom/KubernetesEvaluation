## Download Runtimes

Containerd:

	wget "https://github.com/containerd/containerd/releases/download/v1.7.23/containerd-1.7.23-linux-amd64.tar.gz"
	
	sudo tar Cxzvf /usr/local containerd-1.7.23-linux-amd64.tar.gz
RunC:

	wget "https://github.com/opencontainers/runc/releases/download/v1.1.15/runc.amd64"
	
	sudo install -m 755 runc.amd64 /usr/local/sbin/runc
CNI plugins:

	wget "https://github.com/containernetworking/plugins/releases/download/v1.6.0/cni-plugins-linux-amd64-v1.6.0.tgz"
	
	mkdir -p /opt/cni/bin
	
	sudo tar Cxzvf /opt/cni/bin cni-plugins-linux-amd64-v1.6.0.tgz
