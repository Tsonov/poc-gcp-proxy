package main

import (
	"fmt"
	"log"

	"gcp-proxy-twosigma/cast-mock/proxy"
	"gcp-proxy-twosigma/common"
	"gcp-proxy-twosigma/proto"
)

func main() {
	log.Println("Running sanity checks for cast")
	common.RunGCPSanityChecks()

	requestChan, respChan := make(chan *proto.HttpRequest), make(chan *proto.HttpResponse)

	proxy.RunProxyGRPCServer(requestChan, respChan)
	dispatcher := proxy.NewDispatcher(requestChan, respChan)

	dispatcher.Run()

	fmt.Println("enter to continue")
	fmt.Scanln()

	RunTest(dispatcher)
}
