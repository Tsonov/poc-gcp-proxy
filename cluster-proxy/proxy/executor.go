package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"gcp-proxy-twosigma/proto"
	"golang.org/x/oauth2/google"
)

func getGCPCredential() *google.Credentials {
	defaultCreds, err := google.FindDefaultCredentials(context.Background())
	if err != nil {
		log.Panicf("Unable to find default credentials: %v", err)
	}
	return defaultCreds
}

type Executor struct {
	client *http.Client
}

func NewExecutor() *Executor {
	return &Executor{
		client: &http.Client{},
	}
}

func (e *Executor) DoRequest(protoReq *proto.HttpRequest) (*proto.HttpResponse, error) {
	creds := getGCPCredential()

	token, err := creds.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	req, err := http.NewRequest(protoReq.Method, protoReq.Url, bytes.NewReader(protoReq.Body))
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy http request: %w", err)
	}

	// Set the authorize header manually
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)
	for header, val := range protoReq.Headers {
		if strings.ToLower(header) == "authorization" {
			continue
		}
		req.Header.Add(header, val)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unexpected err for %+v: %w", protoReq, err)
	}

	protoResp := &proto.HttpResponse{
		RequestID: protoReq.RequestID,
		Status:    int32(resp.StatusCode),
		Headers:   make(map[string]string),
		Body: func() []byte {
			if resp.Body == nil {
				return []byte{}
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				panic(fmt.Errorf("failed to serialize body from request: %w", err))
			}
			return body
		}(),
	}
	for header, val := range resp.Header {
		protoResp.Headers[header] = strings.Join(val, ",")
	}

	return protoResp, nil
}
