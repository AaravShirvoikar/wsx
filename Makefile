build-client:
	go build -o ./bin/client ./examples/client/main.go

build-server:
	go build -o ./bin/server ./examples/server/main.go

run-client: build-client
	./bin/client

run-server: build-server
	./bin/server
