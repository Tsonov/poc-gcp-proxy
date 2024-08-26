package proxy

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"gcp-proxy-twosigma/proto"
	"github.com/google/uuid"
)

// Dispatcher should
//	receive a request;
//	give it UID; save the return destination via UID
//	send via proxy
// 	on response resceive - find UID and dispatch back

type Dispatcher struct {
	pendingRequests map[string]chan *proto.HttpResponse
	locker          sync.Mutex

	ProxyRequestChan  chan<- *proto.HttpRequest
	ProxyResponseChan <-chan *proto.HttpResponse
}

func NewDispatcher(requestChan chan<- *proto.HttpRequest, responseChan <-chan *proto.HttpResponse) *Dispatcher {
	return &Dispatcher{
		pendingRequests:   make(map[string]chan *proto.HttpResponse),
		locker:            sync.Mutex{},
		ProxyRequestChan:  requestChan,
		ProxyResponseChan: responseChan,
	}
}

func (d *Dispatcher) Run() {
	go func() {
		log.Println("starting response returning loop")
		for {
			for resp := range d.ProxyResponseChan {
				waiter := d.findWaiterForResponse(resp.RequestID)
				waiter <- resp
				log.Println("Sent a response back to caller")
			}
		}
	}()
}

func (d *Dispatcher) SendRequest(req *proto.HttpRequest) (<-chan *proto.HttpResponse, error) {
	waiter := d.addRequestToWaitingList(req.RequestID)
	d.ProxyRequestChan <- req
	return waiter, nil
}

func (d *Dispatcher) addRequestToWaitingList(requestID string) <-chan *proto.HttpResponse {
	waiter := make(chan *proto.HttpResponse, 1)
	d.locker.Lock()
	d.pendingRequests[requestID] = waiter
	d.locker.Unlock()
	return waiter
}

func (d *Dispatcher) findWaiterForResponse(requestID string) chan *proto.HttpResponse {
	d.locker.Lock()
	val, ok := d.pendingRequests[requestID]
	if !ok {
		log.Panicln("Trying to send a response for non-existent request", requestID)
	}
	delete(d.pendingRequests, requestID)
	d.locker.Unlock()

	return val
}

type HttpOverGrpcRoundTripper struct {
	dispatcher *Dispatcher
}

func NewHttpOverGrpcRoundTripper(dispatcher *Dispatcher) *HttpOverGrpcRoundTripper {
	return &HttpOverGrpcRoundTripper{dispatcher: dispatcher}
}

func (p *HttpOverGrpcRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	requestID := uuid.New().String()

	headers := make(map[string]string)
	for h, v := range request.Header {
		headers[h] = strings.Join(v, ",")
	}
	protoReq := &proto.HttpRequest{
		RequestID: requestID,
		Method:    request.Method,
		Url:       request.URL.String(),
		Headers:   headers,
		Body: func() []byte {
			if request.Body == nil {
				return []byte{}
			}
			body, err := io.ReadAll(request.Body)
			if err != nil {
				panic(fmt.Sprintf("Failed to read body: %v", err))
			}
			return body
		}(),
	}
	waiter, err := p.dispatcher.SendRequest(protoReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	response := <-waiter
	fmt.Println("Received a response back from dispatcher", requestID, response)

	// Convert to response
	resp := &http.Response{
		Status:     http.StatusText(int(response.Status)),
		StatusCode: int(response.Status),
		Header: func() http.Header {
			headers := make(http.Header)
			for key, value := range response.Headers {
				headers[key] = strings.Split(value, ",")
			}
			return headers
		}(),
		Body:          io.NopCloser(bytes.NewReader(response.Body)),
		ContentLength: int64(len(response.Body)),
		Request:       request,
	}

	return resp, nil
}
