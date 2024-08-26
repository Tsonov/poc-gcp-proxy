package main

import (
	"fmt"

	"gcp-proxy-twosigma/common"
)

/*
	Plan:
		1. Sanity checks
			Try to get token/credentials
			Call test endpoint
		2. Create GRPC server that can return stream
		3. Create GRPC client that connects to server and opens stream
		4. In GRPC server, create dummy GCP client that executes operation and expects it to work.


*/

func main() {
	fmt.Println("Hi")

	common.RunGCPSanityChecks()
}
