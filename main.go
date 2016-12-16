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
package main

import (
	"os"

	iptables "github.com/samsung-cnct/gci-iptables-conf-agent/iptables"

	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// GCI Metadata Server Default Values
	gciMetadataFlavorHeader      = "Metadata-Flavor"
	gciMetadataFlavorHeaderValue = "Google"
	gciDefaultScheme             = "http"
	gciDefaultAuthority          = ""

	gciDefaultMetadataURN = "metadata.google.internal/computeMetadata/v1/"
	gciInstanceResource   = "instance/"
	gciProjectResource    = "project/"
	gciAttriburtes        = "attributes/"

	// A directory of custom metadata values passed to the instance during startup or shutdown.
	gciDefaultInstanceAttributes = gciDefaultMetadataURN + gciInstanceResource + gciAttriburtes

	// A directory of custom metadata values set for this project
	gciDefaultProjectAttributes = gciDefaultMetadataURN + gciProjectResource + gciAttriburtes

	// kube-env instance attribute
	gciDefaultKubeEnvAttribute = "kube-env"

	// Optional query parameters
	gciDefaultQueryAltFormat = "alt=json"
	gciDefaultQueryRecursive = "recursive=true"

	// Our primary attribute collection of interest.
	gciDefaultURI = gciDefaultScheme + "://" + gciDefaultInstanceAttributes + gciDefaultKubeEnvAttribute

	// The environment key we wish to locate
	envClusterIPRangeCIDR = "CLUSTER_IP_RANGE"

	// IP Tables NAT Table POSTROUTING MASQUERADE Chain Rule Elements
	natPostRoutingPrefix = "-A POSTROUTING"
	natPostRoutingSuffix = "-m addrtype ! --dst-type LOCAL -j MASQUERADE"
	kubenetNATChainRule  = natPostRoutingPrefix + " ! -d 10.0.0.0/8"
	kubenetSNATComment   = "kubenet: SNAT for outbound traffic from cluster"
	clusterIPSNATComment = " -m comment --comment \"ClusterIP: SNAT for outbound traffic\" "
)

type Indicies struct {
	kubenet, cluster int
}

var (
	clusterIP string
	interval  int
)

func init() {
	log.SetFlags(0)
	log.Print("gci-iptables-conf-agent: initialization")

	if !CheckIPTablesVersion() {
		log.Fatal("Can't continue without a supported version of host system 'iptables' command")
	}

	body, err := getKubEnvInstanceAttributes()
	if err != nil {
		log.Fatal(err)
	}

	clusterIP = getBufferKeyValue(body, envClusterIPRangeCIDR)
	if len(clusterIP) == 0 {
		log.Fatal("Can't continue without a valid value for Cluster IP Range CIDR")
	}
	log.Print(fmt.Sprintf("Working Cluster IP Range CIDR: %s\n", clusterIP))

	interval, err = strconv.Atoi(os.Getenv("IPTABLES_CHECK_INTERVAL"))
	if err != nil {
		interval = 60
	}
	log.Print(fmt.Sprintf("Working iptables check interval: %d seconds\n", interval))
}

func getKubEnvInstanceAttributes() ([]byte, error) {

	req, err := http.NewRequest("GET", gciDefaultURI, nil)
	if err != nil {
		log.Print(fmt.Sprintf("getKubEnvInstanceAttributes: new client request error: %v\n", err))
		return nil, err
	}
	req.Header.Add(gciMetadataFlavorHeader, gciMetadataFlavorHeaderValue)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print(fmt.Sprintf("getKubEnvInstanceAttributes: client do error: %v\n", err))
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Print(fmt.Sprintf("getKubEnvInstanceAttributes: %s: %v\n", gciDefaultURI, err))
		return nil, err
	}

	return b, nil
}

// searches a buffer for a specific key string.  the buffer is a assumed to be a
// byte array composed of multiple input lines delinieated by \n's (e.g. stdout)
func getBufferKeyValue(body []byte, key string) string {
	buf := bytes.Split(body, []byte("\n"))
	keyBytes := []byte(key)
	for i := range buf {
		if bytes.Contains(buf[i], keyBytes) {
			return string(bytes.Fields(buf[i])[1])
		}
	}

	return ""
}

// CheckIPTablesVersion checks to ensure we are workign with a supported
// version of the iptables save and restore table/chain/rule formats.
func CheckIPTablesVersion() bool {
	version := os.Getenv("IPTABLES_VERSION")
	if len(version) == 0 {
		version = iptables.DefaultVersion
	}
	match, outstr, err := iptables.VersionCheckCmd(version)
	if err != nil {
		log.Fatal("Can't execute system command 'iptables --version' - quitting!")
	}
	if match {
		log.Print(fmt.Sprintf("Accepting actual version: %s, satisfies requested: %s", strings.TrimSpace(outstr), version))
		return true
	}
	major := os.Getenv("IPTABLES_MAJOR")
	minor := os.Getenv("IPTABLES_MINOR")
	patch := os.Getenv("IPTABLES_PATCH")
	if major == "*" || minor == "*" {
		// If we don't care about the major or minor version, then
		// that's good enough to continue on with.
		log.Print(fmt.Sprintf("Accepting actual version: %s, by wildcarded major|minor numbers", strings.TrimSpace(outstr)))
		return true
	}
	version = major + "." + minor + "."
	if patch != "*" {
		version += patch
	}
	log.Print(fmt.Sprintf("Checking version: %s, major: %s, minor: %s, patch: %s, version string: %s",
		version, major, minor, patch, strings.TrimSpace(outstr)))

	return strings.Contains(outstr, version)
}

