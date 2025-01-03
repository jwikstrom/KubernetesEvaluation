

# Setup K0s
## Initialize Node
**On node:**

    sudo apt update
    sudo apt upgrade -y

    sudo visudo

Add following at bottom of file:

    jw ALL=(ALL) NOPASSWD: ALL

## SSH Setup

### If windows separate machine:
Run with correct *nodeuser@nodeip*

    ssh-keygen
    type .\.ssh\id_rsa.pub | ssh jw@192.168.1.103 "cat >> .ssh/authorized_keys"

### If linux separate machine:
Run with correct *user@ip*

	ssh-keygen
    ssh-copy-id -i ~/.ssh/id_rsa jw@192.168.1.103

### Access to node:

	ssh jw@192.168.1.103

## Install brew
**On separate:**

    sudo apt update
    sudo apt upgrade -y

    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

Copy the top two commands from installation:

    (echo; echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"') >> /home/joel/.bashrc
    eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"
    
	sudo apt-get install build-essential -y

 	brew install gcc

## Install kubectl
**On separate:**

Choose either latest or specific release:
   

    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
	curl -LO "https://dl.k8s.io/release/$(curl -LO https://dl.k8s.io/release/v1.31.0/bin/linux/amd64/kubectl)/bin/linux/amd64/kubectl"

Then install with:

    sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

## Setup K0sctl (Deprecated -  use k0s installation instead)
**On separate:**
(installed v1.30.3+k0s.0)

    brew install k0sproject/tap/k0sctl

Choose one:

	k0sctl init > k0sctl.yaml
	k0sctl init --k0s > k0sctl.yaml
 	
  	# Or copy paste contents from k0sctl.yaml from this git repo
Then:
	
	k0sctl apply --config k0sctl.yaml -d

    mkdir ~/.kube
    k0sctl kubeconfig --config k0sctl.yaml > ~/.kube/config

Test:

    kubectl cluster-info
    kubectl get nodes
    kubectl get services
    kubectl get pods
    kubectl get all -A
    kubectl describe pod XXX

### Uninstall
To uninstall, run:

	k0sctl reset

# Setup k0s without k0sctl (not currently working)
	curl --proto '=https' --tlsv1.2 https://get.k0s.sh | sudo K0S_VERSION=v1.31.2+k0s.0 sh
	mkdir  -p  /etc/k0s
	sudo sh -c "k0s config create > /etc/k0s/k0s.yaml"
	sudo k0s install controller --single --force -c /etc/k0s/k0s.yaml -v --cri-socket=remote:unix:///var/run/containerd/containerd.sock
	--cri-socket /run/containerd/containerd.sock --pod-network-cidr=10.244.0.0/16
	# --cri-socket=remote:unix:///var/run/containerd/containerd.sock
	sudo systemctl daemon-reload
	sudo k0s start
	
	sudo k0s status
	sudo k0s kubectl get all -A

Re-install:

	sudo systemctl stop k0scontroller
	sudo systemctl disable k0scontroller
	sudo rm /etc/systemd/system/k0scontroller.service
	sudo systemctl daemon-reload

	sudo k0s install controller --single -c /etc/k0s/k0s.yaml -v --cri-socket=remote:unix:///var/run/containerd/containerd.sock


crictl
	
	wget "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.31.1/crictl-v1.31.1-linux-386.tar.gz"
	sudo tar zxvf crictl-v1.31.1-linux-386.tar.gz -C /usr/local/bin/
	

# Setup environment

## Prometheus
**On separate:**

Prometheus is used for monitoring.

	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm repo update
To see latest versions - *optional*:

	helm search repo prometheus-community
To get the values yaml file:

	helm show values prometheus-community/kube-prometheus-stack > promvalues.yaml

Edit the file so that prometheus.service.type = NodePort 
The port should already be *30090*, otherwise edit prometheus.service.nodePort = 30090

Then install with:

	helm install prometheus prometheus-community/kube-prometheus-stack -f promvalues.yaml
### Uninstall
	helm uninstall prometheus prometheus-community/kube-prometheus-stack

## Grafana
On a machine that is on the same subnet as the node with prometheus:

 1. Follow instructions to install grafana on [prometheus docs](https://prometheus.io/docs/visualization/grafana/)
	 - Default Access to Grafana on http://localhost:3000/
 3. Add Data source Prometheus for http://192.168.1.103:30090/
	- http://\<Node Ip>:\<Prometheus NodePort port>
 4. Import dashboards [ID 13332](https://grafana.com/grafana/dashboards/13332-kube-state-metrics-v2/) and [ID 1860](https://grafana.com/grafana/dashboards/1860-node-exporter-full/)

(login admin/joel)

## Install Mosquitto
**On separate:**
	
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

### Testing the mosquitto broker
**On separate:**

	sudo snap install mosquitto

open two terminals and run one command in each:

 	mosquitto_sub -h 192.168.1.103 -p 31883 -t testingtopic
	mosquitto_pub -h 192.168.1.103 -p 31883 -t testingtopic -m "test"
	
# Setup Subscriber
 ## Docker
Build the docker image in the same directory as dockerfile:

	docker build -t  mqtt_subscriber .
	
Test container:

	docker run --rm  mqtt_subscriber
Push to Docker hub:

	docker login
	docker tag mqtt_subscriber bananpannkaka/mqtt_subscriber:latest
	docker push bananpannkaka/mqtt_subscriber:latest

Updating docker image:

	docker build -t bananpannkaka/mqtt_subscriber:latest .
	docker push bananpannkaka/mqtt_subscriber:latest
## Kubernetes

Make sure you have a deployment.yaml file for the subscriber, then run:

	kubectl apply -f subscriberDeployment.yaml

Verify it is up and running then to see logs run:

	kubectl logs -l app=mqtt-subscriber
	#----------OR---------
	kubectl logs <pod-name>

Or to see live logs:

	kubectl logs -f -l app=mqtt-subscriber
	#----------OR---------
	kubectl logs -f <pod-name>

To update pod with a new docker image remove the pod wait until a new pod appears

	kubectl delete pod -l app=mqtt-subscriber
To delete the deployment, do:

	kubectl delete deployment mqtt-subscriber-deployment
