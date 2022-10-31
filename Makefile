
run:
	go run ./server/main.go

proto.gen.server:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		./server/proto/greeter.proto