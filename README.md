# gci-iptables-conf-agent

For GKE managed Kubernetes clusters a potential issue exists when using the GCP 
VPN to bridge two networks each with simmilar private IP address spaces.  The 
current manner in which the kubenet plugin generates NAT iptables creates a 
cluster MASQUERADE NAT rule that prevents correctly routing traffic between the 
two RFC 1918 address spaces.

See for reference the discussion in the Kubernetes Github [Issue 6545](https://github.com/kubernetes/kubernetes/issues/6545)

## Current Solution

The solution provided here is a small Go application that runs as a DaemonSet
and performs the following:

1. Discovers the Cluster IP address CIDR from the [GCI Instance Metadata](https://cloud.google.com/compute/docs/storing-retrieving-metadata). 
2. Read the Kubernetes node host's IP tables rules.
3. Insert a Cluster IP specific rule in the NAT POSTROUTING table that preceeds
   the Kubenet inserted rule.
   * Note: Do not modify/remove Kubenet's rule as Kubenet will try to fix it.  
4. Monitor IP tables rule forever to monitor changes - self healing as needed.

## Building
From source, create the Go static binary:
```
$ mkdir -p "${GOPATH}/src/github.com/samsung-cnct"
$ cd "${GOPATH}/src/github.com/samsung-cnct"
$ git clone https://github.com/samsung-cnct/gci-iptables-conf-agent.git
$ cd gci-iptables-conf-agent
$ CGO_ENABLED=0 GOOS=linux godep go build -a -ldflags '-w' -o gci_iptables_conf_agent
```
## Building the Docker Image
Build and push the docker image, replacing Quay with your target registry.
```
$ docker build --rm --tag quay.io/samsung_cnct/gci-iptables-conf-agent .
$ docker push quay.io/samsung_cnct/gci-iptables-conf-agent:latest
```

## Helm Chart
This project is also packaged as a Helm Chart here: [GCI IPTables Helm Chart](https://github.com/samsung-cnct/k2-charts/tree/master/gci-iptables-conf-agent)

The Helm chart can be deployed in to a GKE Kubernets cluster with the following commands:

```
$ helm repo add cnct http://atlas.cnct.io
$ helm install cnct/gci-iptables-conf-agent
$ helm list
NAME           	REVISION	UPDATED                 	STATUS  	CHART                        
veering-buffoon	1       	Fri Dec 16 16:31:13 2016	DEPLOYED	gci-iptables-conf-agent-0.1.0

rastop:templates sostheim$ helm status veering-buffoon
LAST DEPLOYED: Fri Dec 16 16:31:13 2016
NAMESPACE: default
STATUS: DEPLOYED

RESOURCES:
==> extensions/DaemonSet
NAME             DESIRED   CURRENT   NODE-SELECTOR   AGE
iptables-agent   8         8         <none>          1d
```

## Checking The Results
First, is the agent running on the node that you are checking?
```
$ kubectl get pod | grep iptables-agent | cut -d' ' -f1 | xargs kubectl describe pod | grep Node
Node:		gke-zonar-production-cluster-primary-e8df4da1-4zd1/10.142.0.12
Node:		gke-zonar-production-cluster-primary-e8df4da1-tg0m/10.142.0.7
Node:		gke-zonar-production-cluste-secondary-4f6ea563-lnkv/10.142.0.19
Node:		gke-zonar-production-cluster-primary-e8df4da1-jff9/10.142.0.6
Node:		gke-zonar-production-cluste-secondary-a2b2807c-x3bd/10.142.0.13
Node:		gke-zonar-production-cluste-secondary-a2b2807c-1mkk/10.142.0.16
Node:		gke-zonar-production-cluster-primary-37de06ee-evjx/10.142.0.5
Node:		gke-zonar-production-cluste-secondary-4f6ea563-f7r3/10.142.0.8
Node:		gke-zonar-production-cluster-primary-37de06ee-r30t/10.142.0.17
Node:		gke-zonar-production-cluste-secondary-a2b2807c-gccw/10.142.0.2
Node:		gke-zonar-production-cluste-secondary-4f6ea563-pp5g/10.142.0.9
Node:		gke-zonar-production-cluster-primary-e8df4da1-c6tz/10.142.0.18
Node:		gke-zonar-production-cluste-secondary-4f6ea563-2d46/10.142.0.14
Node:		gke-zonar-production-cluster-primary-37de06ee-655p/10.142.0.11
Node:		gke-zonar-production-cluster-primary-37de06ee-fejp/10.142.0.4
Node:		gke-zonar-production-cluste-secondary-a2b2807c-k0n1/10.142.0.3
```

If the Node in question doesn't show up in the list above, then the agent isn't running on it.

Once you know the node is present, running the following command:

```
$ kubectl logs --tail=10 iptables-agent-1pw1z
Kubenet Rule Index: 80, Cluster IP Rule Index: 79
IP Tables NAT table check: ok
Kubenet Rule Index: 80, Cluster IP Rule Index: 79
IP Tables NAT table check: ok
Kubenet Rule Index: 80, Cluster IP Rule Index: 79
IP Tables NAT table check: ok
Kubenet Rule Index: 80, Cluster IP Rule Index: 79
IP Tables NAT table check: ok
Kubenet Rule Index: 80, Cluster IP Rule Index: 79
IP Tables NAT table check: ok

$ gcloud compute ssh gke-zonar-production-cluster-primary-e8df4da1-4zd1 -- sudo iptables-save | grep -n "SNAT for outbound"
79:-A POSTROUTING ! -d 10.64.0.0/14 -m comment --comment "ClusterIP: SNAT for outbound traffic" -m addrtype ! --dst-type LOCAL -j MASQUERADE
80:-A POSTROUTING ! -d 10.0.0.0/8 -m comment --comment "kubenet: SNAT for outbound traffic from cluster" -m addrtype ! --dst-type LOCAL -j MASQUERADE
```

The results above are correct if the log output is in agreement with the rules dispalyed byt the `ssh | grep -n`.  More specifically, the "ClusterIP: SNAT" rule preceeds the more general rule "kubenet: SNAT" rule in order in the iptables output.
