BUF_VER=1.28.1
PROTOC_GEN_GO_VER=1.28.1
PROTOC_GEN_CONNECT_GO_VER=1.18.1

.PHONY: init
init:
	go install github.com/bufbuild/buf/cmd/buf@v${BUF_VER}
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v${PROTOC_GEN_GO_VER}
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@v${PROTOC_GEN_CONNECT_GO_VER}
	go install github.com/redpanda-data/protoc-gen-go-mcp/cmd/protoc-gen-go-mcp@latest

.PHONY: generate
generate:
	buf generate --template buf.gen.sdk.yaml
	buf generate --template buf.gen.yaml