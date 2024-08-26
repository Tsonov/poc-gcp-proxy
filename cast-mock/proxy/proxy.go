package proxy

import (
	"context"
	"io"
	"log"
	"net"
	"sync"

	"gcp-proxy-twosigma/proto"
	"google.golang.org/grpc"
)

type server struct {
	proto.UnimplementedGCPProxyServerServer

	RequestChan  <-chan *proto.HttpRequest
	ResponseChan chan<- *proto.HttpResponse
}

func (s *server) Proxy(stream proto.GCPProxyServer_ProxyServer) error {
	log.Println("Starting a proxy connection from client")

	var wg sync.WaitGroup

	wg.Add(2)

	// TODO: errs
	go func() {
		defer wg.Done()
		log.Println("Starting request sender")

		for req := range s.RequestChan {
			log.Println("Sending request to client:", req)

			if err := stream.Send(req); err != nil {
				log.Printf("Error sending request: %v\n", err)
			}
		}
	}()

	go func() {
		defer wg.Done()
		log.Println("Starting response receiver")

		for {
			in, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Printf("Error in response receiver: %v\n", err)
				return
			}

			log.Printf("Got a response from client: %v\n", in)
			s.ResponseChan <- in
		}
	}()

	wg.Wait()
	log.Println("Closing proxy connection")
	return nil
}

func (s *server) HelloPing(ctx context.Context, req *proto.Ping) (*proto.Ping, error) {
	log.Printf("Received hello-ping from client: %s", req.Hello)
	return &proto.Ping{Hello: "Hello-Ping from Server"}, nil
}

func RunProxyGRPCServer(requestChan <-chan *proto.HttpRequest, responseChan chan<- *proto.HttpResponse) {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterGCPProxyServerServer(grpcServer, &server{
		RequestChan:  requestChan,
		ResponseChan: responseChan,
	})

	log.Println("gRPC server is running on port :50051")
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
}