// ValidateIPTables checks a iptables-save generated buffer for several
// chacteristics to determine if it meets the needs of our private IP
// address space VPN tunnel routing rules.
func ValidateIPTables(save []byte, clusterIP string) (Indicies, bool) {
	// Validations:
	// 0 - Only validate those rules in the '*nat' iptables-save output (not tested - implied)
	// 1 - If the save buffer contains the 10.0.0.0/8 MASQUERADE rule, then
	//     1.a The save buffer must also contain the derived Cluster IP MASQUERADE rule, and
	//     1.b The 10.0.0.0/8 MASQUERADE rule must come after the Cluster IP MASQUERADE rule
	// If any part of rule 1 is violated, then we make some additional checks to ensure
	// that the rules exactly match the rule format that we expect.

	// -- debug cruft
	// fmt.Fprintf(os.Stdout, ">>>>>>>>>>>>>>>>>>>>>>>>> Begin Input iptables-save Buffer >>>>>>>>>>>>>>>>>>>>>>>>>\n")
	// fmt.Fprintf(os.Stdout, "%s\n", string(save))
	// fmt.Fprintf(os.Stdout, ">>>>>>>>>>>>>>>>>>>>>>>>>> End Input iptables-save Buffer >>>>>>>>>>>>>>>>>>>>>>>>>>\n")

	saveBuf := bytes.Split(save, []byte("\n"))
	clusterIPRule := natPostRoutingPrefix + " ! -d " + clusterIP
	indicies := Indicies{
		iptables.ContainsRulePart(saveBuf, kubenetNATChainRule),
		iptables.ContainsRulePart(saveBuf, clusterIPRule)}

	// Line 0 of every iptables-save buffer is a comment
	// Check 1, 1.a, and 1.b
	if indicies.kubenet > 0 && indicies.cluster > 0 && indicies.kubenet > indicies.cluster {
		// do our extended checking now
		if !strings.HasSuffix(string(saveBuf[indicies.kubenet]), natPostRoutingSuffix) {
			log.Print("Failure: Cluster IP rule suffix is incorrect")
			return indicies, false
		}
		if !strings.HasSuffix(string(saveBuf[indicies.cluster]), natPostRoutingSuffix) {
			log.Print("Failure: Kubenet rule suffix is incorrect.")
			return indicies, false
		}
		return indicies, true
	}
	return indicies, false
}

// ConfigureIPTables forces the iptables-save generated buffer to comply
// with the chacteristics required to meet the needs of our private IP
// address space VPN tunnel routing rules.
func ConfigureIPTables(save []byte, clusterIP string, indicies Indicies) ([]byte, bool) {
	restoreBuf := bytes.Split(save, []byte("\n"))

	// The first fix for all systems we expect to make is to find only the one Kubenet SNAT rule and no Cluster IP SNAT
	if indicies.kubenet > 0 && indicies.cluster == -1 {
		// In this case, we must insert the Cluster IP SNAT Rule, making sure to leave the Kubenet SNAT rule in place,
		// as we know kubenet will forever try to reinsert this rule if it is not present.
		restoreBuf = append(restoreBuf, []byte(""))
		copy(restoreBuf[indicies.kubenet+1:], restoreBuf[indicies.kubenet:])
		restoreBuf[indicies.kubenet] =
			[]byte(natPostRoutingPrefix + " ! -d " + clusterIP + clusterIPSNATComment + natPostRoutingSuffix)

	} else if indicies.kubenet > 0 && indicies.cluster > 0 && indicies.kubenet < indicies.cluster {
		// The easiest of all fixes is to swap the position of Cluster IP SNAT Rule
		// and the Kubenet SNAT rule if they are simply out of order.
		restoreBuf[indicies.kubenet], restoreBuf[indicies.cluster] =
			restoreBuf[indicies.cluster], restoreBuf[indicies.kubenet]
	}

	restore := bytes.Join(restoreBuf, []byte("\n"))
	_, valid := ValidateIPTables(restore, clusterIP)

	return restore, valid
}

func main() {

	log.Print("gci-iptables-conf-agent: main service loop beginning...")
	for {
		time.Sleep(time.Duration(interval) * time.Second)
		ipTables, err := iptables.Save()
		if err != nil {
			log.Print(err)
			continue
		}
		indicies, valid := ValidateIPTables(ipTables, clusterIP)
		log.Print(fmt.Sprintf("Kubenet Rule Index: %d, Cluster IP Rule Index: %d",
			indicies.kubenet, indicies.cluster))
		if valid {
			log.Print("IP Tables NAT table check: ok")
		} else {
			log.Print("Found problem NAT table issue")
			restore, valid := ConfigureIPTables(ipTables, clusterIP, indicies)
			if valid {
				log.Print("NAT table reconfiguration restore buffer created successfully")
				if err := iptables.Restore(restore); err == nil {
					log.Print("iptables-restore successful")
				} else {
					log.Print(fmt.Sprintf("iptables-restore failure: %v", err))
				}
			} else {
				log.Print("NAT table reconfiguration restore buffer creation failed")
			}
		}
	}
}
