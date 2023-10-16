PROTOS := ./proto/src
PROTO_GEN := ./proto/gen

test:
	go test -v ./...

lint:
	golangci-lint run -E goimports,sqlclosecheck,bodyclose,asciicheck,misspell,errorlint --timeout 5m -e "errors.As" -e "errors.Is" ./...

generate-mocks:
	go generate ./...

.PHONY: \
	test \
	lint \
	generate-mocks \
	fuzz \
	clean \
	protoc

fuzz:
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz FuzzExecutionPayloadUnmarshal ./eth
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz FuzzExecutionPayloadMarshalUnmarshal ./eth
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz FuzzOBP01 ./eth

clean:
	find $(PROTO_GEN) -name "*.pb.go" -type f | xargs rm -rf
	mkdir -p $(PROTO_GEN)

protoc: clean
	protoc -I $(PROTOS) \
	--go_out=$(PROTO_GEN) \
	--go_opt=paths=source_relative \
	--go-grpc_out=$(PROTO_GEN) \
	--go-grpc_opt=paths=source_relative \
	$$(find $(PROTOS) -name '*.proto')
