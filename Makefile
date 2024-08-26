
run-proxy:
	go run ./cluster-proxy/


run-cast:
	go run ./cast-mock/

generate-grpc:
	protoc --go_out=./proto --go-grpc_out=./proto ./proto/proxy.proto