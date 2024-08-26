package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"gcp-proxy-twosigma/cast-mock/proxy"
	"gcp-proxy-twosigma/common"
	"google.golang.org/api/option"
	htransport "google.golang.org/api/transport/http"
)

func RunTest(dispatcher *proxy.Dispatcher) {
	sanityCheck(dispatcher)
	gcpCheck(dispatcher)
}

func sanityCheck(dispatcher *proxy.Dispatcher) {
	normalClient := http.DefaultClient
	resp, err := normalClient.Get("https://google.com")
	if err != nil {
		log.Fatalln("failed sanity get of google", err)
	}
	if resp.StatusCode != 200 {
		log.Fatalln("unexpected status code from google", resp.StatusCode)
	}
	log.Println("Got 200 from google with normal client")

	specialClient := http.Client{
		Transport: proxy.NewHttpOverGrpcRoundTripper(dispatcher),
	}
	resp, err = specialClient.Get("https://google.com")
	if err != nil {
		log.Fatalln("failed special get of google", err)
	}
	if resp.StatusCode != 200 {
		log.Fatalln("unexpected status code from google with special client", resp.StatusCode)
	}
	log.Println("Got 200 from google with special client")
}

func gcpCheck(dispatcher *proxy.Dispatcher) {
	httpClient, _, err := htransport.NewClient(context.Background())
	if err != nil {
		log.Panicf("failed creating http client: %v", err)
	}
	httpClient.Transport = proxy.NewHttpOverGrpcRoundTripper(dispatcher)
	client, err := storage.NewClient(
		context.Background(),
		option.WithoutAuthentication(),
		option.WithHTTPClient(httpClient))

	if err != nil {
		log.Panicf("Failed to create client: %v", err)
	}
	defer func(client *storage.Client) {
		err := client.Close()
		if err != nil {
			log.Panicf("Failed to close client: %v", err)
		}
	}(client)

	bucket := client.Bucket(common.TestBucketName)
	bucketAttrsToUpdate := storage.BucketAttrsToUpdate{}
	labelValue := fmt.Sprintf("timestamp-%v", time.Now().Format("2006-01-02-15-04-05"))
	bucketAttrsToUpdate.SetLabel("successful-test-at", labelValue)

	if _, err := bucket.Update(context.Background(), bucketAttrsToUpdate); err != nil {
		log.Panicf("Failed to update label on test bucket: %v", err)
	}

	log.Println("Successfully updated label on test bucket using the proxy")
}
