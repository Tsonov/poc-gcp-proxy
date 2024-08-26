package common

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
)

const (
	TestBucketURL   = "gs://lachezar-2308/"
	TestBucketName  = "lachezar-2308"
	BucketTestLabel = "timestamp"
)

func RunGCPSanityChecks() {
	checkDefaultAuthUsed()
	addLabelToBucket()
}

func checkDefaultAuthUsed() {
	defaultCreds, err := google.FindDefaultCredentials(context.Background())
	if err != nil {
		log.Panicf("Unable to find default credentials: %v", err)
	}

	log.Printf("Found default credentials: [project: %v, typeof(tokensrc): %T, tokensrc: %+v, jsonCreds: %v",
		defaultCreds.ProjectID, defaultCreds.TokenSource, defaultCreds.TokenSource, string(defaultCreds.JSON))
}

func addLabelToBucket() {
	log.Println("Validating if bucket update is accessible")
	client, err := storage.NewClient(context.Background())
	if err != nil {
		log.Panicf("Failed to create client: %v", err)
	}
	defer func(client *storage.Client) {
		err := client.Close()
		if err != nil {
			log.Panicf("Failed to close client: %v", err)
		}
	}(client)

	bucket := client.Bucket(TestBucketName)
	bucketAttrsToUpdate := storage.BucketAttrsToUpdate{}
	labelValue := fmt.Sprintf("timestamp-%v", time.Now().Format("2006-01-02-15-04-05"))
	bucketAttrsToUpdate.SetLabel(BucketTestLabel, labelValue)

	if _, err := bucket.Update(context.Background(), bucketAttrsToUpdate); err != nil {
		log.Panicf("Failed to update label on test bucket: %v", err)
	}

	log.Println("Successfully updated label on test bucket")
}
