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

# Setup environment
