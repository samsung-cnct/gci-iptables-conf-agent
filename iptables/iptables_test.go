//
// Copyright © 2016 Samsung CNCT
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package iptables

import (
	"strings"
	"testing"
)

const (
	validSmallSaveBuf string = `# Generated by iptables-save v1.6.0 on Wed Dec 14 17:46:13 2016
*nat
:PREROUTING ACCEPT [258:12746]
:INPUT ACCEPT [225:10766]
:OUTPUT ACCEPT [7753:477832]
:POSTROUTING ACCEPT [7753:477832]
:DOCKER - [0:0]
:THIS-IS-A-TEST - [0:0]
-A PREROUTING -m addrtype --dst-type LOCAL -j DOCKER
-A OUTPUT ! -d 127.0.0.0/8 -m addrtype --dst-type LOCAL -j DOCKER
-A POSTROUTING -s 172.17.0.0/16 ! -o docker0 -j MASQUERADE
-A DOCKER -i docker0 -j RETURN
COMMIT
# Completed on Wed Dec 14 17:46:13 2016
# Generated by iptables-save v1.6.0 on Wed Dec 14 17:46:13 2016
*filter
:INPUT ACCEPT [30317:23318290]
:FORWARD ACCEPT [0:0]
:OUTPUT ACCEPT [27136:2316209]
:DOCKER - [0:0]
:DOCKER-ISOLATION - [0:0]
:sshguard - [0:0]
-A INPUT -j sshguard
-A FORWARD -j DOCKER-ISOLATION
-A FORWARD -o docker0 -j DOCKER
-A FORWARD -o docker0 -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
-A FORWARD -i docker0 ! -o docker0 -j ACCEPT
-A FORWARD -i docker0 -o docker0 -j ACCEPT
-A DOCKER-ISOLATION -j RETURN
-A sshguard -s 46.148.18.163/32 -j DROP
COMMIT
# Completed on Wed Dec 14 17:46:13 2016`
	validVer1421Buf string = `# Generated by iptables-save v1.4.21 on Thu Dec 15 21:55:31 2016
*filter
:INPUT ACCEPT [3:132]
:FORWARD ACCEPT [0:0]
:OUTPUT ACCEPT [4:164]
:DOCKER - [0:0]
:DOCKER-ISOLATION - [0:0]
:KUBE-FIREWALL - [0:0]
:KUBE-KEEPALIVED-VIP - [0:0]
:KUBE-SERVICES - [0:0]
[900582:466661783] -A INPUT -j KUBE-FIREWALL
[629767:164465256] -A FORWARD -j DOCKER-ISOLATION
[0:0] -A FORWARD -o docker0 -j DOCKER
[0:0] -A FORWARD -o docker0 -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
[0:0] -A FORWARD -i docker0 ! -o docker0 -j ACCEPT
[0:0] -A FORWARD -i docker0 -o docker0 -j ACCEPT
[1096990:208764058] -A OUTPUT -m comment --comment "kubernetes service portals" -j KUBE-SERVICES
[1096990:208764058] -A OUTPUT -j KUBE-FIREWALL
[629767:164465256] -A DOCKER-ISOLATION -j RETURN
[0:0] -A KUBE-FIREWALL -m comment --comment "kubernetes firewall for dropping marked packets" -m mark --mark 0x8000/0x8000 -j DROP
[0:0] -A KUBE-KEEPALIVED-VIP -m set --match-set keepalived dst --return-nomatch ! --update-counters ! --update-subcounters -j DROP
COMMIT
# Completed on Thu Dec 15 21:55:31 2016`
)

func strArrayToByteArrayArray(strArray []string) [][]byte {
	baa := make([][]byte, len(strArray))
	for i := range strArray {
		baa[i] = []byte(strArray[i])
	}
	return baa
}

func TestValidLargeSaveBuf(t *testing.T) {
	strBuf := strings.Split(validSmallSaveBuf, "\n")
	if index := ContainsRulePart(strArrayToByteArrayArray(strBuf), "-A POSTROUTING -s 172.17.0.0/16 "); index < 0 {
		t.Error("Unable to find valid postrouting rule for buffer that contains one")
	}
}

func TestBufVersion160(t *testing.T) {
	strBuf := strings.Split(validSmallSaveBuf, "\n")
	if ver := VersionCheckBuffer(strArrayToByteArrayArray(strBuf)); ver != Version160 {
		t.Error("Unable to valididate correct version for save buffer v1.6.0")
	}
}

func TestBufVersionNot160(t *testing.T) {
	strBuf := strings.Split(validVer1421Buf, "\n")
	if ver := VersionCheckBuffer(strArrayToByteArrayArray(strBuf)); ver == Version160 {
		t.Error("Unable to invalididate incorrect version for save buffer v1.6.0")
	}
}
