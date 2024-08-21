# Setup K0s
## Initialize Node
**On node:**

    sudo apt update
    sudo apt upgrade -y

    sudo visudo

Add following at bottom of file:

    jw ALL=(ALL) NOPASSWD: ALL

## SSH Setup

### If windows host:
Run with correct *user@ip*

    ssh-keygen
    type .\id_rsa.pub | ssh jw@192.168.1.103 "cat >> .ssh/authorized_keys"

### If linux host:
Run with correct *user@ip*

	ssh-keygen
    ssh-copy-id -i ~/.ssh/id_rsa jw@192.168.1.103


## Install brew
**On host:**

    sudo apt update
    sudo apt upgrade -y

    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

Copy the top two commands from installation:

    (echo; echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"') >> /home/joel/.bashrc
    eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"
    
	sudo apt-get install build-essential -y

## Install kubectl

**On host:**

Choose either latest or specific release:
   

    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
	curl -LO "https://dl.k8s.io/release/$(curl -LO https://dl.k8s.io/release/v1.31.0/bin/linux/amd64/kubectl)/bin/linux/amd64/kubectl"

Then install with:

    sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

## Setup K0sctl
**On host:**
(installed v1.30.3+k0s.0)

    brew install k0sproject/tap/k0sctl

Choose one:

	k0sctl init > k0sctl.yaml
	k0sctl init --k0s > k0sctl.yaml
Then:
	
	k0sctl apply --config k0sctl.yaml -d

    mkdir ~/.kube
    k0sctl kubeconfig --config k0sctl.yaml > ~/.kube/config

Test:

    kubectl cluster-info
    kubectl get nodes
    kubectl get services
    kubectl get pods
    kubectl get all
    kubectl describe pod XXX

# Setup environment

## Prometheus - Optional
**On host**
Prometheus should already be installed via k0sctl chart.
This may however be used to see what version to use in earlier chart

	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
 	helm search repo prometheus-community

## Install Mosquitto
**On host**
	
 	helm repo add t3n https://storage.googleapis.com/t3n-helm-charts
 	helm show values t3n/mosquitto > mosquitto_values.yaml
  Edit *mosquitto_values.yaml* ~row 20 - 34

  	service:
	  type: NodePort
	  externalTrafficPolicy: Cluster
	  annotations: {}
	    # metallb.universe.tf/allow-shared-ip: pi-hole
	
	ports:
	  mqtt:
	    port: 1883
	    # sets consistent nodePort, required to set service.type=NodePort
	    nodePort: 31883
	    protocol: TCP
	  websocket:
	    port: 9090
	    protocol: TCP
Then run
		
	helm -n default upgrade --install mqtt -f mosquitto_values.yaml t3n/mosquitto







