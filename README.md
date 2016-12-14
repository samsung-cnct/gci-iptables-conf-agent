# gci-iptables-conf-agent

For GKE managed Kubernetes clusters a potential issue exists when using the GCP 
VPN to bridge two private IP address spaces.  With the current manner in which 
kubenet generates the NAT iptables rules for nodes in the cluster a MASQUERADE 
NAT is applied that prevents correctly routing traffic between the to RFC 1918
spaces.

See for reference the discussion in the Kubernetes Issue: 
https://github.com/kubernetes/kubernetes/issues/6545

## Current Solution

yadda yadda yadda 

## Building

```
mkdir -p "${GOPATH}/src/github.com/samsung-cnct"
cd "${GOPATH}/src/github.com/samsung-cnct"
git clone https://github.com/samsung-cnct/gci-iptables-conf-agent.git
cd gci-iptables-conf-agent
CGO_ENABLED=0 GOOS=linux godep go build -a -ldflags '-w' -o gci_iptables_conf_agent
```
## Building the Docker Image

```
docker build --rm -t gci-iptables-conf-agent .
```
