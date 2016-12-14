# gci-iptables-conf-agent

For GKE managed Kubernetes clusters a potential issue exists when using the GCP 
VPN to bridge two networks each with simmilar private IP address spaces.  The 
current manner in which the kubenet plugin generates NAT iptables creates a 
cluster MASQUERADE NAT rule that prevents correctly routing traffic between the 
two RFC 1918 address spaces.

See for reference the discussion in the Kubernetes Github [Issue 6545]
(https://github.com/kubernetes/kubernetes/issues/6545)

## Current Solution

The solution provided here is a small Go application that runs as a DaemonSet
and performs the following:
1. Discovers the Cluster IP address CIDR from the [GCI Instance Metadata](https://cloud.google.com/compute/docs/storing-retrieving-metadata). 
2. Read the Kubernetes node host's IP tables rules.
3. Insert a Cluster IP specific rule in the NAT POSTROUTING table that preceeds
   the Kubenet inserted rule.
   * Note: we do not remove/overwrite Kubenet's rule as it will try to self heal.  
4. Monitor IP tables rule forever to monitor changes - self healing as needed.

## Building
From source, create the Go static binary:
```
mkdir -p "${GOPATH}/src/github.com/samsung-cnct"
cd "${GOPATH}/src/github.com/samsung-cnct"
git clone https://github.com/samsung-cnct/gci-iptables-conf-agent.git
cd gci-iptables-conf-agent
CGO_ENABLED=0 GOOS=linux godep go build -a -ldflags '-w' -o gci_iptables_conf_agent
```
## Building the Docker Image
Build and push the docker image, replacing Quay with your target registry.
```
docker build --rm -t gci-iptables-conf-agent .
docker push quay.io/samsung_cnct/gci-iptables-conf-agent
```
