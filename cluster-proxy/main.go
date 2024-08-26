package main

import (
	"log"

	"gcp-proxy-twosigma/cluster-proxy/proxy"
	"gcp-proxy-twosigma/common"
)

func main() {
	log.Println("Running sanity checks for proxy")
	common.RunGCPSanityChecks()

	proxy.RunProxyClient()
}
