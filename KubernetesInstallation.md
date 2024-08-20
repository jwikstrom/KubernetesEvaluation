NODE-------------------------------- init

sudo apt update
sudo apt upgrade -y

sudo visudo
(add at bottom of file):
jw ALL=(ALL) NOPASSWD: ALL

(windows host)-------
ssh-keygen
type .\id_rsa.pub | ssh jw@192.168.1.103 "cat >> .ssh/authorized_keys"
---------------------
(linux host)---------
ssh-keygen
ssh-copy-id -i ~/.ssh/id_rsa jw@192.168.1.103
---------------------


HOST-------------------------------- -Install brew
sudo apt update
sudo apt upgrade -y

/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
--copy two commands from installation:
(echo; echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"') >> /home/joel/.bashrc
eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"
    
sudo apt-get install build-essential -y

HOST-------------------------------- Install kubectl (choose one)
   curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
   curl -LO "https://dl.k8s.io/release/$(curl -LO https://dl.k8s.io/release/v1.31.0/bin/linux/amd64/kubectl)/bin/linux/amd64/kubectl"

sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

HOST-------------------------------- -Install & setup k0sctl (installed v1.30.3+k0s.0)
brew install k0sproject/tap/k0sctl
k0sctl init > k0sctl.yaml
(k0sctl init --k0s > k0sctl.yaml)
k0sctl apply --config k0sctl.yaml -d

mkdir ~/.kube
k0sctl kubeconfig --config k0sctl.yaml > ~/.kube/config

kubectl cluster-info
kubectl get nodes

END OF STEP 1: Install K0s

